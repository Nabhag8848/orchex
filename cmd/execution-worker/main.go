package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
	"github.com/nabhag8848/orchex/internal/config"
	"github.com/nabhag8848/orchex/internal/db"
	sqlcdb "github.com/nabhag8848/orchex/internal/db/sqlc"
	"github.com/nabhag8848/orchex/internal/executor"
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

	sandboxClient, err := sandbox.NewLambda(ctx, cfg.Lambda)
	if err != nil {
		log.Fatalf("sandbox: %v", err)
	}

	sqsQueue, err := queue.New(ctx, cfg.SQS)
	if err != nil {
		log.Fatalf("SQS: %v", err)
	}
	deps := workerDeps{
		store:    store,
		sqsQueue: sqsQueue,
		registry: executor.NewRegistry(
			executor.NewStartExecutor(),
			executor.NewConditionalExecutor(),
			executor.NewAPIExecutor(),
			executor.NewFunctionExecutor(sandboxClient),
			executor.NewResponseExecutor(),
		),
	}

	go func() {
		for {
			messages, err := deps.sqsQueue.Receive(ctx)
			if err != nil {
				if ctx.Err() != nil {
					return
				}
				log.Printf("worker: receive node jobs: %v", err)
				continue
			}

			for _, message := range messages {
				go handleNodeJob(ctx, deps, message)
			}
		}
	}()

	go func() {
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

type workerDeps struct {
	store    *db.Store
	sqsQueue *queue.SQS
	registry *executor.Registry
}

func handleNodeJob(ctx context.Context, deps workerDeps, message types.Message) {
	if ctx.Err() != nil {
		return
	}

	var job queue.NodeJobMessage
	if err := json.Unmarshal([]byte(*message.Body), &job); err != nil {
		log.Printf("worker: decode node job: %v", err)
		return
	}

	workflowRun, err := deps.store.GetWorkflowRun(ctx, job.RunID)
	if err != nil {
		log.Printf("worker: get run %s: %v", job.RunID, err)
		return
	}
	if workflowRun.Status == sqlcdb.WorkflowRunStatusPending || workflowRun.Status == sqlcdb.WorkflowRunStatusRunning {
		nodeRow, err := deps.store.GetNodeForExecution(ctx, sqlcdb.GetNodeForExecutionParams{
			WorkflowVersionID: workflowRun.WorkflowVersionID,
			NodeID:            workflowRun.CurrentNodeID,
		})
		if err != nil {
			log.Printf("worker: get current node for run %s: %v", workflowRun.ID, err)
			return
		}

		result, err := deps.registry.Execute(ctx, executor.Node{
			WorkflowVersionID: nodeRow.WorkflowVersionID,
			ID:                nodeRow.ID,
			Name:              nodeRow.Name,
			Type:              executor.NodeType(nodeRow.Type),
			Config:            nodeRow.Config,
		}, workflowRun.LastOutput)
		if err != nil {
			log.Printf("worker: execute node for run %s: %v", workflowRun.ID, err)
			return
		}

		if err := persistNodeResult(ctx, deps.store, workflowRun, nodeRow.ID, result); err != nil {
			log.Printf("worker: persist node result for run %s: %v", workflowRun.ID, err)
			return
		}
		deleteNodeJob(ctx, deps.sqsQueue, message)
	} else {
		deleteNodeJob(ctx, deps.sqsQueue, message)
	}
}

func persistNodeResult(ctx context.Context, store *db.Store, workflowRun sqlcdb.WorkflowRun, nodeID uuid.UUID, result executor.Result) error {
	return store.InTx(ctx, func(q *sqlcdb.Queries) error {
		if result.Completed {
			return q.CompleteWorkflowRun(ctx, sqlcdb.CompleteWorkflowRunParams{
				ID:         workflowRun.ID,
				LastOutput: result.Output,
			})
		}

		nextNodeID, err := q.GetNextNodeForExecution(ctx, sqlcdb.GetNextNodeForExecutionParams{
			WorkflowVersionID: workflowRun.WorkflowVersionID,
			FromNodeID:        nodeID,
			Label:             sqlcdb.EdgeLabel(result.NextEdgeLabel),
		})
		if err != nil {
			return err
		}

		if err := q.AdvanceWorkflowRun(ctx, sqlcdb.AdvanceWorkflowRunParams{
			ID:            workflowRun.ID,
			CurrentNodeID: nextNodeID,
			LastOutput:    result.Output,
		}); err != nil {
			return err
		}

		return q.InsertRunNodeJobOutbox(ctx, sqlcdb.InsertRunNodeJobOutboxParams{
			RunID:             workflowRun.ID,
			WorkflowVersionID: workflowRun.WorkflowVersionID,
			NodeID:            nextNodeID,
		})
	})
}

func deleteNodeJob(ctx context.Context, sqsQueue *queue.SQS, message types.Message) {
	if err := sqsQueue.Delete(ctx, *message.ReceiptHandle); err != nil {
		log.Printf("worker: delete node job: %v", err)
	}
}
