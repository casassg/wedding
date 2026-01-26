# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application with CGO enabled
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags='-w -s -extldflags "-static"' -o server ./cmd/server

FROM alpine:3.19

# Install runtime dependencies
RUN apk add --no-cache ca-certificates sqlite

# Copy application binary and migrations
COPY --from=builder /build/server /app/server
COPY migrations/ddl.sql /app/migrations/ddl.sql

# Create directories and user (but run as root for FUSE)
RUN addgroup -g 1000 app && \
    adduser -D -u 1000 -G app app && \
    mkdir -p /data && \
    chown -R app:app /app /data

EXPOSE 8080 20202

WORKDIR /data

# Run database migrations using sqlite3
RUN sqlite3 /data/wedding.db < /app/migrations/ddl.sql

# Launch the server
CMD ["/app/server", "serve"]
