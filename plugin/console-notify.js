/**
 * console-notify — native toast notifications for the OpenCode console.
 *
 * Fires a notification when (V2 names; V1 names in parentheses):
 *   session.execution.succeeded (session.idle) -> a run finished, waiting for input
 *   session.execution.failed (session.error)  -> a run hit an error
 *   session.execution.interrupted             -> a run was interrupted
 *   permission.asked  -> OpenCode is blocked waiting for a permission answer
 *   form.created (question.asked) -> OpenCode asked a question
 *
 * Child (sub-agent) sessions are filtered out for completion/error so task
 * completions inside a run do not spam notifications. Blocking events
 * (permission/question) always notify, even from children, because they
 * stop all progress until answered.
 *
 * Windows requirement: the `OpenCode.Notifier` AppUserModelID is
 * auto-registered on load in HKCU\Software\Classes\AppUserModelId
 * (HKCU only, no admin rights needed). Manual registration of that key
 * remains a valid fallback if auto-registration is ever blocked.
 *
 * Entrypoint (dual V1/V2, no plugin SDK import needed):
 *   export default { id: "console-notify", server, setup }
 * OpenCode V1 (1.18.29+) calls `server()` and receives the `{ event }`
 * hook; OpenCode V2 calls `setup(ctx)` and the plugin subscribes to
 * `ctx.event` itself. V2 events carry their payload in `event.data`
 * (V1 used `event.properties`); the dispatcher accepts both.
 *
 * Process spawning uses `node:child_process` only, so the module loads
 * with or without a `Bun` global in the host.
 */

import { spawn, spawnSync } from "node:child_process"
import { appendFileSync, readFileSync } from "node:fs"
import { basename, join } from "node:path"
import { homedir } from "node:os"

const PLUGIN_ID = "console-notify"
const AUMID = "OpenCode.Notifier"
const COOLDOWN_MS = 4000
const MAX_LINE = 160
const DEBUG_LOG = join(process.env.TEMP || process.env.TMP || process.env.TMPDIR || (process.platform === "win32" ? "." : "/tmp"), "opencode-console-notify.debug.log")

// V2 closes the event stream when the server restarts; reconnect with a
// bounded doubling delay that resets once events flow again.
const EVENT_RETRY_MIN_MS = 50
const EVENT_RETRY_MAX_MS = 5000

// Config file written by the config command (v2.0.0, customs since 2.2).
// Missing or corrupt means v1 behavior: everything on, English.
function loadPluginSettings() {
  const fallback = { lang: "en", alerts: { sessionIdle: true, sessionError: true, permissionAsked: true, questionAsked: true }, customs: [] }
  try {
    const home = homedir()
    if (!home) return fallback
    const raw = readFileSync(join(home, ".config", "opencode", "plugins", "console-notify.config.json"), "utf8")
    const parsed = JSON.parse(raw)
    const alerts = parsed?.alerts ?? {}
    const customs = Array.isArray(parsed?.customAlerts) ? parsed.customAlerts : []
    return {
      lang: parsed?.lang === "es" ? "es" : "en",
      alerts: {
        sessionIdle: alerts.sessionIdle !== false,
        sessionError: alerts.sessionError !== false,
        permissionAsked: alerts.permissionAsked !== false,
        questionAsked: alerts.questionAsked !== false,
      },
      // Keep only entries that can fire; one bad entry must never break
      // the file (same rule as the CLI normalizer).
      customs: customs.filter((c) => c && typeof c.id === "string" && typeof c.event === "string" && typeof c.title === "string" && c.title.trim() !== "").map((c) => ({
        id: c.id,
        event: c.event,
        title: String(c.title),
        body: typeof c.body === "string" ? c.body : "",
        enabled: c.enabled !== false,
      })),
    }
  } catch {
    return fallback
  }
}

const TEXTS = {
  en: {
    idle: "Task finished — waiting for your input",
    error: "Something went wrong — check the console",
    permission: "Permission needed to continue",
    question: "OpenCode asked you a question",
    interrupted: "Session done",
  },
  es: {
    idle: "Tarea terminada — esperando tu entrada",
    error: "Algo salió mal — revisa la consola",
    permission: "Permiso necesario para continuar",
    question: "OpenCode te hizo una pregunta",
    interrupted: "Sesión terminada",
  },
}

function dbg(message) {
  try {
    appendFileSync(DEBUG_LOG, `${new Date().toISOString()} ${message}\n`)
  } catch {
    // Debug logging must never break anything.
  }
}

// windowsHide (Win32 CREATE_NO_WINDOW) is load-bearing, NOT cosmetic:
// without it PowerShell inherits our console window and `-WindowStyle Hidden`
// runs ShowWindow(GetConsoleWindow(), SW_HIDE) on the USER's console window,
// which is what made a maximized console vanish on every notification.
// With CREATE_NO_WINDOW the child gets its own console with no window, so
// there is nothing to flash and nothing shared to hide.
function spawnHiddenPowerShell(encoded, tag) {
  try {
    const child = spawn(
      "powershell.exe",
      ["-NoProfile", "-NonInteractive", "-EncodedCommand", encoded],
      { stdio: "ignore", windowsHide: true },
    )
    child.on("error", (err) => dbg(`spawn FAIL tag=${tag} err=${String(err?.message ?? err)}`))
    child.on("exit", (code) => {
      if (code !== 0) dbg(`spawn FAIL tag=${tag} exit=${code}`)
    })
    if (typeof child.unref === "function") child.unref()
    return child
  } catch (err) {
    dbg(`spawn THREW tag=${tag} err=${String(err?.message ?? err)}`)
    return undefined
  }
}

/**
 * Ensure the `OpenCode.Notifier` AppUserModelID is registered for the
 * current user (HKCU only, never admin). Idempotent, fire-and-forget:
 * never throws, never blocks plugin loading. The Test-Path /
 * Get-ItemProperty guards make re-running a no-op.
 */
function ensureAumidRegistered() {
  if (process.platform !== "win32") return
  const key = `HKCU:\\Software\\Classes\\AppUserModelId\\${AUMID}`
  const script = [
    `if (-not (Test-Path -LiteralPath '${key}')) { New-Item -Path '${key}' -Force | Out-Null }`,
    `$d = Get-ItemProperty -LiteralPath '${key}' -Name DisplayName -ErrorAction SilentlyContinue`,
    `if ($null -eq $d) { Set-ItemProperty -LiteralPath '${key}' -Name DisplayName -Value 'OpenCode' -Type String }`,
  ].join("\n")
  // UTF-16LE base64 avoids every shell-quoting pitfall for the command line.
  const encoded = Buffer.from(script, "utf16le").toString("base64")
  const child = spawnHiddenPowerShell(encoded, "aumid-register")
  if (!child) return
  child.on("exit", (code) => {
    if (code === 0) dbg(`aumid OK key=${key}`)
  })
}
ensureAumidRegistered()

const lastSent = new Map()
// sessionCache: sessionID -> { isChild, projectID } for sessions already
// resolved through the host. Failures are not cached: a session that is
// temporarily unreadable is retried on the next event (fail open).
const sessionCache = new Map()
// terminalSessions: "done:<id>" / "failed:<id>" markers so one execution
// notifies at most once (failed-then-succeeded, duplicate delivery).
// session.execution.started clears them for the next execution.
const terminalSessions = new Map()
// askedForms / askedPermissions: prompt ids already notified and still
// unanswered, so a re-asked prompt notifies again after reply/cancel.
const askedForms = new Set()
const askedPermissions = new Set()

// XML 1.0 forbids C0/C1 controls and lone surrogates; one illegal character makes LoadXml throw and the toast is lost.
function isXmlCodePointAllowed(cp) {
  return cp === 9 || cp === 10 || cp === 13 || (cp >= 32 && cp <= 0x7e) || (cp >= 0xa0 && cp <= 0xd7ff) || (cp >= 0xe000 && cp <= 0xfffd) || (cp >= 0x10000 && cp <= 0x10ffff)
}

function xmlSafe(value) {
  let out = ""
  for (const ch of String(value)) {
    if (isXmlCodePointAllowed(ch.codePointAt(0))) out += ch
  }
  return out
}

function escapeXml(value) {
  return xmlSafe(value)
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;")
    .replace(/"/g, "&quot;")
    .replace(/'/g, "&apos;")
}

function truncate(value) {
  const text = String(value).replace(/\s+/g, " ").trim()
  if (text.length <= MAX_LINE) return text
  let end = MAX_LINE - 3
  if (end > 0 && text.charCodeAt(end) >= 0xdc00 && text.charCodeAt(end) <= 0xdfff) end -= 1
  return `${text.slice(0, end)}...`
}

// errorDetail extracts a one-line message from a V1 error string or a V2
// SessionError object ({ message, ... }).
function errorDetail(err) {
  if (typeof err === "string") return truncate(err)
  if (err && typeof err.message === "string" && err.message.trim() !== "") return truncate(err.message)
  if (err && typeof err.name === "string" && err.name.trim() !== "") return truncate(err.name)
  return ""
}

// notify-send availability is probed once: most desktops ship
// libnotify-bin, headless boxes do not, and missing must stay silent
// (doctor reports it) rather than throwing on every event.
let notifySendChecked = null
function notifySendAvailable() {
  if (notifySendChecked !== null) return notifySendChecked
  try {
    const probed = spawnSync("notify-send", ["--version"], { stdio: "ignore", windowsHide: true, timeout: 5000 })
    notifySendChecked = !probed.error && probed.status === 0
  } catch {
    notifySendChecked = false
  }
  return notifySendChecked
}

let osascriptChecked = null
function osascriptAvailable() {
  if (osascriptChecked !== null) return osascriptChecked
  try {
    const probed = spawnSync("osascript", ["-e", "return \"ok\""], {
      stdio: "ignore",
      windowsHide: true,
      timeout: 5000,
    })
    osascriptChecked = !probed.error && probed.status === 0
  } catch {
    osascriptChecked = false
  }
  return osascriptChecked
}

// sendToastLinux mirrors sendToast over notify-send: same title/body,
// same fire-and-forget shape. Tags/cooldowns stay in notify(); replace
// semantics are a Windows-only nicety notify-send cannot promise.
function sendToastLinux(title, lines, tag) {
  if (!notifySendAvailable()) {
    dbg(`toast SKIP tag=${tag} reason=no-notify-send`)
    return
  }
  const body = [...lines].filter(Boolean).join("\n")
  try {
    const child = spawn(
      "notify-send",
      ["--app-name=OpenCode", "--urgency=normal", "--expire-time=8000", title, body],
      { stdio: "ignore", windowsHide: true },
    )
    child.on("error", (err) => dbg(`toast FAIL tag=${tag} err=${String(err?.message ?? err)}`))
    child.on("exit", (code) => {
      if (code !== 0) dbg(`toast FAIL tag=${tag} exit=${code}`)
    })
    if (typeof child.unref === "function") child.unref()
    dbg(`toast SENT tag=${tag} title="${title}"`)
  } catch (err) {
    dbg(`toast THREW tag=${tag} err=${String(err?.message ?? err)}`)
  }
}

// sendToastDarwin mirrors sendToast over osascript: same title/body,
// same fire-and-forget shape. Tags/cooldowns stay in notify().
function sendToastDarwin(title, lines, tag) {
  if (!osascriptAvailable()) {
    dbg(`toast SKIP tag=${tag} reason=no-osascript`)
    return
  }
  const body = [...lines].filter(Boolean).join("\n")
  try {
    const child = spawn(
      "osascript",
      ["-e", `display notification ${strAppleScript(body)} with title ${strAppleScript(title)}`],
      { stdio: "ignore", windowsHide: true },
    )
    child.on("error", (err) => dbg(`toast FAIL tag=${tag} err=${String(err?.message ?? err)}`))
    child.on("exit", (code) => {
      if (code !== 0) dbg(`toast FAIL tag=${tag} exit=${code}`)
    })
    if (typeof child.unref === "function") child.unref()
    dbg(`toast SENT tag=${tag} title="${title}"`)
  } catch (err) {
    dbg(`toast THREW tag=${tag} err=${String(err?.message ?? err)}`)
  }
}

function strAppleScript(value) {
  const esc = String(value).replace(/\\/g, "\\\\").replace(/"/g, '\\"')
  return '"' + esc + '"'
}

function sendToast(title, lines, tag) {
  const texts = [title, ...lines].filter(Boolean)
    .map((line) => `<text>${escapeXml(line)}</text>`)
    .join("")
  const xml = `<toast duration="long"><visual><binding template="ToastGeneric">${texts}</binding></visual><audio src="ms-winsoundevent:Notification.Default"/></toast>`
  const script = [
    "[Windows.UI.Notifications.ToastNotificationManager, Windows.UI.Notifications, ContentType = WindowsRuntime] | Out-Null",
    "[Windows.Data.Xml.Dom.XmlDocument, Windows.Data.Xml.Dom.XmlDocument, ContentType = WindowsRuntime] | Out-Null",
    "$x = New-Object Windows.Data.Xml.Dom.XmlDocument",
    `$x.LoadXml('${xml}')`,
    "$t = [Windows.UI.Notifications.ToastNotification]::new($x)",
    `$t.Tag = '${escapeXml(tag)}'`,
    "$t.Group = 'opencode'",
    `[Windows.UI.Notifications.ToastNotificationManager]::CreateToastNotifier('${AUMID}').Show($t)`,
  ].join("\n")
  // UTF-16LE base64 avoids every shell-quoting pitfall for the command line.
  const encoded = Buffer.from(script, "utf16le").toString("base64")
  if (spawnHiddenPowerShell(encoded, tag)) dbg(`toast SENT tag=${tag} title="${title}"`)
}

// Shared notification pipeline: both entrypoints (V1 `server`, V2 `setup`)
// feed (eventType, payload) here. `getSession(sessionID)` resolves a
// session or throws/returns undefined when it cannot be read.
// `locationProjectID` scopes V2 notifications to sessions owned by this
// plugin instance's location: the server delivers every location's events
// to every instance, so without it one event would toast once per loaded
// location. V1 leaves it undefined (the V1 host already scopes events).
function createPipeline({ getSession, directory, locationProjectID }) {
  const projectName = directory ? basename(directory) : ""
  const title = projectName ? `OpenCode \u00b7 ${projectName}` : "OpenCode"

  // resolveSession reads (and caches) the ownership facts every handler
  // needs: whether the session is a sub-agent child and which project it
  // belongs to. Unknown sessions resolve to undefined and handlers fail
  // open: a missed lookup must never swallow a notification.
  const resolveSession = async (sessionID) => {
    if (!sessionID) return undefined
    if (sessionCache.has(sessionID)) return sessionCache.get(sessionID)
    let info
    try {
      const res = await getSession(sessionID)
      const session = res?.data ?? res
      if (!session) return undefined
      info = {
        isChild: Boolean(session?.parentID),
        projectID: typeof session?.projectID === "string" ? session.projectID : undefined,
      }
    } catch {
      return undefined
    }
    sessionCache.set(sessionID, info)
    return info
  }

  const isChild = async (sessionID) => {
    const info = await resolveSession(sessionID)
    return info?.isChild ?? false
  }

  // ownsSession reports whether this instance is the session's owner.
  // It only ever suppresses on positive evidence (both project ids known
  // and different); unknown projects on either side fail open so a missed
  // lookup can duplicate but never swallow a notification.
  const ownsSession = async (sessionID) => {
    if (!locationProjectID) return true
    const info = await resolveSession(sessionID)
    if (!info || !info.projectID) return true
    return info.projectID === locationProjectID
  }

  const notify = (key, lines) => {
    const now = Date.now()
    if (now - (lastSent.get(key) ?? 0) < COOLDOWN_MS) return
    lastSent.set(key, now)
    // One tag per event kind: same-kind toasts replace each other instead of
    // piling up on Windows, while different kinds (question vs idle) never
    // erase one another. notify-send cannot promise replace semantics.
    if (process.platform === "win32") {
      sendToast(title, lines, key.split(":")[0])
      return
    }
    if (process.platform === "linux") {
      sendToastLinux(title, lines, key.split(":")[0])
      return
    }
    if (process.platform === "darwin") {
      sendToastDarwin(title, lines, key.split(":")[0])
    }
  }

  const handleStarted = (p) => {
    const sid = p.sessionID ?? "unknown"
    terminalSessions.delete(`done:${sid}`)
    terminalSessions.delete(`failed:${sid}`)
  }

  const handleDone = async (p) => {
    const settings = loadPluginSettings()
    if (!settings.alerts.sessionIdle) return
    const sid = p.sessionID ?? "unknown"
    if (terminalSessions.has(`done:${sid}`) || terminalSessions.has(`failed:${sid}`)) return
    if (!(await ownsSession(p.sessionID))) return
    if (await isChild(p.sessionID)) return
    terminalSessions.set(`done:${sid}`, true)
    notify(`idle:${sid}`, [TEXTS[settings.lang].idle])
  }

  const handleInterrupted = async (p) => {
    const settings = loadPluginSettings()
    if (!settings.alerts.sessionIdle) return
    const sid = p.sessionID ?? "unknown"
    if (terminalSessions.has(`done:${sid}`) || terminalSessions.has(`failed:${sid}`)) return
    if (!(await ownsSession(p.sessionID))) return
    if (await isChild(p.sessionID)) return
    terminalSessions.set(`done:${sid}`, true)
    notify(`idle:${sid}`, [TEXTS[settings.lang].interrupted])
  }

  const handleFailed = async (p) => {
    const settings = loadPluginSettings()
    if (!settings.alerts.sessionError) return
    const sid = p.sessionID ?? "unknown"
    if (terminalSessions.has(`done:${sid}`) || terminalSessions.has(`failed:${sid}`)) return
    if (!(await ownsSession(p.sessionID))) return
    if (await isChild(p.sessionID)) return
    terminalSessions.set(`failed:${sid}`, true)
    notify(`error:${sid}`, [TEXTS[settings.lang].error, errorDetail(p.error)])
  }

  const handlePermission = async (p) => {
    const settings = loadPluginSettings()
    if (!settings.alerts.permissionAsked) return
    const id = p.id ?? p.sessionID ?? "unknown"
    if (askedPermissions.has(id)) return
    if (!(await ownsSession(p.sessionID))) return
    askedPermissions.add(id)
    notify(`perm:${id}`, [TEXTS[settings.lang].permission])
  }

  const handlePermissionReplied = (p) => {
    const id = p.requestID ?? p.id
    if (id === undefined) return
    askedPermissions.delete(id)
    lastSent.delete(`perm:${id}`)
  }

  const handleQuestion = async (p) => {
    const settings = loadPluginSettings()
    if (!settings.alerts.questionAsked) return
    // V2 nests the prompt under `form`; V1 carries `questions`.
    const f = p.form ?? {}
    const id = f.id ?? p.id ?? p.sessionID ?? "unknown"
    if (askedForms.has(id)) return
    if (!(await ownsSession(f.sessionID ?? p.sessionID))) return
    askedForms.add(id)
    const raw = typeof f.title === "string" && f.title.trim() !== ""
      ? f.title
      : Array.isArray(p.questions) ? p.questions[0]?.question : undefined
    notify(`question:${id}`, [
      TEXTS[settings.lang].question,
      raw ? truncate(raw) : "",
    ])
  }

  const handleFormClosed = (p) => {
    const id = p.form?.id ?? p.id
    if (id === undefined) return
    askedForms.delete(id)
    lastSent.delete(`question:${id}`)
  }

  // Customs fire after the defaults: enabled entries bound to this
  // event type, with the same child-session rule for completion/error
  // families (idle/error on either runtime), everything else always
  // notifies. One tag per custom id so same-kind toasts replace each
  // other; the shared cooldown still applies.
  const CHILD_FILTERED = new Set(["session.idle", "session.error", "session.execution.succeeded", "session.execution.failed"])
  const handleCustoms = async (type, p) => {
    const settings = loadPluginSettings()
    for (const c of settings.customs) {
      if (!c.enabled || c.event !== type) continue
      if (!(await ownsSession(p.sessionID))) continue
      if (CHILD_FILTERED.has(type) && await isChild(p.sessionID)) continue
      const lines = c.body ? [c.title, truncate(c.body)] : [c.title]
      notify(`custom:${c.id}`, lines)
    }
  }

  // Event names from both runtimes route here. V1 emits session.idle,
  // session.error and question.asked; V2 emits session.execution.*,
  // permission.* and form.* (permission.asked exists on both).
  const HANDLED = new Set([
    "session.idle", "session.error", "question.asked",
    "session.execution.started", "session.execution.succeeded",
    "session.execution.failed", "session.execution.interrupted",
    "permission.asked", "permission.replied",
    "form.created", "form.replied", "form.cancelled",
  ])
  const dispatch = async (type, payload) => {
    try {
      const p = payload ?? {}
      if (HANDLED.has(type)) dbg(`DISPATCH type=${type} sid=${p.sessionID ?? "(none)"}`)
      if (type === "session.idle" || type === "session.execution.succeeded") {
        await handleDone(p)
        return handleCustoms(type, p)
      }
      if (type === "session.error" || type === "session.execution.failed") {
        await handleFailed(p)
        return handleCustoms(type, p)
      }
      if (type === "session.execution.interrupted") {
        await handleInterrupted(p)
        return handleCustoms(type, p)
      }
      if (type === "session.execution.started") {
        handleStarted(p)
        return handleCustoms(type, p)
      }
      if (type === "permission.asked") {
        await handlePermission(p)
        return handleCustoms(type, p)
      }
      if (type === "permission.replied") {
        handlePermissionReplied(p)
        return handleCustoms(type, p)
      }
      if (type === "question.asked" || type === "form.created") {
        await handleQuestion(p)
        return handleCustoms(type, p)
      }
      if (type === "form.replied" || type === "form.cancelled") {
        handleFormClosed(p)
        return handleCustoms(type, p)
      }
      return handleCustoms(type, p)
    } catch (err) {
      // Notifications must never break the event pipeline.
      dbg(`EVENT handler error type=${type} err=${String(err?.message ?? err)}`)
    }
  }

  return { dispatch, title }
}

// OpenCode V1 entrypoint (1.18.29+ calls `server()` on definition objects).
async function server({ client, directory }) {
  const pipeline = createPipeline({
    directory,
    getSession: (sessionID) => client.session.get({ path: { id: sessionID } }),
  })
  dbg(`LOADED runtime=v1 directory=${directory ?? "(none)"} title="${pipeline.title}"`)

  return {
    event: async ({ event }) => {
      const payload = event?.properties ?? {}
      return pipeline.dispatch(event?.type, payload)
    },
  }
}

function delayUnlessAborted(ms, signal) {
  return new Promise((resolve) => {
    if (signal.aborted) return resolve()
    const done = () => {
      clearTimeout(timer)
      signal.removeEventListener("abort", done)
      resolve()
    }
    const timer = setTimeout(done, ms)
    signal.addEventListener("abort", done, { once: true })
  })
}

// OpenCode V2 entrypoint: subscribe to the server event stream and feed
// the shared pipeline. V2 events carry the payload in `event.data`.
async function setup(ctx) {
  const directory = ctx.location.directory
  const pipeline = createPipeline({
    directory,
    locationProjectID: ctx.location.project?.id,
    // V1 SDK results carry `{ data, error }`; the V2 client throws instead.
    getSession: (sessionID) => ctx.session.get({ sessionID }),
  })
  dbg(`LOADED runtime=v2 id=${PLUGIN_ID} directory=${directory ?? "(none)"} project=${ctx.location.project?.id ?? "(none)"} title="${pipeline.title}"`)

  const abort = new AbortController()
  let listening = Promise.resolve()
  const cleanup = async () => {
    abort.abort()
    await listening
  }

  try {
    listening = (async () => {
      let retryMs = EVENT_RETRY_MIN_MS
      while (!abort.signal.aborted) {
        try {
          for await (const event of ctx.event.subscribe({ signal: abort.signal })) {
            if (abort.signal.aborted) break
            // Every subscription opens with a server.connected handshake; only a
            // real event proves the stream is healthy enough to reset the backoff.
            if (event?.type !== "server.connected") retryMs = EVENT_RETRY_MIN_MS
            const payload = event?.data ?? event?.properties ?? {}
            await pipeline.dispatch(event?.type, payload)
          }
        } catch (err) {
          // The stream ends or throws when the server restarts; events missed
          // while disconnected are lost, so reconnect below.
          dbg(`SUBSCRIBE error=${String(err?.message ?? err)}`)
        }
        if (abort.signal.aborted) break
        await delayUnlessAborted(retryMs, abort.signal)
        retryMs = Math.min(retryMs * 2, EVENT_RETRY_MAX_MS)
      }
    })()
  } catch (cause) {
    await cleanup()
    throw cause
  }

  return cleanup
}

// V1 (1.18.29+) calls server(); V2 calls setup(). Kept as the single
// default export on purpose: older V1 loaders call every exported
// function as a plugin factory, so no named exports.
export default { id: PLUGIN_ID, server, setup }
