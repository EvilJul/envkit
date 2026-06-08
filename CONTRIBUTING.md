# 贡献指南

感谢你考虑为 EnvKit 做贡献！

## 🚀 开始之前

### 环境要求

- Go 1.21 或更高版本
- Git
- Docker (可选，用于测试容器功能)

### 克隆项目

```bash
git clone https://github.com/fusheng/envkit.git
cd envkit
```

## 🛠️ 开发流程

### 1. 本地开发

```bash
# 安装依赖
go mod download

# 运行程序
go run cmd/envkit/main.go detect

# 构建
go build -o envkit cmd/envkit/main.go

# 运行
./envkit help
```

### 2. 代码结构

```
envkit/
├── cmd/envkit/          # 主程序入口
├── internal/
│   ├── config/          # 配置文件解析
│   ├── detector/        # 系统检测
│   ├── docker/          # Docker 容器管理 (v0.2.0)
│   ├── installer/       # 语言和工具安装器 (v0.2.0)
│   ├── mirror/          # 镜像源配置
│   ├── templates/       # 预设模板
│   └── ui/              # TUI 界面组件 (v0.2.0)
├── templates/           # YAML 模板文件
└── docs/               # 文档
```

### 3. 添加新功能

#### 添加新的语言支持

1. 在 `internal/installer/language.go` 中添加新的 Installer 实现
2. 在 `GetInstaller()` 中注册新语言
3. 更新文档

#### 添加新的工具支持

1. 在 `internal/installer/tool.go` 中添加新的 Installer 实现
2. 在 `GetToolInstaller()` 中注册新工具
3. 更新文档

#### 添加新的数据库支持

1. 在 `internal/docker/manager.go` 中添加 `Start<Database>()` 方法
2. 在 `handleDocker()` 和 `handleInstall()` 中添加对应的 case
3. 更新文档

### 4. 代码规范

- 使用 `gofmt` 格式化代码
- 遵循 Go 的命名约定
- 添加适当的注释（中文）
- 错误处理要完善

### 5. 测试

```bash
# 运行所有测试
go test ./...

# 测试特定模块
go test ./internal/installer/

# 手动测试
./envkit detect
./envkit init
./envkit docker list
```

   提交信息格式：
   - `feat:` 新功能
   - `fix:` Bug 修复
   - `docs:` 文档更新
   - `style:` 代码格式调整
   - `refactor:` 重构
   - `test:` 测试相关
   - `chore:` 构建/工具链相关

5. **推送到 GitHub**
   ```bash
   git push origin feature/my-feature
   ```

6. **创建 Pull Request**
   - 描述你的更改
   - 关联相关 Issue
   - 等待代码审查

## 开发指南

### 环境准备

```bash
# 安装 Go 1.21+
go version

# 克隆项目
git clone https://github.com/fusheng/envkit.git
cd envkit

# 安装依赖
go mod download
```

### 项目结构

```
envkit/
├── cmd/envkit/          # 主程序入口
├── internal/
│   ├── config/          # 配置解析
│   ├── detector/        # 系统检测
│   ├── installer/       # 安装器（待实现）
│   ├── mirror/          # 镜像源配置
│   ├── templates/       # 模板管理
│   └── ui/              # UI组件（待实现）
├── templates/           # YAML模板
└── docs/               # 文档
```

### 添加新的镜像源

1. 在 `internal/mirror/registry.go` 中添加镜像源URL
2. 创建对应的配置器（如 `internal/mirror/xxx.go`）
3. 实现 `Configure()` 和 `Verify()` 方法
4. 在主程序中添加命令处理

示例：
```go
// internal/mirror/maven.go
package mirror

type MavenConfigurator struct {
    registry *Registry
}

func NewMavenConfigurator(registry *Registry) *MavenConfigurator {
    return &MavenConfigurator{registry: registry}
}

func (m *MavenConfigurator) Configure(mirror string) error {
    // 实现配置逻辑
    return nil
}
```

### 添加新的预设模板

1. 在 `templates/` 目录创建新的 YAML 文件
2. 在 `internal/templates/templates.go` 中添加模板类型
3. 使用 `//go:embed` 嵌入模板文件
4. 在 `List()` 方法中添加模板信息

### 运行测试

```bash
# 运行所有测试
go test ./...

# 运行特定包的测试
go test ./internal/config

# 带覆盖率
go test -cover ./...
```

### 本地构建

```bash
# 构建当前平台
go build -o envkit cmd/envkit/main.go

# 多平台构建
./build.sh

# 运行
./envkit version
```

## 代码规范

### Go 代码风格

- 使用 `gofmt` 格式化代码
- 遵循 Go 官方代码规范
- 导出的函数和类型必须有注释
- 错误处理要完整

示例：
```go
// DetectSystem 获取完整的系统信息
func DetectSystem() (*SystemInfo, error) {
    info := &SystemInfo{
        OS:   DetectOS(),
        Arch: DetectArchitecture(),
    }
    
    if info.OS == OSLinux {
        dist, err := detectLinuxDistribution()
        if err != nil {
            return nil, fmt.Errorf("检测Linux发行版失败: %w", err)
        }
        info.Distribution = dist
    }
    
    return info, nil
}
```

### 错误处理

- 使用 `fmt.Errorf` 包装错误
- 为用户友好的错误信息
- 记录详细的调试信息

```go
if err != nil {
    return fmt.Errorf("配置npm镜像源失败: %w", err)
}
```

### 注释规范

- 公开的函数/类型必须有文档注释
- 复杂逻辑需要添加行内注释
- 使用中文注释（面向中国开发者）

## 文档

### 更新文档

如果你的更改影响到用户使用，请同时更新：

- `README.md` - 主文档
- `docs/USAGE.md` - 使用指南
- `CHANGELOG.md` - 变更日志

### 文档风格

- 使用清晰的标题层级
- 提供代码示例
- 使用 emoji 增强可读性（适度）
- 中文文档为主

## 社区准则

- 尊重他人
- 保持友善和专业
- 欢迎新手
- 鼓励提问
- 给予建设性反馈

## 获取帮助

- 创建 Issue 提问
- 查看现有 Issue 和 PR
- 阅读文档和代码

## 许可证

提交代码即表示你同意将代码以 MIT 许可证发布。

---

再次感谢你的贡献！ 🎉
