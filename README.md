## Project Architecture

This is a Go-React SPA (Single Page Application) server that combines:

- **Backend**: Go HTTP server (`main.go`) that serves static files and handles SPA routing
- **Frontend**: React application built with Vite in the `client/` directory
- **Testing**: Go unit tests (in `server/` directory) and Playwright e2e tests (`client/e2e/`)

The Go server serves static files from `./client/dist` by default (configurable via `STATIC_DIR` env var) and falls back to `index.html` for client-side routes, implementing standard SPA behavior.

## Development Commands

### Backend (Go)
- `go run main.go` - Start the Go server (listens on :8080)
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
- Client assets are built to `client/dist/` and served from there

- React app uses React Router for client-side routing
- Static assets are served directly by Go's `http.FileServer`

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

### Docker
- Build: `docker build -t go-react-spa-server .`
- Run: `docker run -p 8080:8080 go-react-spa-server`

