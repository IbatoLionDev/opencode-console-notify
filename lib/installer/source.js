// Adapter: resolves the embedded plugin source path on disk.

import path from 'node:path';
import { fileURLToPath } from 'node:url';
import * as config from '../config/index.js';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

function pluginSourcePath() {
  return path.join(__dirname, '..', '..', 'plugin', config.PLUGIN_FILE_NAME);
}

export { pluginSourcePath };
