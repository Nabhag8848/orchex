package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/labstack/echo/v5"
	"github.com/nabhag8848/orchex/internal/builder"
	"github.com/nabhag8848/orchex/internal/config"
	"github.com/nabhag8848/orchex/internal/db"
	"github.com/nabhag8848/orchex/internal/handler/workflow"
	"github.com/nabhag8848/orchex/internal/logger"
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

	e := builder.NewServer(builder.Deps{
		Workflows: workflow.New(db.NewStore(pool)),
	})

	slog.Info("builder API starting", "address", cfg.HTTPAddr)
	sc := echo.StartConfig{Address: cfg.HTTPAddr}
	if err := sc.Start(ctx, e); err != nil {
		logger.Fatal("start server", "error", err)
	}
}
