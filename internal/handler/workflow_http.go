package handler

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	sqlcdb "github.com/nabhag8848/orchex/internal/db/sqlc"
)

type CreateWorkflowRequest struct {
	Name        string  `json:"name" validate:"required,min=1,max=255"`
	Description *string `json:"description" validate:"omitempty,max=2000"`
}

type UpdateWorkflowRequest struct {
	Name        string  `json:"name" validate:"required,min=1,max=255"`
	Description *string `json:"description" validate:"omitempty,max=2000"`
	Nodes       []Node  `json:"nodes" validate:"dive"`
	Edges       []Edge  `json:"edges" validate:"dive"`
}

func (r *UpdateWorkflowRequest) withEmptyGraph() UpdateWorkflowRequest {
	if r.Nodes == nil {
		r.Nodes = []Node{}
	}
	if r.Edges == nil {
		r.Edges = []Edge{}
	}
	return *r
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
	NodeType string          `json:"node_type" validate:"required"`
	Name     string          `json:"name" validate:"required,min=1"`
	Config   json.RawMessage `json:"config"`
	Position *Position       `json:"position"`
}

func (n Node) coords() (x, y *float64) {
	if n.Position == nil {
		return nil, nil
	}
	px, py := n.Position.X, n.Position.Y
	return &px, &py
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

type WorkflowList struct {
	Items []Workflow `json:"items"`
}

func publishedRequested(version string) (bool, error) {
	switch version {
	case "", "latest":
		return false, nil
	case "published":
		return true, nil
	default:
		return false, fmt.Errorf("version must be latest or published")
	}
}

func workflowFromList(row sqlcdb.ListWorkflowsRow) Workflow {
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

func createdWorkflow(row sqlcdb.CreateWorkflowRow) WorkflowDetail {
	return WorkflowDetail{
		Workflow: Workflow{
			ID:                       row.ID,
			Name:                     row.Name,
			Description:              row.Description,
			Status:                   string(row.Status),
			LatestVersionID:          row.LatestVersionID,
			LatestPublishedVersionID: row.LatestPublishedVersionID,
			CreatedAt:                row.CreatedAt,
			UpdatedAt:                row.UpdatedAt,
			LastPublishedAt:          row.LastPublishedAt,
		},
		Graph: VersionGraph{
			ID:          row.LatestVersionID,
			Version:     1,
			PublishedAt: nil,
			Nodes:       []Node{},
			Edges:       []Edge{},
		},
	}
}
