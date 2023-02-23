package httpapp

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/Nikolay-Yakushev/mango/internal/domain/usercases"
	ports "github.com/Nikolay-Yakushev/mango/internal/ports/driver"
	cfg "github.com/Nikolay-Yakushev/mango/pkg/config"
)

type Adapter struct {
	server      *http.Server
	log         *zap.Logger
	cfg         *cfg.Config
	listener    net.Listener
	description string
	once        sync.Once
	auth        ports.Auth
}

func (a *Adapter) GetDescription() string {
	return a.description
}

func (a *Adapter) Start() error {
	var err error
	
	go func() {
		err = a.server.Serve(a.listener)
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			a.log.Sugar().Errorw("Server startup failed", "reason", err)
		}
	}()
	return err
}

func New(ctx context.Context, cfg *cfg.Config, logger *zap.Logger) (*Adapter, error) {
	listener, err := net.Listen("tcp", cfg.ServerPort)
	if err != nil {
		logger.Sugar().Errorw("Failed to initialize adapter", "error", err)
		return nil, fmt.Errorf("Adapater init failed: %w", err)
	}
	router := gin.New()

	server := &http.Server{
		Handler:      router,
		ReadTimeout:  time.Second * 10,
		WriteTimeout: time.Second * 10,
	}

	auth, err := usercases.New(ctx, logger, cfg)
	if err != nil {
		logger.Sugar().Errorw("Failed to initialize adapter", "error", err)
		return nil, fmt.Errorf("Adapater start failed: %w", err)
	}

	adap := &Adapter{
		description: "Adapter",
		server:      server,
		listener:    listener,
		log:         logger,
		auth:        auth,
		cfg:         cfg,
	}

	adap.initRoutes(router, logger)
	return adap, nil
}

func (a *Adapter) Stop(ctx context.Context) error {
	var err error

	a.once.Do(func(){
			err = a.server.Shutdown(ctx)
		})
	if err !=nil {
		a.log.Sugar().Errorw("failed to stop %s", a.GetDescription(), "Reason", err)
		err := fmt.Errorf("Failed to stop %s. Reason: %w", a.GetDescription(), err)
		return err
	}
	return nil
}
