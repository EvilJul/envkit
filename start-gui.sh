#!/bin/bash

# EnvKit GUI - 快速启动脚本

set -e

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  🚀 EnvKit Desktop - 快速启动"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# 检查应用是否已构建
if [ -d "build/bin/envkit-gui.app" ]; then
    echo "✓ 找到已构建的应用"
    echo ""
    echo "启动 EnvKit..."
    open build/bin/envkit-gui.app
    echo ""
    echo "✅ EnvKit 已启动！"
else
    echo "⚠️  应用尚未构建"
    echo ""
    read -p "是否现在构建应用？(y/N): " confirm

    if [[ $confirm =~ ^[Yy]$ ]]; then
        echo ""
        echo "开始构建..."
        export PATH=$PATH:$(go env GOPATH)/bin
        wails build

        echo ""
        echo "✅ 构建完成！"
        echo ""
        echo "启动 EnvKit..."
        open build/bin/envkit-gui.app
    else
        echo ""
        echo "使用以下命令手动构建："
        echo "  wails build"
        echo ""
        echo "然后运行："
        echo "  open build/bin/envkit-gui.app"
    fi
fi

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
