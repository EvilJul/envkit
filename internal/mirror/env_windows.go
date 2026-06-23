//go:build windows

package mirror

import (
	"fmt"

	"golang.org/x/sys/windows/registry"
)

// setWindowsEnvVar 通过注册表设置 Windows 用户级环境变量（持久化）
func setWindowsEnvVar(name, value string) error {
	key, err := registry.OpenKey(registry.CURRENT_USER, `Environment`, registry.WRITE)
	if err != nil {
		return fmt.Errorf("打开注册表失败: %w", err)
	}
	defer key.Close()

	if err := key.SetStringValue(name, value); err != nil {
		return fmt.Errorf("写入注册表 %s 失败: %w", name, err)
	}
	return nil
}
