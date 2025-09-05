package server

import (
	"compress/gzip"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/andybalholm/brotli"
)

func TestCacheControlMiddleware(t *testing.T) {
	// Create a temporary directory for this test
	tempStaticDir, err := ioutil.TempDir("", "test_static_dir_middleware")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempStaticDir) // Clean up after test

	// Create a dummy favicon.ico in the temporary directory for testing generic static files
	faviconPath := filepath.Join(tempStaticDir, "favicon.ico")
	err = ioutil.WriteFile(faviconPath, []byte("dummy favicon content"), 0644)
	if err != nil {
		t.Fatalf("failed to create dummy favicon.ico: %v", err)
	}

	// Create a config for the middleware
	cfg := &Config{
		StaticDir: tempStaticDir,
		SpaFallbackFile: "index.html",
	}

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
			path:     "/" + cfg.SpaFallbackFile,
			expected: "no-cache, no-store, must-revalidate",
		},
		{
			name:        "non-asset path",
			path:        "/api/data",
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

			handler := CacheControlMiddleware(cfg)(dummyHandler)
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

func TestCompression(t *testing.T) {
	// Create a temporary directory for this test
	tempStaticDir, err := ioutil.TempDir("", "test_static_dir_compression")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempStaticDir) // Clean up after test

	// Temporarily set STATIC_DIR to the temporary directory for testing
	os.Setenv("STATIC_DIR", tempStaticDir)
	defer os.Unsetenv("STATIC_DIR")

	// Create a temporary large file for testing compression
	largeContent := strings.Repeat("a", 2000) // Content larger than typical compression threshold
	tempFilePath := filepath.Join(tempStaticDir, "temp_large_file.txt")
	err = ioutil.WriteFile(tempFilePath, []byte(largeContent), 0644)
	if err != nil {
		t.Fatalf("failed to create temporary file: %v", err)
	}

	handler, _ := SetupHandlers() // Test the full handler chain

	t.Run("should apply Brotli compression when Accept-Encoding is br", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/temp_large_file.txt", nil)
		req.Header.Set("Accept-Encoding", "br")
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
		}

		contentEncoding := rr.Header().Get("Content-Encoding")
		if contentEncoding != "br" {
			t.Errorf("Content-Encoding header mismatch: got %q, want %q", contentEncoding, "br")
		}

		// Decompress the body to verify content
		brReader := brotli.NewReader(rr.Body)

		decompressedBody, err := ioutil.ReadAll(brReader)
		if err != nil {
			t.Fatalf("failed to decompress body with Brotli: %v", err)
		}

		if string(decompressedBody) != largeContent {
			t.Errorf("decompressed body mismatch: got %q, want %q", string(decompressedBody), largeContent)
		}
	})

	t.Run("should apply Gzip compression when Accept-Encoding is gzip (and no br)", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/temp_large_file.txt", nil)
		req.Header.Set("Accept-Encoding", "gzip") // Only gzip
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
		}

		contentEncoding := rr.Header().Get("Content-Encoding")
		if contentEncoding != "gzip" {
			t.Errorf("Content-Encoding header mismatch: got %q, want %q", contentEncoding, "gzip")
		}

		// Decompress the body to verify content
		gzipReader, err := gzip.NewReader(rr.Body)
		if err != nil {
			t.Fatalf("failed to create gzip reader: %v", err)
		}
		defer gzipReader.Close()

		decompressedBody, err := ioutil.ReadAll(gzipReader)
		if err != nil {
			t.Fatalf("failed to decompress body: %v", err)
		}

		if string(decompressedBody) != largeContent {
			t.Errorf("decompressed body mismatch: got %q, want %q", string(decompressedBody), largeContent)
		}
	})

	t.Run("should not apply compression when Accept-Encoding is not present", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/temp_large_file.txt", nil)
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
		}

		contentEncoding := rr.Header().Get("Content-Encoding")
		if contentEncoding != "" {
			t.Errorf("Content-Encoding header should be empty, got %q", contentEncoding)
		}

		if rr.Body.String() != largeContent {
			t.Errorf("body mismatch: got %q, want %q", rr.Body.String(), largeContent)
		}
	})
}