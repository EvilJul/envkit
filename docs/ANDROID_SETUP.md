# Android 开发环境配置指南

EnvKit 提供了一键安装 Android SDK 的能力，通过国内镜像源加速，无需翻墙即可完成 Android 开发环境的搭建。

## 快速安装

```bash
# 方式 1：使用模板
envkit install -f templates/languages/android.yaml

# 方式 2：直接安装
envkit install android
```

## 默认安装内容

| 组件 | 版本/路径 |
|------|---------|
| cmdline-tools | 11076708 (最新稳定版) |
| platform-tools (adb) | 最新 |
| build-tools | 34.0.0 |
| platforms | android-34 |
| JDK | 21 (由 envkit 自动管理) |

## 安装目录

- macOS: `~/Library/Android/sdk`
- Linux: `~/Android/Sdk`
- Windows: `%LOCALAPPDATA%\Android\Sdk`

## 环境变量

安装完成后会自动配置：
- `ANDROID_HOME` - 指向 SDK 根目录
- `ANDROID_SDK_ROOT` - 同上（部分老版本工具使用）
- `PATH` 中追加：
  - `$ANDROID_HOME/platform-tools` (adb)
  - `$ANDROID_HOME/cmdline-tools/latest/bin` (sdkmanager)
  - `$ANDROID_HOME/build-tools/34.0.0` (aapt2, d8)
  - `$ANDROID_HOME/emulator` (emulator)

## 国内镜像源配置

```bash
# Android SDK 镜像源
envkit mirror android aliyun

# Gradle 全局镜像源
envkit mirror gradle aliyun
```

可用的镜像源：
- `aliyun` (推荐) - 阿里云
- `tencent` - 腾讯云
- `huawei` - 华为云
- `tsinghua` - 清华大学

## 验证安装

```bash
# 检查 adb 版本
adb --version

# 检查 sdkmanager
sdkmanager --version

# 查看已安装的 SDK 组件
sdkmanager --list_installed
```

## 卸载

```bash
# 卸载 Android SDK（会清理 ANDROID_HOME、PATH 等环境变量）
envkit uninstall android
```

## 常见问题

### Q: 安装过程中网络超时？
A: 默认使用阿里云镜像源，如遇问题可手动切换：
```bash
envkit mirror android huawei
```

### Q: 如何安装额外的 build-tools 版本？
A: 使用 sdkmanager：
```bash
sdkmanager "build-tools;33.0.2"
sdkmanager "platforms;android-33"
```

### Q: ANDROID_HOME 没有自动设置？
A: 重新加载 shell 配置：
```bash
source ~/.zshrc  # 或 ~/.bashrc
```

## 高级用法

### 自定义安装版本

修改 `templates/languages/android.yaml` 中的 tools 段落，目前暂不支持通过 YAML 自定义版本，默认为最新稳定版。如需自定义，请手动使用 sdkmanager 安装。

### 接受所有 License

安装过程中会自动接受所有 license。如果遇到问题：
```bash
yes | sdkmanager --licenses
```
