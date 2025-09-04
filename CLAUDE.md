# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Architecture

This is a Go-React SPA (Single Page Application) server that combines:

- **Backend**: Go HTTP server (`main.go`) that serves static files and handles SPA routing
- **Frontend**: React application built with Vite in the `client/` directory
- **Testing**: Go unit tests (`main_test.go`) and Playwright e2e tests (`client/e2e/`)

The Go server serves static files from `./client/dist` by default (configurable via `STATIC_DIR` env var) and falls back to `index.html` for client-side routes, implementing standard SPA behavior.

## Development Commands

### Backend (Go)
- `go run main.go` - Start the Go server (listens on :8080)
- `go test` - Run Go unit tests
- `go mod tidy` - Update dependencies

### Frontend (React/Vite)
- `cd client && npm install` - Install frontend dependencies
- `cd client && npm run dev` - Start Vite dev server
- `cd client && npm run build` - Build production assets to `client/dist/`
- `cd client && npm run lint` - Run ESLint
- `cd client && npm run preview` - Preview production build

### End-to-End Testing
- `./run-e2e-tests.sh` - Full e2e test suite (builds client, starts server, runs Playwright tests)
- `cd client && npm run test:e2e` - Alternative e2e test command

### Docker
- Build: `docker build -t go-react-spa-server .`
- Run: `docker run -p 8080:8080 go-react-spa-server`

## Key Implementation Details

- The Go server uses a custom SPA handler that checks file existence before deciding whether to serve static files or fall back to `index.html`
- Client assets are built to `client/dist/` and served from there
- The e2e test script (`run-e2e-tests.sh`) handles the full build-test-cleanup lifecycle
- React app uses React Router for client-side routing
- Static assets are served directly by Go's `http.FileServer`

## Testing Strategy

- **Unit tests**: Go tests in `main_test.go` test the SPA handler logic
- **E2E tests**: Playwright tests in `client/e2e/app.spec.js` test full application behavior
- The test script automatically builds the client, starts the server, and runs Playwright tests