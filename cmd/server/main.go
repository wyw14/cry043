package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/wyw14/cry043/internal/application"
	"github.com/wyw14/cry043/internal/config"
	"github.com/wyw14/cry043/internal/middleware"
	"github.com/wyw14/cry043/internal/repository"
	"github.com/wyw14/cry043/internal/service"
	httptransport "github.com/wyw14/cry043/internal/transport/http"
	"go.uber.org/zap"
)

func main() {
	configuration := config.Load()
	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()

	lifetime, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	pool, err := pgxpool.New(lifetime, configuration.DatabaseURL)
	if err != nil {
		logger.Fatal("open compliance database", zap.Error(err))
	}
	defer pool.Close()

	store := repository.NewPostgres(pool)
	clock := service.Clock{}
	identifiers := &service.IDs{}
	services := httptransport.Services{
		Compliance:  application.NewComplianceService(store, clock, identifiers, &service.LocalScheduler{}),
		Inspections: application.NewInspectionService(store, clock, identifiers),
		Risks:       application.NewRiskService(store, clock),
		Readiness: func() error {
			probe, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			return pool.Ping(probe)
		},
	}
	engine := httptransport.New(services, middleware.RequestBoundary(configuration.Timeout), middleware.RecoverToJSON(logger))
	server := &http.Server{Addr: configuration.Addr, Handler: engine, ReadHeaderTimeout: 5 * time.Second, IdleTimeout: 45 * time.Second}

	go func() {
		logger.Info("compliance control room listening", zap.String("address", configuration.Addr))
		if serveErr := server.ListenAndServe(); serveErr != nil && !errors.Is(serveErr, http.ErrServerClosed) {
			logger.Fatal("serve compliance API", zap.Error(serveErr))
		}
	}()
	<-lifetime.Done()
	shutdown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdown); err != nil {
		logger.Error("graceful shutdown", zap.Error(err))
	}
}
