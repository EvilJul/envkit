@echo off
:: EnvKit Windows 一键安装脚本
:: 作用: 双击运行或在 cmd 中执行，调用 PowerShell 进行安全且自动的安装与环境变量配置。

echo 正在启动 EnvKit 安装程序...
powershell -NoProfile -ExecutionPolicy Bypass -File "%~dp0install.ps1"

echo.
echo 按任意键退出...
pause >nul
