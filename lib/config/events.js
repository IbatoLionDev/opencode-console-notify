// Domain: subscribable OpenCode event catalog for custom notifications.
// The user never alters events, only picks one. Mirrors internal/config.
// Source: the OpenCode V2 session event taxonomy (packages/schema), the
// first-party notification plugin event names, and the observed live
// server stream. High-frequency families carry a noise warning.
const KNOWN_EVENTS = [
  { type: 'session.created', descEn: 'new session starts', descEs: 'inicia una nueva sesión', noisy: false },
  { type: 'session.deleted', descEn: 'session removed', descEs: 'se elimina la sesión', noisy: false },
  { type: 'session.renamed', descEn: 'session renamed', descEs: 'sesión renombrada', noisy: false },
  { type: 'session.moved', descEn: 'session moved', descEs: 'sesión movida', noisy: false },
  { type: 'session.agent.selected', descEn: 'session agent changed', descEs: 'cambió el agente de la sesión', noisy: false },
  { type: 'session.model.selected', descEn: 'session model changed', descEs: 'cambió el modelo de la sesión', noisy: false },
  { type: 'session.metadata.updated', descEn: 'session metadata changed', descEs: 'cambiaron los metadatos de la sesión', noisy: false },
  { type: 'session.permissions', descEn: 'session permission rules replaced', descEs: 'reglas de permiso reemplazadas', noisy: false },
  { type: 'session.viewed', descEn: 'session viewed (idle watermark)', descEs: 'sesión vista (marca de inactividad)', noisy: false },
  { type: 'session.forked', descEn: 'session forked', descEs: 'sesión bifurcada', noisy: false },
  { type: 'session.usage.updated', descEn: 'usage totals updated', descEs: 'uso actualizado', noisy: true },
  { type: 'session.execution.started', descEn: 'a run started', descEs: 'empezó una ejecución', noisy: false },
  { type: 'session.execution.succeeded', descEn: 'run finished, waiting for input', descEs: 'ejecución terminada, esperando tu entrada', noisy: false },
  { type: 'session.execution.failed', descEn: 'run failed', descEs: 'la ejecución falló', noisy: false },
  { type: 'session.execution.interrupted', descEn: 'run interrupted', descEs: 'ejecución interrumpida', noisy: false },
  { type: 'session.inbox.enqueued', descEn: 'prompt admitted', descEs: 'prompt admitido', noisy: false },
  { type: 'session.inbox.delivered', descEn: 'prompt delivered', descEs: 'prompt entregado', noisy: false },
  { type: 'session.inbox.cancelled', descEn: 'prompt cancelled', descEs: 'prompt cancelado', noisy: false },
  { type: 'session.inbox.delivery.changed', descEn: 'prompt delivery changed', descEs: 'cambió la entrega del prompt', noisy: false },
  { type: 'session.instructions.updated', descEn: 'instructions updated', descEs: 'instrucciones actualizadas', noisy: false },
  { type: 'session.synthetic', descEn: 'synthetic message added', descEs: 'mensaje sintético añadido', noisy: false },
  { type: 'session.skill.activated', descEn: 'a skill activated', descEs: 'skill activado', noisy: false },
  { type: 'session.shell.started', descEn: 'session shell started', descEs: 'shell de sesión iniciado', noisy: false },
  { type: 'session.shell.ended', descEn: 'session shell ended', descEs: 'shell de sesión terminado', noisy: false },
  { type: 'session.step.started', descEn: 'agent step started', descEs: 'paso del agente iniciado', noisy: true },
  { type: 'session.step.streamed', descEn: 'agent step streamed', descEs: 'paso del agente transmitido', noisy: true },
  { type: 'session.step.ended', descEn: 'agent step ended', descEs: 'paso del agente terminado', noisy: true },
  { type: 'session.step.failed', descEn: 'agent step failed', descEs: 'paso del agente falló', noisy: false },
  { type: 'session.text.started', descEn: 'assistant text started', descEs: 'texto del asistente iniciado', noisy: false },
  { type: 'session.text.delta', descEn: 'assistant text streamed', descEs: 'texto del asistente transmitido', noisy: true },
  { type: 'session.text.ended', descEn: 'assistant text completed', descEs: 'texto del asistente completado', noisy: false },
  { type: 'session.reasoning.started', descEn: 'thinking started', descEs: 'razonamiento iniciado', noisy: false },
  { type: 'session.reasoning.delta', descEn: 'thinking streamed', descEs: 'razonamiento transmitido', noisy: true },
  { type: 'session.reasoning.ended', descEn: 'thinking completed', descEs: 'razonamiento completado', noisy: false },
  { type: 'session.tool.input.started', descEn: 'tool input started', descEs: 'entrada de herramienta iniciada', noisy: false },
  { type: 'session.tool.input.delta', descEn: 'tool input streamed', descEs: 'entrada de herramienta transmitida', noisy: true },
  { type: 'session.tool.input.ended', descEn: 'tool input completed', descEs: 'entrada de herramienta completada', noisy: false },
  { type: 'session.tool.called', descEn: 'tool called', descEs: 'herramienta invocada', noisy: false },
  { type: 'session.tool.progress', descEn: 'tool progress updated', descEs: 'progreso actualizado', noisy: true },
  { type: 'session.tool.success', descEn: 'tool succeeded', descEs: 'herramienta exitosa', noisy: false },
  { type: 'session.tool.failed', descEn: 'tool failed', descEs: 'herramienta falló', noisy: false },
  { type: 'session.retry.scheduled', descEn: 'retry scheduled', descEs: 'reintento programado', noisy: false },
  { type: 'session.compaction.started', descEn: 'compaction started', descEs: 'compactación iniciada', noisy: false },
  { type: 'session.compaction.delta', descEn: 'compaction streamed', descEs: 'compactación transmitida', noisy: true },
  { type: 'session.compaction.ended', descEn: 'compaction completed', descEs: 'compactación completada', noisy: false },
  { type: 'session.compaction.failed', descEn: 'compaction failed', descEs: 'compactación falló', noisy: false },
  { type: 'session.revert.staged', descEn: 'revert staged', descEs: 'reversión preparada', noisy: false },
  { type: 'session.revert.cleared', descEn: 'staged revert cleared', descEs: 'reversión preparada descartada', noisy: false },
  { type: 'session.revert.committed', descEn: 'revert committed', descEs: 'reversión aplicada', noisy: false },
  { type: 'permission.asked', descEn: 'permission needed to continue', descEs: 'permiso necesario para continuar', noisy: false },
  { type: 'permission.replied', descEn: 'permission request answered', descEs: 'petición de permiso respondida', noisy: false },
  { type: 'form.created', descEn: 'OpenCode asked a question', descEs: 'OpenCode hizo una pregunta', noisy: false },
  { type: 'form.replied', descEn: 'question answered', descEs: 'pregunta respondida', noisy: false },
  { type: 'form.cancelled', descEn: 'question dismissed', descEs: 'pregunta descartada', noisy: false },
  { type: 'server.connected', descEn: 'connected to the server', descEs: 'conectado al servidor', noisy: false },
  { type: 'skill.updated', descEn: 'skill registry changed', descEs: 'registro de skills cambió', noisy: true },
  { type: 'model.updated', descEn: 'model list changed', descEs: 'lista de modelos cambió', noisy: false },
  { type: 'provider.updated', descEn: 'provider list changed', descEs: 'lista de proveedores cambió', noisy: false },
  { type: 'project.updated', descEn: 'project state changed', descEs: 'el proyecto cambió', noisy: false },
  { type: 'agent.updated', descEn: 'agent list changed', descEs: 'lista de agentes cambió', noisy: false },
  { type: 'reference.updated', descEn: 'a reference changed', descEs: 'una referencia cambió', noisy: false },
  { type: 'plugin.updated', descEn: 'plugin list changed', descEs: 'lista de plugins cambió', noisy: false },
  { type: 'command.updated', descEn: 'command list changed', descEs: 'lista de comandos cambió', noisy: false },
  { type: 'websearch.updated', descEn: 'websearch providers changed', descEs: 'proveedores de búsqueda cambiaron', noisy: false },
  { type: 'integration.updated', descEn: 'integration list changed', descEs: 'lista de integraciones cambió', noisy: false },
  { type: 'shell.created', descEn: 'a shell was created', descEs: 'shell creado', noisy: false },
  { type: 'shell.deleted', descEn: 'a shell was removed', descEs: 'shell eliminado', noisy: false },
  { type: 'shell.exited', descEn: 'a shell exited', descEs: 'shell terminado', noisy: false },
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
