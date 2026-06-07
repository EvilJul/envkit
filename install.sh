#!/bin/bash
# EnvKit 安装脚本

set -e

REPO="fusheng/envkit"
VERSION="latest"
INSTALL_DIR="/usr/local/bin"

# 检测操作系统和架构
detect_platform() {
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)

    case "$OS" in
        linux*)
            OS="linux"
            ;;
        darwin*)
            OS="darwin"
            ;;
        msys*|mingw*|cygwin*)
            OS="windows"
            ;;
        *)
            echo "❌ 不支持的操作系统: $OS"
            exit 1
            ;;
    esac

    case "$ARCH" in
        x86_64|amd64)
            ARCH="amd64"
            ;;
        aarch64|arm64)
            ARCH="arm64"
            ;;
        *)
            echo "❌ 不支持的架构: $ARCH"
            exit 1
            ;;
    esac

    echo "检测到系统: $OS-$ARCH"
}

# 下载并安装
install() {
    BINARY_NAME="envkit-${OS}-${ARCH}"

    if [ "$OS" = "windows" ]; then
        BINARY_NAME="${BINARY_NAME}.exe"
        INSTALL_DIR="$HOME/bin"
    fi

    # 创建安装目录
    mkdir -p "$INSTALL_DIR"

    # 下载二进制文件（这里假设从本地dist目录复制，实际应该从GitHub下载）
    echo "📦 正在安装 EnvKit..."

    # TODO: 从 GitHub Releases 下载
    # URL="https://github.com/${REPO}/releases/download/${VERSION}/${BINARY_NAME}"
    # curl -L -o "${INSTALL_DIR}/envkit" "$URL"

    # 临时方案：从本地复制
    if [ -f "dist/${BINARY_NAME}" ]; then
        cp "dist/${BINARY_NAME}" "${INSTALL_DIR}/envkit"
        chmod +x "${INSTALL_DIR}/envkit"
        echo "✅ EnvKit 已安装到: ${INSTALL_DIR}/envkit"
    else
        echo "❌ 找不到二进制文件: dist/${BINARY_NAME}"
        exit 1
    fi

    # 检查是否在 PATH 中
    if ! echo "$PATH" | grep -q "$INSTALL_DIR"; then
        echo ""
        echo "⚠️  警告: $INSTALL_DIR 不在 PATH 中"
        echo "请将以下内容添加到你的 shell 配置文件中："
        echo ""
        echo "  export PATH=\"\$PATH:$INSTALL_DIR\""
        echo ""
    fi

    echo ""
    echo "🎉 安装完成！运行以下命令开始使用:"
    echo ""
    echo "  envkit init"
    echo ""
}

# 主流程
main() {
    echo "🚀 EnvKit 安装程序"
    echo ""

    detect_platform
    install
}

main
