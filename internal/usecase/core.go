package usecase

import (
	"context"
	"log/slog"
	"tg-session/configs"
	"tg-session/internal/domain"
	"tg-session/pkg/logger"
)

type SessionRepository interface {
	Delete(context.Context, domain.SessionID) error
	Save(context.Context, domain.Session) error
	Update(context.Context, domain.Session) error
	Get(context.Context, domain.SessionID) (domain.Session, error)
	GetAll(context.Context) ([]domain.Session, error)
}

type TGClientFactory interface {
	CreateTGClient(string, *slog.Logger) (domain.TGClient, error)
}

type Core struct {
	sessionRepo   SessionRepository
	clientFactory TGClientFactory
	log           logger.Logger
	cfg           configs.Config
}

func New(log *slog.Logger, cfg configs.Config, sr SessionRepository, cf TGClientFactory) *Core {
	return &Core{
		log:         logger.Logger{Inst: log, Name: "Core"},
		cfg:         cfg,
		sessionRepo: sr, clientFactory: cf}
}
