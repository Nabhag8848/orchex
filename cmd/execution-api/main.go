package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/labstack/echo/v5"
	"github.com/nabhag8848/orchex/internal/config"
	"github.com/nabhag8848/orchex/internal/db"
	"github.com/nabhag8848/orchex/internal/execution"
	"github.com/nabhag8848/orchex/internal/handler/run"
	"github.com/nabhag8848/orchex/internal/queue"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := db.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer pool.Close()

	store := db.NewStore(pool)
	sqsQueue, err := queue.New(ctx, cfg.SQS)
	if err != nil {
		log.Fatalf("SQS: %v", err)
	}
	execution.StartRelay(ctx, store, sqsQueue)

	e := execution.NewServer(execution.Deps{
		Runs: run.New(store),
	})

	sc := echo.StartConfig{Address: cfg.HTTPAddr}
	if err := sc.Start(ctx, e); err != nil {
		log.Fatalf("server: %v", err)
	}
}
