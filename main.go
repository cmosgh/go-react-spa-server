package main

import (
	"log"
	"net/http"

	"go-react-spa-server/server" // Import the new server package
)

func runApp() error {
	finalHandler, config := server.SetupHandlers() // Load config and setup handlers

	if err := server.LoadCriticalAssetsIntoCache(config.StaticDir); err != nil {
		log.Printf("Error loading critical assets into cache: %v", err)
		// Continue, as it's not a fatal error if assets are served from disk
	}

	http.Handle("/", finalHandler)
	return server.StartServer(config, finalHandler) // Use StartServer from server package
}

func main() {
	if err := runApp(); err != nil {
		log.Fatal(err)
	}
}