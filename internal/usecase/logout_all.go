package usecase

import (
	"context"
	"tg-session/pkg/logger"
)

func (c Core) LogoutAll(ctx context.Context) error {
	c.log.Log(logger.INFO, "logging out all active sessions")

	sessions, err := c.sessionRepo.GetAll(ctx)
	if err != nil {
		c.log.Log(logger.ERROR, "failed to get all sessions", "err", err.Error())
		return err
	}

	for _, sess := range sessions {
		if err := sess.Close(); err != nil {
			c.log.Log(logger.ERROR, "failed to logout session", "id", sess.SessionID, "err", err.Error())
		}
	}
	return nil
}
