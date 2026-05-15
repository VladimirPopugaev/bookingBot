package domain

type Config struct {
	LogLevel   string         `yaml:"logLevel"`
	Telegram   TelegramConfig `yaml:"telegram"`
	SiteConfig SiteConfig     `yaml:"site"`
	HTTP       HTTPConfig     `yaml:"http"`
}

type TelegramConfig struct {
	BotToken string `yaml:"bot_token"`
	BaseURL  string `yaml:"base_url"`
	Timeout  int    `yaml:"timeout"` // in seconds
}

type SiteConfig struct {
	RequestTimeout     int    `yaml:"request_timeout"`     // in seconds
	MonitoringInterval int    `yaml:"monitoring_interval"` // in seconds
}

type HTTPConfig struct {
	Host       string `yaml:"host"`
	TLSEnabled bool   `yaml:"tls_enabled"`
	CertFile   string `yaml:"cert_file"`
	KeyFile    string `yaml:"key_file"`
}
