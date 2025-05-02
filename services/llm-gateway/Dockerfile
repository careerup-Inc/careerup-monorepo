# --- Build Stage ---
FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

# Install Git so dependencies from VCS can be downloaded
RUN apk add --no-cache git

RUN go mod download && go mod tidy

COPY . .

# Build the application; ensure the output path matches the main package location
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /app/server ./cmd/main.go

# --- Final Stage ---
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the pre-built binary file from the previous stage
COPY --from=builder /app/server .

EXPOSE 9090

CMD ["./server"]