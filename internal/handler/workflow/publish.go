package workflow

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	sqlcdb "github.com/nabhag8848/orchex/internal/db/sqlc"
)

func (h *Handler) publish(ctx context.Context, id uuid.UUID) (PublishWorkflowResponse, error) {
	var result PublishWorkflowResponse
	err := h.store.InTx(ctx, func(q *sqlcdb.Queries) error {
		head, err := q.LockWorkflowForPublish(ctx, id)
		if err != nil {
			return err
		}

		if head.HeadPublishedAt != nil {
			if head.LatestPublishedVersionID == nil || head.LastPublishedAt == nil {
				return fmt.Errorf("published head is missing live pointers")
			}
			result = publishResponse(
				head.ID,
				head.Name,
				head.Description,
				string(head.Status),
				head.LatestVersionID,
				*head.LatestPublishedVersionID,
				head.CreatedAt,
				head.UpdatedAt,
				*head.LastPublishedAt,
				int(head.HeadVersion),
				*head.HeadPublishedAt,
			)
			return nil
		}

		nodeRows, err := q.ListNodesForPublish(ctx, head.LatestVersionID)
		if err != nil {
			return err
		}
		edgeRows, err := q.ListEdgesForPublish(ctx, head.LatestVersionID)
		if err != nil {
			return err
		}
		types, err := loadNodeTypes(ctx, q)
		if err != nil {
			return err
		}
		for _, row := range nodeRows {
			if err := types.validateConfig(row.ID, row.Type, row.Config); err != nil {
				return invalidGraph("%s", err.Error())
			}
		}
		if err := validatePublishGraph(toPublishNodes(nodeRows), toPublishEdges(edgeRows)); err != nil {
			return err
		}

		published, err := q.PublishWorkflowHead(ctx, sqlcdb.PublishWorkflowHeadParams{
			VersionID: head.LatestVersionID,
			ID:        head.ID,
		})
		if err != nil {
			return err
		}
		if published.LatestPublishedVersionID == nil || published.LastPublishedAt == nil {
			return fmt.Errorf("publish did not set live pointers")
		}

		result = publishResponse(
			published.ID,
			published.Name,
			published.Description,
			string(published.Status),
			published.LatestVersionID,
			*published.LatestPublishedVersionID,
			published.CreatedAt,
			published.UpdatedAt,
			*published.LastPublishedAt,
			int(published.PublishedVersion),
			*published.LastPublishedAt,
		)
		return nil
	})
	return result, err
}

func toPublishNodes(rows []sqlcdb.ListNodesForPublishRow) []publishNode {
	out := make([]publishNode, 0, len(rows))
	for _, r := range rows {
		out = append(out, publishNode{
			ID:           r.ID,
			Type:         r.Type,
			MinInDegree:  int(r.MinInDegree),
			MaxInDegree:  int(r.MaxInDegree),
			MinOutDegree: int(r.MinOutDegree),
			MaxOutDegree: int(r.MaxOutDegree),
		})
	}
	return out
}

func toPublishEdges(rows []sqlcdb.ListEdgesForPublishRow) []publishEdge {
	out := make([]publishEdge, 0, len(rows))
	for _, r := range rows {
		out = append(out, publishEdge{
			From:  r.FromNodeID,
			To:    r.ToNodeID,
			Label: string(r.Label),
		})
	}
	return out
}

func publishResponse(
	id uuid.UUID,
	name string,
	description *string,
	status string,
	latestVersionID uuid.UUID,
	publishedVersionID uuid.UUID,
	createdAt time.Time,
	updatedAt time.Time,
	lastPublishedAt time.Time,
	version int,
	publishedAt time.Time,
) PublishWorkflowResponse {
	return PublishWorkflowResponse{
		ID:                       id,
		Name:                     name,
		Description:              description,
		Status:                   status,
		LatestVersionID:          latestVersionID,
		LatestPublishedVersionID: publishedVersionID,
		CreatedAt:                createdAt,
		UpdatedAt:                updatedAt,
		LastPublishedAt:          lastPublishedAt,
		PublishedVersion: PublishedVersion{
			ID:          publishedVersionID,
			Version:     version,
			PublishedAt: publishedAt,
		},
	}
}
