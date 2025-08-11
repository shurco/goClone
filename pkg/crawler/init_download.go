package crawler

import (
	"log"
	"sync"
	"time"

	"github.com/shurco/goClone/pkg/netutil"
)

var (
	// semaphore channel to limit concurrent downloads
	maxConcurrentDownloads = 8
	downloadSemaphore      = make(chan struct{}, maxConcurrentDownloads)
	semMu                  sync.Mutex
)

// SetDownloadConcurrency reconfigures max parallel download workers.
func SetDownloadConcurrency(n int) {
	if n <= 0 {
		return
	}
	semMu.Lock()
	defer semMu.Unlock()
	maxConcurrentDownloads = n
	old := downloadSemaphore
	downloadSemaphore = make(chan struct{}, maxConcurrentDownloads)
	// drain old semaphore if used; not strictly necessary in tests
	_ = old
}

func init() {
	// set default downloader implementation with concurrency limit and retries
	downloadAsset = func(link, projectPath string) {
		// acquire
		downloadSemaphore <- struct{}{}
		defer func() { <-downloadSemaphore }()

		const maxAttempts = 3
		for attempt := 1; attempt <= maxAttempts; attempt++ {
			if err := netutil.Extractor(link, projectPath); err == nil {
				return
			}
			// simple backoff
			time.Sleep(time.Duration(attempt) * 300 * time.Millisecond)
		}
		log.Printf("download failed after retries: %s", link)
	}
}
