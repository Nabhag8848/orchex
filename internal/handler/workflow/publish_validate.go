package workflow

import (
	"fmt"

	"github.com/google/uuid"
)

const nodeTypeStart = "start"
const nodeTypeConditional = "conditional"

const (
	edgeLabelDefault = "default"
	edgeLabelTrue    = "true"
	edgeLabelFalse   = "false"
)

type graphValidationError struct {
	msg string
}

func (e *graphValidationError) Error() string {
	return e.msg
}

func invalidGraph(format string, args ...any) error {
	return &graphValidationError{msg: fmt.Sprintf(format, args...)}
}

type publishNode struct {
	ID           uuid.UUID
	Type         string
	MinInDegree  int
	MaxInDegree  int
	MinOutDegree int
	MaxOutDegree int
}

type publishEdge struct {
	From  uuid.UUID
	To    uuid.UUID
	Label string
}

type publishGraph struct {
	nodes     []publishNode
	byID      map[uuid.UUID]publishNode
	adj       map[uuid.UUID][]uuid.UUID
	inDegree  map[uuid.UUID]int
	outDegree map[uuid.UUID]int
	outLabels map[uuid.UUID][]string
}

func validatePublishGraph(nodes []publishNode, edges []publishEdge) error {
	if len(nodes) == 0 {
		return invalidGraph("workflow graph is empty. Add nodes before publishing")
	}

	g := newPublishGraph(nodes, edges)

	startID, err := uniqueStart(g)
	if err != nil {
		return err
	}
	if err := validateDegrees(g); err != nil {
		return err
	}
	if err := validateEdgeLabels(g); err != nil {
		return err
	}
	if hasCycleKahn(g) {
		return invalidGraph("workflow contains a cycle")
	}
	if err := reachableFromStart(g, startID); err != nil {
		return err
	}
	return nil
}

func newPublishGraph(nodes []publishNode, edges []publishEdge) *publishGraph {
	g := &publishGraph{
		nodes:     nodes,
		byID:      make(map[uuid.UUID]publishNode, len(nodes)),
		adj:       make(map[uuid.UUID][]uuid.UUID, len(nodes)),
		inDegree:  make(map[uuid.UUID]int, len(nodes)),
		outDegree: make(map[uuid.UUID]int, len(nodes)),
		outLabels: make(map[uuid.UUID][]string, len(nodes)),
	}
	for _, n := range nodes {
		g.byID[n.ID] = n
		g.adj[n.ID] = nil
		g.inDegree[n.ID] = 0
		g.outDegree[n.ID] = 0
	}
	for _, e := range edges {
		if _, ok := g.byID[e.From]; !ok {
			continue
		}
		if _, ok := g.byID[e.To]; !ok {
			continue
		}
		g.adj[e.From] = append(g.adj[e.From], e.To)
		g.outDegree[e.From]++
		g.inDegree[e.To]++
		g.outLabels[e.From] = append(g.outLabels[e.From], e.Label)
	}
	return g
}

func uniqueStart(g *publishGraph) (uuid.UUID, error) {
	var starts []uuid.UUID
	for _, n := range g.nodes {
		if n.Type == nodeTypeStart {
			starts = append(starts, n.ID)
		}
	}
	switch len(starts) {
	case 0:
		return uuid.Nil, invalidGraph("workflow must have exactly one start node")
	case 1:
		return starts[0], nil
	default:
		return uuid.Nil, invalidGraph("workflow must have exactly one start node, found %d", len(starts))
	}
}

func validateDegrees(g *publishGraph) error {
	for _, n := range g.nodes {
		in := g.inDegree[n.ID]
		out := g.outDegree[n.ID]
		if in < n.MinInDegree || in > n.MaxInDegree {
			return invalidGraph(
				"node %s (%s): in-degree must be between %d and %d, got %d",
				n.ID, n.Type, n.MinInDegree, n.MaxInDegree, in,
			)
		}
		if out < n.MinOutDegree || out > n.MaxOutDegree {
			return invalidGraph(
				"node %s (%s): out-degree must be between %d and %d, got %d",
				n.ID, n.Type, n.MinOutDegree, n.MaxOutDegree, out,
			)
		}
	}
	return nil
}

func validateEdgeLabels(g *publishGraph) error {
	for _, n := range g.nodes {
		labels := g.outLabels[n.ID]
		switch n.Type {
		case nodeTypeConditional:
			if !hasLabels(labels, edgeLabelTrue, edgeLabelFalse) {
				return invalidGraph(
					"conditional node %s must have outgoing labels true and false",
					n.ID,
				)
			}
		default:
			for _, label := range labels {
				if label != edgeLabelDefault {
					return invalidGraph(
						"node %s (%s) must use label %q, got %q",
						n.ID, n.Type, edgeLabelDefault, label,
					)
				}
			}
		}
	}
	return nil
}

func hasLabels(got []string, want ...string) bool {
	if len(got) != len(want) {
		return false
	}
	seen := make(map[string]struct{}, len(got))
	for _, label := range got {
		seen[label] = struct{}{}
	}
	for _, label := range want {
		if _, ok := seen[label]; !ok {
			return false
		}
	}
	return true
}

// hasCycleKahn peels in-degree 0 nodes (Kahn). Leftover nodes mean a cycle.
func hasCycleKahn(g *publishGraph) bool {
	in := make(map[uuid.UUID]int, len(g.byID))
	for id, deg := range g.inDegree {
		in[id] = deg
	}

	queue := make([]uuid.UUID, 0)
	for id, deg := range in {
		if deg == 0 {
			queue = append(queue, id)
		}
	}

	removed := 0
	for len(queue) > 0 {
		u := queue[0]
		queue = queue[1:]
		removed++
		for _, v := range g.adj[u] {
			in[v]--
			if in[v] == 0 {
				queue = append(queue, v)
			}
		}
	}
	return removed != len(g.byID)
}

func reachableFromStart(g *publishGraph, start uuid.UUID) error {
	visited := make(map[uuid.UUID]struct{}, len(g.byID))
	queue := []uuid.UUID{start}
	visited[start] = struct{}{}

	for len(queue) > 0 {
		u := queue[0]
		queue = queue[1:]
		for _, v := range g.adj[u] {
			if _, ok := visited[v]; ok {
				continue
			}
			visited[v] = struct{}{}
			queue = append(queue, v)
		}
	}

	if len(visited) == len(g.byID) {
		return nil
	}
	for _, n := range g.nodes {
		if _, ok := visited[n.ID]; !ok {
			return invalidGraph("node %s is not reachable from start", n.ID)
		}
	}
	return nil
}
