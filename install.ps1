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
    Write-Host "✅ EnvKit 已从本地复制并成功安装到: $destPath" -ForegroundColor Green
} else {
    Write-Host "📦 正在尝试从 GitHub Releases 下载预编译的二进制文件..." -ForegroundColor Cyan
    $url = "https://github.com/EvilJul/envkit/releases/latest/download/$binaryName"
    
    try {
        # 优先直接下载，限时 5 秒
        Write-Host "正在从 GitHub 直接下载..." -ForegroundColor Gray
        Invoke-WebRequest -Uri $url -OutFile $destPath -UseBasicParsing -TimeoutSec 5 | Out-Null
        Write-Host "✅ EnvKit 已从 GitHub Releases 下载并成功安装到: $destPath" -ForegroundColor Green
    } catch {
        # 尝试国内加速代理下载
        Write-Host "⚠️  从 GitHub 直连下载超时或失败，正在尝试通过国内镜像加速通道下载..." -ForegroundColor Yellow
        $proxyUrl = "https://ghp.ci/" + $url
        try {
            Invoke-WebRequest -Uri $proxyUrl -OutFile $destPath -UseBasicParsing | Out-Null
            Write-Host "✅ EnvKit 已通过加速通道下载并成功安装到: $destPath" -ForegroundColor Green
        } catch {
            if (Get-Command go -ErrorAction SilentlyContinue) {
                Write-Host "⚠️  加速通道下载也失败，检测到本地已安装 Go，正在尝试从源码编译..." -ForegroundColor Yellow
                try {
                    go build -o $destPath ./cmd/envkit/main.go
                    Write-Host "✅ EnvKit 已从源码成功编译并安装到: $destPath" -ForegroundColor Green
                } catch {
                    Write-Host "❌ 从源码编译 EnvKit 失败: $_" -ForegroundColor Red
                    Exit 1
                }
            } else {
                Write-Host "❌ 无法安装 EnvKit。" -ForegroundColor Red
                Write-Host "原因: 从 GitHub 下载失败且本地未安装 Go 编译器。" -ForegroundColor Red
                Write-Host "请确认：" -ForegroundColor Yellow
                Write-Host "  1. 您的网络能够正常访问 GitHub 或加速代理 https://ghp.ci。" -ForegroundColor Yellow
                Write-Host "  2. 您已经在 GitHub 上发布了项目的 Release 版本并上传了二进制包。" -ForegroundColor Yellow
                Exit 1
            }
        }
    }
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
