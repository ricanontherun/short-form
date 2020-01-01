#!/bin/bash

echo "Building Binaries..."

CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 go build -o dist/darwin/amd64/sf

