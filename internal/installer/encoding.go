package installer

import "io"

// NewWindowsConsoleWriter 创建适配 Windows 控制台编码的 Writer
// 在 Windows 上检测控制台代码页，仅在使用 GBK 等非 UTF-8 编码时才转换
// 在 Unix 平台上直接返回原始 Writer
func NewWindowsConsoleWriter(w io.Writer) io.Writer {
	return newWindowsConsoleWriter(w)
}
