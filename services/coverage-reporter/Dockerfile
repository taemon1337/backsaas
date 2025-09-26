# Multi-stage build for Coverage Reporter Service
FROM golang:1.25-alpine AS builder

# Install dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o coverage-reporter .

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/coverage-reporter .

# Copy web templates and static files
COPY --from=builder /app/web ./web

# Create directory for coverage data
RUN mkdir -p /tmp/coverage-data

# Expose port
EXPOSE 8090

# Set environment variables
ENV PORT=8090
ENV STORAGE_TYPE=file
ENV STORAGE_PATH=/tmp/coverage-data

# Run the binary
CMD ["./coverage-reporter"]
