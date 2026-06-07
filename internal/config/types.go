package config

// Config 主配置结构
type Config struct {
	Version   string     `yaml:"version"`
	Name      string     `yaml:"name"`
	Languages []Language `yaml:"languages"`
	Tools     []string   `yaml:"tools"`
	Databases []Database `yaml:"databases"`
	Shell     *Shell     `yaml:"shell,omitempty"`
}

// Language 语言配置
type Language struct {
	Name           string `yaml:"name"`
	Version        string `yaml:"version"`
	Mirror         string `yaml:"mirror,omitempty"`
	PackageManager string `yaml:"package_manager,omitempty"`
}

// Database 数据库配置
type Database struct {
	Name    string `yaml:"name"`
	Version string `yaml:"version,omitempty"`
	Docker  bool   `yaml:"docker,omitempty"` // 是否使用Docker安装
}

// Shell Shell配置
type Shell struct {
	Type    string   `yaml:"type"`    // bash, zsh, fish
	Plugins []string `yaml:"plugins"` // oh-my-zsh, starship, etc.
}

// Mirror 镜像源配置
type Mirror struct {
	Name string `yaml:"name"`
	URL  string `yaml:"url"`
	Type string `yaml:"type"` // npm, pip, go, rust, etc.
}
