package installer

import (
	"io"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

var (
	winConsoleUTF8Once sync.Once
	winConsoleUTF8     bool
)

// NewWindowsConsoleWriter 创建适配 Windows 控制台编码的 Writer。
// 仅当控制台不是 UTF-8（代码页 65001）时做 UTF-8→GBK 转换；UTF-8 控制台直接透传，避免乱码。
func NewWindowsConsoleWriter(w io.Writer) io.Writer {
	if runtime.GOOS != "windows" {
		return w
	}
	if windowsConsoleIsUTF8() {
		return w
	}
	return transform.NewWriter(w, simplifiedchinese.GBK.NewEncoder())
}

// windowsConsoleIsUTF8 检测当前控制台输出代码页是否为 UTF-8 (65001)
func windowsConsoleIsUTF8() bool {
	winConsoleUTF8Once.Do(func() {
		// Windows Terminal 通常为 UTF-8
		if os.Getenv("WT_SESSION") != "" {
			winConsoleUTF8 = true
			return
		}
		out, err := exec.Command("cmd", "/C", "chcp").Output()
		if err != nil {
			// 检测失败时按传统中文控制台处理（应用 GBK）
			winConsoleUTF8 = false
			return
		}
		// 输出示例: "Active code page: 65001" / "活动代码页: 65001"
		winConsoleUTF8 = strings.Contains(string(out), "65001")
	})
	return winConsoleUTF8
}
