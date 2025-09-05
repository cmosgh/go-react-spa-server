package server

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/andybalholm/brotli" // For Brotli compression
)

// cacheControlMiddleware sets appropriate Cache-Control headers for static assets.
func CacheControlMiddleware(config *Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// For assets with content hashes (e.g., in /assets/ or with specific extensions)
			// set a long cache duration and immutable.
			if strings.HasPrefix(r.URL.Path, "/assets/") ||
				strings.HasSuffix(r.URL.Path, ".js") ||
				strings.HasSuffix(r.URL.Path, ".css") ||
				strings.HasSuffix(r.URL.Path, ".png") ||
				strings.HasSuffix(r.URL.Path, ".jpg") ||
				strings.HasSuffix(r.URL.Path, ".jpeg") ||
				strings.HasSuffix(r.URL.Path, ".gif") ||
				strings.HasSuffix(r.URL.Path, ".svg") ||
				strings.HasSuffix(r.URL.Path, ".webp") {
				w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
			} else if r.URL.Path == "/" || r.URL.Path == "/"+config.SpaFallbackFile {
				// For index.html (or custom fallback), set no-cache to ensure fresh content on every visit
				w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
				w.Header().Set("Pragma", "no-cache")
				w.Header().Set("Expires", "0")
			} else {
				// Check if the requested path corresponds to an actual file in the static directory.
				// Only apply default cache control if it's a static file.
				filePath := filepath.Join(config.StaticDir, r.URL.Path)
				if _, err := os.Stat(filePath); err == nil { // File exists
					w.Header().Set("Cache-Control", "public, max-age=3600")
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

// brotliResponseWriter is a wrapper around http.ResponseWriter that compresses data with Brotli.
type brotliResponseWriter struct {
	http.ResponseWriter
	brotliWriter *brotli.Writer
	wroteHeader  bool
}

func (brw *brotliResponseWriter) Write(data []byte) (int, error) {
	if !brw.wroteHeader {
		brw.WriteHeader(http.StatusOK) // Ensure headers are written before first write
	}
	return brw.brotliWriter.Write(data)
}

func (brw *brotliResponseWriter) WriteHeader(statusCode int) {
	if brw.wroteHeader {
		return
	}
	brw.ResponseWriter.Header().Set("Content-Encoding", "br")
	brw.ResponseWriter.Header().Set("Vary", "Accept-Encoding")
	brw.ResponseWriter.WriteHeader(statusCode)
	brw.wroteHeader = true
}

// BrotliHandler compresses responses with Brotli if the client supports it.
func BrotliHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "br") {
			next.ServeHTTP(w, r)
			return
		}

		brWriter := brotli.NewWriter(w)
		defer brWriter.Close()

		brw := &brotliResponseWriter{ResponseWriter: w, brotliWriter: brWriter}
		next.ServeHTTP(brw, r)
	})
}

// CSPMiddleware sets the Content-Security-Policy header if configured.
func CSPMiddleware(config *Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if config.CSPHeader != "" {
				w.Header().Set("Content-Security-Policy", config.CSPHeader)
			}
			next.ServeHTTP(w, r)
		})
	}
}

// HSTSMiddleware sets the Strict-Transport-Security header if configured and the connection is HTTPS.
func HSTSMiddleware(config *Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Only set HSTS header if max-age is configured and the connection is HTTPS
			if config.HSTSMaxAge > 0 && r.TLS != nil {
				w.Header().Set("Strict-Transport-Security", fmt.Sprintf("max-age=%d; includeSubDomains", config.HSTSMaxAge))
			}
			next.ServeHTTP(w, r)
		})
	}
}

// SecurityHeadersMiddleware sets common security headers based on configuration.
func SecurityHeadersMiddleware(config *Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Set X-Content-Type-Options header
			if config.XContentTypeOptions != "" {
				w.Header().Set("X-Content-Type-Options", config.XContentTypeOptions)
			} else {
				w.Header().Set("X-Content-Type-Options", "nosniff") // Default
			}

			// Set X-Frame-Options header
			if config.XFrameOptions != "" {
				w.Header().Set("X-Frame-Options", config.XFrameOptions)
			} else {
				w.Header().Set("X-Frame-Options", "DENY") // Default
			}

			// Set Referrer-Policy header
			if config.ReferrerPolicy != "" {
				w.Header().Set("Referrer-Policy", config.ReferrerPolicy)
			} else {
				w.Header().Set("Referrer-Policy", "no-referrer-when-downgrade") // Default
			}

			// Set Permissions-Policy header
			if config.PermissionsPolicy != "" {
				w.Header().Set("Permissions-Policy", config.PermissionsPolicy)
			} else {
				// Default: disable geolocation, microphone, camera
				w.Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")
			}
			next.ServeHTTP(w, r)
		})
	}
}