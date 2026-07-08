package progress

import (
	"context"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// WailsReporter 通过 Wails runtime 向前端推送进度
type WailsReporter struct {
	ctx       context.Context
	eventName string
	taskID    string
}

// NewWailsReporter 创建 Wails 事件 reporter
func NewWailsReporter(ctx context.Context, eventName, taskID string) *WailsReporter {
	if eventName == "" {
		eventName = "task-progress"
	}
	return &WailsReporter{ctx: ctx, eventName: eventName, taskID: taskID}
}

func (w *WailsReporter) Report(e Event) {
	if w == nil || w.ctx == nil {
		return
	}
	if e.TaskID == "" {
		e.TaskID = w.taskID
	}
	runtime.EventsEmit(w.ctx, w.eventName, e)
}
