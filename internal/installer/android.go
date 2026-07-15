// AndroidInstaller Android SDK 全局安装器
// 通过国内镜像源（腾讯云/阿里云/华为云）下载并安装 Android 命令行工具、platform-tools、build-tools、platforms
// 自动配置 ANDROID_HOME、ANDROID_SDK_ROOT 及 PATH 环境变量
package installer

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/fusheng/envkit/internal/ui"
)

// Android 默认安装参数
const (
	androidDefaultCmdlineToolsVersion = "11076708" // 最新稳定版 cmdline-tools
	androidDefaultBuildToolsVersion   = "34.0.0"
	androidDefaultPlatform            = "android-34"
)

// AndroidInstaller Android SDK 安装器
type AndroidInstaller struct {
	cmdlineToolsVersion string
	buildToolsVersion   string
	platforms           []string
}

// NewAndroidInstaller 创建带自定义参数的 Android 安装器
func NewAndroidInstaller() *AndroidInstaller {
	return &AndroidInstaller{
		cmdlineToolsVersion: androidDefaultCmdlineToolsVersion,
		buildToolsVersion:   androidDefaultBuildToolsVersion,
		platforms:           []string{androidDefaultPlatform},
	}
}

// Install 安装 Android SDK
func (a *AndroidInstaller) Install() error {
	spinner := ui.NewSpinner("正在安装 Android SDK")
	spinner.Start()
	defer spinner.Stop()

	// 计算 SDK 根目录
	sdkRoot, err := a.getSdkRoot()
	if err != nil {
		return err
	}
	ui.Info("Android SDK 安装目录: %s", sdkRoot)

	// 确保目录存在
	if err := os.MkdirAll(sdkRoot, 0755); err != nil {
		return fmt.Errorf("创建 SDK 根目录失败: %w", err)
	}

	// 各平台安装
	var installErr error
	switch runtime.GOOS {
	case "darwin":
		installErr = a.installOnDarwin(sdkRoot)
	case "linux":
		installErr = a.installOnLinux(sdkRoot)
	case "windows":
		installErr = a.installOnWindows(sdkRoot)
	default:
		return fmt.Errorf("不支持的操作系统: %s", runtime.GOOS)
	}
	if installErr != nil {
		return installErr
	}

	// 先装组件，再写 PATH（否则 platform-tools 还不存在会被跳过）
	if err := a.installSdkComponents(sdkRoot); err != nil {
		return fmt.Errorf("安装 SDK 组件失败: %w", err)
	}

	if err := a.setupEnv(sdkRoot); err != nil {
		ui.Warning("配置环境变量失败: %v", err)
	}
	// 组件装完后再次确保 PATH
	_ = a.setupEnv(sdkRoot)

	if !a.IsInstalled() {
		return fmt.Errorf("Android SDK 安装未完成：未检测到 adb（platform-tools）")
	}

	paths := []string{sdkRoot}
	shellLines := []string{
		"# envkit:android",
		"ANDROID_HOME",
		"ANDROID_SDK_ROOT",
	}
	_ = RecordInstallation("android", "tool", a.cmdlineToolsVersion, paths, shellLines)

	ui.Success("Android SDK 安装完成！")
	if runtime.GOOS == "windows" {
		ui.Info("提示: 请重新打开终端使 PATH 生效")
	} else {
		ui.Info("提示: 请重新打开终端或 source ~/.bashrc 使环境变量生效")
	}
	return nil
}

// IsInstalled 检查 Android SDK 是否已安装（ANDROID_HOME 存在且 platform-tools/adb 可用）
func (a *AndroidInstaller) IsInstalled() bool {
	sdkRoot, err := a.getSdkRoot()
	if err != nil {
		return false
	}
	// 检查 ANDROID_HOME 目录是否存在
	if _, err := os.Stat(sdkRoot); err != nil {
		// 兼容环境变量已设置但默认目录不存在的场景
		if env := os.Getenv("ANDROID_HOME"); env != "" {
			if _, err := os.Stat(env); err == nil {
				sdkRoot = env
			} else {
				return false
			}
		} else {
			return false
		}
	}
	// platform-tools/adb 必须存在
	adbPath := filepath.Join(sdkRoot, "platform-tools", "adb")
	if runtime.GOOS == "windows" {
		adbPath += ".exe"
	}
	if _, err := os.Stat(adbPath); err != nil {
		return false
	}
	return true
}

// GetVersion 返回 adb --version 第一行
func (a *AndroidInstaller) GetVersion() string {
	sdkRoot, err := a.getSdkRoot()
	if err != nil {
		return ""
	}
	adbPath := filepath.Join(sdkRoot, "platform-tools", "adb")
	if runtime.GOOS == "windows" {
		adbPath += ".exe"
	}
	cmd := exec.Command(adbPath, "--version")
	output, err := cmd.Output()
	if err != nil {
		// 兜底：直接尝试 PATH 中的 adb
		cmd = exec.Command("adb", "--version")
		output, err = cmd.Output()
		if err != nil {
			return ""
		}
	}
	lines := strings.Split(string(output), "\n")
	if len(lines) > 0 {
		return strings.TrimSpace(lines[0])
	}
	return strings.TrimSpace(string(output))
}

// getSdkRoot 获取当前平台的 SDK 安装根目录
func (a *AndroidInstaller) getSdkRoot() (string, error) {
	// 优先使用已设置的 ANDROID_HOME
	if env := os.Getenv("ANDROID_HOME"); env != "" {
		return env, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	switch runtime.GOOS {
	case "darwin":
		return filepath.Join(home, "Library", "Android", "sdk"), nil
	case "linux":
		return filepath.Join(home, "Android", "Sdk"), nil
	case "windows":
		localAppData := os.Getenv("LOCALAPPDATA")
		if localAppData == "" {
			localAppData = filepath.Join(home, "AppData", "Local")
		}
		return filepath.Join(localAppData, "Android", "Sdk"), nil
	default:
		return "", fmt.Errorf("不支持的操作系统: %s", runtime.GOOS)
	}
}

// installOnDarwin macOS 安装
func (a *AndroidInstaller) installOnDarwin(sdkRoot string) error {
	// 检查依赖
	if err := CheckAndInstallDependencies([]SystemDependency{
		CommonDependencies[0], // curl
		CommonDependencies[2], // unzip
	}); err != nil {
		return err
	}

	// 下载并解压 cmdline-tools
	if err := a.downloadAndExtractCmdlineTools(sdkRoot, "mac"); err != nil {
		return err
	}
	return nil
}

// installOnLinux Linux 安装
func (a *AndroidInstaller) installOnLinux(sdkRoot string) error {
	// 检查依赖
	if err := CheckAndInstallDependencies([]SystemDependency{
		CommonDependencies[0], // curl
		CommonDependencies[2], // unzip
	}); err != nil {
		return err
	}

	// 下载并解压 cmdline-tools
	if err := a.downloadAndExtractCmdlineTools(sdkRoot, "linux"); err != nil {
		return err
	}
	return nil
}

// installOnWindows Windows 安装
func (a *AndroidInstaller) installOnWindows(sdkRoot string) error {
	// Windows 平台，优先尝试 winget 安装（若失败则回退到国内镜像）
	if commandExists("winget") {
		ui.Info("检测到 winget，尝试通过 winget 安装 Android Studio（包含 SDK）...")
		// Android SDK 不在 winget 中直接提供，因此直接走下载流程
	}

	// 下载并解压 cmdline-tools
	return a.downloadAndExtractCmdlineTools(sdkRoot, "win")
}

// downloadAndExtractCmdlineTools 下载并解压 cmdline-tools 到 $SDK_ROOT/cmdline-tools/latest
func (a *AndroidInstaller) downloadAndExtractCmdlineTools(sdkRoot, platform string) error {
	cmdlineDir := filepath.Join(sdkRoot, "cmdline-tools")
	sdkmanagerPath := filepath.Join(cmdlineDir, "latest", "bin", "sdkmanager")
	if runtime.GOOS == "windows" {
		sdkmanagerPath += ".bat"
	}

	// 已安装则跳过
	if _, err := os.Stat(sdkmanagerPath); err == nil {
		ui.Info("检测到 cmdline-tools 已安装，跳过下载: %s", sdkmanagerPath)
		return nil
	}

	if err := os.MkdirAll(cmdlineDir, 0755); err != nil {
		return fmt.Errorf("创建 cmdline-tools 目录失败: %w", err)
	}

	// 构造候选下载 URL 列表（按优先级：腾讯云 > 阿里云 > 华为云 > Google 官方）
	urls := a.buildCmdlineToolsURLs(platform)
	if len(urls) == 0 {
		return fmt.Errorf("未知的平台: %s", platform)
	}

	// 下载到临时目录
	zipPath := filepath.Join(os.TempDir(), fmt.Sprintf("commandlinetools-%s-%s.zip", platform, a.cmdlineToolsVersion))

	// 依次尝试镜像源
	var lastErr error
	downloaded := false
	for _, url := range urls {
		ui.Info("正在从镜像下载 Android cmdline-tools: %s", url)
		if err := a.downloadFromMirror(url, zipPath); err != nil {
			ui.Warning("从 %s 下载失败: %v", url, err)
			lastErr = err
			continue
		}
		downloaded = true
		break
	}
	if !downloaded {
		return fmt.Errorf("所有镜像源下载 cmdline-tools 失败: %w", lastErr)
	}
	defer os.Remove(zipPath)

	// 解压：先解压到临时目录，再移动到 cmdline-tools/latest
	ui.Info("正在解压 cmdline-tools...")
	tmpExtract := filepath.Join(os.TempDir(), fmt.Sprintf("android-cmdline-%s", platform))
	_ = os.RemoveAll(tmpExtract)
	if err := os.MkdirAll(tmpExtract, 0755); err != nil {
		return fmt.Errorf("创建临时解压目录失败: %w", err)
	}
	defer os.RemoveAll(tmpExtract)

	if err := a.extractZip(zipPath, tmpExtract); err != nil {
		return fmt.Errorf("解压 cmdline-tools 失败: %w", err)
	}

	// cmdline-tools zip 解压后的根目录通常名为 "cmdline-tools"
	srcDir := filepath.Join(tmpExtract, "cmdline-tools")
	if _, err := os.Stat(srcDir); err != nil {
		// 兼容部分镜像的目录命名
		entries, _ := os.ReadDir(tmpExtract)
		if len(entries) == 1 && entries[0].IsDir() {
			srcDir = filepath.Join(tmpExtract, entries[0].Name())
		}
	}

	latestDir := filepath.Join(cmdlineDir, "latest")
	_ = os.RemoveAll(latestDir)
	if err := os.MkdirAll(filepath.Dir(latestDir), 0755); err != nil {
		return fmt.Errorf("创建 latest 父目录失败: %w", err)
	}
	if err := os.Rename(srcDir, latestDir); err != nil {
		return fmt.Errorf("移动 cmdline-tools 到 %s 失败: %w", latestDir, err)
	}

	ui.Success("cmdline-tools 已安装至: %s", latestDir)
	return nil
}

// buildCmdlineToolsURLs 根据平台构造 cmdline-tools 下载 URL 列表（多镜像回退）
func (a *AndroidInstaller) buildCmdlineToolsURLs(platform string) []string {
	ver := a.cmdlineToolsVersion
	urls := []string{}

	// 1) 腾讯云
	switch platform {
	case "mac":
		urls = append(urls, fmt.Sprintf("https://mirrors.cloud.tencent.com/AndroidSDK/commandlinetools-mac-%s_latest.zip", ver))
	case "linux":
		urls = append(urls, fmt.Sprintf("https://mirrors.cloud.tencent.com/AndroidSDK/commandlinetools-linux-%s_latest.zip", ver))
	case "win":
		urls = append(urls, fmt.Sprintf("https://mirrors.cloud.tencent.com/AndroidSDK/commandlinetools-win-%s_latest.zip", ver))
	}

	// 2) 阿里云（保留目录形式以备扩展）
	urls = append(urls, "https://mirrors.aliyun.com/android/repository/")

	// 3) 华为云
	urls = append(urls, "https://repo.huaweicloud.com/android/repository/")

	// 4) Google 官方（兜底）
	switch platform {
	case "mac":
		urls = append(urls, fmt.Sprintf("https://dl.google.com/android/repository/commandlinetools-mac-%s_latest.zip", ver))
	case "linux":
		urls = append(urls, fmt.Sprintf("https://dl.google.com/android/repository/commandlinetools-linux-%s_latest.zip", ver))
	case "win":
		urls = append(urls, fmt.Sprintf("https://dl.google.com/android/repository/commandlinetools-win-%s_latest.zip", ver))
	}

	return urls
}

// downloadFromMirror 使用 curl 下载文件到目标路径，带进度提示
func (a *AndroidInstaller) downloadFromMirror(url, dest string) error {
	// 确保目标目录存在
	if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		return err
	}

	// 阿里云/华为云的目录型 URL 单独提示并跳过下载（避免误下载目录索引页）
	if strings.HasSuffix(url, "/") {
		return fmt.Errorf("该 URL 是目录入口，无法直接下载 zip: %s", url)
	}

	// 使用 curl -fSL 下载，失败时显示 HTTP 错误
	if err := runCommand("curl", "-fSL", "--retry", "3", "-o", dest, url); err != nil {
		return err
	}

	// 校验文件大小，避免下载到错误页面
	info, err := os.Stat(dest)
	if err != nil {
		return fmt.Errorf("下载后无法读取文件: %w", err)
	}
	if info.Size() < 1024 {
		// cmdline-tools 通常 > 100MB，小于 1KB 几乎可以肯定是错误页
		_ = os.Remove(dest)
		return fmt.Errorf("下载文件过小 (%d 字节)，可能为错误页面", info.Size())
	}

	return nil
}

// extractZip 解压 zip 文件到目标目录
func (a *AndroidInstaller) extractZip(zipPath, destDir string) error {
	if runtime.GOOS == "windows" {
		// 优先尝试 PowerShell 的 Expand-Archive
		if err := runCommand("powershell", "-NoProfile", "-Command",
			fmt.Sprintf("Expand-Archive -Path '%s' -DestinationPath '%s' -Force", zipPath, destDir)); err == nil {
			return nil
		}
		// 备选 tar（Windows 10 1803+ 自带）
		return runCommand("tar", "-xf", zipPath, "-C", destDir)
	}
	return runCommand("unzip", "-q", zipPath, "-d", destDir)
}

// setupEnv 配置 ANDROID_HOME / ANDROID_SDK_ROOT / PATH 环境变量
func (a *AndroidInstaller) setupEnv(sdkRoot string) error {
	// 当前进程立即生效
	os.Setenv("ANDROID_HOME", sdkRoot)
	os.Setenv("ANDROID_SDK_ROOT", sdkRoot)

	// 需要加入 PATH 的子目录
	dirsToAdd := []string{
		filepath.Join(sdkRoot, "platform-tools"),
		filepath.Join(sdkRoot, "cmdline-tools", "latest", "bin"),
		filepath.Join(sdkRoot, "emulator"),
		filepath.Join(sdkRoot, "build-tools", a.buildToolsVersion),
	}

	for _, dir := range dirsToAdd {
		if _, err := os.Stat(dir); err == nil {
			AddDirToPath(dir)
			_ = PersistPathEnv(dir)
		}
	}

	// 写入 ANDROID_HOME / ANDROID_SDK_ROOT 到 shell 配置文件
	if runtime.GOOS != "windows" {
		return a.persistAndroidEnvVars(sdkRoot)
	}

	// Windows: 始终持久化到 User 环境，不因进程内 os.Setenv 已设置而跳过
	if err := PersistUserEnvVar("ANDROID_HOME", sdkRoot); err != nil {
		ui.Warning("写入 ANDROID_HOME 失败: %v", err)
	}
	if err := PersistUserEnvVar("ANDROID_SDK_ROOT", sdkRoot); err != nil {
		ui.Warning("写入 ANDROID_SDK_ROOT 失败: %v", err)
	}
	ui.Info("ANDROID_HOME / ANDROID_SDK_ROOT 已写入用户环境变量，请重启终端生效")
	return nil
}

// persistAndroidEnvVars 将 ANDROID_HOME / ANDROID_SDK_ROOT 写入 Unix shell 配置文件
func (a *AndroidInstaller) persistAndroidEnvVars(sdkRoot string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	// 尝试用 $HOME 代替绝对路径，使配置更简洁
	pathValue := sdkRoot
	if strings.HasPrefix(sdkRoot, home) {
		pathValue = "$HOME" + strings.TrimPrefix(sdkRoot, home)
	}

	envBlock := fmt.Sprintf(
		"\n# added by envkit (android)\nexport ANDROID_HOME=\"%s\"\nexport ANDROID_SDK_ROOT=\"%s\"\n",
		pathValue, pathValue,
	)

	files := getShellConfigFiles()
	for _, file := range files {
		if _, err := os.Stat(file); err != nil {
			continue
		}
		content, err := os.ReadFile(file)
		if err != nil {
			continue
		}
		// 已存在 ANDROID_HOME 配置则跳过
		if strings.Contains(string(content), "ANDROID_HOME") {
			continue
		}
		f, err := os.OpenFile(file, os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			continue
		}
		_, _ = f.WriteString(envBlock)
		_ = f.Close()
	}
	return nil
}

// installSdkComponents 使用 sdkmanager 接受 license 并安装 platform-tools / build-tools / platforms
func (a *AndroidInstaller) installSdkComponents(sdkRoot string) error {
	sdkmanagerBin := filepath.Join(sdkRoot, "cmdline-tools", "latest", "bin", "sdkmanager")
	if runtime.GOOS == "windows" {
		sdkmanagerBin += ".bat"
	}
	if _, err := os.Stat(sdkmanagerBin); err != nil {
		return fmt.Errorf("未找到 sdkmanager: %s，请确认 cmdline-tools 安装成功", sdkmanagerBin)
	}

	ui.Info("正在接受 Android SDK licenses ...")
	// 使用 yes 自动化 license 确认（pipe 给 sdkmanager --licenses）
	if runtime.GOOS == "windows" {
		// 多份 y：sdkmanager 会连续询问多份 license；.bat 经 cmd /C 调用
		// (echo y& echo y& ...) | "sdkmanager.bat" --licenses --sdk_root=...
		yesParts := make([]string, 0, 40)
		for i := 0; i < 40; i++ {
			yesParts = append(yesParts, "echo y")
		}
		yesPipe := strings.Join(yesParts, "& ")
		cmdLine := fmt.Sprintf("(%s) | \"%s\" --licenses --sdk_root=%s", yesPipe, sdkmanagerBin, sdkRoot)
		if err := runCommand("cmd", "/C", cmdLine); err != nil {
			ui.Warning("接受 licenses 失败（可稍后手动执行）: %v", err)
		}
	} else {
		if err := runCommand("sh", "-c", fmt.Sprintf("yes | \"%s\" --licenses --sdk_root=%s", sdkmanagerBin, sdkRoot)); err != nil {
			ui.Warning("接受 licenses 失败（可稍后手动执行）: %v", err)
		}
	}

	// 组装待安装组件列表
	components := []string{
		"platform-tools",
		fmt.Sprintf("build-tools;%s", a.buildToolsVersion),
	}
	for _, p := range a.platforms {
		components = append(components, fmt.Sprintf("platforms;%s", p))
	}

	// 优化：先检查已安装的组件，跳过已存在的
	installedSet := a.listInstalledComponents(sdkRoot, sdkmanagerBin)
	toInstall := []string{}
	for _, c := range components {
		if installedSet[c] {
			ui.Info("组件已安装，跳过: %s", c)
			continue
		}
		toInstall = append(toInstall, c)
	}

	if len(toInstall) == 0 {
		ui.Info("所有目标 SDK 组件均已安装")
		return nil
	}

	ui.Info("正在安装 SDK 组件: %s", strings.Join(toInstall, ", "))
	// 强制使用国内镜像（dl.google.com 在国内可能不稳定，sdkmanager 默认即走 dl.google.com）
	// 如未来需要切换 sdkmanager 仓库，可在此设置环境变量：
	// os.Setenv("ANDROID_SDK_MANAGER_REPO_OVERRIDE", "https://mirrors.cloud.tencent.com/AndroidSDK/")
	args := append([]string{"--sdk_root=" + sdkRoot}, toInstall...)
	var installErr error
	if runtime.GOOS == "windows" {
		// Windows 上 .bat 必须通过 cmd /C 调用
		installErr = runWindowsScript(sdkmanagerBin, args...)
	} else {
		installErr = runCommand(sdkmanagerBin, args...)
	}
	if installErr != nil {
		return fmt.Errorf("安装 SDK 组件失败: %w", installErr)
	}

	ui.Success("Android SDK 组件安装完成")
	return nil
}

// listInstalledComponents 列出已安装的 SDK 组件，返回 set 以便 O(1) 查找
func (a *AndroidInstaller) listInstalledComponents(sdkRoot, sdkmanagerBin string) map[string]bool {
	set := make(map[string]bool)
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/C", sdkmanagerBin, "--sdk_root="+sdkRoot, "--list_installed")
	} else {
		cmd = exec.Command(sdkmanagerBin, "--sdk_root="+sdkRoot, "--list_installed")
	}
	output, err := cmd.Output()
	if err != nil {
		// 命令失败时直接返回空集（视为全部未安装）
		return set
	}
	// 输出格式示例:
	//   Installed packages:
	//     Path                | Version | Description
	//     platform-tools      | 35.0.0  | ...
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "Installed") || strings.HasPrefix(line, "Path") {
			continue
		}
		parts := strings.Split(line, "|")
		if len(parts) == 0 {
			continue
		}
		pkg := strings.TrimSpace(parts[0])
		if pkg != "" {
			set[pkg] = true
		}
	}
	return set
}
