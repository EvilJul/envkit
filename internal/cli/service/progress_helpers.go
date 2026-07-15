package service

import (
	"github.com/fusheng/envkit/internal/config"
	"github.com/fusheng/envkit/internal/progress"
)

// CountInstallSteps 计算安装总步数（每个 language/tool/database 一步）
func CountInstallSteps(cfg *config.Config) int {
	n := 0
	for _, db := range cfg.Databases {
		if db.Docker {
			n++
		}
	}
	return len(cfg.Languages) + len(cfg.Tools) + n
}

func reportInstallStart(taskID, message string, step, total int) {
	// 开始第 step 步时进度为已完成 (step-1)/total，避免最后一步一开始就显示 100%
	completed := step - 1
	if completed < 0 {
		completed = 0
	}
	progress.ReportStep(taskID, "installing", message, completed, total)
}

func reportInstallDone(taskID, message string, step, total int) {
	// 完成第 step 步后进度为 step/total
	progress.ReportStep(taskID, "done", message, step, total)
}

func reportInstallError(taskID, message string) {
	// Percent < 0：不改动总体进度条（避免失败事件把进度打回 0%）
	progress.Report(progress.Event{
		TaskID:  taskID,
		Stage:   "error",
		Percent: -1,
		Message: message,
	})
}

// re-export for service package - use progress.Report directly in install.go
