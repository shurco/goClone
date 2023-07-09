package crawler

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/geziyor/geziyor"
	"github.com/geziyor/geziyor/client"

	"github.com/shurco/goclone/pkg/arrutil"
	"github.com/shurco/goclone/pkg/fsutil"
	"github.com/shurco/goclone/pkg/netutil"
)

type Flags struct {
	Open        bool
	Serve       bool
	ServePort   int
	UserAgent   string
	ProxyString string
	Cookies     bool
	Robots      bool
}

type filesBase struct {
	pages arrutil.Strings
	css   arrutil.Strings
	js    arrutil.Strings
	img   arrutil.Strings
	font  arrutil.Strings
}

var (
	files filesBase

	projectURL  *url.URL
	projectPath string
	domain      string
)

func CloneSite(ctx context.Context, args []string, flag Flags) error {
	if netutil.IsValidDomain(args[0]) {
		return fmt.Errorf("%q is not valid", args[0])
	}
	domain = args[0]

	var err error
	projectURL, err = url.Parse(domain)
	if err != nil {
		return err
	}

	projectPath = filepath.Join(fsutil.Workdir(), projectURL.Host)

	geziyorOptions := &geziyor.Options{
		AllowedDomains:    []string{projectURL.Host},
		StartURLs:         []string{domain},
		ParseFunc:         quotesParse,
		UserAgent:         flag.UserAgent,
		CookiesDisabled:   flag.Cookies,
		RobotsTxtDisabled: flag.Robots,
		LogDisabled:       true,
	}
	if flag.ProxyString != "" {
		geziyorOptions.ProxyFunc = client.RoundRobinProxy(flag.ProxyString)
	}

	geziyor.NewGeziyor(geziyorOptions).Start()

	fmt.Printf("Pages: %v\n", files.pages.Length())
	fmt.Printf("CSS files: %v\n", files.css.Length())
	fmt.Printf("JS files: %v\n", files.js.Length())
	fmt.Printf("Img files: %v\n", files.img.Length())
	fmt.Printf("Font files: %v\n", files.font.Length())

	if flag.Open {
		cmd := open(projectPath + "/index.html")
		if err := cmd.Start(); err != nil {
			return fmt.Errorf("%v: %w", cmd.Args, err)
		}
	}

	if flag.Serve {
		http.Handle("/", http.FileServer(http.Dir(projectPath)))
		log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", flag.ServePort), nil))
	}

	return nil
}

func quotesParse(g *geziyor.Geziyor, r *client.Response) {
	files = filesBase{}

	body := string(r.Body)
	fmt.Printf("page: %s://%s%s\n", projectURL.Scheme, projectURL.Host, r.Response.Request.URL.Path)

	// search for all link tags that have a rel attribute that is equal to stylesheet - CSS
	r.HTMLDoc.Find("link[rel='stylesheet']").Each(func(i int, s *goquery.Selection) {
		data, exists := s.Attr("href")
		if exists {
			parsedURL, err := url.Parse(data)
			if err != nil {
				fmt.Println("Error parsing URL:", err)
			}

			if parsedURL.Host == projectURL.Host || parsedURL.Host == "" {
				fmt.Println("Css found", "-->", parsedURL)
				if !files.css.Contains(parsedURL.Path) {
					files.css = append(files.css, parsedURL.Path)
					netutil.Extractor(projectURL.String()+parsedURL.Path, projectPath)

					g.Get(r.JoinURL(projectURL.String()+parsedURL.Path), parseCSS)
				}

				body = strings.Replace(body, data, "/assets/css/"+filepath.Base(data), -1)
			}
		}
	})

	// search for all script tags with src attribute -- JS
	r.HTMLDoc.Find("script[src]").Each(func(i int, s *goquery.Selection) {
		data, exists := s.Attr("src")
		if exists {
			parsedURL, err := url.Parse(data)
			if err != nil {
				fmt.Println("Error parsing URL:", err)
			}

			if parsedURL.Host == projectURL.Host || parsedURL.Host == "" {
				fmt.Println("Js found", "-->", parsedURL)
				if !files.js.Contains(parsedURL.Path) {
					files.js = append(files.js, parsedURL.Path)
					netutil.Extractor(projectURL.String()+parsedURL.Path, projectPath)
				}

				body = strings.Replace(body, data, "/assets/js/"+filepath.Base(data), -1)
			}
		}
	})

	r.HTMLDoc.Find("link[rel='preload']").Each(func(i int, s *goquery.Selection) {
		data, exists := s.Attr("href")
		if exists {
			parsedURL, err := url.Parse(data)
			if err != nil {
				fmt.Println("Error parsing URL:", err)
			}

			if parsedURL.Host == projectURL.Host || parsedURL.Host == "" {
				fmt.Println("Js found", "-->", parsedURL)
				if !files.js.Contains(parsedURL.Path) {
					files.js = append(files.js, parsedURL.Path)
					netutil.Extractor(projectURL.String()+parsedURL.Path, projectPath)
				}

				body = strings.Replace(body, data, "/assets/js/"+filepath.Base(data), -1)
			}
		}
	})

	// search for all img tags with src attribute -- Images
	r.HTMLDoc.Find("img[src]").Each(func(i int, s *goquery.Selection) {
		data, exists := s.Attr("src")
		if exists {
			parsedURL, err := url.Parse(data)
			if err != nil {
				fmt.Println("Error parsing URL:", err)
			}
			if strings.HasPrefix(projectURL.String()+parsedURL.Path, "data:image") || strings.HasPrefix(projectURL.String()+parsedURL.Path, "blob:") {
				return
			}

			if parsedURL.Host == projectURL.Host || parsedURL.Host == "" {
				fmt.Println("Img found", "-->", parsedURL)
				if !files.img.Contains(parsedURL.Path) {
					files.img = append(files.img, parsedURL.Path)
					netutil.Extractor(projectURL.String()+parsedURL.Path, projectPath)
				}

				body = strings.Replace(body, data, "/assets/img/"+filepath.Base(data), -1)
			}
		}
	})

	// search for all src in css in code
	r.HTMLDoc.Find("style").Each(func(i int, s *goquery.Selection) {
		data := s.Text()
		body = readCSS(data, body)
	})

	r.HTMLDoc.Find("a").Each(func(i int, s *goquery.Selection) {
		data, exists := s.Attr("href")
		if exists {
			parsedURL, err := url.Parse(data)
			if err != nil {
				fmt.Println("Error parsing URL:", err)
			}

			if (parsedURL.Host == projectURL.Host || parsedURL.Host == "") && parsedURL.Path != "/" {
				if !files.pages.Contains(parsedURL.Path) {
					files.pages = append(files.pages, parsedURL.Path)
				}
			}
		}
	})

	index, err := fsutil.OpenFile(projectPath+r.Response.Request.URL.Path+"/index.html", fsutil.FsCWFlags, 0666)
	if err != nil {
		log.Fatal(err)
	}

	body = strings.Replace(body, domain, "", -1)
	if _, err := fsutil.WriteOSFile(index, body); err != nil {
		log.Fatal(err)
	}

	for _, href := range files.pages {
		g.Get(r.JoinURL(href), quotesParse)
	}
}

// open opens the specified URL in the default browser of the user.
func open(url string) *exec.Cmd {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}
	case "darwin":
		cmd = "open"
	default: // "linux", "freebsd", "openbsd", "netbsd"
		cmd = "xdg-open"
	}
	args = append(args, url)
	return exec.Command(cmd, args...)
}
