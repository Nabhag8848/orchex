package worker

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	sqlcdb "github.com/nabhag8848/orchex/internal/db/sqlc"
	"github.com/nabhag8848/orchex/internal/executor"
)

func (w *Worker) persistNodeResult(ctx context.Context, workflowRun sqlcdb.WorkflowRun, nodeID uuid.UUID, result executor.Result) error {
	var nextNodeID uuid.UUID
	err := w.store.InTx(ctx, func(q *sqlcdb.Queries) error {
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
