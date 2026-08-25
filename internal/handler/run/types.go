package run

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	sqlcdb "github.com/nabhag8848/orchex/internal/db/sqlc"
)

type StartRunRequest struct {
	WorkflowID uuid.UUID       `json:"workflow_id" validate:"required"`
	Payload    json.RawMessage `json:"payload" validate:"required"`
}

type Run struct {
	ID                 uuid.UUID        `json:"id"`
	WorkflowID         uuid.UUID        `json:"workflow_id"`
	WorkflowVersionID  uuid.UUID        `json:"workflow_version_id"`
	Status             string           `json:"status"`
	TriggerType        string           `json:"trigger_type"`
	CurrentNodeID      uuid.UUID        `json:"current_node_id"`
	CurrentNodeAttempt int32            `json:"current_node_attempt"`
	LastOutput         json.RawMessage  `json:"last_output"`
	Error              *json.RawMessage `json:"error"`
	StartedAt          *time.Time       `json:"started_at"`
	PausedAt           *time.Time       `json:"paused_at"`
	CancelledAt        *time.Time       `json:"cancelled_at"`
	CompletedAt        *time.Time       `json:"completed_at"`
	FailedAt           *time.Time       `json:"failed_at"`
	CreatedAt          time.Time        `json:"created_at"`
	UpdatedAt          time.Time        `json:"updated_at"`
}

func runFromRow(row sqlcdb.WorkflowRun) Run {
	return Run{
		ID:                 row.ID,
		WorkflowID:         row.WorkflowID,
		WorkflowVersionID:  row.WorkflowVersionID,
		Status:             string(row.Status),
		TriggerType:        string(row.TriggerType),
		CurrentNodeID:      row.CurrentNodeID,
		CurrentNodeAttempt: row.CurrentNodeAttempt,
		LastOutput:         row.LastOutput,
		Error:              row.Error,
		StartedAt:          row.StartedAt,
		PausedAt:           row.PausedAt,
		CancelledAt:        row.CancelledAt,
		CompletedAt:        row.CompletedAt,
		FailedAt:           row.FailedAt,
		CreatedAt:          row.CreatedAt,
		UpdatedAt:          row.UpdatedAt,
	}
}
