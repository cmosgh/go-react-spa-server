# Phase 2, Feature 2: Customizable SPA Fallback

## Objective
Enhance the server's flexibility by allowing customization of the Single Page Application (SPA) fallback file. This feature aims to improve developer experience and simplify deployment.

### Description
Allow users to specify a different fallback file than `index.html` (e.g., `app.html`) if their Vite application uses a non-standard entry point.

### Implementation Steps for a Senior GoLand Developer

1.  [x] **Configuration Option:**
    *   [x] **Design:** Introduce a new configuration option, preferably an environment variable (e.g., `SPA_FALLBACK_FILE`). Consider integrating this into the existing configuration loading mechanism if a configuration file is already in use or planned.
    *   [x] **Default Value:** The default value should remain `index.html` to maintain backward compatibility.
    *   [x] **Validation:** Implement basic validation to ensure the provided fallback file name is valid (e.g., not empty, doesn't contain path separators).

2.  [x] **Update SPA Handler:**
    *   [x] **Locate SPA Handler:** Identify the existing SPA handler logic responsible for serving `index.html` for non-existent routes. This is likely in `server/handlers.go` or `server/server.go`.
    *   [x] **Dynamic Fallback:** Modify the handler to use the configured `SPA_FALLBACK_FILE` instead of hardcoding `index.html`. This will involve reading the configuration value at server startup.
    *   [x] **Error Handling:** Ensure proper error handling if the specified fallback file does not exist in the static directory.

3.  [x] **Documentation:**
    *   [x] **README.md Update:** Clearly document the new `SPA_FALLBACK_FILE` environment variable in the `README.md`.
    *   [x] **Usage Examples:** Provide examples of how to set this environment variable and its impact on the SPA serving behavior.

## Technical Considerations

*   [x] **Modularity:** Ensure that the configuration reading and application logic for this feature are well-separated and modular.
*   [x] **Error Handling:** Implement comprehensive error handling for file operations, environment variable parsing, and server startup.
*   [x] **Logging:** Use appropriate logging to indicate which configuration values are being used and any issues encountered during configuration loading.
*   [x] **Testing:** Write unit tests for all new configuration parsing, default value handling, and the modified server behavior.