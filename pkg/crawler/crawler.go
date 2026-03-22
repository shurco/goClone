// Package crawler mirrors websites for offline use: it crawls HTML pages, downloads
// linked CSS, JavaScript, images, and fonts, rewrites URLs to local paths, and writes
// a directory tree suitable for opening in a browser.
package crawler

import (
	"context"
	"errors"
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

// CloneSite downloads the site at args[0] (a valid absolute http or https URL) into
// ./<host>/ relative to the process working directory. flag configures proxy, User-Agent,
// cookies, robots.txt handling, optional headless Chrome rendering, asset root, download
// limits, logging, and whether to open the mirror or serve it over HTTP.
//
// The crawl itself does not honour ctx cancellation (Geziyor limitation); ctx is used
// for asset downloads and, when Serve is true, for graceful HTTP server shutdown.
func CloneSite(ctx context.Context, args []string, flag Flags) error {
	if len(args) < 1 {
		return fmt.Errorf("url is required")
	}

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

	u, err := url.ParseRequestURI(args[0])
	if err != nil || u.Scheme == "" || u.Host == "" {
		return fmt.Errorf("%q is not a valid URL", args[0])
	}
	crawlCtx = ctx
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

	// Block until every background asset download goroutine has finished.
	downloadWg.Wait()

	// Auto-hint: if browser rendering was enabled but nothing was captured.
	if flag.BrowserEndpoint != "" {
		filesMu.Lock()
		empty := len(files.pages) == 0 && len(files.css) == 0 && len(files.js) == 0 && len(files.img) == 0 && len(files.font) == 0
		filesMu.Unlock()
		if empty {
			fmt.Println("Hint: JS rendering returned no content. If you run Chrome inside Docker on Linux and see ERR_CONNECTION_REFUSED or ERR_NAME_NOT_RESOLVED:")
			fmt.Println(" - Run Chrome with host networking: docker run --net=host --rm -d --name headless chromedp/headless-shell:stable")
			fmt.Println(" - Or serve a URL reachable from the container (use host IP or a shared Docker network)")
			fmt.Println(" - Ensure you pass the full DevTools WS URL from /json/version when needed")
		}
	}

	filesMu.Lock()
	fmt.Printf("Pages: %v\n", len(files.pages))
	fmt.Printf("CSS files: %v\n", len(files.css))
	fmt.Printf("JS files: %v\n", len(files.js))
	fmt.Printf("Img files: %v\n", len(files.img))
	fmt.Printf("Font files: %v\n", len(files.font))
	filesMu.Unlock()

	if flag.Open {
		openURL := projectPath + "/index.html"
		if flag.Serve {
			openURL = fmt.Sprintf("http://localhost:%d/index.html", flag.ServePort)
		}
		cmd := open(openURL)
		if err := cmd.Start(); err != nil {
			return fmt.Errorf("%v: %w", cmd.Args, err)
		}
	}

	if flag.Serve {
		srv := &http.Server{
			Addr:    fmt.Sprintf(":%d", flag.ServePort),
			Handler: http.FileServer(http.Dir(projectPath)),
		}
		go func() {
			<-ctx.Done()
			_ = srv.Shutdown(context.Background())
		}()
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("serve: %w", err)
		}
	}

	return nil
}

// processSrcset rewrites all URLs inside a srcset-like attribute value and triggers downloads.
func processSrcset(attr string, body string, join func(string) string) string {
	if strings.TrimSpace(attr) == "" {
		return body
	}
	for part := range strings.SplitSeq(attr, ",") {
		p := strings.TrimSpace(part)
		if p == "" {
			continue
		}
		origURL, _, _ := strings.Cut(p, " ")
		abs := join(origURL)
		body = saveAsset("img", abs, origURL, body)
	}
	return body
}

func quotesParse(g *geziyor.Geziyor, r *client.Response) {
	body := string(r.Body)
	urlPath := r.Response.Request.URL.Path
	fmt.Printf("page: %s://%s%s\n", projectURL.Scheme, projectURL.Host, urlPath)

	// CSS files
	r.HTMLDoc.Find("link[rel='stylesheet']").Each(func(i int, s *goquery.Selection) {
		data, exists := s.Attr("href")
		if !exists {
			return
		}
		abs := r.JoinURL(data)
		parsedURL, err := url.Parse(abs)
		if err != nil {
			fmt.Println("Error parsing URL:", err)
			return
		}
		if parsedURL.Host == projectURL.Host || parsedURL.Host == "" {
			if addAsset("css", abs) {
				fmt.Println("Css found", "-->", abs)
				g.Get(r.JoinURL(abs), parseCSS)
			}
			newLink := "/" + netutil.Folders["css"] + "/" + netutil.ReplaceSlashWithDash(parsedURL.Path)
			body = strings.ReplaceAll(body, data, newLink)
		}
	})

	// JS files
	r.HTMLDoc.Find("script[src],script[data-rocket-src]").Each(func(i int, s *goquery.Selection) {
		if data, exists := s.Attr("src"); exists {
			body = saveAsset("js", r.JoinURL(data), data, body)
		}
		if data1, exists1 := s.Attr("data-rocket-src"); exists1 {
			body = saveAsset("js", r.JoinURL(data1), data1, body)
		}
	})

	// Preload assets (fonts, images, scripts, styles)
	r.HTMLDoc.Find("link[rel='preload']").Each(func(i int, s *goquery.Selection) {
		data, ok := s.Attr("href")
		if !ok {
			return
		}
		as, _ := s.Attr("as")
		abs := r.JoinURL(data)
		switch as {
		case "font":
			body = saveAsset("font", abs, data, body)
		case "image":
			body = saveAsset("img", abs, data, body)
		case "script":
			body = saveAsset("js", abs, data, body)
		case "style":
			parsedURL, err := url.Parse(abs)
			if err == nil && (parsedURL.Host == projectURL.Host || parsedURL.Host == "") {
				if addAsset("css", abs) {
					g.Get(r.JoinURL(abs), parseCSS)
				}
				newLink := "/" + netutil.Folders["css"] + "/" + netutil.ReplaceSlashWithDash(parsedURL.Path)
				body = strings.ReplaceAll(body, data, newLink)
			}
		default:
			body = saveAsset("js", abs, data, body)
		}
	})

	// Images and lazy variants
	r.HTMLDoc.Find("img[src],img[data-lazy-src],img[data-lazy-srcset],img[srcset]").Each(func(i int, s *goquery.Selection) {
		if v, ok := s.Attr("data-lazy-srcset"); ok {
			body = processSrcset(v, body, r.JoinURL)
		}
		if v, ok := s.Attr("data-lazy-src"); ok {
			body = saveAsset("img", r.JoinURL(v), v, body)
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
			body = saveAsset("img", r.JoinURL(v), v, body)
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

	// Links to other pages — schedule g.Get immediately when a new page is discovered
	// to avoid the O(n²) pattern of re-queuing all known pages at the end of each call.
	r.HTMLDoc.Find("a").Each(func(i int, s *goquery.Selection) {
		data, exists := s.Attr("href")
		if !exists {
			return
		}
		parsedURL, err := url.Parse(data)
		if err != nil {
			fmt.Println("Error parsing URL:", err)
			return
		}
		if (parsedURL.Host == projectURL.Host || parsedURL.Host == "") && parsedURL.Path != "/" {
			if addAsset("pages", parsedURL.Path) {
				g.Get(r.JoinURL(data), quotesParse)
			}
		}
	})

	if urlPath == "" {
		addAsset("pages", urlPath)
	}

	// Write page to disk
	var filePath string
	cleanURLPath := strings.TrimPrefix(urlPath, "/")
	if cleanURLPath == "" {
		filePath = filepath.Join(projectPath, "index.html")
	} else {
		filePath = filepath.Join(projectPath, cleanURLPath, "index.html")
	}
	index, err := fsutil.OpenFile(filePath, fsutil.FsCWTFlags, 0o666)
	if err != nil {
		log.Printf("quotesParse: open %s: %v", filePath, err)
		return
	}
	body = strings.ReplaceAll(body, domain, "")
	if _, err := fsutil.WriteOSFile(index, body); err != nil {
		log.Printf("quotesParse: write %s: %v", filePath, err)
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
