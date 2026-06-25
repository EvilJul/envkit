//go:build windows

package installer

import (
	"os/exec"
	"syscall"
)

const createNoWindow = 0x08000000

// configureWindowsCommand 为Windows平台配置命令，隐藏控制台窗口
func configureWindowsCommand(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: createNoWindow,
	}
}
