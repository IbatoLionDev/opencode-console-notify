// Line-mode customs panel: numbered list with one-line commands.
// Fullscreen twin lives in customs_panel.js; prompts here read raw lines
// so titles keep their case.
import { t } from '../i18n/dictionary.js';
import { knownEvents } from '../config/events.js';
import {
  addCustom,
  updateCustom,
  removeCustom,
  toggleCustom,
} from '../config/customs.js';
import { saveSettings } from '../config/settings.js';

function customLineDetail(custom) {
  if (!custom.body) {
    return custom.event;
  }
  return `${custom.event} — ${custom.body}`;
}

function renderCustomsLine(settings, lang) {
  const lines = [t(lang, 'customs.title')];
  const list = settings.customAlerts || [];
  if (list.length === 0) {
    lines.push(t(lang, 'customs.empty'));
  }
  list.forEach((c, i) => {
    lines.push(`${i + 1}. ${c.title} [${t(lang, 'custom.tag')}] [${t(lang, c.enabled ? 'state.on' : 'state.off')}]`);
    lines.push(`   ${customLineDetail(c)}`);
  });
  lines.push(t(lang, 'customs.lineFooter'));
  lines.push('');
  return lines.join('\n');
}

function renderCatalogLine(lang) {
  const lines = [t(lang, 'customs.catalogTitle')];
  knownEvents().forEach((e, i) => {
    const desc = lang === 'es' ? e.descEs : e.descEn;
    lines.push(`${i + 1}. ${e.type}: ${desc}${e.noisy ? ' [!]' : ''}`);
  });
  lines.push(t(lang, 'footer.back'));
  lines.push('');
  return lines.join('\n');
}

function customIndex(settings, arg) {
  const n = Number.parseInt(String(arg ?? '').trim(), 10);
  const list = settings.customAlerts || [];
  if (!Number.isInteger(n) || n < 1 || n > list.length) {
    return -1;
  }
  return n - 1;
}

function catalogIndex(event) {
  return knownEvents().findIndex((e) => e.type === event);
}

async function customsLinePickEvent(ctx, askRaw, current) {
  ctx.output.write(renderCatalogLine(ctx.lang));
  let prompt = `${t(ctx.lang, 'customs.pickEvent')}: `;
  if (current) {
    prompt = `${t(ctx.lang, 'customs.pickEvent')} [${current}]: `;
  }
  ctx.output.write(prompt);
  const line = (await askRaw()).trim();
  if (line === '') {
    return current ? catalogIndex(current) : -1;
  }
  const n = Number.parseInt(line, 10);
  const events = knownEvents();
  if (!Number.isInteger(n) || n < 1 || n > events.length) {
    return -1;
  }
  return n - 1;
}

async function customsLineAdd(ctx, askRaw) {
  const idx = await customsLinePickEvent(ctx, askRaw, '');
  if (idx < 0) {
    return 'customs';
  }
  const events = knownEvents();
  ctx.output.write(`${t(ctx.lang, 'customs.titleLabel')}: `);
  const title = await askRaw();
  if (title === '') {
    return 'customs';
  }
  ctx.output.write(`${t(ctx.lang, 'customs.bodyLabel')} (${t(ctx.lang, 'customs.optional')}): `);
  const body = await askRaw();
  try {
    addCustom(ctx.settings, events[idx].type, title, body);
  } catch (err) {
    ctx.output.write(`Error: ${err.message}\n`);
    return 'abort';
  }
  return persistLine(ctx);
}

async function customsLineEdit(ctx, askRaw, idx) {
  const list = ctx.settings.customAlerts || [];
  if (idx < 0 || idx >= list.length) {
    return 'customs';
  }
  const current = list[idx];
  ctx.output.write(`${t(ctx.lang, 'customs.titleLabel')} [${current.title}]: `);
  let title = await askRaw();
  if (title === '') {
    title = current.title;
  }
  ctx.output.write(`${t(ctx.lang, 'customs.bodyLabel')} [${current.body}]: `);
  let body = await askRaw();
  if (body === '' && current.body) {
    body = current.body;
  }
  const eventIdx = await customsLinePickEvent(ctx, askRaw, current.event);
  if (eventIdx < 0) {
    return 'customs';
  }
  const events = knownEvents();
  try {
    const kept = updateCustom(ctx.settings, current.id, events[eventIdx].type, title, body);
    if (!kept) {
      return 'customs';
    }
  } catch (err) {
    ctx.output.write(`Error: ${err.message}\n`);
    return 'abort';
  }
  return persistLine(ctx);
}

async function customsLineDelete(ctx, ask, idx) {
  const list = ctx.settings.customAlerts || [];
  if (idx < 0 || idx >= list.length) {
    return 'customs';
  }
  ctx.output.write(t(ctx.lang, 'customs.confirmDelete').replace('%s', list[idx].title));
  const answer = await ask();
  if (answer !== 'y' && answer !== 's') {
    return 'customs';
  }
  if (!removeCustom(ctx.settings, list[idx].id)) {
    return 'customs';
  }
  return persistLine(ctx);
}

function persistLine(ctx) {
  try {
    saveSettings(ctx.pluginsDir, ctx.settings);
  } catch (err) {
    ctx.output.write(`Error: ${err.message}\n`);
    return 'abort';
  }
  return 'customs';
}

async function runCustomsLineView(ctx, ask, askRaw) {
  for (;;) {
    ctx.output.write(renderCustomsLine(ctx.settings, ctx.lang));
    const next = await dispatchCustomsLine(ctx, ask, askRaw, await ask());
    if (next !== 'customs') {
      return next;
    }
  }
}

// dispatchCustomsLine routes one panel command; 'customs' means stay.
async function dispatchCustomsLine(ctx, ask, askRaw, line) {
  if (isBackLine(line)) {
    return 'menu';
  }
  if (line === 'a') {
    return customsLineAdd(ctx, askRaw);
  }
  if (line === 'c') {
    ctx.output.write(renderCatalogLine(ctx.lang));
    await ask();
    return 'customs';
  }
  if (line.startsWith('e ')) {
    return customsLineEdit(ctx, askRaw, customIndex(ctx.settings, line.slice(2)));
  }
  if (line.startsWith('d ')) {
    return customsLineDelete(ctx, ask, customIndex(ctx.settings, line.slice(2)));
  }
  return toggleCustomLine(ctx, line);
}

function isBackLine(line) {
  return line === 'q' || line === 'esc' || line === 'b' || line === 'back' || line === 'volver';
}

// toggleCustomLine flips the custom at a 1-based position; unknown input
// stays on the panel, save failures abort.
function toggleCustomLine(ctx, line) {
  const idx = customIndex(ctx.settings, line);
  if (idx < 0) {
    return 'customs';
  }
  const list = ctx.settings.customAlerts;
  if (!toggleCustom(ctx.settings, list[idx].id)) {
    return 'customs';
  }
  return persistLine(ctx);
}

export {
  renderCustomsLine,
  renderCatalogLine,
  runCustomsLineView,
};
