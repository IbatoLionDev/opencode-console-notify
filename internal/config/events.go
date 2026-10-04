// Event catalog: every OpenCode event a custom notification can bind
// to. The user never alters events, only picks one. Source: the OpenCode
// plugin docs event list plus the SDK event union; high-frequency
// families carry a noise warning.
package config

// EventInfo describes one subscribable OpenCode event in both UI
// languages. Noisy events can fire several times per minute.
type EventInfo struct {
	Type   string
	DescEn string
	DescEs string
	Noisy  bool
}

var knownEvents = []EventInfo{
	{Type: "session.created", DescEn: "new session starts", DescEs: "inicia una nueva sesión"},
	{Type: "session.updated", DescEn: "session metadata changes", DescEs: "cambian los metadatos de la sesión"},
	{Type: "session.deleted", DescEn: "session removed", DescEs: "se elimina la sesión"},
	{Type: "session.diff", DescEn: "file changes detected", DescEs: "se detectan cambios en archivos"},
	{Type: "session.error", DescEn: "session hit an error", DescEs: "la sesión tuvo un error"},
	{Type: "session.status", DescEn: "agent status changes", DescEs: "cambia el estado del agente", Noisy: true},
	{Type: "session.idle", DescEn: "task finished, waiting for input", DescEs: "tarea terminada, esperando tu entrada"},
	{Type: "session.compacted", DescEn: "session compacted", DescEs: "sesión compactada"},
	{Type: "permission.asked", DescEn: "permission needed to continue", DescEs: "permiso necesario para continuar"},
	{Type: "permission.replied", DescEn: "permission request answered", DescEs: "petición de permiso respondida"},
	{Type: "permission.updated", DescEn: "permission request created or updated", DescEs: "petición de permiso creada o actualizada"},
	{Type: "question.asked", DescEn: "OpenCode asked a question", DescEs: "OpenCode hizo una pregunta"},
	{Type: "question.replied", DescEn: "question answered", DescEs: "pregunta respondida"},
	{Type: "question.rejected", DescEn: "question dismissed", DescEs: "pregunta descartada"},
	{Type: "message.updated", DescEn: "a message changed", DescEs: "un mensaje cambió", Noisy: true},
	{Type: "message.removed", DescEn: "a message was removed", DescEs: "un mensaje fue eliminado"},
	{Type: "message.part.updated", DescEn: "a message part streamed", DescEs: "una parte del mensaje se transmitió", Noisy: true},
	{Type: "message.part.removed", DescEn: "a message part was removed", DescEs: "una parte del mensaje fue eliminada"},
	{Type: "file.edited", DescEn: "a file was edited", DescEs: "un archivo fue editado"},
	{Type: "file.watcher.updated", DescEn: "watched file added, changed or removed", DescEs: "archivo observado añadido, cambiado o eliminado", Noisy: true},
	{Type: "command.executed", DescEn: "a slash command ran", DescEs: "un comando slash se ejecutó"},
	{Type: "todo.updated", DescEn: "the todo list changed", DescEs: "la lista de tareas cambió"},
	{Type: "tui.prompt.append", DescEn: "text appended to the prompt", DescEs: "texto añadido al prompt", Noisy: true},
	{Type: "tui.command.execute", DescEn: "a TUI command ran", DescEs: "un comando TUI se ejecutó"},
	{Type: "tui.toast.show", DescEn: "the TUI showed a toast", DescEs: "la TUI mostró un aviso"},
	{Type: "tui.session.select", DescEn: "another session was selected", DescEs: "otra sesión fue seleccionada"},
	{Type: "lsp.updated", DescEn: "language server state changed", DescEs: "el servidor de lenguaje cambió"},
	{Type: "lsp.client.diagnostics", DescEn: "new diagnostics arrived", DescEs: "llegaron nuevos diagnósticos", Noisy: true},
	{Type: "pty.created", DescEn: "a terminal was created", DescEs: "una terminal fue creada"},
	{Type: "pty.updated", DescEn: "terminal output streamed", DescEs: "salida de terminal transmitida", Noisy: true},
	{Type: "pty.exited", DescEn: "a terminal exited", DescEs: "una terminal terminó"},
	{Type: "pty.deleted", DescEn: "a terminal was removed", DescEs: "una terminal fue eliminada"},
	{Type: "vcs.branch.updated", DescEn: "the git branch changed", DescEs: "la rama git cambió"},
	{Type: "installation.updated", DescEn: "the installation changed", DescEs: "la instalación cambió"},
	{Type: "installation.update.available", DescEn: "a newer OpenCode release exists", DescEs: "existe una versión más nueva de OpenCode"},
	{Type: "project.updated", DescEn: "project state changed", DescEs: "el proyecto cambió"},
	{Type: "project.directories.updated", DescEn: "project directories changed", DescEs: "los directorios del proyecto cambiaron"},
	{Type: "workspace.ready", DescEn: "a workspace is ready", DescEs: "un workspace está listo"},
	{Type: "workspace.failed", DescEn: "a workspace failed", DescEs: "un workspace falló"},
	{Type: "workspace.status", DescEn: "workspace connection changed", DescEs: "la conexión del workspace cambió"},
	{Type: "worktree.ready", DescEn: "a worktree is ready", DescEs: "un worktree está listo"},
	{Type: "worktree.failed", DescEn: "a worktree failed", DescEs: "un worktree falló"},
	{Type: "server.connected", DescEn: "connected to the server", DescEs: "conectado al servidor"},
	{Type: "mcp.tools.changed", DescEn: "an MCP server changed its tools", DescEs: "un servidor MCP cambió sus herramientas"},
	{Type: "mcp.browser.open.failed", DescEn: "opening a browser URL failed", DescEs: "falló abrir una URL en el navegador"},
	{Type: "ide.installed", DescEn: "the IDE extension was installed", DescEs: "la extensión del IDE fue instalada"},
	{Type: "plugin.added", DescEn: "a plugin was added", DescEs: "un plugin fue añadido"},
	{Type: "reference.updated", DescEn: "a reference changed", DescEs: "una referencia cambió"},
}

// KnownEvents returns the catalog copy in stable order.
func KnownEvents() []EventInfo {
	out := make([]EventInfo, len(knownEvents))
	copy(out, knownEvents)
	return out
}

// IsKnownEvent reports whether t is in the catalog.
func IsKnownEvent(t string) bool {
	for _, e := range knownEvents {
		if e.Type == t {
			return true
		}
	}
	return false
}

// EventDescription returns the one-line description in lang (es/en),
// falling back to English and finally to the type itself.
func EventDescription(t, lang string) string {
	for _, e := range knownEvents {
		if e.Type != t {
			continue
		}
		if lang == "es" {
			return e.DescEs
		}
		return e.DescEn
	}
	return t
}

// IsNoisyEvent reports whether t is flagged high-frequency.
func IsNoisyEvent(t string) bool {
	for _, e := range knownEvents {
		if e.Type == t {
			return e.Noisy
		}
	}
	return false
}
