package server

import (
	"fmt" // Added import
	"log"
	"net/http"

	"github.com/NYTimes/gziphandler" // For gzip compression
)



// StartServer encapsulates the server startup logic.
func StartServer(config *Config, handler http.Handler) error {
	addr := fmt.Sprintf(":%d", config.Port) // Construct address from config.Port
	log.Printf("Listening on %s...", addr)
	return http.ListenAndServe(addr, handler)
}

func SetupHandlers() (http.Handler, *Config) {
	config, err := LoadConfig()
	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}

	// Log the static directory being used
	if config.StaticDir == "" {
		config.StaticDir = "./client/dist"
		log.Printf("Using default static directory: %s", config.StaticDir)
	} else {
		log.Printf("Using static directory: %s", config.StaticDir)
	}

	// Log the SPA fallback file being used
	log.Printf("Using SPA fallback file: %s", config.SpaFallbackFile)

	spaHandler := CreateSpaHandler(config) // Use CreateSpaHandler from handlers package

	// Apply caching middleware
	cachedSPAHandler := CacheControlMiddleware(config)(spaHandler) // Use CacheControlMiddleware from middleware package

	// Apply Brotli compression middleware (prioritized)
	brotliCompressedHandler := BrotliHandler(cachedSPAHandler) // Use BrotliHandler from middleware package

	// Apply Gzip compression middleware (fallback)
	finalHandler := gziphandler.GzipHandler(brotliCompressedHandler)

	return finalHandler, config
}
