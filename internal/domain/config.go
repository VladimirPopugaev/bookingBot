package domain

type Config struct {
	LogLevel   string         `yaml:"logLevel"`
	Database   DatabaseConfig `yaml:"postgres"`
	Telegram   TelegramConfig `yaml:"telegram"`
	SiteConfig SiteConfig     `yaml:"site"`
	HTTP       HTTPConfig     `yaml:"http"`
}

type DatabaseConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Name     string `yaml:"name"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	SSLMode  string `yaml:"ssl_mode"`
}

type TelegramConfig struct {
	BotToken string `yaml:"bot_token"`
	BaseURL  string `yaml:"base_url"`
	Timeout  int    `yaml:"timeout"` // in seconds
}

type SiteConfig struct {
	RequestTimeout     int `yaml:"request_timeout"`     // in seconds
	MonitoringInterval int `yaml:"monitoring_interval"` // in seconds
}

type HTTPConfig struct {
	Host       string `yaml:"host"`
	TLSEnabled bool   `yaml:"tls_enabled"`
	CertFile   string `yaml:"cert_file"`
	KeyFile    string `yaml:"key_file"`
}
