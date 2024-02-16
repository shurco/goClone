package netutil

import (
	"log"
	"net"
	"net/http"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/shurco/goclone/pkg/fsutil"
)

var Folders = map[string]string{
	"css":  "assets/css",
	"js":   "assets/js",
	"img":  "assets/img",
	"font": "assets/font",
}

var Extensions = map[string]string{
	".css":   "assets/css",
	".js":    "assets/js",
	".jpg":   "assets/img",
	".jpeg":  "assets/img",
	".gif":   "assets/img",
	".png":   "assets/img",
	".svg":   "assets/img",
	".eot":   "assets/font",
	".otf":   "assets/font",
	".ttf":   "assets/font",
	".woff":  "assets/font",
	".woff2": "assets/font",
}

// Extractor visits a link determines if its a page or sublink
// downloads the contents to a correct directory in project folder
func Extractor(link, projectPath string) {
	resp, err := http.Get(link)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	ext := urlExtension(resp.Request.URL.Path)

	if ext != "" {
		dirPath := "/" + Extensions[ext] + "/"
		if dirPath != "" {
			name := ReplaceSlashWithDash(resp.Request.URL.Path)

			file, err := fsutil.OpenFile(filepath.Join(projectPath, dirPath, name), fsutil.FsCWFlags, 0o666)
			if err != nil {
				log.Fatal(err)
			}
			if _, err := fsutil.WriteOSFile(file, resp.Body); err != nil {
				log.Fatal(err)
			}
		}
	}
}

// GetAssetDir is ...
func GetAssetDir(filename string) string {
	dirPath := "/" + Extensions[urlExtension(filename)] + "/"
	if dirPath != "" {
		return dirPath
	}
	return ""
}

func urlExtension(URL string) string {
	ext := path.Ext(URL)
	if len(ext) > 5 {
		match, _ := regexp.MatchString(`^[a-zA-Z0-9]+$`, ext[1:])
		if !match {
			ext = ext[:2]
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

// ReplaceSlashWithDash is ...
func ReplaceSlashWithDash(input string) string {
	if strings.HasPrefix(input, "/") {
		input = input[1:]
	}

	return strings.ReplaceAll(input, "/", "-")
}
