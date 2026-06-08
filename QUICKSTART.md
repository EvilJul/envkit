# EnvKit 快速开始指南

> v0.2.0 - 一键配置开发环境的跨平台CLI工具

## 🚀 5 分钟快速上手

### 步骤 1: 生成配置

```bash
envkit init
```

选择你需要的环境类型：
1. 🎨 **前端开发** - Node.js + pnpm
2. ⚙️  **后端开发** - Go + Python + Docker
3. 🔥 **全栈开发** - Node.js + Go + Python + Docker

### 步骤 2: 一键安装

```bash
envkit install
```

EnvKit 会自动：
- ✅ 安装所需的编程语言
- ✅ 配置国内镜像源（加速下载）
- ✅ 安装开发工具
- ✅ 启动数据库容器（如果需要）

### 步骤 3: 验证安装

```bash
envkit detect
```

查看已安装的工具和版本。

## 💡 常用命令

```bash
# 检测系统环境
envkit detect

# 单独配置镜像源
envkit mirror npm npmmirror
envkit mirror pip tsinghua

# 启动数据库
envkit docker start postgres 16
envkit docker start redis 7

# 查看运行的容器
envkit docker list

# 获取帮助
envkit help
```

## 🌟 新功能 (v0.2.0)

- ✅ **自动安装语言** - 不再需要手动下载安装包
- ✅ **自动安装工具** - Git/Docker/VSCode 一键安装
- ✅ **Docker 容器管理** - 一键启动开发数据库
- ✅ **TUI 界面** - 彩色输出、进度条、表格显示

## 📚 更多文档

- [完整使用指南](docs/USAGE.md)
- [架构设计](docs/ARCHITECTURE.md)
- [贡献指南](CONTRIBUTING.md)
- [更新日志](CHANGELOG.md)

---

**快速开始就是这么简单！** 🎉
