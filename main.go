package main

import (
	"log"
	"net/http"

	"go-react-spa-server/server" // Import the new server package
)

func runApp() error {
	staticDir := server.GetStaticDir() // Use GetStaticDir from server package

	if err := server.LoadCriticalAssetsIntoCache(staticDir); err != nil {
		log.Printf("Error loading critical assets into cache: %v", err)
		// Continue, as it's not a fatal error if assets are served from disk
	}

	finalHandler := server.SetupHandlers() // Use SetupHandlers from server package
	http.Handle("/", finalHandler)
	return server.StartServer(":8080", nil) // Use StartServer from server package
}

func main() {
	if err := runApp(); err != nil {
		log.Fatal(err)
	}
}