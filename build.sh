#!/bin/bash

set -e

VERSION="0.1.0"
APP_NAME="envkit"

echo "🔨 Building EnvKit v${VERSION}..."

# 清理旧的构建文件
rm -rf dist/
mkdir -p dist/

# 构建不同平台的二进制文件
platforms=(
    "linux/amd64"
    "linux/arm64"
    "darwin/amd64"
    "darwin/arm64"
    "windows/amd64"
)

for platform in "${platforms[@]}"; do
    IFS='/' read -r -a parts <<< "$platform"
    GOOS="${parts[0]}"
    GOARCH="${parts[1]}"

    output_name="${APP_NAME}-${GOOS}-${GOARCH}"

    if [ "$GOOS" = "windows" ]; then
        output_name+=".exe"
    fi

    echo "  Building for ${GOOS}/${GOARCH}..."
    GOOS=$GOOS GOARCH=$GOARCH go build -ldflags="-s -w" -o "dist/${output_name}" ./cmd/envkit
done

echo "✅ Build complete! Binaries are in dist/"
ls -lh dist/
