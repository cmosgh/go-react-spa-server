package server

import (
	"log"
	"net/http"
	"os"

	"github.com/NYTimes/gziphandler" // For gzip compression
)

func GetStaticDir() string {
	// 1. Try to load from config file
	config, err := LoadConfig()
	if err != nil {
		log.Printf("Error loading config file: %v", err)
	}

	staticDir := ""
	if config != nil && config.StaticDir != "" {
		staticDir = config.StaticDir
		log.Printf("Using static directory from config file: %s", staticDir)
	}

	// 2. Check STATIC_DIR environment variable (takes precedence)
	envStaticDir := os.Getenv("STATIC_DIR")
	if envStaticDir != "" {
		staticDir = envStaticDir
		log.Printf("Using static directory from environment variable: %s", staticDir)
	}

	// 3. Default fallback
	if staticDir == "" {
		staticDir = "./client/dist"
		log.Printf("Using default static directory: %s", staticDir)
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
