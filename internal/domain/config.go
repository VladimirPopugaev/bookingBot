package domain

type Config struct {
	LogLevel string         `yaml:"logLevel"`
	Telegram TelegramConfig `yaml:"telegram"`
	SiteConfig SiteConfig       `yaml:"site"`
}

type TelegramConfig struct {
	BotToken string `yaml:"bot_token"`
	BaseURL  string `yaml:"base_url"`
	Timeout  int    `yaml:"timeout"` // in seconds
}

type SiteConfig struct {
	TargetURL      string `yaml:"target_url"`
	RequestTimeout int `yaml:"request_timeout"` // in seconds
	MonitoringInterval int `yaml:"monitoring_interval"` // in seconds
}
