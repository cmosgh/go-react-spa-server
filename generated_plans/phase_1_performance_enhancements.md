# Phase 1: Core Performance Enhancements

**Objective:** Maximize static asset serving speed and efficiency.

## Feature: Enhanced Caching Headers (ETag/Last-Modified/Cache-Control)

- [x] **Description:** Refine the existing `CacheControlMiddleware` to ensure optimal ETag generation, Last-Modified header handling, and granular `Cache-Control` directives. This will leverage browser caching effectively, reducing server load and improving perceived performance for repeat visitors.
- [x] **Implementation Steps:**
    - [x] **Review Current `CacheControlMiddleware`:** Analyze `main.go` to understand the existing implementation of `CacheControlMiddleware`.
    - [x] **Implement ETag Generation:** For each static file served, generate a strong ETag (e.g., based on file content hash or modification time + size).
    - [x] **Handle `If-None-Match` and `If-Modified-Since`:** Implement logic to check incoming `If-None-Match` (for ETag) and `If-Modified-Since` (for Last-Modified) headers. If the content hasn't changed, respond with `304 Not Modified`.
    - [x] **Granular `Cache-Control`:**
        - [x] For hashed assets (e.g., `index-BofCMMuu.css`, `index-CtS_vSNO.js`), set `Cache-Control: public, max-age=31536000, immutable` (or a very long max-age) to ensure aggressive caching by browsers.
        - [x] For `index.html`, set `Cache-Control: no-cache, no-store, must-revalidate` to ensure it's always fetched fresh, as it's the entry point and might change frequently.
        - [x] For other static assets (e.g., `vite.svg`, `horse.svg` if not hashed), set a reasonable `max-age` (e.g., `Cache-Control: public, max-age=3600`).
    - [x] **Unit Tests:** Add comprehensive unit tests for the updated `CacheControlMiddleware` to cover various scenarios (first request, subsequent request with ETag/Last-Modified, different asset types).

## Feature: Brotli Compression Support

- [ ] **Description:** Integrate Brotli compression for text-based assets, offering superior compression ratios compared to Gzip, leading to smaller transfer sizes and faster load times for supported browsers.
- [ ] **Implementation Steps:**
    - [ ] **Research Go Brotli Libraries:** Identify a robust and performant Go library for Brotli compression (e.g., `github.com/andybalholm/brotli`).
    - [ ] **Integrate into Middleware:** Create a new middleware (or extend the existing `GzipCompression` middleware) to check for `Accept-Encoding: br` header.
    - [ ] **Conditional Compression:** If Brotli is supported by the client and the asset type is compressible (HTML, CSS, JS, SVG, JSON), compress the response with Brotli. Fallback to Gzip if Brotli is not supported.
    - [ ] **Content-Encoding Header:** Set `Content-Encoding: br` for Brotli compressed responses.
    - [ ] **Unit Tests:** Add unit tests to verify correct Brotli compression and fallback behavior.

## Feature: In-Memory Caching for Critical Assets

- [ ] **Description:** Implement a simple in-memory cache for frequently accessed, small, and non-changing critical assets like `index.html` (after it's built) and potentially `vite.svg`. This reduces disk I/O and speeds up initial requests.
- [ ] **Implementation Steps:**
    - [ ] **Choose Caching Mechanism:** Use a simple `map` or a small, thread-safe cache library in Go.
    - [ ] **Load Assets into Cache:** During server startup, read `index.html` and other specified critical assets into memory.
    - [ ] **Serve from Cache:** When a request comes for a cached asset, serve it directly from memory.
    - [ ] **Cache Invalidation (Development):** Consider a mechanism to invalidate the cache in development mode if the underlying files change (e.g., by watching file changes, though this might add complexity). For production, the cache can be loaded once at startup.
    - [ ] **Unit Tests:** Test the in-memory caching logic.
