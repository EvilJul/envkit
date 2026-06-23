//go:build !windows

package installer

import "io"

// newWindowsConsoleWriter Unix 平台无需编码转换，直接返回原始 Writer
func newWindowsConsoleWriter(w io.Writer) io.Writer {
	return w
}
