#!/bin/bash

# EnvKit GitHub 快速推送脚本

echo "🚀 EnvKit GitHub 推送助手"
echo ""

# 检查是否已有远程仓库
if git remote get-url origin > /dev/null 2>&1; then
    echo "✓ 检测到已配置的远程仓库:"
    git remote -v
    echo ""
    read -p "是否继续推送? (y/n): " confirm
    if [ "$confirm" != "y" ]; then
        echo "已取消"
        exit 0
    fi
else
    echo "未检测到远程仓库配置"
    echo ""
    echo "请先在 GitHub 创建新仓库，然后输入仓库地址:"
    echo ""
    echo "HTTPS 格式: https://github.com/YOUR_USERNAME/envkit.git"
    echo "SSH 格式 (推荐): git@github.com:YOUR_USERNAME/envkit.git"
    echo ""
    read -p "请输入仓库地址: " repo_url

    if [ -z "$repo_url" ]; then
        echo "❌ 仓库地址不能为空"
        exit 1
    fi

    echo ""
    echo "添加远程仓库..."
    git remote add origin "$repo_url"
    echo "✓ 远程仓库已添加"
fi

echo ""
echo "📊 当前仓库状态:"
echo "  分支: $(git branch --show-current)"
echo "  提交数: $(git rev-list --count HEAD)"
echo "  最新提交: $(git log -1 --oneline)"
echo ""

read -p "确认推送到 GitHub? (y/n): " confirm
if [ "$confirm" != "y" ]; then
    echo "已取消"
    exit 0
fi

echo ""
echo "🚀 开始推送..."
if git push -u origin main; then
    echo ""
    echo "✅ 推送成功!"
    echo ""
    echo "下一步:"
    echo "1. 访问你的 GitHub 仓库查看代码"
    echo "2. 创建 Release (可选): 在仓库页面点击 'Releases' → 'Create a new release'"
    echo "3. 添加仓库话题标签: golang, cli, developer-tools, china, mirror"
    echo ""
    echo "GitHub Actions 会自动运行 CI 测试"
    echo "如果创建 tag (如 v0.1.0)，会自动构建多平台二进制文件"
else
    echo ""
    echo "❌ 推送失败"
    echo ""
    echo "可能的原因:"
    echo "1. 网络问题"
    echo "2. 没有权限（需要配置 SSH key 或 token）"
    echo "3. 仓库不存在"
    echo ""
    echo "解决方法:"
    echo "1. 检查网络连接"
    echo "2. 配置 SSH key: https://docs.github.com/en/authentication"
    echo "3. 或使用 HTTPS + Personal Access Token"
    exit 1
fi
