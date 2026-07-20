package installer

import "testing"

func TestNormalizeComponentName(t *testing.T) {
	cases := map[string]string{
		"code":        "vscode",
		"Code":        "vscode",
		"vscode":      "vscode",
		"nodejs":      "node",
		"python3":     "python",
		"golang":      "go",
		"jdk":         "java",
		"conda":       "miniconda",
		"android-sdk": "android",
		"esp-idf":     "espidf",
		"git":         "git",
	}
	for in, want := range cases {
		if got := NormalizeComponentName(in); got != want {
			t.Errorf("NormalizeComponentName(%q)=%q want %q", in, got, want)
		}
	}
}
