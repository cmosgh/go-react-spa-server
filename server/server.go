package server

import (
	"log"
	"net/http"
	"os"

	"github.com/NYTimes/gziphandler" // For gzip compression
)

func GetStaticDir() string {
	staticDir := os.Getenv("STATIC_DIR")
	if staticDir == "" {
		staticDir = "./client/dist"
	}
	return staticDir
}

// StartServer encapsulates the server startup logic.
func StartServer(addr string, handler http.Handler) error {
	log.Printf("Listening on %s...", addr)
	return http.ListenAndServe(addr, handler)
}

func SetupHandlers() http.Handler {
	staticDir := GetStaticDir()

	spaHandler := CreateSpaHandler(staticDir) // Use CreateSpaHandler from handlers package

	// Apply caching middleware
	cachedSPAHandler := CacheControlMiddleware(spaHandler) // Use CacheControlMiddleware from middleware package

	// Apply Brotli compression middleware (prioritized)
	brotliCompressedHandler := BrotliHandler(cachedSPAHandler) // Use BrotliHandler from middleware package

	// Apply Gzip compression middleware (fallback)
	finalHandler := gziphandler.GzipHandler(brotliCompressedHandler)

	return finalHandler
}
