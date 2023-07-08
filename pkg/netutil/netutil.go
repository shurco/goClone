package netutil

import (
	"log"
	"net"
	"net/http"
	"path"
	"path/filepath"
	"regexp"

	"github.com/shurco/goclone/pkg/fsutil"
)

var (
	extensionDir = map[string]string{
		".css":  "assets/css",
		".js":   "assets/js",
		".jpg":  "assets/img",
		".jpeg": "assets/img",
		".gif":  "assets/img",
		".png":  "assets/img",
		".svg":  "assets/img",
	}
)

// Extractor visits a link determines if its a page or sublink
// downloads the contents to a correct directory in project folder
func Extractor(link string, projectPath string) {
	resp, err := http.Get(link)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	base := path.Base(resp.Request.URL.Path)
	oldExt := filepath.Ext(base)
	ext := urlExtension(resp.Request.URL.Path)

	if ext != "" {
		dirPath := extensionDir[ext]
		if dirPath != "" {
			name := base[0:len(base)-len(oldExt)] + ext

			file, err := fsutil.OpenFile(filepath.Join(projectPath, dirPath, name), fsutil.FsCWFlags, 0666)
			if err != nil {
				log.Fatal(err)
			}
			if _, err := fsutil.WriteOSFile(file, resp.Body); err != nil {
				log.Fatal(err)
			}
		}
	}
}

func urlExtension(URL string) string {
	ext := path.Ext(URL)
	if len(ext) > 5 {
		match, _ := regexp.MatchString(`^[a-zA-Z]+$`, ext[1:])
		if !match {
			ext = ext[:2] // Assuming you want to keep the first character of the extension
		}
	}
	return ext
}

func IsValidDomain(domain string) bool {
	if len(domain) < 1 || len(domain) > 255 {
		return false
	}

	if match, _ := regexp.MatchString(`^([a-zA-Z0-9-]+\.)+[a-zA-Z]{2,}$`, domain); !match {
		return false
	}

	if _, err := net.LookupHost(domain); err != nil {
		return false
	}

	return true
}
