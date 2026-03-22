package crawler

import (
	"context"
	"net/url"
	"slices"
	"sync"
)

// Flags holds CLI-equivalent options for [CloneSite].
type Flags struct {
	// Open launches the mirrored index in the system default browser after the crawl.
	Open bool
	// Serve starts a static file HTTP server on ServePort and blocks until ctx is cancelled.
	Serve bool
	// ServePort is the TCP port for the optional static server (default in CLI: 8088).
	ServePort int
	// UserAgent is sent with HTTP requests (Geziyor and asset downloads).
	UserAgent string
	// ProxyString is a proxy list/URL accepted by Geziyor's RoundRobinProxy; empty disables.
	ProxyString string
	// Cookies, when true, disables sending cookies (Geziyor CookiesDisabled).
	Cookies bool
	// Robots, when true, disables robots.txt checks (Geziyor RobotsTxtDisabled).
	Robots bool
	// BrowserEndpoint is a Chrome DevTools WebSocket URL for rendered pages; empty uses HTTP only.
	BrowserEndpoint string
	// AssetsRoot is the top-level folder name for css/js/img/font under the mirror (e.g. "assets").
	AssetsRoot string
	// MaxConcurrentWorkers limits parallel asset downloads (semaphore size).
	MaxConcurrentWorkers int
	// MaxDownloadMB is the per-asset size cap in megabytes for downloads.
	MaxDownloadMB int
	// HTTPTimeoutSeconds sets the HTTP client timeout for asset fetches.
	HTTPTimeoutSeconds int
	// Verbose enables Geziyor logging when true.
	Verbose bool
}

var (
	files   filesBase
	filesMu sync.Mutex

	projectURL  *url.URL
	projectPath string
	domain      string

	// crawlCtx is the context from the current CloneSite call; used by download goroutines
	// to respect cancellation. Defaults to Background for safety in tests.
	crawlCtx = context.Background()

	// downloadAsset schedules an asset download asynchronously. It is set by init() in
	// init_download.go and may be replaced by tests. The function is goroutine-safe.
	downloadAsset func(link, projectPath string)
)

type filesBase struct {
	pages []string
	css   []string
	js    []string
	img   []string
	font  []string
}

// addAsset appends link to the slice identified by kind under filesMu and returns true
// when the link was newly added, false if already known.
// Valid kinds: "css", "js", "img", "font", "pages".
func addAsset(kind, link string) bool {
	filesMu.Lock()
	defer filesMu.Unlock()

	var list *[]string
	switch kind {
	case "css":
		list = &files.css
	case "js":
		list = &files.js
	case "img":
		list = &files.img
	case "font":
		list = &files.font
	case "pages":
		list = &files.pages
	default:
		return false
	}

	if slices.Contains(*list, link) {
		return false
	}
	*list = append(*list, link)
	return true
}
