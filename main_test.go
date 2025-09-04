package main

import (
	"compress/gzip"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/NYTimes/gziphandler"
)

// spaHandler serves a single-page application.
// It serves static files from the staticDir, and for any path that doesn't match a file,
// it serves the index.html file.
// NOTE: This is the original spaHandler from main_test.go, not the one from main.go
// as main.go's logic is now embedded in the main function's handler setup.
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

func TestCacheControlMiddleware(t *testing.T) {
	// A dummy handler to pass to the middleware
	dummyHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	tests := []struct {
		name        string
		path        string
		expected    string
		notExpected string
	}{
		{
			name:     "hashed JS asset",
			path:     "/assets/index-CtS_vSNO.js",
			expected: "public, max-age=31536000, immutable",
		},
		{
			name:     "CSS asset",
			path:     "/style.css",
			expected: "public, max-age=31536000, immutable",
		},
		{
			name:     "PNG image asset",
			path:     "/image.png",
			expected: "public, max-age=31536000, immutable",
		},
		{
			name:     "root path",
			path:     "/",
			expected: "no-cache, no-store, must-revalidate",
		},
		{
			name:     "index.html path",
			path:     "/index.html",
			expected: "no-cache, no-store, must-revalidate",
		},
		{
			name:        "non-asset path",
			path:        "/api/data",
			notExpected: "max-age", // Should not have a cache-control header set by this middleware
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.path, nil)
			rr := httptest.NewRecorder()

			handler := cacheControlMiddleware(dummyHandler)
			handler.ServeHTTP(rr, req)

			cacheControl := rr.Header().Get("Cache-Control")
			if tt.expected != "" && cacheControl != tt.expected {
				t.Errorf("Cache-Control header mismatch for %s: got %q, want %q", tt.path, cacheControl, tt.expected)
			}
			if tt.notExpected != "" && strings.Contains(cacheControl, tt.notExpected) {
				t.Errorf("Cache-Control header should not contain %q for %s, but got %q", tt.notExpected, tt.path, cacheControl)
			}

			if strings.Contains(tt.expected, "no-cache") {
				if rr.Header().Get("Pragma") != "no-cache" {
					t.Errorf("Pragma header mismatch for %s: got %q, want %q", tt.path, rr.Header().Get("Pragma"), "no-cache")
				}
				if rr.Header().Get("Expires") != "0" {
					t.Errorf("Expires header mismatch for %s: got %q, want %q", tt.path, rr.Header().Get("Expires"), "0")
				}
			}
		})
	}
}

func TestGzipCompression(t *testing.T) {
	// Create a dummy handler that serves some content
	dummyContent := strings.Repeat("This is some content to be gzipped. ", 100) // Make it larger than 1400 bytes
	dummyHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte(dummyContent))
	})

	// Apply the gzip handler
	gzippedHandler := gziphandler.GzipHandler(dummyHandler)

	t.Run("should gzip content when Accept-Encoding is gzip", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Accept-Encoding", "gzip")
		rr := httptest.NewRecorder()

		gzippedHandler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
		}

		contentEncoding := rr.Header().Get("Content-Encoding")
		if contentEncoding != "gzip" {
			t.Errorf("Content-Encoding header mismatch: got %q, want %q", contentEncoding, "gzip")
		}

		// Decompress the body to verify content
		reader, err := gzip.NewReader(rr.Body)
		if err != nil {
			t.Fatalf("failed to create gzip reader: %v", err)
		}
		defer reader.Close()

		decompressedBody, err := ioutil.ReadAll(reader)
		if err != nil {
			t.Fatalf("failed to decompress body: %v", err)
		}

		if string(decompressedBody) != dummyContent {
			t.Errorf("decompressed body mismatch: got %q, want %q", string(decompressedBody), dummyContent)
		}
	})

	t.Run("should not gzip content when Accept-Encoding is not gzip", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		rr := httptest.NewRecorder()

		gzippedHandler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
		}

		contentEncoding := rr.Header().Get("Content-Encoding")
		if contentEncoding != "" {
			t.Errorf("Content-Encoding header should be empty, got %q", contentEncoding)
		}

		if rr.Body.String() != dummyContent {
			t.Errorf("body mismatch: got %q, want %q", rr.Body.String(), dummyContent)
		}
	})
}

func TestGetStaticDir(t *testing.T) {
	t.Run("STATIC_DIR is set", func(t *testing.T) {
		os.Setenv("STATIC_DIR", "/tmp/custom/dist")
		defer os.Unsetenv("STATIC_DIR") // Clean up after test

		dir := getStaticDir()
		expected := "/tmp/custom/dist"
		if dir != expected {
			t.Errorf("getStaticDir() returned %q, want %q when STATIC_DIR is set", dir, expected)
		}
	})

	t.Run("STATIC_DIR is not set", func(t *testing.T) {
		os.Unsetenv("STATIC_DIR") // Ensure it's not set
		dir := getStaticDir()
		expected := "./client/dist"
		if dir != expected {
			t.Errorf("getStaticDir() returned %q, want %q when STATIC_DIR is not set", dir, expected)
		}
	})
}