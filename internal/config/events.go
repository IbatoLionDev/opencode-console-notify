// Event catalog: every OpenCode event a custom notification can bind
// to. The user never alters events, only picks one. Source: the OpenCode
// V2 session event taxonomy, the first-party notification plugin event
// names, and the observed live server stream; high-frequency families
// carry a noise warning.
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
	{Type: "session.deleted", DescEn: "session removed", DescEs: "se elimina la sesión"},
	{Type: "session.renamed", DescEn: "session renamed", DescEs: "sesión renombrada"},
	{Type: "session.moved", DescEn: "session moved", DescEs: "sesión movida"},
	{Type: "session.agent.selected", DescEn: "session agent changed", DescEs: "cambió el agente de la sesión"},
	{Type: "session.model.selected", DescEn: "session model changed", DescEs: "cambió el modelo de la sesión"},
	{Type: "session.metadata.updated", DescEn: "session metadata changed", DescEs: "cambiaron los metadatos de la sesión"},
	{Type: "session.permissions", DescEn: "session permission rules replaced", DescEs: "reglas de permiso reemplazadas"},
	{Type: "session.viewed", DescEn: "session viewed (idle watermark)", DescEs: "sesión vista (marca de inactividad)"},
	{Type: "session.forked", DescEn: "session forked", DescEs: "sesión bifurcada"},
	{Type: "session.usage.updated", DescEn: "usage totals updated", DescEs: "uso actualizado", Noisy: true},
	{Type: "session.execution.started", DescEn: "a run started", DescEs: "empezó una ejecución"},
	{Type: "session.execution.succeeded", DescEn: "run finished, waiting for input", DescEs: "ejecución terminada, esperando tu entrada"},
	{Type: "session.execution.failed", DescEn: "run failed", DescEs: "la ejecución falló"},
	{Type: "session.execution.interrupted", DescEn: "run interrupted", DescEs: "ejecución interrumpida"},
	{Type: "session.inbox.enqueued", DescEn: "prompt admitted", DescEs: "prompt admitido"},
	{Type: "session.inbox.delivered", DescEn: "prompt delivered", DescEs: "prompt entregado"},
	{Type: "session.inbox.cancelled", DescEn: "prompt cancelled", DescEs: "prompt cancelado"},
	{Type: "session.inbox.delivery.changed", DescEn: "prompt delivery changed", DescEs: "cambió la entrega del prompt"},
	{Type: "session.instructions.updated", DescEn: "instructions updated", DescEs: "instrucciones actualizadas"},
	{Type: "session.synthetic", DescEn: "synthetic message added", DescEs: "mensaje sintético añadido"},
	{Type: "session.skill.activated", DescEn: "a skill activated", DescEs: "skill activado"},
	{Type: "session.shell.started", DescEn: "session shell started", DescEs: "shell de sesión iniciado"},
	{Type: "session.shell.ended", DescEn: "session shell ended", DescEs: "shell de sesión terminado"},
	{Type: "session.step.started", DescEn: "agent step started", DescEs: "paso del agente iniciado", Noisy: true},
	{Type: "session.step.streamed", DescEn: "agent step streamed", DescEs: "paso del agente transmitido", Noisy: true},
	{Type: "session.step.ended", DescEn: "agent step ended", DescEs: "paso del agente terminado", Noisy: true},
	{Type: "session.step.failed", DescEn: "agent step failed", DescEs: "paso del agente falló"},
	{Type: "session.text.started", DescEn: "assistant text started", DescEs: "texto del asistente iniciado"},
	{Type: "session.text.delta", DescEn: "assistant text streamed", DescEs: "texto del asistente transmitido", Noisy: true},
	{Type: "session.text.ended", DescEn: "assistant text completed", DescEs: "texto del asistente completado"},
	{Type: "session.reasoning.started", DescEn: "thinking started", DescEs: "razonamiento iniciado"},
	{Type: "session.reasoning.delta", DescEn: "thinking streamed", DescEs: "razonamiento transmitido", Noisy: true},
	{Type: "session.reasoning.ended", DescEn: "thinking completed", DescEs: "razonamiento completado"},
	{Type: "session.tool.input.started", DescEn: "tool input started", DescEs: "entrada de herramienta iniciada"},
	{Type: "session.tool.input.delta", DescEn: "tool input streamed", DescEs: "entrada de herramienta transmitida", Noisy: true},
	{Type: "session.tool.input.ended", DescEn: "tool input completed", DescEs: "entrada de herramienta completada"},
	{Type: "session.tool.called", DescEn: "tool called", DescEs: "herramienta invocada"},
	{Type: "session.tool.progress", DescEn: "tool progress updated", DescEs: "progreso actualizado", Noisy: true},
	{Type: "session.tool.success", DescEn: "tool succeeded", DescEs: "herramienta exitosa"},
	{Type: "session.tool.failed", DescEn: "tool failed", DescEs: "herramienta falló"},
	{Type: "session.retry.scheduled", DescEn: "retry scheduled", DescEs: "reintento programado"},
	{Type: "session.compaction.started", DescEn: "compaction started", DescEs: "compactación iniciada"},
	{Type: "session.compaction.delta", DescEn: "compaction streamed", DescEs: "compactación transmitida", Noisy: true},
	{Type: "session.compaction.ended", DescEn: "compaction completed", DescEs: "compactación completada"},
	{Type: "session.compaction.failed", DescEn: "compaction failed", DescEs: "compactación falló"},
	{Type: "session.revert.staged", DescEn: "revert staged", DescEs: "reversión preparada"},
	{Type: "session.revert.cleared", DescEn: "staged revert cleared", DescEs: "reversión preparada descartada"},
	{Type: "session.revert.committed", DescEn: "revert committed", DescEs: "reversión aplicada"},
	{Type: "permission.asked", DescEn: "permission needed to continue", DescEs: "permiso necesario para continuar"},
	{Type: "permission.replied", DescEn: "permission request answered", DescEs: "petición de permiso respondida"},
	{Type: "form.created", DescEn: "OpenCode asked a question", DescEs: "OpenCode hizo una pregunta"},
	{Type: "form.replied", DescEn: "question answered", DescEs: "pregunta respondida"},
	{Type: "form.cancelled", DescEn: "question dismissed", DescEs: "pregunta descartada"},
	{Type: "server.connected", DescEn: "connected to the server", DescEs: "conectado al servidor"},
	{Type: "skill.updated", DescEn: "skill registry changed", DescEs: "registro de skills cambió", Noisy: true},
	{Type: "model.updated", DescEn: "model list changed", DescEs: "lista de modelos cambió"},
	{Type: "provider.updated", DescEn: "provider list changed", DescEs: "lista de proveedores cambió"},
	{Type: "project.updated", DescEn: "project state changed", DescEs: "el proyecto cambió"},
	{Type: "agent.updated", DescEn: "agent list changed", DescEs: "lista de agentes cambió"},
	{Type: "reference.updated", DescEn: "a reference changed", DescEs: "una referencia cambió"},
	{Type: "plugin.updated", DescEn: "plugin list changed", DescEs: "lista de plugins cambió"},
	{Type: "command.updated", DescEn: "command list changed", DescEs: "lista de comandos cambió"},
	{Type: "websearch.updated", DescEn: "websearch providers changed", DescEs: "proveedores de búsqueda cambiaron"},
	{Type: "integration.updated", DescEn: "integration list changed", DescEs: "lista de integraciones cambió"},
	{Type: "shell.created", DescEn: "a shell was created", DescEs: "shell creado"},
	{Type: "shell.deleted", DescEn: "a shell was removed", DescEs: "shell eliminado"},
	{Type: "shell.exited", DescEn: "a shell exited", DescEs: "shell terminado"},
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
