#!/bin/bash
# EnvKit GUI - 跨平台快速启动脚本

set -e

ROOT_DIR="$(cd "$(dirname "$0")" && pwd)"
cd "$ROOT_DIR"

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  EnvKit Desktop - 快速启动"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

OS="$(uname -s 2>/dev/null || echo unknown)"

find_built_app() {
  if [ -d "build/bin/envkit-gui.app" ]; then
    echo "build/bin/envkit-gui.app"
    return 0
  fi
  if [ -x "build/bin/envkit-gui" ]; then
    echo "build/bin/envkit-gui"
    return 0
  fi
  if [ -x "build/bin/envkit-gui.exe" ]; then
    echo "build/bin/envkit-gui.exe"
    return 0
  fi
  # 兼容其它输出名
  if [ -x "build/bin/EnvKit" ]; then
    echo "build/bin/EnvKit"
    return 0
  fi
  return 1
}

launch_app() {
  local app="$1"
  case "$OS" in
    Darwin)
      if [ -d "$app" ]; then
        open "$app"
      else
        "$app" &
      fi
      ;;
    Linux|*)
      if [ -d "$app" ]; then
        echo "错误: 在 Linux 上不能直接 open .app 目录: $app"
        return 1
      fi
      "$app" &
      ;;
  esac
}

if APP_PATH="$(find_built_app)"; then
  echo "✓ 找到已构建的应用: $APP_PATH"
  echo ""
  echo "启动 EnvKit..."
  launch_app "$APP_PATH"
  echo ""
  echo "✅ EnvKit 已启动（后台运行）"
else
  echo "⚠️  应用尚未构建"
  echo ""
  if [ -t 0 ]; then
    read -r -p "是否现在构建应用？(y/N): " confirm
  else
    confirm="n"
  fi

  if [[ "${confirm:-}" =~ ^[Yy]$ ]]; then
    echo ""
    echo "开始构建..."
    export PATH="${PATH}:$(go env GOPATH 2>/dev/null)/bin"
    if ! command -v wails >/dev/null 2>&1; then
      echo "错误: 未找到 wails 命令。请先安装: go install github.com/wailsapp/wails/v2/cmd/wails@latest"
      exit 1
    fi
    # 确保前端产物存在
    if [ ! -f frontend/dist/index.html ] || [ ! -d frontend/node_modules ]; then
      (cd frontend && npm install && npm run build) || true
    fi
    wails build

    if APP_PATH="$(find_built_app)"; then
      echo ""
      echo "✅ 构建完成，正在启动..."
      launch_app "$APP_PATH"
    else
      echo "构建完成，但未找到可执行文件，请检查 build/bin/"
      exit 1
    fi
  else
    echo ""
    echo "使用以下命令手动构建："
    echo "  wails build"
    echo ""
    echo "然后运行："
    case "$OS" in
      Darwin) echo "  open build/bin/envkit-gui.app  或  ./start-gui.sh" ;;
      *)      echo "  ./build/bin/envkit-gui  或  ./start-gui.sh" ;;
    esac
  fi
fi

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
