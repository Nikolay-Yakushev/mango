package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	application "github.com/Nikolay-Yakushev/mango/internal/app"
	cfg "github.com/Nikolay-Yakushev/mango/pkg/config"
	"github.com/Nikolay-Yakushev/mango/pkg/logger"
)

const appDesc string = "mango"

func main() {
	cfg := cfg.New()
	logger, err := logger.New(cfg)
	if err != nil {
		log.Fatalf("failed to initialize logger %s", err.Error())
	}

	ctx, cancel := signal.NotifyContext(
		context.Background(), syscall.SIGKILL, os.Interrupt,
		syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT,
	)
	defer cancel()

	logger.Sugar().Infof("Starting %s application", appDesc)
	app, err := application.New(logger, cfg)
	if err != nil {
		logger.Sugar().Fatalw(
			"Failed to initialize app",
			"Reason", err.Error(),
		)
	}

	logger.Sugar().Debug("Start up (app=%s) compleated", appDesc)
	app.Start(ctx)
	<-ctx.Done()

	stopCtx, stopCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer stopCancel()

	app.Stop(stopCtx)

}
