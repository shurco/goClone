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
	"github.com/shurco/goclone/pkg/fsutil"
	"github.com/shurco/goclone/pkg/netutil"
)

func parseCSS(g *geziyor.Geziyor, r *client.Response) {
	body := string(r.Body)
	base := path.Base(r.Request.URL.Path)

	index, err := fsutil.OpenFile(projectPath+"/assets/css/"+base, fsutil.FsCWFlags, 0666)
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
			parsedURL, err := url.Parse(match[1])
			if err != nil {
				fmt.Println("Error parsing URL:", err)
			}

			if parsedURL.Host == projectURL.Host || parsedURL.Host == "" {
				folder := netutil.GetAssetDir(parsedURL.Path)
				switch folder {
				case "assets/font":
					if !files.font.Contains(parsedURL.Path) {
						files.font = append(files.font, parsedURL.Path)
						netutil.Extractor(projectURL.String()+parsedURL.Path, projectPath)
					}

				case "assets/img":
					if !files.img.Contains(parsedURL.Path) {
						files.img = append(files.img, parsedURL.Path)
						netutil.Extractor(projectURL.String()+parsedURL.Path, projectPath)
					}
				}
				body = strings.Replace(body, match[1], "/"+folder+"/"+filepath.Base(match[1]), -1)
			}
		}
	}
	return body
}
