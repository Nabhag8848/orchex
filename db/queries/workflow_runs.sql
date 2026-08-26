-- name: GetWorkflowRun :one
SELECT *
FROM workflow_runs
WHERE id = $1;

-- Published workflow plus that version's Start node.
-- LEFT JOIN keeps a published row when Start is missing so the handler can 500
-- instead of treating it as an unpublished 404.
-- COALESCE: sqlc types nodes.id as NOT NULL even on this LEFT JOIN, so a
-- missing Start would fail Scan. The zero UUID is the missing-Start marker.
-- name: GetPublishedWorkflowForStart :one
SELECT
    w.id,
    w.latest_published_version_id,
    COALESCE(s.id, '00000000-0000-0000-0000-000000000000'::uuid) AS start_node_id
FROM workflows w
LEFT JOIN (
    SELECT n.workflow_version_id, n.id
    FROM nodes n
    INNER JOIN node_types nt ON nt.id = n.node_type_id
    WHERE nt.type = 'start'
) s ON s.workflow_version_id = w.latest_published_version_id
WHERE w.id = $1
  AND w.status = 'published';

-- name: InsertWorkflowRun :one
INSERT INTO workflow_runs (
    workflow_id,
    workflow_version_id,
    status,
    trigger_type,
    current_node_id,
    current_node_attempt,
    last_output
) VALUES (
    sqlc.arg('workflow_id'),
    sqlc.arg('workflow_version_id'),
    'pending',
    'manual',
    sqlc.arg('current_node_id'),
    1,
    sqlc.arg('last_output')
)
RETURNING *;

-- name: InsertRunNodeJobOutbox :exec
INSERT INTO run_node_jobs_outbox (
    run_id,
    workflow_version_id,
    node_id,
    attempt
) VALUES (
    sqlc.arg('run_id'),
    sqlc.arg('workflow_version_id'),
    sqlc.arg('node_id'),
    1
);

-- name: WorkflowRunExists :one
SELECT EXISTS(
    SELECT 1
    FROM workflow_runs
    WHERE id = $1
);

-- Soft pause: pending|running → paused; already paused is idempotent (keep paused_at).
-- Handler checks WorkflowRunExists first; no row here means terminal status → 409.
-- name: PauseWorkflowRun :one
UPDATE workflow_runs
SET
    status = 'paused',
    paused_at = COALESCE(paused_at, now())
WHERE id = $1
  AND status IN ('pending', 'running', 'paused')
RETURNING *;
