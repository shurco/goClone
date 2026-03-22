package crawler

import (
	"log"
	"sync"
	"time"

	"github.com/shurco/goClone/pkg/netutil"
)

var (
	maxConcurrentDownloads = 8
	downloadSemaphore      = make(chan struct{}, maxConcurrentDownloads)
	semMu                  sync.Mutex
	// downloadWg tracks in-flight background downloads; call downloadWg.Wait() to
	// block until all assets scheduled via downloadAsset have finished.
	downloadWg sync.WaitGroup
)

// SetDownloadConcurrency replaces the global download semaphore capacity so at most n
// asset downloads run concurrently. Values n <= 0 are ignored.
func SetDownloadConcurrency(n int) {
	if n <= 0 {
		return
	}
	semMu.Lock()
	defer semMu.Unlock()
	maxConcurrentDownloads = n
	downloadSemaphore = make(chan struct{}, maxConcurrentDownloads)
}

func init() {
	downloadAsset = func(link, path string) {
		downloadWg.Add(1)
		go func() {
			defer downloadWg.Done()
			downloadSemaphore <- struct{}{}
			defer func() { <-downloadSemaphore }()

			const maxAttempts = 3
			for attempt := 1; attempt <= maxAttempts; attempt++ {
				if crawlCtx.Err() != nil {
					return // context cancelled — stop retrying
				}
				if err := netutil.Extractor(crawlCtx, link, path); err == nil {
					return
				}
				time.Sleep(time.Duration(attempt) * 300 * time.Millisecond)
			}
			log.Printf("download failed after retries: %s", link)
		}()
	}
}
