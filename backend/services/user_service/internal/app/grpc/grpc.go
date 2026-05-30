package grpcServ

import (
	"context"
	"fmt"
	"net"

	userv1 "github.com/sudo-odner/minor-shared/pkg/pb/user/v1"
	"github.com/sudo-odner/minor/backend/services/user_service/internal/config"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type Server struct {
	cfg    *config.ServerGRPC
	log    *zap.Logger
	server *grpc.Server
}

func New(cfg *config.ServerGRPC, log *zap.Logger, handler userv1.UserServiceServer) *Server {
	gRPCServer := grpc.NewServer()
	userv1.RegisterUserServiceServer(gRPCServer, handler)
	return &Server{
		cfg:    cfg,
		log:    log,
		server: gRPCServer,
	}
}

func (s *Server) Run() error {
	const op = "app.grpc.Run"
	s.log.Info("starting gRPC server", zap.String("addr", s.cfg.Address))

	l, err := net.Listen("tcp", s.cfg.Address)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	if err := s.server.Serve(l); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	s.log.Info("stopping gRPC server", zap.String("addr", s.cfg.Address))
	stopped := make(chan struct{})
	go func() {
		s.server.GracefulStop()
		close(stopped)
	}()

	select {
	case <-ctx.Done():
		s.log.Warn("graceful shutdown timed out, forcing stop")
		s.server.Stop()
		return ctx.Err()
	case <-stopped:
		return nil
	}
}
