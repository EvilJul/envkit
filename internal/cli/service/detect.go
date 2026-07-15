package service

import (
	"github.com/fusheng/envkit/internal/detector"
)

// DetectReport 环境检测报告
type DetectReport struct {
	System    detector.SystemInfo
	Languages map[string]*detector.Tool
	Tools     map[string]*detector.Tool
	Managers  map[string]*detector.Tool
}

// DetectEnvironment 检测系统环境
func DetectEnvironment() DetectReport {
	sys := detector.DetectSystem()
	var system detector.SystemInfo
	if sys != nil {
		system = *sys
	}
	return DetectReport{
		System:    system,
		Languages: detector.DetectLanguages(),
		Tools:     detector.DetectTools(),
		Managers:  detector.DetectPackageManagers(),
	}
}

// StatusLabel 返回安装状态文本（无 ANSI，供 TUI 使用）
func StatusLabel(tool *detector.Tool) string {
	if tool != nil && tool.Installed {
		return "已安装 (" + tool.Version + ")"
	}
	return "未安装"
}

// StatusLabelANSI 返回带颜色的安装状态（CLI Plain 模式）
func StatusLabelANSI(tool *detector.Tool, green, yellow func(string) string) string {
	if tool != nil && tool.Installed {
		return green("✓ 已安装 (" + tool.Version + ")")
	}
	return yellow("✗ 未安装")
}
