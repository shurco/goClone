package crawler

import (
	"fmt"
	"log"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/geziyor/geziyor"
	"github.com/geziyor/geziyor/client"
	"github.com/shurco/goClone/pkg/fsutil"
	"github.com/shurco/goClone/pkg/netutil"
)

// cssURLRegexp matches url(...) references inside CSS content.
var cssURLRegexp = regexp.MustCompile(`url\((.*?)\)`)

func parseCSS(_ *geziyor.Geziyor, r *client.Response) {
	body := string(r.Body)
	// Use ReplaceSlashWithDash so the filename matches what quotesParse rewrites in HTML.
	name := netutil.ReplaceSlashWithDash(r.Request.URL.Path)
	cssDir := netutil.Folders["css"]

	index, err := fsutil.OpenFile(filepath.Join(projectPath, cssDir, name), fsutil.FsCWTFlags, 0o666)
	if err != nil {
		log.Printf("parseCSS: open %s: %v", name, err)
		return
	}
	if _, err := fsutil.WriteOSFile(index, readCSS(body, body, r.Request.URL)); err != nil {
		log.Printf("parseCSS: write %s: %v", name, err)
	}
}

func readCSS(data, body string, base *url.URL) string {
	for rawLine := range strings.SplitSeq(data, "\n") {
		line := strings.TrimSpace(rawLine)
		for _, match := range cssURLRegexp.FindAllStringSubmatch(line, -1) {
			original := match[1]
			clean := strings.ReplaceAll(strings.ReplaceAll(original, `'`, ""), `"`, "")
			parsedURL, err := url.Parse(clean)
			if err != nil {
				fmt.Println("Error parsing URL:", err)
				continue
			}

			resolved := base.ResolveReference(parsedURL)
			if resolved.Scheme == "data" || resolved.Scheme == "blob" {
				continue
			}
			if resolved.Host != projectURL.Host && resolved.Host != "" {
				continue
			}

			folder := netutil.GetAssetDir(resolved.Path)
			if folder == "" {
				continue
			}
			link := resolved.String()

			switch folder {
			case netutil.Folders["font"]:
				if addAsset("font", link) {
					fmt.Println("Font found", "-->", link)
					downloadAsset(link, projectPath)
				}
			case netutil.Folders["img"]:
				if addAsset("img", link) {
					fmt.Println("Img found", "-->", link)
					downloadAsset(link, projectPath)
				}
			}

			// Use ReplaceSlashWithDash so the CSS reference matches the filename
			// that Extractor writes (consistent with ReplaceSlashWithDash naming).
			newLink := "/" + strings.TrimPrefix(folder, "/") + "/" + netutil.ReplaceSlashWithDash(resolved.Path)
			body = strings.ReplaceAll(body, original, newLink)
		}
	}
	return body
}
