package usecase

import (
	"context"
	"tg-session/internal/domain"
)

func (c Core) Subscribe(ctx context.Context, sessionID domain.SessionID) (<-chan domain.Message, error) {
	session, err := c.sessionRepo.Get(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	return session.Updates(), nil
}
