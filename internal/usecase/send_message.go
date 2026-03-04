package usecase

import (
	"context"
	"fmt"
	"tg-session/internal/domain"
	"tg-session/pkg/logger"
)

func (c Core) SendMessage(ctx context.Context, sessionID domain.SessionID, msg domain.Message) (int64, error) {
	c.log.Log(logger.INFO, "sending message", "username", msg.From, "text", msg.Text)
	session, err := c.sessionRepo.Get(ctx, sessionID)
	if err != nil {
		return 0, err
	}
	id, err := session.SendText(msg)
	if err != nil {
		c.log.Log(logger.ERROR, "failed to send text", "sessionID", sessionID, "text", msg, "err", err.Error())
		return 0, fmt.Errorf("failed to send message")
	}
	c.log.Log(logger.INFO, "sending message done")
	return id, nil
}
