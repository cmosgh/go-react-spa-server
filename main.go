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
		filePath := filepath.Join(staticDir, r.URL.Path)

		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			http.ServeFile(w, r, filepath.Join(staticDir, "index.html"))
			return
		}

		fs.ServeHTTP(w, r)
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
