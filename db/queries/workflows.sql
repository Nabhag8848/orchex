-- name: GetWorkflowHead :one
SELECT
    id,
    latest_published_version_id,
    latest_version_id
FROM workflows
WHERE id = $1
  AND status != 'archived'
FOR UPDATE;

-- Locks the workflow and its head version so publish and a concurrent save
-- cannot interleave. Index: workflows_pkey, then versions PK by latest_version_id.
-- name: LockWorkflowForPublish :one
SELECT
    w.id,
    w.name,
    w.description,
    w.status,
    w.latest_published_version_id,
    w.latest_version_id,
    w.created_at,
    w.updated_at,
    w.last_published_at,
    v.version AS head_version,
    v.published_at AS head_published_at
FROM workflows w
INNER JOIN workflow_versions v ON v.id = w.latest_version_id
WHERE w.id = sqlc.arg('id')
  AND w.status != 'archived'
FOR UPDATE OF w, v;

-- Single statement: stamp the draft head, then point the live pointers at it.
-- Guards keep a stale caller from publishing a version that is no longer the head.
-- name: PublishWorkflowHead :one
WITH published_version AS (
    UPDATE workflow_versions v
    SET published_at = now()
    WHERE v.id = sqlc.arg('version_id')
      AND v.workflow_id = sqlc.arg('id')
      AND v.published_at IS NULL
    RETURNING v.id, v.version, v.published_at
)
UPDATE workflows w
SET
    status = 'published',
    latest_published_version_id = pv.id,
    last_published_at = pv.published_at
FROM published_version pv
WHERE w.id = sqlc.arg('id')
  AND w.latest_version_id = pv.id
  AND w.status != 'archived'
RETURNING
    w.id,
    w.name,
    w.description,
    w.status,
    w.latest_version_id,
    w.latest_published_version_id,
    w.created_at,
    w.updated_at,
    w.last_published_at,
    pv.version AS published_version;

-- name: GetWorkflowDetail :one
SELECT
    w.id,
    w.name,
    w.description,
    w.status,
    w.latest_published_version_id,
    w.latest_version_id,
    w.created_at,
    w.updated_at,
    w.last_published_at,
    v.id AS graph_id,
    v.version AS graph_version,
    v.published_at AS graph_published_at,
    CAST(COALESCE((
        SELECT jsonb_agg(
            jsonb_build_object(
                'id', n.id,
                'node_type', nt.type,
                'name', n.name,
                'config', n.config,
                'position', CASE
                    WHEN n.position_x IS NULL AND n.position_y IS NULL THEN NULL
                    ELSE jsonb_build_object('x', n.position_x, 'y', n.position_y)
                END
            )
            ORDER BY n.created_at, n.id
        )
        FROM nodes n
        INNER JOIN node_types nt ON nt.id = n.node_type_id
        WHERE n.workflow_version_id = v.id
    ), '[]'::jsonb) AS jsonb) AS nodes,
    CAST(COALESCE((
        SELECT jsonb_agg(
            jsonb_build_object(
                'id', e.id,
                'from_node_id', e.from_node_id,
                'to_node_id', e.to_node_id,
                'label', e.label
            )
            ORDER BY e.created_at, e.id
        )
        FROM workflow_edges e
        WHERE e.workflow_version_id = v.id
    ), '[]'::jsonb) AS jsonb) AS edges
FROM workflows w
INNER JOIN workflow_versions v
    ON v.id = CASE
        WHEN sqlc.arg('published')::bool THEN w.latest_published_version_id
        ELSE w.latest_version_id
    END
WHERE w.id = sqlc.arg('id')
  AND w.status != 'archived';

-- name: ListWorkflows :many
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
WHERE status != 'archived'
ORDER BY updated_at DESC, id;

-- name: WorkflowExists :one
SELECT EXISTS(
    SELECT 1
    FROM workflows
    WHERE id = $1
      AND status != 'archived'
);

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

-- name: UpdateWorkflow :exec
UPDATE workflows
SET
    name = sqlc.arg('name'),
    description = sqlc.narg('description'),
    latest_version_id = sqlc.arg('latest_version_id')
WHERE id = sqlc.arg('id')
  AND status != 'archived';

-- name: ArchiveWorkflow :one
WITH existing AS (
    SELECT w.id, w.status
    FROM workflows w
    WHERE w.id = sqlc.arg('id')
),
updated AS (
    UPDATE workflows w
    SET status = 'archived'
    FROM existing
    WHERE w.id = existing.id
      AND existing.status != 'archived'
    RETURNING w.id
)
SELECT
    EXISTS(SELECT 1 FROM existing) AS found,
    EXISTS(SELECT 1 FROM updated) AS archived;
