package worker

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	sqlcdb "github.com/nabhag8848/orchex/internal/db/sqlc"
	"github.com/nabhag8848/orchex/internal/executor"
	"github.com/nabhag8848/orchex/internal/queue"
)

func (w *Worker) handleNodeJob(ctx context.Context, message types.Message) {
	if ctx.Err() != nil {
		return
	}

	var job queue.NodeJobMessage
	if err := json.Unmarshal([]byte(*message.Body), &job); err != nil {
		slog.Warn("decode node job", "error", err)
		return
	}
	slog.Debug("execute node job", "run_id", job.RunID, "node_id", job.NodeID, "attempt", job.Attempt)

	workflowRun, err := w.store.GetWorkflowRun(ctx, job.RunID)
	if err != nil {
		slog.Warn("get workflow run", "run_id", job.RunID, "error", err)
		return
	}
	if workflowRun.Status == sqlcdb.WorkflowRunStatusPending || workflowRun.Status == sqlcdb.WorkflowRunStatusRunning {
		nodeRow, err := w.store.GetNodeForExecution(ctx, sqlcdb.GetNodeForExecutionParams{
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

		result, err := w.executor.Execute(ctx, executor.Node{
			WorkflowVersionID: nodeRow.WorkflowVersionID,
			ID:                nodeRow.ID,
			Name:              nodeRow.Name,
			Type:              executor.NodeType(nodeRow.Type),
			Config:            nodeRow.Config,
		}, workflowRun.LastOutput)
		if err != nil {
			slog.Warn("execute node", "run_id", workflowRun.ID, "node_id", nodeRow.ID, "node_type", nodeRow.Type, "error", err)
			if err := w.store.FailWorkflowRun(ctx, sqlcdb.FailWorkflowRunParams{
				ID:           workflowRun.ID,
				ErrorMessage: err.Error(),
			}); err != nil {
				slog.Warn("fail workflow run", "run_id", workflowRun.ID, "error", err)
			}
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

		if err := w.persistNodeResult(ctx, workflowRun, nodeRow.ID, result); err != nil {
			slog.Warn("persist node result", "run_id", workflowRun.ID, "node_id", nodeRow.ID, "error", err)
			return
		}
		w.deleteNodeJob(ctx, message)
	} else {
		slog.Debug("delete node job for inactive run", "run_id", workflowRun.ID, "status", workflowRun.Status)
		w.deleteNodeJob(ctx, message)
	}
}

func (w *Worker) deleteNodeJob(ctx context.Context, message types.Message) {
	if err := w.queue.Delete(ctx, *message.ReceiptHandle); err != nil {
		slog.Warn("delete node job", "error", err)
	}
}
