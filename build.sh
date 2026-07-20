#!/bin/bash
# 本地交叉编译全平台 CLI，产物命名与 install.sh / GitHub Release 一致

set -euo pipefail

APP_NAME="envkit"
# 优先：环境变量 VERSION → git tag → 默认 0.2.3
if [ -z "${VERSION:-}" ]; then
  if git describe --tags --exact-match 2>/dev/null | grep -q '^v'; then
    VERSION="$(git describe --tags --exact-match | sed 's/^v//')"
  else
    VERSION="0.2.3"
  fi
fi

echo "🔨 Building EnvKit CLI v${VERSION}..."

rm -rf dist/
mkdir -p dist/

# 与 .github/workflows/release.yml matrix 保持一致
platforms=(
  "linux/amd64"
  "linux/arm64"
  "darwin/amd64"
  "darwin/arm64"
  "windows/amd64"
  "windows/arm64"
)

for platform in "${platforms[@]}"; do
  IFS='/' read -r GOOS GOARCH <<< "${platform}"

  output_name="${APP_NAME}-${GOOS}-${GOARCH}"
  if [ "${GOOS}" = "windows" ]; then
    output_name="${output_name}.exe"
  fi

  echo "  → ${GOOS}/${GOARCH} → dist/${output_name}"
  CGO_ENABLED=0 GOOS="${GOOS}" GOARCH="${GOARCH}" go build -trimpath \
    -ldflags="-s -w -X github.com/fusheng/envkit/internal/cli.Version=${VERSION}" \
    -o "dist/${output_name}" \
    ./cmd/envkit
done

echo ""
echo "✅ Build complete! Binaries:"
ls -lh dist/
echo ""
echo "Expected install download names:"
echo "  envkit-linux-amd64 | envkit-linux-arm64"
echo "  envkit-darwin-amd64 | envkit-darwin-arm64"
echo "  envkit-windows-amd64.exe | envkit-windows-arm64.exe"
