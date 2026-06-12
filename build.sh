#!/usr/bin/env bash
set -e

APP=flux
DIR="./cmd/flux"

echo "Building linux amd64..."
GOOS=linux GOARCH=amd64 go build -o $APP-linux-amd64 $DIR

echo "Building linux arm64..."
GOOS=linux GOARCH=arm64 go build -o $APP-linux-arm64 $DIR

echo "Building darwin amd64..."
GOOS=darwin GOARCH=amd64 go build -o $APP-darwin-amd64 $DIR

echo "Building darwin arm64..."
GOOS=darwin GOARCH=arm64 go build -o $APP-darwin-arm64 $DIR

echo "Building windows amd64..."
GOOS=windows GOARCH=amd64 go build -o $APP-windows-amd64.exe $DIR

echo "Building windows arm64..."
GOOS=windows GOARCH=arm64 go build -o $APP-windows-arm64.exe $DIR

echo "Generating checksums..."
sha256sum $APP-linux-amd64 > $APP-linux-amd64.sha256
sha256sum $APP-linux-arm64 > $APP-linux-arm64.sha256
sha256sum $APP-darwin-amd64 > $APP-darwin-amd64.sha256
sha256sum $APP-darwin-arm64 > $APP-darwin-arm64.sha256
sha256sum $APP-windows-amd64.exe > $APP-windows-amd64.sha256
sha256sum $APP-windows-arm64.exe > $APP-windows-arm64.sha256

echo "Done."
