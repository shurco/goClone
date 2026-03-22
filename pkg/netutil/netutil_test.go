package netutil

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func Test_SetAssetRoot_UpdatesMappings(t *testing.T) {
	origRoot := AssetRoot
	origFolders := map[string]string{"css": Folders["css"], "js": Folders["js"], "img": Folders["img"], "font": Folders["font"]}
	defer func() { SetAssetRoot(origRoot); Folders = origFolders }()

	SetAssetRoot("static")
	if AssetRoot != "static" {
		t.Fatalf("asset root not set")
	}
	if Folders["css"] != "static/css" || Folders["js"] != "static/js" || Folders["img"] != "static/img" || Folders["font"] != "static/font" {
		t.Fatalf("folders not updated: %+v", Folders)
	}
	if GetAssetDir("/path/a.css") != "static/css" {
		t.Fatalf("GetAssetDir not using new root")
	}
}

func Test_ReplaceSlashWithDash(t *testing.T) {
	if got := ReplaceSlashWithDash("/a/b/c.png"); got != "a-b-c.png" {
		t.Fatalf("unexpected: %s", got)
	}
}

func Test_GetAssetDir(t *testing.T) {
	if dir := GetAssetDir("/img/a.png"); !strings.HasSuffix(dir, "/img") {
		t.Fatalf("expected img dir, got %s", dir)
	}
	if dir := GetAssetDir("/unknown/file.xyz"); dir != "" {
		t.Fatalf("expected empty dir for unknown ext, got %s", dir)
	}
}

func Test_urlExtension_StripsQuery(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{"/a/b.svg?v=123", ".svg"},
		{"/a/b.png#anchor", ".png"},
		{"/a/b.woff2?foo=bar&baz=1", ".woff2"},
		{"/a/b.js", ".js"},
		{"/a/b.css?version=42", ".css"},
		{"/no-ext/path", ""},
	}
	for _, tc := range cases {
		if got := urlExtension(tc.input); got != tc.want {
			t.Errorf("urlExtension(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

func Test_GetAssetDir_NewExtensions(t *testing.T) {
	cases := []struct {
		file string
		want string
	}{
		{"/img/photo.webp", "assets/img"},
		{"/icons/favicon.ico", "assets/img"},
		{"/img/hero.avif", "assets/img"},
	}
	for _, tc := range cases {
		if got := GetAssetDir(tc.file); got != tc.want {
			t.Errorf("GetAssetDir(%q) = %q, want %q", tc.file, got, tc.want)
		}
	}
}

func Test_Extractor_SuccessPNG(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/img/a.png", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("PNG"))
	})
	ts := httptest.NewServer(mux)
	defer ts.Close()

	tmp := t.TempDir()
	if err := Extractor(t.Context(), ts.URL+"/img/a.png", tmp); err != nil {
		t.Fatalf("extractor error: %v", err)
	}
	p := filepath.Join(tmp, Folders["img"], "img-a.png")
	if _, err := os.Stat(p); err != nil {
		t.Fatalf("expected file downloaded: %v", err)
	}
}

func Test_Extractor_StatusError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/missing.png", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
	})
	ts := httptest.NewServer(mux)
	defer ts.Close()

	if err := Extractor(t.Context(), ts.URL+"/missing.png", t.TempDir()); err == nil {
		t.Fatalf("expected error on 404")
	}
}

func Test_Extractor_TooLarge(t *testing.T) {
	orig := maxDownloadBytes
	defer func() { maxDownloadBytes = orig }()
	SetMaxDownloadBytes(5)

	mux := http.NewServeMux()
	mux.HandleFunc("/big.bin", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "10")
		_, _ = w.Write([]byte("0123456789"))
	})
	ts := httptest.NewServer(mux)
	defer ts.Close()

	if err := Extractor(t.Context(), ts.URL+"/big.bin", t.TempDir()); err == nil {
		t.Fatalf("expected too large error")
	}
}

func Test_Extractor_QueryParamURL(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/assets/logo.svg", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("<svg/>"))
	})
	ts := httptest.NewServer(mux)
	defer ts.Close()

	tmp := t.TempDir()
	// URL with query parameter — extension must be resolved correctly.
	if err := Extractor(t.Context(), ts.URL+"/assets/logo.svg?v=42", tmp); err != nil {
		t.Fatalf("extractor error: %v", err)
	}
	p := filepath.Join(tmp, Folders["img"], "assets-logo.svg")
	if _, err := os.Stat(p); err != nil {
		t.Fatalf("expected file %s to exist: %v", p, err)
	}
}
