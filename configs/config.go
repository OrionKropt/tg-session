package configs

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	TelegramAPPID   int
	TelegramAPPHash string
	PeerDBName      string
	LogLevel        string
}

func NewConfig() *Config {
	return &Config{LogLevel: "info"}
}

func (cfg *Config) ReadConfig() {
	appID, err := strconv.Atoi(os.Getenv("TELEGRAM_APP_ID"))
	if err == nil {
		cfg.TelegramAPPID = appID
	}
	cfg.TelegramAPPHash = os.Getenv("TELEGRAM_APP_HASH")
	cfg.PeerDBName = os.Getenv("TELEGRAM_PEER_DATABASE_NAME")
	cfg.LogLevel = os.Getenv("LOG_LEVEL")
}

func (cfg *Config) Validate() error {
	if cfg.TelegramAPPID == 0 {
		return fmt.Errorf("TELEGRAM_APP_ID must be set")
	}
	if cfg.TelegramAPPHash == "" {
		return fmt.Errorf("TELEGRAM_APP_HASH must be set")
	}
	if cfg.PeerDBName == "" {
		return fmt.Errorf("TELEGRAM_PEER_DATABASE_NAME must be set")
	}

	return nil
}
