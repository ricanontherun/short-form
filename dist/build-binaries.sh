#!/bin/bash

echo "Building Binaries..."

echo "Building for Mac..."
CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 go build -o dist/darwin/sf
echo "Done"

echo "Building for Linux..."
CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o dist/linux/sf
echo "Done"

echo "Building for Windows..."
CGO_ENABLED=1 GOOS=windows GOARCH=amd64 go build -o dist/windows/sf
echo "Done"

