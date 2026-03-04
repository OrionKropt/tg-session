package app

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"tg-session/configs"
	"tg-session/internal/controller/grpc"
	"tg-session/internal/infrastructure/repository/in_memory_storage"
	tg "tg-session/internal/infrastructure/telegram"
	"tg-session/internal/usecase"
	"time"
)

func Run(cfg configs.Config, log *slog.Logger) {
	sessionRepo := in_memory_storage.NewSessionRepository()
	clientFactory := tg.NewTGClientFactory(cfg.TelegramAPPID, cfg.TelegramAPPHash)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	core := usecase.New(log, cfg, &sessionRepo, &clientFactory)
	server := grpc.NewServer(log, core)
	server.Init("0.0.0.0:50051")
	if err := server.RunGRPC(ctx); err != nil {
		log.Error("failed to start server", "err", err.Error())
	}
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer shutdownCancel()
	_ = core.LogoutAll(shutdownCtx)
}
