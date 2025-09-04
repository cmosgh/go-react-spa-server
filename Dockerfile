# Stage 1: Build the Go binary
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Copy Go module files
COPY go.mod ./
# Tidy downloads dependencies
RUN go mod tidy

# Copy the source code
COPY *.go ./

# Build the Go app
RUN CGO_ENABLED=0 GOOS=linux go build -o /go-react-spa-server

# Stage 2: Create the final image
FROM alpine:latest

WORKDIR /app

# Copy the built binary from the builder stage
COPY --from=builder /go-react-spa-server .



# Expose port 8080
EXPOSE 8080

# Command to run the executable
CMD ["/app/go-react-spa-server"]