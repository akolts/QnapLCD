#!/bin/bash
# Build QnapLCD binaries using Docker.
# Produces: ./qnaplcd and ./qnaplcd-test (static Linux amd64 binaries).
set -e

IMAGE_NAME="qnaplcd-builder"
VERSION="${1:-$(git describe --tags --always --dirty 2>/dev/null || echo dev)}"

echo "Building version: $VERSION"
docker build --build-arg VERSION="$VERSION" -t "$IMAGE_NAME" .

echo "Extracting binaries..."
CONTAINER_ID=$(docker create "$IMAGE_NAME")
docker cp "$CONTAINER_ID:/qnaplcd" ./qnaplcd
docker cp "$CONTAINER_ID:/qnaplcd-test" ./qnaplcd-test
docker rm "$CONTAINER_ID" > /dev/null

echo ""
echo "Built successfully:"
ls -lh qnaplcd qnaplcd-test
