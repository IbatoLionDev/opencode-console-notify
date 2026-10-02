// Notice is the best-effort update hint: when the npm registry reports
// a newer release, doctor prints one line. All failures are silent by
// design so offline machines never see doctor break.
package doctor

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/IbatoLionDev/opencode-console-notify/internal/version"
)

// registryURL is the npm endpoint serving {"version": "..."} for latest.
var registryURL = "https://registry.npmjs.org/opencode-console-notify/latest"

// registryClient is a blocking client with a short timeout so offline
// machines never hang doctor; tests swap registryURL for httptest servers.
var registryClient = &http.Client{Timeout: 3 * time.Second}

// fetchLatestVersion returns the newest published version, or "" on any
// fetch/parse failure (callers treat "" as "no notice").
func fetchLatestVersion() string {
	resp, err := registryClient.Get(registryURL)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return ""
	}
	var data struct {
		Version string `json:"version"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return ""
	}
	return data.Version
}

// maybePrintUpdateNotice prints a one-line upgrade hint when the registry
// reports a newer version. All failures are silent by design.
func maybePrintUpdateNotice(out io.Writer) {
	latest := fetchLatestVersion()
	if latest == "" {
		return
	}
	if version.Compare(latest, version.Version) > 0 {
		fmt.Fprintf(out, "Update available: %s -> %s — run \"opencode-notify upgrade\".\n", version.Version, latest)
	}
}
