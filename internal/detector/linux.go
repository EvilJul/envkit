package detector

import (
	"os"
	"os/exec"
	"strings"
)

// detectLinuxDistribution 检测Linux发行版
func detectLinuxDistribution() string {
	// 尝试读取 /etc/os-release
	data, err := os.ReadFile("/etc/os-release")
	if err == nil {
		content := string(data)
		if strings.Contains(content, "ubuntu") || strings.Contains(content, "Ubuntu") {
			return "ubuntu"
		}
		if strings.Contains(content, "debian") || strings.Contains(content, "Debian") {
			return "debian"
		}
		if strings.Contains(content, "fedora") || strings.Contains(content, "Fedora") {
			return "fedora"
		}
		if strings.Contains(content, "centos") || strings.Contains(content, "CentOS") {
			return "centos"
		}
		if strings.Contains(content, "arch") || strings.Contains(content, "Arch") {
			return "arch"
		}
		if strings.Contains(content, "opensuse") || strings.Contains(content, "openSUSE") {
			return "opensuse"
		}
	}

	// 尝试使用 lsb_release 命令
	cmd := exec.Command("lsb_release", "-is")
	output, err := cmd.Output()
	if err == nil {
		distro := strings.ToLower(strings.TrimSpace(string(output)))
		if distro != "" {
			return distro
		}
	}

	return "unknown"
}
