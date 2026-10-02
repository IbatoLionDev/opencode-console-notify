// Adapter: atomic file write (tmp+rename in the same dir so readers
// never see a half-written file).

import crypto from 'node:crypto';
import fs from 'node:fs';
import path from 'node:path';
import * as config from '../config/index.js';

// Atomic tmp+rename in the same dir: readers never see a half-written file.
function writeFileAtomic(dest, data) {
  const dir = path.dirname(dest);
  const tmpName = path.join(dir, `${config.PLUGIN_FILE_NAME}.${crypto.randomUUID()}.tmp`);
  try {
    fs.writeFileSync(tmpName, data, { mode: 0o644 });
    fs.renameSync(tmpName, dest);
  } catch (err) {
    try {
      fs.unlinkSync(tmpName);
    } catch {
    }
    throw err;
  }
}

export { writeFileAtomic };
