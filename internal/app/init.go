package application

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	cfg "github.com/Nikolay-Yakushev/mango/pkg/config"
	httpapp "github.com/Nikolay-Yakushev/mango/internal/adapters/http"
)

type Closeable interface {
	Stop(ctx context.Context)error
	GetDescription()string
}

type App struct {
	log         *zap.Logger
	description string
	cfg         *cfg.Config
	closeables  []Closeable
}

func (a *App) GetDescription() string {
	return a.description
}

func New(logger *zap.Logger, cfg *cfg.Config) (*App, error) {
	var closeables []Closeable
	namedLogger := logger.Named("mango")
	a := &App{
		log: namedLogger, 
		description: "Mango",
		cfg: cfg,
		closeables: closeables,
	}
	return a, nil
}

func (a *App) Start(ctx context.Context) error {
	webapp, err := httpapp.New(ctx, a.cfg, a.log)
	if err != nil {
		a.log.Sugar().Errorw("Failed to start webapp", "reason", err)
		err = fmt.Errorf("Failed to start webapp. Reason %w", err)
		return err
	}
	a.closeables = append(a.closeables, webapp)
	webapp.Start()
	return nil
}

type CloseResult struct {
	Err    error
	Entity Closeable
}

func (a *App) Stop(ctx context.Context) error {
	defer a.log.Sync()

	for idx :=range a.closeables{
		closeable := a.closeables[idx]
		errCh := make(chan CloseResult, 1)

		go func(entity Closeable){
			err := entity.Stop(ctx)
			errCh <-CloseResult{Err: err, Entity: entity}

		}(closeable)

		select {

			case <-ctx.Done():
				return ctx.Err()

			case closeRes := <-errCh:
				if closeRes.Err != nil{
					a.log.Sugar().Errorf("Failed to stop %s", closeRes.Entity.GetDescription())
					continue
				}
				a.log.Sugar().Infof("Successefully stopped `%s`", closeRes.Entity.GetDescription())
		}
		a.log.Sugar().Info("Successefully stopped components")
	}
	return nil
}
