#!/bin/bash

# Exit immediately if a command exits with a non-zero status.
set -e

# Build the client
echo "Building client..."
cd client
npm install
npm run build
cd ..

# Kill any process running on port 8080
echo "Checking for and killing processes on port 8080..."
PORT=8081
PID=$(lsof -t -i:$PORT || true)
if [ -n "$PID" ]; then
    echo "Killing process $PID on port $PORT"
    kill -9 $PID
    sleep 1 # Give it a moment to terminate
else
    echo "No process found on port $PORT"
fi

# Start the server in the background
echo "Starting server..."
go run main.go &
# Get the process ID of the server
SERVER_PID=$!

# Function to kill the server
cleanup() {
    echo "Stopping server..."
    # Use kill -0 to check if the process exists before trying to kill it
    if kill -0 $SERVER_PID > /dev/null 2>&1; then
        echo "Killing server process $SERVER_PID..."
        kill $SERVER_PID
        wait $SERVER_PID || true # Wait for the process to terminate, ignore errors if already dead
    else
        echo "Server process $SERVER_PID not found or already stopped."
    fi
}

# Trap exit signals to ensure the server is killed
trap cleanup EXIT

# Wait for the server to be ready
echo "Waiting for server to be ready on http://localhost:8081..."
for i in $(seq 1 10); do
    if curl -s http://localhost:8081 > /dev/null; then
        echo "Server is ready!"
        break
    else
        echo "Server not ready yet, waiting... ($i/10)"
        sleep 2
    fi
    if [ $i -eq 10 ]; then
        echo "Server did not become ready in time. Exiting."
        exit 1
    fi
done

# Run the e2e tests
echo "Running e2e tests..."
cd client
npx playwright install --with-deps
npx playwright test
