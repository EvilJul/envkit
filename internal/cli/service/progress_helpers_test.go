package service

import (
	"testing"

	"github.com/fusheng/envkit/internal/progress"
)

type captureReporter struct {
	events []progress.Event
}

func (c *captureReporter) Report(e progress.Event) {
	c.events = append(c.events, e)
}

func TestReportInstallStartNotFullBeforeDone(t *testing.T) {
	cap := &captureReporter{}
	old := progress.SetReporter(cap)
	defer progress.SetReporter(old)

	// 单步安装：开始不得 100%
	reportInstallStart("node", "正在安装 node...", 1, 1)
	if len(cap.events) != 1 {
		t.Fatalf("events=%d", len(cap.events))
	}
	if cap.events[0].Percent != 0 {
		t.Fatalf("start percent=%v want 0", cap.events[0].Percent)
	}

	reportInstallDone("node", "node 安装成功", 1, 1)
	if cap.events[1].Percent != 100 {
		t.Fatalf("done percent=%v want 100", cap.events[1].Percent)
	}
}

func TestReportInstallStartLastOfMany(t *testing.T) {
	cap := &captureReporter{}
	old := progress.SetReporter(cap)
	defer progress.SetReporter(old)

	// 3 步中开始第 3 步：应为 ~66%，不是 100%
	reportInstallStart("rust", "正在安装 rust...", 3, 3)
	if len(cap.events) != 1 {
		t.Fatalf("events=%d", len(cap.events))
	}
	p := cap.events[0].Percent
	if p >= 99 {
		t.Fatalf("start last step percent=%v, should be < 100", p)
	}
	if p < 66 || p > 67 {
		t.Fatalf("start last step percent=%v want ~66.67", p)
	}

	reportInstallDone("rust", "rust 安装成功", 3, 3)
	if cap.events[1].Percent != 100 {
		t.Fatalf("done last step percent=%v want 100", cap.events[1].Percent)
	}
}

func TestReportInstallErrorDoesNotResetPercent(t *testing.T) {
	cap := &captureReporter{}
	old := progress.SetReporter(cap)
	defer progress.SetReporter(old)

	reportInstallError("go", "安装失败")
	if cap.events[0].Percent >= 0 {
		t.Fatalf("error percent=%v want < 0 (no bar update)", cap.events[0].Percent)
	}
}
