//go:build !windows

package main

// getWindowsUserEnvVars 在非 Windows 平台上返回空列表
func getWindowsUserEnvVars() []map[string]string {
	return nil
}

// setWindowsEnvVar 在非 Windows 平台上为空实现
func setWindowsEnvVar(key, value string) error {
	return nil
}

// deleteWindowsEnvVar 在非 Windows 平台上为空实现
func deleteWindowsEnvVar(key string) error {
	return nil
}

// getWindowsUserPath 在非 Windows 平台上返回 nil
func getWindowsUserPath() []string {
	return nil
}
