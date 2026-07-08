package ui

import (
	"fmt"
	"strings"
	"time"
)

// ProgressBar 进度条结构
type ProgressBar struct {
	total   int
	current int
	width   int
	prefix  string
	started time.Time
}

// NewProgressBar 创建新的进度条
func NewProgressBar(total int, prefix string) *ProgressBar {
	return &ProgressBar{
		total:   total,
		current: 0,
		width:   40,
		prefix:  prefix,
		started: time.Now(),
	}
}

// Increment 增加进度
func (p *ProgressBar) Increment() {
	p.current++
	p.render()
}

// Set 设置当前进度
func (p *ProgressBar) Set(current int) {
	p.current = current
	p.render()
}

// Finish 完成进度条
func (p *ProgressBar) Finish() {
	p.current = p.total
	p.render()
	fmt.Println()
}

// render 渲染进度条
func (p *ProgressBar) render() {
	percent := float64(p.current) / float64(p.total)
	filled := int(percent * float64(p.width))

	bar := strings.Repeat("█", filled) + strings.Repeat("░", p.width-filled)

	elapsed := time.Since(p.started)
	var eta string
	if p.current > 0 {
		totalTime := elapsed.Seconds() * float64(p.total) / float64(p.current)
		remaining := totalTime - elapsed.Seconds()
		eta = fmt.Sprintf("ETA: %ds", int(remaining))
	}

	fmt.Printf("\r%s [%s] %d/%d (%.0f%%) %s",
		p.prefix, bar, p.current, p.total, percent*100, eta)
}

// Spinner 旋转加载器
type Spinner struct {
	frames  []string
	current int
	running bool
	message string
}

// NewSpinner 创建新的旋转加载器
func NewSpinner(message string) *Spinner {
	return &Spinner{
		frames:  []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
		current: 0,
		running: false,
		message: message,
	}
}

// Start 启动旋转加载器
func (s *Spinner) Start() {
	if IsSilentMode() {
		return
	}
	s.running = true
	go func() {
		for s.running {
			fmt.Printf("\r%s%s %s%s", ColorCyan, s.frames[s.current], s.message, ColorReset)
			s.current = (s.current + 1) % len(s.frames)
			time.Sleep(100 * time.Millisecond)
		}
	}()
}

// Stop 停止旋转加载器
func (s *Spinner) Stop() {
	if IsSilentMode() {
		return
	}
	s.running = false
	time.Sleep(150 * time.Millisecond) // 等待最后一帧显示
	fmt.Print("\r" + strings.Repeat(" ", 50) + "\r") // 清除行
}

// UpdateMessage 更新加载器消息
func (s *Spinner) UpdateMessage(message string) {
	s.message = message
}
