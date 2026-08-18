-- name: UpsertEdges :exec
INSERT INTO workflow_edges (
    id,
    workflow_version_id,
    from_node_id,
    to_node_id,
    label
)
SELECT
    (e->>'id')::uuid,
    sqlc.arg('workflow_version_id'),
    (e->>'from_node_id')::uuid,
    (e->>'to_node_id')::uuid,
    (e->>'label')::edge_label
FROM jsonb_array_elements(sqlc.arg('edges')::jsonb) AS t(e)
ON CONFLICT (workflow_version_id, id) DO UPDATE SET
    from_node_id = EXCLUDED.from_node_id,
    to_node_id = EXCLUDED.to_node_id,
    label = EXCLUDED.label;

-- cardinality 0 (empty payload) deletes every edge in the version.
-- name: DeleteEdgesNotIn :exec
DELETE FROM workflow_edges
WHERE workflow_version_id = sqlc.arg('workflow_version_id')
  AND (
      cardinality(sqlc.arg('ids')::uuid[]) = 0
      OR id <> ALL (sqlc.arg('ids')::uuid[])
  );
