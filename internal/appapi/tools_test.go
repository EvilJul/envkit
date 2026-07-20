package appapi

import "testing"

func TestGetToolsNamesMatchInstaller(t *testing.T) {
	// 安装器支持的清单键，GetTools 的 Name 必须对齐
	allowed := map[string]bool{
		"git": true, "docker": true, "vscode": true, "uv": true, "miniconda": true,
		"kubectl": true, "minikube": true, "espidf": true, "android": true,
	}
	// 禁止再出现的「仅检测、不可安装」项
	forbidden := map[string]bool{"vim": true, "curl": true, "wget": true, "code": true}

	tools := GetTools()
	if len(tools) == 0 {
		t.Fatal("GetTools returned empty")
	}
	seen := map[string]bool{}
	for _, tool := range tools {
		if forbidden[tool.Name] {
			t.Errorf("GetTools contains forbidden name %q", tool.Name)
		}
		if !allowed[tool.Name] {
			t.Errorf("GetTools unexpected name %q (not in installer set)", tool.Name)
		}
		seen[tool.Name] = true
	}
	if !seen["vscode"] {
		t.Error("GetTools must include vscode (not code) for uninstall alignment")
	}
	if seen["code"] {
		t.Error("GetTools must not use code as Name; use vscode")
	}
}
