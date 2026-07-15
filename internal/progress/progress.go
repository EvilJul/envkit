package progress

// Event 任务进度事件（TUI / Wails / CLI 共用）
type Event struct {
	TaskID  string  `json:"taskId"`
	Stage   string  `json:"stage"`
	Percent float64 `json:"percent"`
	Message string  `json:"message"`
}

// Reporter 进度上报接口
type Reporter interface {
	Report(Event)
}

// NopReporter 忽略进度（CLI 脚本模式默认）
type NopReporter struct{}

func (NopReporter) Report(Event) {}

var defaultReporter Reporter = NopReporter{}

// SetReporter 设置全局 reporter，返回上一个 reporter 便于 defer 恢复
func SetReporter(r Reporter) Reporter {
	old := defaultReporter
	if r == nil {
		defaultReporter = NopReporter{}
	} else {
		defaultReporter = r
	}
	return old
}

// GetReporter 返回当前 reporter
func GetReporter() Reporter {
	return defaultReporter
}

// Report 向当前 reporter 发送事件
func Report(e Event) {
	if defaultReporter != nil {
		defaultReporter.Report(e)
	}
}

// ReportStep 按步骤上报（current/total 转为 0–100 百分比）。
// current 表示「已完成」步数（0..total）；开始某步时传 completed=step-1，完成时传 step。
func ReportStep(taskID, stage, message string, current, total int) {
	if total <= 0 {
		total = 1
	}
	if current < 0 {
		current = 0
	}
	if current > total {
		current = total
	}
	pct := float64(current) / float64(total) * 100
	if pct > 100 {
		pct = 100
	}
	Report(Event{
		TaskID:  taskID,
		Stage:   stage,
		Percent: pct,
		Message: message,
	})
}

// StepPercent 计算步骤进度百分比（与 ReportStep 一致，便于测试）
func StepPercent(completed, total int) float64 {
	if total <= 0 {
		total = 1
	}
	if completed < 0 {
		completed = 0
	}
	if completed > total {
		completed = total
	}
	return float64(completed) / float64(total) * 100
}
