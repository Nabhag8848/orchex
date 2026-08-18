package handler

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	sqlcdb "github.com/nabhag8848/orchex/internal/db/sqlc"
)

func (h *WorkflowHandler) save(ctx context.Context, id uuid.UUID, req UpdateWorkflowRequest, types nodeTypes) (WorkflowDetail, error) {
	var detail WorkflowDetail
	err := h.store.InTx(ctx, func(q *sqlcdb.Queries) error {
		head, err := q.GetWorkflowHead(ctx, id)
		if err != nil {
			return err
		}

		versionID := head.LatestVersionID
		fork := needsDraftFork(head)
		if fork {
			ver, err := q.CreateWorkflowVersion(ctx, head.ID)
			if err != nil {
				return fmt.Errorf("create workflow version: %w", err)
			}
			versionID = ver.ID
		}

		if err := persistGraph(ctx, q, versionID, req.Nodes, req.Edges, types, fork); err != nil {
			return fmt.Errorf("persist graph: %w", err)
		}

		if err := q.UpdateWorkflow(ctx, sqlcdb.UpdateWorkflowParams{
			Name:            req.Name,
			Description:     req.Description,
			LatestVersionID: versionID,
			ID:              head.ID,
		}); err != nil {
			return err
		}

		detail, err = getDetail(ctx, q, head.ID, false)
		return err
	})
	return detail, err
}

func needsDraftFork(head sqlcdb.GetWorkflowHeadRow) bool {
	return head.LatestPublishedVersionID != nil && *head.LatestPublishedVersionID == head.LatestVersionID
}

func persistGraph(ctx context.Context, q *sqlcdb.Queries, versionID uuid.UUID, nodes []Node, edges []Edge, types nodeTypes, fork bool) error {
	nodeJSON, err := encodeNodes(nodes, types)
	if err != nil {
		return err
	}
	if err := q.UpsertNodes(ctx, sqlcdb.UpsertNodesParams{
		WorkflowVersionID: versionID,
		Nodes:             nodeJSON,
	}); err != nil {
		return err
	}
	if err := q.DeleteNodesNotIn(ctx, sqlcdb.DeleteNodesNotInParams{
		WorkflowVersionID: versionID,
		Ids:               mapIDs(nodes, func(n Node) uuid.UUID { return n.ID }),
	}); err != nil {
		return err
	}

	edgeJSON, err := encodeEdges(edges)
	if err != nil {
		return err
	}
	if err := q.UpsertEdges(ctx, sqlcdb.UpsertEdgesParams{
		WorkflowVersionID: versionID,
		Edges:             edgeJSON,
	}); err != nil {
		return err
	}
	if err := q.DeleteEdgesNotIn(ctx, sqlcdb.DeleteEdgesNotInParams{
		WorkflowVersionID: versionID,
		Ids:               mapIDs(edges, func(e Edge) uuid.UUID { return e.ID }),
	}); err != nil {
		return err
	}

	if !fork {
		return q.TouchWorkflowVersion(ctx, versionID)
	}
	return nil
}

type nodeRecord struct {
	ID         uuid.UUID `json:"id"`
	NodeTypeID uuid.UUID `json:"node_type_id"`
	Name       string    `json:"name"`
	PositionX  *float64  `json:"position_x"`
	PositionY  *float64  `json:"position_y"`
}

type edgeRecord struct {
	ID         uuid.UUID `json:"id"`
	FromNodeID uuid.UUID `json:"from_node_id"`
	ToNodeID   uuid.UUID `json:"to_node_id"`
	Label      string    `json:"label"`
}

func encodeNodes(nodes []Node, types nodeTypes) (json.RawMessage, error) {
	rows := make([]nodeRecord, 0, len(nodes))
	for _, n := range nodes {
		typeID, ok := types.id(n.NodeType)
		if !ok {
			return nil, fmt.Errorf("unknown node_type %q", n.NodeType)
		}
		x, y := n.coords()
		rows = append(rows, nodeRecord{
			ID:         n.ID,
			NodeTypeID: typeID,
			Name:       n.Name,
			PositionX:  x,
			PositionY:  y,
		})
	}
	return json.Marshal(rows)
}

func encodeEdges(edges []Edge) (json.RawMessage, error) {
	rows := make([]edgeRecord, 0, len(edges))
	for _, e := range edges {
		rows = append(rows, edgeRecord{
			ID:         e.ID,
			FromNodeID: e.FromNodeID,
			ToNodeID:   e.ToNodeID,
			Label:      string(edgeLabel(e.Label)),
		})
	}
	return json.Marshal(rows)
}

func mapIDs[T any](items []T, id func(T) uuid.UUID) []uuid.UUID {
	out := make([]uuid.UUID, 0, len(items))
	for _, item := range items {
		out = append(out, id(item))
	}
	return out
}
