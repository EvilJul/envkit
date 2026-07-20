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

// refreshProcessPathFromRegistry 从 Windows 用户与系统 PATH 合并刷新当前进程 PATH。
// winget 安装后新路径往往只写入注册表，当前进程不会自动可见。
func refreshProcessPathFromRegistry() {
	if runtime.GOOS != "windows" {
		return
	}
	// PowerShell: 合并 Machine + User Path
	ps := `[Environment]::GetEnvironmentVariable('Path','Machine') + ';' + [Environment]::GetEnvironmentVariable('Path','User')`
	cmd := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", ps)
	out, err := cmd.Output()
	if err != nil {
		return
	}
	newPath := strings.TrimSpace(string(out))
	if newPath == "" {
		return
	}
	os.Setenv("PATH", newPath)
}

// PersistPathEnv 将环境变量修改写入 shell 配置文件（Unix）或系统环境变量（Windows）
func PersistPathEnv(dir string) error {
	return PersistPathEnvTagged(dir, "envkit")
}

// PersistPathEnvTagged 写入 PATH，并使用可识别标签（便于按组件卸载）
func PersistPathEnvTagged(dir, tag string) error {
	if runtime.GOOS == "windows" {
		return persistPathEnvWindows(dir)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	pathValue := dir
	if strings.HasPrefix(dir, home) {
		pathValue = "$HOME" + strings.TrimPrefix(dir, home)
	}
	if tag == "" {
		tag = "envkit"
	}
	exportCmd := fmt.Sprintf("\n# envkit:%s\nexport PATH=\"%s:$PATH\"\n", tag, pathValue)

	files := getShellConfigFiles()
	if len(files) == 0 {
		return fmt.Errorf("未找到可写入的 shell 配置文件")
	}

	// 确保至少一个配置文件存在（无则创建首选）
	primary := files[0]
	if _, err := os.Stat(primary); os.IsNotExist(err) {
		if err := os.WriteFile(primary, []byte("# created by envkit\n"), 0644); err != nil {
			return fmt.Errorf("创建 shell 配置失败 %s: %w", primary, err)
		}
	}

	written := 0
	var lastErr error
	for _, file := range files {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			continue
		}
		content, err := os.ReadFile(file)
		if err != nil {
			lastErr = err
			continue
		}
		// 已包含该目录或标签则跳过
		if strings.Contains(string(content), dir) || strings.Contains(string(content), pathValue) {
			if strings.Contains(string(content), "# envkit:"+tag) || strings.Contains(string(content), pathValue) {
				written++
				continue
			}
		}
		f, err := os.OpenFile(file, os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			lastErr = err
			continue
		}
		if _, err := f.WriteString(exportCmd); err != nil {
			_ = f.Close()
			lastErr = err
			continue
		}
		_ = f.Close()
		written++
	}
	if written == 0 {
		if lastErr != nil {
			return fmt.Errorf("写入 PATH 到 shell 配置失败: %w", lastErr)
		}
		return fmt.Errorf("未能将 PATH 写入任何 shell 配置文件")
	}
	return nil
}

// PersistUserEnvVar 将环境变量持久化到 Windows 用户级环境（User）。
// 始终写入，不因当前进程已 Setenv 而跳过。优先 PowerShell Environment API，失败时回退 setx。
func PersistUserEnvVar(name, value string) error {
	if runtime.GOOS != "windows" {
		return fmt.Errorf("PersistUserEnvVar 仅支持 Windows")
	}
	if err := userEnvSetWindows(name, value); err != nil {
		// 回退 setx
		if err2 := exec.Command("setx", name, value).Run(); err2 != nil {
			return fmt.Errorf("持久化用户环境变量 %s 失败: %w", name, err)
		}
	}
	return nil
}

// userEnvGetWindows 读取 Windows 用户级环境变量（HKCU）
func userEnvGetWindows(key string) (string, error) {
	escapedKey := strings.ReplaceAll(key, "'", "''")
	script := fmt.Sprintf("[Environment]::GetEnvironmentVariable('%s','User')", escapedKey)
	cmd := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", script)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("读取用户环境变量 %s 失败: %w (%s)", key, err, strings.TrimSpace(string(out)))
	}
	return strings.TrimRight(string(out), "\r\n"), nil
}

// userEnvSetWindows 写入 Windows 用户级环境变量（HKCU）
func userEnvSetWindows(key, value string) error {
	escapedKey := strings.ReplaceAll(key, "'", "''")
	escapedValue := strings.ReplaceAll(value, "'", "''")
	script := fmt.Sprintf("[Environment]::SetEnvironmentVariable('%s', '%s', 'User')", escapedKey, escapedValue)
	cmd := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", script)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("写入用户环境变量 %s 失败: %w (%s)", key, err, strings.TrimSpace(string(out)))
	}
	return nil
}

// getUserPathWindows 读取 Windows 用户级 Path（不含 Machine / 进程合并 PATH）
func getUserPathWindows() (string, error) {
	return userEnvGetWindows("Path")
}

// setUserPathWindows 写入 Windows 用户级 Path
func setUserPathWindows(path string) error {
	return userEnvSetWindows("Path", path)
}

// pathListContainsWindows 判断 dir 是否已在以 ; 分隔的 Path 中（大小写不敏感）
func pathListContainsWindows(pathList, dir string) bool {
	dir = strings.TrimRight(strings.TrimSpace(dir), `\`)
	if dir == "" {
		return false
	}
	for _, p := range strings.Split(pathList, ";") {
		p = strings.TrimRight(strings.TrimSpace(p), `\`)
		if p != "" && strings.EqualFold(p, dir) {
			return true
		}
	}
	return false
}

// persistPathEnvWindows 将 dir 追加到 Windows 用户级 Path（不使用 setx / 不全量写进程 PATH）
func persistPathEnvWindows(dir string) error {
	// 当前进程立即生效
	AddDirToPath(dir)

	userPath, err := getUserPathWindows()
	if err != nil {
		return err
	}

	if pathListContainsWindows(userPath, dir) {
		return nil
	}

	newPath := dir
	if trimmed := strings.TrimSpace(userPath); trimmed != "" {
		// 去掉尾部分隔符再追加，避免 ";;"
		trimmed = strings.TrimRight(trimmed, ";")
		newPath = trimmed + ";" + dir
	}

	if err := setUserPathWindows(newPath); err != nil {
		return err
	}

	ui.Info("PATH 已更新，请重启终端使其生效")
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

	names := []string{"fnm"}
	if runtime.GOOS == "windows" {
		names = append(names, "fnm.exe")
	}

	for _, dir := range dirs {
		for _, name := range names {
			fnmPath := filepath.Join(dir, name)
			if _, err := os.Stat(fnmPath); err == nil {
				AddDirToPath(dir)
				_ = PersistPathEnv(dir)
				persistFnmShellIntegration(dir)
				return nil
			}
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

// locateAndAddUvToPath 定位 uv 二进制，加入当前进程 PATH，并持久化到 shell 配置。
// 注意：即使 commandExists("uv") 已为 true，仍会写入 shell，避免「检测已安装但终端找不到」的情况。
func locateAndAddUvToPath() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	dirs := []string{
		filepath.Join(home, ".local", "bin"),
		filepath.Join(home, ".cargo", "bin"),
	}

	names := []string{"uv"}
	if runtime.GOOS == "windows" {
		names = append(names, "uv.exe")
	}

	for _, dir := range dirs {
		for _, name := range names {
			uvPath := filepath.Join(dir, name)
			if _, err := os.Stat(uvPath); err == nil {
				AddDirToPath(dir)
				if err := PersistPathEnvTagged(dir, "uv"); err != nil {
					ui.Warning("uv 已找到，但写入 shell PATH 失败: %v（请手动将 %s 加入 PATH）", err, dir)
				}
				return nil
			}
		}
	}

	if commandExists("uv") {
		// 在 PATH 中但非标准目录：仍尽量把 ~/.local/bin 写进配置（官方脚本默认位置）
		localBin := filepath.Join(home, ".local", "bin")
		AddDirToPath(localBin)
		_ = PersistPathEnvTagged(localBin, "uv")
		return nil
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
	names := []string{"rustup"}
	if runtime.GOOS == "windows" {
		names = append(names, "rustup.exe")
	}
	for _, name := range names {
		rustupPath := filepath.Join(dir, name)
		if _, err := os.Stat(rustupPath); err == nil {
			AddDirToPath(dir)
			_ = PersistPathEnv(dir)
			return nil
		}
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
