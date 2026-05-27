package grpcServ

import "context"

type App struct {
	log        any
	reposipory any
	broker     any
	httpServer any
	grpcServer any
}

func New() *App {
	const op = "app.New"
	// TODO: Init repository Postgres
	// TODO: Init broker Nuts
	// TODO: Init service
	// TODO: Init HTTP server
	// TODO: Init gRPC server
	return nil
}

func (a *App) Run() {
}

func (a *App) Stop(ctx context.Context) {
}
