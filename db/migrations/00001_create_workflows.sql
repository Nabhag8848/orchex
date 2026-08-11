-- +goose Up
-- +goose StatementBegin
CREATE TYPE workflow_status AS ENUM ('draft', 'published', 'archived');

CREATE TABLE workflows (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name text NOT NULL,
    description text,
    status workflow_status NOT NULL DEFAULT 'draft',
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    last_published_at timestamptz
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS workflows;
DROP TYPE IF EXISTS workflow_status;
-- +goose StatementEnd
