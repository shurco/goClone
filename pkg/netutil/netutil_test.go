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

func Test_Extractor_SuccessPNG(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/img/a.png", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("PNG"))
	})
	ts := httptest.NewServer(mux)
	defer ts.Close()

	tmp := t.TempDir()
	url := ts.URL + "/img/a.png"
	if err := Extractor(url, tmp); err != nil {
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

	tmp := t.TempDir()
	if err := Extractor(ts.URL+"/missing.png", tmp); err == nil {
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

	tmp := t.TempDir()
	if err := Extractor(ts.URL+"/big.bin", tmp); err == nil {
		t.Fatalf("expected too large error")
	}
}
