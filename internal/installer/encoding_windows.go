//go:build windows

package installer

import (
	"io"
	"syscall"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

var (
	kernel32         = syscall.NewLazyDLL("kernel32.dll")
	procGetConsoleCP = kernel32.NewProc("GetConsoleCP")
)

// getConsoleCodePage 获取当前控制台的代码页
func getConsoleCodePage() uint {
	ret, _, _ := procGetConsoleCP.Call()
	return uint(ret)
}

// newWindowsConsoleWriter 为 Windows 平台创建编码转换 Writer
// 仅在控制台使用 GBK (936) 等非 UTF-8 编码时才进行转换
// 现代 Windows Terminal 和 VS Code 终端默认使用 UTF-8 (65001)，无需转换
func newWindowsConsoleWriter(w io.Writer) io.Writer {
	cp := getConsoleCodePage()
	// UTF-8 (65001) 或未确定 (0) 不需要转换
	if cp == 65001 || cp == 0 {
		return w
	}
	// 只有中文代码页 (936=GBK, 950=Big5) 才做编码转换
	if cp == 936 || cp == 950 {
		return transform.NewWriter(w, simplifiedchinese.GBK.NewEncoder())
	}
	return w
}
