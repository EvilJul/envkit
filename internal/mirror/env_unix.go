//go:build !windows

package mirror

// setWindowsEnvVar 在非 Windows 平台上为空实现
func setWindowsEnvVar(name, value string) error {
	return nil
}
