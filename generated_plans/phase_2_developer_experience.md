# Phase 2: Developer Experience & Configuration

**Objective:** Make the server easier to configure and use for developers.

## Feature: Flexible Static Directory Configuration

- [x] **Description:** Enhance the `STATIC_DIR` environment variable to allow for more flexible configuration, potentially supporting multiple static directories or a configuration file.
- [x] **Implementation Steps:**
    - [x] **Configuration File Support:** Introduce support for a simple configuration file (e.g., `config.json` or `config.yaml`) that can define the static directory path. Environment variables should still take precedence.
    - [x] **Update `main.go`:** Modify the server initialization to read from the new configuration source.
    - [x] **Documentation:** Clearly document the new configuration options in the README.md.

## Feature: Customizable SPA Fallback

- [x] **Description:** Allow users to specify a different fallback file than `index.html` (e.g., `app.html`) if their Vite application uses a non-standard entry point.
- [x] **Implementation Steps:**
    - [x] **Configuration Option:** Add a new configuration option (e.g., `SPA_FALLBACK_FILE` environment variable or in the config file) to specify the fallback HTML file name.
    - [x] **Update SPA Handler:** Modify the SPA handler logic to use the configured fallback file instead of hardcoding `index.html`.
    - [x] **Documentation:** Document this new configuration option in the `README.md`.

## Feature: Configurable Port via Environment Variable

- [x] **Description:** Allow the server's listening port to be configured via an environment variable (e.g., `PORT`). This simplifies deployment in containerized environments like Docker, where port mapping is common.
- [x] **Implementation Steps:**
    - [x] **Environment Variable Check:** Modify the server initialization in `main.go` (or `server.go`) to check for a `PORT` environment variable.
    - [x] **Default Port:** If the `PORT` environment variable is not set, the server should default to a standard port (e.g., `8080`).
    - [x] **Server Listen Address:** Use the configured port when setting up the server's listen address.
    - [x] **Documentation:** Document this new configuration option in the `README.md`, including how to use it with Docker.

## Feature: Optimized Docker Deployment

- [ ] **Description:** Create an optimized Docker deployment strategy for the Go-React SPA server, focusing on generating the smallest possible Docker image suitable for Kubernetes and other container orchestration platforms. This includes proper documentation for building and deploying the Docker image.
- [ ] **Implementation Steps:**
    - [ ] **Review Dockerfile:** Analyze the existing `Dockerfile` for potential optimizations (e.g., multi-stage builds, smaller base images, removing unnecessary files).
    - [ ] **Implement Multi-stage Build:** Refactor the `Dockerfile` to use multi-stage builds to separate build-time dependencies from runtime dependencies, significantly reducing the final image size.
    - [ ] **Choose Minimal Base Image:** Select a minimal base image for the final stage (e.g., `scratch` or `alpine`) to further reduce the image footprint.
    - [ ] **Optimize Go Build:** Ensure the Go application is built with optimizations for size and static linking within the Docker environment.
    - [ ] **Document Dockerfile:** Add comments to the `Dockerfile` explaining each step and its purpose.
    - [ ] **Create Deployment Documentation:** Develop comprehensive documentation (e.g., in `README.md` or a new `DEPLOYMENT.md`) detailing how to build the Docker image, push it to a registry, and deploy it to Kubernetes or other container platforms, including example YAML configurations.
    - [ ] **Verify Image Size:** Measure and report the size of the optimized Docker image.