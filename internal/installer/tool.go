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

	if err == nil {
		RecordInstallation("git", "tool", "latest", nil, nil)
	}
	return err
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
	RecordInstallation("docker", "tool", "latest", nil, nil)
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
		if err == nil {
			RecordInstallation("vscode", "tool", "stable", []string{"/Applications/Visual Studio Code.app", "~/.local/bin/code", "/usr/local/bin/code"}, nil)
		}
	case "linux":
		err = v.installOnLinux()
		if err == nil {
			RecordInstallation("vscode", "tool", "stable", nil, nil)
		}
	case "windows":
		err = installWithWinget("Microsoft.VisualStudioCode")
		if err == nil {
			RecordInstallation("vscode", "tool", "stable", nil, nil)
		}
	default:
		return fmt.Errorf("不支持的操作系统: %s", runtime.GOOS)
	}
	return err
}

func (v *VSCodeInstaller) installOnDarwin() error {
	// 尝试通过 Homebrew Cask 安装
	err := installWithBrew("--cask", "visual-studio-code")
	if err == nil {
		return nil
	}

	// 备用方案：直接从 VSCode 官网下载 zip 并安装
	ui.Warning("通过 Homebrew Cask 安装 VSCode 失败 (%v)，正在尝试直接从官方下载...", err)

	downloadURL := "https://update.code.visualstudio.com/latest/darwin-universal/stable"
	zipPath := "/tmp/vscode.zip"

	downloadCmd := exec.Command("curl", "-L", "-o", zipPath, downloadURL)
	if err := downloadCmd.Run(); err != nil {
		return fmt.Errorf("下载 VSCode 失败: %w", err)
	}

	// 解压
	unzipCmd := exec.Command("unzip", "-q", zipPath, "-d", "/tmp/vscode-extracted")
	if err := unzipCmd.Run(); err != nil {
		os.Remove(zipPath)
		return fmt.Errorf("解压 VSCode 失败: %w", err)
	}

	// 移动至 /Applications
	destPath := "/Applications/Visual Studio Code.app"
	_ = exec.Command("rm", "-rf", destPath).Run()

	moveCmd := exec.Command("mv", "/tmp/vscode-extracted/Visual Studio Code.app", destPath)
	if err := moveCmd.Run(); err != nil {
		// 备用 sudo 移动
		_ = exec.Command("sudo", "mv", "/tmp/vscode-extracted/Visual Studio Code.app", destPath).Run()
	}

	// 移除 macOS quarantine 属性，防止首次运行时 Gatekeeper 弹窗阻塞终端
	_ = exec.Command("xattr", "-cr", destPath).Run()

	// 清理
	os.Remove(zipPath)
	_ = exec.Command("rm", "-rf", "/tmp/vscode-extracted").Run()

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
	case "miniconda", "conda":
		return &MinicondaInstaller{}
	case "kubectl":
		return &KubectlInstaller{}
	case "minikube":
		return &MinikubeInstaller{}
	default:
		return nil
	}
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
		binDir = filepath.Join(prefix, "Scripts")
	}
	AddDirToPath(binDir)
	_ = PersistPathEnv(binDir)

	// 运行 conda init
	condaBin := filepath.Join(binDir, "conda")
	if runtime.GOOS == "windows" {
		condaBin = filepath.Join(prefix, "condabin", "conda.bat")
	}

	initCmd := exec.Command(condaBin, "init", "--all")
	_ = initCmd.Run()

	RecordInstallation("miniconda", "tool", "latest", []string{"~/miniconda3", "~/.condarc"}, []string{"# >>> conda initialize >>>", "# added by envkit"})

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
		downloadURL = "https://mirrors.tuna.tsinghua.edu.cn/anaconda/miniconda/Miniconda3-latest-Windows-x86_64.exe"
		return m.installWindows(downloadURL, prefix)

	default:
		return fmt.Errorf("不支持的操作系统: %s", runtime.GOOS)
	}
}

func (m *MinicondaInstaller) installUnix(downloadURL, prefix string) error {
	shPath := "/tmp/miniconda.sh"

	// 下载安装脚本
	downloadCmd := exec.Command("curl", "-fsSL", "-o", shPath, downloadURL)
	if err := downloadCmd.Run(); err != nil {
		return fmt.Errorf("下载 Miniconda 失败: %w", err)
	}
	defer os.Remove(shPath)

	// 静默安装
	_ = os.RemoveAll(prefix)
	installCmd := exec.Command("sh", shPath, "-b", "-p", prefix)
	if err := installCmd.Run(); err != nil {
		return fmt.Errorf("执行 Miniconda 安装脚本失败: %w", err)
	}

	return nil
}

func (m *MinicondaInstaller) installWindows(downloadURL, prefix string) error {
	exePath := filepath.Join(os.Getenv("TEMP"), "miniconda.exe")

	// 下载安装包
	downloadCmd := exec.Command("curl", "-fsSL", "-o", exePath, downloadURL)
	if err := downloadCmd.Run(); err != nil {
		return fmt.Errorf("下载 Miniconda 失败: %w", err)
	}
	defer os.Remove(exePath)

	// 静默安装
	_ = os.RemoveAll(prefix)
	installCmd := exec.Command(exePath, "/S", "/RegisterPython=0", "/D="+prefix)
	if err := installCmd.Run(); err != nil {
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
		if err == nil {
			RecordInstallation("kubectl", "tool", "v1.30.0", []string{"~/.local/bin/kubectl"}, []string{"# added by envkit"})
		}
	case "linux":
		err = k.installOnLinux()
		if err == nil {
			RecordInstallation("kubectl", "tool", "v1.30.0", []string{"~/.local/bin/kubectl"}, []string{"# added by envkit"})
		}
	case "windows":
		err = installWithWinget("Kubernetes.kubectl")
		if err == nil {
			RecordInstallation("kubectl", "tool", "v1.30.0", nil, nil)
		}
	default:
		return fmt.Errorf("不支持的操作系统: %s", runtime.GOOS)
	}
	return err
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
	downloadCmd := exec.Command("curl", "-fsSL", "-o", destPath, downloadURL)
	if err := downloadCmd.Run(); err != nil {
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
		if err == nil {
			RecordInstallation("minikube", "tool", "latest", []string{"~/.local/bin/minikube"}, []string{"# added by envkit"})
		}
	case "linux":
		err = m.installOnLinux()
		if err == nil {
			RecordInstallation("minikube", "tool", "latest", []string{"~/.local/bin/minikube"}, []string{"# added by envkit"})
		}
	case "windows":
		err = installWithWinget("Kubernetes.Minikube")
		if err == nil {
			RecordInstallation("minikube", "tool", "latest", nil, nil)
		}
	default:
		return fmt.Errorf("不支持的操作系统: %s", runtime.GOOS)
	}
	return err
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
	downloadCmd := exec.Command("curl", "-fsSL", "-o", destPath, downloadURL)
	if err := downloadCmd.Run(); err != nil {
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
