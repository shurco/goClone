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
	// arrange environment
	u, _ := url.Parse("https://example.com")
	projectURL = u
	domain = u.Scheme + "://" + u.Host
	projectPath = t.TempDir()
	files = filesBase{}

	// server that tracks concurrent in-flight requests
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

	// set low concurrency
	SetDownloadConcurrency(3)

	// act: schedule many downloads
	const jobs = 10
	var wg sync.WaitGroup
	wg.Add(jobs)
	for i := 0; i < jobs; i++ {
		go func(i int) {
			defer wg.Done()
			link := ts.URL + "/img/file" + time.Now().Format("150405.000") + ".png"
			// call package-level downloader (uses semaphore) directly
			downloadAsset(link, projectPath)
		}(i)
	}
	wg.Wait()

	// assert
	if maxInflight > 3 {
		t.Fatalf("expected max concurrency <= 3, got %d", maxInflight)
	}

	// and files should have been written under assets/img
	imgDir := filepath.Join(projectPath, "assets/img")
	if _, err := http.Dir(projectPath).Open("assets/img"); err != nil {
		t.Fatalf("expected assets/img directory to exist: %v", err)
	}
	_ = imgDir
}
