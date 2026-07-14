package installer

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/fusheng/envkit/internal/detector"
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

	if err != nil {
		return err
	}
	switch runtime.GOOS {
	case "linux":
		RecordInstallation("node", "language", version, []string{"~/.fnm", "~/.local/share/fnm"}, []string{"# fnm integration", "# added by envkit"})
	default:
		RecordInstallation("node", "language", version, nil, []string{"# added by envkit"})
	}
	return nil
}

func (n *NodeInstaller) installWithBrew(version string) error {
	// 检查是否安装了 Homebrew
	if !commandExists("brew") {
		return fmt.Errorf("请先安装 Homebrew: https://brew.sh")
	}

	return runCommand("brew", "install", "node@"+version)
}

func (n *NodeInstaller) installWithFnm(version string) error {
	// 检查是否安装了 fnm
	if !commandExists("fnm") {
		ui.Info("正在安装 fnm (Fast Node Manager)...")
		if err := runCommand("sh", "-c", "curl -fsSL https://fnm.vercel.app/install | bash"); err != nil {
			return fmt.Errorf("安装 fnm 失败: %w", err)
		}
		
		if err := locateAndAddFnmToPath(); err != nil {
			return err
		}
	}

	// 使用 fnm 安装 Node.js
	if err := runCommand("fnm", "install", version); err != nil {
		return fmt.Errorf("安装 Node.js 失败: %w", err)
	}

	// 设置为默认版本
	if err := runCommand("fnm", "default", version); err != nil {
		return fmt.Errorf("设置默认 Node.js 版本失败: %w", err)
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

	if err != nil {
		return err
	}
	switch runtime.GOOS {
	case "linux":
		RecordInstallation("python", "language", version, []string{"~/.local/bin/uv", "~/.local/bin/uvx", "~/.local/share/uv"}, []string{"# added by envkit"})
	default:
		RecordInstallation("python", "language", version, nil, []string{"# added by envkit"})
	}
	return nil
}

func (p *PythonInstaller) installWithBrew(version string) error {
	if !commandExists("brew") {
		return fmt.Errorf("请先安装 Homebrew: https://brew.sh")
	}

	return runCommand("brew", "install", "python@"+version)
}

func (p *PythonInstaller) installWithUv(version string) error {
	// 检查是否安装了 uv
	if !commandExists("uv") {
		ui.Info("正在安装 uv (Python 包管理器)...")
		if err := runCommand("sh", "-c", "curl -LsSf https://astral.sh/uv/install.sh | sh"); err != nil {
			return fmt.Errorf("安装 uv 失败: %w", err)
		}
		
		if err := locateAndAddUvToPath(); err != nil {
			return err
		}
	}

	// 使用 uv 安装 Python
	if err := runCommand("uv", "python", "install", version); err != nil {
		return fmt.Errorf("安装 Python 失败: %w", err)
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
		// 用户级安装到 ~/.local/go，避免 sudo /usr/local（TUI/无密码场景会假失败或假成功）
		err = g.installUserLocal(version, "linux")
	case "windows":
		err = g.installWithWinget(version)
	default:
		return fmt.Errorf("不支持的操作系统: %s", runtime.GOOS)
	}

	if err != nil {
		return err
	}
	if !g.IsInstalled() {
		return fmt.Errorf("Go 安装流程结束，但未检测到 go 命令（请检查 ~/.local/go/bin 是否在 PATH）")
	}
	return nil
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
	return g.installUserLocal(version, "darwin")
}

// installUserLocal 将 Go 安装到 ~/.local/go（无需 root）
func (g *GoInstaller) installUserLocal(version, goos string) error {
	ui.Info("正在下载 Go %s (%s/%s)...", version, goos, runtime.GOARCH)

	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	destDir := filepath.Join(home, ".local")
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return err
	}

	tarPath := filepath.Join(destDir, "go.tar.gz")
	urls := goDownloadURLs(version, goos, runtime.GOARCH)

	var downloadErr error
	for _, downloadURL := range urls {
		ui.Info("尝试: %s", downloadURL)
		if err := runCommand("curl", "-fsSL", "--connect-timeout", "30", "--max-time", "600", "-o", tarPath, downloadURL); err != nil {
			downloadErr = err
			continue
		}
		downloadErr = nil
		break
	}
	if downloadErr != nil {
		return fmt.Errorf("下载 Go 失败: %w", downloadErr)
	}

	// 校验 tarball 大小，避免下到 HTML 错误页仍继续解压
	if st, err := os.Stat(tarPath); err != nil || st.Size() < 1_000_000 {
		_ = os.Remove(tarPath)
		return fmt.Errorf("下载的 Go 包无效或过小")
	}

	ui.Info("正在解压到 ~/.local/go ...")
	goDir := filepath.Join(destDir, "go")
	_ = os.RemoveAll(goDir)

	if err := runCommand("tar", "-C", destDir, "-xzf", tarPath); err != nil {
		_ = os.Remove(tarPath)
		return fmt.Errorf("解压 Go 失败: %w", err)
	}
	_ = os.Remove(tarPath)

	binDir := filepath.Join(goDir, "bin")
	goBin := filepath.Join(binDir, "go")
	if runtime.GOOS == "windows" {
		goBin += ".exe"
	}
	if _, err := os.Stat(goBin); err != nil {
		return fmt.Errorf("解压后未找到 %s", goBin)
	}

	AddDirToPath(binDir)
	if err := PersistPathEnv(binDir); err != nil {
		ui.Warning("写入 shell PATH 失败: %v（当前进程已临时生效）", err)
	}

	RecordInstallation("go", "language", version, []string{"~/.local/go"}, []string{"# added by envkit"})
	ui.Success("Go %s 已安装到 ~/.local/go（新终端自动生效 PATH）", version)
	return nil
}

func goDownloadURLs(version, goos, arch string) []string {
	name := fmt.Sprintf("go%s.%s-%s.tar.gz", version, goos, arch)
	return []string{
		"https://golang.google.cn/dl/" + name,
		"https://mirrors.aliyun.com/golang/" + name,
		"https://go.dev/dl/" + name,
	}
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
	return strings.TrimSpace(string(output))
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
		ui.Info("正在安装 rustup...")
		var cmd *exec.Cmd
		if runtime.GOOS == "windows" {
			cmd = exec.Command("winget", "install", "--id", "Rustlang.Rustup", "--silent", "--accept-package-agreements", "--accept-source-agreements")
		} else {
			cmd = exec.Command("sh", "-c", "curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y")
		}
		cmd.Env = append(os.Environ(),
			"RUSTUP_DIST_SERVER=https://rsproxy.cn",
			"RUSTUP_UPDATE_ROOT=https://rsproxy.cn/rustup",
		)
		err = cmd.Run()
		if err != nil {
			if !r.IsInstalled() {
				return fmt.Errorf("安装 rustup 失败: %w", err)
			}
		}

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
		cmd := exec.Command("rustup", "install", version)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Env = append(os.Environ(),
			"RUSTUP_DIST_SERVER=https://rsproxy.cn",
			"RUSTUP_UPDATE_ROOT=https://rsproxy.cn/rustup",
		)
		err = cmd.Run()
	}

	if err != nil {
		return err
	}
	paths := []string{"~/.cargo", "~/.rustup"}
	RecordInstallation("rust", "language", version, paths, []string{"# added by envkit"})
	return nil
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

// commandExists 检查命令是否存在（复用 detector 的 PATH 回退逻辑）
func commandExists(cmd string) bool {
	return detector.CheckCommand(cmd)
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
		err = j.installWithSdkman(version)
	case "windows":
		err = exec.Command("winget", "install", "Microsoft.OpenJDK."+version).Run()
	default:
		return fmt.Errorf("不支持的操作系统: %s", runtime.GOOS)
	}

	if err != nil {
		return err
	}
	switch runtime.GOOS {
	case "darwin", "linux":
		RecordInstallation("java", "language", version, []string{"~/.sdkman"}, []string{"SDKMAN_DIR", "sdkman-init.sh"})
	default:
		RecordInstallation("java", "language", version, nil, nil)
	}
	return nil
}

func (j *JavaInstaller) installWithSdkman(version string) error {
	// 1. 检查并安装 SDKMAN!
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	sdkmanDir := filepath.Join(home, ".sdkman")
	if _, err := os.Stat(sdkmanDir); os.IsNotExist(err) {
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
			} else if commandExists("brew") {
				if err := runCommand("brew", "install", "zip", "unzip"); err != nil {
					ui.Warning("无法自动安装 zip/unzip 依赖，SDKMAN! 安装可能会失败。请手动安装：brew install zip unzip")
				}
			} else {
				ui.Warning("无法确定当前系统的包管理器，请手动安装 zip 和 unzip 后重试。")
			}
		}

		ui.Info("正在安装 SDKMAN!...")
		if err := runCommand("sh", "-c", "curl -s \"https://get.sdkman.io\" | bash"); err != nil {
			return fmt.Errorf("安装 SDKMAN! 失败: %w", err)
		}
	}

	// 2. 映射版本名称。SDKMAN 上的版本名如 "21-open" (OpenJDK 21)
	sdkVersion := version
	if version == "21" || version == "17" || version == "11" || version == "8" {
		sdkVersion = version + "-open"
	}

	ui.Info("正在使用 SDKMAN! 安装 Java %s...", sdkVersion)
	if err := runCommand("bash", "-c", fmt.Sprintf("source %s/.sdkman/bin/sdkman-init.sh && sdk install java %s", home, sdkVersion)); err != nil {
		return fmt.Errorf("SDKMAN 安装 Java 失败: %w", err)
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

	if err != nil {
		return err
	}
	switch runtime.GOOS {
	case "darwin", "linux":
		RecordInstallation("bun", "language", version, []string{"~/.bun"}, []string{"# bun", "BUN_INSTALL"})
	default:
		RecordInstallation("bun", "language", version, nil, nil)
	}
	if !b.IsInstalled() {
		return fmt.Errorf("Bun 安装流程结束，但未检测到 bun 命令（请检查 ~/.bun/bin 是否在 PATH，或重新打开终端）")
	}
	return nil
}

func (b *BunInstaller) installWithCurl() error {
	// 确保 unzip 依赖存在
	if !commandExists("unzip") {
		ui.Info("检测到系统缺少 unzip 依赖，正在尝试安装...")
		if commandExists("apt-get") {
			_ = runCommand("sudo", "apt-get", "update")
			if err := runCommand("sudo", "apt-get", "install", "-y", "unzip"); err != nil {
				ui.Warning("无法自动安装 unzip 依赖，Bun 安装可能会失败。请手动安装：sudo apt-get install -y unzip")
			}
		} else if commandExists("yum") {
			if err := runCommand("sudo", "yum", "install", "-y", "unzip"); err != nil {
				ui.Warning("无法自动安装 unzip 依赖，Bun 安装可能会失败。请手动安装：sudo yum install -y unzip")
			}
		} else if commandExists("dnf") {
			if err := runCommand("sudo", "dnf", "install", "-y", "unzip"); err != nil {
				ui.Warning("无法自动安装 unzip 依赖，Bun 安装可能会失败。请手动安装：sudo dnf install -y unzip")
			}
		} else if commandExists("pacman") {
			if err := runCommand("sudo", "pacman", "-S", "--noconfirm", "unzip"); err != nil {
				ui.Warning("无法自动安装 unzip 依赖，Bun 安装可能会失败。请手动安装：sudo pacman -S unzip")
			}
		} else if commandExists("brew") {
			if err := runCommand("brew", "install", "unzip"); err != nil {
				ui.Warning("无法自动安装 unzip 依赖，Bun 安装可能会失败。请手动安装：brew install unzip")
			}
		} else {
			ui.Warning("无法确定当前系统的包管理器，请手动安装 unzip 后重试。")
		}
	}

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
	bunBin := filepath.Join(bunBinDir, "bun")
	if _, err := os.Stat(bunBin); err != nil {
		return fmt.Errorf("Bun 安装脚本结束，但未找到 %s", bunBin)
	}
	AddDirToPath(bunBinDir)
	if err := PersistPathEnv(bunBinDir); err != nil {
		ui.Warning("写入 shell PATH 失败: %v（当前进程已临时生效）", err)
	}
	return nil
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
