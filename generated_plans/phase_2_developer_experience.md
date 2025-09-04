# Phase 2: Developer Experience & Configuration

**Objective:** Make the server easier to configure and use for developers.

## Feature: Flexible Static Directory Configuration

- [ ] **Description:** Enhance the `STATIC_DIR` environment variable to allow for more flexible configuration, potentially supporting multiple static directories or a configuration file.
- [ ] **Implementation Steps:**
    - [ ] **Configuration File Support:** Introduce support for a simple configuration file (e.g., `config.json` or `config.yaml`) that can define the static directory path. Environment variables should still take precedence.
    - [ ] **Update `main.go`:** Modify the server initialization to read from the new configuration source.
    - [ ] **Documentation:** Clearly document the new configuration options.

## Feature: Customizable SPA Fallback

- [ ] **Description:** Allow users to specify a different fallback file than `index.html` (e.g., `app.html`) if their Vite application uses a non-standard entry point.
- [ ] **Implementation Steps:**
    - [ ] **Configuration Option:** Add a new configuration option (e.g., `SPA_FALLBACK_FILE` environment variable or in the config file) to specify the fallback HTML file name.
    - [ ] **Update SPA Handler:** Modify the SPA handler logic to use the configured fallback file instead of hardcoding `index.html`.
    - [ ] **Documentation:** Document this new configuration option.
