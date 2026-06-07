# 贡献指南

感谢你对 EnvKit 项目的关注！我们欢迎所有形式的贡献。

## 贡献方式

### 报告 Bug

如果你发现了 bug，请创建一个 Issue，包含以下信息：

- 操作系统和版本
- EnvKit 版本 (`envkit version`)
- 复现步骤
- 期望行为
- 实际行为
- 相关日志或截图

### 提出新功能

如果你有好的想法，欢迎创建 Feature Request Issue，包含：

- 功能描述
- 使用场景
- 预期效果
- 可能的实现方案（可选）

### 提交代码

1. **Fork 项目**
   ```bash
   git clone https://github.com/你的用户名/envkit.git
   cd envkit
   ```

2. **创建特性分支**
   ```bash
   git checkout -b feature/my-feature
   ```

3. **编写代码**
   - 遵循项目代码风格
   - 添加必要的注释
   - 确保代码通过测试

4. **提交更改**
   ```bash
   git add .
   git commit -m "feat: 添加某某功能"
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
