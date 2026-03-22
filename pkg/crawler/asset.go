package crawler

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/shurco/goClone/pkg/netutil"
)

// saveAsset records a newly discovered asset of the given kind ("js", "img", "font"),
// schedules its download, and rewrites original in body to the local mirror path.
// If the link belongs to a different host it is left unchanged.
func saveAsset(kind, absLink, original, body string) string {
	parsedURL, err := url.Parse(absLink)
	if err != nil {
		fmt.Println("Error parsing URL:", err)
		return body
	}
	if parsedURL.Host != projectURL.Host && parsedURL.Host != "" {
		return body
	}

	if addAsset(kind, absLink) {
		fmt.Printf("%s found --> %s\n", kind, absLink)
		downloadAsset(absLink, projectPath)
	}

	newLink := "/" + netutil.Folders[kind] + "/" + netutil.ReplaceSlashWithDash(parsedURL.Path)
	return strings.ReplaceAll(body, original, newLink)
}
