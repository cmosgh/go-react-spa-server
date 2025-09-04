package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
)

func main() {
	staticDir := os.Getenv("STATIC_DIR")
	if staticDir == "" {
		staticDir = "./client/dist"
	}

	// The file server for static assets
	fs := http.FileServer(http.Dir(staticDir))

	// This handler will serve the SPA
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
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

	log.Println("Listening on :8080...")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}
