package grpc

import (
	"context"
	pb "tg-session/pkg/api/telegram/v1"
	"tg-session/pkg/logger"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) CreateSession(ctx context.Context, in *pb.CreateSessionRequest) (*pb.CreateSessionResponse, error) {
	session, qrURL, err := s.uc.CreateSession(ctx)
	if err != nil {
		s.log.Log(logger.ERROR, "failed to create session", "err", err.Error())
		return nil, status.Errorf(codes.Internal, "could not initialize session: %v", err)
	}
	return &pb.CreateSessionResponse{
		SessionId: string(session.SessionID),
		QrCode:    qrURL,
	}, nil
}
