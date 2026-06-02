package grpcapp

import (
	"fmt"
	"net"

	authv1 "github.com/sudo-odner/minor-shared/pkg/pb/auth/v1"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type App struct {
	log        *zap.Logger
	grpcServer *grpc.Server
	port       int
}

func New(log *zap.Logger, authService authv1.AuthServiceServer, port int) *App {
	gServer := grpc.NewServer()

	authv1.RegisterAuthServiceServer(gServer, authService)

	return &App{
		log:        log,
		grpcServer: gServer,
		port:       port,
	}
}

func (a *App) MustRun() {
	if err := a.Run(); err != nil {
		panic(err)
	}
}

func (a *App) Run() error {
	const op = "grpcapp.Run"

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", a.port))
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	a.log.Info("grpc server started", zap.String("addr", l.Addr().String()))

	if err := a.grpcServer.Serve(l); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (a *App) Stop() {
	a.log.Info("stopping grpc server", zap.Int("port", a.port))
	a.grpcServer.GracefulStop()
}
