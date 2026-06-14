#!/bin/bash

# EnvKit GUI 项目初始化脚本
# 用途：快速搭建 Wails + Svelte 开发环境

set -e

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  EnvKit Desktop GUI - 项目初始化"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# 检查依赖
check_dependency() {
    if ! command -v $1 &> /dev/null; then
        echo "❌ 未找到 $1"
        echo "   请先安装: $2"
        exit 1
    else
        echo "✓ 检测到 $1"
    fi
}

echo "1️⃣  检查依赖..."
check_dependency "go" "https://go.dev/dl/"
check_dependency "node" "https://nodejs.org/"
check_dependency "npm" "https://nodejs.org/"

# 检查 Wails CLI
if ! command -v wails &> /dev/null; then
    echo ""
    echo "📦 安装 Wails CLI..."
    go install github.com/wailsapp/wails/v2/cmd/wails@latest

    # 添加 GOPATH/bin 到 PATH 提示
    echo ""
    echo "⚠️  请确保 \$GOPATH/bin 在你的 PATH 中"
    echo "   macOS/Linux: export PATH=\$PATH:\$(go env GOPATH)/bin"
    echo "   然后重新运行此脚本"
    exit 0
else
    echo "✓ 检测到 Wails CLI"
fi

# 运行 Wails 检查
echo ""
echo "2️⃣  运行 Wails 系统检查..."
wails doctor

echo ""
read -p "是否继续初始化 GUI 项目? (y/N): " confirm
if [[ ! $confirm =~ ^[Yy]$ ]]; then
    echo "已取消"
    exit 0
fi

# 项目根目录
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$PROJECT_ROOT"

# 初始化 Wails 项目
echo ""
echo "3️⃣  初始化 Wails 项目..."

# 创建 GUI 目录结构
mkdir -p cmd/gui
mkdir -p frontend/src/lib/components
mkdir -p frontend/src/lib/stores
mkdir -p frontend/src/routes

# 如果 wails.json 不存在，初始化项目
if [ ! -f "wails.json" ]; then
    echo "正在创建 wails.json..."
    cat > wails.json << 'EOF'
{
  "$schema": "https://wails.io/schemas/config.v2.json",
  "name": "envkit-gui",
  "outputfilename": "envkit-gui",
  "frontend:install": "npm install",
  "frontend:build": "npm run build",
  "frontend:dev:watcher": "npm run dev",
  "frontend:dev:serverUrl": "auto",
  "author": {
    "name": "EnvKit",
    "email": "dev@envkit.io"
  },
  "info": {
    "companyName": "EnvKit",
    "productName": "EnvKit",
    "productVersion": "0.2.0",
    "copyright": "Copyright © 2024",
    "comments": "轻量级跨平台开发环境管理工具"
  }
}
EOF
fi

# 创建前端 package.json
if [ ! -f "frontend/package.json" ]; then
    echo ""
    echo "4️⃣  初始化前端项目..."
    cat > frontend/package.json << 'EOF'
{
  "name": "envkit-gui-frontend",
  "version": "0.2.0",
  "type": "module",
  "scripts": {
    "dev": "vite",
    "build": "vite build",
    "preview": "vite preview"
  },
  "dependencies": {
    "svelte": "^4.2.8"
  },
  "devDependencies": {
    "@sveltejs/vite-plugin-svelte": "^3.0.1",
    "vite": "^5.0.11"
  }
}
EOF

    # 安装前端依赖
    cd frontend
    npm install
    cd ..
fi

# 创建 Vite 配置
if [ ! -f "frontend/vite.config.js" ]; then
    cat > frontend/vite.config.js << 'EOF'
import { defineConfig } from 'vite'
import { svelte } from '@sveltejs/vite-plugin-svelte'

export default defineConfig({
  plugins: [svelte()],
  server: {
    port: 34115
  }
})
EOF
fi

# 创建 Svelte 配置
if [ ! -f "frontend/svelte.config.js" ]; then
    cat > frontend/svelte.config.js << 'EOF'
import { vitePreprocess } from '@sveltejs/vite-plugin-svelte'

export default {
  preprocess: vitePreprocess()
}
EOF
fi

# 创建主 Go 文件
if [ ! -f "cmd/gui/main.go" ]; then
    echo ""
    echo "5️⃣  创建 GUI 入口文件..."
    cat > cmd/gui/main.go << 'EOF'
package main

import (
	"context"
	"embed"
	"fmt"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/mac"

	"github.com/fusheng/envkit/internal/detector"
)

//go:embed all:frontend/dist
var assets embed.FS

type App struct {
	ctx context.Context
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, EnvKit is ready!", name)
}

func (a *App) GetSystemInfo() map[string]string {
	sys := detector.DetectSystem()
	return map[string]string{
		"os":           string(sys.OS),
		"arch":         string(sys.Architecture),
		"distribution": sys.Distribution,
	}
}

func main() {
	app := NewApp()

	err := wails.Run(&options.App{
		Title:  "EnvKit",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 255, G: 255, B: 255, A: 1},
		OnStartup:        app.startup,
		Bind: []interface{}{
			app,
		},
		Mac: &mac.Options{
			TitleBar: mac.TitleBarHiddenInset(),
			About: &mac.AboutInfo{
				Title:   "EnvKit",
				Message: "轻量级跨平台开发环境管理工具\nv0.2.0",
			},
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
EOF
fi

# 创建基础前端文件
if [ ! -f "frontend/src/App.svelte" ]; then
    echo ""
    echo "6️⃣  创建前端基础文件..."

    cat > frontend/src/App.svelte << 'EOF'
<script>
  import { onMount } from 'svelte';
  import { Greet, GetSystemInfo } from '../wailsjs/go/main/App.js';

  let name = 'World';
  let greetMsg = '';
  let systemInfo = {};

  onMount(async () => {
    systemInfo = await GetSystemInfo();
  });

  async function greet() {
    greetMsg = await Greet(name);
  }
</script>

<main>
  <div class="container">
    <h1>EnvKit Desktop</h1>

    <div class="section">
      <h2>System Information</h2>
      <div class="info">
        <p><strong>OS:</strong> {systemInfo.os || 'Loading...'}</p>
        <p><strong>Architecture:</strong> {systemInfo.arch || 'Loading...'}</p>
        {#if systemInfo.distribution}
          <p><strong>Distribution:</strong> {systemInfo.distribution}</p>
        {/if}
      </div>
    </div>

    <div class="section">
      <h2>Test API</h2>
      <input type="text" bind:value={name} placeholder="Enter your name" />
      <button on:click={greet}>Greet</button>
      {#if greetMsg}
        <p class="result">{greetMsg}</p>
      {/if}
    </div>
  </div>
</main>

<style>
  :global(body) {
    margin: 0;
    padding: 0;
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', 'Roboto', sans-serif;
  }

  main {
    background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
    min-height: 100vh;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 20px;
  }

  .container {
    background: white;
    border-radius: 12px;
    padding: 40px;
    max-width: 600px;
    width: 100%;
    box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
  }

  h1 {
    margin: 0 0 30px 0;
    color: #1d1d1f;
    font-size: 32px;
  }

  h2 {
    color: #6e6e73;
    font-size: 18px;
    margin: 0 0 16px 0;
  }

  .section {
    margin-bottom: 30px;
  }

  .info {
    background: #f5f5f7;
    padding: 16px;
    border-radius: 8px;
  }

  .info p {
    margin: 8px 0;
    color: #1d1d1f;
  }

  input {
    width: 100%;
    padding: 12px;
    border: 1px solid #d1d1d6;
    border-radius: 6px;
    font-size: 14px;
    margin-bottom: 12px;
  }

  button {
    background: #007aff;
    color: white;
    padding: 12px 24px;
    border: none;
    border-radius: 6px;
    font-size: 14px;
    cursor: pointer;
    transition: opacity 0.2s;
  }

  button:hover {
    opacity: 0.8;
  }

  .result {
    margin-top: 16px;
    padding: 12px;
    background: #28a74520;
    color: #28a745;
    border-radius: 6px;
  }
</style>
EOF

    cat > frontend/src/main.js << 'EOF'
import App from './App.svelte'

const app = new App({
  target: document.getElementById('app'),
})

export default app
EOF

    cat > frontend/index.html << 'EOF'
<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <title>EnvKit</title>
  </head>
  <body>
    <div id="app"></div>
    <script type="module" src="/src/main.js"></script>
  </body>
</html>
EOF
fi

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "✅ 初始化完成！"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "📝 下一步："
echo ""
echo "  1. 启动开发服务器："
echo "     wails dev"
echo ""
echo "  2. 构建生产版本："
echo "     wails build"
echo ""
echo "  3. 查看文档："
echo "     docs/GUI_DESIGN.md          - 设计方案"
echo "     docs/GUI_IMPLEMENTATION.md  - 实现指南"
echo "     docs/GUI_VISUAL_SPEC.md     - 视觉规范"
echo ""
echo "🎉 开始构建你的桌面应用吧！"
echo ""
