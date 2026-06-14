# EnvKit Scripts

自动化脚本集合，用于项目初始化、构建和部署。

---

## 可用脚本

### 1. init-gui.sh

**用途：** 快速初始化桌面客户端项目

**功能：**
- 检查系统依赖（Go, Node.js, npm）
- 安装 Wails CLI（如果未安装）
- 创建项目结构
- 生成基础代码
- 安装前端依赖

**使用方法：**
```bash
./scripts/init-gui.sh
```

**输出：**
```
envkit/
├── cmd/gui/main.go          # GUI 入口
├── frontend/                # 前端代码
│   ├── src/
│   ├── package.json
│   └── vite.config.js
└── wails.json               # Wails 配置
```

---

## 开发工作流

### 开发模式
```bash
# 1. 初始化项目
./scripts/init-gui.sh

# 2. 启动开发服务器
wails dev

# 3. 浏览器自动打开，修改代码即时生效
```

### 生产构建
```bash
# 构建当前平台
wails build

# 构建 macOS (Universal Binary)
wails build -platform darwin/universal

# 构建 Windows
wails build -platform windows/amd64

# 构建 Linux
wails build -platform linux/amd64
```

---

## 常见问题

### Q: 运行 init-gui.sh 提示找不到 wails 命令？

**A:** 需要将 Go bin 目录添加到 PATH：

```bash
# macOS/Linux
export PATH=$PATH:$(go env GOPATH)/bin

# 永久添加（zsh）
echo 'export PATH=$PATH:$(go env GOPATH)/bin' >> ~/.zshrc
source ~/.zshrc

# 永久添加（bash）
echo 'export PATH=$PATH:$(go env GOPATH)/bin' >> ~/.bashrc
source ~/.bashrc
```

### Q: npm install 失败？

**A:** 尝试更换镜像源：

```bash
npm config set registry https://registry.npmmirror.com
```

### Q: wails dev 启动失败？

**A:** 运行系统检查：

```bash
wails doctor
```

查看缺少哪些依赖并按提示安装。

---

## 脚本开发规范

### 新增脚本的要求

1. **文件命名**
   - 使用小写字母和连字符
   - 以 `.sh` 结尾（Shell 脚本）
   - 例如：`init-gui.sh`, `build-all.sh`

2. **文件头部**
   ```bash
   #!/bin/bash
   
   # 脚本简短描述
   # 用途：详细说明
   
   set -e  # 遇到错误立即退出
   ```

3. **输出规范**
   ```bash
   echo "✓ 成功信息"
   echo "❌ 错误信息"
   echo "⚠️  警告信息"
   echo "📦 正在处理..."
   echo "━━━━━━━━━━━━━━" # 分隔线
   ```

4. **权限**
   ```bash
   chmod +x scripts/your-script.sh
   ```

---

## 未来规划

### 待添加的脚本

- `build-all.sh` - 一键构建所有平台
- `package.sh` - 打包发布版本
- `test-gui.sh` - 运行 GUI 测试
- `clean.sh` - 清理构建产物
- `release.sh` - 自动发布流程

---

## 贡献

欢迎提交新的自动化脚本！

**提交前请确保：**
- 脚本可以在 macOS/Linux 上正常运行
- 添加必要的错误处理
- 提供清晰的输出信息
- 更新此 README
