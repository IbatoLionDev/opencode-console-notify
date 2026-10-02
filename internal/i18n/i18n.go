// Package i18n holds the config TUI and default toast dictionaries.
// English is the fallback; Spanish covers v2.0.0. No dependencies.
package i18n

// Supported languages for the config command.
const (
	LangEN = "en"
	LangES = "es"
)

// Normalize returns lang when supported, English otherwise.
func Normalize(lang string) string {
	if lang == LangES {
		return LangES
	}
	return LangEN
}

var en = map[string]string{
	"app.title":         "OpenCode Notify Config",
	"menu.language":     "Language",
	"menu.update":       "Update",
	"menu.info":         "Info",
	"menu.alerts":       "Alerts",
	"menu.exit":         "Exit",
	"footer.menu":       "Up/Down navigate | Enter open | Q/Esc exit",
	"footer.back":       "Esc/Q back",
	"footer.toggle":     "Up/Down navigate | Space on/off | Esc/Q back",
	"lang.title":        "Language: applies to this TUI and default notification texts",
	"lang.current":      "Current",
	"alerts.title":      "Alerts: Space toggles, changes save immediately",
	"alert.sessionIdle": "Task finished",
	"alert.sessionError": "Something went wrong",
	"alert.permissionAsked": "Permission needed",
	"alert.questionAsked":   "Question asked",
	"trigger.sessionIdle":     "session.idle: task finished, waiting for input (child sessions filtered)",
	"trigger.sessionError":    "session.error: something failed (child sessions filtered)",
	"trigger.permissionAsked": "permission.asked: permission needed to continue (always notifies)",
	"trigger.questionAsked":   "question.asked: OpenCode asked a question (always notifies)",
	"toast.idle":       "Task finished — waiting for your input",
	"toast.error":       "Something went wrong — check the console",
	"toast.permission":  "Permission needed to continue",
	"toast.question":    "OpenCode asked you a question",
	"state.on":          "on",
	"state.off":         "off",
}

var es = map[string]string{
	"app.title":         "Configuración de OpenCode Notify",
	"menu.language":     "Idioma",
	"menu.update":       "Actualizar",
	"menu.info":         "Info",
	"menu.alerts":       "Alertas",
	"menu.exit":         "Salir",
	"footer.menu":       "Arriba/Abajo navegar | Enter abrir | Q/Esc salir",
	"footer.back":       "Esc/Q volver",
	"footer.toggle":     "Arriba/Abajo navegar | Espacio activar/desactivar | Esc/Q volver",
	"lang.title":        "Idioma: aplica a esta TUI y a los textos de notificación por defecto",
	"lang.current":      "Actual",
	"alerts.title":      "Alertas: Espacio cambia, el cambio se guarda al momento",
	"alert.sessionIdle": "Tarea terminada",
	"alert.sessionError": "Algo salió mal",
	"alert.permissionAsked": "Permiso necesario",
	"alert.questionAsked":   "Pregunta recibida",
	"trigger.sessionIdle":     "session.idle: tarea terminada, esperando tu entrada (sesiones hijas filtradas)",
	"trigger.sessionError":    "session.error: algo falló (sesiones hijas filtradas)",
	"trigger.permissionAsked": "permission.asked: permiso necesario para continuar (siempre notifica)",
	"trigger.questionAsked":   "question.asked: OpenCode hizo una pregunta (siempre notifica)",
	"toast.idle":       "Tarea terminada — esperando tu entrada",
	"toast.error":       "Algo salió mal — revisa la consola",
	"toast.permission":  "Permiso necesario para continuar",
	"toast.question":    "OpenCode te hizo una pregunta",
	"state.on":          "activada",
	"state.off":         "desactivada",
}

// T returns the string for key in lang, falling back to English and
// finally to the key itself when missing.
func T(lang, key string) string {
	if v, ok := dict(Normalize(lang))[key]; ok {
		return v
	}
	if v, ok := en[key]; ok {
		return v
	}
	return key
}

func dict(lang string) map[string]string {
	if lang == LangES {
		return es
	}
	return en
}

// Supported reports whether lang is a v2.0.0 language.
func Supported(lang string) bool {
	return lang == LangEN || lang == LangES
}
