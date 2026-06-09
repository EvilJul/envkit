package installer

import (
	"fmt"
	"runtime"

	"github.com/fusheng/envkit/internal/ui"
)

// SystemDependency 系统依赖
type SystemDependency struct {
	Name        string
	Commands    []string          // 可能的命令名称
	PackageName map[string]string // 包管理器 -> 包名
	Required    bool              // 是否必需
}

var CommonDependencies = []SystemDependency{
	{
		Name:     "curl",
		Commands: []string{"curl"},
		PackageName: map[string]string{
			"apt":    "curl",
			"yum":    "curl",
			"dnf":    "curl",
			"pacman": "curl",
			"brew":   "curl",
		},
		Required: true,
	},
	{
		Name:     "tar",
		Commands: []string{"tar"},
		PackageName: map[string]string{
			"apt":    "tar",
			"yum":    "tar",
			"dnf":    "tar",
			"pacman": "tar",
		},
		Required: true,
	},
	{
		Name:     "unzip",
		Commands: []string{"unzip"},
		PackageName: map[string]string{
			"apt":    "unzip",
			"yum":    "unzip",
			"dnf":    "unzip",
			"pacman": "unzip",
		},
		Required: false,
	},
	{
		Name:     "zip",
		Commands: []string{"zip"},
		PackageName: map[string]string{
			"apt":    "zip",
			"yum":    "zip",
			"dnf":    "zip",
			"pacman": "zip",
		},
		Required: false,
	},
	{
		Name:     "bash",
		Commands: []string{"bash"},
		PackageName: map[string]string{
			"apt":    "bash",
			"yum":    "bash",
			"dnf":    "bash",
			"pacman": "bash",
		},
		Required: true,
	},
	{
		Name:     "gpg",
		Commands: []string{"gpg", "gpg2"},
		PackageName: map[string]string{
			"apt":    "gnupg",
			"yum":    "gnupg2",
			"dnf":    "gnupg2",
			"pacman": "gnupg",
		},
		Required: false,
	},
}

// CheckAndInstallDependencies 检查并安装系统依赖
func CheckAndInstallDependencies(deps []SystemDependency) error {
	if runtime.GOOS == "windows" {
		return nil // Windows 不检查系统工具
	}

	var missing []SystemDependency

	for _, dep := range deps {
		found := false
		for _, cmd := range dep.Commands {
			if commandExists(cmd) {
				found = true
				break
			}
		}
		if !found {
			missing = append(missing, dep)
		}
	}

	if len(missing) == 0 {
		return nil
	}

	// 检测包管理器
	pkgManager := detectPackageManager()
	if pkgManager == "" {
		var names []string
		for _, dep := range missing {
			names = append(names, dep.Name)
		}
		return fmt.Errorf("缺少系统依赖: %v，请手动安装", names)
	}

	// 尝试自动安装
	ui.Info("检测到缺少系统依赖，正在尝试安装...")
	for _, dep := range missing {
		pkgName, ok := dep.PackageName[pkgManager]
		if !ok {
			if dep.Required {
				return fmt.Errorf("无法为包管理器 %s 找到 %s 的包名", pkgManager, dep.Name)
			}
			continue
		}

		ui.Info("正在安装 %s...", dep.Name)
		if err := installSystemPackage(pkgManager, pkgName); err != nil {
			if dep.Required {
				return fmt.Errorf("安装 %s 失败: %w", dep.Name, err)
			}
			ui.Warning("安装 %s 失败: %v", dep.Name, err)
		} else {
			ui.Success("%s 安装成功", dep.Name)
		}
	}

	return nil
}

// detectPackageManager 检测系统包管理器
func detectPackageManager() string {
	managers := []string{"apt-get", "yum", "dnf", "pacman", "brew"}
	for _, mgr := range managers {
		if commandExists(mgr) {
			if mgr == "apt-get" {
				return "apt"
			}
			return mgr
		}
	}
	return ""
}

// installSystemPackage 安装系统包
func installSystemPackage(manager, packageName string) error {
	switch manager {
	case "apt":
		_ = runCommand("sudo", "apt-get", "update")
		return runCommand("sudo", "apt-get", "install", "-y", packageName)
	case "yum":
		return runCommand("sudo", "yum", "install", "-y", packageName)
	case "dnf":
		return runCommand("sudo", "dnf", "install", "-y", packageName)
	case "pacman":
		return runCommand("sudo", "pacman", "-S", "--noconfirm", packageName)
	case "brew":
		return runCommand("brew", "install", packageName)
	default:
		return fmt.Errorf("不支持的包管理器: %s", manager)
	}
}
