package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/labstack/echo/v5"
	"github.com/nabhag8848/orchex/internal/config"
	"github.com/nabhag8848/orchex/internal/db"
	"github.com/nabhag8848/orchex/internal/sandbox"
	"github.com/nabhag8848/orchex/internal/worker"
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

	go func() {
		sandboxClient, err := sandbox.NewLambda(ctx, cfg.Lambda)
		if err != nil {
			log.Printf("sandbox: %v", err)
			return
		}
		result, err := sandboxClient.Invoke(ctx, "return { ping: true };", json.RawMessage(`{"data":{}}`), 5000)
		if err != nil {
			log.Printf("sandbox: startup invoke: %v", err)
			return
		}
		log.Printf("sandbox: startup invoke succeeded: %s", result)
	}()

	e := worker.NewServer()
	sc := echo.StartConfig{Address: cfg.HTTPAddr}
	if err := sc.Start(ctx, e); err != nil {
		log.Fatalf("server: %v", err)
	}
}
