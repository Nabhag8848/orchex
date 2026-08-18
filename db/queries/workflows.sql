-- name: GetWorkflow :one
SELECT
    id,
    name,
    description,
    status,
    latest_published_version_id,
    latest_version_id,
    created_at,
    updated_at,
    last_published_at
FROM workflows
WHERE id = $1
  AND status != 'archived';

-- name: CreateWorkflow :one
WITH new_workflow AS (
    INSERT INTO workflows (name, description, latest_version_id)
    VALUES (sqlc.arg('name'), sqlc.narg('description'), gen_random_uuid())
    RETURNING
        id,
        name,
        description,
        status,
        latest_published_version_id,
        latest_version_id,
        created_at,
        updated_at,
        last_published_at
),
new_version AS (
    INSERT INTO workflow_versions (id, workflow_id)
    SELECT latest_version_id, id
    FROM new_workflow
    RETURNING id
)
SELECT
    w.id,
    w.name,
    w.description,
    w.status,
    w.latest_published_version_id,
    w.latest_version_id,
    w.created_at,
    w.updated_at,
    w.last_published_at
FROM new_workflow w
INNER JOIN new_version v ON v.id = w.latest_version_id;
