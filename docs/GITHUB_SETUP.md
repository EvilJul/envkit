# 推送到 GitHub 指南

## 步骤 1: 在 GitHub 上创建新仓库

1. 访问 https://github.com/new
2. 填写仓库信息：
   - **Repository name**: `envkit`
   - **Description**: `🚀 一键配置开发环境的跨平台CLI工具 - 专为中国开发者优化`
   - **Public** 或 **Private**: 选择你想要的可见性
   - **不要勾选** "Initialize this repository with a README" (我们已经有了)

3. 点击 "Create repository"

## 步骤 2: 推送代码到 GitHub

GitHub 会显示命令，或者你可以直接运行：

```bash
# 添加远程仓库（将 YOUR_USERNAME 替换为你的 GitHub 用户名）
git remote add origin https://github.com/YOUR_USERNAME/envkit.git

# 或使用 SSH（推荐）
git remote add origin git@github.com:YOUR_USERNAME/envkit.git

# 推送代码
git push -u origin main
```

## 步骤 3: 创建首个 Release（可选）

在 GitHub 仓库页面：

1. 点击 "Releases" → "Create a new release"
2. 填写：
   - **Tag**: `v0.1.0`
   - **Release title**: `EnvKit v0.1.0 - MVP Release`
   - **Description**: 复制 CHANGELOG.md 中的内容
3. 点击 "Publish release"

GitHub Actions 会自动构建多平台二进制文件并附加到 Release。

## 步骤 4: 验证

推送成功后，你应该能看到：

- ✅ 完整的源代码
- ✅ README.md 在首页正确显示
- ✅ GitHub Actions 开始运行 CI
- ✅ 如果创建了 tag，会自动构建 Release

## 当前仓库状态

```bash
# 查看当前状态
git status

# 查看提交历史
git log --oneline

# 查看远程仓库
git remote -v
```

## 本地仓库信息

- **分支**: main
- **提交数**: 3
- **文件数**: 33
- **代码行数**: ~2860 行（包括注释和文档）

## 下一步

推送成功后，你可以：

1. 在 GitHub 仓库设置中添加话题标签：
   - `golang`
   - `cli`
   - `developer-tools`
   - `china`
   - `mirror`
   - `development-environment`

2. 添加仓库描述和网站链接

3. 配置 GitHub Pages（如果需要文档网站）

4. 启用 Issues 和 Discussions

## 常见问题

### Q: 如何更改远程仓库地址？

```bash
git remote set-url origin 新的仓库地址
```

### Q: 如何推送到不同的分支？

```bash
git push origin branch-name
```

### Q: 如何删除远程仓库的连接？

```bash
git remote remove origin
```
