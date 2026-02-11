package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"source.hodakov.me/hdkv/faketunes/internal/application"
	"source.hodakov.me/hdkv/faketunes/internal/domains"
	"source.hodakov.me/hdkv/faketunes/internal/domains/cacher"
	"source.hodakov.me/hdkv/faketunes/internal/domains/filesystem"
	"source.hodakov.me/hdkv/faketunes/internal/domains/transcoder"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	app := application.New(ctx)

	app.Logger().Info("Starting faketunes...")

	err := app.InitConfig()
	if err != nil {
		app.Logger().Fatal(err)
	}

	app.InitLogger()

	app.RegisterDomain(domains.FilesystemName, filesystem.New(app))
	app.RegisterDomain(domains.CacherName, cacher.New(app))
	app.RegisterDomain(domains.TranscoderName, transcoder.New(app))

	err = app.ConnectDependencies()
	if err != nil {
		app.Logger().Fatal(err)
	}

	err = app.StartDomains()
	if err != nil {
		app.Logger().Fatal(err)
	}

	app.Logger().Info("Started faketunes")

	// CTRL+C handler.
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt)

	shutdownDone := make(chan bool, 1)

	go func() {
		signalThing := <-interrupt
		if signalThing == syscall.SIGTERM || signalThing == syscall.SIGINT {
			app.Logger().WithField("signal", signalThing.String()).
				Info("Got terminating signal, shutting down...")

			cancel()

			shutdownDone <- true
		}
	}()

	<-shutdownDone
	os.Exit(0)
}
