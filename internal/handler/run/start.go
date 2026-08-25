package run

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	sqlcdb "github.com/nabhag8848/orchex/internal/db/sqlc"
)

var (
	errMissingStartNode        = errors.New("published workflow is missing a start node")
	errPublishedWithoutVersion = errors.New("published workflow is missing live version")
)

func (h *Handler) start(ctx context.Context, id uuid.UUID, req StartRunRequest) (Run, error) {
	lastOutput, err := marshalStartOutput(req.Payload)
	if err != nil {
		return Run{}, err
	}

	var created sqlcdb.WorkflowRun
	err = h.store.InTx(ctx, func(q *sqlcdb.Queries) error {
		wf, err := q.GetPublishedWorkflowForStart(ctx, id)
		if err != nil {
			return err
		}
		if wf.LatestPublishedVersionID == nil {
			return errPublishedWithoutVersion
		}
		if wf.StartNodeID == uuid.Nil {
			return errMissingStartNode
		}
		versionID := *wf.LatestPublishedVersionID
		startNodeID := wf.StartNodeID

		created, err = q.InsertWorkflowRun(ctx, sqlcdb.InsertWorkflowRunParams{
			WorkflowID:        wf.ID,
			WorkflowVersionID: versionID,
			CurrentNodeID:     startNodeID,
			LastOutput:        lastOutput,
		})
		if err != nil {
			return err
		}

		return q.InsertRunNodeJobOutbox(ctx, sqlcdb.InsertRunNodeJobOutboxParams{
			RunID:             created.ID,
			WorkflowVersionID: versionID,
			NodeID:            startNodeID,
		})
	})
	if err != nil {
		return Run{}, err
	}
	return runFromRow(created), nil
}

type startOutput struct {
	Data startOutputData `json:"data"`
}

type startOutputData struct {
	Payload json.RawMessage `json:"payload"`
}

func marshalStartOutput(payload json.RawMessage) (json.RawMessage, error) {
	return json.Marshal(startOutput{
		Data: startOutputData{Payload: payload},
	})
}
