package worker

import (
	"context"
	"encoding/json"
	"log/slog"
	"sync"

	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/google/uuid"
	sqlcdb "github.com/nabhag8848/orchex/internal/db/sqlc"
	"github.com/nabhag8848/orchex/internal/executor"
)

// Queue supplies node jobs and acknowledges processed messages.
type Queue interface {
	Receive(context.Context) ([]types.Message, error)
	Delete(context.Context, string) error
}

// Store provides run checkpoints and atomic result persistence.
type Store interface {
	GetWorkflowRun(context.Context, uuid.UUID) (sqlcdb.WorkflowRun, error)
	GetNodeForExecution(context.Context, sqlcdb.GetNodeForExecutionParams) (sqlcdb.GetNodeForExecutionRow, error)
	FailWorkflowRun(context.Context, sqlcdb.FailWorkflowRunParams) error
	InTx(context.Context, func(*sqlcdb.Queries) error) error
}

// NodeExecutor dispatches a node to its registered executor.
type NodeExecutor interface {
	Execute(context.Context, executor.Node, json.RawMessage) (executor.Result, error)
}

// Worker polls for jobs and executes each received message independently.
type Worker struct {
	store    Store
	queue    Queue
	executor NodeExecutor
}

func New(store Store, queue Queue, executor NodeExecutor) *Worker {
	return &Worker{store: store, queue: queue, executor: executor}
}

// Run processes each batch concurrently and waits before receiving another.
// Queue.Receive supplies up to ten jobs with a 30-second visibility timeout.
func (w *Worker) Run(ctx context.Context) {
	for ctx.Err() == nil {
		messages, err := w.queue.Receive(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			slog.Warn("receive node jobs", "error", err)
			continue
		}

		slog.Debug("received node jobs", "count", len(messages))
		var wg sync.WaitGroup
		for _, message := range messages {
			wg.Add(1)
			go func() {
				defer wg.Done()
				w.handleNodeJob(ctx, message)
			}()
		}
		wg.Wait()
	}
}
