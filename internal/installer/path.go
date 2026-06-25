package installer

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// AddDirToPath 将目录添加到当前进程的 PATH 中
func AddDirToPath(dir string) {
	path := os.Getenv("PATH")
	dirs := filepath.SplitList(path)
	for _, d := range dirs {
		if d == dir {
			return
		}
	}
	os.Setenv("PATH", dir+string(os.PathListSeparator)+path)
}

// PersistPathEnv 将环境变量修改写入 shell 配置文件（Unix）或系统环境变量（Windows）
func PersistPathEnv(dir string) error {
	if runtime.GOOS == "windows" {
		return persistPathEnvWindows(dir)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	// 尝试用 $HOME 代替绝对路径，使配置更简洁
	pathValue := dir
	if strings.HasPrefix(dir, home) {
		pathValue = "$HOME" + strings.TrimPrefix(dir, home)
	}

	exportCmd := fmt.Sprintf("\n# added by envkit\nexport PATH=\"%s:$PATH\"\n", pathValue)

	// 获取 shell 配置文件
	files := getShellConfigFiles()

	for _, file := range files {
		if _, err := os.Stat(file); err == nil {
			content, err := os.ReadFile(file)
			if err == nil {
				// 避免重复写入
				if !strings.Contains(string(content), dir) && !strings.Contains(string(content), pathValue) {
					f, err := os.OpenFile(file, os.O_APPEND|os.O_WRONLY, 0644)
					if err == nil {
						_, _ = f.WriteString(exportCmd)
						_ = f.Close()
					}
				}
			}
		}
	}
	return nil
}

// getDefaultShell 获取当前用户的默认 shell
func getDefaultShell() string {
	shell := os.Getenv("SHELL")
	if shell != "" {
		return filepath.Base(shell)
	}

	// 降级到 dscl 查询（macOS）
	if runtime.GOOS == "darwin" {
		cmd := exec.Command("dscl", ".", "-read", os.Getenv("HOME"), "UserShell")
		out, err := cmd.Output()
		if err == nil {
			parts := strings.Fields(string(out))
			if len(parts) >= 2 {
				return filepath.Base(parts[1])
			}
		}
	}

	return "bash"
}

// getShellConfigFiles 获取 shell 配置文件列表
func getShellConfigFiles() []string {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil
	}

	if runtime.GOOS == "windows" {
		return nil
	}

	defaultShell := getDefaultShell()

	// 优先配置默认 shell 的配置文件
	var files []string

	if defaultShell == "zsh" {
		files = []string{
			filepath.Join(home, ".zshrc"),
			filepath.Join(home, ".zprofile"),
		}
	} else if defaultShell == "bash" {
		files = []string{
			filepath.Join(home, ".bashrc"),
			filepath.Join(home, ".bash_profile"),
			filepath.Join(home, ".profile"),
		}
	} else {
		// 降级方案：尝试所有
		files = []string{
			filepath.Join(home, ".bashrc"),
			filepath.Join(home, ".zshrc"),
			filepath.Join(home, ".profile"),
		}
	}

	return files
}

// locateAndAddFnmToPath 定位并添加 fnm 到 PATH
func locateAndAddFnmToPath() error {
	if commandExists("fnm") {
		return nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	dirs := []string{
		filepath.Join(home, ".fnm"),
		filepath.Join(home, ".local", "share", "fnm"),
		filepath.Join(home, "Library", "Application Support", "fnm"),
	}

	for _, dir := range dirs {
		fnmPath := filepath.Join(dir, "fnm")
		if _, err := os.Stat(fnmPath); err == nil {
			AddDirToPath(dir)
			_ = PersistPathEnv(dir)
			persistFnmShellIntegration(dir)
			return nil
		}
	}

	return fmt.Errorf("未找到 fnm 二进制文件，请确认安装成功并已加入 PATH")
}

// persistFnmShellIntegration 配置 fnm shell 集成
func persistFnmShellIntegration(fnmDir string) {
	home, err := os.UserHomeDir()
	if err != nil {
		return
	}

	pathValue := fnmDir
	if strings.HasPrefix(fnmDir, home) {
		pathValue = "$HOME" + strings.TrimPrefix(fnmDir, home)
	}

	integrationCmd := fmt.Sprintf("\n# fnm integration\nexport PATH=\"%s:$PATH\"\neval \"$(fnm env --use-on-cd)\"\n", pathValue)

	files := getShellConfigFiles()

	for _, file := range files {
		if _, err := os.Stat(file); err == nil {
			content, err := os.ReadFile(file)
			if err == nil {
				if !strings.Contains(string(content), "fnm env") {
					f, err := os.OpenFile(file, os.O_APPEND|os.O_WRONLY, 0644)
					if err == nil {
						_, _ = f.WriteString(integrationCmd)
						_ = f.Close()
					}
				}
			}
		}
	}
}

// applyFnmEnv 获取并应用 fnm 环境变量到当前进程
func applyFnmEnv() error {
	cmd := exec.Command("fnm", "env", "--shell", "bash")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("运行 fnm env 失败 (%v): %s", err, string(out))
	}

	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "export ") {
			continue
		}
		parts := strings.SplitN(strings.TrimPrefix(line, "export "), "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := parts[0]
		val := strings.Trim(parts[1], `"'`)

		if key == "PATH" {
			val = strings.Replace(val, "$PATH", os.Getenv("PATH"), -1)
			os.Setenv("PATH", val)
		} else {
			os.Setenv(key, val)
		}
	}
	return nil
}

// locateAndAddUvToPath 定位并添加 uv 到 PATH
func locateAndAddUvToPath() error {
	if commandExists("uv") {
		return nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	dirs := []string{
		filepath.Join(home, ".local", "bin"),
		filepath.Join(home, ".cargo", "bin"),
	}

	for _, dir := range dirs {
		uvName := "uv"
		if runtime.GOOS == "windows" {
			uvName = "uv.exe"
		}
		uvPath := filepath.Join(dir, uvName)
		if _, err := os.Stat(uvPath); err == nil {
			AddDirToPath(dir)
			_ = PersistPathEnv(dir)
			return nil
		}
	}

	return fmt.Errorf("未找到 uv 二进制文件，请确认安装成功并已加入 PATH")
}

// applyUvPythonEnv 定位并应用 uv 安装的 Python 到 PATH
func applyUvPythonEnv(version string) error {
	cmd := exec.Command("uv", "python", "find", version)
	out, err := cmd.Output()
	if err != nil {
		cmd = exec.Command("uv", "python", "find")
		out, err = cmd.Output()
		if err != nil {
			return fmt.Errorf("运行 uv python find 失败: %w", err)
		}
	}

	pythonPath := strings.TrimSpace(string(out))
	if pythonPath == "" {
		return fmt.Errorf("未找到 Python 路径")
	}

	binDir := filepath.Dir(pythonPath)
	AddDirToPath(binDir)
	_ = PersistPathEnv(binDir)
	return nil
}

// locateAndAddRustupToPath 定位并添加 rustup (cargo) 到 PATH
func locateAndAddRustupToPath() error {
	if commandExists("rustup") {
		return nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	dir := filepath.Join(home, ".cargo", "bin")
	rustupPath := filepath.Join(dir, "rustup")
	if _, err := os.Stat(rustupPath); err == nil {
		AddDirToPath(dir)
		_ = PersistPathEnv(dir)
		return nil
	}

	return fmt.Errorf("未找到 rustup 二进制文件，请确认安装成功并已加入 PATH")
}

// locateAndAddBrewGoToPath 定位并通过 Homebrew 添加 Go 到 PATH
func locateAndAddBrewGoToPath(version string) {
	cmd := exec.Command("brew", "--prefix", "go@"+version)
	out, err := cmd.Output()
	if err != nil {
		cmd = exec.Command("brew", "--prefix", "go")
		out, err = cmd.Output()
	}

	if err == nil {
		prefix := strings.TrimSpace(string(out))
		if prefix != "" {
			binDir := filepath.Join(prefix, "bin")
			AddDirToPath(binDir)
			_ = PersistPathEnv(binDir)
		}
	}
}

// persistFnmMirrorConfig 将fnm国内镜像配置持久化到shell配置文件
func persistFnmMirrorConfig() {
	mirrorConfig := "\n# fnm 国内镜像配置 (added by envkit)\nexport FNM_NODE_DIST_MIRROR=\"https://registry.npmmirror.com/-/binary/node\"\n"

	files := getShellConfigFiles()
	for _, file := range files {
		if _, err := os.Stat(file); err == nil {
			content, err := os.ReadFile(file)
			if err == nil {
				// 避免重复写入
				if !strings.Contains(string(content), "FNM_NODE_DIST_MIRROR") {
					f, err := os.OpenFile(file, os.O_APPEND|os.O_WRONLY, 0644)
					if err == nil {
						_, _ = f.WriteString(mirrorConfig)
						_ = f.Close()
					}
				}
			}
		}
	}
}
