# EnvKit 使用指南

> EnvKit v0.2.0 - 一键配置开发环境的跨平台CLI工具

## 快速开始

### 1. 交互式生成配置文件

```bash
envkit init
```

选择预设模板（前端/后端/全栈），自动生成 `dev-env.yaml` 配置文件。

### 2. 自动安装开发环境

```bash
envkit install -f dev-env.yaml
```

根据配置文件自动完成：
- ✅ 安装编程语言（Node.js/Python/Go/Rust）
- ✅ 配置国内镜像源
- ✅ 安装开发工具（Git/Docker/VSCode）
- ✅ 启动数据库容器（PostgreSQL/Redis/MySQL/MongoDB）

### 3. 检测系统环境

```bash
envkit detect
```

以表格形式查看当前系统已安装的编程语言、开发工具和包管理器。

### 4. 管理 Docker 容器

```bash
# 启动数据库容器
envkit docker start postgres 16
envkit docker start redis 7

# 查看容器列表
envkit docker list

# 停止容器
envkit docker stop envkit-postgres

# 删除容器
envkit docker remove envkit-postgres
```

## 自动安装语言

EnvKit v0.2.0 支持自动安装以下编程语言：

### Node.js

```bash
# 通过配置文件安装
# dev-env.yaml:
languages:
  - name: node
    version: "20.x"
    mirror: npmmirror
```

**安装方式：**
- macOS: Homebrew
- Linux: fnm (Fast Node Manager)
- Windows: winget

### Python

```bash
languages:
  - name: python
    version: "3.12"
    mirror: tsinghua
```

**安装方式：**
- macOS: Homebrew
- Linux: uv (高性能 Python 包管理器)
- Windows: winget

### Go

```bash
languages:
  - name: go
    version: "1.23"
    mirror: goproxy
```

**安装方式：**
- macOS: Homebrew
- Linux: 官方安装包
- Windows: winget

### Rust

```bash
languages:
  - name: rust
    version: "stable"
    mirror: ustc
```

**安装方式：**
- 所有平台: rustup

## 自动安装工具

### Git

自动安装 Git 版本控制系统。

**安装方式：**
- macOS: Homebrew
- Linux: apt-get
- Windows: winget

### Docker

自动安装 Docker 容器引擎。

**安装方式：**
- Linux: 官方脚本
- macOS/Windows: 提示手动下载 Docker Desktop

### VSCode

自动安装 Visual Studio Code 编辑器。

**安装方式：**
- macOS: Homebrew Cask
- Linux: .deb 包
- Windows: winget

## Docker 容器管理

### 启动数据库容器

#### PostgreSQL

```bash
envkit docker start postgres 16
```

**连接信息：**
- 主机: localhost
- 端口: 5432
- 用户: postgres
- 密码: postgres
- 数据卷: envkit-postgres-data

#### Redis

```bash
envkit docker start redis 7
```

**连接信息：**
- 主机: localhost
- 端口: 6379
- 数据卷: envkit-redis-data
- 持久化: 已启用（AOF）

#### MySQL

```bash
envkit docker start mysql 8
```

**连接信息：**
- 主机: localhost
- 端口: 3306
- 用户: root
- 密码: mysql
- 数据卷: envkit-mysql-data

#### MongoDB

```bash
envkit docker start mongo 7
```

**连接信息：**
- 主机: localhost
- 端口: 27017
- 数据卷: envkit-mongodb-data
- 认证: 无（开发环境）

### 容器管理命令

```bash
# 列出所有 envkit 管理的容器
envkit docker list

# 停止容器
envkit docker stop envkit-postgres

# 删除容器（会询问是否删除数据卷）
envkit docker remove envkit-postgres

# 删除容器并删除数据卷
# 在提示时输入 'y'
envkit docker remove envkit-postgres
# 是否同时删除数据卷? (y/N): y
```

### Node.js (npm)

```bash
envkit mirror npm npmmirror
```

支持的镜像源：
- `npmmirror` - 淘宝镜像（推荐）
- `tencent` - 腾讯云镜像
- `huawei` - 华为云镜像

配置后会同时设置 npm、yarn 和 pnpm 的镜像源。

### Python (pip)

```bash
envkit mirror pip tsinghua
```

支持的镜像源：
- `tsinghua` - 清华大学镜像（推荐）
- `aliyun` - 阿里云镜像
- `ustc` - 中科大镜像
- `douban` - 豆瓣镜像

配置后会创建 `~/.pip/pip.conf` (Linux/macOS) 或 `%APPDATA%\pip\pip.ini` (Windows)。

### Go

```bash
envkit mirror go goproxy
```

支持的镜像源：
- `goproxy` - goproxy.cn（推荐）
- `aliyun` - 阿里云镜像
- `qiniu` - 七牛云镜像

配置后会设置 `GOPROXY` 环境变量。

### Rust

```bash
envkit mirror rust ustc
```

支持的镜像源：
- `ustc` - 中科大镜像（推荐）
- `sjtu` - 上海交大镜像
- `tsinghua` - 清华大学镜像
- `bytedance` - 字节跳动镜像

配置后会创建 `~/.cargo/config` 文件。

## 配置文件格式

`dev-env.yaml` 示例：

```yaml
version: "1.0"
name: "My Development Stack"

languages:
  - name: node
    version: "20.x"
    mirror: npmmirror
    package_manager: pnpm
  
  - name: python
    version: "3.12"
    mirror: tsinghua
  
  - name: go
    version: "1.23"
    mirror: goproxy

tools:
  - git
  - docker
  - vscode

databases:
  - name: postgresql
    version: "16"
    docker: true
  
  - name: redis
    version: "7"
    docker: true

shell:
  type: zsh
  plugins:
    - oh-my-zsh
```

## 预设模板

### 前端开发环境
- Node.js 20.x
- pnpm 包管理器
- 自动配置 npm 淘宝镜像

### 后端开发环境
- Go 1.23
- Python 3.12
- Docker
- PostgreSQL 16 (Docker)
- Redis 7 (Docker)

### 全栈开发环境
- Node.js 20.x + Go 1.23 + Python 3.12
- Docker
- PostgreSQL + Redis

## 常见问题

### 1. 镜像源配置后不生效？

**npm:**
```bash
npm config get registry
```

**pip:**
```bash
pip config list
```

**go:**
```bash
go env GOPROXY
```

### 2. 需要重新配置镜像源？

直接再次运行 mirror 命令即可覆盖：
```bash
envkit mirror npm npmmirror
```

### 3. 如何恢复官方源？

暂时不支持自动恢复，可以手动配置：

**npm:**
```bash
npm config set registry https://registry.npmjs.org/
```

**pip:**
删除 `~/.pip/pip.conf` 文件

**go:**
```bash
go env -w GOPROXY=https://proxy.golang.org,direct
```

## 当前版本限制

EnvKit v0.1.0 是 MVP 版本，目前支持的功能：

✅ 系统检测（操作系统、架构、已安装工具）
✅ 配置文件生成（3个预设模板）
✅ 国内镜像源配置（npm/pip/go/rust）
✅ 跨平台支持（Windows/macOS/Linux）

🚧 正在开发中的功能：

⏸️ 自动安装编程语言和工具
⏸️ TUI 交互界面（进度条、实时日志）
⏸️ Docker 容器安装数据库
⏸️ dotfiles 管理
⏸️ Shell 环境配置（oh-my-zsh等）

## 贡献

欢迎提交 Issue 和 Pull Request！

## 许可证

MIT License
