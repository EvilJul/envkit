# EnvKit

🚀 一键配置开发环境的跨平台CLI工具，专为中国开发者优化

![Version](https://img.shields.io/badge/version-0.2.0-blue)
![Go](https://img.shields.io/badge/go-1.21+-00ADD8?logo=go)
![License](https://img.shields.io/badge/license-MIT-green)

## ✨ 特性

- ✅ **跨平台支持** - Windows/macOS/Linux 全平台支持
- ✅ **自动安装语言** - 自动安装 Node.js/Python/Go/Rust (v0.2.0 新增)
- ✅ **自动安装工具** - 自动安装 Git/Docker/VSCode (v0.2.0 新增)
- ✅ **Docker 容器管理** - 一键启动 PostgreSQL/Redis/MySQL/MongoDB (v0.2.0 新增)
- ✅ **TUI 界面** - 彩色输出、进度条、表格显示 (v0.2.0 新增)
- ✅ **预设模板** - 前端/后端/全栈开发环境一键生成
- ✅ **国内镜像源** - 自动配置 npm/pip/go/rust 等国内镜像
- ✅ **智能检测** - 自动检测已安装的语言和工具
- ✅ **配置管理** - 基于 YAML 的声明式配置

## 🚀 快速开始

### 下载安装

从 [Releases](https://github.com/fusheng/envkit/releases) 下载适合你系统的二进制文件，或者使用 Go 安装：

```bash
go install github.com/fusheng/envkit/cmd/envkit@latest
```

### 基本使用

```bash
# 1. 生成配置文件（选择预设模板）
envkit init

# 2. 安装并配置环境（自动安装语言、工具、启动容器）
envkit install -f dev-env.yaml

# 3. 检测当前系统环境
envkit detect

# 4. 管理 Docker 容器
envkit docker start postgres 16  # 启动 PostgreSQL
envkit docker list              # 查看所有容器
```

### 单独配置镜像源

```bash
# 配置 npm 镜像源
envkit mirror npm npmmirror

# 配置 pip 镜像源
envkit mirror pip tsinghua

# 配置 Go 代理
envkit mirror go goproxy

# 配置 Rust 镜像源
envkit mirror rust ustc
```

## 📖 预设模板

### 🎨 前端开发环境
- Node.js 20.x
- pnpm 包管理器
- npm 淘宝镜像

```bash
envkit init  # 选择 1
```

### ⚙️ 后端开发环境
- Go 1.23
- Python 3.12
- Docker
- PostgreSQL + Redis (Docker)

```bash
envkit init  # 选择 2
```

### 🔥 全栈开发环境
- Node.js + Go + Python
- Docker
- 数据库支持

```bash
envkit init  # 选择 3
```

## 🌏 支持的镜像源

### Node.js (npm/pnpm/yarn)
- **npmmirror** (淘宝镜像) - 推荐 ⭐
- **tencent** (腾讯云)
- **huawei** (华为云)

### Python (pip)
- **tsinghua** (清华大学) - 推荐 ⭐
- **aliyun** (阿里云)
- **ustc** (中科大)
- **douban** (豆瓣)

### Go
- **goproxy** (goproxy.cn) - 推荐 ⭐
- **aliyun** (阿里云)
- **qiniu** (七牛云)

### Rust (cargo)
- **ustc** (中科大) - 推荐 ⭐
- **sjtu** (上海交大)
- **tsinghua** (清华大学)
- **bytedance** (字节跳动)

## 📝 配置文件示例

```yaml
version: "1.0"
name: "Full Stack Development"

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

## 📚 文档

详细使用文档请查看 [docs/USAGE.md](docs/USAGE.md)

## 🛠️ 开发

### 环境要求

- Go 1.21 或更高版本

### 本地开发

```bash
# 克隆项目
git clone https://github.com/fusheng/envkit.git
cd envkit

# 安装依赖
go mod download

# 运行
go run cmd/envkit/main.go

# 构建
go build -o envkit cmd/envkit/main.go

# 多平台构建
./build.sh
```

### 项目结构

```
envkit/
├── cmd/
│   └── envkit/          # 主程序入口
├── internal/
│   ├── config/          # 配置文件解析
│   ├── detector/        # 系统检测
│   ├── mirror/          # 镜像源配置
│   ├── templates/       # 预设模板
│   └── ui/              # 用户界面
├── templates/           # YAML 模板文件
├── docs/               # 文档
├── build.sh            # 构建脚本
└── README.md
```

## 🎯 开发路线图

### v0.1.0 - MVP
- ✅ 系统检测
- ✅ 配置文件管理
- ✅ 国内镜像源配置
- ✅ 预设模板

### v0.2.0 (当前版本) - 自动化安装
- ✅ 自动安装编程语言 (Node.js/Python/Go/Rust)
- ✅ 自动安装开发工具 (Git/Docker/VSCode)
- ✅ Docker 容器管理 (PostgreSQL/Redis/MySQL/MongoDB)
- ✅ TUI 交互界面（彩色输出、进度条、表格）

### v0.3.0 (计划中)
- ⏸️ 单元测试覆盖
- ⏸️ dotfiles 管理
- ⏸️ Shell 环境配置
- ⏸️ 云端配置同步
- ⏸️ 插件系统

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

## 📄 许可证

MIT License - 详见 [LICENSE](LICENSE) 文件

## 🙏 致谢

感谢所有提供国内镜像源的组织和个人。

---

Made with ❤️ for Chinese Developers
