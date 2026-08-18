-- name: UpsertNodes :exec
INSERT INTO nodes (
    id,
    workflow_version_id,
    node_type_id,
    name,
    config,
    position_x,
    position_y
)
SELECT
    (e->>'id')::uuid,
    sqlc.arg('workflow_version_id'),
    (e->>'node_type_id')::uuid,
    e->>'name',
    '{}'::jsonb,
    NULLIF(e->>'position_x', '')::double precision,
    NULLIF(e->>'position_y', '')::double precision
FROM jsonb_array_elements(sqlc.arg('nodes')::jsonb) AS t(e)
ON CONFLICT (workflow_version_id, id) DO UPDATE SET
    node_type_id = EXCLUDED.node_type_id,
    name = EXCLUDED.name,
    config = EXCLUDED.config,
    position_x = EXCLUDED.position_x,
    position_y = EXCLUDED.position_y;

-- cardinality 0 (empty payload) deletes every node in the version.
-- name: DeleteNodesNotIn :exec
DELETE FROM nodes
WHERE workflow_version_id = sqlc.arg('workflow_version_id')
  AND (
      cardinality(sqlc.arg('ids')::uuid[]) = 0
      OR id <> ALL (sqlc.arg('ids')::uuid[])
  );
