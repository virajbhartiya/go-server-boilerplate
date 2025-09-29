# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache gcc musl-dev

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o server ./cmd/api/main.go

# Final stage
FROM alpine:latest

WORKDIR /app

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata

# Copy binary from builder
COPY --from=builder /app/server .

# Copy migrations
COPY --from=builder /app/internal/database/migrations ./internal/database/migrations

# Expose port
EXPOSE 8080

# Run the application
CMD ["./server"]