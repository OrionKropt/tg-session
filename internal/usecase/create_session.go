package usecase

import (
	"context"
	"path/filepath"
	"tg-session/internal/domain"
	"tg-session/pkg/logger"
)

func (c Core) CreateSession(ctx context.Context) (session domain.Session, qrURL string, err error) {
	c.log.Log(logger.INFO, "creating session")

	sessionCtx, cancel := context.WithCancel(context.Background())
	session = domain.NewSession(sessionCtx, cancel)

	dbPath := filepath.Join(c.cfg.PeerDBName, string(session.SessionID))

	client, err := c.clientFactory.CreateTGClient(dbPath, c.log.Inst)
	if err != nil {
		return domain.Session{}, "", err
	}
	if err := client.Start(); err != nil {
		c.log.Log(logger.ERROR, "failed to start telegram client", "err", err.Error())
		return domain.Session{}, "", err
	}

	_ = session.WithClient(client)

	qrURL, err = client.AuthQRCode(sessionCtx)
	if err != nil {
		c.log.Log(logger.ERROR, "failed to auth via QR code", "error", err.Error())
		_ = session.Close()
		return domain.Session{}, "", err
	}

	err = c.sessionRepo.Save(ctx, session)
	if err != nil {
		c.log.Log(logger.ERROR, "failed to create session", "err", err.Error())
		_ = session.Close()
		return domain.Session{}, "", err
	}
	c.log.Log(logger.INFO, "creating session done")
	return session, qrURL, err
}
