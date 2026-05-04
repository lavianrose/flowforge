package dag

import (
	"errors"
	"fmt"

	"github.com/lavianrose/flowforge/internal/models"
)

var (
	ErrCycleDetected     = errors.New("cycle detected in workflow")
	ErrInvalidNode       = errors.New("invalid node definition")
	ErrInvalidEdge       = errors.New("invalid edge definition")
	ErrDuplicateEdge     = errors.New("duplicate edge detected")
	ErrSelfLoop          = errors.New("self-loop detected")
	ErrDisconnectedGraph = errors.New("disconnected graph detected")
)

type Validator struct{}

func NewValidator() *Validator {
	return &Validator{}
}

// Validate validates the workflow definition
func (v *Validator) Validate(def models.WorkflowDef) error {
	if err := v.validateNodes(def.Nodes); err != nil {
		return err
	}

	if err := v.validateEdges(def.Edges, def.Nodes); err != nil {
		return err
	}

	if err := v.detectCycles(def); err != nil {
		return err
	}

	return nil
}

// validateNodes checks if all nodes have valid structure
func (v *Validator) validateNodes(nodes []models.WorkflowNode) error {
	if len(nodes) == 0 {
		return errors.New("workflow must have at least one node")
	}

	nodeIDs := make(map[string]bool)
	for i, node := range nodes {
		if node.ID == "" {
			return fmt.Errorf("node %d: missing ID", i)
		}

		if node.Type == "" {
			return fmt.Errorf("node %s: missing type", node.ID)
		}

		if node.Name == "" {
			return fmt.Errorf("node %s: missing name", node.ID)
		}

		// Check for duplicate node IDs
		if nodeIDs[node.ID] {
			return fmt.Errorf("duplicate node ID: %s", node.ID)
		}
		nodeIDs[node.ID] = true

		// Validate node type
		validTypes := map[string]bool{
			"http":      true,
			"delay":     true,
			"script":    true,
			"condition": true,
		}

		if !validTypes[node.Type] {
			return fmt.Errorf("node %s: invalid type '%s'", node.ID, node.Type)
		}

		// Validate configuration based on type
		if err := v.validateNodeConfig(&node); err != nil {
			return fmt.Errorf("node %s: %w", node.ID, err)
		}
	}

	return nil
}

// validateNodeConfig validates node configuration based on type
func (v *Validator) validateNodeConfig(node *models.WorkflowNode) error {
	switch node.Type {
	case "http":
		if node.Config == nil {
			return errors.New("http node requires config")
		}
		if _, ok := node.Config["url"]; !ok {
			return errors.New("http node requires 'url' in config")
		}
		if _, ok := node.Config["method"]; !ok {
			return errors.New("http node requires 'method' in config")
		}

	case "delay":
		if node.Config == nil {
			return errors.New("delay node requires config")
		}
		if _, ok := node.Config["seconds"]; !ok {
			return errors.New("delay node requires 'seconds' in config")
		}

	case "script":
		if node.Config == nil {
			return errors.New("script node requires config")
		}
		if _, ok := node.Config["code"]; !ok {
			return errors.New("script node requires 'code' in config")
		}
		if lang, ok := node.Config["language"].(string); ok && lang != "" {
			validLangs := map[string]bool{"python": true, "javascript": true, "template": true}
			if !validLangs[lang] {
				return fmt.Errorf("script node: unsupported language '%s'", lang)
			}
		}

	case "condition":
		if node.Config == nil {
			return errors.New("condition node requires config")
		}
		if _, ok := node.Config["expression"]; !ok {
			return errors.New("condition node requires 'expression' in config")
		}
	}

	return nil
}

// validateEdges checks if all edges are valid
func (v *Validator) validateEdges(edges []models.WorkflowEdge, nodes []models.WorkflowNode) error {
	nodeIDs := make(map[string]bool)
	for _, node := range nodes {
		nodeIDs[node.ID] = true
	}

	edgeKeys := make(map[string]bool)
	for i, edge := range edges {
		if edge.ID == "" {
			return fmt.Errorf("edge %d: missing ID", i)
		}

		if edge.Source == "" {
			return fmt.Errorf("edge %s: missing source", edge.ID)
		}

		if edge.Target == "" {
			return fmt.Errorf("edge %s: missing target", edge.ID)
		}

		// Check for self-loops
		if edge.Source == edge.Target {
			return fmt.Errorf("%w: %s", ErrSelfLoop, edge.ID)
		}

		// Check if source and target nodes exist
		if !nodeIDs[edge.Source] {
			return fmt.Errorf("edge %s: source node '%s' does not exist", edge.ID, edge.Source)
		}

		if !nodeIDs[edge.Target] {
			return fmt.Errorf("edge %s: target node '%s' does not exist", edge.ID, edge.Target)
		}

		// Check for duplicate edges
		key := fmt.Sprintf("%s->%s", edge.Source, edge.Target)
		if edgeKeys[key] {
			return fmt.Errorf("%w: %s", ErrDuplicateEdge, key)
		}
		edgeKeys[key] = true
	}

	return nil
}

// detectCycles checks if the workflow graph contains cycles using DFS
func (v *Validator) detectCycles(def models.WorkflowDef) error {
	// Build adjacency list
	adj := make(map[string][]string)
	inDegree := make(map[string]int)

	// Initialize all nodes
	for _, node := range def.Nodes {
		adj[node.ID] = []string{}
		inDegree[node.ID] = 0
	}

	// Build graph from edges
	for _, edge := range def.Edges {
		adj[edge.Source] = append(adj[edge.Source], edge.Target)
		inDegree[edge.Target]++
	}

	// Kahn's algorithm for cycle detection
	queue := make([]string, 0)
	for nodeID, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, nodeID)
		}
	}

	visitedCount := 0
	for len(queue) > 0 {
		node := queue[0]
		queue = queue[1:]
		visitedCount++

		for _, neighbor := range adj[node] {
			inDegree[neighbor]--
			if inDegree[neighbor] == 0 {
				queue = append(queue, neighbor)
			}
		}
	}

	// If we haven't visited all nodes, there's a cycle
	if visitedCount != len(def.Nodes) {
		return ErrCycleDetected
	}

	return nil
}

// TopologicalSort returns nodes in topological order
func (v *Validator) TopologicalSort(def models.WorkflowDef) ([]string, error) {
	// Build adjacency list
	adj := make(map[string][]string)
	inDegree := make(map[string]int)

	// Initialize all nodes
	for _, node := range def.Nodes {
		adj[node.ID] = []string{}
		inDegree[node.ID] = 0
	}

	// Build graph from edges
	for _, edge := range def.Edges {
		adj[edge.Source] = append(adj[edge.Source], edge.Target)
		inDegree[edge.Target]++
	}

	// Kahn's algorithm for topological sort
	queue := make([]string, 0)
	for nodeID, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, nodeID)
		}
	}

	result := make([]string, 0)
	for len(queue) > 0 {
		node := queue[0]
		queue = queue[1:]
		result = append(result, node)

		for _, neighbor := range adj[node] {
			inDegree[neighbor]--
			if inDegree[neighbor] == 0 {
				queue = append(queue, neighbor)
			}
		}
	}

	// If we haven't visited all nodes, there's a cycle
	if len(result) != len(def.Nodes) {
		return nil, ErrCycleDetected
	}

	return result, nil
}

// GetExecutionLevels returns nodes grouped by execution level (can run in parallel)
func (v *Validator) GetExecutionLevels(def models.WorkflowDef) ([][]string, error) {
	order, err := v.TopologicalSort(def)
	if err != nil {
		return nil, err
	}

	// Build adjacency list and reverse adjacency list
	adj := make(map[string][]string)
	inDegree := make(map[string]int)
	maxDist := make(map[string]int)

	for _, node := range def.Nodes {
		adj[node.ID] = []string{}
		inDegree[node.ID] = 0
		maxDist[node.ID] = 0
	}

	for _, edge := range def.Edges {
		adj[edge.Source] = append(adj[edge.Source], edge.Target)
		inDegree[edge.Target]++
	}

	// Calculate longest path for each node
	for _, node := range order {
		for _, neighbor := range adj[node] {
			if maxDist[neighbor] < maxDist[node]+1 {
				maxDist[neighbor] = maxDist[node] + 1
			}
		}
	}

	// Group by level
	levels := make(map[int][]string)
	maxLevel := 0
	for nodeID, level := range maxDist {
		levels[level] = append(levels[level], nodeID)
		if level > maxLevel {
			maxLevel = level
		}
	}

	// Build result
	result := make([][]string, maxLevel+1)
	for i := 0; i <= maxLevel; i++ {
		result[i] = levels[i]
	}

	return result, nil
}
