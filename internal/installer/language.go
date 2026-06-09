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

	var err error
	switch runtime.GOOS {
	case "darwin":
		err = n.installWithBrew(version)
	case "linux":
		err = n.installWithFnm(version)
	case "windows":
		err = n.installWithWinget(version)
	default:
		return fmt.Errorf("不支持的操作系统: %s", runtime.GOOS)
	}

	if err == nil || n.IsInstalled() {
		switch runtime.GOOS {
		case "linux":
			RecordInstallation("node", "language", version, []string{"~/.fnm", "~/.local/share/fnm"}, []string{"# fnm integration", "# added by envkit"})
		default:
			RecordInstallation("node", "language", version, nil, []string{"# added by envkit"})
		}
		return nil
	}
	return err
}

func (n *NodeInstaller) installWithBrew(version string) error {
	// 检查是否安装了 Homebrew
	if !commandExists("brew") {
		return fmt.Errorf("请先安装 Homebrew: https://brew.sh")
	}

	return runCommand("brew", "install", "node@"+version)
}

func (n *NodeInstaller) installWithFnm(version string) error {
	// Windows 平台不支持此安装方式
	if runtime.GOOS == "windows" {
		return fmt.Errorf("Windows 平台请使用 winget 或从官网下载安装包")
	}

	// 检查依赖
	if err := CheckAndInstallDependencies([]SystemDependency{
		CommonDependencies[0], // curl
		CommonDependencies[4], // bash
	}); err != nil {
		return err
	}

	// 检查是否安装了 fnm
	if !commandExists("fnm") {
		ui.Info("正在安装 fnm (Fast Node Manager)...")

		// 尝试多种安装方式
		installURLs := []string{
			"https://fnm.vercel.app/install",
			"https://raw.githubusercontent.com/Schniz/fnm/master/.ci/install.sh",
		}

		var installErr error
		for i, url := range installURLs {
			if i > 0 {
				ui.Info("尝试备用安装源...")
			}
			installCmd := exec.Command("bash", "-c", fmt.Sprintf("curl -fsSL %s | bash", url))
			installCmd.Stdout = os.Stdout
			installCmd.Stderr = os.Stderr
			installCmd.Stdin = os.Stdin
			installErr = installCmd.Run()
			if installErr == nil {
				break
			}
		}

		if installErr != nil {
			return fmt.Errorf("安装 fnm 失败: %w", installErr)
		}

		if err := locateAndAddFnmToPath(); err != nil {
			return err
		}
	}

	// 使用 fnm 安装 Node.js，添加镜像支持
	ui.Info("正在使用 fnm 安装 Node.js %s...", version)

	// 设置 Node.js 镜像环境变量
	cmd := exec.Command("fnm", "install", version)
	cmd.Env = append(os.Environ(),
		"NODE_MIRROR=https://npmmirror.com/mirrors/node",
		"NVM_NODEJS_ORG_MIRROR=https://npmmirror.com/mirrors/node",
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		// 如果指定版本失败，尝试 latest
		ui.Warning("安装 Node.js %s 失败，尝试安装最新版本...", version)
		cmd = exec.Command("fnm", "install", "latest")
		cmd.Env = append(os.Environ(),
			"NODE_MIRROR=https://npmmirror.com/mirrors/node",
			"NVM_NODEJS_ORG_MIRROR=https://npmmirror.com/mirrors/node",
		)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("安装 Node.js 失败: %w", err)
		}
	}

	// 设置为默认版本
	if err := runCommand("fnm", "default", version); err != nil {
		// 如果设置默认版本失败，尝试使用 fnm use
		_ = runCommand("fnm", "use", version)
	}

	// 应用 fnm 环境变量到当前进程
	if err := applyFnmEnv(); err != nil {
		ui.Warning("应用 fnm 环境变量失败: %v", err)
	}

	return nil
}

func (n *NodeInstaller) installWithWinget(version string) error {
	if !commandExists("winget") {
		return fmt.Errorf("请先安装 winget")
	}

	return runCommand("winget", "install", "OpenJS.NodeJS")
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

	var err error
	switch runtime.GOOS {
	case "darwin":
		err = p.installWithBrew(version)
	case "linux":
		err = p.installWithUv(version)
	case "windows":
		err = p.installWithWinget(version)
	default:
		return fmt.Errorf("不支持的操作系统: %s", runtime.GOOS)
	}

	if err == nil || p.IsInstalled() {
		switch runtime.GOOS {
		case "linux":
			RecordInstallation("python", "language", version, []string{"~/.local/bin/uv", "~/.local/bin/uvx", "~/.local/share/uv"}, []string{"# added by envkit"})
		default:
			RecordInstallation("python", "language", version, nil, []string{"# added by envkit"})
		}
		return nil
	}
	return err
}

func (p *PythonInstaller) installWithBrew(version string) error {
	if !commandExists("brew") {
		return fmt.Errorf("请先安装 Homebrew: https://brew.sh")
	}

	return runCommand("brew", "install", "python@"+version)
}

func (p *PythonInstaller) installWithUv(version string) error {
	// Windows 平台不支持此安装方式
	if runtime.GOOS == "windows" {
		return fmt.Errorf("Windows 平台请使用 winget 或从官网下载安装包")
	}

	// 检查依赖
	if err := CheckAndInstallDependencies([]SystemDependency{
		CommonDependencies[0], // curl
		CommonDependencies[4], // bash
	}); err != nil {
		return err
	}

	// 检查是否安装了 uv
	if !commandExists("uv") {
		ui.Info("正在安装 uv (Python 包管理器)...")

		// 尝试多种安装方式
		installCmds := [][]string{
			{"bash", "-c", "curl -LsSf https://astral.sh/uv/install.sh | sh"},
			{"bash", "-c", "curl -LsSf https://raw.githubusercontent.com/astral-sh/uv/main/install.sh | sh"},
		}

		var installErr error
		for i, cmdParts := range installCmds {
			if i > 0 {
				ui.Info("尝试备用安装源...")
			}
			installCmd := exec.Command(cmdParts[0], cmdParts[1:]...)
			installCmd.Stdout = os.Stdout
			installCmd.Stderr = os.Stderr
			installCmd.Stdin = os.Stdin
			installErr = installCmd.Run()
			if installErr == nil {
				break
			}
		}

		if installErr != nil {
			return fmt.Errorf("安装 uv 失败: %w", installErr)
		}

		if err := locateAndAddUvToPath(); err != nil {
			return err
		}
	}

	// 使用 uv 安装 Python
	ui.Info("正在使用 uv 安装 Python %s...", version)
	cmd := exec.Command("uv", "python", "install", version)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		// 如果指定版本失败，尝试安装最新的稳定版本
		ui.Warning("安装 Python %s 失败，尝试安装 Python 3...", version)
		cmd = exec.Command("uv", "python", "install", "3")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("安装 Python 失败: %w", err)
		}
	}

	// 应用 uv python 环境变量到当前进程
	if err := applyUvPythonEnv(version); err != nil {
		ui.Warning("应用 uv python 环境变量失败: %v", err)
	}

	return nil
}

func (p *PythonInstaller) installWithWinget(version string) error {
	if !commandExists("winget") {
		return fmt.Errorf("请先安装 winget")
	}

	return runCommand("winget", "install", "Python.Python."+version)
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

	var err error
	switch runtime.GOOS {
	case "darwin":
		err = g.installWithBrew(version)
	case "linux":
		err = g.installFromSource(version)
	case "windows":
		err = g.installWithWinget(version)
	default:
		return fmt.Errorf("不支持的操作系统: %s", runtime.GOOS)
	}

	if err == nil || g.IsInstalled() {
		if err != nil {
			RecordInstallation("go", "language", version, nil, []string{"# added by envkit"})
		}
		return nil
	}
	return err
}

func (g *GoInstaller) installWithBrew(version string) error {
	if !commandExists("brew") {
		return g.installFromSourceDarwin(version)
	}

	// Homebrew uses major.minor version (e.g. go@1.22) instead of patch version (e.g. go@1.22.0)
	brewVersion := version
	parts := strings.Split(version, ".")
	if len(parts) >= 2 {
		brewVersion = parts[0] + "." + parts[1]
	}

	if err := runCommand("brew", "install", "go@"+brewVersion); err != nil {
		ui.Warning("通过 Homebrew 安装 Go 失败 (%v)，正在尝试直接从官方镜像源下载...", err)
		return g.installFromSourceDarwin(version)
	}
	locateAndAddBrewGoToPath(brewVersion)
	RecordInstallation("go", "language", version, nil, []string{"# added by envkit"})
	return nil
}

func (g *GoInstaller) installFromSourceDarwin(version string) error {
	ui.Info("正在从官方镜像下载 Go %s...", version)

	// 检查依赖
	if err := CheckAndInstallDependencies([]SystemDependency{
		CommonDependencies[0], // curl
		CommonDependencies[1], // tar
	}); err != nil {
		return err
	}

	arch := runtime.GOARCH
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	destDir := filepath.Join(home, ".local")
	_ = os.MkdirAll(destDir, 0755)

	tarPath := filepath.Join(destDir, "go.tar.gz")

	// 尝试多个下载源
	downloadURLs := []string{
		fmt.Sprintf("https://golang.google.cn/dl/go%s.darwin-%s.tar.gz", version, arch),
		fmt.Sprintf("https://go.dev/dl/go%s.darwin-%s.tar.gz", version, arch),
		fmt.Sprintf("https://golang.org/dl/go%s.darwin-%s.tar.gz", version, arch),
	}

	var downloadErr error
	for i, downloadURL := range downloadURLs {
		if i > 0 {
			ui.Info("尝试备用下载源 %d...", i)
		}
		downloadErr = runCommand("curl", "-fsSL", "--connect-timeout", "10", "--max-time", "300", "-o", tarPath, downloadURL)
		if downloadErr == nil {
			break
		}
	}

	if downloadErr != nil {
		return fmt.Errorf("下载 Go 失败: %w", downloadErr)
	}

	ui.Info("正在解压安装...")
	goDir := filepath.Join(destDir, "go")
	_ = os.RemoveAll(goDir)

	if err := runCommand("tar", "-C", destDir, "-xzf", tarPath); err != nil {
		os.Remove(tarPath)
		return fmt.Errorf("解压 Go 失败: %w", err)
	}

	os.Remove(tarPath)

	binDir := filepath.Join(goDir, "bin")
	AddDirToPath(binDir)
	_ = PersistPathEnv(binDir)

	RecordInstallation("go", "language", version, []string{"~/.local/go"}, []string{"# added by envkit"})

	ui.Success("Go %s 安装成功并已配置环境变量！", version)
	return nil
}

func (g *GoInstaller) installFromSource(version string) error {
	ui.Info("从官方镜像下载 Go %s...", version)

	// 检查依赖
	if err := CheckAndInstallDependencies([]SystemDependency{
		CommonDependencies[0], // curl
		CommonDependencies[1], // tar
	}); err != nil {
		return err
	}

	// 下载 Go 安装包
	arch := runtime.GOARCH
	tmpFile := filepath.Join(os.TempDir(), "go.tar.gz")

	// 尝试多个下载源
	downloadURLs := []string{
		fmt.Sprintf("https://golang.google.cn/dl/go%s.linux-%s.tar.gz", version, arch),
		fmt.Sprintf("https://go.dev/dl/go%s.linux-%s.tar.gz", version, arch),
		fmt.Sprintf("https://golang.org/dl/go%s.linux-%s.tar.gz", version, arch),
	}

	var downloadErr error
	for i, downloadURL := range downloadURLs {
		if i > 0 {
			ui.Info("尝试备用下载源 %d...", i)
		}
		downloadErr = runCommand("curl", "-fsSL", "--connect-timeout", "10", "--max-time", "300", "-o", tmpFile, downloadURL)
		if downloadErr == nil {
			break
		}
	}

	if downloadErr != nil {
		return fmt.Errorf("下载失败: %w", downloadErr)
	}

	// 解压安装 - 优先尝试用户目录，避免需要 sudo
	ui.Info("正在安装...")
	home, _ := os.UserHomeDir()
	userInstallDir := filepath.Join(home, ".local")
	goDir := filepath.Join(userInstallDir, "go")

	// 先尝试安装到用户目录（Windows 上忽略权限参数）
	_ = os.MkdirAll(userInstallDir, 0755)
	_ = os.RemoveAll(goDir) // 清理旧版本

	if err := runCommand("tar", "-C", userInstallDir, "-xzf", tmpFile); err == nil {
		// 用户目录安装成功
		os.Remove(tmpFile)
		binDir := filepath.Join(goDir, "bin")
		AddDirToPath(binDir)
		_ = PersistPathEnv(binDir)
		RecordInstallation("go", "language", version, []string{"~/.local/go"}, []string{"# added by envkit"})
		ui.Success("Go %s 安装成功并已配置环境变量！", version)
		return nil
	}

	// 用户目录失败，尝试系统目录（需要 sudo）
	ui.Info("尝试安装到系统目录（需要管理员权限）...")
	if err := runCommand("sudo", "rm", "-rf", "/usr/local/go"); err != nil {
		ui.Warning("清理旧版本失败: %v", err)
	}

	if err := runCommand("sudo", "tar", "-C", "/usr/local", "-xzf", tmpFile); err != nil {
		os.Remove(tmpFile)
		return fmt.Errorf("安装失败: %w", err)
	}

	// 清理临时文件
	os.Remove(tmpFile)

	AddDirToPath("/usr/local/go/bin")
	_ = PersistPathEnv("/usr/local/go/bin")

	RecordInstallation("go", "language", version, []string{"/usr/local/go"}, []string{"# added by envkit"})

	ui.Success("Go %s 安装成功并已配置环境变量！", version)
	return nil
}

func (g *GoInstaller) installWithWinget(version string) error {
	if !commandExists("winget") {
		return fmt.Errorf("请先安装 winget")
	}

	return runCommand("winget", "install", "GoLang.Go")
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
	var err error
	if !commandExists("rustup") {
		spinner.Stop() // 安装 rustup 可能需要交互
		ui.Info("正在安装 rustup...")
		var cmd *exec.Cmd
		if runtime.GOOS == "windows" {
			cmd = exec.Command("winget", "install", "--id", "Rustlang.Rustup", "--silent", "--accept-package-agreements", "--accept-source-agreements")
		} else {
			// 尝试多个安装源
			installURLs := []string{
				"https://rsproxy.cn/rustup-init.sh",
				"https://sh.rustup.rs",
			}

			installed := false
			for i, url := range installURLs {
				if i > 0 {
					ui.Info("尝试备用安装源...")
				}
				cmd = exec.Command("sh", "-c", fmt.Sprintf("curl --proto '=https' --tlsv1.2 -sSf %s | sh -s -- -y", url))
				cmd.Env = append(os.Environ(),
					"RUSTUP_DIST_SERVER=https://rsproxy.cn",
					"RUSTUP_UPDATE_ROOT=https://rsproxy.cn/rustup",
				)
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				cmd.Stdin = os.Stdin
				err = cmd.Run()
				if err == nil {
					installed = true
					break
				}
			}

			if !installed && !r.IsInstalled() {
				return fmt.Errorf("安装 rustup 失败: %w", err)
			}
		}
		spinner.Start()

		if err == nil {
			if err = locateAndAddRustupToPath(); err != nil {
				if !r.IsInstalled() {
					return err
				}
			}
		}
	}

	if commandExists("rustup") {
		// 安装指定版本
		spinner.Stop()
		ui.Info("正在使用 rustup 安装 Rust %s...", version)
		cmd := exec.Command("rustup", "install", version)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		cmd.Env = append(os.Environ(),
			"RUSTUP_DIST_SERVER=https://rsproxy.cn",
			"RUSTUP_UPDATE_ROOT=https://rsproxy.cn/rustup",
		)
		err = cmd.Run()

		// 如果指定版本失败，尝试安装 stable
		if err != nil && version != "stable" {
			ui.Warning("安装 Rust %s 失败，尝试安装 stable 版本...", version)
			cmd = exec.Command("rustup", "install", "stable")
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Stdin = os.Stdin
			cmd.Env = append(os.Environ(),
				"RUSTUP_DIST_SERVER=https://rsproxy.cn",
				"RUSTUP_UPDATE_ROOT=https://rsproxy.cn/rustup",
			)
			err = cmd.Run()
		}
		spinner.Start()
	}

	if err == nil || r.IsInstalled() {
		paths := []string{"~/.cargo", "~/.rustup"}
		RecordInstallation("rust", "language", version, paths, []string{"# added by envkit"})
		return nil
	}
	return err
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
	case "java", "jdk":
		return &JavaInstaller{}
	case "bun":
		return &BunInstaller{}
	default:
		return nil
	}
}

// JavaInstaller Java (JDK) 安装器
type JavaInstaller struct{}

func (j *JavaInstaller) Install(version string) error {
	spinner := ui.NewSpinner("正在安装 Java (JDK) " + version)
	spinner.Start()
	defer spinner.Stop()

	var err error
	switch runtime.GOOS {
	case "darwin", "linux":
		// 需要交互式 sudo，先停止 spinner
		spinner.Stop()
		err = j.installWithSdkman(version)
		spinner.Start()
	case "windows":
		err = exec.Command("winget", "install", "Microsoft.OpenJDK."+version).Run()
	default:
		return fmt.Errorf("不支持的操作系统: %s", runtime.GOOS)
	}

	if err == nil || j.IsInstalled() {
		switch runtime.GOOS {
		case "darwin", "linux":
			RecordInstallation("java", "language", version, []string{"~/.sdkman"}, []string{"SDKMAN_DIR", "sdkman-init.sh"})
		default:
			RecordInstallation("java", "language", version, nil, nil)
		}
		return nil
	}
	return err
}

func (j *JavaInstaller) installWithSdkman(version string) error {
	// 1. 检查并安装 SDKMAN!
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	sdkmanDir := filepath.Join(home, ".sdkman")
	sdkmanInitScript := filepath.Join(sdkmanDir, "bin", "sdkman-init.sh")

	// 检查 SDKMAN 是否已正确安装
	if _, err := os.Stat(sdkmanInitScript); os.IsNotExist(err) {
		// 先清理可能存在的不完整安装
		if _, err := os.Stat(sdkmanDir); err == nil {
			ui.Info("检测到不完整的 SDKMAN 安装，正在清理...")
			_ = os.RemoveAll(sdkmanDir)
		}

		// 检查 zip/unzip 依赖
		if !commandExists("zip") || !commandExists("unzip") {
			ui.Info("检测到系统缺少 zip 或 unzip 依赖，正在尝试安装...")
			if commandExists("apt-get") {
				_ = runCommand("sudo", "apt-get", "update")
				if err := runCommand("sudo", "apt-get", "install", "-y", "zip", "unzip"); err != nil {
					ui.Warning("无法自动安装 zip/unzip 依赖，SDKMAN! 安装可能会失败。请手动安装：sudo apt-get install -y zip unzip")
				}
			} else if commandExists("yum") {
				if err := runCommand("sudo", "yum", "install", "-y", "zip", "unzip"); err != nil {
					ui.Warning("无法自动安装 zip/unzip 依赖，SDKMAN! 安装可能会失败。请手动安装：sudo yum install -y zip unzip")
				}
			} else if commandExists("dnf") {
				if err := runCommand("sudo", "dnf", "install", "-y", "zip", "unzip"); err != nil {
					ui.Warning("无法自动安装 zip/unzip 依赖，SDKMAN! 安装可能会失败。请手动安装：sudo dnf install -y zip unzip")
				}
			} else if commandExists("pacman") {
				if err := runCommand("sudo", "pacman", "-S", "--noconfirm", "zip", "unzip"); err != nil {
					ui.Warning("无法自动安装 zip/unzip 依赖，SDKMAN! 安装可能会失败。请手动安装：sudo pacman -S zip unzip")
				}
			} else if commandExists("brew") {
				if err := runCommand("brew", "install", "zip", "unzip"); err != nil {
					ui.Warning("无法自动安装 zip/unzip 依赖，SDKMAN! 安装可能会失败。请手动安装：brew install zip unzip")
				}
			} else {
				ui.Warning("无法确定当前系统的包管理器，请手动安装 zip 和 unzip 后重试。")
			}
		}

		ui.Info("正在安装 SDKMAN!...")
		// 尝试多种方式安装 SDKMAN
		var installErr error

		// 方式 1: 直接从官方安装
		installCmd := exec.Command("bash", "-c", "curl -s \"https://get.sdkman.io\" | bash")
		installCmd.Env = append(os.Environ(), "SDKMAN_DIR="+sdkmanDir)
		installCmd.Stdout = os.Stdout
		installCmd.Stderr = os.Stderr
		installCmd.Stdin = os.Stdin
		installErr = installCmd.Run()

		// 验证安装是否成功
		if installErr == nil {
			if _, err := os.Stat(sdkmanInitScript); os.IsNotExist(err) {
				installErr = fmt.Errorf("SDKMAN 安装脚本未找到")
			}
		}

		if installErr != nil {
			ui.Warning("官方源安装失败 (%v)，尝试手动下载安装...", installErr)

			// 清理失败的安装
			_ = os.RemoveAll(sdkmanDir)

			// 方式 2: 手动下载并解压 SDKMAN
			downloadURLs := []string{
				"https://get.sdkman.io?rcupdate=false",
				"https://raw.githubusercontent.com/sdkman/sdkman-cli/master/res/install.sh",
			}

			downloaded := false
			for i, url := range downloadURLs {
				if i > 0 {
					ui.Info("尝试备用下载源...")
				}

				// 下载安装脚本
				tmpScript := filepath.Join(os.TempDir(), "sdkman-install.sh")
				if err := runCommand("curl", "-fsSL", "--connect-timeout", "10", "-o", tmpScript, url); err == nil {
					// 执行安装脚本
					installCmd = exec.Command("bash", tmpScript)
					installCmd.Env = append(os.Environ(), "SDKMAN_DIR="+sdkmanDir)
					installCmd.Stdout = os.Stdout
					installCmd.Stderr = os.Stderr
					installCmd.Stdin = os.Stdin
					if installCmd.Run() == nil {
						// 验证安装
						if _, err := os.Stat(sdkmanInitScript); err == nil {
							downloaded = true
							os.Remove(tmpScript)
							break
						}
					}
					os.Remove(tmpScript)
				}
			}

			if !downloaded {
				return fmt.Errorf("所有 SDKMAN 安装方式均失败，请手动安装 SDKMAN: curl -s \"https://get.sdkman.io\" | bash")
			}
		}

		// 配置 SDKMAN 使用国内镜像（尝试性配置，失败不影响）
		configFile := filepath.Join(sdkmanDir, "etc", "config")
		if content, err := os.ReadFile(configFile); err == nil {
			configStr := string(content)
			// 替换为国内镜像源
			configStr = strings.ReplaceAll(configStr, "sdkman_api_url=https://api.sdkman.io/2", "sdkman_api_url=https://sdkman.bmpi.dev/2")
			configStr = strings.ReplaceAll(configStr, "sdkman_candidatesapi_url=https://api.sdkman.io/2/candidates/all", "sdkman_candidatesapi_url=https://sdkman.bmpi.dev/2/candidates/all")
			_ = os.WriteFile(configFile, []byte(configStr), 0644)
		}
	}

	// 2. 映射版本名称。SDKMAN 上的版本名如 "21-open" (OpenJDK 21)
	sdkVersion := version
	if version == "21" || version == "17" || version == "11" || version == "8" {
		sdkVersion = version + "-open"
	}

	ui.Info("正在使用 SDKMAN! 安装 Java %s...", sdkVersion)

	// 尝试安装 Java，如果失败则尝试不同的版本标识符
	installScript := fmt.Sprintf("source %s && sdk install java %s", sdkmanInitScript, sdkVersion)
	cmd := exec.Command("bash", "-c", installScript)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Env = append(os.Environ(), "SDKMAN_DIR="+sdkmanDir)

	err = cmd.Run()
	if err != nil {
		// 如果安装失败，尝试其他版本标识符
		alternativeVersions := []string{
			sdkVersion,
			version + ".0.0-open",
			version + "-tem",
			version + "-zulu",
		}

		for _, altVer := range alternativeVersions {
			if altVer == sdkVersion {
				continue // 跳过已经尝试过的
			}
			ui.Info("尝试安装 Java %s...", altVer)
			installScript = fmt.Sprintf("source %s && sdk install java %s", sdkmanInitScript, altVer)
			cmd = exec.Command("bash", "-c", installScript)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Stdin = os.Stdin
			cmd.Env = append(os.Environ(), "SDKMAN_DIR="+sdkmanDir)

			if cmd.Run() == nil {
				err = nil
				break
			}
		}

		if err != nil {
			return fmt.Errorf("SDKMAN 安装 Java 失败: %w", err)
		}
	}

	// 3. 配置环境变量
	return persistSdkmanIntegration(home)
}

func persistSdkmanIntegration(home string) error {
	files := []string{
		filepath.Join(home, ".bashrc"),
		filepath.Join(home, ".zshrc"),
		filepath.Join(home, ".profile"),
	}

	integrationCmd := "\n# added by envkit (sdkman)\nexport SDKMAN_DIR=\"$HOME/.sdkman\"\n[[ -s \"$HOME/.sdkman/bin/sdkman-init.sh\" ]] && source \"$HOME/.sdkman/bin/sdkman-init.sh\"\n"

	for _, file := range files {
		if _, err := os.Stat(file); err == nil {
			content, err := os.ReadFile(file)
			if err == nil {
				if !strings.Contains(string(content), "sdkman-init.sh") {
					f, err := os.OpenFile(file, os.O_APPEND|os.O_WRONLY, 0644)
					if err == nil {
						_, _ = f.WriteString(integrationCmd)
						_ = f.Close()
					}
				}
			}
		}
	}
	return nil
}

func (j *JavaInstaller) IsInstalled() bool {
	return commandExists("java")
}

func (j *JavaInstaller) GetVersion() string {
	cmd := exec.Command("java", "-version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return ""
	}
	lines := strings.Split(string(output), "\n")
	if len(lines) > 0 {
		return strings.TrimSpace(lines[0])
	}
	return string(output)
}

// BunInstaller Bun 安装器
type BunInstaller struct{}

func (b *BunInstaller) Install(version string) error {
	spinner := ui.NewSpinner("正在安装 Bun " + version)
	spinner.Start()
	defer spinner.Stop()

	var err error
	switch runtime.GOOS {
	case "darwin", "linux":
		err = b.installWithCurl()
	case "windows":
		err = exec.Command("winget", "install", "Jarred-Sumner.Bun").Run()
	default:
		return fmt.Errorf("不支持的操作系统: %s", runtime.GOOS)
	}

	if err == nil || b.IsInstalled() {
		switch runtime.GOOS {
		case "darwin", "linux":
			RecordInstallation("bun", "language", version, []string{"~/.bun"}, []string{"# bun", "BUN_INSTALL"})
		default:
			RecordInstallation("bun", "language", version, nil, nil)
		}
		return nil
	}
	return err
}

func (b *BunInstaller) installWithCurl() error {
	// 尝试多个安装源
	installURLs := []string{
		"https://bun.sh/install",
		"https://github.com/oven-sh/bun/releases/latest/download/install.sh",
	}

	var installErr error
	for i, url := range installURLs {
		if i > 0 {
			ui.Info("尝试备用安装源...")
		}
		cmd := exec.Command("bash", "-c", fmt.Sprintf("curl -fsSL %s | bash", url))
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		installErr = cmd.Run()
		if installErr == nil {
			break
		}
	}

	if installErr != nil {
		return fmt.Errorf("安装 Bun 失败: %w", installErr)
	}

	// 将 ~/.bun/bin 加入 path 并持久化
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	bunBinDir := filepath.Join(home, ".bun", "bin")
	AddDirToPath(bunBinDir)
	return PersistPathEnv(bunBinDir)
}

func (b *BunInstaller) IsInstalled() bool {
	return commandExists("bun")
}

func (b *BunInstaller) GetVersion() string {
	cmd := exec.Command("bun", "--version")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}
