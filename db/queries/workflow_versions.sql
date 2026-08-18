-- name: CreateWorkflowVersion :one
INSERT INTO workflow_versions (workflow_id, version)
SELECT sqlc.arg('workflow_id'), COALESCE(MAX(version), 0) + 1
FROM workflow_versions
WHERE workflow_id = sqlc.arg('workflow_id')
RETURNING
    id,
    workflow_id,
    version,
    created_at,
    last_updated_at,
    published_at;

-- name: TouchWorkflowVersion :exec
UPDATE workflow_versions
SET version = version
WHERE id = $1;
