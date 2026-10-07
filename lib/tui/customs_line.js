// Line-mode customs panel: numbered list with one-line commands.
// Fullscreen twin lives in customs_panel.js; prompts here read raw lines
// so titles keep their case. Mutations live in customs_line_edit.js.
import { t } from '../i18n/dictionary.js';
import {
  toggleCustom,
} from '../config/customs.js';
import {
  renderCatalogLine as catalogRender,
  customsLinePickEvent,
  customsLineAdd,
  customsLineEdit,
  customsLineDelete,
  persistLine,
} from './customs_line_edit.js';

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

function customIndex(settings, arg) {
  const n = Number.parseInt(String(arg ?? '').trim(), 10);
  const list = settings.customAlerts || [];
  if (!Number.isInteger(n) || n < 1 || n > list.length) {
    return -1;
  }
  return n - 1;
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
    ctx.output.write(catalogRender(ctx.lang));
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
  runCustomsLineView,
};
export {
  renderCatalogLine,
} from './customs_line_edit.js';
