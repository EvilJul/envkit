package detector

import (
	"os/exec"
	"strings"
)

// Tool 表示一个工具的检测结果
type Tool struct {
	Name      string
	Installed bool
	Version   string
	Path      string
}

// CheckCommand 检查命令是否存在
func CheckCommand(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

// GetCommandVersion 获取命令的版本号
func GetCommandVersion(name string, versionFlag string) string {
	if !CheckCommand(name) {
		return ""
	}

	cmd := exec.Command(name, versionFlag)
	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	return strings.TrimSpace(string(output))
}

// DetectTool 检测单个工具
func DetectTool(name string, versionFlag string) *Tool {
	tool := &Tool{
		Name:      name,
		Installed: CheckCommand(name),
	}

	if tool.Installed {
		path, _ := exec.LookPath(name)
		tool.Path = path
		tool.Version = GetCommandVersion(name, versionFlag)
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
		"ruby":   DetectTool("ruby", "--version"),
		"php":    DetectTool("php", "--version"),
	}

	return languages
}

// DetectTools 检测常见开发工具
func DetectTools() map[string]*Tool {
	tools := map[string]*Tool{
		"git":    DetectTool("git", "--version"),
		"docker": DetectTool("docker", "--version"),
		"code":   DetectTool("code", "--version"),
		"vim":    DetectTool("vim", "--version"),
		"curl":   DetectTool("curl", "--version"),
		"wget":   DetectTool("wget", "--version"),
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
