package ui

// SilentRenderer 不输出任何内容的渲染器（TUI 模式）
type SilentRenderer struct{}

// NewSilentRenderer 创建静默渲染器
func NewSilentRenderer() Renderer {
	return &SilentRenderer{}
}

type noopSpinnerHandle struct{}

func (SilentRenderer) Success(format string, args ...interface{}) {}
func (SilentRenderer) Error(format string, args ...interface{})   {}
func (SilentRenderer) Warning(format string, args ...interface{}) {}
func (SilentRenderer) Info(format string, args ...interface{})    {}
func (SilentRenderer) NewSpinner(message string) SpinnerHandle    { return noopSpinnerHandle{} }

func (noopSpinnerHandle) Stop()                          {}
func (noopSpinnerHandle) UpdateMessage(message string) {}
