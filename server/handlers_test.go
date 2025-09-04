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
	_ = os.DevNull
	_ = filepath.Separator
}

func TestSpaHandler(t *testing.T) {
	// Create a temporary directory for this test
	tempStaticDir, err := ioutil.TempDir("", "test_static_dir_handlers")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempStaticDir) // Clean up after test

	// Create dummy index.html in the temporary directory
	dummyIndexHTMLPath := filepath.Join(tempStaticDir, "index.html")
	err = ioutil.WriteFile(dummyIndexHTMLPath, []byte("<html><head><title>Vite + React</title></head><body><div id=\"root\"></div></body></html>"), 0644)
	if err != nil {
		t.Fatalf("failed to create dummy index.html: %v", err)
	}
	

	// Create dummy assets directory and a JS file
	dummyAssetsDir := filepath.Join(tempStaticDir, "assets")
	err = os.MkdirAll(dummyAssetsDir, 0755)
	if err != nil {
		t.Fatalf("failed to create dummy assets dir: %v", err)
	}
	dummyJSFilePath := filepath.Join(dummyAssetsDir, "test.js")
	err = ioutil.WriteFile(dummyJSFilePath, []byte("console.log('test');"), 0644)
	if err != nil {
		t.Fatalf("failed to create dummy JS file: %v", err)
	}

	// Create dummy vite.svg
	dummyViteSVGPath := filepath.Join(tempStaticDir, "vite.svg")
	err = ioutil.WriteFile(dummyViteSVGPath, []byte("<svg></svg>"), 0644)
	if err != nil {
		t.Fatalf("failed to create dummy vite.svg: %v", err)
	}

	// The handler to test
	handler := CreateSpaHandler(tempStaticDir)

	// Load critical assets into cache for this test handler
	// Clear cache first to ensure isolation
	for k := range inMemoryCache {
		delete(inMemoryCache, k)
	}

	err = LoadCriticalAssetsIntoCache(tempStaticDir)
	if err != nil {
		t.Fatalf("failed to load critical assets into cache for TestSpaHandler: %v", err)
	}

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
		if !strings.Contains(body, "<title>Vite + React</title>") {
			t.Errorf("body should contain the Vite + React title")
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
		// Use the dummy JS file created in the temporary directory
		req := httptest.NewRequest("GET", "/assets/test.js", nil)
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
