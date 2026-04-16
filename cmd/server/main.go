package main

import (
	"context"
	"novaserver/application"
	"os"
	"os/signal"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	app, err := application.New(ctx, 7777)
	if err != nil {
		panic(err)
	}
	err = app.Run()
	if err != nil {
		panic(err)
	}
	<-ctx.Done()
}
