# EnvKit 桌面客户端 - 项目总览

## 快速开始

```bash
# 1. 运行初始化脚本
./scripts/init-gui.sh

# 2. 启动开发服务器
wails dev

# 3. 构建生产版本
wails build
```

---

## 设计核心

### 设计哲学
**"工具感优先，零AI风味"**

- ✅ 像系统设置一样实用
- ✅ 像终端模拟器一样直接
- ✅ 像 VS Code 一样专业
- ❌ 不要紫蓝渐变
- ❌ 不要毛玻璃效果
- ❌ 不要花哨动画
- ❌ 不要低信息密度

### 技术选型

| 技术 | 选择 | 原因 |
|------|------|------|
| 框架 | **Wails v2** | Go 原生集成，体积小（<10MB），性能好 |
| UI | **Svelte** | 轻量、编译时优化、无虚拟DOM |
| 样式 | **Tailwind CSS** | Utility-first，无运行时 |
| 图标 | **Lucide Icons** | SVG，tree-shakeable |

**为什么不用 Electron？**
- 打包体积 150MB+（Wails 仅 10MB）
- 内存占用高（Wails 使用系统 WebView）
- 启动慢

**为什么不用 Tauri？**
- Rust 技术栈与现有 Go 代码割裂
- Wails 更简单，社区活跃

---

## 项目结构

```
envkit/
├── cmd/
│   ├── envkit/              # CLI 版本（已有）
│   └── gui/                 # GUI 版本（新增）
│       └── main.go          # GUI 入口
│
├── internal/                # 共享代码（CLI 和 GUI 复用）
│   ├── installer/           # 安装逻辑
│   ├── detector/            # 检测逻辑
│   ├── docker/              # Docker 管理
│   ├── mirror/              # 镜像源配置
│   └── config/              # 配置解析
│
├── frontend/                # Svelte 前端
│   ├── src/
│   │   ├── App.svelte       # 主应用
│   │   ├── lib/
│   │   │   ├── components/  # UI 组件
│   │   │   │   ├── Sidebar.svelte
│   │   │   │   ├── LanguageCard.svelte
│   │   │   │   ├── ToolCard.svelte
│   │   │   │   └── DatabaseTable.svelte
│   │   │   └── stores/      # 状态管理
│   │   │       └── app.js
│   │   └── routes/          # 页面路由
│   │       ├── Languages.svelte
│   │       ├── Tools.svelte
│   │       ├── Database.svelte
│   │       └── Settings.svelte
│   ├── package.json
│   └── vite.config.js
│
├── docs/
│   ├── GUI_DESIGN.md        # 设计方案
│   ├── GUI_IMPLEMENTATION.md # 实现指南
│   └── GUI_VISUAL_SPEC.md   # 视觉规范
│
├── scripts/
│   └── init-gui.sh          # 快速初始化脚本
│
└── wails.json               # Wails 配置
```

---

## 界面预览

### 布局
```
┌─────────────────────────────────────────────────┐
│  Title Bar (原生)                                │
├─────────┬───────────────────────────────────────┤
│ Sidebar │           Main Content                │
│         │                                        │
│ [Logo]  │  ┌──────────────────────────────┐   │
│ EnvKit  │  │                              │   │
│         │  │      Languages / Tools       │   │
│ ◉ Languages│  │      Database / Settings     │   │
│ ○ Tools  │  │                              │   │
│ ○ Database│  └──────────────────────────────┘   │
│ ○ Settings│                                     │
│         │  [Status Bar]                         │
└─────────┴───────────────────────────────────────┘
```

### 配色（Light Mode）
```
背景：    #ffffff (纯白)
侧边栏：  #f5f5f7 (浅灰)
边框：    #d1d1d6
文本：    #1d1d1f
链接：    #007aff (苹果蓝)
成功：    #28a745 (绿)
警告：    #ff9500 (橙)
错误：    #dc3545 (红)
```

### 字体
```
macOS:   -apple-system (SF Pro)
Windows: Segoe UI
Linux:   System UI
尺寸：   13px (正文), 15px (标题), 11px (辅助)
```

---

## 核心功能

### 1. Languages（语言环境管理）
- 展示已安装/未安装语言
- 一键安装/卸载
- 配置镜像源
- 版本管理

### 2. Tools（开发工具管理）
- 网格布局展示工具卡片
- 安装/卸载 Git、Docker、VSCode 等
- 显示版本信息

### 3. Database（数据库容器管理）
- 表格展示 Docker 容器
- 启动/停止/删除容器
- 显示运行状态

### 4. Settings（设置）
- 默认镜像源配置
- 启动选项
- 高级设置

---

## API 设计

### Go 后端暴露的方法
```go
// 语言管理
func (a *App) GetLanguages() []Language
func (a *App) InstallLanguage(name, version string) error
func (a *App) UninstallLanguage(name string) error

// 工具管理
func (a *App) GetTools() []Tool
func (a *App) InstallTool(name string) error
func (a *App) UninstallTool(name string) error

// 数据库管理
func (a *App) GetDatabases() []Database
func (a *App) StartDatabase(name, version string) error
func (a *App) StopDatabase(containerName string) error

// 系统信息
func (a *App) GetSystemInfo() map[string]string
```

### 前端调用示例
```javascript
import { GetLanguages, InstallLanguage } from '../wailsjs/go/main/App'

// 获取语言列表
const languages = await GetLanguages()

// 安装 Node.js
await InstallLanguage('node', '20.11.1')
```

---

## 性能指标

### 目标
```
启动时间：   < 500ms (冷启动)
            < 200ms (热启动)

内存占用：   30-50 MB (初始)
            50-80 MB (稳定运行)

安装包体积： 8-12 MB (所有平台)
```

### 对比 Electron
```
              Wails    Electron
启动时间      500ms    2-3s
内存占用      50MB     150-300MB
安装包体积    10MB     150MB+
```

---

## 开发流程

### 1. 初始化项目
```bash
./scripts/init-gui.sh
```

### 2. 开发模式
```bash
wails dev
```
- 热重载
- 开发工具
- 实时预览

### 3. 构建生产版本
```bash
# 当前平台
wails build

# 指定平台
wails build -platform darwin/universal  # macOS (Intel + Apple Silicon)
wails build -platform windows/amd64     # Windows
wails build -platform linux/amd64       # Linux
```

### 4. 输出位置
```
build/bin/
├── envkit-gui.app          # macOS
├── envkit-gui.exe          # Windows
└── envkit-gui              # Linux
```

---

## 与 CLI 版本的关系

### 代码复用
```
✅ 完全复用 internal/ 下的所有模块
✅ 共享配置文件 (~/.envkit/)
✅ 共享安装记录 (manifest.json)
✅ 两个版本可以同时使用
```

### 差异
```
CLI 版本：
- 命令行交互
- 适合脚本自动化
- 服务器环境

GUI 版本：
- 图形界面
- 适合日常使用
- 桌面环境
```

---

## 路线图

### Phase 1: MVP（当前目标）
- [x] 项目框架搭建
- [ ] 语言环境管理页面
- [ ] 工具管理页面
- [ ] 数据库容器管理页面
- [ ] 基础设置页面

### Phase 2: 增强
- [ ] 配置文件导入/导出
- [ ] 多环境切换
- [ ] 安装日志查看
- [ ] 更新检查

### Phase 3: 高级功能
- [ ] 批量操作
- [ ] 自定义脚本
- [ ] 云端同步
- [ ] 团队共享配置

---

## 常见问题

### Q: 为什么选择 Wails 而不是 Electron？
**A:** 体积小（10MB vs 150MB）、性能好、内存占用低、与现有 Go 代码无缝集成。

### Q: 支持哪些操作系统？
**A:** macOS 10.13+, Windows 10+, Linux (主流发行版)

### Q: 能否与 CLI 版本共存？
**A:** 可以，两者共享配置和安装记录，可以互相配合使用。

### Q: 需要安装依赖吗？
**A:** 开发需要 Go + Node.js，用户使用不需要任何依赖。

### Q: 如何调试？
**A:** `wails dev` 会自动打开 DevTools，可以像调试网页一样调试。

---

## 相关文档

- **[GUI_DESIGN.md](./GUI_DESIGN.md)** - 完整设计方案和架构说明
- **[GUI_IMPLEMENTATION.md](./GUI_IMPLEMENTATION.md)** - 详细实现指南和代码示例
- **[GUI_VISUAL_SPEC.md](./GUI_VISUAL_SPEC.md)** - 视觉规范和组件库

---

## 贡献指南

### 设计原则
1. **工具感优先** - 像系统工具，不像 AI 产品
2. **性能第一** - 快速启动，低内存占用
3. **简洁实用** - 无多余装饰，高信息密度
4. **跨平台一致** - 遵循各平台的原生规范

### 代码规范
- Go: `gofmt` + `golint`
- JavaScript: ESLint + Prettier
- Svelte: 遵循官方风格指南

---

## 许可证

MIT License - 与主项目保持一致

---

## 联系方式

- **项目主页**: https://github.com/fusheng/envkit
- **Issues**: https://github.com/fusheng/envkit/issues
- **文档**: https://github.com/fusheng/envkit/tree/main/docs

---

**开始构建属于你的开发环境管理桌面应用！** 🚀
