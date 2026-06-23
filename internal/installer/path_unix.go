//go:build !windows

package installer

// persistPathEnvWindows 在非 Windows 平台上为空实现
// 实际的 PATH 持久化由 PersistPathEnv 中的 shell 配置文件写入逻辑处理
func persistPathEnvWindows(dir string) error {
	return nil
}
