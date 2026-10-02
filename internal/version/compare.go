// Compare holds version comparison shared by the cli and doctor
// packages: numeric segments, pre-release sorts below its plain release.
package version

// coreSegments strips build metadata and pre-release suffixes, then parses
// the numeric dot-separated segments (non-numeric parts count as 0).
func coreSegments(v string) []int {
	core := v
	if i := indexByte(core, '+'); i >= 0 {
		core = core[:i]
	}
	if i := indexByte(core, '-'); i >= 0 {
		core = core[:i]
	}
	parts := splitDots(core)
	out := make([]int, 0, len(parts))
	for _, p := range parts {
		out = append(out, parseNum(p))
	}
	return out
}

func indexByte(s string, c byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == c {
			return i
		}
	}
	return -1
}

func splitDots(s string) []string {
	var parts []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '.' {
			parts = append(parts, s[start:i])
			start = i + 1
		}
	}
	parts = append(parts, s[start:])
	return parts
}

func parseNum(s string) int {
	n := 0
	found := false
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			if !found {
				return 0
			}
			break
		}
		found = true
		n = n*10 + int(s[i]-'0')
	}
	if !found {
		return 0
	}
	return n
}

func hasPreRelease(v string) bool {
	core := v
	if i := indexByte(core, '+'); i >= 0 {
		core = core[:i]
	}
	return indexByte(core, '-') >= 0
}

// Compare returns a negative number when a < b, 0 when equal, and a
// positive number when a > b. A pre-release sorts below its plain release.
func Compare(a, b string) int {
	pa := coreSegments(a)
	pb := coreSegments(b)
	maxLen := len(pa)
	if len(pb) > maxLen {
		maxLen = len(pb)
	}
	for i := 0; i < maxLen; i++ {
		x, y := 0, 0
		if i < len(pa) {
			x = pa[i]
		}
		if i < len(pb) {
			y = pb[i]
		}
		if x != y {
			return x - y
		}
	}
	preA := hasPreRelease(a)
	preB := hasPreRelease(b)
	switch {
	case preA && !preB:
		return -1
	case !preA && preB:
		return 1
	default:
		return 0
	}
}
