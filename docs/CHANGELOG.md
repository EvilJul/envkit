# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.2.3] - 2026-07-20

### Fixed
- **uv PATH 持久化**：安装后即使已能检测到 uv，也会将 `~/.local/bin` 写入 shell 配置
- 重复执行 `envkit install uv` 会补写 PATH，避免「list 显示已安装但终端找不到 uv」

### Changed
- 版本默认与 Release 对齐为 **0.2.3**

## [0.2.2] - 2026-07-20

### Added
- 独立工具 **uv**（Astral）：CLI/GUI 可单独安装与卸载；Python 安装复用同一安装器
- `GetDatabases` 对接 Docker 容器列表，GUI 数据库页可显示运行中的容器
- 安装清单组件名归一化（`code`→`vscode` 等），避免装卸名不一致

### Fixed
- GUI 假按钮：移除无后端的 vim/curl/wget、MariaDB/ClickHouse
- 清单 `RecordInstallation` 失败不再静默忽略
- Docker 容器存在性检测改为精确匹配
- Environment 页：Linux bash 优先 `.bashrc`，PATH 使用系统分隔符
- `install.sh`：源码树有 Go 时优先本地编译，避免覆盖为旧 GitHub Release
- StartStack 未完整实现时返回明确错误，避免假成功
- 进度 Reporter 并发安全；重复 GUI 入口 build ignore；frontend/dist 占位

### Changed
- 版本默认与 Release 对齐为 **0.2.2**

## [0.1.0] - 2026-06-07

### Added

#### Core Features
- 跨平台支持（Windows/macOS/Linux）
- 系统检测功能（OS、架构、Linux发行版）
- 已安装工具检测（语言、开发工具、包管理器）
- YAML 配置文件解析和验证
- 配置文件导出功能

#### 镜像源配置
- npm/pnpm/yarn 国内镜像源配置（淘宝、腾讯、华为）
- pip 国内镜像源配置（清华、阿里、中科大、豆瓣）
- Go proxy 配置（goproxy.cn、阿里、七牛）
- Rust cargo 镜像源配置（中科大、上交、清华、字节）
- 镜像源注册表系统

#### 预设模板
- 前端开发环境模板（Node.js + pnpm）
- 后端开发环境模板（Go + Python + Docker）
- 全栈开发环境模板（Node.js + Go + Python + Docker）

#### CLI 命令
- `envkit init` - 交互式生成配置文件
- `envkit install` - 根据配置文件安装环境
- `envkit detect` - 检测系统环境
- `envkit mirror` - 单独配置镜像源
- `envkit version` - 显示版本信息
- `envkit help` - 显示帮助信息

#### 文档
- README.md 主文档
- docs/USAGE.md 使用指南
- LICENSE MIT 许可证
- CHANGELOG.md 变更日志

#### 构建
- 多平台构建脚本 (build.sh)
- 支持 Linux/macOS/Windows (amd64/arm64)

### Technical Details
- Go 1.21+ 支持
- gopkg.in/yaml.v3 配置解析
- embed 嵌入模板文件
- 约 1365 行 Go 代码

## [0.2.0] - 2026-06-08

### Added

#### 自动安装功能
- 自动安装 Node.js (通过 fnm/brew/winget)
- 自动安装 Python (通过 uv/brew/winget)
- 自动安装 Go (通过 brew/官方安装包/winget)
- 自动安装 Rust (通过 rustup)
- 自动安装 Git (通过 brew/apt/winget)
- 自动安装 Docker (Linux 通过官方脚本)
- 自动安装 VSCode (通过 brew/deb包/winget)

#### Docker 容器管理
- 启动 PostgreSQL 容器
- 启动 Redis 容器
- 启动 MySQL 容器
- 启动 MongoDB 容器
- 列出所有容器
- 停止/删除容器
- 数据卷管理

#### TUI 界面增强
- 彩色输出 (Success/Error/Warning/Info)
- 进度条显示
- 旋转加载器
- 表格显示
- 分区标题

#### 新增命令
- `envkit docker start <db> <version>` - 启动数据库容器
- `envkit docker stop <container>` - 停止容器
- `envkit docker list` - 列出容器
- `envkit docker remove <container>` - 删除容器

### Changed
- 版本号更新为 0.2.0
- `envkit install` 现在会自动安装语言和工具
- `envkit detect` 使用表格显示检测结果
- 改进错误提示和用户反馈

### Technical Details
- 新增 internal/ui 模块 (~300 行)
- 新增 internal/installer 模块 (~400 行)
- 新增 internal/docker 模块 (~300 行)
- 总代码量约 2194 行

## [Unreleased] / 0.3.0

### 新增功能
- ✨ 全局安装 Android SDK 开发环境（cmdline-tools、platform-tools、build-tools、platforms）
- ✨ 通过国内镜像源（阿里云/腾讯云/华为云）下载 Android SDK
- ✨ 全局 Gradle 国内镜像源配置（aliyun）
- ✨ 同步支持 CLI 和 GUI 两种方式管理 Android SDK
- ✨ 新增命令 `envkit install android`、`envkit mirror android` 和 `envkit mirror gradle`

### 改进
- 更新 detector 支持检测 adb 和 sdkmanager
- GUI 桌面客户端 Tools 和 Settings 页面添加 Android 管理入口

### 文档
- 新增 `templates/languages/android.yaml` 模板
- 新增 `templates/examples/android-mobile.yaml` 移动端开发示例
- 新增 `docs/ANDROID_SETUP.md` Android 开发环境配置指南

## [Unreleased]

### Planned for v0.3.0
- dotfiles 管理和同步
- Shell 环境配置（oh-my-zsh, starship）
- 云端配置同步
- 插件系统
- GUI 界面（可选）

---

[0.1.0]: https://github.com/fusheng/envkit/releases/tag/v0.1.0
