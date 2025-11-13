#!/bin/bash

# Build script for NCCL Test Web

set -e

echo "Building frontend..."
cd web
npm install
npm run build
cd ..

echo "Copying frontend dist to internal/web..."
rm -rf internal/web/dist
cp -r web/dist internal/web/

echo "Building Go binary..."
go build -o nccl-test-web cmd/server/main.go

echo "Build complete! Run with: ./nccl-test-web"
