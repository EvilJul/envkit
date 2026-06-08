package installer

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"

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

	switch runtime.GOOS {
	case "darwin":
		return installWithBrew("git")
	case "linux":
		return installWithApt("git")
	case "windows":
		return installWithWinget("Git.Git")
	default:
		return fmt.Errorf("不支持的操作系统: %s", runtime.GOOS)
	}
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
	ui.Info("正在安装 Docker Engine...")

	// 更新包索引
	updateCmd := exec.Command("sudo", "apt-get", "update")
	if err := updateCmd.Run(); err != nil {
		return fmt.Errorf("更新包索引失败: %w", err)
	}

	// 安装依赖
	prereqCmd := exec.Command("sudo", "apt-get", "install", "-y",
		"ca-certificates", "curl", "gnupg", "lsb-release")
	if err := prereqCmd.Run(); err != nil {
		return fmt.Errorf("安装依赖失败: %w", err)
	}

	// 添加 Docker GPG 密钥
	keyCmd := exec.Command("bash", "-c",
		"curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg")
	if err := keyCmd.Run(); err != nil {
		return fmt.Errorf("添加 GPG 密钥失败: %w", err)
	}

	// 添加 Docker 仓库
	repoCmd := exec.Command("bash", "-c",
		`echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null`)
	if err := repoCmd.Run(); err != nil {
		return fmt.Errorf("添加仓库失败: %w", err)
	}

	// 更新包索引
	updateCmd2 := exec.Command("sudo", "apt-get", "update")
	if err := updateCmd2.Run(); err != nil {
		return fmt.Errorf("更新包索引失败: %w", err)
	}

	// 安装 Docker
	installCmd := exec.Command("sudo", "apt-get", "install", "-y",
		"docker-ce", "docker-ce-cli", "containerd.io", "docker-compose-plugin")
	installCmd.Stdout = os.Stdout
	installCmd.Stderr = os.Stderr
	if err := installCmd.Run(); err != nil {
		return fmt.Errorf("安装 Docker 失败: %w", err)
	}

	// 启动 Docker 服务
	startCmd := exec.Command("sudo", "systemctl", "start", "docker")
	if err := startCmd.Run(); err != nil {
		ui.Warning("启动 Docker 服务失败: %v", err)
	}

	// 启用开机自启
	enableCmd := exec.Command("sudo", "systemctl", "enable", "docker")
	if err := enableCmd.Run(); err != nil {
		ui.Warning("设置开机自启失败: %v", err)
	}

	ui.Success("Docker 安装成功！")
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

	switch runtime.GOOS {
	case "darwin":
		return installWithBrew("--cask", "visual-studio-code")
	case "linux":
		return v.installOnLinux()
	case "windows":
		return installWithWinget("Microsoft.VisualStudioCode")
	default:
		return fmt.Errorf("不支持的操作系统: %s", runtime.GOOS)
	}
}

func (v *VSCodeInstaller) installOnLinux() error {
	// 下载并安装 VSCode .deb 包
	ui.Info("正在下载 VSCode...")

	downloadCmd := exec.Command("curl", "-fsSL", "-o", "/tmp/vscode.deb",
		"https://code.visualstudio.com/sha/download?build=stable&os=linux-deb-x64")
	if err := downloadCmd.Run(); err != nil {
		return fmt.Errorf("下载失败: %w", err)
	}

	ui.Info("正在安装...")
	installCmd := exec.Command("sudo", "dpkg", "-i", "/tmp/vscode.deb")
	if err := installCmd.Run(); err != nil {
		// 修复依赖
		fixCmd := exec.Command("sudo", "apt-get", "install", "-f", "-y")
		fixCmd.Run()
	}

	// 清理临时文件
	os.Remove("/tmp/vscode.deb")

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
	cmd := exec.Command("brew", fullArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func installWithApt(packageName string) error {
	// 更新包索引
	updateCmd := exec.Command("sudo", "apt-get", "update")
	if err := updateCmd.Run(); err != nil {
		return fmt.Errorf("更新包索引失败: %w", err)
	}

	// 安装包
	installCmd := exec.Command("sudo", "apt-get", "install", "-y", packageName)
	installCmd.Stdout = os.Stdout
	installCmd.Stderr = os.Stderr
	return installCmd.Run()
}

func installWithWinget(packageId string) error {
	if !commandExists("winget") {
		return fmt.Errorf("请先安装 winget")
	}

	cmd := exec.Command("winget", "install", packageId)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
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
	default:
		return nil
	}
}
