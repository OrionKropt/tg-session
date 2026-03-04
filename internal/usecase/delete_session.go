package usecase

import (
	"context"
	"tg-session/internal/domain"
	"tg-session/pkg/logger"
)

func (c Core) DeleteSession(ctx context.Context, sessionID domain.SessionID) (err error) {
	c.log.Log(logger.INFO, "deleting session")
	session, err := c.sessionRepo.Get(ctx, sessionID)
	if err != nil {
		return err
	}
	if err := session.Close(); err != nil {
		c.log.Log(logger.ERROR, "failed to close session", "sessionID", sessionID, "err", err.Error())
		return err
	}
	c.log.Log(logger.INFO, "deleting session done")
	return c.sessionRepo.Delete(ctx, session.SessionID)
}
