package detector

import (
	"runtime"
)

type OSType string

const (
	OSWindows OSType = "windows"
	OSMacOS   OSType = "darwin"
	OSLinux   OSType = "linux"
	OSUnknown OSType = "unknown"
)

type Architecture string

const (
	ArchAMD64   Architecture = "amd64"
	ArchARM64   Architecture = "arm64"
	Arch386     Architecture = "386"
	ArchUnknown Architecture = "unknown"
)

type SystemInfo struct {
	OS           OSType
	Architecture Architecture
	Distribution string // For Linux: ubuntu, debian, fedora, arch, etc.
}

// DetectOS 检测当前操作系统
func DetectOS() OSType {
	switch runtime.GOOS {
	case "windows":
		return OSWindows
	case "darwin":
		return OSMacOS
	case "linux":
		return OSLinux
	default:
		return OSUnknown
	}
}

// DetectArchitecture 检测系统架构
func DetectArchitecture() Architecture {
	switch runtime.GOARCH {
	case "amd64":
		return ArchAMD64
	case "arm64":
		return ArchARM64
	case "386":
		return Arch386
	default:
		return ArchUnknown
	}
}

// DetectSystem 获取完整的系统信息
func DetectSystem() *SystemInfo {
	info := &SystemInfo{
		OS:           DetectOS(),
		Architecture: DetectArchitecture(),
	}

	// 对于Linux系统，尝试检测发行版
	if info.OS == OSLinux {
		info.Distribution = detectLinuxDistribution()
	}

	return info
}

// IsWindows 判断是否为Windows系统
func (s *SystemInfo) IsWindows() bool {
	return s.OS == OSWindows
}

// IsMacOS 判断是否为macOS系统
func (s *SystemInfo) IsMacOS() bool {
	return s.OS == OSMacOS
}

// IsLinux 判断是否为Linux系统
func (s *SystemInfo) IsLinux() bool {
	return s.OS == OSLinux
}

// IsARM 判断是否为ARM架构
func (s *SystemInfo) IsARM() bool {
	return s.Architecture == ArchARM64
}
