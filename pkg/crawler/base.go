package crawler

import "net/url"

type Flags struct {
	Open            bool
	Serve           bool
	ServePort       int
	UserAgent       string
	ProxyString     string
	Cookies         bool
	Robots          bool
	BrowserEndpoint string
}

var (
	files filesBase

	projectURL  *url.URL
	projectPath string
	domain      string
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
