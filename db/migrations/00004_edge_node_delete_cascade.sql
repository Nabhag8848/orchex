-- Edge FKs cascade on node delete. Unique (name) and (from, label) are
-- deferred to COMMIT so a save can upsert then delete extras in one txn.

-- +goose Up
-- +goose StatementBegin

ALTER TABLE workflow_edges
    DROP CONSTRAINT workflow_edges_from_fkey,
    DROP CONSTRAINT workflow_edges_to_fkey;

ALTER TABLE workflow_edges
    ADD CONSTRAINT workflow_edges_from_fkey
        FOREIGN KEY (workflow_version_id, from_node_id)
        REFERENCES nodes (workflow_version_id, id)
        ON DELETE CASCADE,
    ADD CONSTRAINT workflow_edges_to_fkey
        FOREIGN KEY (workflow_version_id, to_node_id)
        REFERENCES nodes (workflow_version_id, id)
        ON DELETE CASCADE;

ALTER TABLE nodes
    DROP CONSTRAINT uq_nodes_version_name;

ALTER TABLE nodes
    ADD CONSTRAINT uq_nodes_version_name
        UNIQUE (workflow_version_id, name)
        DEFERRABLE INITIALLY DEFERRED;

ALTER TABLE workflow_edges
    DROP CONSTRAINT uq_workflow_edges_source_label;

ALTER TABLE workflow_edges
    ADD CONSTRAINT uq_workflow_edges_source_label
        UNIQUE (workflow_version_id, from_node_id, label)
        DEFERRABLE INITIALLY DEFERRED;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE nodes
    DROP CONSTRAINT uq_nodes_version_name;

ALTER TABLE nodes
    ADD CONSTRAINT uq_nodes_version_name
        UNIQUE (workflow_version_id, name);

ALTER TABLE workflow_edges
    DROP CONSTRAINT uq_workflow_edges_source_label;

ALTER TABLE workflow_edges
    ADD CONSTRAINT uq_workflow_edges_source_label
        UNIQUE (workflow_version_id, from_node_id, label);

ALTER TABLE workflow_edges
    DROP CONSTRAINT workflow_edges_from_fkey,
    DROP CONSTRAINT workflow_edges_to_fkey;

ALTER TABLE workflow_edges
    ADD CONSTRAINT workflow_edges_from_fkey
        FOREIGN KEY (workflow_version_id, from_node_id)
        REFERENCES nodes (workflow_version_id, id),
    ADD CONSTRAINT workflow_edges_to_fkey
        FOREIGN KEY (workflow_version_id, to_node_id)
        REFERENCES nodes (workflow_version_id, id);

-- +goose StatementEnd
