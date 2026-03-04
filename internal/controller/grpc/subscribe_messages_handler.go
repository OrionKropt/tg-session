package grpc

import (
	"tg-session/internal/domain"
	pb "tg-session/pkg/api/telegram/v1"
	"tg-session/pkg/logger"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) SubscribeMessages(in *pb.SubscribeMessagesRequest, stream grpc.ServerStreamingServer[pb.MessageUpdate]) error {
	ctx := stream.Context()
	sessionID := domain.SessionID(in.SessionId)
	updates, err := s.uc.Subscribe(ctx, sessionID)
	if err != nil {
		s.log.Log(logger.ERROR, "failed to subscribe", "sessionID", sessionID, "err", err.Error())
		return status.Error(codes.Internal, err.Error())
	}

	s.log.Log(logger.INFO, "gRPC stream opened", "session_id", sessionID)
	for {
		select {
		case <-ctx.Done():
			s.log.Log(logger.INFO, "gRPC stream closed", "session_id", sessionID)
			return ctx.Err()

		case msg, ok := <-updates:
			if !ok {
				return status.Error(codes.Aborted, "session closed")
			}

			err := stream.Send(&pb.MessageUpdate{
				MessageId: int64(msg.ID),
				From:      msg.From,
				Text:      msg.Text,
				Timestamp: msg.CreatedAt.Unix(),
			})

			if err != nil {
				s.log.Log(logger.ERROR, "stream send failed", "err", err.Error())
				return status.Error(codes.Unavailable, "failed to deliver message to stream")
			}
		}
	}
}
