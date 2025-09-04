package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/NYTimes/gziphandler" // For gzip compression
)

// cacheControlMiddleware sets appropriate Cache-Control headers for static assets.
func cacheControlMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// For assets with content hashes (e.g., in /assets/ or with specific extensions)
		// set a long cache duration and immutable.
		if strings.HasPrefix(r.URL.Path, "/assets/") ||
			strings.HasSuffix(r.URL.Path, ".js") ||
			strings.HasSuffix(r.URL.Path, ".css") ||
			strings.HasSuffix(r.URL.Path, ".png") ||
			strings.HasSuffix(r.URL.Path, ".jpg") ||
			strings.HasSuffix(r.URL.Path, ".jpeg") ||
			strings.HasSuffix(r.URL.Path, ".gif") ||
			strings.HasSuffix(r.URL.Path, ".svg") ||
			strings.HasSuffix(r.URL.Path, ".webp") {
			w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		} else if r.URL.Path == "/" || r.URL.Path == "/index.html" {
			// For index.html, set no-cache to ensure fresh content on every visit
			w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
			w.Header().Set("Pragma", "no-cache")
			w.Header().Set("Expires", "0")
		} else {
			// Check if the requested path corresponds to an actual file in the static directory.
			// Only apply default cache control if it's a static file.
			staticDir := getStaticDir()
			filePath := filepath.Join(staticDir, r.URL.Path)
			if _, err := os.Stat(filePath); err == nil { // File exists
				w.Header().Set("Cache-Control", "public, max-age=3600")
			}
		}
		next.ServeHTTP(w, r)
	})
}

func getStaticDir() string {
	staticDir := os.Getenv("STATIC_DIR")
	if staticDir == "" {
		staticDir = "./client/dist"
	}
	return staticDir
}

// startServer encapsulates the server startup logic.
func startServer(addr string, handler http.Handler) error {
	log.Printf("Listening on %s...", addr)
	return http.ListenAndServe(addr, handler)
}

// runApp sets up and starts the HTTP server.
// createSpaHandler creates an http.Handler that serves static files
// and falls back to index.html for client-side routes.
func createSpaHandler(staticDir string) http.Handler {
	fs := http.FileServer(http.Dir(staticDir))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
			// If file doesn't exist even after fallback (shouldn't happen for index.html),
			// or other error, let http.FileServer handle it or return 404.
			http.NotFound(w, r)
			return
		}

		// Generate ETag
		etag := fmt.Sprintf("%x-%x", fileInfo.ModTime().Unix(), fileInfo.Size())
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

func setupHandlers() http.Handler {
	staticDir := getStaticDir()

	spaHandler := createSpaHandler(staticDir)

	// Apply caching middleware
	cachedSPAHandler := cacheControlMiddleware(spaHandler)

	// Apply gzip compression middleware
	finalHandler := gziphandler.GzipHandler(cachedSPAHandler)

	return finalHandler
}

func runApp() error {
	finalHandler := setupHandlers()
	http.Handle("/", finalHandler)
	return startServer(":8080", nil)
}

func main() {
	if err := runApp(); err != nil {
		log.Fatal(err)
	}
}
