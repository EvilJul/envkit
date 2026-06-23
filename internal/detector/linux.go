package detector

import (
	"bufio"
	"os"
	"os/exec"
	"strings"
)

// detectLinuxDistribution 检测Linux发行版
func detectLinuxDistribution() string {
	// 尝试读取 /etc/os-release，优先解析 ID= 字段
	data, err := os.ReadFile("/etc/os-release")
	if err == nil {
		id := parseOSReleaseID(string(data))
		if id != "" {
			return id
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

// parseOSReleaseID 从 /etc/os-release 内容中解析 ID 字段
// 优先匹配 ID= 行，避免 ID_LIKE= 等字段导致误判
func parseOSReleaseID(content string) string {
	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "ID=") {
			id := strings.Trim(strings.TrimPrefix(line, "ID"), `"'`)
			return strings.ToLower(id)
		}
	}
	// fallback: 用 Contains 做模糊匹配
	lower := strings.ToLower(content)
	knownDistros := []string{"ubuntu", "debian", "fedora", "centos", "arch", "opensuse", "alpine", "linuxmint", "pop"}
	for _, distro := range knownDistros {
		if strings.Contains(lower, distro) {
			return distro
		}
	}
	return ""
}
