package ui

// SpinnerHandle 旋转加载器句柄
type SpinnerHandle interface {
	Stop()
	UpdateMessage(message string)
}

// Renderer CLI 输出渲染器（Plain 用于脚本模式，Silent 用于 TUI）
type Renderer interface {
	Success(format string, args ...interface{})
	Error(format string, args ...interface{})
	Warning(format string, args ...interface{})
	Info(format string, args ...interface{})
	NewSpinner(message string) SpinnerHandle
}

var defaultRenderer Renderer = &plainRenderer{}

// SetRenderer 设置全局渲染器（TUI 模式注入 SilentRenderer）
func SetRenderer(r Renderer) {
	if r != nil {
		defaultRenderer = r
	}
}

// GetRenderer 返回当前渲染器
func GetRenderer() Renderer {
	return defaultRenderer
}

// ResetRenderer 恢复默认 Plain 渲染器
func ResetRenderer() {
	defaultRenderer = &plainRenderer{}
}

// IsSilentMode 当前是否为静默渲染模式（TUI）
func IsSilentMode() bool {
	_, ok := defaultRenderer.(*SilentRenderer)
	return ok
}
