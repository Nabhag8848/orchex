-- name: ListNodeTypes :many
SELECT id, type, config_schema
FROM node_types;
