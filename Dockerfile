# Multi-stage Dockerfile optimized for Fly.io deployment
# Stage 1: Build the application
FROM golang:1.25 AS build

WORKDIR /go/src/boilerplate

# Copy go mod files first for better layer caching
COPY go.mod go.sum ./

# Download dependencies (cached layer if go.mod/go.sum unchanged)
RUN go mod download

# Copy source code
COPY . .

# Build the application as a statically linked binary
# This allows it to run on Alpine Linux without CGO dependencies
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -a -installsuffix cgo \
    -ldflags="-w -s" \
    -o app ./cmd/server

# Stage 2: Create minimal runtime image
FROM alpine:latest AS release

# Install dumb-init for proper signal handling and ca-certificates for HTTPS
RUN apk --no-cache add dumb-init ca-certificates

WORKDIR /app

# Copy only the binary from build stage (keeps image small)
COPY --from=build /go/src/boilerplate/app .

# Make binary executable
RUN chmod +x /app/app

# Expose port 3000 (default for the application)
EXPOSE 3000

# Use dumb-init to handle signals properly (important for graceful shutdown)
ENTRYPOINT ["/usr/bin/dumb-init", "--"]

# Run the application
CMD ["./app"]