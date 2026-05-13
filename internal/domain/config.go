package domain

type Config struct {
	LogLevel string         `yaml:"logLevel"`
	Telegram TelegramConfig `yaml:"telegram"`
}

type TelegramConfig struct {
	BotToken string `yaml:"bot_token"`
	BaseURL  string `yaml:"base_url"`
	Timeout  int    `yaml:"timeout"` // in seconds
}
