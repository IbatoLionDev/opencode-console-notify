/**
 * console-notify — native Windows toast notifications for the OpenCode console.
 *
 * Fires a Windows notification when:
 *   session.idle      -> main session finished a turn and is waiting for input
 *   session.error     -> session hit an error
 *   permission.asked  -> OpenCode is blocked waiting for a permission answer
 *   question.asked    -> OpenCode asked a question
 *
 * Child (sub-agent) sessions are filtered out for idle/error so task
 * completions inside a run do not spam notifications. Blocking events
 * (permission/question) always notify, even from children, because they
 * stop all progress until answered.
 *
 * Windows requirement: the `OpenCode.Notifier` AppUserModelID is
 * auto-registered on load in HKCU\Software\Classes\AppUserModelId
 * (HKCU only, no admin rights needed). Manual registration of that key
 * remains a valid fallback if auto-registration is ever blocked.
 */

import { appendFileSync, readFileSync } from "node:fs"
import { basename, join } from "node:path"
import { homedir } from "node:os"

const AUMID = "OpenCode.Notifier"
const COOLDOWN_MS = 4000
const MAX_LINE = 160
const DEBUG_LOG = join(process.env.TEMP || process.env.TMP || ".", "opencode-console-notify.debug.log")

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
  },
  es: {
    idle: "Tarea terminada — esperando tu entrada",
    error: "Algo salió mal — revisa la consola",
    permission: "Permiso necesario para continuar",
    question: "OpenCode te hizo una pregunta",
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
function spawnHiddenPowerShell(encoded) {
  return Bun.spawn(
    ["powershell.exe", "-NoProfile", "-NonInteractive", "-EncodedCommand", encoded],
    { stdin: "ignore", stdout: "ignore", stderr: "ignore", windowsHide: true },
  )
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
  // UTF-16LE base64 avoids every shell-quoting pitfall between Bun and PowerShell.
  const encoded = Buffer.from(script, "utf16le").toString("base64")
  try {
    const proc = spawnHiddenPowerShell(encoded)
    Promise.resolve(proc.exited).then(
      (code) => {
        if (code !== 0) dbg(`aumid FAIL key=${key} exit=${code}`)
        else dbg(`aumid OK key=${key}`)
      },
      (err) => dbg(`aumid FAIL key=${key} err=${String(err?.message ?? err)}`),
    )
  } catch (err) {
    dbg(`aumid THREW key=${key} err=${String(err?.message ?? err)}`)
  }
}
ensureAumidRegistered()

const lastSent = new Map()
const childCache = new Map()

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
  // UTF-16LE base64 avoids every shell-quoting pitfall between Bun and PowerShell.
  const encoded = Buffer.from(script, "utf16le").toString("base64")
  try {
    const proc = spawnHiddenPowerShell(encoded)
    Promise.resolve(proc.exited).then(
      (code) => {
        if (code !== 0) dbg(`toast FAIL tag=${tag} exit=${code}`)
      },
      (err) => dbg(`toast FAIL tag=${tag} err=${String(err?.message ?? err)}`),
    )
  } catch (err) {
    dbg(`toast THREW tag=${tag} err=${String(err?.message ?? err)}`)
  }
}

export default async ({ client, $, directory }) => {
  const projectName = directory ? basename(directory) : ""
  const title = projectName ? `OpenCode \u00b7 ${projectName}` : "OpenCode"
  dbg(`LOADED directory=${directory ?? "(none)"} title="${title}" has$=${typeof $}`)

  const isChild = async (sessionID) => {
    if (!sessionID) return false
    if (childCache.has(sessionID)) return childCache.get(sessionID)
    let child = false
    try {
      const res = await client.session.get({ path: { id: sessionID } })
      const session = res?.data ?? res
      child = Boolean(session?.parentID)
    } catch {
      child = false // fail open: a missed filter must never swallow a notification
    }
    childCache.set(sessionID, child)
    return child
  }

  const notify = (key, lines) => {
    if (process.platform !== "win32") return
    const now = Date.now()
    if (now - (lastSent.get(key) ?? 0) < COOLDOWN_MS) return
    lastSent.set(key, now)
    // One tag per event kind: same-kind toasts replace each other instead of
    // piling up, while different kinds (question vs idle) never erase one another.
    sendToast(title, lines, key.split(":")[0])
  }

  const handleIdle = async (p) => {
    const settings = loadPluginSettings()
    if (!settings.alerts.sessionIdle) return
    if (await isChild(p.sessionID)) return
    notify(`idle:${p.sessionID ?? "unknown"}`, [TEXTS[settings.lang].idle])
  }

  const handleError = async (p) => {
    const settings = loadPluginSettings()
    if (!settings.alerts.sessionError) return
    if (await isChild(p.sessionID)) return
    const detail = typeof p.error === "string" ? truncate(p.error) : ""
    notify(`error:${p.sessionID ?? "unknown"}`, [TEXTS[settings.lang].error, detail])
  }

  const handlePermission = (p) => {
    const settings = loadPluginSettings()
    if (!settings.alerts.permissionAsked) return
    notify(`perm:${p.id ?? p.sessionID ?? "unknown"}`, [TEXTS[settings.lang].permission])
  }

  const handleQuestion = (p) => {
    const settings = loadPluginSettings()
    if (!settings.alerts.questionAsked) return
    const raw = Array.isArray(p.questions) ? p.questions[0]?.question : undefined
    notify(`question:${p.id ?? p.sessionID ?? "unknown"}`, [
      TEXTS[settings.lang].question,
      raw ? truncate(raw) : "",
    ])
  }

  // Customs fire after the defaults: enabled entries bound to this
  // event type, with the same child-session rule (idle/error filtered,
  // everything else always notifies). One tag per custom id so same-kind
  // toasts replace each other; the shared cooldown still applies.
  const handleCustoms = async (type, p) => {
    const settings = loadPluginSettings()
    for (const c of settings.customs) {
      if (!c.enabled || c.event !== type) continue
      if ((type === "session.idle" || type === "session.error") && await isChild(p.sessionID)) continue
      const lines = c.body ? [c.title, truncate(c.body)] : [c.title]
      notify(`custom:${c.id}`, lines)
    }
  }

  return {
    event: async ({ event }) => {
      try {
        const p = event.properties ?? {}
        if (event.type === "session.idle") {
          await handleIdle(p)
          return handleCustoms(event.type, p)
        }
        if (event.type === "session.error") {
          await handleError(p)
          return handleCustoms(event.type, p)
        }
        if (event.type === "permission.asked") {
          handlePermission(p)
          return handleCustoms(event.type, p)
        }
        if (event.type === "question.asked") {
          handleQuestion(p)
          return handleCustoms(event.type, p)
        }
        return handleCustoms(event.type, p)
      } catch (err) {
        // Notifications must never break the event pipeline.
        dbg(`EVENT handler error type=${event?.type} err=${String(err?.message ?? err)}`)
      }
    },
  }
}
