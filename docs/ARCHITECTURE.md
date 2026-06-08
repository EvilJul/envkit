# EnvKit 架构设计文档

> v0.2.0 架构说明

## 📐 整体架构

```
┌─────────────────────────────────────────────────────────────┐
│                     CLI 命令层 (main.go)                      │
│  init | install | detect | mirror | docker | version | help │
└──────────────┬──────────────────────────────┬────────────────┘
               │                              │
    ┌──────────▼──────────┐      ┌───────────▼────────────┐
    │   核心业务模块        │      │    UI 表现层           │
    │                     │      │                        │
    │  • installer        │      │  • colors              │
    │  • docker           │      │  • progress            │
    │  • mirror           │      │  • table               │
    │  • detector         │      │                        │
    │  • config           │      └────────────────────────┘
    │  • templates        │
    └──────────┬──────────┘
               │
    ┌──────────▼──────────┐
    │   系统调用层          │
    │                     │
    │  • exec.Command     │
    │  • os/file          │
    │  • Docker API       │
    └─────────────────────┘
```

## 🧩 模块设计

### 1. CLI 命令层 (`cmd/envkit/main.go`)

**职责：**
- 解析命令行参数
- 路由到对应的处理函数
- 统一错误处理和退出

**主要函数：**
- `main()` - 程序入口
- `handleInit()` - 处理 init 命令
- `handleInstall()` - 处理 install 命令
- `handleDetect()` - 处理 detect 命令
- `handleMirror()` - 处理 mirror 命令
- `handleDocker()` - 处理 docker 命令
- `printUsage()` - 显示帮助信息

### 2. 安装器模块 (`internal/installer/`)

**设计模式：** 策略模式 + 工厂模式

**接口定义：**
```go
type LanguageInstaller interface {
    Install(version string) error
    IsInstalled() bool
    GetVersion() string
}

type ToolInstaller interface {
    Install() error
    IsInstalled() bool
    GetVersion() string
}
```

**实现：**
- `NodeInstaller` - Node.js 安装器
- `PythonInstaller` - Python 安装器
- `GoInstaller` - Go 安装器
- `RustInstaller` - Rust 安装器
- `GitInstaller` - Git 安装器
- `DockerInstaller` - Docker 安装器
- `VSCodeInstaller` - VSCode 安装器

**跨平台策略：**
```
Linux   → fnm/uv/官方包/apt
macOS   → brew
Windows → winget
```

### 3. Docker 管理模块 (`internal/docker/`)

**职责：**
- 管理数据库容器的生命周期
- 处理容器配置和数据卷
- 提供容器操作接口

**主要方法：**
- `StartPostgreSQL()` - 启动 PostgreSQL
- `StartRedis()` - 启动 Redis
- `StartMySQL()` - 启动 MySQL
- `StartMongoDB()` - 启动 MongoDB
- `StopContainer()` - 停止容器
- `RemoveContainer()` - 删除容器
- `ListContainers()` - 列出容器

**容器命名规范：**
```
envkit-<database>
例如: envkit-postgres, envkit-redis
```

**数据卷命名规范：**
```
envkit-<database>-data
例如: envkit-postgres-data
```

### 4. UI 模块 (`internal/ui/`)

**职责：**
- 提供统一的输出格式
- 彩色输出和进度显示
- 表格渲染

**组件：**

#### colors.go
- `Success()` - 成功消息（绿色 ✓）
- `Error()` - 错误消息（红色 ✗）
- `Warning()` - 警告消息（黄色 ⚠）
- `Info()` - 信息消息（蓝色 ℹ）
- `Bold()`, `Green()`, `Red()`, 等 - 文本装饰

#### progress.go
- `ProgressBar` - 进度条（带 ETA）
- `Spinner` - 旋转加载器

#### table.go
- `Table` - 表格渲染器
- `PrintHeader()` - 打印标题
- `PrintSection()` - 打印分区

### 5. 镜像源模块 (`internal/mirror/`)

**职责：**
- 管理各语言的镜像源配置
- 提供镜像源注册表

**configurator 模式：**
```go
type Configurator interface {
    Configure(mirrorName string) error
}
```

**实现：**
- `NPMConfigurator` - npm/pnpm/yarn
- `PipConfigurator` - pip
- `GoConfigurator` - GOPROXY
- `RustConfigurator` - cargo

### 6. 检测模块 (`internal/detector/`)

**职责：**
- 检测操作系统和架构
- 检测已安装的语言和工具
- 检测包管理器

**主要功能：**
- `DetectSystem()` - 系统信息检测
- `DetectLanguages()` - 语言检测
- `DetectTools()` - 工具检测
- `DetectPackageManagers()` - 包管理器检测

### 7. 配置模块 (`internal/config/`)

**职责：**
- 解析 YAML 配置文件
- 提供配置数据结构

**数据结构：**
```go
type Config struct {
    Version   string
    Name      string
    Languages []Language
    Tools     []string
    Databases []Database
    Shell     Shell
}
```

### 8. 模板模块 (`internal/templates/`)

**职责：**
- 管理预设模板
- 提供模板列表和获取接口

**预设模板：**
- Frontend (前端开发)
- Backend (后端开发)
- Fullstack (全栈开发)

## 🔄 数据流

### install 命令流程

```
用户执行 envkit install
         │
         ▼
    解析配置文件
         │
         ▼
    检测系统信息
         │
         ├─→ 安装语言
         │   ├─ 检查是否已安装
         │   ├─ 选择安装方式
         │   ├─ 执行安装
         │   └─ 配置镜像源
         │
         ├─→ 安装工具
         │   ├─ 检查是否已安装
         │   ├─ 选择安装方式
         │   └─ 执行安装
         │
         └─→ 启动容器
             ├─ 检查 Docker 状态
             ├─ 检查容器是否存在
             └─ 创建/启动容器
```

### docker 命令流程

```
用户执行 envkit docker start postgres 16
                │
                ▼
         检查 Docker 运行状态
                │
                ▼
         检查容器是否已存在
                │
        ┌───────┴───────┐
        │               │
     已存在          不存在
        │               │
    启动容器        创建容器
        │               │
        └───────┬───────┘
                ▼
         显示连接信息
```

## 🎯 设计原则

### 1. 接口隔离
每个模块都有清晰的接口定义，降低耦合度。

### 2. 单一职责
每个模块只负责一个特定功能。

### 3. 开闭原则
通过接口和工厂模式，易于扩展新的语言和工具支持。

### 4. 跨平台抽象
通过条件编译和运行时检测实现跨平台支持。

### 5. 用户体验优先
- 彩色输出提高可读性
- 进度反馈增强交互性
- 错误提示清晰友好

## 🔌 扩展点

### 添加新语言

1. 实现 `LanguageInstaller` 接口
2. 在 `GetInstaller()` 中注册
3. 更新文档

### 添加新工具

1. 实现 `ToolInstaller` 接口
2. 在 `GetToolInstaller()` 中注册
3. 更新文档

### 添加新数据库

1. 在 `ContainerManager` 中添加 `Start<DB>()` 方法
2. 在命令处理中添加对应的 case
3. 更新文档

### 添加新镜像源

1. 在 `Registry` 中注册新镜像
2. 更新 `Configurator` 实现
3. 更新文档

## 📝 代码约定

### 命名规范

- **包名：** 小写，简短，描述性
- **接口名：** 以 -er 结尾（Installer, Configurator）
- **函数名：** 驼峰命名，动词开头
- **常量：** 大写字母，下划线分隔

### 错误处理

- 使用 `fmt.Errorf()` 包装错误
- 在适当的层级处理错误
- 向用户提供清晰的错误信息

### 注释规范

- 包级别：在 package 声明前说明包的用途
- 导出函数：必须有注释
- 复杂逻辑：添加行内注释说明

## 🔒 安全考虑

1. **命令注入防护：** 使用 `exec.Command()` 分离参数
2. **权限检查：** 某些操作需要提示用户权限
3. **数据验证：** 验证用户输入和配置文件
4. **密码管理：** 容器默认密码应提示用户修改

---

**版本：** v0.2.0  
**更新日期：** 2026-06-08
