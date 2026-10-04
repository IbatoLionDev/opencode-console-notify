// Domain: subscribable OpenCode event catalog for custom notifications.
// The user never alters events, only picks one. Mirrors internal/config.
const KNOWN_EVENTS = [
  { type: 'session.created', descEn: 'new session starts', descEs: 'inicia una nueva sesión', noisy: false },
  { type: 'session.updated', descEn: 'session metadata changes', descEs: 'cambian los metadatos de la sesión', noisy: false },
  { type: 'session.deleted', descEn: 'session removed', descEs: 'se elimina la sesión', noisy: false },
  { type: 'session.diff', descEn: 'file changes detected', descEs: 'se detectan cambios en archivos', noisy: false },
  { type: 'session.error', descEn: 'session hit an error', descEs: 'la sesión tuvo un error', noisy: false },
  { type: 'session.status', descEn: 'agent status changes', descEs: 'cambia el estado del agente', noisy: true },
  { type: 'session.idle', descEn: 'task finished, waiting for input', descEs: 'tarea terminada, esperando tu entrada', noisy: false },
  { type: 'session.compacted', descEn: 'session compacted', descEs: 'sesión compactada', noisy: false },
  { type: 'permission.asked', descEn: 'permission needed to continue', descEs: 'permiso necesario para continuar', noisy: false },
  { type: 'permission.replied', descEn: 'permission request answered', descEs: 'petición de permiso respondida', noisy: false },
  { type: 'permission.updated', descEn: 'permission request created or updated', descEs: 'petición de permiso creada o actualizada', noisy: false },
  { type: 'question.asked', descEn: 'OpenCode asked a question', descEs: 'OpenCode hizo una pregunta', noisy: false },
  { type: 'question.replied', descEn: 'question answered', descEs: 'pregunta respondida', noisy: false },
  { type: 'question.rejected', descEn: 'question dismissed', descEs: 'pregunta descartada', noisy: false },
  { type: 'message.updated', descEn: 'a message changed', descEs: 'un mensaje cambió', noisy: true },
  { type: 'message.removed', descEn: 'a message was removed', descEs: 'un mensaje fue eliminado', noisy: false },
  { type: 'message.part.updated', descEn: 'a message part streamed', descEs: 'una parte del mensaje se transmitió', noisy: true },
  { type: 'message.part.removed', descEn: 'a message part was removed', descEs: 'una parte del mensaje fue eliminada', noisy: false },
  { type: 'file.edited', descEn: 'a file was edited', descEs: 'un archivo fue editado', noisy: false },
  { type: 'file.watcher.updated', descEn: 'watched file added, changed or removed', descEs: 'archivo observado añadido, cambiado o eliminado', noisy: true },
  { type: 'command.executed', descEn: 'a slash command ran', descEs: 'un comando slash se ejecutó', noisy: false },
  { type: 'todo.updated', descEn: 'the todo list changed', descEs: 'la lista de tareas cambió', noisy: false },
  { type: 'tui.prompt.append', descEn: 'text appended to the prompt', descEs: 'texto añadido al prompt', noisy: true },
  { type: 'tui.command.execute', descEn: 'a TUI command ran', descEs: 'un comando TUI se ejecutó', noisy: false },
  { type: 'tui.toast.show', descEn: 'the TUI showed a toast', descEs: 'la TUI mostró un aviso', noisy: false },
  { type: 'tui.session.select', descEn: 'another session was selected', descEs: 'otra sesión fue seleccionada', noisy: false },
  { type: 'lsp.updated', descEn: 'language server state changed', descEs: 'el servidor de lenguaje cambió', noisy: false },
  { type: 'lsp.client.diagnostics', descEn: 'new diagnostics arrived', descEs: 'llegaron nuevos diagnósticos', noisy: true },
  { type: 'pty.created', descEn: 'a terminal was created', descEs: 'una terminal fue creada', noisy: false },
  { type: 'pty.updated', descEn: 'terminal output streamed', descEs: 'salida de terminal transmitida', noisy: true },
  { type: 'pty.exited', descEn: 'a terminal exited', descEs: 'una terminal terminó', noisy: false },
  { type: 'pty.deleted', descEn: 'a terminal was removed', descEs: 'una terminal fue eliminada', noisy: false },
  { type: 'vcs.branch.updated', descEn: 'the git branch changed', descEs: 'la rama git cambió', noisy: false },
  { type: 'installation.updated', descEn: 'the installation changed', descEs: 'la instalación cambió', noisy: false },
  { type: 'installation.update.available', descEn: 'a newer OpenCode release exists', descEs: 'existe una versión más nueva de OpenCode', noisy: false },
  { type: 'project.updated', descEn: 'project state changed', descEs: 'el proyecto cambió', noisy: false },
  { type: 'project.directories.updated', descEn: 'project directories changed', descEs: 'los directorios del proyecto cambiaron', noisy: false },
  { type: 'workspace.ready', descEn: 'a workspace is ready', descEs: 'un workspace está listo', noisy: false },
  { type: 'workspace.failed', descEn: 'a workspace failed', descEs: 'un workspace falló', noisy: false },
  { type: 'workspace.status', descEn: 'workspace connection changed', descEs: 'la conexión del workspace cambió', noisy: false },
  { type: 'worktree.ready', descEn: 'a worktree is ready', descEs: 'un worktree está listo', noisy: false },
  { type: 'worktree.failed', descEn: 'a worktree failed', descEs: 'un worktree falló', noisy: false },
  { type: 'server.connected', descEn: 'connected to the server', descEs: 'conectado al servidor', noisy: false },
  { type: 'mcp.tools.changed', descEn: 'an MCP server changed its tools', descEs: 'un servidor MCP cambió sus herramientas', noisy: false },
  { type: 'mcp.browser.open.failed', descEn: 'opening a browser URL failed', descEs: 'falló abrir una URL en el navegador', noisy: false },
  { type: 'ide.installed', descEn: 'the IDE extension was installed', descEs: 'la extensión del IDE fue instalada', noisy: false },
  { type: 'plugin.added', descEn: 'a plugin was added', descEs: 'un plugin fue añadido', noisy: false },
  { type: 'reference.updated', descEn: 'a reference changed', descEs: 'una referencia cambió', noisy: false },
];

function knownEvents() {
  return KNOWN_EVENTS.map((e) => ({ ...e }));
}

function isKnownEvent(type) {
  return KNOWN_EVENTS.some((e) => e.type === type);
}

function eventDescription(type, lang) {
  const found = KNOWN_EVENTS.find((e) => e.type === type);
  if (!found) {
    return type;
  }
  return lang === 'es' ? found.descEs : found.descEn;
}

function isNoisyEvent(type) {
  const found = KNOWN_EVENTS.find((e) => e.type === type);
  return found ? found.noisy : false;
}

export {
  knownEvents,
  isKnownEvent,
  eventDescription,
  isNoisyEvent,
};
