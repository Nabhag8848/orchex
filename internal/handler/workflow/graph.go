package workflow

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	sqlcdb "github.com/nabhag8848/orchex/internal/db/sqlc"
	"github.com/santhosh-tekuri/jsonschema/v6"
)

type nodeTypes struct {
	idByName     map[string]uuid.UUID
	schemaByName map[string]*jsonschema.Schema
	names        []string
}

func loadNodeTypes(ctx context.Context, q *sqlcdb.Queries) (nodeTypes, error) {
	rows, err := q.ListNodeTypes(ctx)
	if err != nil {
		return nodeTypes{}, err
	}

	compiler := jsonschema.NewCompiler()
	compiler.DefaultDraft(jsonschema.Draft2020)
	compiler.AssertFormat()

	nt := nodeTypes{
		idByName:     make(map[string]uuid.UUID, len(rows)),
		schemaByName: make(map[string]*jsonschema.Schema, len(rows)),
		names:        make([]string, 0, len(rows)),
	}
	for _, row := range rows {
		sch, err := compileConfigSchema(compiler, row.Type, row.ConfigSchema)
		if err != nil {
			return nodeTypes{}, err
		}
		nt.idByName[row.Type] = row.ID
		nt.schemaByName[row.Type] = sch
		nt.names = append(nt.names, row.Type)
	}
	return nt, nil
}

func compileConfigSchema(compiler *jsonschema.Compiler, typeName string, raw json.RawMessage) (*jsonschema.Schema, error) {
	if len(raw) == 0 {
		raw = json.RawMessage(`{}`)
	}
	doc, err := jsonschema.UnmarshalJSON(bytes.NewReader(raw))
	if err != nil {
		return nil, fmt.Errorf("parse config_schema for %q: %w", typeName, err)
	}
	url := "orchex://node-types/" + typeName + "/config"
	if err := compiler.AddResource(url, doc); err != nil {
		return nil, fmt.Errorf("add config_schema for %q: %w", typeName, err)
	}
	sch, err := compiler.Compile(url)
	if err != nil {
		return nil, fmt.Errorf("compile config_schema for %q: %w", typeName, err)
	}
	return sch, nil
}

func (nt nodeTypes) id(name string) (uuid.UUID, bool) {
	id, ok := nt.idByName[name]
	return id, ok
}

func (nt nodeTypes) known() string {
	return strings.Join(nt.names, ", ")
}

func (nt nodeTypes) validateConfig(nodeID uuid.UUID, nodeType string, config json.RawMessage) error {
	sch, ok := nt.schemaByName[nodeType]
	if !ok {
		return fmt.Errorf("node %s has unknown node_type %q. Use one of: %s", nodeID, nodeType, nt.known())
	}
	if len(config) == 0 {
		config = json.RawMessage(`{}`)
	}
	inst, err := jsonschema.UnmarshalJSON(bytes.NewReader(config))
	if err != nil {
		return fmt.Errorf("node %s (%s): config is not valid JSON: %v", nodeID, nodeType, err)
	}
	if err := sch.Validate(inst); err != nil {
		return fmt.Errorf("node %s (%s): config does not match config_schema: %v", nodeID, nodeType, err)
	}
	return nil
}

func validateGraph(nodes []Node, edges []Edge, types nodeTypes) error {
	ids, err := uniqueNodes(nodes, types)
	if err != nil {
		return err
	}
	return uniqueEdges(edges, ids)
}

func uniqueNodes(nodes []Node, types nodeTypes) (map[uuid.UUID]struct{}, error) {
	ids := make(map[uuid.UUID]struct{}, len(nodes))
	names := make(map[string]uuid.UUID, len(nodes))

	for _, n := range nodes {
		if n.ID == uuid.Nil {
			return nil, fmt.Errorf("every node must include a client-generated id")
		}
		if _, exists := ids[n.ID]; exists {
			return nil, fmt.Errorf("duplicate node id %s. Each node id must be unique in this graph", n.ID)
		}
		ids[n.ID] = struct{}{}

		if strings.TrimSpace(n.Name) == "" {
			return nil, fmt.Errorf("node %s is missing a name", n.ID)
		}
		if other, exists := names[n.Name]; exists {
			return nil, fmt.Errorf("nodes %s and %s share the name %q. Node names must be unique in this graph", other, n.ID, n.Name)
		}
		names[n.Name] = n.ID

		if _, ok := types.id(n.NodeType); !ok {
			return nil, fmt.Errorf("node %s has unknown node_type %q. Use one of: %s", n.ID, n.NodeType, types.known())
		}
		if err := types.validateConfig(n.ID, n.NodeType, n.Config); err != nil {
			return nil, err
		}
	}
	return ids, nil
}

func uniqueEdges(edges []Edge, nodes map[uuid.UUID]struct{}) error {
	ids := make(map[uuid.UUID]struct{}, len(edges))
	bySourceLabel := make(map[string]uuid.UUID, len(edges))

	for _, e := range edges {
		if e.ID == uuid.Nil {
			return fmt.Errorf("every edge must include a client-generated id")
		}
		if _, exists := ids[e.ID]; exists {
			return fmt.Errorf("duplicate edge id %s. Each edge id must be unique in this graph", e.ID)
		}
		ids[e.ID] = struct{}{}

		if _, ok := nodes[e.FromNodeID]; !ok {
			return fmt.Errorf("edge %s references from_node_id %s which is not in this graph. Add that node to nodes, or remove/retarget the edge", e.ID, e.FromNodeID)
		}
		if _, ok := nodes[e.ToNodeID]; !ok {
			return fmt.Errorf("edge %s references to_node_id %s which is not in this graph. Add that node to nodes, or remove/retarget the edge", e.ID, e.ToNodeID)
		}

		label := edgeLabel(e.Label)
		if !validEdgeLabel(label) {
			return fmt.Errorf("edge %s has invalid label %q. Use default, true, or false", e.ID, e.Label)
		}

		key := e.FromNodeID.String() + ":" + string(label)
		if other, exists := bySourceLabel[key]; exists {
			return fmt.Errorf("node %s already has an outgoing edge with label %q (edges %s and %s). Each node may have only one edge per label (default, or true/false on a Conditional). Change the label, remove the extra edge, or pick a different source", e.FromNodeID, label, other, e.ID)
		}
		bySourceLabel[key] = e.ID
	}
	return nil
}

func edgeLabel(raw string) sqlcdb.EdgeLabel {
	if raw == "" {
		return sqlcdb.EdgeLabelDefault
	}
	return sqlcdb.EdgeLabel(raw)
}

func validEdgeLabel(label sqlcdb.EdgeLabel) bool {
	switch label {
	case sqlcdb.EdgeLabelDefault, sqlcdb.EdgeLabelTrue, sqlcdb.EdgeLabelFalse:
		return true
	default:
		return false
	}
}

func getDetail(ctx context.Context, q *sqlcdb.Queries, id uuid.UUID, published bool) (WorkflowDetail, error) {
	row, err := q.GetWorkflowDetail(ctx, sqlcdb.GetWorkflowDetailParams{
		Published: published,
		ID:        id,
	})
	if err != nil {
		return WorkflowDetail{}, err
	}
	return detailFromRow(row)
}

func detailFromRow(row sqlcdb.GetWorkflowDetailRow) (WorkflowDetail, error) {
	nodes, err := decodeJSON(row.Nodes, []Node{})
	if err != nil {
		return WorkflowDetail{}, fmt.Errorf("decode nodes: %w", err)
	}
	edges, err := decodeJSON(row.Edges, []Edge{})
	if err != nil {
		return WorkflowDetail{}, fmt.Errorf("decode edges: %w", err)
	}

	return WorkflowDetail{
		Workflow: Workflow{
			ID:                       row.ID,
			Name:                     row.Name,
			Description:              row.Description,
			Status:                   string(row.Status),
			LatestVersionID:          row.LatestVersionID,
			LatestPublishedVersionID: row.LatestPublishedVersionID,
			CreatedAt:                row.CreatedAt,
			UpdatedAt:                row.UpdatedAt,
			LastPublishedAt:          row.LastPublishedAt,
		},
		Graph: VersionGraph{
			ID:          row.GraphID,
			Version:     int(row.GraphVersion),
			PublishedAt: row.GraphPublishedAt,
			Nodes:       nodes,
			Edges:       edges,
		},
	}, nil
}

func decodeJSON[T any](raw json.RawMessage, fallback T) (T, error) {
	if len(raw) == 0 {
		return fallback, nil
	}
	var out T
	if err := json.Unmarshal(raw, &out); err != nil {
		return fallback, err
	}
	return out, nil
}
