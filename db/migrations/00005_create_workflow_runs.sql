-- Execution tables: runs (checkpoint + last hop output) and the node-job outbox.

-- +goose Up
-- +goose StatementBegin

CREATE TYPE trigger_type AS ENUM ('manual', 'webhook', 'scheduler');
CREATE TYPE workflow_run_status AS ENUM (
    'pending',
    'running',
    'paused',
    'failed',
    'completed',
    'cancelled'
);

CREATE TABLE workflow_runs (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    workflow_id uuid NOT NULL REFERENCES workflows (id),
    workflow_version_id uuid NOT NULL REFERENCES workflow_versions (id),
    status workflow_run_status NOT NULL DEFAULT 'pending',
    trigger_type trigger_type NOT NULL DEFAULT 'manual',
    current_node_id uuid NOT NULL,
    current_node_attempt int NOT NULL DEFAULT 1,
    last_output jsonb NOT NULL DEFAULT '{}',
    error jsonb,
    started_at timestamptz,
    paused_at timestamptz,
    cancelled_at timestamptz,
    completed_at timestamptz,
    failed_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT uq_workflow_runs_id_version UNIQUE (id, workflow_version_id),
    CONSTRAINT workflow_runs_checkpoint_fkey
        FOREIGN KEY (workflow_version_id, current_node_id)
        REFERENCES nodes (workflow_version_id, id)
);

CREATE TABLE run_node_jobs_outbox (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    run_id uuid NOT NULL,
    workflow_version_id uuid NOT NULL,
    node_id uuid NOT NULL,
    attempt int NOT NULL DEFAULT 1,
    available_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT run_node_jobs_outbox_run_fkey
        FOREIGN KEY (run_id, workflow_version_id)
        REFERENCES workflow_runs (id, workflow_version_id),
    CONSTRAINT run_node_jobs_outbox_node_fkey
        FOREIGN KEY (workflow_version_id, node_id)
        REFERENCES nodes (workflow_version_id, id)
);

CREATE TRIGGER trg_workflow_runs_touch_updated_at
    BEFORE UPDATE ON workflow_runs
    FOR EACH ROW
    EXECUTE FUNCTION touch_updated_at();

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS run_node_jobs_outbox;
DROP TABLE IF EXISTS workflow_runs;

DROP TYPE IF EXISTS workflow_run_status;
DROP TYPE IF EXISTS trigger_type;

-- +goose StatementEnd
