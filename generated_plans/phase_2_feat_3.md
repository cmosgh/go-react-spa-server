# Phase 2, Feature 3: Configurable Port via Environment Variable

## Objective
Enhance the server's flexibility by configuring the server's listening port via environment variables. This feature aims to improve developer experience and simplify deployment, especially in containerized environments.

### Description
Allow the server's listening port to be configured via an environment variable (e.g., `PORT`). This simplifies deployment in containerized environments like Docker, where port mapping is common.

### Implementation Steps for a Senior GoLand Developer

1.  [ ] **Environment Variable Check:**
    *   [ ] **Locate Server Initialization:** Identify the server initialization code, likely in `main.go` or `server/server.go`, where the server's listen address is defined.
    *   [ ] **Read Environment Variable:** Use `os.Getenv("PORT")` to read the value of the `PORT` environment variable.
    *   [ ] **Type Conversion:** Convert the string value from the environment variable to an integer. Handle potential errors during conversion (e.g., non-numeric input).

2.  [ ] **Default Port:**
    *   [ ] **Fallback Logic:** If the `PORT` environment variable is not set or is invalid, the server should default to a standard port (e.g., `8080`). This ensures the server can still start without explicit port configuration.

3.  [ ] **Server Listen Address:**
    *   [ ] **Dynamic Address:** Construct the server's listen address (e.g., `":" + port`) using the configured or default port.
    *   [ ] **Error Handling:** Implement robust error handling for `http.ListenAndServe` in case the chosen port is already in use or invalid.

4.  [ ] **Documentation:**
    *   [ ] **README.md Update:** Document the new `PORT` environment variable in the `README.md`.
    *   [ ] **Docker Examples:** Provide clear examples of how to use this environment variable when running the application in a Docker container (e.g., `docker run -p 80:8080 -e PORT=8080 ...`).

## Technical Considerations

*   [ ] **Modularity:** Ensure that the configuration reading and application logic for this feature are well-separated and modular.
*   [ ] **Error Handling:** Implement comprehensive error handling for file operations, environment variable parsing, and server startup.
*   [ ] **Logging:** Use appropriate logging to indicate which configuration values are being used and any issues encountered during configuration loading.
*   [ ] **Testing:** Write unit tests for all new configuration parsing, default value handling, and the modified server behavior.