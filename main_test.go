package main

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// spaHandler serves a single-page application.
// It serves static files from the staticDir, and for any path that doesn't match a file,
// it serves the index.html file.
func spaHandler(staticDir string) http.Handler {
	// The file server for static assets
	fs := http.FileServer(http.Dir(staticDir))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
}

func TestSpaHandler(t *testing.T) {
	// The handler to test
	handler := spaHandler("./static")

	// --- Test Cases ---

	t.Run("serves index.html for the root route", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
		}

		body := rr.Body.String()
		if !strings.Contains(body, "<title>Vite + React</title>") {
			t.Errorf("body should contain the Vite + React title")
		}
		if !strings.Contains(body, `<div id="root"></div>`) {
			t.Errorf(`body should contain '<div id="root"></div>'`)
		}
	})

	t.Run("serves index.html for a non-existent client-side route", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/some/client/route", nil)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
		}

		body := rr.Body.String()
		if !strings.Contains(body, "<title>Vite + React</title>") {
			t.Errorf("body should contain the Vite + React title")
		}
	})

	t.Run("serves an existing static asset", func(t *testing.T) {
		// Find the actual asset file to test against.
		files, err := ioutil.ReadDir("./static/assets")
		if err != nil {
			t.Fatalf("could not read static/assets directory: %v", err)
		}
		var jsFile string
		for _, file := range files {
			if strings.HasSuffix(file.Name(), ".js") {
				jsFile = file.Name()
				break
			}
		}
		if jsFile == "" {
			t.Fatal("could not find a JS asset file to test")
		}

		req := httptest.NewRequest("GET", "/assets/"+jsFile, nil)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
		}

		// Check for a content type header
		contentType := rr.Header().Get("Content-Type")
		if !strings.Contains(contentType, "javascript") {
			t.Errorf("wrong content type for JS file: got %q, want it to contain 'javascript'", contentType)
		}

		if rr.Body.Len() == 0 {
			t.Errorf("body of JS asset should not be empty")
		}
	})
}
