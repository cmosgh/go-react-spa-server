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
	"time"

	"github.com/NYTimes/gziphandler"
)

func TestSpaHandler(t *testing.T) {
	// The handler to test
	handler := createSpaHandler("./static")

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

	t.Run("serves existing static asset with ETag and Last-Modified headers", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/vite.svg", nil)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
		}

		etag := rr.Header().Get("ETag")
		if etag == "" {
			t.Errorf("ETag header not found")
		}

		lastModified := rr.Header().Get("Last-Modified")
		if lastModified == "" {
			t.Errorf("Last-Modified header not found")
		}

		contentType := rr.Header().Get("Content-Type")
		if !strings.Contains(contentType, "image/svg+xml") {
			t.Errorf("wrong content type for SVG file: got %q, want it to contain 'image/svg+xml'", contentType)
		}
	})

	t.Run("returns 304 Not Modified for ETag match", func(t *testing.T) {
		// First request to get ETag
		req1 := httptest.NewRequest("GET", "/vite.svg", nil)
		rr1 := httptest.NewRecorder()
		handler.ServeHTTP(rr1, req1)
		etag := rr1.Header().Get("ETag")
		if etag == "" {
			t.Fatal("ETag not found in first response")
		}

		// Second request with If-None-Match
		req2 := httptest.NewRequest("GET", "/vite.svg", nil)
		req2.Header.Set("If-None-Match", etag)
		rr2 := httptest.NewRecorder()
		handler.ServeHTTP(rr2, req2)

		if status := rr2.Code; status != http.StatusNotModified {
			t.Errorf("handler returned wrong status code: got %v want %v, body: %s", status, http.StatusNotModified, rr2.Body.String())
		}
		if rr2.Body.Len() != 0 {
			t.Errorf("body should be empty for 304 response, got %d bytes", rr2.Body.Len())
		}
	})

	t.Run("returns 304 Not Modified for If-Modified-Since match", func(t *testing.T) {
		// First request to get Last-Modified
		req1 := httptest.NewRequest("GET", "/vite.svg", nil)
		rr1 := httptest.NewRecorder()
		handler.ServeHTTP(rr1, req1)
		lastModified := rr1.Header().Get("Last-Modified")
		if lastModified == "" {
			t.Fatal("Last-Modified not found in first response")
		}

		// Second request with If-Modified-Since
		req2 := httptest.NewRequest("GET", "/vite.svg", nil)
		req2.Header.Set("If-Modified-Since", lastModified)
		rr2 := httptest.NewRecorder()
		handler.ServeHTTP(rr2, req2)

		if status := rr2.Code; status != http.StatusNotModified {
			t.Errorf("handler returned wrong status code: got %v want %v, body: %s", status, http.StatusNotModified, rr2.Body.String())
		}
		if rr2.Body.Len() != 0 {
			t.Errorf("body should be empty for 304 response, got %d bytes", rr2.Body.Len())
		}
	})

	t.Run("returns 200 OK if ETag does not match", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/vite.svg", nil)
		req.Header.Set("If-None-Match", "invalid-etag")
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
		}
		if rr.Body.Len() == 0 {
			t.Errorf("body should not be empty for 200 OK response")
		}
	})

	t.Run("returns 200 OK if If-Modified-Since is older than actual modification", func(t *testing.T) {
		// Get current Last-Modified
		req1 := httptest.NewRequest("GET", "/vite.svg", nil)
		rr1 := httptest.NewRecorder()
		handler.ServeHTTP(rr1, req1)
		currentLastModified := rr1.Header().Get("Last-Modified")
		if currentLastModified == "" {
			t.Fatal("Last-Modified not found in first response")
		}

		// Parse current Last-Modified and subtract a day
		oldTime, err := http.ParseTime(currentLastModified)
		if err != nil {
			t.Fatalf("failed to parse Last-Modified time: %v", err)
		}
		oldTime = oldTime.Add(-24 * time.Hour)
		oldModifiedSince := oldTime.Format(http.TimeFormat)

		req2 := httptest.NewRequest("GET", "/vite.svg", nil)
		req2.Header.Set("If-Modified-Since", oldModifiedSince)
		rr2 := httptest.NewRecorder()
		handler.ServeHTTP(rr2, req2)

		if status := rr2.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
		}
		if rr2.Body.Len() == 0 {
			t.Errorf("body should not be empty for 200 OK response")
		}
	})
}

func TestCacheControlMiddleware(t *testing.T) {
	// Temporarily set STATIC_DIR to a known value for testing
	os.Setenv("STATIC_DIR", "./static")
	defer os.Unsetenv("STATIC_DIR") // Clean up after test

	// Create a dummy favicon.ico for testing generic static files
	faviconPath := filepath.Join("./static", "favicon.ico")
	err := ioutil.WriteFile(faviconPath, []byte("dummy favicon content"), 0644)
	if err != nil {
		t.Fatalf("failed to create dummy favicon.ico: %v", err)
	}
	defer os.Remove(faviconPath) // Clean up after test

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
			name:     "non-asset path",
			path:     "/api/data",
			notExpected: "max-age", // Should not have a cache-control header set by this middleware
		},
		{
			name:     "generic static file",
			path:     "/favicon.ico",
			expected: "public, max-age=3600",
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

func TestStartServer(t *testing.T) {
	// Create a dummy handler
	dummyHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	t.Run("server fails to start with invalid address", func(t *testing.T) {
		// Call startServer with an invalid address that will cause ListenAndServe to fail immediately
		err := startServer(":invalid_port", dummyHandler)
		if err == nil {
			t.Errorf("startServer() did not return an error for invalid address")
		}
	})
}

func TestSetupHandlers(t *testing.T) {
	// Temporarily set STATIC_DIR to a known value for testing
	os.Setenv("STATIC_DIR", "./static")
	defer os.Unsetenv("STATIC_DIR")

	handler := setupHandlers()

	t.Run("should apply cache control headers for assets", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/assets/some.js", nil)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		cacheControl := rr.Header().Get("Cache-Control")
		expected := "public, max-age=31536000, immutable"
		if cacheControl != expected {
			t.Errorf("Cache-Control header mismatch: got %q, want %q", cacheControl, expected)
		}
	})

	t.Run("should apply no-cache for index.html", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/index.html", nil)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		cacheControl := rr.Header().Get("Cache-Control")
		expected := "no-cache, no-store, must-revalidate"
		if cacheControl != expected {
			t.Errorf("Cache-Control header mismatch: got %q, want %q", cacheControl, expected)
		}
		if rr.Header().Get("Pragma") != "no-cache" {
			t.Errorf("Pragma header mismatch: got %q, want %q", rr.Header().Get("Pragma"), "no-cache")
		}
		if rr.Header().Get("Expires") != "0" {
			t.Errorf("Expires header mismatch: got %q, want %q", rr.Header().Get("Expires"), "0")
		}
	})

	t.Run("should gzip content when Accept-Encoding is gzip", func(t *testing.T) {
		// Create a temporary large file for testing gzip
		largeContent := strings.Repeat("a", 2000) // Content larger than typical gzip threshold
		tempFilePath := filepath.Join("./static", "temp_large_file.txt")
		err := ioutil.WriteFile(tempFilePath, []byte(largeContent), 0644)
		if err != nil {
			t.Fatalf("failed to create temporary file: %v", err)
		}
		defer os.Remove(tempFilePath) // Clean up the temporary file

		req := httptest.NewRequest("GET", "/temp_large_file.txt", nil) // Request the temporary file
		req.Header.Set("Accept-Encoding", "gzip")
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		contentEncoding := rr.Header().Get("Content-Encoding")
		if contentEncoding != "gzip" {
			t.Errorf("Content-Encoding header mismatch: got %q, want %q", contentEncoding, "gzip")
		}

		// Verify content is gzipped by attempting to decompress
		reader, err := gzip.NewReader(rr.Body)
		if err != nil {
			t.Fatalf("failed to create gzip reader: %v", err)
		}
		defer reader.Close()
		decompressedBody, err := ioutil.ReadAll(reader)
		if err != nil {
			t.Fatalf("failed to decompress body: %v", err)
		}
		if string(decompressedBody) != largeContent {
			t.Errorf("decompressed body mismatch: got %q, want %q", string(decompressedBody), largeContent)
		}
	})
}
