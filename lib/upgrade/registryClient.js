// Adapter: npm registry client serving the latest published version.
// Only dependency is node:https; comparison stays in versioning.js.

import { get } from 'node:https';

const PACKAGE_NAME = 'opencode-console-notify';
const REGISTRY_URL = 'https://registry.npmjs.org/opencode-console-notify/latest';
const FETCH_TIMEOUT_MS = 10000;

// Fetch the latest published version from the npm registry.
function fetchLatest(timeoutMs = FETCH_TIMEOUT_MS) {
  return new Promise((resolve, reject) => {
    const req = get(REGISTRY_URL, (res) => {
      const status = res.statusCode || 0;
      if (status < 200 || status >= 300) {
        res.resume();
        reject(new Error(`registry request failed with status ${status}`));
        return;
      }
      let body = '';
      res.setEncoding('utf8');
      res.on('data', (chunk) => {
        body += chunk;
      });
      res.on('end', () => {
        let data;
        try {
          data = JSON.parse(body);
        } catch (err) {
          reject(new Error(`parse registry response: ${err.message}`));
          return;
        }
        if (!data || typeof data.version !== 'string' || data.version === '') {
          reject(new Error('registry response has no version field'));
          return;
        }
        resolve(data.version);
      });
    });
    req.on('error', (err) => {
      reject(new Error(`registry request failed: ${err.message}`));
    });
    req.setTimeout(timeoutMs, () => {
      req.destroy(new Error(`registry request timed out after ${timeoutMs}ms`));
    });
  });
}

export {
  PACKAGE_NAME,
  REGISTRY_URL,
  FETCH_TIMEOUT_MS,
  fetchLatest,
};
