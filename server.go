package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/tirathawat/assessment/di"
	"github.com/tirathawat/assessment/logs"
)

func main() {
	server, cleanup, err := di.InitializeApplication()
	defer cleanup()
	if err != nil {
		logs.Error().Err(err).Msg("Cannot initialize application")
		panic(err)
	}

	server.Run()
	logs.Info().Msgf("Application started at port %s", server.Port())

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	if err := server.Shutdown(); err != nil {
		logs.Error().Err(err).Msg("Cannot shutdown application")
	}
}
