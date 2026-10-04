// Adapter: fullscreen config loop (keyboard navigation, in-place redraw).
// Views live in fullscreen_views.js; line mode (tui.js runTui) stays as
// the fallback for pipes and scripts.

import readline from 'node:readline';
import { loadSettings } from '../config/settings.js';
import { normalizeLang } from '../i18n/dictionary.js';
import {
  enterFrame,
  leaveFrame,
  parseKeyPress,
} from '../screen/screen.js';
import { runTui } from './tui.js';
import { runMenuView, runInfoView } from './fullscreen_views.js';
import { runAlertsView } from './fullscreen_alerts.js';
import { runCustomsView } from './customs_panel.js';
import { runLanguageView } from './fullscreen_language.js';

async function runFullscreen(pluginsDir, input, output, upgrade) {
  let settings;
  try {
    settings = loadSettings(pluginsDir);
  } catch (err) {
    output.write(`Error: ${err.message}\n`);
    return 1;
  }
  const ctx = {
    pluginsDir,
    settings,
    lang: normalizeLang(settings.lang),
    output,
    cols: output.columns || 80,
    rows: output.rows || 24,
  };
  input.setRawMode(true);
  readline.emitKeypressEvents(input);
  output.write(enterFrame());
  const ask = () => new Promise((resolve) => {
    input.once('keypress', (ch, key) => resolve(parseKeyPress(key?.name || '', key?.sequence || ch || '')));
  });
  const menuSel = { value: 0 };
  const alertsSel = { value: 0 };
  const customsSel = { value: 0 };
  let view = 'menu';
  try {
    for (;;) {
      if (view === 'menu') {
        view = await runMenuView(ctx, ask, menuSel);
      } else if (view === 'alerts') {
        view = await runAlertsView(ctx, ask, alertsSel);
      } else if (view === 'customs') {
        view = await runCustomsView(ctx, ask, customsSel);
      } else if (view === 'language') {
        view = await runLanguageView(ctx, ask);
      } else if (view === 'info') {
        view = await runInfoView(ctx, ask);
      } else if (view === 'update') {
        output.write(leaveFrame());
        input.setRawMode(false);
        if (typeof upgrade === 'function') {
          return await upgrade(output);
        }
        return 0;
      } else if (view === 'exit') {
        return 0;
      } else {
        return 1;
      }
    }
  } finally {
    try {
      input.setRawMode(false);
    } catch {
    }
    output.write(leaveFrame());
  }
}

// runSmartTui opens the fullscreen loop on a real console and the line
// mode everywhere else (pipes, scripts, probes).
async function runSmartTui(pluginsDir, input, output, upgrade) {
  if (input?.isTTY && output?.isTTY) {
    return await runFullscreen(pluginsDir, input, output, upgrade);
  }
  return await runTui(pluginsDir, input, output, upgrade);
}

export {
  runFullscreen,
  runSmartTui,
};
export {
  menuTarget,
  menuShortcut,
} from './fullscreen_views.js';
export {
  digitIndex,
  alertField,
} from './fullscreen_alerts.js';
