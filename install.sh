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

    # 临时方案：从本地复制、从 GitHub 下载或从源码编译
    if [ -f "dist/${BINARY_NAME}" ]; then
        cp "dist/${BINARY_NAME}" "${INSTALL_DIR}/envkit"
        chmod +x "${INSTALL_DIR}/envkit"
        echo "✅ EnvKit 已从本地复制并成功安装到: ${INSTALL_DIR}/envkit"
    else
        echo "📦 正在尝试从 GitHub Releases 下载预编译的二进制文件..."
        if [ "$VERSION" = "latest" ]; then
            URL="https://github.com/${REPO}/releases/latest/download/${BINARY_NAME}"
        else
            URL="https://github.com/${REPO}/releases/download/${VERSION}/${BINARY_NAME}"
        fi

        if curl -fsSL -o "${INSTALL_DIR}/envkit" "$URL"; then
            chmod +x "${INSTALL_DIR}/envkit"
            echo "✅ EnvKit 已成功从 GitHub Releases 下载并安装到: ${INSTALL_DIR}/envkit"
        elif command -v go >/dev/null 2>&1; then
            echo "⚠️  从 GitHub 下载失败，检测到本地已安装 Go，正在尝试从源码编译..."
            if go build -o "${INSTALL_DIR}/envkit" ./cmd/envkit/main.go; then
                chmod +x "${INSTALL_DIR}/envkit"
                echo "✅ EnvKit 已从源码成功编译并安装到: ${INSTALL_DIR}/envkit"
            else
                echo "❌ 从源码编译 EnvKit 失败"
                exit 1
            fi
        else
            echo "❌ 无法安装 EnvKit。"
            echo "原因: 从 GitHub 下载失败 (链接: $URL) 且本地未安装 Go 编译器。"
            echo "请确认："
            echo "  1. 你的网络能够正常访问 GitHub。"
            echo "  2. 你已经在 GitHub 上发布了项目的 Release 版本并上传了二进制包 (仓库: $REPO)。"
            exit 1
        fi
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
