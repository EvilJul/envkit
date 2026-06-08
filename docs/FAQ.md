# 常见问题 (FAQ)

## 安装相关

### Q: 支持哪些操作系统？
**A:** EnvKit 支持：
- macOS (Intel/Apple Silicon)
- Linux (Ubuntu/Debian/CentOS/Fedora/Arch)
- Windows 10/11

### Q: 需要管理员权限吗？
**A:** 部分操作需要：
- Linux: 安装软件包需要 sudo
- macOS: 首次安装 Homebrew 需要密码
- Windows: winget 可能需要管理员权限

### Q: 可以离线使用吗？
**A:** 不可以。EnvKit 需要网络连接来：
- 下载语言安装包
- 拉取 Docker 镜像
- 配置镜像源

## 语言安装

### Q: 为什么安装失败？
**A:** 常见原因：
1. **网络问题** - 检查网络连接
2. **权限不足** - 某些操作需要 sudo
3. **已安装其他版本** - 可能有冲突
4. **磁盘空间不足** - 确保有足够空间

### Q: 支持安装特定版本吗？
**A:** 是的，在配置文件中指定版本：
```yaml
languages:
  - name: node
    version: "18.x"  # 指定 18.x 版本
```

### Q: 如何卸载已安装的语言？
**A:** EnvKit 目前不提供卸载功能，请使用系统的包管理器：
- macOS: `brew uninstall node`
- Linux: `apt remove nodejs`
- Windows: `winget uninstall NodeJS`

## 镜像源配置

### Q: 镜像源配置后不生效？
**A:** 检查步骤：
1. 验证配置：
   ```bash
   npm config get registry  # npm
   pip config list          # pip
   go env GOPROXY          # go
   ```
2. 重启终端
3. 检查文件权限

### Q: 如何恢复官方源？
**A:** 
- **npm**: `npm config set registry https://registry.npmjs.org/`
- **pip**: 删除 `~/.pip/pip.conf`
- **go**: `go env -w GOPROXY=https://proxy.golang.org,direct`

### Q: 支持企业内部镜像源吗？
**A:** 目前不支持，但可以在配置后手动修改配置文件。

## Docker 容器

### Q: 为什么容器启动失败？
**A:** 检查：
1. Docker 是否运行：`docker info`
2. 端口是否被占用：`lsof -i :5432`
3. 磁盘空间是否充足

### Q: 如何修改容器密码？
**A:** 删除容器重新创建：
```bash
envkit docker remove envkit-postgres
# 然后修改 manager.go 中的默认密码或传入环境变量
```

### Q: 容器数据会丢失吗？
**A:** 不会。EnvKit 使用数据卷存储数据：
- `envkit-postgres-data`
- `envkit-redis-data`
- `envkit-mysql-data`
- `envkit-mongodb-data`

删除容器时会询问是否删除数据卷。

### Q: 如何连接容器中的数据库？
**A:** 
```bash
# PostgreSQL
psql -h localhost -U postgres -d postgres
# 密码: postgres

# Redis
redis-cli -h localhost

# MySQL
mysql -h localhost -u root -p
# 密码: mysql

# MongoDB
mongosh mongodb://localhost:27017
```

## 性能和体验

### Q: 安装速度慢？
**A:** 
1. 已配置国内镜像源
2. 首次安装需要下载较大文件
3. 后续安装会快很多

### Q: 如何查看详细日志？
**A:** EnvKit v0.2.0 提供实时进度显示，详细日志功能在开发中。

## 故障排除

### Q: 命令找不到？
**A:** 确保：
1. envkit 在 PATH 中
2. 使用 `./envkit` 或绝对路径
3. 文件有执行权限：`chmod +x envkit`

### Q: 权限被拒绝？
**A:** 
- Linux/macOS: 使用 `sudo envkit install`
- Windows: 以管理员身份运行

### Q: 报告 bug？
**A:** 在 GitHub 创建 Issue:
https://github.com/fusheng/envkit/issues

包含：
- 操作系统和版本
- EnvKit 版本
- 完整错误信息
- 复现步骤

## 其他

### Q: 支持代理吗？
**A:** EnvKit 会使用系统代理设置：
- HTTP_PROXY
- HTTPS_PROXY
- NO_PROXY

### Q: 有GUI版本吗？
**A:** 目前只有 CLI 版本，GUI 在 v0.3.0+ 路线图中。

### Q: 如何更新 EnvKit？
**A:** 
```bash
# Go 安装
go install github.com/fusheng/envkit/cmd/envkit@latest

# 二进制安装
# 下载新版本覆盖旧版本
```

---

**还有问题？** 欢迎在 [GitHub Issues](https://github.com/fusheng/envkit/issues) 提问！
