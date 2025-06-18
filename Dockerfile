# -------- Build Stage --------
    FROM golang:1.22-alpine AS builder

    # Enable Go modules and disable CGO for static build
    ENV CGO_ENABLED=0 GOOS=linux
    
    # Set working directory inside container
    WORKDIR /app
    
    # Copy go.mod and go.sum first (for layer caching)
    COPY go.mod go.sum ./
    RUN go mod download
    
    # Copy the rest of the application code
    COPY . .

    # Setup cache mounts
    ENV GOCACHE=/root/.cache/go-build
    
    # Build the binary, stripping debugging tools from binary, caching build contents to docker cache
    RUN --mount=type=cache,target=/root/.cache/go-build \
    go build -ldflags="-s -w" -o exchange-rate-service ./cmd/server
    
    # -------- Run Stage --------
    FROM alpine:latest
    
    # Set working directory
    WORKDIR /app
    
    # Copy the built binary from builder stage
    COPY --from=builder /app/exchange-rate-service .
    
    # Expose app port 
    EXPOSE 8080
    
    # Run the app
    CMD ["./exchange-rate-service"]
    