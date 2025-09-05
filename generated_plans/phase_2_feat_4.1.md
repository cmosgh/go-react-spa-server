# Phase 2, Feature 4.1: Docker E2E Test Plan for External SPA Volume

## Objective:
Verify that the Go SPA server Docker image correctly serves a Single Page Application (SPA) from an externally mounted volume.

## Prerequisites:
- [ ] Docker installed and running.
- [ ] Node.js and npm installed (to build the sample SPA).
- [ ] The Go SPA server Docker image built (e.g., `go-spa-server:latest`).

## Test Steps:

### 1. Prepare a Sample SPA:
- [ ] Ensure you have a built SPA. For this test, we'll assume your existing `client` directory can be built.
- [ ] Navigate to the `client` directory: `cd client`
- [ ] Install dependencies: `npm install`
- [ ] Build the SPA: `npm run build`
- [ ] The built SPA will be in `client/dist/`. Note the absolute path to this directory (e.g., `/Users/moshu/development/pdfthumbnailpro/go-react-spa-server/client/dist`). Let's call this `SPA_DIST_PATH`.

### 2. Run the Docker Container with Volume Mount:
- [ ] Navigate back to the project root: `cd ..`
- [ ] Run the Docker container, mounting the `SPA_DIST_PATH` to `/app/spa-static` inside the container, and setting the `STATIC_DIR` environment variable to `/app/spa-static`.
  ```bash
  docker run -d -p 8080:8080 \
    -v /path/to/your/spa/dist:/app/spa-static \
    -e STATIC_DIR=/app/spa-static \
    --name go-spa-test-container go-spa-server:latest
  ```
  *   **Important:** Replace `/path/to/your/spa/dist` with the actual absolute path to your `client/dist` directory.

### 3. Verify SPA is Served:
- [ ] Wait a few seconds for the container to start.
- [ ] Use `curl` or a web browser to access the SPA.
- [ ] **Check the root path (index.html):**
  ```bash
  curl http://localhost:8080/
  ```
  *Expected:* The output should be the HTML content of your `index.html` from the mounted SPA.
- [ ] **Check a static asset (e.g., a CSS or JS file from your SPA build):**
  *   You'll need to know a specific file name from your `client/dist` (e.g., `index-BkDSiPRN.css` or `index-Djraj8qp.js` from your `static/assets` directory).
  ```bash
  curl http://localhost:8080/assets/index-BkDSiPRN.css # Replace with an actual asset path
  ```
  *Expected:* The output should be the content of that CSS or JS file.
- [ ] **Check a non-existent path (SPA fallback):**
  ```bash
  curl http://localhost:8080/some-non-existent-route
  ```
  *Expected:* The output should still be the HTML content of your `index.html` (due to SPA fallback).

### 4. Clean Up:
- [ ] Stop and remove the Docker container:
  ```bash
  docker stop go-spa-test-container
  docker rm go-spa-test-container
  ```
