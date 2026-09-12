package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/labstack/echo/v5"
	"github.com/nabhag8848/orchex/internal/config"
	"github.com/nabhag8848/orchex/internal/db"
	"github.com/nabhag8848/orchex/internal/execution"
	"github.com/nabhag8848/orchex/internal/handler/run"
	"github.com/nabhag8848/orchex/internal/logger"
	"github.com/nabhag8848/orchex/internal/queue"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		logger.Configure("info")
		logger.Fatal("load config", "error", err)
	}
	logger.Configure(cfg.LogLevel)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := db.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Fatal("connect database", "error", err)
	}
	defer pool.Close()

	store := db.NewStore(pool)
	sqsQueue, err := queue.New(ctx, cfg.SQS)
	if err != nil {
		logger.Fatal("create SQS client", "error", err)
	}
	execution.StartRelay(ctx, store, sqsQueue)

	e := execution.NewServer(execution.Deps{
		Runs: run.New(store),
	})

	slog.Info("execution API starting", "address", cfg.HTTPAddr)
	sc := echo.StartConfig{Address: cfg.HTTPAddr}
	if err := sc.Start(ctx, e); err != nil {
		logger.Fatal("start server", "error", err)
	}
}
