# 跨平台兼容性优化总结

本次优化针对 Windows、Linux 和 macOS 三个平台的兼容性问题进行了全面修复。

## 修复概览

### ✅ HIGH 优先级问题（已修复）

#### 1. Windows 中文字符编码问题
**问题**: Windows 控制台使用 GBK 编码，导致中文输出乱码

**修复内容**:
- 新增 `internal/installer/encoding.go` 文件
- 实现 `NewWindowsConsoleWriter()` 函数进行 UTF-8 到 GBK 的编码转换
- 修改 `runCommand()` 函数，在 Windows 平台自动应用编码转换
- 添加依赖 `golang.org/x/text v0.14.0`

**影响文件**:
- `go.mod`
- `internal/installer/encoding.go` (新建)
- `internal/installer/tool.go`

#### 2. Linux 系统工具依赖检测缺失
**问题**: 在最小化 Linux 系统上缺少必需工具（curl, tar, unzip, bash, gpg）导致安装失败

**修复内容**:
- 新增 `internal/installer/dependencies.go` 文件
- 实现 `CheckAndInstallDependencies()` 自动检测和安装系统依赖
- 实现 `detectPackageManager()` 自动识别包管理器（apt, yum, dnf, pacman, brew）
- 在所有安装函数中添加依赖检测调用

**修复位置**:
- `internal/installer/dependencies.go` (新建)
- `internal/installer/language.go`: `installWithFnm()`, `installWithUv()`, `installFromSourceDarwin()`, `installFromSource()`
- `internal/installer/tool.go`: `installOnLinux()` (Docker), `installOnDarwin()` (VSCode), `installOnLinux()` (VSCode), `installUnix()` (Miniconda)

#### 3. Windows/Unix 平台路径分离
**问题**: 使用 bash 脚本安装方式在 Windows 上不适用

**修复内容**:
- 在 `installWithFnm()` 和 `installWithUv()` 开头添加 Windows 平台检测
- Windows 平台返回明确错误提示，引导用户使用 winget 或官网安装包

**影响文件**:
- `internal/installer/language.go`

---

### ✅ MEDIUM 优先级问题（已修复）

#### 4. Windows PATH 环境变量持久化
**问题**: Unix shell 配置文件方式不适用于 Windows

**修复内容**:
- 新增 `persistPathEnvWindows()` 函数使用 `setx` 命令修改 Windows 用户环境变量
- 修改 `PersistPathEnv()` 函数，根据平台自动选择持久化方式
- 新增 `getShellConfigFiles()` 函数统一获取 shell 配置文件列表

**影响文件**:
- `internal/installer/path.go`

#### 5. 临时文件路径硬编码
**问题**: 硬编码 `/tmp` 路径仅适用于 Unix，Windows 使用不同临时目录

**修复内容**:
- 所有临时文件路径改为使用 `os.TempDir()`
- 修复的文件和函数：
  - `language.go`: `installFromSourceDarwin()`, `installFromSource()`
  - `tool.go`: `installOnDarwin()` (VSCode), `installOnLinux()` (VSCode), `installUnix()` (Miniconda), `installWindows()` (Miniconda)

**影响文件**:
- `internal/installer/language.go`
- `internal/installer/tool.go`

#### 6. macOS 默认 shell 检测
**问题**: macOS Catalina+ 默认使用 zsh，但代码未区分用户默认 shell

**修复内容**:
- 新增 `getDefaultShell()` 函数检测当前用户默认 shell
- 修改 `getShellConfigFiles()` 优先配置默认 shell 的配置文件
- 支持从 `$SHELL` 环境变量和 `dscl` 命令获取 shell 信息

**影响文件**:
- `internal/installer/path.go`

#### 7. Windows winget 版本检测
**问题**: 未检测 winget 是否存在及其版本

**修复内容**:
- 修改 `installWithWinget()` 函数
- 添加 winget 存在性检查
- 添加版本检测并输出到控制台
- 提供更友好的错误提示

**影响文件**:
- `internal/installer/tool.go`

#### 8. 符号链接处理
**问题**: `os.RemoveAll()` 会删除符号链接的目标，而非链接本身

**修复内容**:
- 修改 `UninstallComponent()` 函数
- 使用 `os.Lstat()` 检测符号链接
- 对符号链接使用 `os.Remove()`，普通文件/目录使用 `os.RemoveAll()`

**影响文件**:
- `internal/installer/manifest.go`

#### 9. 路径转换
**问题**: 配置中的路径未考虑 Windows 反斜杠

**修复内容**:
- 在 `RecordInstallation()` 中使用 `filepath.FromSlash()` 转换路径
- 使用 `filepath.Clean()` 规范化路径

**影响文件**:
- `internal/installer/manifest.go`

#### 10. Shell 配置文件统一管理
**问题**: 多处硬编码 shell 配置文件列表

**修复内容**:
- 统一使用 `getShellConfigFiles()` 函数
- 修复位置：
  - `CleanShellConfigs()`
  - `persistFnmShellIntegration()`

**影响文件**:
- `internal/installer/manifest.go`
- `internal/installer/path.go`

#### 11. 文件权限注释
**问题**: Unix 风格权限在 Windows 上被忽略，但代码未说明

**修复内容**:
- 在 `installFromSource()` 中添加注释说明 Windows 忽略权限参数

**影响文件**:
- `internal/installer/language.go`

---

## 新增文件

1. **internal/installer/encoding.go** - Windows 控制台编码转换
2. **internal/installer/dependencies.go** - Linux 系统依赖自动检测和安装

## 新增依赖

```go
golang.org/x/text v0.14.0
```

## 测试结果

✅ 编译成功
✅ 基本命令运行正常
✅ 版本显示正确

## 建议后续优化

### 长期优化（未在本次修复）

1. **国际化支持** - 根据系统语言环境调整错误信息
2. **macOS 架构检测增强** - 处理 Rosetta 2 转译和 Universal Binary
3. **Linux 发行版特性检测** - 针对不同发行版（如 RHEL 系列的 EPEL）提供特定支持
4. **macOS Homebrew 路径动态获取** - 使用 `brew --prefix` 替代硬编码路径

## 关键改进点

1. **平台感知**: 所有关键操作都根据 `runtime.GOOS` 选择合适的实现
2. **依赖前置**: 在使用系统工具前检测并自动安装
3. **编码正确**: Windows 控制台输出自动转换编码
4. **路径规范**: 统一使用 `filepath` 包处理路径，支持跨平台
5. **错误友好**: 提供明确的错误信息和解决建议

## 验证建议

### Windows 平台
- [ ] 测试中文路径和中文输出
- [ ] 测试 PATH 环境变量持久化（需重启终端验证）
- [ ] 测试 winget 安装流程

### Linux 平台
- [ ] 在最小化系统（如 Docker Alpine）上测试依赖自动安装
- [ ] 测试不同包管理器（apt, yum, dnf, pacman）
- [ ] 测试临时文件在 `/tmp` 不可用时的降级方案

### macOS 平台
- [ ] 在 Intel 和 Apple Silicon 上测试
- [ ] 测试 bash 和 zsh 环境
- [ ] 测试 Homebrew 路径检测

---

**优化完成时间**: 2026-06-09
**优化者**: Claude Code (Opus 4.7)
