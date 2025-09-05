# Phase 2, Feature 4: Optimized Docker Deployment - Technical Solution (Revised)

## Objective:
To create an optimized Docker deployment strategy for the Go server, focusing on generating the smallest possible Docker image suitable for Kubernetes and other container orchestration platforms. This image will *only* contain the Go backend. The Single Page Application (SPA) will be served from an external volume mounted at runtime, configured via an environment variable.

## Technical Steps:

1.  [ ] **Analyze Existing Dockerfile:**
    *   [ ] Examine the current `Dockerfile` located at `.Dockerfile`.
    *   [ ] Identify areas for improvement, such as:
        *   [ ] Base image selection.
        *   [ ] Unnecessary files being copied into the image.
        *   [ ] Lack of multi-stage build.
        *   [ ] **Crucially, remove any steps related to building or copying the React frontend.**

2.  [ ] **Implement Multi-stage Build (Go Backend Only):**
    *   [ ] **Build Stage:**
        *   [ ] Use a larger, build-friendly image (e.g., `golang:latest` or `golang:alpine`) for compiling the Go application.
        *   [ ] Copy `go.mod` and `go.sum` first, then `go get` dependencies to leverage Docker layer caching.
        *   [ ] Copy the Go source code and build the Go binary, ensuring it's statically linked (`CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .`).
    *   [ ] **Final (Runtime) Stage:**
        *   [ ] Use a minimal base image (e.g., `scratch` or `alpine`).
        *   [ ] Copy *only* the compiled Go binary from the build stage.
        *   [ ] Set the entry point to run the Go binary.

3.  [ ] **Optimize Go Build for Docker:**
    *   [ ] Ensure `CGO_ENABLED=0` is set during the Go build process to create a statically linked binary, which can then be run on a `scratch` or `alpine` image without external C dependencies.
    *   [ ] Use `go build -ldflags "-s -w"` to strip debug information and symbol tables, further reducing the binary size.

4.  [ ] **Update `.dockerignore`:**
    *   [ ] Add entries to `.dockerignore` to prevent unnecessary files (e.g., `node_modules` from the root, `.git`, `.idea`, `test-results`, `generated_plans`, and *all client-side code*) from being copied into the build context, speeding up Docker builds.

5.  [ ] **Document Dockerfile:**
    *   [ ] Add clear, concise comments within the `Dockerfile` explaining each stage, command, and the rationale behind specific optimizations.

6.  [ ] **Create Deployment Documentation:**
    *   [ ] Create or update `DEPLOYMENT.md` in the project root.
    *   [ ] **Building the Docker Image:** Provide step-by-step instructions on how to build the optimized Docker image (Go backend only).
    *   [ ] **Pushing to a Registry:** Explain how to tag the image and push it to a Docker registry.
    *   [ ] **Kubernetes Deployment Example:**
        *   [ ] Provide a basic Kubernetes Deployment YAML example.
        *   [ ] Include a Service YAML for exposing the application.
        *   [ ] **Crucially, demonstrate how to mount an external volume for the SPA static files and how to configure the Go server to use this volume (e.g., via `STATIC_DIR` environment variable).**
        *   [ ] Explain how to apply these YAMLs using `kubectl`.
        *   [ ] Mention considerations for production deployments.
    *   [ ] **Other Container Platforms:** Briefly mention deployment to other platforms.

7.  [ ] **Verify Image Size:**
    *   [ ] After implementing the changes, run `docker images` to compare the size of the new optimized image (Go backend only).
    *   [ ] Include the size comparison in the `DEPLOYMENT.md` to highlight the optimization benefits.
