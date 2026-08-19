package workflow

import (
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestValidatePublishGraph(t *testing.T) {
	start := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	api := uuid.MustParse("22222222-2222-4222-8222-222222222222")
	resp := uuid.MustParse("33333333-3333-4333-8333-333333333333")
	cond := uuid.MustParse("44444444-4444-4444-8444-444444444444")
	respFalse := uuid.MustParse("55555555-5555-4555-8555-555555555555")
	fnA := uuid.MustParse("66666666-6666-4666-8666-666666666666")
	fnB := uuid.MustParse("77777777-7777-4777-8777-777777777777")

	startNode := node(start, "start", 0, 0, 1, 1)
	apiNode := node(api, "api", 1, 1, 1, 1)
	respNode := node(resp, "response", 1, 1, 0, 0)
	condNode := node(cond, "conditional", 1, 1, 2, 2)
	respFalseNode := node(respFalse, "response", 1, 1, 0, 0)
	fnANode := node(fnA, "function", 1, 1, 1, 1)
	fnBNode := node(fnB, "function", 1, 1, 1, 1)

	tests := []struct {
		name    string
		nodes   []publishNode
		edges   []publishEdge
		wantErr string
	}{
		{
			name:    "empty",
			nodes:   nil,
			wantErr: "workflow graph is empty",
		},
		{
			name:  "valid linear",
			nodes: []publishNode{startNode, apiNode, respNode},
			edges: []publishEdge{
				edge(start, api, edgeLabelDefault),
				edge(api, resp, edgeLabelDefault),
			},
		},
		{
			name:  "valid conditional",
			nodes: []publishNode{startNode, condNode, respNode, respFalseNode},
			edges: []publishEdge{
				edge(start, cond, edgeLabelDefault),
				edge(cond, resp, edgeLabelTrue),
				edge(cond, respFalse, edgeLabelFalse),
			},
		},
		{
			name:  "missing start",
			nodes: []publishNode{apiNode, respNode},
			edges: []publishEdge{
				edge(api, resp, edgeLabelDefault),
			},
			wantErr: "exactly one start node",
		},
		{
			name:    "bad degree",
			nodes:   []publishNode{startNode, respNode},
			edges:   nil,
			wantErr: "out-degree",
		},
		{
			name:  "conditional wrong labels",
			nodes: []publishNode{startNode, condNode, respNode, respFalseNode},
			edges: []publishEdge{
				edge(start, cond, edgeLabelDefault),
				edge(cond, resp, edgeLabelDefault),
				edge(cond, respFalse, edgeLabelTrue),
			},
			wantErr: "true and false",
		},
		{
			// Degree-valid island A↔B is a cycle (and unreachable from Start).
			name:  "cycle",
			nodes: []publishNode{startNode, apiNode, respNode, fnANode, fnBNode},
			edges: []publishEdge{
				edge(start, api, edgeLabelDefault),
				edge(api, resp, edgeLabelDefault),
				edge(fnA, fnB, edgeLabelDefault),
				edge(fnB, fnA, edgeLabelDefault),
			},
			wantErr: "cycle",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePublishGraph(tt.nodes, tt.edges)
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("expected error containing %q", tt.wantErr)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("error %q does not contain %q", err.Error(), tt.wantErr)
			}
		})
	}
}

func node(id uuid.UUID, typ string, minIn, maxIn, minOut, maxOut int) publishNode {
	return publishNode{
		ID:           id,
		Type:         typ,
		MinInDegree:  minIn,
		MaxInDegree:  maxIn,
		MinOutDegree: minOut,
		MaxOutDegree: maxOut,
	}
}

func edge(from, to uuid.UUID, label string) publishEdge {
	return publishEdge{From: from, To: to, Label: label}
}
