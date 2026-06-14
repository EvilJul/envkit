# EnvKit 桌面客户端设计方案

## 设计原则

### 核心理念
- **工具感优先**：像系统设置、终端模拟器那样实用、直接
- **轻量化架构**：启动快、内存占用小、安装包体积控制在 10MB 以内
- **原生体验**：遵循各平台的视觉规范，不追求"统一外观"
- **零装饰主义**：禁止渐变、毛玻璃、过度圆角、花哨动画

### 反模式（避免的设计）
- ❌ AI 产品的那种"未来感"设计（紫蓝渐变、发光效果）
- ❌ 过度使用卡片和阴影
- ❌ 动画过于丰富
- ❌ 信息密度过低（一屏只显示几个项目）

### 参考对象
- ✅ macOS 系统偏好设置
- ✅ VS Code / Sublime Text
- ✅ Linear（简洁、高效）
- ✅ Spotify Desktop（功能清晰）

---

## 技术栈选型

### 框架：Wails v2
**为什么选 Wails？**
```
✓ Go 原生集成（无需重写后端逻辑）
✓ 原生性能（系统 WebView，不打包 Chromium）
✓ 体积小（<10MB，Electron 通常 >150MB）
✓ 原生菜单、对话框、系统托盘
✓ 支持 macOS 10.13+, Windows 10+, Linux
```

### 前端技术栈
```
UI 框架：Svelte（比 React/Vue 更轻量，编译时优化）
样式方案：Tailwind CSS（utility-first，无运行时）
图标库：Lucide Icons（SVG，tree-shakeable）
状态管理：Svelte Stores（内置，无需额外库）
```

### 为什么不用 Electron/Tauri？
```
Electron：打包体积过大（150MB+），内存占用高
Tauri：   Rust 技术栈，与现有 Go 代码库割裂
```

---

## 界面架构

### 布局结构
```
┌─────────────────────────────────────────────────┐
│  [Title Bar - 原生样式]                          │
├─────────┬───────────────────────────────────────┤
│         │                                         │
│  侧边栏  │           主内容区                      │
│         │                                         │
│  [图标]  │  ┌───────────────────────────────┐   │
│  Languages│  │                               │   │
│  Tools   │  │      内容区域                  │   │
│  Database│  │                               │   │
│  Settings│  └───────────────────────────────┘   │
│         │                                         │
│         │  [底部状态栏]                           │
└─────────┴───────────────────────────────────────┘
```

### 侧边栏设计
```css
宽度：200px 固定
背景：系统原生背景色（macOS: #f5f5f7, Windows: #f3f3f3）
字体：系统默认字体，13px
间距：紧凑布局，项目间距 2px
图标：16x16，单色，与文本对齐
悬停：浅灰背景 (#e8e8e8)，无圆角
选中：深灰背景 (#d1d1d6)，无高光
```

---

## 页面设计

### 1. Languages（语言环境）

**布局：列表视图**
```
┌─────────────────────────────────────────────┐
│  Languages                         [+ Add]  │
├─────────────────────────────────────────────┤
│                                             │
│  ┌─ Node.js ──────────────────── ✓ v20.11.1┐
│  │  Mirror: npmmirror                  [⚙]│
│  │  Package Manager: pnpm                  │
│  └────────────────────────────────────────┘
│                                             │
│  ┌─ Python ──────────────────────── ✗ Not │
│  │  Recommended: 3.10.11            [Install]│
│  └────────────────────────────────────────┘
│                                             │
│  ┌─ Go ──────────────────────────── ✓ v1.22│
│  │  Mirror: goproxy                    [⚙]│
│  └────────────────────────────────────────┘
│                                             │
└─────────────────────────────────────────────┘
```

**状态指示器**
```
✓ 已安装：绿色勾
✗ 未安装：灰色叉
⚠ 需要更新：黄色叹号
```

### 2. Tools（开发工具）

**布局：网格视图（3列）**
```
┌─────────────────────────────────────────────┐
│  Tools                                      │
├─────────────────────────────────────────────┤
│                                             │
│  ┌────────┐  ┌────────┐  ┌────────┐       │
│  │  Git   │  │ Docker │  │ VSCode │       │
│  │  ✓ 2.40│  │ ✗ N/A  │  │ ✓ 1.85 │       │
│  └────────┘  └────────┘  └────────┘       │
│     [⚙]      [Install]      [⚙]           │
│                                             │
│  ┌────────┐  ┌────────┐  ┌────────┐       │
│  │ Kubectl│  │Minikube│  │ Conda  │       │
│  │ ✗ N/A  │  │ ✗ N/A  │  │ ✗ N/A  │       │
│  └────────┘  └────────┘  └────────┘       │
│  [Install]  [Install]   [Install]         │
│                                             │
└─────────────────────────────────────────────┘
```

### 3. Database（数据库容器）

**布局：表格视图**
```
┌──────────────────────────────────────────────────┐
│  Database Containers              [+ Start New]  │
├──────────────────────────────────────────────────┤
│                                                  │
│  Name              Type       Version   Status  │
│  ────────────────────────────────────────────── │
│  envkit-postgres   PostgreSQL  16      ● Running│
│  envkit-redis      Redis       7       ● Running│
│  envkit-mysql      MySQL       8.0     ○ Stopped│
│                                                  │
│  Right-click for: Stop | Restart | Remove      │
└──────────────────────────────────────────────────┘
```

### 4. Settings（设置）

**布局：表单视图**
```
┌─────────────────────────────────────────────┐
│  Settings                                   │
├─────────────────────────────────────────────┤
│                                             │
│  General                                    │
│  ├─ Launch at startup         [ ]          │
│  └─ Minimize to tray          [✓]          │
│                                             │
│  Mirror                                     │
│  ├─ Default npm mirror        [npmmirror▾] │
│  ├─ Default pip mirror        [tsinghua ▾] │
│  └─ Default go mirror         [goproxy  ▾] │
│                                             │
│  Advanced                                   │
│  ├─ Installation directory    [/usr/local] │
│  └─ Show expert options       [ ]          │
│                                             │
└─────────────────────────────────────────────┘
```

---

## 配色方案

### Light Mode（默认）
```css
--background:       #ffffff   /* 纯白 */
--surface:          #f5f5f5   /* 浅灰 */
--surface-hover:    #ebebeb   /* 悬停 */
--border:           #d1d1d6   /* 边框 */
--text-primary:     #1d1d1f   /* 主文本 */
--text-secondary:   #6e6e73   /* 次要文本 */
--text-tertiary:    #86868b   /* 三级文本 */

--success:          #28a745   /* 绿色 - 已安装 */
--warning:          #ffc107   /* 黄色 - 警告 */
--error:            #dc3545   /* 红色 - 错误 */
--info:             #007aff   /* 蓝色 - 信息 */
```

### Dark Mode（可选）
```css
--background:       #1e1e1e
--surface:          #2d2d2d
--surface-hover:    #3a3a3a
--border:           #3e3e42
--text-primary:     #ffffff
--text-secondary:   #adadb8
--text-tertiary:    #86868b
```

---

## 交互细节

### 安装流程
```
1. 点击 [Install] 按钮
2. 弹出确认对话框（原生系统对话框）
   ┌────────────────────────────┐
   │ Install Node.js 20.11.1?   │
   │                            │
   │ [ ] Configure npm mirror   │
   │                            │
   │     [Cancel]  [Install]    │
   └────────────────────────────┘
3. 安装中显示进度
   Installing Node.js... ████░░░░░░ 45%
4. 完成后自动刷新状态
```

### 卸载流程
```
1. 右键菜单 → Uninstall
2. 确认对话框（红色警告）
   ┌────────────────────────────┐
   │ ⚠ Uninstall Node.js?       │
   │                            │
   │ This will remove:          │
   │ • /usr/local/node          │
   │ • Environment variables    │
   │                            │
   │     [Cancel]  [Uninstall]  │
   └────────────────────────────┘
```

### 状态更新
```
底部状态栏实时显示：
┌────────────────────────────────────────┐
│ Last scan: 2 min ago  |  Docker: ● Up │
└────────────────────────────────────────┘

每 30 秒自动刷新一次状态（后台扫描）
```

---

## 性能指标

### 启动时间
```
冷启动：< 500ms
热启动：< 200ms
```

### 内存占用
```
初始：30-50 MB
稳定运行：50-80 MB
```

### 安装包体积
```
macOS (.dmg):   8-12 MB
Windows (.exe):  8-12 MB
Linux (.deb):    8-12 MB
```

---

## 开发路线图

### Phase 1: MVP（核心功能）
```
✓ 语言环境列表 + 安装/卸载
✓ 工具列表 + 安装/卸载
✓ 数据库容器管理
✓ 系统扫描
```

### Phase 2: 增强功能
```
□ 配置文件导入/导出
□ 多环境管理（切换不同配置）
□ 日志查看器
□ 更新检查
```

### Phase 3: 高级功能
```
□ 批量操作
□ 自定义安装脚本
□ 远程环境同步
```

---

## 技术实现细节

### 目录结构
```
envkit-gui/
├── frontend/           # Svelte 前端
│   ├── src/
│   │   ├── lib/       # 组件库
│   │   ├── routes/    # 页面路由
│   │   └── stores/    # 状态管理
│   └── package.json
├── main.go            # Wails 主程序
├── app.go             # 应用逻辑
└── wails.json         # Wails 配置
```

### Go 后端接口
```go
type App struct {
    ctx context.Context
}

// 暴露给前端的方法
func (a *App) GetLanguages() []Language
func (a *App) InstallLanguage(name, version string) error
func (a *App) UninstallLanguage(name string) error
func (a *App) GetTools() []Tool
func (a *App) GetDatabases() []Database
func (a *App) StartDatabase(name, version string) error
```

### 前端调用示例
```javascript
import { GetLanguages, InstallLanguage } from '../wailsjs/go/main/App'

// 获取语言列表
const languages = await GetLanguages()

// 安装语言
await InstallLanguage('node', '20.11.1')
```

---

## 与 CLI 版本的关系

### 代码复用
```
桌面版复用 internal/ 下的所有模块：
✓ installer/*
✓ detector/*
✓ docker/*
✓ mirror/*
✓ config/*

只需新增：
+ gui/           # GUI 特定代码
+ frontend/      # 前端代码
```

### 统一配置
```yaml
# 两个版本共享同一配置文件
~/.envkit/config.yaml
~/.envkit/manifest.json
```

---

## 待解决问题

1. **权限提升**
   - macOS/Linux: 需要 sudo 时弹出系统密码框
   - Windows: UAC 提示

2. **后台任务**
   - 安装/卸载是长时间操作
   - 需要显示实时进度和日志

3. **多实例处理**
   - CLI 和 GUI 同时运行时的冲突

---

## 下一步

1. 搭建 Wails 项目框架
2. 实现核心 UI 组件
3. 对接现有 Go 后端
4. 测试跨平台兼容性
