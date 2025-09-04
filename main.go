package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	

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
		}
		next.ServeHTTP(w, r)
	})
}

func main() {
	staticDir := os.Getenv("STATIC_DIR")
	if staticDir == "" {
		staticDir = "./client/dist"
	}

	// The file server for static assets
	fs := http.FileServer(http.Dir(staticDir))

	// Create a handler that serves the SPA, applying caching and gzip
	spaHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Construct the path to the requested file in the static directory
		filePath := filepath.Join(staticDir, r.URL.Path)

		// Check if a file exists at the constructed path.
		// If not, it's likely a client-side route.
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			// File does not exist, serve index.html
			http.ServeFile(w, r, filepath.Join(staticDir, "index.html"))
			return
		}

		// File exists, let the file server handle it.
		fs.ServeHTTP(w, r)
	})

	// Apply caching middleware
	cachedSPAHandler := cacheControlMiddleware(spaHandler)

	// Apply gzip compression middleware
	finalHandler := gziphandler.GzipHandler(cachedSPAHandler)

	// Register the final handler
	http.Handle("/", finalHandler)

	log.Println("Listening on :8080...")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}