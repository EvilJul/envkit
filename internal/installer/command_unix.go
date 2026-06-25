//go:build !windows

package installer

import "os/exec"

// configureWindowsCommand Unix平台空实现
func configureWindowsCommand(cmd *exec.Cmd) {
	// Unix平台不需要窗口隐藏配置
}
