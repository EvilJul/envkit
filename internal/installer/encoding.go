package installer

import (
	"io"
	"runtime"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

// NewWindowsConsoleWriter 创建适配 Windows 控制台编码的 Writer
// Windows 控制台可能使用 GBK 编码，需要从 UTF-8 转换
func NewWindowsConsoleWriter(w io.Writer) io.Writer {
	if runtime.GOOS != "windows" {
		return w
	}
	// Windows 控制台可能使用 GBK 编码
	return transform.NewWriter(w, simplifiedchinese.GBK.NewEncoder())
}
