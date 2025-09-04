package server

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// CreateSpaHandler creates an http.Handler that serves static files
// and falls back to index.html for client-side routes.
func CreateSpaHandler(staticDir string) http.Handler {
	fs := http.FileServer(http.Dir(staticDir))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Try to serve from in-memory cache first
		cachePath := r.URL.Path
		if cachePath == "/" {
			cachePath = "/index.html"
		}
		if cachedAsset, ok := GetCachedAsset(cachePath); ok { // Use GetCachedAsset from cache package
			// Set Content-Type
			w.Header().Set("Content-Type", cachedAsset.MimeType)

			// Generate ETag for cached content
			etag := fmt.Sprintf("\"%x-%x\"", cachedAsset.ModTime.Unix(), cachedAsset.Size)
			w.Header().Set("ETag", etag)

			// Set Last-Modified header for cached content
			w.Header().Set("Last-Modified", cachedAsset.ModTime.Format(http.TimeFormat))

			// Check If-None-Match for cached content
			ifNoneMatch := r.Header.Get("If-None-Match")
			if ifNoneMatch != "" && ifNoneMatch == etag {
				w.WriteHeader(http.StatusNotModified)
				return
			}

			// Check If-Modified-Since for cached content
			ifModifiedSince := r.Header.Get("If-Modified-Since")
			if ifModifiedSince != "" {
				t, err := http.ParseTime(ifModifiedSince)
				if err == nil && cachedAsset.ModTime.Before(t.Add(1*time.Second)) {
					w.WriteHeader(http.StatusNotModified)
					return
				}
			}

			
		w.Write(cachedAsset.Content)
			return
		}

		// Original logic for serving from disk if not in cache
		requestedPath := filepath.Join(staticDir, r.URL.Path)
		serveFilePath := requestedPath // Assume requested path initially

		// Check if the requested file exists, otherwise fallback to index.html
		_, err := os.Stat(requestedPath)
		if os.IsNotExist(err) {
			serveFilePath = filepath.Join(staticDir, "index.html")
		}

		// Get file info for ETag and Last-Modified
		fileInfo, err := os.Stat(serveFilePath)
		if err != nil {
			http.NotFound(w, r)
			return
		}

		// Generate ETag
		etag := fmt.Sprintf("\"%x-%x\"", fileInfo.ModTime().Unix(), fileInfo.Size())
		w.Header().Set("ETag", etag)

		// Set Last-Modified header
		w.Header().Set("Last-Modified", fileInfo.ModTime().Format(http.TimeFormat))

		// Check If-None-Match
		ifNoneMatch := r.Header.Get("If-None-Match")
		if ifNoneMatch != "" && ifNoneMatch == etag {
			w.WriteHeader(http.StatusNotModified)
			return
		}

		// Check If-Modified-Since
		ifModifiedSince := r.Header.Get("If-Modified-Since")
		if ifModifiedSince != "" {
				t, err := http.ParseTime(ifModifiedSince)
				if err == nil && fileInfo.ModTime().Before(t.Add(1*time.Second)) { // Add 1 second tolerance
					w.WriteHeader(http.StatusNotModified)
					return
				}
			}

		// If not 304, serve the file
		if serveFilePath == requestedPath {
			fs.ServeHTTP(w, r) // Serve the requested file
		} else {
			http.ServeFile(w, r, serveFilePath) // Serve index.html fallback
		}
	})
}
