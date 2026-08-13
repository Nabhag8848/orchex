-- +goose Up
-- +goose StatementBegin

CREATE TYPE node_category AS ENUM ('trigger', 'logic', 'action', 'terminal');
CREATE TYPE edge_label AS ENUM ('default', 'true', 'false');

CREATE TABLE node_types (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    type text NOT NULL UNIQUE,
    category node_category NOT NULL,
    display_name text NOT NULL,
    min_in_degree int NOT NULL,
    max_in_degree int NOT NULL,
    min_out_degree int NOT NULL,
    max_out_degree int NOT NULL,
    config_schema jsonb NOT NULL DEFAULT '{}',
    input_schema jsonb NOT NULL DEFAULT '{}',
    output_schema jsonb NOT NULL DEFAULT '{}',
    error_schema jsonb NOT NULL DEFAULT '{}',
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE workflow_versions (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    workflow_id uuid NOT NULL REFERENCES workflows (id),
    version int NOT NULL DEFAULT 1,
    created_at timestamptz NOT NULL DEFAULT now(),
    last_updated_at timestamptz NOT NULL DEFAULT now(),
    published_at timestamptz,
    CONSTRAINT uq_workflow_versions_workflow_version UNIQUE (workflow_id, version)
);

-- Partial unique index: at most one draft (published_at IS NULL) per workflow.
-- Not representable in DBML.
CREATE UNIQUE INDEX uq_workflow_versions_one_draft
    ON workflow_versions (workflow_id)
    WHERE published_at IS NULL;

ALTER TABLE workflows
    ADD COLUMN latest_published_version_id uuid,
    ADD COLUMN latest_version_id uuid NOT NULL;

-- Circular FKs with workflow_versions: deferred so create can insert
-- workflow → version → set pointers in one transaction.
ALTER TABLE workflows
    ADD CONSTRAINT workflows_latest_version_id_fkey
        FOREIGN KEY (latest_version_id)
        REFERENCES workflow_versions (id)
        DEFERRABLE INITIALLY DEFERRED;

ALTER TABLE workflows
    ADD CONSTRAINT workflows_latest_published_version_id_fkey
        FOREIGN KEY (latest_published_version_id)
        REFERENCES workflow_versions (id)
        DEFERRABLE INITIALLY DEFERRED;

CREATE TABLE nodes (
    id uuid NOT NULL DEFAULT gen_random_uuid(),
    workflow_version_id uuid NOT NULL REFERENCES workflow_versions (id),
    node_type_id uuid NOT NULL REFERENCES node_types (id),
    name text NOT NULL,
    config jsonb NOT NULL DEFAULT '{}',
    position_x double precision,
    position_y double precision,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (workflow_version_id, id),
    CONSTRAINT uq_nodes_version_name UNIQUE (workflow_version_id, name)
);

CREATE TABLE workflow_edges (
    id uuid NOT NULL DEFAULT gen_random_uuid(),
    workflow_version_id uuid NOT NULL REFERENCES workflow_versions (id),
    from_node_id uuid NOT NULL,
    to_node_id uuid NOT NULL,
    label edge_label NOT NULL DEFAULT 'default',
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (workflow_version_id, id),
    CONSTRAINT uq_workflow_edges_source_label UNIQUE (workflow_version_id, from_node_id, label),
    CONSTRAINT workflow_edges_from_fkey
        FOREIGN KEY (workflow_version_id, from_node_id)
        REFERENCES nodes (workflow_version_id, id),
    CONSTRAINT workflow_edges_to_fkey
        FOREIGN KEY (workflow_version_id, to_node_id)
        REFERENCES nodes (workflow_version_id, id)
);

CREATE FUNCTION touch_updated_at()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$;

CREATE FUNCTION touch_last_updated_at()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
    NEW.last_updated_at = now();
    RETURN NEW;
END;
$$;

CREATE TRIGGER trg_node_types_touch_updated_at
    BEFORE UPDATE ON node_types
    FOR EACH ROW
    EXECUTE FUNCTION touch_updated_at();

CREATE TRIGGER trg_workflows_touch_updated_at
    BEFORE UPDATE ON workflows
    FOR EACH ROW
    EXECUTE FUNCTION touch_updated_at();

CREATE TRIGGER trg_nodes_touch_updated_at
    BEFORE UPDATE ON nodes
    FOR EACH ROW
    EXECUTE FUNCTION touch_updated_at();

CREATE TRIGGER trg_workflow_edges_touch_updated_at
    BEFORE UPDATE ON workflow_edges
    FOR EACH ROW
    EXECUTE FUNCTION touch_updated_at();

CREATE TRIGGER trg_workflow_versions_touch_last_updated_at
    BEFORE UPDATE ON workflow_versions
    FOR EACH ROW
    EXECUTE FUNCTION touch_last_updated_at();

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TRIGGER IF EXISTS trg_workflows_touch_updated_at ON workflows;

DROP TABLE IF EXISTS workflow_edges;
DROP TABLE IF EXISTS nodes;

ALTER TABLE workflows
    DROP CONSTRAINT IF EXISTS workflows_latest_published_version_id_fkey;
ALTER TABLE workflows
    DROP CONSTRAINT IF EXISTS workflows_latest_version_id_fkey;
ALTER TABLE workflows
    DROP COLUMN IF EXISTS latest_version_id;
ALTER TABLE workflows
    DROP COLUMN IF EXISTS latest_published_version_id;

DROP TABLE IF EXISTS workflow_versions;
DROP TABLE IF EXISTS node_types;

DROP FUNCTION IF EXISTS touch_last_updated_at();
DROP FUNCTION IF EXISTS touch_updated_at();

DROP TYPE IF EXISTS edge_label;
DROP TYPE IF EXISTS node_category;

-- +goose StatementEnd
