package crawler

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/shurco/goclone/pkg/netutil"
)

func saveIMG(parsedURL *url.URL, body string) string {
	if parsedURL.Host == projectURL.Host || parsedURL.Host == "" {
		link := domain + strings.ReplaceAll("/"+parsedURL.Path, "//", "/")

		if !contains(files.img, link) {
			fmt.Println("Img found", "-->", link)
			files.img = append(files.img, link)
			go netutil.Extractor(link, projectPath)
		}

		newLink := "/" + netutil.Folders["img"] + "/" + netutil.ReplaceSlashWithDash(parsedURL.Path)
		return strings.Replace(body, link, newLink, -1)
	}
	return body
}
