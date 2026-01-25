#!/bin/bash

set -e

ROOT_DIR="$(cd "$(dirname "$0")" && pwd)"

echo "Starting local wedding stack..."
echo "- Backend: http://localhost:8081"
echo "- Frontend: http://localhost:1313/"
echo "Press Ctrl+C to stop"
echo ""

cleanup() {
    if [ -n "$HUGO_PID" ]; then
        kill "$HUGO_PID" 2>/dev/null || true
    fi
    if [ -n "$BACKEND_PID" ]; then
        kill "$BACKEND_PID" 2>/dev/null || true
    fi
}

trap cleanup EXIT INT TERM

cd "$ROOT_DIR"

if [ -f "$ROOT_DIR/bin/activate-hermit" ]; then
    . "$ROOT_DIR/bin/activate-hermit"
fi

ALLOWED_ORIGINS="http://localhost:1313" "$ROOT_DIR/backend/dev.sh" &
BACKEND_PID=$!

echo "Starting Hugo server..."
hugo server &
HUGO_PID=$!

wait "$HUGO_PID"
