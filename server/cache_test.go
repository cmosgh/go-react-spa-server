package server

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func init() {
	// Dummy usage to prevent "imported and not used" errors
	_ = time.Now()
}

func TestInMemoryCaching(t *testing.T) {
	// Clear cache before and after the entire test suite
	for k := range inMemoryCache {
		delete(inMemoryCache, k)
	}
	t.Cleanup(func() {
		for k := range inMemoryCache {
			delete(inMemoryCache, k)
		}
	})

	// Create a temporary directory for this test
	tempStaticDir, err := ioutil.TempDir("", "test_static_dir_cache")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempStaticDir) // Clean up after test

	cfg := &Config{
		StaticDir:       tempStaticDir,
		SpaFallbackFile: "index.html",
	}

	// Create dummy index.html and vite.svg in the temporary directory
	dummyIndexHTMLPath := filepath.Join(tempStaticDir, "index.html")
	err = ioutil.WriteFile(dummyIndexHTMLPath, []byte("<html><head><title>Test Index</title></head><body></body></html>"), 0644)
	if err != nil {
		t.Fatalf("failed to create dummy index.html: %v", err)
	}

	dummyViteSVGPath := filepath.Join(tempStaticDir, "vite.svg")
	err = ioutil.WriteFile(dummyViteSVGPath, []byte("<svg></svg>"), 0644)
	if err != nil {
		t.Fatalf("failed to create dummy vite.svg: %v", err)
	}

	t.Run("loadCriticalAssetsIntoCache loads assets correctly", func(t *testing.T) {
		// Clear cache before testing and ensure it's clean after
		for k := range inMemoryCache {
			delete(inMemoryCache, k)
		}
		defer func() {
			for k := range inMemoryCache {
				delete(inMemoryCache, k)
			}
		}()

		err := LoadCriticalAssetsIntoCache(cfg.StaticDir) // Pass temp dir
		if err != nil {
			t.Fatalf("loadCriticalAssetsIntoCache failed: %v", err)
		}

		if len(inMemoryCache) != 2 {
			t.Errorf("Expected 2 assets in cache, got %d", len(inMemoryCache))
		}

		// Verify index.html
		idxAsset, ok := inMemoryCache["/index.html"]
		if !ok {
			t.Error("/index.html not found in cache")
		} else {
			if !strings.Contains(string(idxAsset.Content), "<title>Test Index</title>") {
				t.Errorf("index.html content mismatch")
			}
			if idxAsset.MimeType != "text/html; charset=utf-8" {
				t.Errorf("index.html mime type mismatch: got %q", idxAsset.MimeType)
			}
			if idxAsset.Size == 0 {
				t.Errorf("index.html size is 0")
			}
			if idxAsset.ModTime.IsZero() {
				t.Errorf("index.html ModTime is zero")
			}
		}

		// Verify vite.svg
		viteAsset, ok := inMemoryCache["/vite.svg"]
		if !ok {
			t.Error("/vite.svg not found in cache")
		} else {
			if string(viteAsset.Content) != "<svg></svg>" {
				t.Errorf("vite.svg content mismatch")
			}
			if viteAsset.MimeType != "image/svg+xml" {
				t.Errorf("vite.svg mime type mismatch: got %q", viteAsset.MimeType)
			}
			if viteAsset.Size == 0 {
				t.Errorf("vite.svg size is 0")
			}
			if viteAsset.ModTime.IsZero() {
				t.Errorf("vite.svg ModTime is zero")
			}
		}
	})

	t.Run("loadCriticalAssetsIntoCache handles missing assets gracefully", func(t *testing.T) {
		// Clear cache before testing and ensure it's clean after
		for k := range inMemoryCache {
			delete(inMemoryCache, k)
		}
		defer func() {
			for k := range inMemoryCache {
				delete(inMemoryCache, k)
			}
		}()

		// Create a temporary directory for this test to ensure no actual files interfere
		tempDir, err := ioutil.TempDir("", "test_static_dir")
		if err != nil {
			t.Fatalf("failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tempDir)

		// Call loadCriticalAssetsIntoCache with a directory that has no critical assets
		err = LoadCriticalAssetsIntoCache(tempDir)
		if err != nil {
			t.Fatalf("loadCriticalAssetsIntoCache failed: %v", err)
		}

		if len(inMemoryCache) != 0 {
			t.Errorf("Expected 0 assets in cache, got %d", len(inMemoryCache))
		}
	})

	t.Run("createSpaHandler serves index.html from cache", func(t *testing.T) {
		// Ensure cache is populated
		LoadCriticalAssetsIntoCache(cfg.StaticDir)

		handler := CreateSpaHandler(cfg)
		req := httptest.NewRequest("GET", "/index.html", nil)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
		}
		if !strings.Contains(rr.Body.String(), "<title>Test Index</title>") {
			t.Errorf("index.html content mismatch when served from cache")
		}
		if rr.Header().Get("Content-Type") != "text/html; charset=utf-8" {
			t.Errorf("Content-Type mismatch for cached index.html: got %q", rr.Header().Get("Content-Type"))
		}
		if rr.Header().Get("ETag") == "" {
			t.Errorf("ETag header missing for cached index.html")
		}
		if rr.Header().Get("Last-Modified") == "" {
			t.Errorf("Last-Modified header missing for cached index.html")
		}
	})

	t.Run("createSpaHandler serves vite.svg from cache", func(t *testing.T) {
		// Ensure cache is populated
		LoadCriticalAssetsIntoCache(cfg.StaticDir)

		handler := CreateSpaHandler(cfg)
		req := httptest.NewRequest("GET", "/vite.svg", nil)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
		}
		if rr.Body.String() != "<svg></svg>" {
			t.Errorf("vite.svg content mismatch when served from cache")
		}
		if rr.Header().Get("Content-Type") != "image/svg+xml" {
			t.Errorf("Content-Type mismatch for cached vite.svg: got %q", rr.Header().Get("Content-Type"))
		}
		if rr.Header().Get("ETag") == "" {
			t.Errorf("ETag header missing for cached vite.svg")
		}
		if rr.Header().Get("Last-Modified") == "" {
			t.Errorf("Last-Modified header missing for cached vite.svg")
		}
	})

	t.Run("createSpaHandler returns 304 for cached index.html with ETag match", func(t *testing.T) {
		LoadCriticalAssetsIntoCache(cfg.StaticDir)
		handler := CreateSpaHandler(cfg)

		// First request to get ETag
		req1 := httptest.NewRequest("GET", "/index.html", nil)
		rr1 := httptest.NewRecorder()
		handler.ServeHTTP(rr1, req1)
		etag := rr1.Header().Get("ETag")
		if etag == "" {
			t.Fatal("ETag not found in first response")
		}

		// Second request with If-None-Match
		req2 := httptest.NewRequest("GET", "/index.html", nil)
		req2.Header.Set("If-None-Match", etag)
		rr2 := httptest.NewRecorder()
		handler.ServeHTTP(rr2, req2)

		if status := rr2.Code; status != http.StatusNotModified {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusNotModified)
		}
		if rr2.Body.Len() != 0 {
			t.Errorf("body should be empty for 304 response, got %d bytes", rr2.Body.Len())
		}
	})

	t.Run("createSpaHandler returns 304 for cached vite.svg with If-Modified-Since match", func(t *testing.T) {
		LoadCriticalAssetsIntoCache(cfg.StaticDir)
		handler := CreateSpaHandler(cfg)

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
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusNotModified)
		}
		if rr2.Body.Len() != 0 {
			t.Errorf("body should be empty for 304 response, got %d bytes", rr2.Body.Len())
		}
	})
}
