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

	"github.com/shurco/goClone/pkg/fsutil"
	"github.com/shurco/goClone/pkg/netutil"
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
	if flag.BrowserEndpoint != "" {
		geziyorOptions.BrowserEndpoint = flag.BrowserEndpoint
		geziyorOptions.StartRequestsFunc = func(g *geziyor.Geziyor) {
			g.GetRendered(domain, g.Opt.ParseFunc)
		}
	}

	geziyor.NewGeziyor(geziyorOptions).Start()

	fmt.Printf("Pages: %v\n", len(files.pages))
	fmt.Printf("CSS files: %v\n", len(files.css))
	fmt.Printf("JS files: %v\n", len(files.js))
	fmt.Printf("Img files: %v\n", len(files.img))
	fmt.Printf("Font files: %v\n", len(files.font))

	projectPath = filepath.Join(fsutil.Workdir(), projectURL.Host)

	if flag.Open {
		url := projectPath + "/index.html"
		if flag.Serve {
			url = fmt.Sprintf("http://localhost:%d/index.html", flag.ServePort)
		}
		cmd := open(url)
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
	urlPath := r.Response.Request.URL.Path
	fmt.Printf("page: %s://%s%s\n", projectURL.Scheme, projectURL.Host, urlPath)

	// search for all link tags that have a rel attribute that is equal to stylesheet - CSS
	r.HTMLDoc.Find("link[rel='stylesheet']").Each(func(i int, s *goquery.Selection) {
		data, exists := s.Attr("href")
		if exists {
			parsedURL, err := url.Parse(data)
			if err != nil {
				fmt.Println("Error parsing URL:", err)
				return
			}

			if parsedURL.Host == projectURL.Host || parsedURL.Host == "" {
				link := domain + strings.ReplaceAll("/"+parsedURL.Path, "//", "/")

				if !contains(files.css, link) {
					fmt.Println("Css found", "-->", link)
					files.css = append(files.css, link)
					go netutil.Extractor(link, projectPath)
					g.Get(r.JoinURL(link), parseCSS)
				}

				newLink := netutil.Folders["css"] + "/" + netutil.ReplaceSlashWithDash(parsedURL.Path)
				body = strings.Replace(body, data, newLink, -1)
			}
		}
	})

	// search for all script tags with src attribute -- JS
	r.HTMLDoc.Find("script[src],script[data-rocket-src]").Each(func(i int, s *goquery.Selection) {
		data, exists := s.Attr("src")
		if exists {
			parsedURL, err := url.Parse(data)
			if err != nil {
				fmt.Println("Error parsing URL:", err)
				return
			}
			body = saveJS(parsedURL, body)

			/*
				if parsedURL.Host == projectURL.Host || parsedURL.Host == "" {
					link := domain + strings.ReplaceAll("/"+parsedURL.Path, "//", "/")

					if !files.js.Contains(link) {
						fmt.Println("Js found", "-->", link)
						files.js = append(files.js, link)
						go netutil.Extractor(link, projectPath)
					}

					newLink := "/" + netutil.Folders["js"] + "/" + netutil.ReplaceSlashWithDash(parsedURL.Path)
					body = strings.Replace(body, data, newLink, -1)
				}
			*/
		}

		data1, exists1 := s.Attr("data-rocket-src")
		if exists1 {
			parsedURL, err := url.Parse(data1)
			if err != nil {
				fmt.Println("Error parsing URL:", err)
				return
			}
			body = saveJS(parsedURL, body)

			/*
				if parsedURL.Host == projectURL.Host || parsedURL.Host == "" {
					link := domain + strings.ReplaceAll("/"+parsedURL.Path, "//", "/")

					if !files.js.Contains(link) {
						fmt.Println("Js found", "-->", link)
						files.js = append(files.js, link)
						go netutil.Extractor(link, projectPath)
					}

					newLink := "/" + netutil.Folders["js"] + "/" + netutil.ReplaceSlashWithDash(parsedURL.Path)
					body = strings.Replace(body, data1, newLink, -1)
				}
			*/
		}
	})

	r.HTMLDoc.Find("link[rel='preload']").Each(func(i int, s *goquery.Selection) {
		data, exists := s.Attr("href")
		if exists {
			parsedURL, err := url.Parse(data)
			if err != nil {
				fmt.Println("Error parsing URL:", err)
				return
			}
			body = saveJS(parsedURL, body)

			/*
				if parsedURL.Host == projectURL.Host || parsedURL.Host == "" {
					link := domain + strings.ReplaceAll("/"+parsedURL.Path, "//", "/")

					if !files.js.Contains(link) {
						fmt.Println("Js found", "-->", link)
						files.js = append(files.js, link)
						go netutil.Extractor(link, projectPath)
					}

					newLink := "/" + netutil.Folders["js"] + "/" + netutil.ReplaceSlashWithDash(parsedURL.Path)
					body = strings.Replace(body, data, newLink, -1)
				}
			*/
		}
	})

	// search for all img tags with src attribute -- Images
	r.HTMLDoc.Find("img[src],img[data-lazy-src],img[data-lazy-srcset]").Each(func(i int, s *goquery.Selection) {
		data1, exists1 := s.Attr("data-lazy-srcset")
		if exists1 {
			urls := strings.Split(data1, ", ")
			for _, line := range urls {
				linkTmp := strings.Split(line, " ")
				parsedURL, err := url.Parse(linkTmp[0])
				if err != nil {
					fmt.Println("Error parsing URL:", err)
					return
				}
				body = saveIMG(parsedURL, body)

				/*
					if parsedURL.Host == projectURL.Host || parsedURL.Host == "" {
						link := domain + strings.ReplaceAll("/"+parsedURL.Path, "//", "/")

						if !files.img.Contains(link) {
							fmt.Println("Img found", "-->", link)
							files.img = append(files.img, link)
							go netutil.Extractor(link, projectPath)
						}

						newLink := "/" + netutil.Folders["img"] + "/" + netutil.ReplaceSlashWithDash(parsedURL.Path)
						body = strings.Replace(body, linkTmp[0], newLink, -1)
					}
				*/
			}
		}

		data2, exists2 := s.Attr("data-lazy-src")
		if exists2 {
			urls := strings.Split(data2, ", ")
			for _, line := range urls {
				linkTmp := strings.Split(line, " ")
				parsedURL, err := url.Parse(linkTmp[0])
				if err != nil {
					fmt.Println("Error parsing URL:", err)
					return
				}
				body = saveIMG(parsedURL, body)

				/*
					if parsedURL.Host == projectURL.Host || parsedURL.Host == "" {
						link := domain + strings.ReplaceAll("/"+parsedURL.Path, "//", "/")

						if !files.img.Contains(link) {
							fmt.Println("Img found", "-->", link)
							files.img = append(files.img, link)
							go netutil.Extractor(link, projectPath)
						}

						newLink := "/" + netutil.Folders["img"] + "/" + netutil.ReplaceSlashWithDash(parsedURL.Path)
						body = strings.Replace(body, linkTmp[0], newLink, -1)
					}
				*/
			}
		}

		data, exists := s.Attr("src")
		if exists {
			parsedURL, err := url.Parse(data)
			if err != nil {
				fmt.Println("Error parsing URL:", err)
				return
			}
			if parsedURL.Scheme == "data" || parsedURL.Scheme == "blob" {
				return
			}
			body = saveIMG(parsedURL, body)

			/*
				if parsedURL.Host == projectURL.Host || parsedURL.Host == "" {
					link := domain + strings.ReplaceAll("/"+parsedURL.Path, "//", "/")

					if !files.img.Contains(link) {
						fmt.Println("Img found", "-->", link)
						files.img = append(files.img, link)
						go netutil.Extractor(link, projectPath)
					}

					newLink := "/assets/img/" + netutil.ReplaceSlashWithDash(parsedURL.Path)
					body = strings.Replace(body, data, newLink, -1)
				}
			*/
		}
	})

	// search for all src in css in code
	r.HTMLDoc.Find("style").Each(func(i int, s *goquery.Selection) {
		data := s.Text()
		body = readCSS(data, body)
	})

	// search all links pages
	r.HTMLDoc.Find("a").Each(func(i int, s *goquery.Selection) {
		data, exists := s.Attr("href")
		if exists {
			parsedURL, err := url.Parse(data)
			if err != nil {
				fmt.Println("Error parsing URL:", err)
				return
			}

			if (parsedURL.Host == projectURL.Host || parsedURL.Host == "") && parsedURL.Path != "/" {
				if !contains(files.pages, parsedURL.Path) {
					files.pages = append(files.pages, parsedURL.Path)
				}
			}
		}
	})

	if urlPath == "" && !contains(files.pages, urlPath) {
		files.pages = append(files.pages, urlPath)
	}

	// Fix the file path issue - handle empty urlPath properly
	var filePath string
	cleanURLPath := strings.TrimPrefix(urlPath, "/")
	if cleanURLPath == "" {
		filePath = filepath.Join(projectPath, "index.html")
	} else {
		filePath = filepath.Join(projectPath, cleanURLPath, "index.html")
	}

	index, err := fsutil.OpenFile(filePath, fsutil.FsCWFlags, 0o666)
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
