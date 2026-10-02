// Pure domain: numeric semver comparison. No I/O.

// Numeric semver comparison: negative when a < b, 0 when equal,
// positive when a > b.
function compareVersions(a, b) {
  const core = (v) => v
    .split('+')[0]
    .split('-')[0]
    .split('.')
    .map((part) => {
      const n = Number.parseInt(part, 10);
      return Number.isNaN(n) ? 0 : n;
    });
  const pa = core(a);
  const pb = core(b);
  const len = Math.max(pa.length, pb.length);
  for (let i = 0; i < len; i++) {
    const x = pa[i] || 0;
    const y = pb[i] || 0;
    if (x !== y) {
      return x - y;
    }
  }
  // A pre-release (e.g. 1.2.0-beta) sorts below its plain release.
  const preA = a.includes('-');
  const preB = b.includes('-');
  if (preA && !preB) {
    return -1;
  }
  if (!preA && preB) {
    return 1;
  }
  return 0;
}

export { compareVersions };
