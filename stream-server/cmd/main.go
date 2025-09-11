package main

import (
	"context"
	"os"
	"os/signal"
	"stream-server/internal/logger"
	"stream-server/internal/server"
	"stream-server/internal/streaming"
	"time"
)

func main() {

	rm := &streaming.RoomManager{Rooms: make(map[string]*streaming.Room)}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)

	log, ctx := logger.InitLogger("panic", ctx)
	serv := server.NewServer(log, rm)

	serv.SetupServer("8000")

	serv.RegisterRoutes()

	go func() {
		if err := serv.StartServer(); err != nil {
			log.Fatal().Err(err).Msg("fail to start the server")
		}
	}()

	<-ctx.Done()

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)

	if err := serv.StopServer(ctx); err != nil {
		log.Fatal().Err(err).Msg("server force to shutdown")
	}

	stop()
	cancel()

	log.Info().Msg("server excited gracefully")

}
