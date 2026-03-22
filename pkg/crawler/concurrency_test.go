package crawler

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func Test_DownloadConcurrencyLimit(t *testing.T) {
	u, _ := url.Parse("https://example.com")
	projectURL = u
	domain = u.Scheme + "://" + u.Host
	projectPath = t.TempDir()
	crawlCtx = t.Context()
	files = filesBase{}

	// Server that tracks concurrent in-flight requests.
	var inflight int32
	var maxInflight int32
	mux := http.NewServeMux()
	mux.HandleFunc("/img/", func(w http.ResponseWriter, r *http.Request) {
		cur := atomic.AddInt32(&inflight, 1)
		for {
			old := atomic.LoadInt32(&maxInflight)
			if cur <= old || atomic.CompareAndSwapInt32(&maxInflight, old, cur) {
				break
			}
		}
		time.Sleep(120 * time.Millisecond)
		atomic.AddInt32(&inflight, -1)
		_, _ = w.Write([]byte("PNG"))
	})
	ts := httptest.NewServer(mux)
	defer ts.Close()

	SetDownloadConcurrency(3)

	const jobs = 10
	var wg sync.WaitGroup
	for range jobs {
		wg.Go(func() {
			link := ts.URL + "/img/file" + time.Now().Format("150405.000") + ".png"
			downloadAsset(link, projectPath)
		})
	}
	// Wait for all wg.Go goroutines (they launch inner download goroutines via downloadWg).
	wg.Wait()
	// Wait for all inner download goroutines to complete.
	downloadWg.Wait()

	if maxInflight > 3 {
		t.Fatalf("expected max concurrency <= 3, got %d", maxInflight)
	}

	if _, err := http.Dir(projectPath).Open("assets/img"); err != nil {
		t.Fatalf("expected assets/img directory to exist: %v", err)
	}
	_ = filepath.Join(projectPath, "assets/img")
}
