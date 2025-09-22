# E2E Integration Test Runner
FROM golang:1.23-alpine AS builder

# Install dependencies
RUN apk add --no-cache git ca-certificates

# Set working directory
WORKDIR /app

# Copy go mod files
COPY runner/go.mod runner/go.sum ./
RUN go mod download

# Copy source code
COPY runner/ ./
COPY system.yaml /workspace/tests/e2e/system.yaml

# Build the test runner
RUN go build -o e2e-runner main.go

# Runtime image
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache ca-certificates curl jq

# Copy the test runner
COPY --from=builder /app/e2e-runner /usr/local/bin/
COPY --from=builder /workspace/tests/e2e/system.yaml /workspace/tests/e2e/

# Create results directory
RUN mkdir -p /workspace/tests/e2e/results

# Set working directory
WORKDIR /workspace/tests/e2e

# Run the E2E tests
CMD ["e2e-runner"]
