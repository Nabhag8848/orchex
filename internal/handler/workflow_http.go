package handler

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	sqlcdb "github.com/nabhag8848/orchex/internal/db/sqlc"
)

type CreateWorkflowRequest struct {
	Name        string  `json:"name" validate:"required,min=1,max=255"`
	Description *string `json:"description" validate:"omitempty,max=2000"`
}

type Workflow struct {
	ID                       uuid.UUID  `json:"id"`
	Name                     string     `json:"name"`
	Description              *string    `json:"description"`
	Status                   string     `json:"status"`
	LatestVersionID          uuid.UUID  `json:"latest_version_id"`
	LatestPublishedVersionID *uuid.UUID `json:"latest_published_version_id"`
	CreatedAt                time.Time  `json:"created_at"`
	UpdatedAt                time.Time  `json:"updated_at"`
	LastPublishedAt          *time.Time `json:"last_published_at"`
}

type Position struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

type Node struct {
	ID       uuid.UUID       `json:"id"`
	NodeType string          `json:"node_type"`
	Name     string          `json:"name"`
	Config   json.RawMessage `json:"config"`
	Position *Position       `json:"position"`
}

type Edge struct {
	ID         uuid.UUID `json:"id"`
	FromNodeID uuid.UUID `json:"from_node_id"`
	ToNodeID   uuid.UUID `json:"to_node_id"`
	Label      string    `json:"label"`
}

type VersionGraph struct {
	ID          uuid.UUID  `json:"id"`
	Version     int        `json:"version"`
	PublishedAt *time.Time `json:"published_at"`
	Nodes       []Node     `json:"nodes"`
	Edges       []Edge     `json:"edges"`
}

type WorkflowDetail struct {
	Workflow
	Graph VersionGraph `json:"graph"`
}

func workflowFromCreate(row sqlcdb.CreateWorkflowRow) Workflow {
	return Workflow{
		ID:                       row.ID,
		Name:                     row.Name,
		Description:              row.Description,
		Status:                   string(row.Status),
		LatestVersionID:          row.LatestVersionID,
		LatestPublishedVersionID: row.LatestPublishedVersionID,
		CreatedAt:                row.CreatedAt,
		UpdatedAt:                row.UpdatedAt,
		LastPublishedAt:          row.LastPublishedAt,
	}
}

func workflowFromGet(row sqlcdb.GetWorkflowRow) Workflow {
	return Workflow{
		ID:                       row.ID,
		Name:                     row.Name,
		Description:              row.Description,
		Status:                   string(row.Status),
		LatestVersionID:          row.LatestVersionID,
		LatestPublishedVersionID: row.LatestPublishedVersionID,
		CreatedAt:                row.CreatedAt,
		UpdatedAt:                row.UpdatedAt,
		LastPublishedAt:          row.LastPublishedAt,
	}
}

func emptyDraftGraph(versionID uuid.UUID) VersionGraph {
	return VersionGraph{
		ID:          versionID,
		Version:     1,
		PublishedAt: nil,
		Nodes:       []Node{},
		Edges:       []Edge{},
	}
}
