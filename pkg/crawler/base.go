package crawler

import "net/url"

type Flags struct {
	Open                 bool
	Serve                bool
	ServePort            int
	UserAgent            string
	ProxyString          string
	Cookies              bool
	Robots               bool
	BrowserEndpoint      string
	AssetsRoot           string
	MaxConcurrentWorkers int
	MaxDownloadMB        int
	HTTPTimeoutSeconds   int
	Verbose              bool
}

var (
	files filesBase

	projectURL  *url.URL
	projectPath string
	domain      string

	// downloadAsset allows injecting a stub in tests instead of real network I/O
	downloadAsset = func(link, projectPath string) { // default delegates to netutil.Extractor
		// replaced at init in crawler package to avoid import cycle here
	}
)

type filesBase struct {
	pages []string
	css   []string
	js    []string
	img   []string
	font  []string
}

func contains(slice []string, sub string) bool {
	for _, s := range slice {
		if s == sub {
			return true
		}
	}
	return false
}
