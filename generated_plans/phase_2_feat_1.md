# Phase 2, Feature 1: Flexible Static Directory Configuration

## Objective
Enhance the `STATIC_DIR` environment variable handling to support more flexible static file serving, including the ability to specify a configuration file for defining the static directory path. This will improve developer experience by providing more robust configuration options.

## Implementation Steps for a Senior GoLand Developer

1.  [x] **Introduce Configuration File Support:**
    *   [x] **Design:** Decide on a configuration file format. Given the simplicity, `.go-spa-server-config.json` is a straightforward choice for this feature.
    *   [x] **Structure:** Define a simple JSON structure, e.g., `{"static_dir": "/path/to/static"}`.
    *   [x] **Parsing Logic:** Implement a function to read and parse this configuration file. Consider using Go's `encoding/json` package.
    *   [x] **Precedence:** Ensure that environment variables (`STATIC_DIR`) take precedence over values read from the configuration file. This allows for easy overrides in different deployment environments.

2.  [x] **Update `main.go` (or relevant server initialization logic):**
    *   [x] **Load Order:** Modify the server's startup sequence to first attempt to read the static directory from the configuration file.
    *   [x] **Environment Variable Check:** After attempting to load from the config file, check for the `STATIC_DIR` environment variable. If present, it should override the configuration file value.
    *   [x] **Default Fallback:** If neither is provided, maintain the existing default behavior (e.g., serving from a `client/dist` or `static` directory).
    *   [x] **Error Handling:** Implement robust error handling for file reading and parsing, providing clear log messages for developers.

3.  [ ] **Documentation:**
    *   [ ] **README.md Update:** Clearly document the new configuration options in the `README.md` file.
    *   [ ] **Examples:** Provide examples of how to use both the environment variable and the new configuration file.
    *   [ ] **Precedence Explanation:** Explicitly state the precedence order (environment variable > config file > default).

## Technical Considerations

*   [x] **Modularity:** Ensure the configuration loading logic is modular and testable.
*   [x] **Logging:** Use appropriate logging levels to inform developers about the configuration source being used (e.g., "Using static directory from environment variable," "Using static directory from config file").
*   [x] **Testing:** Write unit tests for the new configuration parsing and loading logic to ensure correctness and robustness.