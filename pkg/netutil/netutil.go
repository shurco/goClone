// Package netutil downloads remote assets into the mirror layout and maps file
// extensions to subdirectory paths under the configurable asset root.
package netutil

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/shurco/goClone/pkg/fsutil"
)

var (
	// AssetRoot is the root directory name for storing assets.
	AssetRoot = "assets"
	// Folders maps logical types to directories under AssetRoot.
	Folders = map[string]string{
		"css":  AssetRoot + "/css",
		"js":   AssetRoot + "/js",
		"img":  AssetRoot + "/img",
		"font": AssetRoot + "/font",
	}
	// Extensions maps file extensions to target directories.
	Extensions = map[string]string{
		".css":   AssetRoot + "/css",
		".js":    AssetRoot + "/js",
		".jpg":   AssetRoot + "/img",
		".jpeg":  AssetRoot + "/img",
		".gif":   AssetRoot + "/img",
		".png":   AssetRoot + "/img",
		".svg":   AssetRoot + "/img",
		".webp":  AssetRoot + "/img",
		".ico":   AssetRoot + "/img",
		".avif":  AssetRoot + "/img",
		".eot":   AssetRoot + "/font",
		".otf":   AssetRoot + "/font",
		".ttf":   AssetRoot + "/font",
		".woff":  AssetRoot + "/font",
		".woff2": AssetRoot + "/font",
	}
	defaultUserAgent       = "goclone"
	maxDownloadBytes int64 = 50 * 1024 * 1024 // 50MB
	httpClient             = &http.Client{Timeout: 20 * time.Second}
)

// SetAssetRoot updates AssetRoot and reconfigures all folders/ext mappings.
func SetAssetRoot(root string) {
	if strings.TrimSpace(root) == "" {
		return
	}
	AssetRoot = strings.Trim(root, "/")
	Folders["css"] = AssetRoot + "/css"
	Folders["js"] = AssetRoot + "/js"
	Folders["img"] = AssetRoot + "/img"
	Folders["font"] = AssetRoot + "/font"
	Extensions[".css"] = AssetRoot + "/css"
	Extensions[".js"] = AssetRoot + "/js"
	Extensions[".jpg"] = AssetRoot + "/img"
	Extensions[".jpeg"] = AssetRoot + "/img"
	Extensions[".gif"] = AssetRoot + "/img"
	Extensions[".png"] = AssetRoot + "/img"
	Extensions[".svg"] = AssetRoot + "/img"
	Extensions[".webp"] = AssetRoot + "/img"
	Extensions[".ico"] = AssetRoot + "/img"
	Extensions[".avif"] = AssetRoot + "/img"
	Extensions[".eot"] = AssetRoot + "/font"
	Extensions[".otf"] = AssetRoot + "/font"
	Extensions[".ttf"] = AssetRoot + "/font"
	Extensions[".woff"] = AssetRoot + "/font"
	Extensions[".woff2"] = AssetRoot + "/font"
}

// SetDefaultUserAgent sets the UA string for asset downloads.
func SetDefaultUserAgent(ua string) {
	if strings.TrimSpace(ua) != "" {
		defaultUserAgent = ua
	}
}

// SetMaxDownloadBytes sets a per-asset size limit for downloads.
func SetMaxDownloadBytes(n int64) {
	if n > 0 {
		maxDownloadBytes = n
	}
}

// SetHTTPTimeout sets the HTTP client timeout for asset fetches.
func SetHTTPTimeout(d time.Duration) {
	if d > 0 {
		httpClient.Timeout = d
	}
}

// Extractor downloads link and writes it to the appropriate directory under projectPath.
// The request is made with ctx so callers can cancel in-flight downloads.
func Extractor(ctx context.Context, link, projectPath string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, link, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", defaultUserAgent)

	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("unexpected status %d for %s", resp.StatusCode, link)
	}

	// Check Content-Length before creating any file on disk.
	if cl := resp.ContentLength; cl > 0 && cl > maxDownloadBytes {
		return fmt.Errorf("content too large: %d > %d", cl, maxDownloadBytes)
	}

	ext := urlExtension(resp.Request.URL.Path)
	if ext == "" {
		return errors.New("unknown file extension")
	}
	dirPath := Extensions[ext]
	if dirPath == "" {
		return fmt.Errorf("no dir mapping for extension %s", ext)
	}
	name := ReplaceSlashWithDash(resp.Request.URL.Path)
	full := filepath.Join(projectPath, dirPath, name)

	// Path traversal guard: ensure full path stays under projectPath.
	rel, err := filepath.Rel(projectPath, full)
	if err != nil || strings.HasPrefix(rel, "..") {
		return errors.New("invalid output path")
	}

	file, err := fsutil.OpenFile(full, fsutil.FsCWTFlags, 0o666)
	if err != nil {
		return err
	}
	defer file.Close()

	if _, err := io.Copy(file, io.LimitReader(resp.Body, maxDownloadBytes)); err != nil {
		return err
	}
	return nil
}

// GetAssetDir returns the asset directory for a given filename by extension.
func GetAssetDir(filename string) string {
	return Extensions[urlExtension(filename)]
}

// urlExtension returns the file extension from a URL path, stripping query and
// fragment before calling path.Ext so that "/a.svg?v=1" correctly returns ".svg".
func urlExtension(rawURL string) string {
	if i := strings.IndexAny(rawURL, "?#"); i >= 0 {
		rawURL = rawURL[:i]
	}
	return path.Ext(rawURL)
}

// IsValidDomain reports whether domain looks like a multi-label hostname and resolves
// via DNS ([net.LookupHost]). It returns false for empty, overlong, or malformed names.
func IsValidDomain(domain string) bool {
	if len(domain) < 1 || len(domain) > 255 {
		return false
	}

	if match, _ := regexp.MatchString(`^([a-zA-Z0-9-]+\.)+[a-zA-Z]{2,}$`, domain); !match {
		return false
	}

	if _, err := net.LookupHost(domain); err != nil {
		return false
	}

	return true
}

// ReplaceSlashWithDash turns a URL path into a flat filename fragment by stripping a
// leading slash and replacing each '/' with '-', e.g. "/a/b.css" -> "a-b.css".
func ReplaceSlashWithDash(input string) string {
	input = strings.TrimPrefix(input, "/")
	return strings.ReplaceAll(input, "/", "-")
}
