# Stage 1: Build the Go binary
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Copy Go module files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the Go source code
COPY *.go ./
COPY server/ ./server/

# Build the Go application with optimizations for size and static linking
# CGO_ENABLED=0: Disables CGO, ensuring a statically linked binary
# GOOS=linux: Specifies the target operating system
# -ldflags "-s -w": Strips debug information and symbol tables to reduce binary size
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags "-s -w" -o /go-react-spa-server .

# Stage 2: Create the final, minimal image
FROM alpine:latest

WORKDIR /app

# Copy the compiled Go binary from the builder stage
COPY --from=builder /go-react-spa-server .

# Expose port 8080, which is the default port for the Go server
EXPOSE 8080

# Command to run the executable
# The Go application will serve the static files from a configurable path (e.g., via STATIC_DIR env var)
CMD ["/app/go-react-spa-server"]