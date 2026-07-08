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
	progress.ReportStep(taskID, "installing", message, step, total)
}

func reportInstallDone(taskID, message string, step, total int) {
	progress.ReportStep(taskID, "done", message, step, total)
}

func reportInstallError(taskID, message string) {
	progress.Report(progress.Event{
		TaskID:  taskID,
		Stage:   "error",
		Percent: 0,
		Message: message,
	})
}

// re-export for service package - use progress.Report directly in install.go
