package crawler

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"testing"
)

func setEnvNoStub(t *testing.T, raw string) func() {
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
	return func() {}
}

func TestIntegration_DownloadsAssets(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/img/a.png", func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte("PNG")) })
	mux.HandleFunc("/js/app.js", func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte("console.log(1)")) })
	mux.HandleFunc("/css/site.css", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("body{background:url('../img/a.png')}"))
	})
	ts := httptest.NewServer(mux)
	defer ts.Close()

	cleanup := setEnvNoStub(t, ts.URL)
	defer cleanup()

	_ = saveAsset("img", ts.URL+"/img/a.png", "/img/a.png", "")
	_ = saveAsset("js", ts.URL+"/js/app.js", "/js/app.js", "")
	_ = readCSS("body{background:url('../img/a.png')}", "style", mustParseURL(t, ts.URL+"/css/site.css"))

	// Wait for all background downloads to complete instead of polling.
	downloadWg.Wait()

	check := func(p string) bool { _, err := os.Stat(p); return err == nil }
	imgPath := filepath.Join(projectPath, "assets/img", "img-a.png")
	jsPath := filepath.Join(projectPath, "assets/js", "js-app.js")
	// Image referenced from CSS resolves to /img/a.png -> "img-a.png"
	cssAsset := filepath.Join(projectPath, "assets/img", "img-a.png")

	if !check(imgPath) || !check(jsPath) || !check(cssAsset) {
		t.Fatalf("expected assets to be downloaded: img=%v js=%v cssAsset=%v",
			check(imgPath), check(jsPath), check(cssAsset))
	}
}

func mustParseURL(t *testing.T, s string) *url.URL {
	t.Helper()
	u, err := url.Parse(s)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	return u
}
