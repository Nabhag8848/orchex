-- name: GetWorkflow :one
SELECT
    id,
    name,
    description,
    status,
    created_at,
    updated_at,
    last_published_at
FROM workflows
WHERE id = $1
  AND status != 'archived';
