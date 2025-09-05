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
	"github.com/stretchr/testify/assert"
)

// createTempConfigFile creates a temporary config file for testing.
func createTempConfigFile(t *testing.T, content string) (string, func()) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, ".go-spa-server-config.json")
	err := os.WriteFile(configPath, []byte(content), 0644)
	assert.NoError(t, err)

	// Change to the temporary directory to simulate running from there
	originalDir, _ := os.Getwd()
	err = os.Chdir(tempDir)
	assert.NoError(t, err)

	return configPath, func() {
		os.Chdir(originalDir) // Change back to original directory
	}
}



func TestStartServer(t *testing.T) {
	// Create a dummy handler
	dummyHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	t.Run("server fails to start with invalid address", func(t *testing.T) {
		// Call StartServer with an invalid address that will cause ListenAndServe to fail immediately
		err := StartServer(":invalid_port", dummyHandler)
		assert.Error(t, err)
	})
}

func TestSetupHandlers(t *testing.T) {
	// Create a temporary directory for this test
	tempStaticDir, err := ioutil.TempDir("", "test_static_dir_setup_handlers")
	assert.NoError(t, err)
	defer os.RemoveAll(tempStaticDir) // Clean up after test

	// Create a temporary large file for testing gzip
	largeContent := strings.Repeat("a", 2000) // Content larger than typical gzip threshold
	tempFilePath := filepath.Join(tempStaticDir, "temp_large_file.txt")
	err = ioutil.WriteFile(tempFilePath, []byte(largeContent), 0644)
	assert.NoError(t, err)

	// Create dummy index.html and custom.html for fallback tests
	indexHTMLPath := filepath.Join(tempStaticDir, "index.html")
	assert.NoError(t, ioutil.WriteFile(indexHTMLPath, []byte("<html><body>Index HTML</body></html>"), 0644))
	customHTMLPath := filepath.Join(tempStaticDir, "custom.html")
	assert.NoError(t, ioutil.WriteFile(customHTMLPath, []byte("<html><body>Custom HTML</body></html>"), 0644))

	// Test with default SPA fallback
	t.Run("default SPA fallback", func(t *testing.T) {
		t.Setenv("STATIC_DIR", tempStaticDir)
		t.Setenv("SPA_FALLBACK_FILE", "") // Ensure default
		defer os.Unsetenv("STATIC_DIR")
		defer os.Unsetenv("SPA_FALLBACK_FILE")

		handler, cfg := SetupHandlers()
		assert.Equal(t, "index.html", cfg.SpaFallbackFile)

		// Test non-existent route falls back to index.html
		req := httptest.NewRequest("GET", "/non-existent-route", nil)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Contains(t, rr.Body.String(), "Index HTML")

		// Test index.html cache control
		req = httptest.NewRequest("GET", "/index.html", nil)
		rr = httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		assert.Equal(t, "no-cache, no-store, must-revalidate", rr.Header().Get("Cache-Control"))
	})

	// Test with custom SPA fallback
	t.Run("custom SPA fallback", func(t *testing.T) {
		t.Setenv("STATIC_DIR", tempStaticDir)
		t.Setenv("SPA_FALLBACK_FILE", "custom.html")
		defer os.Unsetenv("STATIC_DIR")
		defer os.Unsetenv("SPA_FALLBACK_FILE")

		handler, cfg := SetupHandlers()
		assert.Equal(t, "custom.html", cfg.SpaFallbackFile)

		// Test non-existent route falls back to custom.html
		req := httptest.NewRequest("GET", "/another-non-existent-route", nil)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Contains(t, rr.Body.String(), "Custom HTML")

		// Test custom.html cache control
		req = httptest.NewRequest("GET", "/custom.html", nil)
		rr = httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		assert.Equal(t, "no-cache, no-store, must-revalidate", rr.Header().Get("Cache-Control"))
	})

	t.Run("should apply cache control headers for assets", func(t *testing.T) {
		t.Setenv("STATIC_DIR", tempStaticDir)
		defer os.Unsetenv("STATIC_DIR")

		handler, _ := SetupHandlers()

		req := httptest.NewRequest("GET", "/assets/some.js", nil)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		cacheControl := rr.Header().Get("Cache-Control")
		assert.Equal(t, "public, max-age=31536000, immutable", cacheControl)
	})

	t.Run("should gzip content when Accept-Encoding is gzip", func(t *testing.T) {
		t.Setenv("STATIC_DIR", tempStaticDir)
		defer os.Unsetenv("STATIC_DIR")

		handler, _ := SetupHandlers()

		req := httptest.NewRequest("GET", "/temp_large_file.txt", nil) // Request the temporary file
		req.Header.Set("Accept-Encoding", "gzip")
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		contentEncoding := rr.Header().Get("Content-Encoding")
		assert.Equal(t, "gzip", contentEncoding)

		// Verify content is gzipped by attempting to decompress
		reader, err := gzip.NewReader(rr.Body)
		assert.NoError(t, err)
		defer reader.Close()
		decompressedBody, err := ioutil.ReadAll(reader)
		assert.NoError(t, err)

		assert.Equal(t, largeContent, string(decompressedBody))
	})
}