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
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/geziyor/geziyor"
	"github.com/geziyor/geziyor/client"

	"github.com/shurco/goClone/pkg/fsutil"
	"github.com/shurco/goClone/pkg/netutil"
)

func CloneSite(ctx context.Context, args []string, flag Flags) error {
	if len(args) < 1 {
		return fmt.Errorf("url is required")
	}

	// configure assets root and UA
	if strings.TrimSpace(flag.AssetsRoot) != "" {
		netutil.SetAssetRoot(flag.AssetsRoot)
	}
	netutil.SetDefaultUserAgent(flag.UserAgent)
	if flag.HTTPTimeoutSeconds > 0 {
		netutil.SetHTTPTimeout(time.Duration(flag.HTTPTimeoutSeconds) * time.Second)
	}
	if flag.MaxDownloadMB > 0 {
		netutil.SetMaxDownloadBytes(int64(flag.MaxDownloadMB) * 1024 * 1024)
	}
	if flag.MaxConcurrentWorkers > 0 {
		SetDownloadConcurrency(flag.MaxConcurrentWorkers)
	}

	// validate URL and initialize globals
	u, err := url.ParseRequestURI(args[0])
	if err != nil || u.Scheme == "" || u.Host == "" {
		return fmt.Errorf("%q is not a valid URL", args[0])
	}
	domain = fmt.Sprintf("%s://%s", u.Scheme, u.Host)
	projectURL = u
	projectPath = filepath.Join(fsutil.Workdir(), projectURL.Host)
	files = filesBase{}

	geziyorOptions := &geziyor.Options{
		AllowedDomains:    []string{projectURL.Host},
		StartURLs:         []string{u.String()},
		ParseFunc:         quotesParse,
		UserAgent:         flag.UserAgent,
		CookiesDisabled:   flag.Cookies,
		RobotsTxtDisabled: flag.Robots,
		LogDisabled:       !flag.Verbose,
	}
	if flag.ProxyString != "" {
		geziyorOptions.ProxyFunc = client.RoundRobinProxy(flag.ProxyString)
	}
	if flag.BrowserEndpoint != "" {
		geziyorOptions.BrowserEndpoint = flag.BrowserEndpoint
		geziyorOptions.StartRequestsFunc = func(g *geziyor.Geziyor) {
			g.GetRendered(u.String(), g.Opt.ParseFunc)
		}
	}

	geziyor.NewGeziyor(geziyorOptions).Start()

	// Auto-hint: if browser rendering was enabled but nothing captured, suggest Linux Docker networking fixes
	if flag.BrowserEndpoint != "" && len(files.pages) == 0 && len(files.css) == 0 && len(files.js) == 0 && len(files.img) == 0 && len(files.font) == 0 {
		fmt.Println("Hint: JS rendering returned no content. If you run Chrome inside Docker on Linux and see ERR_CONNECTION_REFUSED or ERR_NAME_NOT_RESOLVED:")
		fmt.Println(" - Run Chrome with host networking: docker run --net=host --rm -d --name headless chromedp/headless-shell:stable")
		fmt.Println(" - Or serve a URL reachable from the container (use host IP or a shared Docker network)")
		fmt.Println(" - Ensure you pass the full DevTools WS URL from /json/version when needed")
	}

	fmt.Printf("Pages: %v\n", len(files.pages))
	fmt.Printf("CSS files: %v\n", len(files.css))
	fmt.Printf("JS files: %v\n", len(files.js))
	fmt.Printf("Img files: %v\n", len(files.img))
	fmt.Printf("Font files: %v\n", len(files.font))

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

// processSrcset rewrites all URLs inside a srcset-like attribute value and triggers downloads
func processSrcset(attr string, body string, join func(string) string) string {
	if strings.TrimSpace(attr) == "" {
		return body
	}
	parts := strings.Split(attr, ",")
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		// p can be "url [descriptor]"
		sp := strings.SplitN(p, " ", 2)
		origURL := sp[0]
		abs := join(origURL)
		body = saveIMG(abs, origURL, body)
	}
	return body
}

func quotesParse(g *geziyor.Geziyor, r *client.Response) {
	body := string(r.Body)
	urlPath := r.Response.Request.URL.Path
	fmt.Printf("page: %s://%s%s\n", projectURL.Scheme, projectURL.Host, urlPath)

	// CSS files
	r.HTMLDoc.Find("link[rel='stylesheet']").Each(func(i int, s *goquery.Selection) {
		if data, exists := s.Attr("href"); exists {
			abs := r.JoinURL(data)
			parsedURL, err := url.Parse(abs)
			if err != nil {
				fmt.Println("Error parsing URL:", err)
				return
			}
			if parsedURL.Host == projectURL.Host || parsedURL.Host == "" {
				link := abs
				if !contains(files.css, link) {
					fmt.Println("Css found", "-->", link)
					files.css = append(files.css, link)
					go downloadAsset(link, projectPath)
					g.Get(r.JoinURL(link), parseCSS)
				}
				newLink := "/" + netutil.Folders["css"] + "/" + netutil.ReplaceSlashWithDash(parsedURL.Path)
				body = strings.Replace(body, data, newLink, -1)
			}
		}
	})

	// JS files
	r.HTMLDoc.Find("script[src],script[data-rocket-src]").Each(func(i int, s *goquery.Selection) {
		if data, exists := s.Attr("src"); exists {
			body = saveJS(r.JoinURL(data), data, body)
		}
		if data1, exists1 := s.Attr("data-rocket-src"); exists1 {
			body = saveJS(r.JoinURL(data1), data1, body)
		}
	})

	// Preload assets (handle fonts/css/js)
	r.HTMLDoc.Find("link[rel='preload']").Each(func(i int, s *goquery.Selection) {
		data, ok := s.Attr("href")
		if !ok {
			return
		}
		as, _ := s.Attr("as")
		abs := r.JoinURL(data)
		switch as {
		case "font", "image":
			body = saveIMG(abs, data, body)
		case "script":
			body = saveJS(abs, data, body)
		case "style":
			// treat like CSS link
			parsedURL, err := url.Parse(abs)
			if err == nil {
				if !contains(files.css, abs) {
					files.css = append(files.css, abs)
					go downloadAsset(abs, projectPath)
					g.Get(r.JoinURL(abs), parseCSS)
				}
				newLink := "/" + netutil.Folders["css"] + "/" + netutil.ReplaceSlashWithDash(parsedURL.Path)
				body = strings.Replace(body, data, newLink, -1)
			}
		default:
			// fallback to JS
			body = saveJS(abs, data, body)
		}
	})

	// Images and lazy variants
	r.HTMLDoc.Find("img[src],img[data-lazy-src],img[data-lazy-srcset],img[srcset]").Each(func(i int, s *goquery.Selection) {
		if v, ok := s.Attr("data-lazy-srcset"); ok {
			body = processSrcset(v, body, r.JoinURL)
		}
		if v, ok := s.Attr("data-lazy-src"); ok {
			body = saveIMG(r.JoinURL(v), v, body)
		}
		if v, ok := s.Attr("srcset"); ok {
			body = processSrcset(v, body, r.JoinURL)
		}
		if v, ok := s.Attr("src"); ok {
			parsed, err := url.Parse(v)
			if err != nil {
				fmt.Println("Error parsing URL:", err)
				return
			}
			if parsed.Scheme == "data" || parsed.Scheme == "blob" {
				return
			}
			body = saveIMG(r.JoinURL(v), v, body)
		}
	})

	// picture <source srcset>
	r.HTMLDoc.Find("picture source[srcset]").Each(func(i int, s *goquery.Selection) {
		if v, ok := s.Attr("srcset"); ok {
			body = processSrcset(v, body, r.JoinURL)
		}
	})

	// Inline CSS blocks
	r.HTMLDoc.Find("style").Each(func(i int, s *goquery.Selection) {
		data := s.Text()
		body = readCSS(data, body, r.Response.Request.URL)
	})

	// Links to other pages
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

	// Write page to disk
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
