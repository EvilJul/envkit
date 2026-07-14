# EnvKit Windows 安装脚本
# 用法: .\install.ps1
# 可选环境变量: ENVKIT_MIRROR, ENVKIT_REPO, ENVKIT_VERSION

$ErrorActionPreference = "Stop"

$Repo = if ($env:ENVKIT_REPO) { $env:ENVKIT_REPO } else { "EvilJul/envkit" }
$Version = if ($env:ENVKIT_VERSION) { $env:ENVKIT_VERSION } else { "latest" }
$CustomMirror = $env:ENVKIT_MIRROR
$TimeoutSec = 300

Write-Host "🚀 EnvKit Windows 安装程序" -ForegroundColor Cyan
Write-Host ""

$arch = $env:PROCESSOR_ARCHITECTURE
if ($arch -eq "ARM64") {
    $binaryName = "envkit-windows-arm64.exe"
} else {
    $binaryName = "envkit-windows-amd64.exe"
}

$installDir = Join-Path $HOME "bin"
if (-not (Test-Path $installDir)) {
    New-Item -ItemType Directory -Path $installDir | Out-Null
}
$destPath = Join-Path $installDir "envkit.exe"
$srcPath = Join-Path $PSScriptRoot "dist\$binaryName"

function Get-ReleaseUrl {
    if ($Version -eq "latest") {
        return "https://github.com/$Repo/releases/latest/download/$binaryName"
    }
    return "https://github.com/$Repo/releases/download/$Version/$binaryName"
}

function Test-ValidBinary([string]$Path) {
    if (-not (Test-Path $Path)) { return $false }
    $len = (Get-Item $Path).Length
    if ($len -lt 1000000) { return $false }
    $bytes = Get-Content -Path $Path -Encoding Byte -TotalCount 2
    # MZ
    return ($bytes[0] -eq 0x4D -and $bytes[1] -eq 0x5A)
}

function Try-Download([string]$Url, [string]$OutFile) {
    Write-Host "  → $Url" -ForegroundColor Gray
    try {
        Invoke-WebRequest -Uri $Url -OutFile $OutFile -UseBasicParsing -TimeoutSec $TimeoutSec
        if (Test-ValidBinary $OutFile) { return $true }
        Write-Host "    下载内容无效，跳过" -ForegroundColor Yellow
        Remove-Item -Force $OutFile -ErrorAction SilentlyContinue
    } catch {
        Write-Host "    失败: $($_.Exception.Message)" -ForegroundColor Yellow
        Remove-Item -Force $OutFile -ErrorAction SilentlyContinue
    }
    return $false
}

Write-Host "📦 正在安装 EnvKit..." -ForegroundColor Cyan

if (Test-Path $srcPath) {
    Copy-Item -Path $srcPath -Destination $destPath -Force
    Write-Host "✅ 已从本地 dist 安装: $destPath" -ForegroundColor Green
} else {
    $base = Get-ReleaseUrl
    $urls = New-Object System.Collections.Generic.List[string]
    [void]$urls.Add($base)
    if ($CustomMirror) {
        $prefix = $CustomMirror.TrimEnd("/")
        [void]$urls.Add("$prefix/$base")
    }
    foreach ($p in @(
            "https://ghfast.top/",
            "https://gh-proxy.com/",
            "https://ghproxy.net/",
            "https://mirror.ghproxy.com/",
            "https://gitdl.cn/"
        )) {
        [void]$urls.Add("$p$base")
    }

    $ok = $false
    foreach ($u in $urls) {
        if (Try-Download $u $destPath) {
            Write-Host "✅ 下载成功" -ForegroundColor Green
            $ok = $true
            break
        }
    }

    if (-not $ok) {
        if (Get-Command go -ErrorAction SilentlyContinue) {
            Write-Host "📦 使用本地 Go 从源码编译..." -ForegroundColor Cyan
            Push-Location $PSScriptRoot
            try {
                $env:CGO_ENABLED = "0"
                go build -trimpath -ldflags="-s -w" -o $destPath ./cmd/envkit
                Write-Host "✅ 已从源码编译: $destPath" -ForegroundColor Green
                $ok = $true
            } finally {
                Pop-Location
            }
        }
    }

    if (-not $ok) {
        Write-Host "❌ 安装失败：网络下载失败且无可用 Go。" -ForegroundColor Red
        Write-Host "可设置 ENVKIT_MIRROR=https://ghfast.top/ 后重试" -ForegroundColor Yellow
        Write-Host "或手动下载: $base" -ForegroundColor Yellow
        exit 1
    }
}

$userPath = [Environment]::GetEnvironmentVariable("Path", "User")
if (-not (($userPath -split ";") -contains $installDir)) {
    $newUserPath = ($userPath + ";" + $installDir) -replace ";+", ";"
    [Environment]::SetEnvironmentVariable("Path", $newUserPath, "User")
    Write-Host "✅ 已写入用户 PATH（请重开终端）" -ForegroundColor Green
}

Write-Host ""
Write-Host "🎉 安装完成，请重开终端后运行: envkit version" -ForegroundColor Green
Write-Host ""
