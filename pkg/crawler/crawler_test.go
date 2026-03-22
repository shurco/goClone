package crawler

import (
	"net/url"
	"testing"
)

func withEnv(t *testing.T, raw string) func() {
	t.Helper()
	u, err := url.Parse(raw)
	if err != nil {
		t.Fatalf("parse url: %v", err)
	}
	projectURL = u
	domain = u.Scheme + "://" + u.Host
	projectPath = t.TempDir()
	crawlCtx = t.Context()
	files = filesBase{}
	old := downloadAsset
	downloadAsset = func(link, projectPath string) {}
	return func() { downloadAsset = old }
}

func Test_saveAsset_JS_RewriteAndCollect(t *testing.T) {
	cleanup := withEnv(t, "https://example.com")
	defer cleanup()

	body := `<script src="/static/app.js"></script>`
	abs := "https://example.com/static/app.js"
	out := saveAsset("js", abs, "/static/app.js", body)
	if out == body {
		t.Fatalf("body not rewritten")
	}
	if len(files.js) != 1 || files.js[0] != abs {
		t.Fatalf("js not collected: %+v", files.js)
	}
}

func Test_saveAsset_IMG_RewriteAndCollect(t *testing.T) {
	cleanup := withEnv(t, "https://example.com")
	defer cleanup()

	body := `<img src="/img/logo.png">`
	abs := "https://example.com/img/logo.png"
	out := saveAsset("img", abs, "/img/logo.png", body)
	if out == body {
		t.Fatalf("body not rewritten")
	}
	if len(files.img) != 1 || files.img[0] != abs {
		t.Fatalf("img not collected: %+v", files.img)
	}
}

func Test_readCSS_ResolvesRelativeAndRewrites(t *testing.T) {
	cleanup := withEnv(t, "https://example.com")
	defer cleanup()

	css := `body{background:url('../img/bg.jpg')}\n@font-face{src:url("/fonts/a.woff2")}`
	pageBody := "<style>" + css + "</style>"
	base, _ := url.Parse("https://example.com/assets/css/site.css")
	out := readCSS(css, pageBody, base)
	if out == pageBody {
		t.Fatalf("css references not rewritten")
	}
	if len(files.img) != 1 || len(files.font) != 1 {
		t.Fatalf("assets not collected: img=%d font=%d", len(files.img), len(files.font))
	}
}

func Test_readCSS_SkipsDataAndExternal(t *testing.T) {
	cleanup := withEnv(t, "https://example.com")
	defer cleanup()

	css := `body{background:url('data:image/png;base64,abc')}\n.div{background:url('https://external.com/a.png')}`
	pageBody := "<style>" + css + "</style>"
	base, _ := url.Parse("https://example.com/assets/css/site.css")
	out := readCSS(css, pageBody, base)
	if out != pageBody {
		t.Fatalf("data/external urls should not be rewritten")
	}
	if len(files.img) != 0 && len(files.font) != 0 {
		t.Fatalf("no assets should be collected for data/external")
	}
}

func Test_addAsset_Deduplication(t *testing.T) {
	cleanup := withEnv(t, "https://example.com")
	defer cleanup()

	if !addAsset("js", "https://example.com/a.js") {
		t.Fatalf("expected true on first add")
	}
	if addAsset("js", "https://example.com/a.js") {
		t.Fatalf("expected false on duplicate add")
	}
	if len(files.js) != 1 {
		t.Fatalf("expected 1 js entry, got %d", len(files.js))
	}
}

func Test_processSrcset_RewritesAll(t *testing.T) {
	cleanup := withEnv(t, "https://example.com")
	defer cleanup()
	in := `<img srcset="/img/a.png 1x, /img/b.png 2x">`
	body := in
	out := processSrcset("/img/a.png 1x, /img/b.png 2x", body, func(s string) string { return projectURL.String() + s })
	if out == body {
		t.Fatalf("srcset not rewritten")
	}
	if len(files.img) != 2 {
		t.Fatalf("expected 2 images collected, got %d", len(files.img))
	}
}

func Test_Preload_AsFont(t *testing.T) {
	cleanup := withEnv(t, "https://example.com")
	defer cleanup()
	body := `<link rel="preload" as="font" href="/font/a.woff2">`
	abs := "https://example.com/font/a.woff2"
	out := saveAsset("font", abs, "/font/a.woff2", body)
	if out == body {
		t.Fatalf("preload font not handled")
	}
	if len(files.font) != 1 {
		t.Fatalf("expected 1 font collected, got %d", len(files.font))
	}
}
