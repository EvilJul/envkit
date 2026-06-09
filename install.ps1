# EnvKit Windows 安装脚本
# 用法: 在 PowerShell 中运行: .\install.ps1

$ErrorActionPreference = "Stop"

Write-Host "🚀 EnvKit Windows 安装程序" -ForegroundColor Cyan
Write-Host ""

# 1. 检测系统架构
$arch = $env:PROCESSOR_ARCHITECTURE
if ($arch -eq "AMD64") {
    $binaryName = "envkit-windows-amd64.exe"
} else {
    Write-Host "⚠️  未检测到 64位 Intel/AMD 架构 (检测到: $arch)，将默认使用 AMD64 版本。" -ForegroundColor Yellow
    $binaryName = "envkit-windows-amd64.exe"
}

# 2. 设置并创建安装目录 (~/bin)
$installDir = Join-Path $HOME "bin"
if (-not (Test-Path $installDir)) {
    New-Item -ItemType Directory -Path $installDir | Out-Null
    Write-Host "创建安装目录: $installDir" -ForegroundColor Gray
}

# 3. 复制二进制文件
$srcPath = Join-Path $PSScriptRoot "dist\$binaryName"
$destPath = Join-Path $installDir "envkit.exe"

Write-Host "📦 正在安装 EnvKit..." -ForegroundColor Cyan

if (Test-Path $srcPath) {
    Copy-Item -Path $srcPath -Destination $destPath -Force
    Write-Host "✅ EnvKit 已成功安装到: $destPath" -ForegroundColor Green
} elseif (Get-Command go -ErrorAction SilentlyContinue) {
    Write-Host "⚠️  未在 dist/ 中找到预编译二进制文件，检测到本地已安装 Go，正在从源码编译..." -ForegroundColor Yellow
    try {
        go build -o $destPath ./cmd/envkit/main.go
        Write-Host "✅ EnvKit 已从源码成功编译并安装到: $destPath" -ForegroundColor Green
    } catch {
        Write-Host "❌ 从源码编译 EnvKit 失败: $_" -ForegroundColor Red
        Exit 1
    }
} else {
    Write-Host "❌ 无法安装 EnvKit。" -ForegroundColor Red
    Write-Host "原因: 无法在当前目录的 dist 文件夹中找到预编译二进制文件 $srcPath 且本地未安装 Go 编译器。" -ForegroundColor Red
    Write-Host "解决办法:" -ForegroundColor Yellow
    Write-Host "  1. 安装 Go 并运行 'go build -o dist/$binaryName ./cmd/envkit' 编译后再运行此脚本。" -ForegroundColor Yellow
    Write-Host "  2. 或者从 GitHub Releases 下载预编译版本: https://github.com/EvilJul/envkit/releases" -ForegroundColor Yellow
    Exit 1
}

# 4. 配置用户环境变量 PATH
$userPath = [Environment]::GetEnvironmentVariable("Path", "User")
$pathList = $userPath -split ";"

# 检查是否已包含
if ($pathList -notcontains $installDir) {
    Write-Host "正在将安装路径添加到用户环境变量 PATH..." -ForegroundColor Cyan
    $newUserPath = $userPath + ";" + $installDir
    # 过滤掉多余的分号
    $newUserPath = $newUserPath -replace ";+", ";"
    [Environment]::SetEnvironmentVariable("Path", $newUserPath, "User")
    Write-Host "✅ 环境变量已成功更新！" -ForegroundColor Green
    Write-Host "⚠️  提示: 您需要重启当前的 PowerShell 窗口或编辑器以加载新的 PATH 环境变量。" -ForegroundColor Yellow
} else {
    Write-Host "ℹ 环境变量 PATH 中已包含安装路径，无需更新。" -ForegroundColor Gray
}

Write-Host ""
Write-Host "🎉 安装完成！请重新打开终端并运行以下命令开始使用:" -ForegroundColor Green
Write-Host "  envkit init" -ForegroundColor Cyan
Write-Host ""
