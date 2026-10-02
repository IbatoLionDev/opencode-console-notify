// Adapter: best-effort update notice. Silent on every failure so offline
// machines never see doctor break.

import { fetchLatest, readLocalVersion, compareVersions } from '../upgrade/index.js';

async function maybePrintUpdateNotice(out) {
  try {
    const latest = await fetchLatest(3000);
    const local = readLocalVersion();
    if (compareVersions(latest, local) > 0) {
      out.write(`Update available: ${local} -> ${latest} — run "opencode-console-notify upgrade".\n`);
    }
  } catch {
    // silent: offline must never break doctor
  }
}

export { maybePrintUpdateNotice };
