@echo off
chcp 65001 >nul
REM EnvKit Windows Installation Script
REM Active code page set to UTF-8 to support Chinese display and paths

echo 正在启动 EnvKit 安装程序...
powershell -NoProfile -ExecutionPolicy Bypass -File "%~dp0install.ps1"

echo.
echo 按任意键退出...
pause >nul
