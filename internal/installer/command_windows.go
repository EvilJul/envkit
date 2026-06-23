//go:build windows

package installer

import (
	"os/exec"
	"syscall"
)

// configureWindowsCommand 为Windows平台配置命令，隐藏控制台窗口
func configureWindowsCommand(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: syscall.CREATE_NO_WINDOW,
	}
}
