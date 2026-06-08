package installer

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"github.com/fusheng/envkit/internal/ui"
)

// LanguageInstaller 语言安装器接口
type LanguageInstaller interface {
	Install(version string) error
	IsInstalled() bool
	GetVersion() string
}

// NodeInstaller Node.js 安装器
type NodeInstaller struct{}

func (n *NodeInstaller) Install(version string) error {
	spinner := ui.NewSpinner("正在安装 Node.js " + version)
	spinner.Start()
	defer spinner.Stop()

	switch runtime.GOOS {
	case "darwin":
		return n.installWithBrew(version)
	case "linux":
		return n.installWithFnm(version)
	case "windows":
		return n.installWithWinget(version)
	default:
		return fmt.Errorf("不支持的操作系统: %s", runtime.GOOS)
	}
}

func (n *NodeInstaller) installWithBrew(version string) error {
	// 检查是否安装了 Homebrew
	if !commandExists("brew") {
		return fmt.Errorf("请先安装 Homebrew: https://brew.sh")
	}

	cmd := exec.Command("brew", "install", "node@"+version)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (n *NodeInstaller) installWithFnm(version string) error {
	// 检查是否安装了 fnm
	if !commandExists("fnm") {
		ui.Info("正在安装 fnm (Fast Node Manager)...")
		installCmd := exec.Command("curl", "-fsSL", "https://fnm.vercel.app/install", "|", "bash")
		if err := installCmd.Run(); err != nil {
			return fmt.Errorf("安装 fnm 失败: %w", err)
		}
	}

	// 使用 fnm 安装 Node.js
	cmd := exec.Command("fnm", "install", version)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("安装 Node.js 失败: %w", err)
	}

	// 设置为默认版本
	defaultCmd := exec.Command("fnm", "default", version)
	return defaultCmd.Run()
}

func (n *NodeInstaller) installWithWinget(version string) error {
	if !commandExists("winget") {
		return fmt.Errorf("请先安装 winget")
	}

	cmd := exec.Command("winget", "install", "OpenJS.NodeJS")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (n *NodeInstaller) IsInstalled() bool {
	return commandExists("node")
}

func (n *NodeInstaller) GetVersion() string {
	cmd := exec.Command("node", "--version")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return string(output)
}

// PythonInstaller Python 安装器
type PythonInstaller struct{}

func (p *PythonInstaller) Install(version string) error {
	spinner := ui.NewSpinner("正在安装 Python " + version)
	spinner.Start()
	defer spinner.Stop()

	switch runtime.GOOS {
	case "darwin":
		return p.installWithBrew(version)
	case "linux":
		return p.installWithUv(version)
	case "windows":
		return p.installWithWinget(version)
	default:
		return fmt.Errorf("不支持的操作系统: %s", runtime.GOOS)
	}
}

func (p *PythonInstaller) installWithBrew(version string) error {
	if !commandExists("brew") {
		return fmt.Errorf("请先安装 Homebrew: https://brew.sh")
	}

	cmd := exec.Command("brew", "install", "python@"+version)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (p *PythonInstaller) installWithUv(version string) error {
	// 检查是否安装了 uv
	if !commandExists("uv") {
		ui.Info("正在安装 uv (Python 包管理器)...")
		installCmd := exec.Command("curl", "-LsSf", "https://astral.sh/uv/install.sh", "|", "sh")
		if err := installCmd.Run(); err != nil {
			return fmt.Errorf("安装 uv 失败: %w", err)
		}
	}

	// 使用 uv 安装 Python
	cmd := exec.Command("uv", "python", "install", version)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (p *PythonInstaller) installWithWinget(version string) error {
	if !commandExists("winget") {
		return fmt.Errorf("请先安装 winget")
	}

	cmd := exec.Command("winget", "install", "Python.Python."+version)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (p *PythonInstaller) IsInstalled() bool {
	return commandExists("python3") || commandExists("python")
}

func (p *PythonInstaller) GetVersion() string {
	var cmd *exec.Cmd
	if commandExists("python3") {
		cmd = exec.Command("python3", "--version")
	} else {
		cmd = exec.Command("python", "--version")
	}
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return string(output)
}

// GoInstaller Go 安装器
type GoInstaller struct{}

func (g *GoInstaller) Install(version string) error {
	spinner := ui.NewSpinner("正在安装 Go " + version)
	spinner.Start()
	defer spinner.Stop()

	switch runtime.GOOS {
	case "darwin":
		return g.installWithBrew(version)
	case "linux":
		return g.installFromSource(version)
	case "windows":
		return g.installWithWinget(version)
	default:
		return fmt.Errorf("不支持的操作系统: %s", runtime.GOOS)
	}
}

func (g *GoInstaller) installWithBrew(version string) error {
	if !commandExists("brew") {
		return fmt.Errorf("请先安装 Homebrew: https://brew.sh")
	}

	cmd := exec.Command("brew", "install", "go@"+version)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (g *GoInstaller) installFromSource(version string) error {
	ui.Info("从官方下载 Go %s...", version)

	// 下载 Go 安装包
	arch := runtime.GOARCH
	downloadURL := fmt.Sprintf("https://go.dev/dl/go%s.linux-%s.tar.gz", version, arch)

	downloadCmd := exec.Command("curl", "-fsSL", "-o", "/tmp/go.tar.gz", downloadURL)
	if err := downloadCmd.Run(); err != nil {
		return fmt.Errorf("下载失败: %w", err)
	}

	// 解压安装
	ui.Info("正在安装...")
	extractCmd := exec.Command("sudo", "tar", "-C", "/usr/local", "-xzf", "/tmp/go.tar.gz")
	if err := extractCmd.Run(); err != nil {
		return fmt.Errorf("安装失败: %w", err)
	}

	// 清理临时文件
	os.Remove("/tmp/go.tar.gz")

	ui.Success("Go %s 安装成功！请将 /usr/local/go/bin 添加到 PATH", version)
	return nil
}

func (g *GoInstaller) installWithWinget(version string) error {
	if !commandExists("winget") {
		return fmt.Errorf("请先安装 winget")
	}

	cmd := exec.Command("winget", "install", "GoLang.Go")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (g *GoInstaller) IsInstalled() bool {
	return commandExists("go")
}

func (g *GoInstaller) GetVersion() string {
	cmd := exec.Command("go", "version")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return string(output)
}

// RustInstaller Rust 安装器
type RustInstaller struct{}

func (r *RustInstaller) Install(version string) error {
	spinner := ui.NewSpinner("正在安装 Rust " + version)
	spinner.Start()
	defer spinner.Stop()

	// Rust 通常通过 rustup 安装
	if !commandExists("rustup") {
		ui.Info("正在安装 rustup...")
		installCmd := exec.Command("curl", "--proto", "=https", "--tlsv1.2", "-sSf", "https://sh.rustup.rs", "|", "sh", "-s", "--", "-y")
		if err := installCmd.Run(); err != nil {
			return fmt.Errorf("安装 rustup 失败: %w", err)
		}
	}

	// 安装指定版本
	cmd := exec.Command("rustup", "install", version)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (r *RustInstaller) IsInstalled() bool {
	return commandExists("rustc")
}

func (r *RustInstaller) GetVersion() string {
	cmd := exec.Command("rustc", "--version")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return string(output)
}

// commandExists 检查命令是否存在
func commandExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

// GetInstaller 获取语言安装器
func GetInstaller(language string) LanguageInstaller {
	switch language {
	case "node", "nodejs":
		return &NodeInstaller{}
	case "python", "python3":
		return &PythonInstaller{}
	case "go", "golang":
		return &GoInstaller{}
	case "rust":
		return &RustInstaller{}
	default:
		return nil
	}
}
