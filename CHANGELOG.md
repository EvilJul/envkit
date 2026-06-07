# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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

## [Unreleased]

### Planned for v0.2.0
- 自动安装编程语言（fnm for Node.js, uv for Python）
- 自动安装开发工具（git, docker, vscode）
- Docker 容器管理数据库
- TUI 交互界面（进度条、实时日志）

### Planned for v0.3.0
- dotfiles 管理和同步
- Shell 环境配置（oh-my-zsh, starship）
- 云端配置同步
- 插件系统
- GUI 界面（可选）

---

[0.1.0]: https://github.com/fusheng/envkit/releases/tag/v0.1.0
