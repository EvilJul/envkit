# ESP-IDF 全平台环境安装配置与部署教程

ESP-IDF（乐鑫物联网开发框架）是乐鑫面向 ESP32 系列芯片的官方 SDK。乐鑫推出了全新的 **ESP-IDF 安装管理器 (ESP-IDF Installation Manager, 简称 EIM)**，支持在 Windows、macOS 与 Linux 上通过各平台原生包管理器一键安装工具链、Python 环境与 ESP-IDF 本体，避免了手动配置依赖的繁琐过程。

---

## 1. 原生包管理器安装 EIM

根据你的操作系统，使用对应的包管理器命令进行安装。

### Windows
可以使用 Windows 包管理器 `winget` 一键部署：
```powershell
# 安装带图形化界面 (GUI) 的版本
winget install Espressif.EIM

# 或者仅安装命令行 (CLI) 版本
winget install Espressif.EIM-CLI
```

### macOS
使用 Homebrew 进行安装：
```bash
# 添加乐鑫官方 Homebrew Tap
brew tap espressif/eim

# 安装带图形化界面 (GUI) 的版本
brew install --cask eim-gui

# 或者仅安装命令行 (CLI) 版本
brew install eim
```

### Linux (Debian / Ubuntu / Deepin)
使用 APT 官方软件源安装：
```bash
# 1. 添加 EIM APT 软件源到系统
echo "deb [trusted=yes] https://dl.espressif.com/dl/eim/apt/ stable main" | \
    sudo tee /etc/apt/sources.list.d/espressif.list

# 2. 更新软件包索引
sudo apt update

# 3. 安装命令行版本
sudo apt install eim-cli

# 4. (可选) 安装图形化界面版本
sudo apt install eim
```

### Linux (CentOS / RHEL / Fedora)
使用 RPM 软件源安装：
```bash
# 1. 安装 RPM 仓库配置
sudo dnf install https://dl.espressif.com/dl/eim/rpm/eim-repo-latest.noarch.rpm

# 2. 安装命令行版本
sudo dnf install eim-cli

# 3. (可选) 安装图形化界面版本
sudo dnf install eim
```

> [!NOTE]
> EIM 会自动在后台为你下载并配置符合要求的 CMake、Git、Python 以及编译芯片所需的 Xtensa 和 RISC-V 交叉编译器，无需手动干预。

---

## 2. 导出环境变量并运行

安装完成后，你需要在使用开发工具链的命令行会话中，导出 EIM 配置好的环境变量。

```bash
# 激活 ESP-IDF 环境变量
# 如果是 macOS/Linux，运行以下命令（默认安装在 ~/.espressif 目录下）：
. $HOME/.espressif/export.sh

# 如果是 Windows, 打开 ESP-IDF PowerShell/CMD 终端，或者运行：
# %USERPROFILE%\.espressif\export.bat
```

验证是否安装成功：
```bash
idf.py --version
```
输出包含 `ESP-IDF vX.X.X` 即代表安装成功。

---

## 3. 创建与部署首个工程 (Hello World)

激活环境变量后，即可进行项目的创建、配置、构建与烧录。

### Step 1: 创建新项目
```bash
# 创建一个名为 hello_world 的新项目
idf.py create-project hello_world
cd hello_world
```

### Step 2: 设定目标芯片 (Target)
ESP-IDF 支持乐鑫多款芯片，如 ESP32、ESP32-S2、ESP32-S3、ESP32-C3、ESP32-C6 等。在构建前需选择对应芯片：
```bash
# 设置目标芯片为 ESP32-C6 (支持 Wi-Fi 6, BLE, Zigbee)
idf.py set-target esp32c6
```

### Step 3: 配置项目外设 (可选)
如果需要微调 Wi-Fi 账号、系统堆栈大小等，可以打开配置菜单：
```bash
idf.py menuconfig
```

### Step 4: 编译项目
```bash
# 启动 CMake & Ninja 构建工程
idf.py build
```

### Step 5: 烧录并打开串口监视器
将 ESP32 开发板通过 USB 连接至电脑，执行以下命令进行一键烧录并实时查看串口日志：
```bash
# -p 指定你的开发板串口设备
# Windows 通常为 COM3 等
# macOS/Linux 为 /dev/ttyUSB0 或 /dev/cu.usbserial-xxx
idf.py -p /dev/ttyUSB0 flash monitor
```
> [!TIP]
> 串口监视器中，使用快捷键 `Ctrl + ]` 可以退出监控模式返回终端。

---

## 4. IDE 插件支持

如果你不喜欢在命令行中操作，可以在安装好 EIM 后，使用以下编辑器插件进行可视化开发：

| 编辑器 / IDE | 推荐插件名称 | 主要功能特点 |
| :--- | :--- | :--- |
| **VS Code** | `espressif.esp-idf-extension` | 支持一键配置、菜单选项修改、快捷键烧录与分区表可视化编辑 |
| **CLion** | `CLion 内置 CMake` | 拥有极强的高级代码导航、语法高亮、自动提示与外设硬件调试能力 |
| **Eclipse** | `IDF Eclipse Plugin` | 面向 Eclipse 传统用户的官方插件支持 |

---

## 5. 面向专用场景的扩展框架

乐鑫在 ESP-IDF 之上针对特定场景构建了更高级别的 SDK，你可以根据产品需求进行集成：

* **ESP-Matter**： Matter 智能家居连接协议栈，与乐鑫芯片深度对齐。
* **ESP-Zigbee-SDK**： 针对 ESP32-C6, ESP32-H2 等芯片的官方 Zigbee 3.0 协议支持。
* **ESP-BROOKESIA**： 针对带屏 AIoT 设备的智能人机交互（HMI）开发框架。
* **ESP-WHO**： 面向摄像头芯片的轻量级人脸识别与图像检测库。
