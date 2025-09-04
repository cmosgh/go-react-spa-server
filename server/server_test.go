package server

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"path/filepath"

	"compress/gzip"
	"io/ioutil"
)

func TestGetStaticDir(t *testing.T) {
	t.Run("STATIC_DIR is set", func(t *testing.T) {
		os.Setenv("STATIC_DIR", "/tmp/custom/dist")
		defer os.Unsetenv("STATIC_DIR") // Clean up after test

		dir := GetStaticDir()
		expected := "/tmp/custom/dist"
		if dir != expected {
			t.Errorf("GetStaticDir() returned %q, want %q when STATIC_DIR is set", dir, expected)
		}
	})

	t.Run("STATIC_DIR is not set", func(t *testing.T) {
		os.Unsetenv("STATIC_DIR") // Ensure it's not set
		dir := GetStaticDir()
		expected := "./client/dist"
		if dir != expected {
			t.Errorf("GetStaticDir() returned %q, want %q when STATIC_DIR is not set", dir, expected)
		}
	})
}

func TestStartServer(t *testing.T) {
	// Create a dummy handler
	dummyHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	t.Run("server fails to start with invalid address", func(t *testing.T) {
		// Call StartServer with an invalid address that will cause ListenAndServe to fail immediately
		err := StartServer(":invalid_port", dummyHandler)
		if err == nil {
			t.Errorf("StartServer() did not return an error for invalid address")
		}
	})
}

func TestSetupHandlers(t *testing.T) {
	// Create a temporary directory for this test
	tempStaticDir, err := ioutil.TempDir("", "test_static_dir_setup_handlers")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempStaticDir) // Clean up after test

	// Temporarily set STATIC_DIR to the temporary directory for testing
	os.Setenv("STATIC_DIR", tempStaticDir)
	defer os.Unsetenv("STATIC_DIR")

	// Create a temporary large file for testing gzip
	largeContent := strings.Repeat("a", 2000) // Content larger than typical gzip threshold
	tempFilePath := filepath.Join(tempStaticDir, "temp_large_file.txt")
	err = ioutil.WriteFile(tempFilePath, []byte(largeContent), 0644)
	if err != nil {
		t.Fatalf("failed to create temporary file: %v", err)
	}

	handler := SetupHandlers()

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
		tempFilePath := filepath.Join(tempStaticDir, "temp_large_file.txt")
		err = ioutil.WriteFile(tempFilePath, []byte(largeContent), 0644)
		if err != nil {
			t.Fatalf("failed to create temporary file: %v", err)
		}

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