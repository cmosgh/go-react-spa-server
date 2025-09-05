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

func TestGetStaticDir(t *testing.T) {
	t.Run("STATIC_DIR is set", func(t *testing.T) {
		t.Setenv("STATIC_DIR", "/tmp/custom/dist")
		dir := GetStaticDir()
		assert.Equal(t, "/tmp/custom/dist", dir)
	})

	t.Run("STATIC_DIR is not set and no config file", func(t *testing.T) {
		t.Setenv("STATIC_DIR", "") // Ensure it's not set
		dir := GetStaticDir()
		assert.Equal(t, "./client/dist", dir)
	})

	t.Run("Config file exists, no env var", func(t *testing.T) {
		cleanup := func() {}
		_, cleanup = createTempConfigFile(t, `{"static_dir": "./config_static"}`)
		defer cleanup()

		t.Setenv("STATIC_DIR", "") // Ensure env var is not set
		dir := GetStaticDir()
		assert.Equal(t, "./config_static", dir)
	})

	t.Run("Env var takes precedence over config file", func(t *testing.T) {
		cleanup := func() {}
		_, cleanup = createTempConfigFile(t, `{"static_dir": "./config_static"}`)
		defer cleanup()

		t.Setenv("STATIC_DIR", "./env_static")
		dir := GetStaticDir()
		assert.Equal(t, "./env_static", dir)
	})

	t.Run("Invalid config file, no env var (falls back to default)", func(t *testing.T) {
		cleanup := func() {}
		_, cleanup = createTempConfigFile(t, `{"static_dir": "./config_static"`) // Invalid JSON
		defer cleanup()

		t.Setenv("STATIC_DIR", "") // Ensure env var is not set
		dir := GetStaticDir()
		assert.Equal(t, "./client/dist", dir)
	})

	t.Run("Invalid config file, env var set (env var takes precedence)", func(t *testing.T) {
		cleanup := func() {}
		_, cleanup = createTempConfigFile(t, `{"static_dir": "./config_static"`) // Invalid JSON
		defer cleanup()

		t.Setenv("STATIC_DIR", "./env_static")
		dir := GetStaticDir()
		assert.Equal(t, "./env_static", dir)
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
		assert.Error(t, err)
	})
}

func TestSetupHandlers(t *testing.T) {
	// Create a temporary directory for this test
	tempStaticDir, err := ioutil.TempDir("", "test_static_dir_setup_handlers")
	assert.NoError(t, err)
	defer os.RemoveAll(tempStaticDir) // Clean up after test

	// Temporarily set STATIC_DIR to the temporary directory for testing
	t.Setenv("STATIC_DIR", tempStaticDir)

	// Create a temporary large file for testing gzip
	largeContent := strings.Repeat("a", 2000) // Content larger than typical gzip threshold
	tempFilePath := filepath.Join(tempStaticDir, "temp_large_file.txt")
	err = ioutil.WriteFile(tempFilePath, []byte(largeContent), 0644)
	assert.NoError(t, err)

	handler := SetupHandlers()

	t.Run("should apply cache control headers for assets", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/assets/some.js", nil)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		cacheControl := rr.Header().Get("Cache-Control")
		assert.Equal(t, "public, max-age=31536000, immutable", cacheControl)
	})

	t.Run("should apply no-cache for index.html", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/index.html", nil)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		cacheControl := rr.Header().Get("Cache-Control")
		assert.Equal(t, "no-cache, no-store, must-revalidate", cacheControl)
		assert.Equal(t, "no-cache", rr.Header().Get("Pragma"))
		assert.Equal(t, "0", rr.Header().Get("Expires"))
	})

	t.Run("should gzip content when Accept-Encoding is gzip", func(t *testing.T) {
		// Create a temporary large file for testing gzip
		largeContent := strings.Repeat("a", 2000) // Content larger than typical gzip threshold
		tempFilePath := filepath.Join(tempStaticDir, "temp_large_file.txt")
		err = ioutil.WriteFile(tempFilePath, []byte(largeContent), 0644)
		assert.NoError(t, err)

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