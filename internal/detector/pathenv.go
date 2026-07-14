package detector

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

var pathInitOnce sync.Once

// InitDetectionEnvironment 合并登录 shell 的 PATH 与常见工具安装目录，供 GUI/CLI 检测使用。
func InitDetectionEnvironment() {
	pathInitOnce.Do(func() {
		merged := mergePathLists(
			loginShellPATH(),
			os.Getenv("PATH"),
			strings.Join(toolSearchDirs(), string(os.PathListSeparator)),
		)
		if merged != "" {
			os.Setenv("PATH", merged)
		}
	})
}

func loginShellPATH() string {
	if runtime.GOOS == "windows" {
		return ""
	}

	shell := os.Getenv("SHELL")
	if shell == "" {
		if runtime.GOOS == "darwin" {
			shell = "/bin/zsh"
		} else {
			shell = "/bin/bash"
		}
	}

	cmd := exec.Command(shell, "-ilc", `printf %s "$PATH"`)
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func toolSearchDirs() []string {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil
	}

	dirs := []string{
		filepath.Join(home, ".bun", "bin"),
		filepath.Join(home, ".cargo", "bin"),
		filepath.Join(home, ".local", "bin"),
		filepath.Join(home, ".local", "go", "bin"),
		filepath.Join(home, "miniconda3", "bin"),
		filepath.Join(home, "miniconda3", "condabin"),
		filepath.Join(home, ".fnm"),
		filepath.Join(home, ".local", "share", "fnm"),
		filepath.Join(home, "Library", "Application Support", "fnm"),
		filepath.Join(home, ".sdkman", "candidates", "java", "current", "bin"),
		filepath.Join(home, ".pyenv", "shims"),
		"/usr/local/go/bin",
		"/opt/homebrew/bin",
		"/usr/local/bin",
		"/usr/local/Homebrew/bin",
		filepath.Join(home, "Library", "Android", "sdk", "platform-tools"),
		filepath.Join(home, "Library", "Android", "sdk", "cmdline-tools", "latest", "bin"),
		filepath.Join(home, "Android", "Sdk", "platform-tools"),
		filepath.Join(home, "Android", "Sdk", "cmdline-tools", "latest", "bin"),
		filepath.Join(home, "Android", "platform-tools"),
	}

	if runtime.GOOS == "windows" {
		localAppData := os.Getenv("LOCALAPPDATA")
		if localAppData == "" {
			localAppData = filepath.Join(home, "AppData", "Local")
		}
		userProfile := os.Getenv("USERPROFILE")
		if userProfile == "" {
			userProfile = home
		}
		programFiles := os.Getenv("ProgramFiles")
		if programFiles == "" {
			programFiles = `C:\Program Files`
		}

		dirs = append(dirs,
			filepath.Join(localAppData, "Android", "Sdk", "platform-tools"),
			filepath.Join(localAppData, "Android", "Sdk", "cmdline-tools", "latest", "bin"),
			filepath.Join(userProfile, "miniconda3", "Scripts"),
			filepath.Join(userProfile, "miniconda3", "condabin"),
			filepath.Join(home, "miniconda3", "Scripts"),
			filepath.Join(programFiles, "Go", "bin"),
			filepath.Join(localAppData, "Programs", "Python"),
		)

		// 官方 Python 安装器: %LOCALAPPDATA%\Programs\Python\Python3x{,\Scripts}
		pythonRoot := filepath.Join(localAppData, "Programs", "Python")
		if entries, err := os.ReadDir(pythonRoot); err == nil {
			for _, e := range entries {
				if e.IsDir() {
					base := filepath.Join(pythonRoot, e.Name())
					dirs = append(dirs, base, filepath.Join(base, "Scripts"))
				}
			}
		}
	}

	nvmRoot := filepath.Join(home, ".nvm", "versions", "node")
	if entries, err := os.ReadDir(nvmRoot); err == nil {
		for _, e := range entries {
			if e.IsDir() {
				dirs = append(dirs, filepath.Join(nvmRoot, e.Name(), "bin"))
			}
		}
	}

	return dirs
}

func mergePathLists(parts ...string) string {
	seen := make(map[string]struct{})
	var out []string
	for _, p := range parts {
		for _, dir := range filepath.SplitList(p) {
			dir = strings.TrimSpace(dir)
			if dir == "" {
				continue
			}
			if _, ok := seen[dir]; ok {
				continue
			}
			seen[dir] = struct{}{}
			out = append(out, dir)
		}
	}
	return strings.Join(out, string(os.PathListSeparator))
}
