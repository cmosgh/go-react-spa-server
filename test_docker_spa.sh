#!/bin/bash

set -e

# --- Configuration ---
IMAGE_NAME="go-spa-server:latest"
CONTAINER_NAME="go-spa-test-container"
HOST_PORT="8081"
CONTAINER_PORT="8081"

# --- Cleanup previous runs ---
echo "--- Cleaning up any previous Docker container runs ---"
if docker ps -a --format '{{.Names}}' | grep -q "^${CONTAINER_NAME}\\\$"; then
  echo "Stopping and removing existing container: ${CONTAINER_NAME}"
  docker stop ${CONTAINER_NAME} > /dev/null
  docker rm ${CONTAINER_NAME} > /dev/null
else
  echo "No existing container ${CONTAINER_NAME} found."
fi

# --- Step 1: Prepare the Sample SPA ---
echo "--- Building the React SPA ---"
pushd client > /dev/null
npm install
npm run build
popd > /dev/null

SPA_DIST_PATH="$(pwd)/client/dist"
echo "SPA build output path: ${SPA_DIST_PATH}"

# --- Step 2: Run the Docker Container with Volume Mount ---
echo "--- Running Docker container ---"
docker run -d -p ${HOST_PORT}:${CONTAINER_PORT} \
  -v "${SPA_DIST_PATH}:/app/spa-static" \
  -e STATIC_DIR="/app/spa-static" \
  --name ${CONTAINER_NAME} ${IMAGE_NAME}

echo "Waiting for container to start..."
sleep 5 # Give the server some time to start

# --- Step 3: Verify SPA is Served ---
echo "--- Verifying SPA is served ---"

# Check the root path (index.html)
echo "Checking root path (index.html)..."
if curl -s http://localhost:${HOST_PORT}/ | grep -q "<div id=\"root\">"; then
  echo "SUCCESS: index.html served correctly."
else
  echo "FAILURE: index.html not served as expected."
  docker logs ${CONTAINER_NAME}
  docker stop ${CONTAINER_NAME} > /dev/null
  docker rm ${CONTAINER_NAME} > /dev/null
  exit 1
fi

# Check a static asset (assuming vite.svg exists in client/dist/)
echo "Checking static asset (vite.svg)..."
if curl -s http://localhost:${HOST_PORT}/vite.svg | grep -q "<svg"; then # Simple check for SVG content
  echo "SUCCESS: vite.svg served correctly."
else
  echo "FAILURE: vite.svg not served as expected."
  docker logs ${CONTAINER_NAME}
  docker stop ${CONTAINER_NAME} > /dev/null
  docker rm ${CONTAINER_NAME} > /dev/null
  exit 1
fi

# Check a non-existent path (SPA fallback)
echo "Checking non-existent path (SPA fallback)..."
if curl -s http://localhost:${HOST_PORT}/some-non-existent-route | grep -q "<div id=\"root\">"; then
  echo "SUCCESS: SPA fallback to index.html working."
else
  echo "FAILURE: SPA fallback not working as expected."
  docker logs ${CONTAINER_NAME}
  docker stop ${CONTAINER_NAME} > /dev/null
  docker rm ${CONTAINER_NAME} > /dev/null
  exit 1
fi

echo "--- All SPA serving tests passed! ---"

# --- Step 4: Clean Up ---
echo "--- Cleaning up Docker container ---"
docker stop ${CONTAINER_NAME} > /dev/null
docker rm ${CONTAINER_NAME} > /dev/null
echo "Cleanup complete."
