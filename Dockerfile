FROM golang:1.20-alpine AS builder

WORKDIR /app

# Install dependencies
RUN apk add --no-cache ca-certificates git

# Copy go mod files first for better caching
COPY go.mod go.sum* ./
RUN go mod download

# Copy the source code
COPY . .

# Build the application with optimization flags
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o /app/untie-server \
    ./src/api

# Use distroless as minimal base image
FROM gcr.io/distroless/static-debian11

WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/untie-server /app/

# Copy templates and static files
COPY --from=builder /app/src/web/templates /app/web/templates
COPY --from=builder /app/src/web/static /app/web/static

# Set user to non-root
USER nonroot:nonroot

# Expose default port
EXPOSE 8080

# Set environment variable for release mode
ENV ENV=production
ENV PORT=8080

# Run the binary
ENTRYPOINT ["/app/untie-server"]
