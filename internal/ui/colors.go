package ui

import "fmt"

// ANSI 颜色代码
const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
	ColorPurple = "\033[35m"
	ColorCyan   = "\033[36m"
	ColorWhite  = "\033[37m"
	ColorBold   = "\033[1m"
)

// 彩色输出函数
func Success(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Printf("%s✓ %s%s\n", ColorGreen, msg, ColorReset)
}

func Error(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Printf("%s✗ %s%s\n", ColorRed, msg, ColorReset)
}

func Warning(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Printf("%s⚠ %s%s\n", ColorYellow, msg, ColorReset)
}

func Info(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Printf("%sℹ %s%s\n", ColorBlue, msg, ColorReset)
}

func Bold(text string) string {
	return fmt.Sprintf("%s%s%s", ColorBold, text, ColorReset)
}

func Green(text string) string {
	return fmt.Sprintf("%s%s%s", ColorGreen, text, ColorReset)
}

func Red(text string) string {
	return fmt.Sprintf("%s%s%s", ColorRed, text, ColorReset)
}

func Yellow(text string) string {
	return fmt.Sprintf("%s%s%s", ColorYellow, text, ColorReset)
}

func Blue(text string) string {
	return fmt.Sprintf("%s%s%s", ColorBlue, text, ColorReset)
}

func Cyan(text string) string {
	return fmt.Sprintf("%s%s%s", ColorCyan, text, ColorReset)
}
