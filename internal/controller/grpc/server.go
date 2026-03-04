package grpc

import (
	"context"
	"log/slog"
	"net"
	"tg-session/internal/domain"
	pb "tg-session/pkg/api/telegram/v1"
	"tg-session/pkg/logger"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type UseCase interface {
	CreateSession(context.Context) (domain.Session, string, error)
	DeleteSession(context.Context, domain.SessionID) error
	SendMessage(context.Context, domain.SessionID, domain.Message) (int64, error)
	Subscribe(context.Context, domain.SessionID) (<-chan domain.Message, error)
}

type Server struct {
	pb.UnimplementedTGSessionServer
	uc     UseCase
	server *grpc.Server
	log    *logger.Logger
	addr   string
}

func NewServer(inst *slog.Logger, uc UseCase) *Server {
	return &Server{uc: uc, log: &logger.Logger{Inst: inst, Name: "gRPC Server"}}
}

func InterceptorLogger(l *slog.Logger) logging.Logger {
	return logging.LoggerFunc(func(ctx context.Context, lvl logging.Level, msg string, fields ...any) {
		l.Log(ctx, slog.Level(lvl), msg, fields...)
	})
}

func (s *Server) Init(addr string) {
	s.addr = addr
	logOpts := []logging.Option{
		logging.WithLogOnEvents(logging.StartCall, logging.FinishCall),
	}

	opts := []grpc.ServerOption{grpc.ChainUnaryInterceptor(
		logging.UnaryServerInterceptor(InterceptorLogger(s.log.Inst), logOpts...))}

	s.server = grpc.NewServer(opts...)
	pb.RegisterTGSessionServer(s.server, s)

	reflection.Register(s.server)
}

func (s *Server) RunGRPC(ctx context.Context) error {
	lis, err := net.Listen("tcp", s.addr)
	if err != nil {
		s.log.Log(logger.ERROR, "failed to listen", "err", err.Error())
		return err
	}

	go func() {
		<-ctx.Done()
		s.Stop()
	}()

	s.log.Log(logger.INFO, "server listening at "+s.addr)
	if err := s.server.Serve(lis); err != nil {
		s.log.Log(logger.ERROR, "failed to serve", "err", err.Error())
		return err
	}
	return nil
}

func (s *Server) Stop() {
	s.log.Log(logger.INFO, "stopping gRPC server...")
	s.server.GracefulStop()
}
