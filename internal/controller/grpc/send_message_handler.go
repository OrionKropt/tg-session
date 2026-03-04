package grpc

import (
	"context"
	"tg-session/internal/domain"
	pb "tg-session/pkg/api/telegram/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) SendMessage(ctx context.Context, in *pb.SendMessageRequest) (*pb.SendMessageResponse, error) {
	msg := domain.Message{
		From: in.Peer,
		Text: in.Text,
	}
	id, err := s.uc.SendMessage(ctx, domain.SessionID(in.SessionId), msg)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.SendMessageResponse{
		MessageId: id,
	}, nil
}
