//go:build windows

package main

import (
	"fmt"
	"strings"

	"golang.org/x/sys/windows/registry"
)

// getWindowsUserEnvVars 从注册表读取用户级环境变量
func getWindowsUserEnvVars() []map[string]string {
	var vars []map[string]string

	key, err := registry.OpenKey(registry.CURRENT_USER, `Environment`, registry.READ)
	if err != nil {
		return vars
	}
	defer key.Close()

	names, err := key.ReadValueNames(-1)
	if err != nil {
		return vars
	}

	for _, name := range names {
		val, _, err := key.GetStringValue(name)
		if err != nil {
			continue
		}
		// 跳过 PATH 变量（单独处理）
		if strings.EqualFold(name, "Path") {
			continue
		}
		vars = append(vars, map[string]string{
			"key":    name,
			"value":  val,
			"source": "注册表 (用户)",
			"scope":  "user",
		})
	}
	return vars
}

// setWindowsEnvVar 通过注册表设置用户级环境变量
func setWindowsEnvVar(key, value string) error {
	k, err := registry.OpenKey(registry.CURRENT_USER, `Environment`, registry.WRITE)
	if err != nil {
		return fmt.Errorf("打开注册表失败: %w", err)
	}
	defer k.Close()

	if err := k.SetStringValue(key, value); err != nil {
		return fmt.Errorf("写入注册表失败: %w", err)
	}
	return nil
}

// deleteWindowsEnvVar 通过注册表删除用户级环境变量
func deleteWindowsEnvVar(key string) error {
	k, err := registry.OpenKey(registry.CURRENT_USER, `Environment`, registry.WRITE)
	if err != nil {
		return fmt.Errorf("打开注册表失败: %w", err)
	}
	defer k.Close()

	if err := k.DeleteValue(key); err != nil && err != registry.ErrNotExist {
		return fmt.Errorf("删除注册表键失败: %w", err)
	}
	return nil
}

// getWindowsUserPath 从注册表读取用户级 PATH
func getWindowsUserPath() []string {
	key, err := registry.OpenKey(registry.CURRENT_USER, `Environment`, registry.READ)
	if err != nil {
		return nil
	}
	defer key.Close()

	path, _, err := key.GetStringValue("Path")
	if err != nil {
		return nil
	}

	var entries []string
	for _, entry := range strings.Split(path, ";") {
		entry = strings.TrimSpace(entry)
		if entry != "" {
			entries = append(entries, entry)
		}
	}
	return entries
}
