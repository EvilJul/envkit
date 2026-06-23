//go:build windows

package installer

import (
	"fmt"
	"strings"

	"golang.org/x/sys/windows/registry"
	"github.com/fusheng/envkit/internal/ui"
)

// persistPathEnvWindows 通过 Windows 注册表持久化用户 PATH 环境变量
// 相比 setx 命令，注册表方式没有 1024 字符限制，且能正确分离用户 PATH 和系统 PATH
func persistPathEnvWindows(dir string) error {
	// 从注册表读取用户级 PATH（不会包含系统 PATH）
	key, err := registry.OpenKey(registry.CURRENT_USER, `Environment`, registry.READ)
	if err != nil {
		return fmt.Errorf("打开注册表失败: %w", err)
	}

	userPath, _, err := key.GetStringValue("Path")
	key.Close()
	if err != nil && err != registry.ErrNotExist {
		return fmt.Errorf("读取用户 PATH 失败: %w", err)
	}

	// 检查是否已存在
	pathEntries := strings.Split(userPath, ";")
	for _, entry := range pathEntries {
		if strings.EqualFold(strings.TrimSpace(entry), dir) {
			return nil // 已存在，无需重复添加
		}
	}

	// 追加新路径
	newPath := dir
	if userPath != "" {
		newPath = userPath + ";" + dir
	}

	// 写回注册表
	key, err = registry.OpenKey(registry.CURRENT_USER, `Environment`, registry.WRITE)
	if err != nil {
		return fmt.Errorf("打开注册表写入失败: %w", err)
	}
	defer key.Close()

	if err := key.SetStringValue("Path", newPath); err != nil {
		return fmt.Errorf("写入注册表 PATH 失败: %w", err)
	}

	ui.Info("PATH 已更新，请重启终端使其生效")
	return nil
}
