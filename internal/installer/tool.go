package installer

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/fusheng/envkit/internal/ui"
)

// ToolInstaller 工具安装器接口
type ToolInstaller interface {
	Install() error
	IsInstalled() bool
	GetVersion() string
}

// GitInstaller Git 安装器
type GitInstaller struct{}

func (g *GitInstaller) Install() error {
	spinner := ui.NewSpinner("正在安装 Git")
	spinner.Start()
	defer spinner.Stop()

	var err error
	switch runtime.GOOS {
	case "darwin":
		err = installWithBrew("git")
	case "linux":
		err = installWithApt("git")
	case "windows":
		err = installWithWinget("Git.Git")
	default:
		return fmt.Errorf("不支持的操作系统: %s", runtime.GOOS)
	}

	if err != nil {
		return err
	}
	if !g.IsInstalled() {
		return fmt.Errorf("Git 安装完成但未检测到 git 命令")
	}
	if err := recordInstall("git", "tool", "latest", nil, []string{"# envkit:git"}); err != nil {
		return err
	}
	return nil
}

func (g *GitInstaller) IsInstalled() bool {
	return commandExists("git")
}

func (g *GitInstaller) GetVersion() string {
	cmd := exec.Command("git", "--version")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return string(output)
}

// DockerInstaller Docker 安装器
type DockerInstaller struct{}

func (d *DockerInstaller) Install() error {
	spinner := ui.NewSpinner("正在安装 Docker")
	spinner.Start()
	defer spinner.Stop()

	switch runtime.GOOS {
	case "darwin":
		ui.Warning("macOS 上请从 https://www.docker.com/products/docker-desktop 下载 Docker Desktop")
		return fmt.Errorf("请手动安装 Docker Desktop")
	case "linux":
		return d.installOnLinux()
	case "windows":
		ui.Warning("Windows 上请从 https://www.docker.com/products/docker-desktop 下载 Docker Desktop")
		return fmt.Errorf("请手动安装 Docker Desktop")
	default:
		return fmt.Errorf("不支持的操作系统: %s", runtime.GOOS)
	}
}

func (d *DockerInstaller) installOnLinux() error {
	if !commandExists("apt-get") {
		return fmt.Errorf("当前 Docker 自动安装仅支持 Debian/Ubuntu 系统。其他 Linux 发行版请参考官方文档手动安装: https://docs.docker.com/engine/install/")
	}

	// 检查依赖
	if err := CheckAndInstallDependencies([]SystemDependency{
		CommonDependencies[0], // curl
		CommonDependencies[5], // gpg
		{
			Name:     "lsb-release",
			Commands: []string{"lsb_release"},
			PackageName: map[string]string{
				"apt": "lsb-release",
			},
			Required: true,
		},
	}); err != nil {
		return err
	}

	ui.Info("正在安装 Docker Engine...")

	// 更新包索引
	if err := runCommand("sudo", "apt-get", "update"); err != nil {
		return fmt.Errorf("更新包索引失败: %w", err)
	}

	// 安装依赖
	if err := runCommand("sudo", "apt-get", "install", "-y",
		"ca-certificates", "curl", "gnupg", "lsb-release"); err != nil {
		return fmt.Errorf("安装依赖失败: %w", err)
	}

	// 添加 Docker GPG 密钥
	if err := runCommand("bash", "-c",
		"curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg"); err != nil {
		return fmt.Errorf("添加 GPG 密钥失败: %w", err)
	}

	// 添加 Docker 仓库
	if err := runCommand("bash", "-c",
		`echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null`); err != nil {
		return fmt.Errorf("添加仓库失败: %w", err)
	}

	// 更新包索引
	if err := runCommand("sudo", "apt-get", "update"); err != nil {
		return fmt.Errorf("更新包索引失败: %w", err)
	}

	// 安装 Docker
	if err := runCommand("sudo", "apt-get", "install", "-y",
		"docker-ce", "docker-ce-cli", "containerd.io", "docker-compose-plugin"); err != nil {
		return fmt.Errorf("安装 Docker 失败: %w", err)
	}

	// 启动 Docker 服务
	if err := runCommand("sudo", "systemctl", "start", "docker"); err != nil {
		ui.Warning("启动 Docker 服务失败: %v", err)
	}

	// 启用开机自启
	if err := runCommand("sudo", "systemctl", "enable", "docker"); err != nil {
		ui.Warning("设置开机自启失败: %v", err)
	}

	if !d.IsInstalled() {
		return fmt.Errorf("Docker 安装流程结束但未检测到 docker 命令")
	}
	ui.Success("Docker 安装成功！")
	if err := recordInstall("docker", "tool", "latest", nil, []string{"# envkit:docker"}); err != nil {
		return err
	}
	ui.Info("提示: 运行 'sudo usermod -aG docker $USER' 将当前用户添加到 docker 组")
	return nil
}

func (d *DockerInstaller) IsInstalled() bool {
	return commandExists("docker")
}

func (d *DockerInstaller) GetVersion() string {
	cmd := exec.Command("docker", "--version")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return string(output)
}

// VSCodeInstaller VSCode 安装器
type VSCodeInstaller struct{}

func (v *VSCodeInstaller) Install() error {
	spinner := ui.NewSpinner("正在安装 VSCode")
	spinner.Start()
	defer spinner.Stop()

	var err error
	switch runtime.GOOS {
	case "darwin":
		err = v.installOnDarwin()
	case "linux":
		err = v.installOnLinux()
	case "windows":
		err = installWithWinget("Microsoft.VisualStudioCode")
	default:
		return fmt.Errorf("不支持的操作系统: %s", runtime.GOOS)
	}

	if err != nil {
		return err
	}
	if !v.IsInstalled() {
		return fmt.Errorf("VSCode 安装完成但未检测到 code 命令")
	}
	switch runtime.GOOS {
	case "darwin":
		if err := recordInstall("vscode", "tool", "stable", []string{"/Applications/Visual Studio Code.app", "~/.local/bin/code", "/usr/local/bin/code"}, []string{"# envkit:vscode"}); err != nil {
			return err
		}
	default:
		if err := recordInstall("vscode", "tool", "stable", nil, []string{"# envkit:vscode"}); err != nil {
			return err
		}
	}
	return nil
}

func (v *VSCodeInstaller) installOnDarwin() error {
	// 尝试通过 Homebrew Cask 安装
	err := installWithBrew("--cask", "visual-studio-code")
	if err == nil {
		return nil
	}

	// 备用方案：直接从 VSCode 官网下载 zip 并安装
	ui.Warning("通过 Homebrew Cask 安装 VSCode 失败 (%v)，正在尝试直接从官方下载...", err)

	// 检查依赖
	if err := CheckAndInstallDependencies([]SystemDependency{
		CommonDependencies[0], // curl
		CommonDependencies[2], // unzip
	}); err != nil {
		return err
	}

	downloadURL := "https://update.code.visualstudio.com/latest/darwin-universal/stable"
	zipPath := filepath.Join(os.TempDir(), "vscode.zip")

	if err := runCommand("curl", "-L", "-o", zipPath, downloadURL); err != nil {
		return fmt.Errorf("下载 VSCode 失败: %w", err)
	}

	// 解压
	extractDir := filepath.Join(os.TempDir(), "vscode-extracted")
	if err := runCommand("unzip", "-q", zipPath, "-d", extractDir); err != nil {
		os.Remove(zipPath)
		return fmt.Errorf("解压 VSCode 失败: %w", err)
	}

	// 移动至 /Applications
	destPath := "/Applications/Visual Studio Code.app"
	_ = exec.Command("rm", "-rf", destPath).Run()

	if err := runCommand("mv", filepath.Join(extractDir, "Visual Studio Code.app"), destPath); err != nil {
		// 备用 sudo 移动
		_ = runCommand("sudo", "mv", filepath.Join(extractDir, "Visual Studio Code.app"), destPath)
	}

	// 移除 macOS quarantine 属性，防止首次运行时 Gatekeeper 弹窗阻塞终端
	_ = exec.Command("xattr", "-cr", destPath).Run()

	// 清理
	os.Remove(zipPath)
	_ = exec.Command("rm", "-rf", extractDir).Run()

	// 自动创建 code 软链接至 ~/.local/bin，以实现终端 code 命令可用
	home, errHome := os.UserHomeDir()
	if errHome == nil {
		localBin := filepath.Join(home, ".local", "bin")
		_ = os.MkdirAll(localBin, 0755)
		linkPath := filepath.Join(localBin, "code")
		_ = exec.Command("rm", "-f", linkPath).Run()
		_ = exec.Command("ln", "-s", "/Applications/Visual Studio Code.app/Contents/Resources/app/bin/code", linkPath).Run()

		// 同时尝试软链到 /usr/local/bin (如果可用)
		_ = exec.Command("ln", "-s", "/Applications/Visual Studio Code.app/Contents/Resources/app/bin/code", "/usr/local/bin/code").Run()
	}

	return nil
}

func (v *VSCodeInstaller) installOnLinux() error {
	if !commandExists("dpkg") {
		return fmt.Errorf("当前 VSCode 自动安装仅支持 Debian/Ubuntu 系统。其他 Linux 发行版请前往官网手动下载并安装: https://code.visualstudio.com/")
	}

	// 检查依赖
	if err := CheckAndInstallDependencies([]SystemDependency{
		CommonDependencies[0], // curl
	}); err != nil {
		return err
	}

	// 下载并安装 VSCode .deb 包
	ui.Info("正在下载 VSCode...")

	debPath := filepath.Join(os.TempDir(), "vscode.deb")
	osKey := "linux-deb-x64"
	switch runtime.GOARCH {
	case "arm64", "aarch64":
		osKey = "linux-deb-arm64"
	case "arm":
		osKey = "linux-deb-armhf"
	}
	if err := runCommand("curl", "-fsSL", "-o", debPath,
		"https://code.visualstudio.com/sha/download?build=stable&os="+osKey); err != nil {
		return fmt.Errorf("下载失败 (%s): %w", osKey, err)
	}

	ui.Info("正在安装...")
	if err := runCommand("sudo", "dpkg", "-i", debPath); err != nil {
		// 修复依赖
		_ = runCommand("sudo", "apt-get", "install", "-f", "-y")
	}

	// 清理临时文件
	os.Remove(debPath)

	return nil
}

func (v *VSCodeInstaller) IsInstalled() bool {
	return commandExists("code")
}

func (v *VSCodeInstaller) GetVersion() string {
	cmd := exec.Command("code", "--version")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return string(output)
}

// 辅助函数

func installWithBrew(args ...string) error {
	if !commandExists("brew") {
		return fmt.Errorf("请先安装 Homebrew: https://brew.sh")
	}

	fullArgs := append([]string{"install"}, args...)
	return runCommand("brew", fullArgs...)
}

func installWithApt(packageName string) error {
	if !commandExists("apt-get") {
		// 自动适配其它常见 Linux 包管理器
		if commandExists("yum") {
			return runCommand("sudo", "yum", "install", "-y", packageName)
		} else if commandExists("dnf") {
			return runCommand("sudo", "dnf", "install", "-y", packageName)
		} else if commandExists("pacman") {
			return runCommand("sudo", "pacman", "-S", "--noconfirm", packageName)
		}
		return fmt.Errorf("不支持的 Linux 发行版或缺少包管理器。请手动安装: %s", packageName)
	}

	// 更新包索引
	if err := runCommand("sudo", "apt-get", "update"); err != nil {
		return fmt.Errorf("更新包索引失败: %w", err)
	}

	// 安装包
	return runCommand("sudo", "apt-get", "install", "-y", packageName)
}

// installWithWinget 通过 winget 静默安装指定包（最新版）。
// 安装成功后会尝试从 User+Machine PATH 刷新当前进程 PATH；是否安装成功由调用方 IsInstalled 判断。
func installWithWinget(packageId string) error {
	return installWithWingetVersion(packageId, "")
}

// installWithWingetVersion 通过 winget 静默安装指定包，version 非空且非 "latest" 时附加 --version。
func installWithWingetVersion(packageId, version string) error {
	if !commandExists("winget") {
		return fmt.Errorf("winget 未安装。请从 Microsoft Store 安装 App Installer")
	}

	// 检查 winget 版本
	cmd := exec.Command("winget", "--version")
	out, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("无法获取 winget 版本: %w", err)
	}
	ui.Info("检测到 winget 版本: %s", strings.TrimSpace(string(out)))

	args := []string{
		"install",
		"--id", packageId,
		"-e",
		"--silent",
		"--accept-package-agreements",
		"--accept-source-agreements",
		"--disable-interactivity",
	}
	if ver := strings.TrimSpace(version); ver != "" && !strings.EqualFold(ver, "latest") {
		args = append(args, "--version", ver)
	}

	if err := runCommand("winget", args...); err != nil {
		return err
	}

	// winget 安装后新二进制可能仅在注册表 PATH 中，刷新当前进程 PATH
	refreshProcessPathFromRegistry()
	return nil
}

// runWindowsScript 通过 cmd /C 执行 .bat/.cmd 等脚本（Windows 上直接 exec .bat 常失败）。
func runWindowsScript(script string, args ...string) error {
	cmdArgs := append([]string{"/C", script}, args...)
	return runCommand("cmd", cmdArgs...)
}

// GetToolInstaller 获取工具安装器
func GetToolInstaller(tool string) ToolInstaller {
	switch tool {
	case "git":
		return &GitInstaller{}
	case "docker":
		return &DockerInstaller{}
	case "vscode", "code":
		return &VSCodeInstaller{}
	case "uv":
		return &UvInstaller{}
	case "miniconda", "conda":
		return &MinicondaInstaller{}
	case "kubectl":
		return &KubectlInstaller{}
	case "minikube":
		return &MinikubeInstaller{}
	case "espidf", "esp-idf":
		return &EspIdfInstaller{}
	case "android", "android-sdk":
		return &AndroidInstaller{}
	default:
		return nil
	}
}

// UvInstaller Astral uv（极速 Python 包/项目管理器，亦可安装 CPython）
type UvInstaller struct{}

func (u *UvInstaller) Install() error {
	spinner := ui.NewSpinner("正在安装 uv")
	spinner.Start()
	defer spinner.Stop()

	already := u.IsInstalled()
	if already {
		ui.Info("uv 已安装: %s", strings.TrimSpace(u.GetVersion()))
	} else {
		var err error
		switch runtime.GOOS {
		case "darwin", "linux":
			err = u.installUnix()
		case "windows":
			err = u.installWindows()
		default:
			return fmt.Errorf("不支持的操作系统: %s", runtime.GOOS)
		}
		if err != nil {
			return err
		}
	}

	// 无论是否已安装，都确保 ~/.local/bin 写入 shell（官方脚本默认安装位置）
	if err := locateAndAddUvToPath(); err != nil && !u.IsInstalled() {
		return err
	}
	if !u.IsInstalled() {
		return fmt.Errorf("uv 安装完成但未检测到 uv 命令（请检查 PATH 或重开终端）")
	}
	if already {
		ui.Success("已修复 PATH 配置，重开终端后可直接使用 uv")
	} else {
		ui.Success("uv 安装成功: %s", strings.TrimSpace(u.GetVersion()))
	}
	return u.recordManifest()
}

func (u *UvInstaller) recordManifest() error {
	paths := []string{"~/.local/bin/uv", "~/.local/bin/uvx", "~/.cargo/bin/uv"}
	if runtime.GOOS == "windows" {
		paths = []string{"~/.local/bin/uv.exe", "~/.local/bin/uvx.exe"}
	}
	return recordInstall("uv", "tool", "latest", paths, []string{"# envkit:uv"})
}

func (u *UvInstaller) installUnix() error {
	// 依赖 curl；最小化系统可能缺失
	if err := CheckAndInstallDependencies([]SystemDependency{CommonDependencies[0]}); err != nil {
		ui.Warning("系统依赖检测: %v", err)
	}
	ui.Info("通过官方安装脚本安装 uv...")
	if err := runCommand("sh", "-c", "curl -LsSf https://astral.sh/uv/install.sh | sh"); err != nil {
		return fmt.Errorf("安装 uv 失败: %w", err)
	}
	return nil
}

func (u *UvInstaller) installWindows() error {
	// 优先 winget（包 ID: astral-sh.uv）
	if commandExists("winget") {
		ui.Info("尝试通过 winget 安装 uv...")
		if err := installWithWinget("astral-sh.uv"); err == nil {
			refreshProcessPathFromRegistry()
			if u.IsInstalled() {
				return nil
			}
		} else {
			ui.Warning("winget 安装 uv 失败 (%v)，改用官方 PowerShell 脚本", err)
		}
	}
	ui.Info("通过官方 PowerShell 脚本安装 uv...")
	if err := runCommand("powershell", "-NoProfile", "-ExecutionPolicy", "Bypass",
		"-Command", "irm https://astral.sh/uv/install.ps1 | iex"); err != nil {
		return fmt.Errorf("安装 uv 失败: %w", err)
	}
	refreshProcessPathFromRegistry()
	return nil
}

func (u *UvInstaller) IsInstalled() bool {
	return commandExists("uv")
}

func (u *UvInstaller) GetVersion() string {
	cmd := exec.Command("uv", "--version")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

// MinicondaInstaller Miniconda 安装器
type MinicondaInstaller struct{}

func (m *MinicondaInstaller) Install() error {
	spinner := ui.NewSpinner("正在安装 Miniconda")
	spinner.Start()
	defer spinner.Stop()

	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	prefix := filepath.Join(home, "miniconda3")

	// 1. 下载并安装
	if err := m.installMiniconda(prefix); err != nil {
		return err
	}

	// 2. 配置清华镜像源 (.condarc)
	if err := m.configureTsinghuaMirror(); err != nil {
		ui.Warning("配置清华镜像源失败: %v", err)
	}

	// 3. 配置环境变量
	binDir := filepath.Join(prefix, "bin")
	if runtime.GOOS == "windows" {
		// Windows Miniconda: Scripts + condabin 均需在 PATH（目录存在时才写入）
		scriptsDir := filepath.Join(prefix, "Scripts")
		condabinDir := filepath.Join(prefix, "condabin")
		if st, err := os.Stat(scriptsDir); err == nil && st.IsDir() {
			AddDirToPath(scriptsDir)
			_ = PersistPathEnv(scriptsDir)
		}
		if st, err := os.Stat(condabinDir); err == nil && st.IsDir() {
			AddDirToPath(condabinDir)
			_ = PersistPathEnv(condabinDir)
		}
		// winget 安装可能写到其它位置，再刷一次注册表 PATH
		refreshProcessPathFromRegistry()
		binDir = scriptsDir
	} else {
		AddDirToPath(binDir)
		_ = PersistPathEnv(binDir)
	}

	// 运行 conda init
	if runtime.GOOS == "windows" {
		condaBat := filepath.Join(prefix, "condabin", "conda.bat")
		if _, err := os.Stat(condaBat); err == nil {
			_ = runWindowsScript(condaBat, "init", "--all")
		} else if commandExists("conda") {
			_ = runWindowsScript("conda", "init", "--all")
		}
	} else {
		condaBin := filepath.Join(binDir, "conda")
		initCmd := exec.Command(condaBin, "init", "--all")
		_ = initCmd.Run()
	}

	if !m.IsInstalled() {
		return fmt.Errorf("Miniconda 安装完成但未检测到 conda 命令（请检查 PATH 或重开终端）")
	}

	if err := recordInstall("miniconda", "tool", "latest", []string{"~/miniconda3", "~/.condarc"}, []string{"# >>> conda initialize >>>", "# envkit:miniconda"}); err != nil {
		return err
	}

	return nil
}

func (m *MinicondaInstaller) installMiniconda(prefix string) error {
	var downloadURL string
	arch := runtime.GOARCH

	switch runtime.GOOS {
	case "darwin":
		if arch == "arm64" {
			downloadURL = "https://mirrors.tuna.tsinghua.edu.cn/anaconda/miniconda/Miniconda3-latest-MacOSX-arm64.sh"
		} else {
			downloadURL = "https://mirrors.tuna.tsinghua.edu.cn/anaconda/miniconda/Miniconda3-latest-MacOSX-x86_64.sh"
		}
		return m.installUnix(downloadURL, prefix)

	case "linux":
		if arch == "arm64" {
			downloadURL = "https://mirrors.tuna.tsinghua.edu.cn/anaconda/miniconda/Miniconda3-latest-Linux-aarch64.sh"
		} else {
			downloadURL = "https://mirrors.tuna.tsinghua.edu.cn/anaconda/miniconda/Miniconda3-latest-Linux-x86_64.sh"
		}
		return m.installUnix(downloadURL, prefix)

	case "windows":
		if arch == "arm64" {
			// 优先 arm64 安装包；若镜像无此包再回退 x86_64 或 winget
			downloadURL = "https://mirrors.tuna.tsinghua.edu.cn/anaconda/miniconda/Miniconda3-latest-Windows-arm64.exe"
			if err := m.installWindows(downloadURL, prefix); err == nil {
				return nil
			}
			ui.Warning("arm64 Miniconda 安装包失败，尝试 winget...")
			if err := installWithWinget("Anaconda.Miniconda3"); err == nil {
				refreshProcessPathFromRegistry()
				return nil
			}
			ui.Warning("winget 安装 Miniconda 失败，回退 x86_64 安装包...")
		}
		downloadURL = "https://mirrors.tuna.tsinghua.edu.cn/anaconda/miniconda/Miniconda3-latest-Windows-x86_64.exe"
		return m.installWindows(downloadURL, prefix)

	default:
		return fmt.Errorf("不支持的操作系统: %s", runtime.GOOS)
	}
}

func (m *MinicondaInstaller) installUnix(downloadURL, prefix string) error {
	// 检查依赖
	if err := CheckAndInstallDependencies([]SystemDependency{
		CommonDependencies[0], // curl
		CommonDependencies[4], // bash
	}); err != nil {
		return err
	}

	shPath := filepath.Join(os.TempDir(), "miniconda.sh")

	// 下载安装脚本
	if err := runCommand("curl", "-fsSL", "-o", shPath, downloadURL); err != nil {
		return fmt.Errorf("下载 Miniconda 失败: %w", err)
	}
	defer os.Remove(shPath)

	// 静默安装
	_ = os.RemoveAll(prefix)
	if err := runCommand("sh", shPath, "-b", "-p", prefix); err != nil {
		return fmt.Errorf("执行 Miniconda 安装脚本失败: %w", err)
	}

	return nil
}

func (m *MinicondaInstaller) installWindows(downloadURL, prefix string) error {
	exePath := filepath.Join(os.Getenv("TEMP"), "miniconda.exe")

	// 下载安装包
	if err := runCommand("curl", "-fsSL", "-o", exePath, downloadURL); err != nil {
		return fmt.Errorf("下载 Miniconda 失败: %w", err)
	}
	defer os.Remove(exePath)

	// 静默安装
	_ = os.RemoveAll(prefix)
	if err := runCommand(exePath, "/S", "/RegisterPython=0", "/D="+prefix); err != nil {
		return fmt.Errorf("执行 Miniconda 安装程序失败: %w", err)
	}

	return nil
}

func (m *MinicondaInstaller) configureTsinghuaMirror() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	condarcPath := filepath.Join(home, ".condarc")

	content := `channels:
  - defaults
show_channel_urls: true
default_channels:
  - https://mirrors.tuna.tsinghua.edu.cn/anaconda/pkgs/main
  - https://mirrors.tuna.tsinghua.edu.cn/anaconda/pkgs/r
  - https://mirrors.tuna.tsinghua.edu.cn/anaconda/pkgs/msys2
custom_channels:
  conda-forge: https://mirrors.tuna.tsinghua.edu.cn/anaconda/cloud
  msys2: https://mirrors.tuna.tsinghua.edu.cn/anaconda/cloud
  bioconda: https://mirrors.tuna.tsinghua.edu.cn/anaconda/cloud
  menpo: https://mirrors.tuna.tsinghua.edu.cn/anaconda/cloud
  pytorch: https://mirrors.tuna.tsinghua.edu.cn/anaconda/cloud
  pytorch-lts: https://mirrors.tuna.tsinghua.edu.cn/anaconda/cloud
  simpleitk: https://mirrors.tuna.tsinghua.edu.cn/anaconda/cloud
  deepmodeling: https://mirrors.tuna.tsinghua.edu.cn/anaconda/cloud/
`
	return os.WriteFile(condarcPath, []byte(content), 0644)
}

func (m *MinicondaInstaller) IsInstalled() bool {
	return commandExists("conda")
}

func (m *MinicondaInstaller) GetVersion() string {
	cmd := exec.Command("conda", "--version")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

// KubectlInstaller Kubectl 安装器
type KubectlInstaller struct{}

func (k *KubectlInstaller) Install() error {
	spinner := ui.NewSpinner("正在安装 kubectl")
	spinner.Start()
	defer spinner.Stop()

	var err error
	switch runtime.GOOS {
	case "darwin":
		err = k.installOnDarwin()
	case "linux":
		err = k.installOnLinux()
	case "windows":
		err = installWithWinget("Kubernetes.kubectl")
	default:
		return fmt.Errorf("不支持的操作系统: %s", runtime.GOOS)
	}

	if err != nil {
		return err
	}
	if !k.IsInstalled() {
		return fmt.Errorf("kubectl 安装完成但未检测到命令")
	}
	switch runtime.GOOS {
	case "darwin", "linux":
		if err := recordInstall("kubectl", "tool", "v1.30.0", []string{"~/.local/bin/kubectl"}, []string{"# envkit:kubectl"}); err != nil {
			return err
		}
	default:
		if err := recordInstall("kubectl", "tool", "v1.30.0", nil, []string{"# envkit:kubectl"}); err != nil {
			return err
		}
	}
	return nil
}

func (k *KubectlInstaller) installOnDarwin() error {
	err := installWithBrew("kubernetes-cli")
	if err == nil {
		return nil
	}

	ui.Warning("通过 Homebrew 安装 kubectl 失败 (%v)，正在尝试直接下载...", err)
	arch := runtime.GOARCH
	downloadURL := fmt.Sprintf("https://dl.k8s.io/release/v1.30.0/bin/darwin/%s/kubectl", arch)
	return k.downloadAndInstallUnix(downloadURL)
}

func (k *KubectlInstaller) installOnLinux() error {
	arch := runtime.GOARCH
	downloadURL := fmt.Sprintf("https://dl.k8s.io/release/v1.30.0/bin/linux/%s/kubectl", arch)
	return k.downloadAndInstallUnix(downloadURL)
}

func (k *KubectlInstaller) downloadAndInstallUnix(downloadURL string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	binDir := filepath.Join(home, ".local", "bin")
	_ = os.MkdirAll(binDir, 0755)

	destPath := filepath.Join(binDir, "kubectl")
	if err := runCommand("curl", "-fsSL", "-o", destPath, downloadURL); err != nil {
		return fmt.Errorf("下载 kubectl 失败: %w", err)
	}

	if err := os.Chmod(destPath, 0755); err != nil {
		return fmt.Errorf("为 kubectl 赋予执行权限失败: %w", err)
	}

	AddDirToPath(binDir)
	_ = PersistPathEnv(binDir)
	return nil
}

func (k *KubectlInstaller) IsInstalled() bool {
	return commandExists("kubectl")
}

func (k *KubectlInstaller) GetVersion() string {
	cmd := exec.Command("kubectl", "version", "--client", "--output=yaml")
	output, err := cmd.CombinedOutput()
	if err != nil {
		cmd = exec.Command("kubectl", "version", "--client")
		output, err = cmd.CombinedOutput()
		if err != nil {
			return ""
		}
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "gitVersion:") {
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				return strings.Trim(parts[1], ` "'`)
			}
		}
	}
	return strings.TrimSpace(string(output))
}

// MinikubeInstaller Minikube 安装器
type MinikubeInstaller struct{}

func (m *MinikubeInstaller) Install() error {
	spinner := ui.NewSpinner("正在安装 minikube")
	spinner.Start()
	defer spinner.Stop()

	var err error
	switch runtime.GOOS {
	case "darwin":
		err = m.installOnDarwin()
	case "linux":
		err = m.installOnLinux()
	case "windows":
		err = installWithWinget("Kubernetes.Minikube")
	default:
		return fmt.Errorf("不支持的操作系统: %s", runtime.GOOS)
	}

	if err != nil {
		return err
	}
	if !m.IsInstalled() {
		return fmt.Errorf("minikube 安装完成但未检测到命令")
	}
	switch runtime.GOOS {
	case "darwin", "linux":
		if err := recordInstall("minikube", "tool", "latest", []string{"~/.local/bin/minikube"}, []string{"# envkit:minikube"}); err != nil {
			return err
		}
	default:
		if err := recordInstall("minikube", "tool", "latest", nil, []string{"# envkit:minikube"}); err != nil {
			return err
		}
	}
	return nil
}

func (m *MinikubeInstaller) installOnDarwin() error {
	err := installWithBrew("minikube")
	if err == nil {
		return nil
	}

	ui.Warning("通过 Homebrew 安装 minikube 失败 (%v)，正在尝试直接下载...", err)
	arch := runtime.GOARCH
	downloadURL := fmt.Sprintf("https://storage.googleapis.com/minikube/releases/latest/minikube-darwin-%s", arch)
	return m.downloadAndInstallUnix(downloadURL)
}

func (m *MinikubeInstaller) installOnLinux() error {
	arch := runtime.GOARCH
	downloadURL := fmt.Sprintf("https://storage.googleapis.com/minikube/releases/latest/minikube-linux-%s", arch)
	return m.downloadAndInstallUnix(downloadURL)
}

func (m *MinikubeInstaller) downloadAndInstallUnix(downloadURL string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	binDir := filepath.Join(home, ".local", "bin")
	_ = os.MkdirAll(binDir, 0755)

	destPath := filepath.Join(binDir, "minikube")
	if err := runCommand("curl", "-fsSL", "-o", destPath, downloadURL); err != nil {
		return fmt.Errorf("下载 minikube 失败: %w", err)
	}

	if err := os.Chmod(destPath, 0755); err != nil {
		return fmt.Errorf("为 minikube 赋予执行权限失败: %w", err)
	}

	AddDirToPath(binDir)
	_ = PersistPathEnv(binDir)
	return nil
}

func (m *MinikubeInstaller) IsInstalled() bool {
	return commandExists("minikube")
}

func (m *MinikubeInstaller) GetVersion() string {
	cmd := exec.Command("minikube", "version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

// EspIdfInstaller ESP-IDF (EIM) 安装器
type EspIdfInstaller struct{}

func (e *EspIdfInstaller) Install() error {
	spinner := ui.NewSpinner("正在安装 ESP-IDF (EIM)")
	spinner.Start()
	defer spinner.Stop()

	var err error
	switch runtime.GOOS {
	case "darwin":
		err = e.installOnDarwin()
	case "linux":
		err = e.installOnLinux()
	case "windows":
		err = installWithWinget("Espressif.EIM-CLI")
	default:
		return fmt.Errorf("不支持的操作系统: %s", runtime.GOOS)
	}

	if err != nil {
		return err
	}
	if !e.IsInstalled() {
		return fmt.Errorf("ESP-IDF/EIM 安装完成但未检测到 eim 命令")
	}
	if err := recordInstall("espidf", "tool", "latest", []string{"~/.espressif"}, []string{"# envkit:espidf", ".espressif/export.sh"}); err != nil {
		return err
	}
	return nil
}

func (e *EspIdfInstaller) installOnDarwin() error {
	if !commandExists("brew") {
		return fmt.Errorf("请先安装 Homebrew: https://brew.sh")
	}

	// 添加 EIM tap
	_ = runCommand("brew", "tap", "espressif/eim")
	// 安装 CLI 版本
	if err := runCommand("brew", "install", "eim"); err != nil {
		// 双重检查：如果命令已经存在，则认为安装成功
		if e.IsInstalled() {
			return nil
		}
		return err
	}
	return nil
}

func (e *EspIdfInstaller) installOnLinux() error {
	if commandExists("apt-get") {
		ui.Info("正在配置 ESP-IDF APT 软件源...")
		if err := runCommand("bash", "-c", `echo "deb [trusted=yes] https://dl.espressif.com/dl/eim/apt/ stable main" | sudo tee /etc/apt/sources.list.d/espressif.list`); err != nil {
			return fmt.Errorf("添加 APT 软件源失败: %w", err)
		}

		_ = runCommand("sudo", "apt-get", "update")

		ui.Info("正在通过 APT 安装 eim-cli...")
		return runCommand("sudo", "apt-get", "install", "-y", "eim-cli")
	} else if commandExists("dnf") {
		ui.Info("正在通过 DNF 安装 RPM 仓库配置...")
		if err := runCommand("sudo", "dnf", "install", "-y", "https://dl.espressif.com/dl/eim/rpm/eim-repo-latest.noarch.rpm"); err != nil {
			return fmt.Errorf("安装 RPM 仓库配置失败: %w", err)
		}

		ui.Info("正在通过 DNF 安装 eim-cli...")
		return runCommand("sudo", "dnf", "install", "-y", "eim-cli")
	}

	return fmt.Errorf("不支持的 Linux 包管理器（需要 apt 或 dnf）")
}

func (e *EspIdfInstaller) IsInstalled() bool {
	if commandExists("eim") {
		return true
	}
	if runtime.GOOS == "darwin" {
		if fileExists("/opt/homebrew/bin/eim") {
			AddDirToPath("/opt/homebrew/bin")
			return true
		}
		if fileExists("/usr/local/bin/eim") {
			AddDirToPath("/usr/local/bin")
			return true
		}
		if commandExists("brew") && exec.Command("brew", "list", "eim").Run() == nil {
			out, err := exec.Command("brew", "--prefix", "eim").Output()
			if err == nil {
				prefix := strings.TrimSpace(string(out))
				if prefix != "" {
					binDir := filepath.Join(prefix, "bin")
					if fileExists(filepath.Join(binDir, "eim")) {
						AddDirToPath(binDir)
						return true
					}
				}
			}
			return true
		}
	}
	return false
}

func (e *EspIdfInstaller) GetVersion() string {
	cmd := exec.Command("eim", "--version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// runCommand 执行外部命令并捕获输出。
// TUI 静默模式下禁止写 stdout/stdin，避免破坏 alt-screen 与 sudo 假交互。
func runCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)

	var stdoutBuf, stderrBuf bytes.Buffer

	if ui.IsSilentMode() {
		cmd.Stdout = &stdoutBuf
		cmd.Stderr = &stderrBuf
		cmd.Stdin = nil
	} else {
		stdoutWriter := NewWindowsConsoleWriter(os.Stdout)
		stderrWriter := NewWindowsConsoleWriter(os.Stderr)
		cmd.Stdout = io.MultiWriter(stdoutWriter, &stdoutBuf)
		cmd.Stderr = io.MultiWriter(stderrWriter, &stderrBuf)
		cmd.Stdin = os.Stdin
	}

	err := cmd.Run()
	if err != nil {
		stderrStr := strings.TrimSpace(stderrBuf.String())
		if stderrStr == "" {
			stderrStr = strings.TrimSpace(stdoutBuf.String())
		}
		if stderrStr != "" {
			return fmt.Errorf("%w\n【错误原因/详情】:\n%s", err, stderrStr)
		}
		return err
	}
	return nil
}
