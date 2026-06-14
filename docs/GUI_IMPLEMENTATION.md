# EnvKit 桌面客户端实现指南

## 快速开始

### 1. 安装 Wails CLI
```bash
# macOS/Linux
go install github.com/wailsapp/wails/v2/cmd/wails@latest

# 验证安装
wails doctor
```

### 2. 初始化项目
```bash
cd envkit
wails init -n envkit-gui -t svelte
```

### 3. 项目结构
```
envkit/
├── cmd/
│   ├── envkit/        # CLI 版本
│   └── gui/           # GUI 版本（新增）
│       └── main.go
├── internal/          # 共享代码
│   ├── installer/
│   ├── detector/
│   └── ...
├── frontend/          # Svelte 前端
│   ├── src/
│   │   ├── App.svelte
│   │   ├── lib/
│   │   │   ├── components/
│   │   │   │   ├── Sidebar.svelte
│   │   │   │   ├── LanguageCard.svelte
│   │   │   │   ├── ToolCard.svelte
│   │   │   │   └── DatabaseTable.svelte
│   │   │   └── stores/
│   │   │       └── app.js
│   │   └── routes/
│   │       ├── Languages.svelte
│   │       ├── Tools.svelte
│   │       ├── Database.svelte
│   │       └── Settings.svelte
│   ├── package.json
│   └── vite.config.js
└── wails.json
```

---

## 核心代码实现

### 1. Go 后端（cmd/gui/main.go）

```go
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
	"github.com/fusheng/envkit/internal/installer"
	"github.com/fusheng/envkit/internal/docker"
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

// Language 语言环境信息
type Language struct {
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	Version     string `json:"version"`
	Installed   bool   `json:"installed"`
	Mirror      string `json:"mirror"`
}

// Tool 工具信息
type Tool struct {
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	Version     string `json:"version"`
	Installed   bool   `json:"installed"`
}

// Database 数据库容器信息
type Database struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Version string `json:"version"`
	Status  string `json:"status"` // running, stopped
}

// GetLanguages 获取所有语言环境
func (a *App) GetLanguages() []Language {
	detected := detector.DetectLanguages()
	
	languages := []Language{
		{
			Name:        "node",
			DisplayName: "Node.js",
			Version:     "",
			Installed:   false,
		},
		{
			Name:        "python",
			DisplayName: "Python",
			Version:     "",
			Installed:   false,
		},
		{
			Name:        "go",
			DisplayName: "Go",
			Version:     "",
			Installed:   false,
		},
		{
			Name:        "rust",
			DisplayName: "Rust",
			Version:     "",
			Installed:   false,
		},
		{
			Name:        "java",
			DisplayName: "Java",
			Version:     "",
			Installed:   false,
		},
		{
			Name:        "bun",
			DisplayName: "Bun",
			Version:     "",
			Installed:   false,
		},
	}

	// 填充实际状态
	for i := range languages {
		key := languages[i].Name
		if key == "node" {
			key = "node"
		} else if key == "rust" {
			key = "rustc"
		}
		
		if tool := detected[key]; tool != nil && tool.Installed {
			languages[i].Installed = true
			languages[i].Version = tool.Version
		}
	}

	return languages
}

// InstallLanguage 安装语言环境
func (a *App) InstallLanguage(name string, version string) error {
	langInstaller := installer.GetInstaller(name)
	if langInstaller == nil {
		return fmt.Errorf("不支持的语言: %s", name)
	}

	return langInstaller.Install(version)
}

// UninstallLanguage 卸载语言环境
func (a *App) UninstallLanguage(name string) error {
	return installer.UninstallComponent(name)
}

// GetTools 获取所有工具
func (a *App) GetTools() []Tool {
	detected := detector.DetectTools()
	
	tools := []Tool{
		{Name: "git", DisplayName: "Git", Version: "", Installed: false},
		{Name: "docker", DisplayName: "Docker", Version: "", Installed: false},
		{Name: "code", DisplayName: "VSCode", Version: "", Installed: false},
		{Name: "conda", DisplayName: "Miniconda", Version: "", Installed: false},
		{Name: "kubectl", DisplayName: "Kubectl", Version: "", Installed: false},
		{Name: "minikube", DisplayName: "Minikube", Version: "", Installed: false},
	}

	for i := range tools {
		if tool := detected[tools[i].Name]; tool != nil && tool.Installed {
			tools[i].Installed = true
			tools[i].Version = tool.Version
		}
	}

	return tools
}

// InstallTool 安装工具
func (a *App) InstallTool(name string) error {
	toolInstaller := installer.GetToolInstaller(name)
	if toolInstaller == nil {
		return fmt.Errorf("不支持的工具: %s", name)
	}

	return toolInstaller.Install()
}

// UninstallTool 卸载工具
func (a *App) UninstallTool(name string) error {
	return installer.UninstallComponent(name)
}

// GetDatabases 获取数据库容器列表
func (a *App) GetDatabases() []Database {
	// TODO: 实现获取 Docker 容器列表
	return []Database{}
}

// StartDatabase 启动数据库容器
func (a *App) StartDatabase(name string, version string) error {
	dockerMgr := docker.NewContainerManager()
	
	if !dockerMgr.IsDockerRunning() {
		return fmt.Errorf("Docker 未运行")
	}

	switch name {
	case "postgres", "postgresql":
		return dockerMgr.StartPostgreSQL(version, "postgres")
	case "redis":
		return dockerMgr.StartRedis(version)
	case "mysql":
		return dockerMgr.StartMySQL(version, "mysql")
	case "mongodb", "mongo":
		return dockerMgr.StartMongoDB(version)
	default:
		return fmt.Errorf("不支持的数据库: %s", name)
	}
}

// StopDatabase 停止数据库容器
func (a *App) StopDatabase(containerName string) error {
	dockerMgr := docker.NewContainerManager()
	return dockerMgr.StopContainer(containerName)
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
				Message: "轻量级跨平台开发环境管理工具",
			},
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
```

### 2. 前端主应用（frontend/src/App.svelte）

```svelte
<script>
  import { onMount } from 'svelte';
  import Sidebar from './lib/components/Sidebar.svelte';
  import Languages from './routes/Languages.svelte';
  import Tools from './routes/Tools.svelte';
  import Database from './routes/Database.svelte';
  import Settings from './routes/Settings.svelte';
  import { currentPage } from './lib/stores/app.js';

  let page = 'languages';

  currentPage.subscribe(value => {
    page = value;
  });
</script>

<main>
  <div class="app-container">
    <Sidebar />
    <div class="content">
      {#if page === 'languages'}
        <Languages />
      {:else if page === 'tools'}
        <Tools />
      {:else if page === 'database'}
        <Database />
      {:else if page === 'settings'}
        <Settings />
      {/if}
    </div>
  </div>
</main>

<style>
  :global(body) {
    margin: 0;
    padding: 0;
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', 'Roboto', 'Helvetica', 'Arial', sans-serif;
    font-size: 13px;
    color: #1d1d1f;
    background-color: #ffffff;
  }

  .app-container {
    display: flex;
    height: 100vh;
    overflow: hidden;
  }

  .content {
    flex: 1;
    overflow-y: auto;
    padding: 20px;
    background-color: #ffffff;
  }
</style>
```

### 3. 侧边栏组件（frontend/src/lib/components/Sidebar.svelte）

```svelte
<script>
  import { currentPage } from '../stores/app.js';

  function navigate(page) {
    currentPage.set(page);
  }
</script>

<aside class="sidebar">
  <div class="logo">
    <svg width="24" height="24" viewBox="0 0 24 24" fill="none">
      <rect x="3" y="3" width="7" height="7" rx="1" fill="#007aff"/>
      <rect x="14" y="3" width="7" height="7" rx="1" fill="#007aff"/>
      <rect x="3" y="14" width="7" height="7" rx="1" fill="#007aff"/>
      <rect x="14" y="14" width="7" height="7" rx="1" fill="#007aff"/>
    </svg>
    <span>EnvKit</span>
  </div>

  <nav>
    <button 
      class="nav-item" 
      class:active={$currentPage === 'languages'}
      on:click={() => navigate('languages')}
    >
      <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor">
        <path d="M4 19.5A2.5 2.5 0 0 1 6.5 17H20"/>
        <path d="M6.5 2H20v20H6.5A2.5 2.5 0 0 1 4 19.5v-15A2.5 2.5 0 0 1 6.5 2z"/>
      </svg>
      Languages
    </button>

    <button 
      class="nav-item" 
      class:active={$currentPage === 'tools'}
      on:click={() => navigate('tools')}
    >
      <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor">
        <path d="M14.7 6.3a1 1 0 0 0 0 1.4l1.6 1.6a1 1 0 0 0 1.4 0l3.77-3.77a6 6 0 0 1-7.94 7.94l-6.91 6.91a2.12 2.12 0 0 1-3-3l6.91-6.91a6 6 0 0 1 7.94-7.94l-3.76 3.76z"/>
      </svg>
      Tools
    </button>

    <button 
      class="nav-item" 
      class:active={$currentPage === 'database'}
      on:click={() => navigate('database')}
    >
      <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor">
        <ellipse cx="12" cy="5" rx="9" ry="3"/>
        <path d="M21 12c0 1.66-4 3-9 3s-9-1.34-9-3"/>
        <path d="M3 5v14c0 1.66 4 3 9 3s9-1.34 9-3V5"/>
      </svg>
      Database
    </button>

    <button 
      class="nav-item" 
      class:active={$currentPage === 'settings'}
      on:click={() => navigate('settings')}
    >
      <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor">
        <circle cx="12" cy="12" r="3"/>
        <path d="M12 1v6m0 6v6m9-9h-6m-6 0H3"/>
      </svg>
      Settings
    </button>
  </nav>
</aside>

<style>
  .sidebar {
    width: 200px;
    background-color: #f5f5f7;
    border-right: 1px solid #d1d1d6;
    display: flex;
    flex-direction: column;
    padding: 16px 0;
  }

  .logo {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 0 16px 16px 16px;
    font-weight: 600;
    font-size: 15px;
    color: #1d1d1f;
  }

  nav {
    display: flex;
    flex-direction: column;
    gap: 2px;
    padding: 0 8px;
  }

  .nav-item {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 8px 12px;
    background: none;
    border: none;
    border-radius: 4px;
    cursor: pointer;
    color: #1d1d1f;
    font-size: 13px;
    text-align: left;
    transition: background-color 0.1s;
  }

  .nav-item:hover {
    background-color: #ebebeb;
  }

  .nav-item.active {
    background-color: #d1d1d6;
    font-weight: 500;
  }

  .nav-item svg {
    stroke-width: 2;
  }
</style>
```

### 4. 语言页面（frontend/src/routes/Languages.svelte）

```svelte
<script>
  import { onMount } from 'svelte';
  import { GetLanguages, InstallLanguage, UninstallLanguage } from '../../wailsjs/go/main/App';
  
  let languages = [];
  let loading = false;

  onMount(async () => {
    await loadLanguages();
  });

  async function loadLanguages() {
    loading = true;
    try {
      languages = await GetLanguages();
    } catch (err) {
      console.error('Failed to load languages:', err);
    }
    loading = false;
  }

  async function install(lang) {
    if (!confirm(`Install ${lang.displayName}?`)) return;
    
    loading = true;
    try {
      await InstallLanguage(lang.name, lang.version || 'latest');
      await loadLanguages();
    } catch (err) {
      alert(`Installation failed: ${err}`);
    }
    loading = false;
  }

  async function uninstall(lang) {
    if (!confirm(`Uninstall ${lang.displayName}? This will remove all files and environment variables.`)) return;
    
    loading = true;
    try {
      await UninstallLanguage(lang.name);
      await loadLanguages();
    } catch (err) {
      alert(`Uninstallation failed: ${err}`);
    }
    loading = false;
  }
</script>

<div class="page">
  <div class="header">
    <h1>Languages</h1>
    <button class="btn-refresh" on:click={loadLanguages} disabled={loading}>
      {loading ? 'Loading...' : 'Refresh'}
    </button>
  </div>

  <div class="language-list">
    {#each languages as lang}
      <div class="language-card">
        <div class="card-header">
          <div class="language-info">
            <h3>{lang.displayName}</h3>
            {#if lang.installed}
              <span class="status installed">✓ Installed</span>
              <span class="version">{lang.version}</span>
            {:else}
              <span class="status not-installed">✗ Not Installed</span>
            {/if}
          </div>
          <div class="actions">
            {#if lang.installed}
              <button class="btn-secondary" on:click={() => uninstall(lang)} disabled={loading}>
                Uninstall
              </button>
            {:else}
              <button class="btn-primary" on:click={() => install(lang)} disabled={loading}>
                Install
              </button>
            {/if}
          </div>
        </div>
      </div>
    {/each}
  </div>
</div>

<style>
  .page {
    max-width: 1200px;
  }

  .header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 20px;
  }

  h1 {
    font-size: 24px;
    font-weight: 600;
    margin: 0;
  }

  .language-list {
    display: flex;
    flex-direction: column;
    gap: 12px;
  }

  .language-card {
    background: #f5f5f5;
    border: 1px solid #e5e5e5;
    border-radius: 6px;
    padding: 16px;
  }

  .card-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .language-info {
    display: flex;
    align-items: center;
    gap: 12px;
  }

  .language-info h3 {
    font-size: 15px;
    font-weight: 500;
    margin: 0;
  }

  .status {
    padding: 2px 8px;
    border-radius: 4px;
    font-size: 12px;
  }

  .status.installed {
    background: #28a74520;
    color: #28a745;
  }

  .status.not-installed {
    background: #86868b20;
    color: #86868b;
  }

  .version {
    font-size: 12px;
    color: #6e6e73;
  }

  .actions {
    display: flex;
    gap: 8px;
  }

  button {
    padding: 6px 16px;
    border: none;
    border-radius: 4px;
    font-size: 13px;
    cursor: pointer;
    transition: opacity 0.1s;
  }

  button:hover:not(:disabled) {
    opacity: 0.8;
  }

  button:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .btn-primary {
    background: #007aff;
    color: white;
  }

  .btn-secondary {
    background: #e5e5e5;
    color: #1d1d1f;
  }

  .btn-refresh {
    background: #f5f5f5;
    color: #1d1d1f;
    border: 1px solid #d1d1d6;
  }
</style>
```

### 5. 状态管理（frontend/src/lib/stores/app.js）

```javascript
import { writable } from 'svelte/store';

export const currentPage = writable('languages');
export const loading = writable(false);
export const systemInfo = writable({
  os: '',
  arch: '',
  distribution: ''
});
```

---

## 构建和运行

### 开发模式
```bash
cd envkit
wails dev
```

### 构建生产版本
```bash
# 构建当前平台
wails build

# 构建所有平台
wails build -platform darwin/universal  # macOS
wails build -platform windows/amd64     # Windows
wails build -platform linux/amd64       # Linux
```

### 输出位置
```
build/bin/
├── envkit-gui.app          # macOS
├── envkit-gui.exe          # Windows
└── envkit-gui              # Linux
```

---

## 下一步优化

1. **进度反馈**
   - 实现安装进度条
   - 显示实时日志

2. **错误处理**
   - 友好的错误提示
   - 详细的错误日志

3. **性能优化**
   - 后台刷新状态
   - 缓存机制

4. **用户体验**
   - 暗色模式
   - 快捷键支持
   - 系统托盘
