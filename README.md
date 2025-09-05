## Project Architecture

This is a Go SPA (Single Page Application) server that combines:

- **Backend**: Go HTTP server (`main.go`) that serves static files and handles SPA routing
- **Frontend**: React application built with Vite in the `client/` directory (served via external volume at runtime)

The Go server serves static files from a configurable directory (via `STATIC_DIR` env var) and falls back to `index.html` for client-side routes, implementing standard SPA behavior. The Docker image for this server *only* contains the Go backend; the React frontend is expected to be mounted as an external volume at runtime.

## Development Commands

### Backend (Go)
- `go run main.go` - Start the Go server (listens on :8081)
- `go test ./...` - Run Go unit tests (including subpackages)
- `go test -v -cover ./...` - Run Go unit tests with coverage
- `go mod tidy` - Update dependencies

### Frontend (React/Vite)
- `cd client && npm install` - Install frontend dependencies
- `cd client && npm run dev` - Start Vite dev server
- `cd client && npm run build` - Build production assets to `client/dist/`
- `cd client && npm run lint` - Run ESLint
- `cd client && npm run preview` - Preview production build



## Key Implementation Details

- The Go server uses a custom SPA handler that checks file existence before deciding whether to serve static files or fall back to `index.html`
- React app uses React Router for client-side routing
- Static assets are served directly by Go's `http.FileServer`

## Configuration

### Static Directory Configuration

The server serves static files from a configurable directory. This directory is where the mounted SPA static files are expected to reside. The lookup order for the static directory is as follows:

1.  **`STATIC_DIR` Environment Variable**: If set, this environment variable takes the highest precedence. This is particularly useful when mounting an external volume containing your SPA.
    ```bash
    STATIC_DIR=/path/to/your/static/files go run main.go
    ```

2.  **`.go-spa-server-config.json` File**: If a file named `.go-spa-server-config.json` exists in the current working directory, the `static_dir` field within this JSON file will be used.
    Example `.go-spa-server-config.json`:
    ```json
    {
      "static_dir": "./custom_static_build"
    }
    ```

3.  **Default Path**: If neither the environment variable nor the configuration file specifies a static directory, the server defaults to serving files from `./client/dist`.

### SPA Fallback File Configuration

The server serves Single Page Applications (SPAs) by falling back to a specific HTML file (e.g., `index.html`) for client-side routes. You can customize this fallback file.

1.  **`SPA_FALLBACK_FILE` Environment Variable**: If set, this environment variable specifies the name of the HTML file to use as the SPA fallback. This takes precedence over the configuration file.
    ```bash
    SPA_FALLBACK_FILE=app.html go run main.go
    ```

2.  **`.go-spa-server-config.json` File**: If a file named `.go-spa-server-config.json` exists, the `spa_fallback_file` field within this JSON file will be used.
    Example `.go-spa-server-config.json`:
    ```json
    {
      "spa_fallback_file": "app.html"
    }
    ```

3.  **Default File**: If neither the environment variable nor the configuration file specifies a fallback file, the server defaults to `index.html`.

### Configurable Port

The server's listening port can be configured via an environment variable or a configuration file.

1.  **`PORT` Environment Variable**: If set, this environment variable specifies the port on which the server will listen. This takes precedence over the configuration file.
    ```bash
    PORT=9000 go run main.go
    ```
    For Docker:
    ```bash
    docker run -p 80:8081 -e PORT=8081 go-spa-server
    ```

2.  **`.go-spa-server-config.json` File**: If a file named `.go-spa-server-config.json` exists, the `port` field within this JSON file will be used.
    Example `.go-spa-server-config.json` with custom port:
    ```json
    {
      "port": 9090
    }
    ```

3.  **Default Port**: If neither the environment variable nor the configuration file specifies a port, the server defaults to `8081`.

## Testing

This project includes both Go unit tests for the backend and Playwright end-to-end (e2e) tests for the full application.

### Running Go Unit Tests

To run the Go unit tests with code coverage, navigate to the project root and execute:

```bash
go test -v -cover ./...
```

This will run all Go tests and report the code coverage.

### Running End-to-End (e2e) Tests

The e2e tests are run using Playwright and are orchestrated by a shell script.

**Prerequisites:**
- Node.js and npm (for client dependencies and Playwright)
- Go (for the backend server)

**Steps to run e2e tests:**

1.  **Install client dependencies:**
    ```bash
    cd client
    npm install
    cd ..
    ```
2.  **Run the e2e test script:**
    ```bash
    ./run-e2e-tests.sh
    ```
    This script will:
    - Build the React client.
    - Start the Go server in the background.
    - Install Playwright browsers (if not already installed).
    - Execute the Playwright tests.
    - Clean up by stopping the Go server.

Alternatively, you can run the Playwright tests directly after building the client and starting the server manually:

1.  **Build client:**
    ```bash
    cd client
    npm install
    npm run build
    cd ..
    ```
2.  **Start server (in a separate terminal):**
    ```bash
    go run main.go
    ```
3.  **Run Playwright tests:**
    ```bash
    cd client
    npx playwright install --with-deps
    npm run test:e2e
    cd ..
    ```

## Optimized Docker Deployment

An optimized Docker deployment strategy has been implemented for the Go SPA server. The Docker image now *only* contains the Go backend, resulting in a significantly smaller image size (currently **14.6MB**).

The Single Page Application (SPA) is expected to be served from an external volume mounted at runtime, configured via the `STATIC_DIR` environment variable.

For detailed instructions on building, pushing, and deploying the optimized Docker image, including Kubernetes examples with volume mounts, refer to: [`DEPLOYMENT.md`](./DEPLOYMENT.md)

For a comprehensive end-to-end test plan to verify the Docker image with an attached SPA volume, refer to: [`generated_plans/phase_2_feat_4.1.md`](./generated_plans/phase_2_feat_4.1.md)

