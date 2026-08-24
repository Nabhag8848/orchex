package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/nabhag8848/orchex/internal/config"
	"github.com/nabhag8848/orchex/internal/db"
	"github.com/nabhag8848/orchex/internal/queue"
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

	if cfg.FunctionSandboxARN == "" || cfg.AWSEndpointURL != "" {
		log.Printf("sandbox: skip invoke (local)")
	} else {
		sb, err := sandbox.New(ctx, cfg.FunctionSandboxARN, cfg.AWSRegion, cfg.AWSEndpointURL)
		if err != nil {
			log.Printf("sandbox: client: %v", err)
		} else {
			pctx, cancel := context.WithTimeout(ctx, 15*time.Second)
			out, err := sb.Ping(pctx)
			cancel()
			if err != nil {
				log.Printf("sandbox: invoke failed: %v", err)
			} else {
				log.Printf("sandbox: invoke ok payload=%s", out)
			}
		}
	}

	if cfg.SQSQueueURL == "" {
		log.Printf("sqs: SQS_QUEUE_URL unset; skip polling (local)")
	} else {
		q, err := queue.New(ctx, cfg.SQSQueueURL, cfg.AWSRegion, cfg.AWSEndpointURL)
		if err != nil {
			log.Fatalf("sqs: %v", err)
		}
		go poll(ctx, q)
	}

	e := worker.NewServer()
	sc := echo.StartConfig{Address: cfg.HTTPAddr}
	if err := sc.Start(ctx, e); err != nil {
		log.Fatalf("server: %v", err)
	}
}

func poll(ctx context.Context, q *queue.Client) {
	log.Printf("sqs: polling")
	for {
		if err := ctx.Err(); err != nil {
			return
		}

		msgs, err := q.Receive(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			log.Printf("sqs: receive failed: %v", err)
			select {
			case <-ctx.Done():
				return
			case <-time.After(2 * time.Second):
			}
			continue
		}

		for _, m := range msgs {
			log.Printf("sqs: received message_id=%s body=%s", m.MessageID, m.Body)
			if err := q.Delete(ctx, m.ReceiptHandle); err != nil {
				log.Printf("sqs: delete failed message_id=%s: %v", m.MessageID, err)
				continue
			}
			log.Printf("sqs: deleted message_id=%s", m.MessageID)
		}
	}
}
