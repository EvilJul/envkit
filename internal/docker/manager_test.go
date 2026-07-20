package docker

import (
	"strings"
	"testing"
)

func TestContainerExistsExactMatchLogic(t *testing.T) {
	// 模拟 docker filter 子串匹配输出的精确行判断
	name := "envkit-postgres"
	outputs := []struct {
		raw  string
		want bool
	}{
		{"envkit-postgres\n", true},
		{"envkit-postgres-backup\n", false},
		{"envkit-postgres\nenvkit-postgres-backup\n", true},
		{"envkit-mysql\n", false},
		{"", false},
	}
	for _, tc := range outputs {
		got := false
		for _, line := range strings.Split(tc.raw, "\n") {
			if strings.TrimSpace(line) == name {
				got = true
				break
			}
		}
		if got != tc.want {
			t.Errorf("raw=%q got %v want %v", tc.raw, got, tc.want)
		}
	}
}
