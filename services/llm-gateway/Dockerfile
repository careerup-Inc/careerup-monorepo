# --- Build Stage ---
    FROM golang:1.24-alpine AS builder

    WORKDIR /app
    
    # Install git for private repo access if needed
    # RUN apk add --no-cache git
    
    # Copy go.mod and go.sum files
    COPY go.mod go.sum ./
    # Download dependencies
    # Consider using go mod vendor if preferred
    RUN go mod download
    
    # Copy the entire service source code
    COPY . .
    
    # Build the application
    RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /app/server ./cmd/main.go
    
    # --- Final Stage ---
    FROM alpine:latest
    
    RUN apk --no-cache add ca-certificates
    
    WORKDIR /root/
    
    # Copy the pre-built binary file from the builder stage
    COPY --from=builder /app/server .
    # Copy .env file if needed at runtime (Alternatively, use Docker Compose env vars)
    # COPY .env .
    
    # Expose the default port the application listens on (defined in .env or default)
    EXPOSE 9090
    
    # Command to run the executable
    CMD ["./server"]
    