package config

// LanguageVersionOption 可供 UI 选择的版本项
type LanguageVersionOption struct {
	Version string `json:"version"`
	Size    string `json:"size"`
}

// LanguageCatalogEntry 语言目录元数据（与前端展示字段对齐）
type LanguageCatalogEntry struct {
	Name              string
	DisplayName       string
	Description       string
	Icon              string
	DefaultVersion    string
	DefaultMirror     string
	MirrorOptions     []string
	AvailableVersions []LanguageVersionOption
}

// LanguageCatalog 返回 EnvKit 支持的语言列表及 UI 元数据
func LanguageCatalog() []LanguageCatalogEntry {
	return []LanguageCatalogEntry{
		{
			Name: "node", DisplayName: "Node.js", Icon: "node",
			Description:    "基于 Chrome V8 的 JavaScript 运行时",
			DefaultVersion: "20.11.1", DefaultMirror: "npmmirror",
			MirrorOptions: []string{"npmmirror", "official"},
			AvailableVersions: []LanguageVersionOption{
				{Version: "20.11.0 LTS", Size: "85 MB"},
				{Version: "18.19.0 LTS", Size: "82 MB"},
				{Version: "21.6.0", Size: "88 MB"},
			},
		},
		{
			Name: "python", DisplayName: "Python", Icon: "python",
			Description:    "解释型、动态类型、多范式语言",
			DefaultVersion: "3.12.1", DefaultMirror: "tsinghua",
			MirrorOptions: []string{"tsinghua", "aliyun", "ustc", "official"},
			AvailableVersions: []LanguageVersionOption{
				{Version: "3.12.1", Size: "65 MB"},
				{Version: "3.11.7", Size: "60 MB"},
				{Version: "3.10.13", Size: "58 MB"},
			},
		},
		{
			Name: "go", DisplayName: "Go", Icon: "go",
			Description:    "静态类型编译语言，擅长并发与系统编程",
			DefaultVersion: "1.22.0", DefaultMirror: "goproxy",
			MirrorOptions: []string{"goproxy", "aliyun", "official"},
			AvailableVersions: []LanguageVersionOption{
				{Version: "1.22.1", Size: "180 MB"},
				{Version: "1.22.0", Size: "178 MB"},
				{Version: "1.21.5", Size: "175 MB"},
			},
		},
		{
			Name: "rust", DisplayName: "Rust", Icon: "rust",
			Description:    "注重安全与并发的系统级语言",
			DefaultVersion: "stable", DefaultMirror: "ustc",
			MirrorOptions: []string{"ustc", "tsinghua", "official"},
			AvailableVersions: []LanguageVersionOption{
				{Version: "1.75.0", Size: "250 MB"},
				{Version: "1.74.1", Size: "248 MB"},
				{Version: "1.73.0", Size: "245 MB"},
			},
		},
		{
			Name: "java", DisplayName: "Java", Icon: "java",
			Description:    "广泛使用的企业级面向对象语言",
			DefaultVersion: "21", DefaultMirror: "aliyun",
			MirrorOptions: []string{"aliyun", "tsinghua", "official"},
			AvailableVersions: []LanguageVersionOption{
				{Version: "21 LTS", Size: "190 MB"},
				{Version: "17 LTS", Size: "180 MB"},
				{Version: "11 LTS", Size: "170 MB"},
			},
		},
		{
			Name: "bun", DisplayName: "Bun", Icon: "bun",
			Description:    "新一代一体化 JavaScript 运行时与工具链",
			DefaultVersion: "latest", DefaultMirror: "npmmirror",
			MirrorOptions: []string{"npmmirror", "official"},
			AvailableVersions: []LanguageVersionOption{
				{Version: "1.1.0", Size: "45 MB"},
				{Version: "1.0.30", Size: "43 MB"},
			},
		},
	}
}
