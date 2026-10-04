// Domain: custom alerts (user text bound to catalog events).
// Validation is catalog-strict; titles and bodies truncate to
// toast-safe length. Mirrors internal/config/customs.go.
import { isKnownEvent } from './events.js';

const CUSTOM_TEXT_LIMIT = 160;

function truncateCustomText(s) {
  const chars = [...String(s ?? '')];
  if (chars.length <= CUSTOM_TEXT_LIMIT) {
    return String(s ?? '');
  }
  return chars.slice(0, CUSTOM_TEXT_LIMIT).join('');
}

function validateCustom(event, title) {
  if (!isKnownEvent(event)) {
    throw new Error(`unknown event "${event}" (see --list-events)`);
  }
  if (String(title ?? '').trim() === '') {
    throw new Error('custom title must not be empty');
  }
}

// normalizeCustoms drops entries that can never fire so one bad entry
// never breaks the file.
function normalizeCustoms(list) {
  return list
    .filter((c) => c && typeof c === 'object')
    .map((c) => ({
      id: String(c.id ?? ''),
      event: String(c.event ?? ''),
      title: String(c.title ?? ''),
      body: String(c.body ?? ''),
      enabled: c.enabled !== false,
    }))
    .filter((c) => {
      try {
        validateCustom(c.event, c.title);
        return true;
      } catch {
        return false;
      }
    });
}

function nextCustomId(settings) {
  let max = 0;
  for (const c of settings.customAlerts || []) {
    const m = /^custom-(\d+)$/.exec(c.id || '');
    if (m && Number(m[1]) > max) {
      max = Number(m[1]);
    }
  }
  return `custom-${max + 1}`;
}

function findCustom(settings, id) {
  return (settings.customAlerts || []).find((c) => c.id === id);
}

function addCustom(settings, event, title, body) {
  validateCustom(event, title);
  const custom = {
    id: nextCustomId(settings),
    event: String(event).trim(),
    title: truncateCustomText(String(title).trim()),
    body: truncateCustomText(String(body ?? '').trim()),
    enabled: true,
  };
  settings.customAlerts = [...(settings.customAlerts || []), custom];
  return custom.id;
}

function updateCustom(settings, id, event, title, body) {
  validateCustom(event, title);
  const custom = findCustom(settings, id);
  if (!custom) {
    return false;
  }
  custom.event = String(event).trim();
  custom.title = truncateCustomText(String(title).trim());
  custom.body = truncateCustomText(String(body ?? '').trim());
  return true;
}

function removeCustom(settings, id) {
  const list = settings.customAlerts || [];
  const idx = list.findIndex((c) => c.id === id);
  if (idx < 0) {
    return false;
  }
  settings.customAlerts = [...list.slice(0, idx), ...list.slice(idx + 1)];
  return true;
}

function toggleCustom(settings, id) {
  const custom = findCustom(settings, id);
  if (!custom) {
    return false;
  }
  custom.enabled = !custom.enabled;
  return true;
}

function customsForEvent(settings, event) {
  return (settings.customAlerts || []).filter((c) => c.enabled && c.event === event);
}

export {
  CUSTOM_TEXT_LIMIT,
  truncateCustomText,
  validateCustom,
  normalizeCustoms,
  nextCustomId,
  findCustom,
  addCustom,
  updateCustom,
  removeCustom,
  toggleCustom,
  customsForEvent,
};
