package crawler

import (
	"fmt"
	"log"
	"net/url"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/geziyor/geziyor"
	"github.com/geziyor/geziyor/client"
	"github.com/shurco/goClone/pkg/fsutil"
	"github.com/shurco/goClone/pkg/netutil"
)

func parseCSS(g *geziyor.Geziyor, r *client.Response) {
	body := string(r.Body)
	base := path.Base(r.Request.URL.Path)

	index, err := fsutil.OpenFile(filepath.Join(projectPath, "assets/css", base), fsutil.FsCWFlags, 0o666)
	if err != nil {
		log.Fatal(err)
	}

	if _, err := fsutil.WriteOSFile(index, readCSS(body, body, r.Request.URL)); err != nil {
		log.Fatal(err)
	}
}

func readCSS(data, body string, base *url.URL) string {
	lines := strings.Split(data, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		regExp, err := regexp.Compile(`url\((.*?)\)`)
		if err != nil {
			fmt.Println("Error compiling regex pattern:", err)
		}

		matches := regExp.FindAllStringSubmatch(line, -1)
		for _, match := range matches {
			original := match[1]
			clean := strings.ReplaceAll(strings.ReplaceAll(original, `'`, ""), `"`, "")
			parsedURL, err := url.Parse(clean)
			if err != nil {
				fmt.Println("Error parsing URL:", err)
				continue
			}

			// resolve relative to base (CSS file or page URL)
			resolved := base.ResolveReference(parsedURL)
			if resolved.Scheme == "data" || resolved.Scheme == "blob" {
				continue
			}

			if resolved.Host == projectURL.Host || resolved.Host == "" {
				folder := netutil.GetAssetDir(resolved.Path)
				if folder == "" {
					continue
				}
				link := resolved.String()

				switch folder {
				case netutil.Folders["font"], "assets/font":
					if !contains(files.font, link) {
						fmt.Println("Font found", "-->", link)
						files.font = append(files.font, link)
						downloadAsset(link, projectPath)
					}
				case netutil.Folders["img"], "assets/img":
					if !contains(files.img, link) {
						fmt.Println("Img found", "-->", link)
						files.img = append(files.img, link)
						downloadAsset(link, projectPath)
					}
				}
				newLink := "/" + strings.TrimPrefix(folder, "/") + "/" + filepath.Base(resolved.Path)
				body = strings.Replace(body, original, newLink, -1)
			}
		}
	}
	return body
}
