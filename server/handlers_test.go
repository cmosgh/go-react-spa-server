package server

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)



func TestSpaHandler_CustomFallbackFile(t *testing.T) {
	// Create a temporary directory for this test
	tempStaticDir, err := ioutil.TempDir("", "test_static_dir_custom_fallback")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempStaticDir) // Clean up after test

	// Create a custom fallback HTML file
	customFallbackHTMLPath := filepath.Join(tempStaticDir, "app.html")
	err = ioutil.WriteFile(customFallbackHTMLPath, []byte("<html><body>Custom App Fallback</body></html>"), 0644)
	if err != nil {
		t.Fatalf("failed to create custom fallback html: %v", err)
	}

	// Create a config for the handler with custom fallback
	cfg := &Config{
		StaticDir: tempStaticDir,
		SpaFallbackFile: "app.html",
	}

	// The handler to test
	handler := CreateSpaHandler(cfg)

	// Load critical assets into cache for this test handler
	// Clear cache first to ensure isolation
	for k := range inMemoryCache {
		delete(inMemoryCache, k)
	}

	err = LoadCriticalAssetsIntoCache(cfg.StaticDir)
	if err != nil {
		t.Fatalf("failed to load critical assets into cache for TestSpaHandler_CustomFallbackFile: %v", err)
	}

	// Test that a non-existent route falls back to app.html
	req := httptest.NewRequest("GET", "/non-existent-route", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	body := rr.Body.String()
	if !strings.Contains(body, "Custom App Fallback") {
		t.Errorf("body should contain the custom fallback content")
	}

	// Test that the root route also serves the custom fallback
	req = httptest.NewRequest("GET", "/", nil)
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code for root: got %v want %v", status, http.StatusOK)
	}

	body = rr.Body.String()
	if !strings.Contains(body, "Custom App Fallback") {
		t.Errorf("body should contain the custom fallback content for root")
	}
}

func TestSpaHandler_DefaultFallbackFile(t *testing.T) {
	// Create a temporary directory for this test
	tempStaticDir, err := ioutil.TempDir("", "test_static_dir_default_fallback")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempStaticDir) // Clean up after test

	// Create an index.html file
	indexHTMLPath := filepath.Join(tempStaticDir, "index.html")
	err = ioutil.WriteFile(indexHTMLPath, []byte("<html><body>Default Index Fallback</body></html>"), 0644)
	if err != nil {
		t.Fatalf("failed to create index.html: %v", err)
	}

	// Create a config for the handler with default fallback (index.html)
	cfg := &Config{
		StaticDir: tempStaticDir,
		SpaFallbackFile: "index.html", // Explicitly set to default for clarity
	}

	// The handler to test
	handler := CreateSpaHandler(cfg)

	// Clear cache first to ensure isolation
	for k := range inMemoryCache {
		delete(inMemoryCache, k)
	}

	err = LoadCriticalAssetsIntoCache(cfg.StaticDir)
	if err != nil {
		t.Fatalf("failed to load critical assets into cache for TestSpaHandler_DefaultFallbackFile: %v", err)
	}

	// Test that a non-existent route falls back to index.html
	req := httptest.NewRequest("GET", "/non-existent-route", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	body := rr.Body.String()
	if !strings.Contains(body, "Default Index Fallback") {
		t.Errorf("body should contain the default index fallback content")
	}

	// Test that the root route also serves the default fallback
	req = httptest.NewRequest("GET", "/", nil)
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code for root: got %v want %v", status, http.StatusOK)
	}

	body = rr.Body.String()
	if !strings.Contains(body, "Default Index Fallback") {
		t.Errorf("body should contain the default index fallback content for root")
	}
}