#!/bin/bash
# EnvKit 安装脚本：本地 dist → 多源下载 → 源码编译

set -euo pipefail

REPO="${ENVKIT_REPO:-EvilJul/envkit}"
VERSION="${ENVKIT_VERSION:-latest}"
INSTALL_DIR="/usr/local/bin"
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CONNECT_TIMEOUT="${ENVKIT_CONNECT_TIMEOUT:-30}"
MAX_TIME="${ENVKIT_MAX_TIME:-300}"
# 可选：自定义镜像前缀，例如 https://ghfast.top/ 或 https://mirror.ghproxy.com/
# 会对完整 GitHub URL 做前缀代理
CUSTOM_MIRROR="${ENVKIT_MIRROR:-}"

detect_platform() {
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)

    case "$OS" in
        linux*) OS="linux" ;;
        darwin*) OS="darwin" ;;
        msys*|mingw*|cygwin*) OS="windows" ;;
        *)
            echo "❌ 不支持的操作系统: $OS"
            exit 1
            ;;
    esac

    case "$ARCH" in
        x86_64|amd64) ARCH="amd64" ;;
        aarch64|arm64) ARCH="arm64" ;;
        *)
            echo "❌ 不支持的架构: $ARCH"
            exit 1
            ;;
    esac

    echo "检测到系统: $OS-$ARCH"
}

resolve_install_dir() {
    if [ "$OS" = "windows" ]; then
        INSTALL_DIR="$HOME/bin"
    elif [ -w "/usr/local/bin" ] 2>/dev/null; then
        INSTALL_DIR="/usr/local/bin"
    else
        INSTALL_DIR="$HOME/.local/bin"
    fi
    mkdir -p "$INSTALL_DIR"
}

binary_name() {
    local name="envkit-${OS}-${ARCH}"
    if [ "$OS" = "windows" ]; then
        name="${name}.exe"
    fi
    echo "$name"
}

release_url() {
    local name
    name="$(binary_name)"
    if [ "$VERSION" = "latest" ]; then
        echo "https://github.com/${REPO}/releases/latest/download/${name}"
    else
        echo "https://github.com/${REPO}/releases/download/${VERSION}/${name}"
    fi
}

# 校验下载结果：体积 + 可执行/格式，避免镜像返回 HTML 错误页仍当成功
is_valid_binary() {
    local path="$1"
    [ -f "$path" ] || return 1
    local size
    size=$(wc -c <"$path" | tr -d ' ')
    # CLI 压缩后约 6MB+，过小基本是错误页
    if [ "$size" -lt 1000000 ]; then
        return 1
    fi
    if command -v file >/dev/null 2>&1; then
        case "$(file -b "$path" 2>/dev/null || true)" in
            *ELF*|*Mach-O*|*PE32*|*executable*|*shared\ object*) return 0 ;;
            *) return 1 ;;
        esac
    fi
    # 无 file 命令时：ELF 魔数 \x7fELF 或 PE MZ
    local magic
    magic=$(head -c 4 "$path" | od -An -tx1 | tr -d ' \n')
    case "$magic" in
        7f454c46|4d5a*) return 0 ;; # ELF / MZ
    esac
    # Mach-O 常见魔数
    magic=$(head -c 4 "$path" | od -An -tx1 | tr -d ' \n')
    case "$magic" in
        cffaedfe|feedfacf|cefaedfe|feedface) return 0 ;;
    esac
    return 1
}

download_one() {
    local url="$1"
    local dest="$2"
    echo "  → $url"
    rm -f "$dest"
    if curl -fL --retry 3 --retry-delay 2 --retry-all-errors \
        --connect-timeout "$CONNECT_TIMEOUT" --max-time "$MAX_TIME" \
        -o "$dest" "$url" 2>/tmp/envkit-curl.err; then
        if is_valid_binary "$dest"; then
            return 0
        fi
        echo "    下载内容无效（可能是镜像错误页），跳过"
    else
        local err
        err=$(tail -n 1 /tmp/envkit-curl.err 2>/dev/null || true)
        echo "    失败: ${err:-curl error}"
    fi
    rm -f "$dest"
    return 1
}

# 构建候选 URL 列表：直连 + 自定义镜像 + 多个公共加速前缀
candidate_urls() {
    local base
    base="$(release_url)"
    echo "$base"
    if [ -n "$CUSTOM_MIRROR" ]; then
        # 允许用户写带/或不带/的前缀
        echo "${CUSTOM_MIRROR%/}/${base}"
    fi
    # 国内常用 GitHub 加速（前缀代理整链 URL）
    local prefixes=(
        "https://ghfast.top/"
        "https://gh-proxy.com/"
        "https://ghproxy.net/"
        "https://mirror.ghproxy.com/"
        "https://gitdl.cn/"
    )
    local p
    for p in "${prefixes[@]}"; do
        echo "${p}${base}"
    done
}

download_binary() {
    local dest="$1"
    local url
    echo "📦 从 GitHub Releases 下载 $(binary_name) ..."
    while IFS= read -r url; do
        [ -n "$url" ] || continue
        if download_one "$url" "$dest"; then
            echo "✅ 下载成功: $url"
            return 0
        fi
    done < <(candidate_urls)
    return 1
}

build_from_source() {
    local dest="$1"
    if ! command -v go >/dev/null 2>&1; then
        return 1
    fi
    if [ ! -f "${PROJECT_ROOT}/cmd/envkit/main.go" ]; then
        return 1
    fi
    local ver short
    short="$(cd "${PROJECT_ROOT}" && git rev-parse --short HEAD 2>/dev/null || echo "dev")"
    ver="0.2.0+${short}"
    echo "📦 使用本地 Go 从源码编译 (version=${ver})..."
    # Version 注入 internal/cli，便于确认装的是哪次提交
    (cd "${PROJECT_ROOT}" && CGO_ENABLED=0 go build -trimpath \
        -ldflags="-s -w -X github.com/fusheng/envkit/internal/cli.Version=${ver}" \
        -o "$dest" ./cmd/envkit)
}

has_source_tree() {
    [ -f "${PROJECT_ROOT}/cmd/envkit/main.go" ] && [ -f "${PROJECT_ROOT}/go.mod" ]
}

install() {
    local name dest tmp
    name="$(binary_name)"
    dest="${INSTALL_DIR}/envkit"
    if [ "$OS" = "windows" ]; then
        dest="${INSTALL_DIR}/envkit.exe"
    fi
    tmp="${dest}.tmp.$$"

    echo "📦 正在安装 EnvKit..."
    echo "   目标路径: $dest"

    # 优先级说明：
    # - ENVKIT_USE_RELEASE=1：强制走 dist / GitHub Release（CI/无 Go 环境）
    # - 默认：若本机有 Go 且仓库是源码树 → 优先从源码编译（避免 git pull 后仍装上旧 Release）
    # - 否则：dist → 下载 Release → 源码兜底
    local installed=0

    if [ "${ENVKIT_USE_RELEASE:-}" = "1" ]; then
        echo "   模式: ENVKIT_USE_RELEASE=1（优先 dist / GitHub Release）"
        if [ -f "${PROJECT_ROOT}/dist/${name}" ]; then
            cp "${PROJECT_ROOT}/dist/${name}" "$dest"
            chmod +x "$dest"
            echo "✅ 已从本地 dist 安装: $dest"
            installed=1
        elif download_binary "$tmp"; then
            mv -f "$tmp" "$dest"
            chmod +x "$dest"
            echo "✅ 已从 GitHub Release 安装: $dest"
            installed=1
        elif build_from_source "$dest"; then
            chmod +x "$dest"
            echo "✅ 已从源码编译安装: $dest"
            installed=1
        fi
    else
        if command -v go >/dev/null 2>&1 && has_source_tree; then
            echo "   模式: 源码优先（检测到 Go + 本地仓库）"
            if build_from_source "$dest"; then
                chmod +x "$dest"
                echo "✅ 已从源码编译安装: $dest"
                installed=1
            fi
        fi
        if [ "$installed" -eq 0 ] && [ -f "${PROJECT_ROOT}/dist/${name}" ]; then
            cp "${PROJECT_ROOT}/dist/${name}" "$dest"
            chmod +x "$dest"
            echo "✅ 已从本地 dist 安装: $dest"
            installed=1
        fi
        if [ "$installed" -eq 0 ] && download_binary "$tmp"; then
            mv -f "$tmp" "$dest"
            chmod +x "$dest"
            echo "✅ 已从 GitHub Release 安装: $dest"
            installed=1
        fi
        if [ "$installed" -eq 0 ] && build_from_source "$dest"; then
            chmod +x "$dest"
            echo "✅ 已从源码编译安装: $dest"
            installed=1
        fi
    fi

    if [ "$installed" -eq 0 ]; then
        rm -f "$tmp"
        echo "❌ 安装失败"
        echo "原因: 网络下载均失败，且本机无可用 Go 编译器 / 源码树。"
        echo "可尝试:"
        echo "  1. 设置镜像: export ENVKIT_MIRROR=https://ghfast.top/"
        echo "  2. 放宽超时: export ENVKIT_CONNECT_TIMEOUT=60 ENVKIT_MAX_TIME=600"
        echo "  3. 安装 Go 后重试: https://go.dev/dl/"
        echo "  4. 手动编译: go build -o envkit ./cmd/envkit && sudo cp envkit /usr/local/bin/"
        echo "  5. 手动下载: $(release_url)"
        exit 1
    fi

    if ! echo "$PATH" | tr ':' '\n' | grep -qx "$INSTALL_DIR"; then
        echo ""
        echo "⚠️  $INSTALL_DIR 不在 PATH 中，请加入 shell 配置:"
        echo "  export PATH=\"\$PATH:$INSTALL_DIR\""
    fi

    echo ""
    echo "🎉 安装完成"
    echo "   二进制: $dest"
    if command -v "$dest" >/dev/null 2>&1 || [ -x "$dest" ]; then
        echo -n "   版本: "
        "$dest" version 2>/dev/null || true
        echo -n "   uv 是否在 list 中: "
        if "$dest" list 2>/dev/null | grep -qE '[[:space:]]uv[[:space:]]'; then
            echo "是 ✓"
        else
            echo "否 ✗（请确认源码是否包含 uv 提交）"
        fi
    fi
    echo ""
}

main() {
    echo "🚀 EnvKit 安装程序"
    echo ""
    detect_platform
    resolve_install_dir
    install
}

main
