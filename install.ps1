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
} else {
    Write-Host "❌ 无法在当前目录的 dist 文件夹中找到: $srcPath" -ForegroundColor Red
    Write-Host "请确保你在克隆的项目根目录下运行此脚本。" -ForegroundColor Yellow
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
