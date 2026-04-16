package application

import (
	"context"
	"fmt"
	"novaserver/server"
)

type Application struct {
	ctx context.Context
	srv *server.Server
}

func New(ctx context.Context, port uint) (*Application, error) {
	app := &Application{
		ctx: ctx,
	}
	var err error
	app.srv, err = server.NewServer(fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}

	return app, nil
}

func (app *Application) Run() error {
	app.srv.Run(app.ctx)
	return nil
}
