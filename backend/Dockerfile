# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /build

# Install build dependencies for CGO (required by golang-migrate sqlite3 driver)
RUN apk add --no-cache build-base

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application with CGO enabled
RUN CGO_ENABLED=1 GOOS=linux go build -o server ./cmd/server

# Runtime stage
FROM flyio/litefs:0.5 AS litefs

FROM alpine:3.19

# Install runtime dependencies
RUN apk add --no-cache ca-certificates sqlite fuse3

# Copy LiteFS binary from official image
COPY --from=litefs /usr/local/bin/litefs /usr/local/bin/litefs

# Copy application binary
COPY --from=builder /build/server /app/server
COPY migrations /app/migrations

# Copy LiteFS configuration
COPY litefs.yml /etc/litefs.yml

# Create directories and user (but run as root for FUSE)
RUN mkdir -p /data/litefs /litefs && \
    addgroup -g 1000 app && \
    adduser -D -u 1000 -G app app && \
    chown -R app:app /app /data /litefs

EXPOSE 8080 20202

# LiteFS will run migrations and start the server (runs as root for FUSE)
ENTRYPOINT ["litefs", "mount"]
