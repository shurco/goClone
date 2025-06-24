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

	if _, err := fsutil.WriteOSFile(index, readCSS(body, body)); err != nil {
		log.Fatal(err)
	}
}

func readCSS(data, body string) string {
	lines := strings.Split(data, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		regExp, err := regexp.Compile(`url\((.*?)\)`)
		if err != nil {
			fmt.Println("Error compiling regex pattern:", err)
		}

		matches := regExp.FindAllStringSubmatch(line, -1)
		for _, match := range matches {
			link := strings.ReplaceAll(match[1], `'`, "")
			link = strings.ReplaceAll(link, `"`, "")
			parsedURL, err := url.Parse(link)
			if err != nil {
				fmt.Println("Error parsing URL:", err)
			}

			if parsedURL.Host == projectURL.Host || parsedURL.Host == "" {
				folder := netutil.GetAssetDir(parsedURL.Path)
				link := domain + strings.ReplaceAll("/"+parsedURL.Path, "//", "/")

				switch folder {
				case netutil.Folders["font"]:
					if !contains(files.font, link) {
						fmt.Println("Font found", "-->", link)
						files.font = append(files.font, link)
						netutil.Extractor(link, projectPath)
					}

				case netutil.Folders["img"]:
					if !contains(files.img, link) {
						fmt.Println("Img found", "-->", link)
						files.img = append(files.img, link)
						netutil.Extractor(link, projectPath)
					}
				}
				body = strings.Replace(body, link, "/"+folder+"/"+filepath.Base(link), -1)
			}
		}
	}
	return body
}
