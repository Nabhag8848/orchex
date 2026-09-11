package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/labstack/echo/v5"
	"github.com/nabhag8848/orchex/internal/config"
	"github.com/nabhag8848/orchex/internal/db"
	sqlcdb "github.com/nabhag8848/orchex/internal/db/sqlc"
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
	store := db.NewStore(pool)

	sqsQueue, err := queue.New(ctx, cfg.SQS)
	if err != nil {
		log.Fatalf("SQS: %v", err)
	}

	go func() {
		for {
			messages, err := sqsQueue.Receive(ctx)
			if err != nil {
				if ctx.Err() != nil {
					return
				}
				log.Printf("worker: receive node jobs: %v", err)
				continue
			}

			for _, message := range messages {
				go handleNodeJob(ctx, store, sqsQueue, message)
			}
		}
	}()

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

func handleNodeJob(ctx context.Context, store *db.Store, sqsQueue *queue.SQS, message types.Message) {
	if ctx.Err() != nil {
		return
	}

	var job queue.NodeJobMessage
	if err := json.Unmarshal([]byte(*message.Body), &job); err != nil {
		log.Printf("worker: decode node job: %v", err)
		return
	}

	workflowRun, err := store.GetWorkflowRun(ctx, job.RunID)
	if err != nil {
		log.Printf("worker: get run %s: %v", job.RunID, err)
		return
	}
	if workflowRun.Status == sqlcdb.WorkflowRunStatusPending || workflowRun.Status == sqlcdb.WorkflowRunStatusRunning {
		log.Printf("worker: execute node job: %s", *message.Body)
		deleteNodeJob(ctx, sqsQueue, message)
	} else {
		deleteNodeJob(ctx, sqsQueue, message)
	}
}

func deleteNodeJob(ctx context.Context, sqsQueue *queue.SQS, message types.Message) {
	if err := sqsQueue.Delete(ctx, *message.ReceiptHandle); err != nil {
		log.Printf("worker: delete node job: %v", err)
	}
}
