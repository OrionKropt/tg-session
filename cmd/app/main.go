package main

import (
	"log/slog"
	"os"
	"tg-session/configs"
	"tg-session/internal/app"
	"tg-session/pkg/logger"
)

func initialize() (cfg *configs.Config, log *slog.Logger, err error) {
	cfg = configs.NewConfig()
	cfg.ReadConfig()
	logHandler := logger.NewLogHandler(os.Stdout, logger.LogHandlerOptions{
		SlogOpts: slog.HandlerOptions{
			Level: logger.ParseLevel(cfg.LogLevel),
		},
	})
	log = slog.New(logHandler)
	return cfg, log, nil
}

func main() {
	cfg, log, err := initialize()
	if err != nil {
		panic(err)
	}
	if err := cfg.Validate(); err != nil {
		panic(err)
	}
	log.Info("Initializing tg session service")
	app.Run(*cfg, log)
}
