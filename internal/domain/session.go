package domain

import (
	"context"
	"tg-session/internal/infrastructure/uuid"
	"time"
)

type TGClient interface {
	Start() (err error)
	Stop(ctx context.Context) (err error)
	SendText(context.Context, string, string) (int64, error)
	AuthQRCode(ctx context.Context) (string, error)
	Updates() <-chan Message
}

type SessionID string

type Session struct {
	SessionID SessionID
	client    TGClient
	ctx       context.Context
	cancel    context.CancelFunc
}

func NewSession(ctx context.Context, cancel context.CancelFunc) Session {
	return Session{
		SessionID: SessionID(uuid.Generate().String()),
		ctx:       ctx,
		cancel:    cancel,
	}
}

func (s *Session) WithClient(client TGClient) *Session {
	s.client = client
	return s
}

func (s *Session) Close() error {
	s.cancel()
	stopCtx, stopCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer stopCancel()
	return s.client.Stop(stopCtx)
}

func (s *Session) Start() {
	go func() {
		if err := s.client.Start(); err != nil {
			return
		}
	}()
}

func (s *Session) Updates() <-chan Message {
	return s.client.Updates()
}

func (s *Session) SendText(msg Message) (int64, error) {
	return s.client.SendText(s.ctx, msg.From, msg.Text)
}
