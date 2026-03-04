package grpc

import (
	"context"
	"tg-session/internal/domain"
	pb "tg-session/pkg/api/telegram/v1"
	"tg-session/pkg/logger"
)

func (s *Server) DeleteSession(ctx context.Context, in *pb.DeleteSessionRequest) (*pb.DeleteSessionResponse, error) {
	if err := s.uc.DeleteSession(ctx, domain.SessionID(in.SessionId)); err != nil {
		s.log.Log(logger.ERROR, "failed to delete session ", "sessionID", in.SessionId, "err", err.Error())
		return nil, err
	}
	return &pb.DeleteSessionResponse{}, nil
}
