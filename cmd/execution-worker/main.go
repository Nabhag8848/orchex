package main

import (
	"context"
	"encoding/json"
	"log/slog"
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
	"github.com/nabhag8848/orchex/internal/logger"
	"github.com/nabhag8848/orchex/internal/queue"
	"github.com/nabhag8848/orchex/internal/sandbox"
	"github.com/nabhag8848/orchex/internal/worker"
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

	sandboxClient, err := sandbox.NewLambda(ctx, cfg.Lambda)
	if err != nil {
		logger.Fatal("create sandbox client", "error", err)
	}

	sqsQueue, err := queue.New(ctx, cfg.SQS)
	if err != nil {
		logger.Fatal("create SQS client", "error", err)
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
				slog.Warn("receive node jobs", "error", err)
				continue
			}

			slog.Debug("received node jobs", "count", len(messages))
			for _, message := range messages {
				go handleNodeJob(ctx, deps, message)
			}
		}
	}()

	go func() {
		result, err := sandboxClient.Invoke(ctx, "return { ping: true };", json.RawMessage(`{"data":{}}`), 5000)
		if err != nil {
			slog.Warn("sandbox startup invoke failed", "error", err)
			return
		}
		slog.Debug("sandbox startup invoke succeeded", "result", string(result))
	}()

	e := worker.NewServer()
	slog.Info("execution worker starting", "address", cfg.HTTPAddr)
	sc := echo.StartConfig{Address: cfg.HTTPAddr}
	if err := sc.Start(ctx, e); err != nil {
		logger.Fatal("start server", "error", err)
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
		slog.Warn("decode node job", "error", err)
		return
	}
	slog.Debug("execute node job", "run_id", job.RunID, "node_id", job.NodeID, "attempt", job.Attempt)

	workflowRun, err := deps.store.GetWorkflowRun(ctx, job.RunID)
	if err != nil {
		slog.Warn("get workflow run", "run_id", job.RunID, "error", err)
		return
	}
	if workflowRun.Status == sqlcdb.WorkflowRunStatusPending || workflowRun.Status == sqlcdb.WorkflowRunStatusRunning {
		nodeRow, err := deps.store.GetNodeForExecution(ctx, sqlcdb.GetNodeForExecutionParams{
			WorkflowVersionID: workflowRun.WorkflowVersionID,
			NodeID:            workflowRun.CurrentNodeID,
		})
		if err != nil {
			slog.Warn("get current node", "run_id", workflowRun.ID, "error", err)
			return
		}
		slog.Debug("node execution started",
			"run_id", workflowRun.ID,
			"node_id", nodeRow.ID,
			"node_type", nodeRow.Type,
			"input", string(workflowRun.LastOutput),
		)

		result, err := deps.registry.Execute(ctx, executor.Node{
			WorkflowVersionID: nodeRow.WorkflowVersionID,
			ID:                nodeRow.ID,
			Name:              nodeRow.Name,
			Type:              executor.NodeType(nodeRow.Type),
			Config:            nodeRow.Config,
		}, workflowRun.LastOutput)
		if err != nil {
			slog.Warn("execute node", "run_id", workflowRun.ID, "node_id", nodeRow.ID, "node_type", nodeRow.Type, "error", err)
			return
		}
		slog.Debug("node execution completed",
			"run_id", workflowRun.ID,
			"node_id", nodeRow.ID,
			"node_type", nodeRow.Type,
			"output", string(result.Output),
			"next_edge_label", result.NextEdgeLabel,
			"completed", result.Completed,
		)

		if err := persistNodeResult(ctx, deps.store, workflowRun, nodeRow.ID, result); err != nil {
			slog.Warn("persist node result", "run_id", workflowRun.ID, "node_id", nodeRow.ID, "error", err)
			return
		}
		deleteNodeJob(ctx, deps.sqsQueue, message)
	} else {
		slog.Debug("delete node job for inactive run", "run_id", workflowRun.ID, "status", workflowRun.Status)
		deleteNodeJob(ctx, deps.sqsQueue, message)
	}
}

func persistNodeResult(ctx context.Context, store *db.Store, workflowRun sqlcdb.WorkflowRun, nodeID uuid.UUID, result executor.Result) error {
	var nextNodeID uuid.UUID
	err := store.InTx(ctx, func(q *sqlcdb.Queries) error {
		if result.Completed {
			return q.CompleteWorkflowRun(ctx, sqlcdb.CompleteWorkflowRunParams{
				ID:         workflowRun.ID,
				LastOutput: result.Output,
			})
		}

		var err error
		nextNodeID, err = q.GetNextNodeForExecution(ctx, sqlcdb.GetNextNodeForExecutionParams{
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
	if err != nil {
		return err
	}

	if result.Completed {
		slog.Debug("workflow run completed", "run_id", workflowRun.ID, "node_id", nodeID, "output", string(result.Output))
		return nil
	}

	slog.Debug("workflow run advanced",
		"run_id", workflowRun.ID,
		"from_node_id", nodeID,
		"to_node_id", nextNodeID,
		"next_edge_label", result.NextEdgeLabel,
	)
	slog.Debug("queued next node job", "run_id", workflowRun.ID, "node_id", nextNodeID)
	return nil
}

func deleteNodeJob(ctx context.Context, sqsQueue *queue.SQS, message types.Message) {
	if err := sqsQueue.Delete(ctx, *message.ReceiptHandle); err != nil {
		slog.Warn("delete node job", "error", err)
	}
}
