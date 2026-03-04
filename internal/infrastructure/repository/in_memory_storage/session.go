package in_memory_storage

import (
	"context"
	"fmt"
	"sync"
	"tg-session/internal/domain"
)

type Session struct {
	mu       sync.RWMutex
	sessions map[domain.SessionID]domain.Session
}

func NewSessionRepository() Session {
	return Session{
		sessions: make(map[domain.SessionID]domain.Session, 1024),
	}
}

func (s *Session) Save(ctx context.Context, session domain.Session) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, ok := s.sessions[session.SessionID]
	if ok {
		return fmt.Errorf("session already exists: %s", session.SessionID)
	}
	s.sessions[session.SessionID] = session
	return nil
}

func (s *Session) Update(ctx context.Context, session domain.Session) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	session, ok := s.sessions[session.SessionID]
	if !ok {
		return fmt.Errorf("session not found: %s", session.SessionID)
	}
	s.sessions[session.SessionID] = session
	return nil
}

func (s *Session) Delete(ctx context.Context, id domain.SessionID) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, ok := s.sessions[id]
	if !ok {
		return fmt.Errorf("session not found: %s", id)
	}
	delete(s.sessions, id)
	return nil
}

func (s *Session) Get(ctx context.Context, id domain.SessionID) (domain.Session, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	session, ok := s.sessions[id]
	if !ok {
		return domain.Session{}, fmt.Errorf("session not found: %s", id)
	}
	return session, nil
}

func (s *Session) GetAll(ctx context.Context) ([]domain.Session, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	sessions := make([]domain.Session, 0, len(s.sessions))
	for _, session := range s.sessions {
		sessions = append(sessions, session)
	}
	return sessions, nil
}
