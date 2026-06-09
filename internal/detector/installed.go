package detector

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// Tool 表示一个工具的检测结果
type Tool struct {
	Name      string
	Installed bool
	Version   string
	Path      string
}

// lookPathWithFallback 尝试在系统的 PATH 中查找命令，如果找不到，则在常见的默认安装路径中查找
func lookPathWithFallback(name string) (string, error) {
	path, err := exec.LookPath(name)
	if err == nil {
		return path, nil
	}

	home, errDir := os.UserHomeDir()
	if errDir != nil {
		return "", err
	}

	fallbacks := []string{
		filepath.Join(home, ".cargo", "bin"),
		filepath.Join(home, ".local", "bin"),
		filepath.Join(home, ".local", "go", "bin"),
		filepath.Join(home, "miniconda3", "bin"),
		filepath.Join(home, "miniconda3", "condabin"),
		filepath.Join(home, ".fnm"),
		filepath.Join(home, ".local", "share", "fnm"),
		filepath.Join(home, "Library", "Application Support", "fnm"),
		"/usr/local/go/bin",
	}

	for _, dir := range fallbacks {
		binPath := filepath.Join(dir, name)
		if runtime.GOOS == "windows" {
			binPath += ".exe"
		}
		if _, err := os.Stat(binPath); err == nil {
			return binPath, nil
		}
	}

	return "", err
}

// CheckCommand 检查命令是否存在
func CheckCommand(name string) bool {
	_, err := lookPathWithFallback(name)
	return err == nil
}

// GetCommandVersion 获取命令的版本号
func GetCommandVersion(name string, versionFlag string) string {
	path, err := lookPathWithFallback(name)
	if err != nil {
		return ""
	}

	args := strings.Fields(versionFlag)
	cmd := exec.Command(path, args...)
	// Some commands (like java) output version info to stderr, so we use CombinedOutput
	output, err := cmd.CombinedOutput()
	if err != nil {
		return ""
	}

	cmdName := filepath.Base(name)
	return cleanVersionString(cmdName, string(output))
}

// cleanVersionString 清理并精简版本号字符串，确保在 TUI 表格中友好显示
func cleanVersionString(name string, raw string) string {
	lines := strings.Split(raw, "\n")
	if len(lines) == 0 {
		return ""
	}

	// 获取第一个非空行
	firstLine := ""
	for _, l := range lines {
		l = strings.TrimSpace(l)
		if l != "" {
			firstLine = l
			break
		}
	}
	if firstLine == "" {
		return ""
	}

	// 针对常用命令进行精炼，只提取核心版本号
	switch name {
	case "vim":
		// 例如: "VIM - Vi IMproved 9.0 (2022 Jun 28 ...)" -> "9.0"
		if idx := strings.Index(firstLine, "VIM - Vi IMproved "); idx >= 0 {
			vStr := firstLine[idx+len("VIM - Vi IMproved "):]
			if spaceIdx := strings.Index(vStr, " "); spaceIdx >= 0 {
				return vStr[:spaceIdx]
			}
			return vStr
		}
	case "curl":
		// 例如: "curl 8.7.1 (x86_64-apple-darwin21.0) ..." -> "8.7.1"
		parts := strings.Fields(firstLine)
		if len(parts) >= 2 && parts[0] == "curl" {
			return parts[1]
		}
	case "git":
		// 例如: "git version 2.37.1 (Apple Git-137.1)" -> "2.37.1"
		parts := strings.Fields(firstLine)
		if len(parts) >= 3 && parts[0] == "git" && parts[1] == "version" {
			return parts[2]
		}
	case "go":
		// 例如: "go version go1.22.0 darwin/amd64" -> "1.22.0"
		parts := strings.Fields(firstLine)
		if len(parts) >= 3 && parts[0] == "go" && parts[1] == "version" {
			return strings.TrimPrefix(parts[2], "go")
		}
	case "docker":
		// 例如: "Docker version 20.10.7, build f0df350" -> "20.10.7"
		parts := strings.Fields(firstLine)
		if len(parts) >= 3 && parts[0] == "Docker" && parts[1] == "version" {
			return strings.TrimSuffix(parts[2], ",")
		}
	case "pip", "pip3":
		// 例如: "pip 25.0.1 from ... (python 3.12)" -> "25.0.1"
		parts := strings.Fields(firstLine)
		if len(parts) >= 2 && parts[0] == "pip" {
			return parts[1]
		}
	case "ruby":
		// 例如: "ruby 2.6.10p210 (2022-04-12 revision 67958) ..." -> "2.6.10p210"
		parts := strings.Fields(firstLine)
		if len(parts) >= 2 && parts[0] == "ruby" {
			return parts[1]
		}
	case "python", "python3":
		// 例如: "Python 3.12.10" -> "3.12.10"
		parts := strings.Fields(firstLine)
		if len(parts) >= 2 && parts[0] == "Python" {
			return parts[1]
		}
	case "java":
		// 例如: "openjdk version \"17.0.10\" 2024-01-16" -> "17.0.10"
		parts := strings.Fields(firstLine)
		for _, part := range parts {
			if strings.HasPrefix(part, "\"") && strings.HasSuffix(part, "\"") {
				return strings.Trim(part, "\"")
			}
		}
	case "kubectl":
		// 例如: "Client Version: v1.30.0" -> "1.30.0"
		if idx := strings.Index(firstLine, "Client Version: "); idx >= 0 {
			vStr := strings.TrimPrefix(firstLine[idx+len("Client Version: "):], "v")
			return vStr
		}
	case "minikube":
		// 例如: "minikube version: v1.33.1" -> "1.33.1"
		if idx := strings.Index(firstLine, "minikube version: "); idx >= 0 {
			vStr := strings.TrimPrefix(firstLine[idx+len("minikube version: "):], "v")
			return vStr
		}
	}

	// 通用清洗规则：
	// 1. 去除括号及括号内的多余编译环境说明
	if idx := strings.Index(firstLine, "("); idx >= 0 {
		firstLine = strings.TrimSpace(firstLine[:idx])
	}
	// 2. 去除逗号及其后面的多余说明
	if idx := strings.Index(firstLine, ","); idx >= 0 {
		firstLine = strings.TrimSpace(firstLine[:idx])
	}
	// 3. 限制最长为 40 字符，以防表格布局错乱
	if len(firstLine) > 40 {
		firstLine = firstLine[:37] + "..."
	}

	return firstLine
}



// DetectTool 检测单个工具
func DetectTool(name string, versionFlag string) *Tool {
	path, err := lookPathWithFallback(name)
	tool := &Tool{
		Name:      name,
		Installed: err == nil,
		Path:      path,
	}

	if tool.Installed {
		tool.Version = GetCommandVersion(name, versionFlag)
		// 特殊处理 macOS 下的 java 占位符/存根：如果版本号获取为空，说明实际未安装 Java 运行时
		if name == "java" && tool.Version == "" {
			tool.Installed = false
		}
	}

	return tool
}

// DetectLanguages 检测常见编程语言
func DetectLanguages() map[string]*Tool {
	languages := map[string]*Tool{
		"node":   DetectTool("node", "--version"),
		"npm":    DetectTool("npm", "--version"),
		"python": DetectTool("python3", "--version"),
		"pip":    DetectTool("pip3", "--version"),
		"go":     DetectTool("go", "version"),
		"rustc":  DetectTool("rustc", "--version"),
		"cargo":  DetectTool("cargo", "--version"),
		"java":   DetectTool("java", "-version"),
		"bun":    DetectTool("bun", "--version"),
		"ruby":   DetectTool("ruby", "--version"),
		"php":    DetectTool("php", "--version"),
	}

	return languages
}

// DetectTools 检测常见开发工具
func DetectTools() map[string]*Tool {
	tools := map[string]*Tool{
		"git":      DetectTool("git", "--version"),
		"docker":   DetectTool("docker", "--version"),
		"code":     DetectTool("code", "--version"),
		"vim":      DetectTool("vim", "--version"),
		"curl":     DetectTool("curl", "--version"),
		"wget":     DetectTool("wget", "--version"),
		"conda":    DetectTool("conda", "--version"),
		"kubectl":  DetectTool("kubectl", "version --client"),
		"minikube": DetectTool("minikube", "version"),
	}

	return tools
}

// DetectPackageManagers 检测包管理器
func DetectPackageManagers() map[string]*Tool {
	managers := make(map[string]*Tool)

	// 通用包管理器
	managers["brew"] = DetectTool("brew", "--version")

	// Linux包管理器
	managers["apt"] = DetectTool("apt", "--version")
	managers["yum"] = DetectTool("yum", "--version")
	managers["dnf"] = DetectTool("dnf", "--version")
	managers["pacman"] = DetectTool("pacman", "--version")
	managers["zypper"] = DetectTool("zypper", "--version")

	// Windows包管理器
	managers["winget"] = DetectTool("winget", "--version")
	managers["choco"] = DetectTool("choco", "--version")
	managers["scoop"] = DetectTool("scoop", "--version")

	return managers
}
