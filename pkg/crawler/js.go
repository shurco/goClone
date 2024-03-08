package crawler

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/shurco/goClone/pkg/netutil"
)

func saveJS(parsedURL *url.URL, body string) string {
	if parsedURL.Host == projectURL.Host || parsedURL.Host == "" {
		link := domain + strings.ReplaceAll("/"+parsedURL.Path, "//", "/")

		if !contains(files.js, link) {
			fmt.Println("Js found", "-->", link)
			files.js = append(files.js, link)
			go netutil.Extractor(link, projectPath)
		}

		newLink := "/" + netutil.Folders["js"] + "/" + netutil.ReplaceSlashWithDash(parsedURL.Path)
		return strings.Replace(body, link, newLink, -1)
	}
	return body
}
