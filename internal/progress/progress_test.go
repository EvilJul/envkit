package progress

import "testing"

func TestStepPercentStartVsDone(t *testing.T) {
	// 单组件：开始应为 0%，完成应为 100%
	if got := StepPercent(0, 1); got != 0 {
		t.Fatalf("start 1/1: got %v want 0", got)
	}
	if got := StepPercent(1, 1); got != 100 {
		t.Fatalf("done 1/1: got %v want 100", got)
	}

	// 三组件：开始第 3 步 = 2/3 ≈ 66.67%，完成第 3 步 = 100%
	start3 := StepPercent(2, 3)
	if start3 < 66 || start3 > 67 {
		t.Fatalf("start step 3/3: got %v want ~66.67", start3)
	}
	if got := StepPercent(3, 3); got != 100 {
		t.Fatalf("done 3/3: got %v want 100", got)
	}

	// 开始第 1 步
	if got := StepPercent(0, 3); got != 0 {
		t.Fatalf("start 1/3: got %v want 0", got)
	}
	// 完成第 1 步
	if got := StepPercent(1, 3); got < 33 || got > 34 {
		t.Fatalf("done 1/3: got %v want ~33.33", got)
	}
}
