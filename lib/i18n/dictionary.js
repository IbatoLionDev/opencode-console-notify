// Domain: config TUI + default toast dictionaries. English fallback.

const LANG_EN = 'en';
const LANG_ES = 'es';

function normalizeLang(lang) {
  return lang === LANG_ES ? LANG_ES : LANG_EN;
}

const en = {
  'app.title': 'OpenCode Notify Config',
  'menu.language': 'Language',
  'menu.update': 'Update',
  'menu.info': 'Info',
  'menu.alerts': 'Alerts',
  'menu.customs': 'Customs',
  'menu.exit': 'Exit',
  'footer.menu': 'Up/Down navigate | Enter open | Q/Esc exit',
  'footer.back': 'Esc/Q back',
  'footer.toggle': 'Up/Down navigate | Space on/off | Esc/Q back',
  'lang.title': 'Language: applies to this TUI and default notification texts',
  'lang.current': 'Current',
  'alerts.title': 'Alerts: Space toggles, changes save immediately',
  'alert.sessionIdle': 'Task finished',
  'alert.sessionError': 'Something went wrong',
  'alert.permissionAsked': 'Permission needed',
  'alert.questionAsked': 'Question asked',
  'trigger.sessionIdle': 'session.idle: task finished, waiting for input (child sessions filtered)',
  'trigger.sessionError': 'session.error: something failed (child sessions filtered)',
  'trigger.permissionAsked': 'permission.asked: permission needed to continue (always notifies)',
  'trigger.questionAsked': 'question.asked: OpenCode asked a question (always notifies)',
  'toast.idle': 'Task finished — waiting for your input',
  'toast.error': 'Something went wrong — check the console',
  'toast.permission': 'Permission needed to continue',
  'toast.question': 'OpenCode asked you a question',
  'state.on': 'on',
  'state.off': 'off',
  'custom.tag': 'custom',
  'customs.title': 'Custom notifications',
  'customs.empty': 'No customs yet — press A to add one',
  'customs.footer': 'Up/Down navigate | Space toggle | A add E edit D delete C catalog | Q back',
  'customs.pickEvent': 'Pick an event by number',
  'customs.titleLabel': 'Title',
  'customs.bodyLabel': 'Body',
  'customs.confirmDelete': 'Delete "%s"? (y/n): ',
  'customs.catalogTitle': 'Subscribable events (! = noisy)',
  'customs.optional': 'empty = none',
};

const es = {
  'app.title': 'Configuración de OpenCode Notify',
  'menu.language': 'Idioma',
  'menu.update': 'Actualizar',
  'menu.info': 'Info',
  'menu.alerts': 'Alertas',
  'menu.customs': 'Custom',
  'menu.exit': 'Salir',
  'footer.menu': 'Arriba/Abajo navegar | Enter abrir | Q/Esc salir',
  'footer.back': 'Esc/Q volver',
  'footer.toggle': 'Arriba/Abajo navegar | Espacio activar/desactivar | Esc/Q volver',
  'lang.title': 'Idioma: aplica a esta TUI y a los textos de notificación por defecto',
  'lang.current': 'Actual',
  'alerts.title': 'Alertas: Espacio cambia, el cambio se guarda al momento',
  'alert.sessionIdle': 'Tarea terminada',
  'alert.sessionError': 'Algo salió mal',
  'alert.permissionAsked': 'Permiso necesario',
  'alert.questionAsked': 'Pregunta recibida',
  'trigger.sessionIdle': 'session.idle: tarea terminada, esperando tu entrada (sesiones hijas filtradas)',
  'trigger.sessionError': 'session.error: algo falló (sesiones hijas filtradas)',
  'trigger.permissionAsked': 'permission.asked: permiso necesario para continuar (siempre notifica)',
  'trigger.questionAsked': 'question.asked: OpenCode hizo una pregunta (siempre notifica)',
  'toast.idle': 'Tarea terminada — esperando tu entrada',
  'toast.error': 'Algo salió mal — revisa la consola',
  'toast.permission': 'Permiso necesario para continuar',
  'toast.question': 'OpenCode te hizo una pregunta',
  'state.on': 'activada',
  'state.off': 'desactivada',
  'custom.tag': 'custom',
  'customs.title': 'Notificaciones custom',
  'customs.empty': 'Aún no hay customs — pulsa A para crear una',
  'customs.footer': 'Arriba/Abajo navegar | Espacio activar/desactivar | A añadir E editar D eliminar C catálogo | Q volver',
  'customs.pickEvent': 'Elige un evento por número',
  'customs.titleLabel': 'Título',
  'customs.bodyLabel': 'Texto',
  'customs.confirmDelete': '¿Eliminar "%s"? (s/n): ',
  'customs.catalogTitle': 'Eventos disponibles (! = ruidoso)',
  'customs.optional': 'vacío = ninguno',
};

function t(lang, key) {
  const dict = normalizeLang(lang) === LANG_ES ? es : en;
  if (dict[key] !== undefined) {
    return dict[key];
  }
  if (en[key] !== undefined) {
    return en[key];
  }
  return key;
}

export {
  LANG_EN,
  LANG_ES,
  normalizeLang,
  t,
};
