package crawler

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/shurco/goClone/pkg/netutil"
)

func saveIMG(absLink string, original string, body string) string {
	parsedURL, err := url.Parse(absLink)
	if err != nil {
		fmt.Println("Error parsing URL:", err)
		return body
	}
	if parsedURL.Host == projectURL.Host || parsedURL.Host == "" {
		link := absLink

		if !contains(files.img, link) {
			fmt.Println("Img found", "-->", link)
			files.img = append(files.img, link)
			go downloadAsset(link, projectPath)
		}

		newLink := "/" + netutil.Folders["img"] + "/" + netutil.ReplaceSlashWithDash(parsedURL.Path)
		return strings.Replace(body, original, newLink, -1)
	}
	return body
}
