package dag

import (
	"testing"

	"github.com/lavianrose/flowforge/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidator_Validate_Success(t *testing.T) {
	v := NewValidator()

	def := models.WorkflowDef{
		Nodes: []models.WorkflowNode{
			{
				ID:   "node1",
				Type: "http",
				Name: "HTTP Request",
				Config: map[string]interface{}{
					"url":    "https://api.example.com",
					"method": "GET",
				},
			},
		},
		Edges: []models.WorkflowEdge{},
	}

	err := v.Validate(def)
	assert.NoError(t, err)
}

func TestValidator_Validate_NoNodes(t *testing.T) {
	v := NewValidator()

	def := models.WorkflowDef{
		Nodes: []models.WorkflowNode{},
		Edges: []models.WorkflowEdge{},
	}

	err := v.Validate(def)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "at least one node")
}

func TestValidator_Validate_MissingNodeID(t *testing.T) {
	v := NewValidator()

	def := models.WorkflowDef{
		Nodes: []models.WorkflowNode{
			{
				Type: "http",
				Name: "HTTP Request",
				Config: map[string]interface{}{
					"url":    "https://api.example.com",
					"method": "GET",
				},
			},
		},
		Edges: []models.WorkflowEdge{},
	}

	err := v.Validate(def)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing ID")
}

func TestValidator_Validate_MissingNodeType(t *testing.T) {
	v := NewValidator()

	def := models.WorkflowDef{
		Nodes: []models.WorkflowNode{
			{
				ID:   "node1",
				Name: "Node",
				Config: map[string]interface{}{
					"url": "https://api.example.com",
				},
			},
		},
		Edges: []models.WorkflowEdge{},
	}

	err := v.Validate(def)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing type")
}

func TestValidator_Validate_MissingNodeName(t *testing.T) {
	v := NewValidator()

	def := models.WorkflowDef{
		Nodes: []models.WorkflowNode{
			{
				ID:   "node1",
				Type: "http",
				Config: map[string]interface{}{
					"url":    "https://api.example.com",
					"method": "GET",
				},
			},
		},
		Edges: []models.WorkflowEdge{},
	}

	err := v.Validate(def)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing name")
}

func TestValidator_Validate_DuplicateNodeID(t *testing.T) {
	v := NewValidator()

	def := models.WorkflowDef{
		Nodes: []models.WorkflowNode{
			{
				ID:   "node1",
				Type: "http",
				Name: "HTTP 1",
				Config: map[string]interface{}{
					"url":    "https://api.example.com",
					"method": "GET",
				},
			},
			{
				ID:   "node1",
				Type: "delay",
				Name: "Delay",
				Config: map[string]interface{}{
					"seconds": 5,
				},
			},
		},
		Edges: []models.WorkflowEdge{},
	}

	err := v.Validate(def)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "duplicate node ID")
}

func TestValidator_Validate_InvalidNodeType(t *testing.T) {
	v := NewValidator()

	def := models.WorkflowDef{
		Nodes: []models.WorkflowNode{
			{
				ID:   "node1",
				Type: "invalid_type",
				Name: "Invalid Node",
				Config: map[string]interface{}{
					"some": "config",
				},
			},
		},
		Edges: []models.WorkflowEdge{},
	}

	err := v.Validate(def)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid type")
}

func TestValidator_Validate_HTTPNode_MissingConfig(t *testing.T) {
	v := NewValidator()

	def := models.WorkflowDef{
		Nodes: []models.WorkflowNode{
			{
				ID:   "node1",
				Type: "http",
				Name: "HTTP Request",
			},
		},
		Edges: []models.WorkflowEdge{},
	}

	err := v.Validate(def)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "requires config")
}

func TestValidator_Validate_HTTPNode_MissingURL(t *testing.T) {
	v := NewValidator()

	def := models.WorkflowDef{
		Nodes: []models.WorkflowNode{
			{
				ID:   "node1",
				Type: "http",
				Name: "HTTP Request",
				Config: map[string]interface{}{
					"method": "GET",
				},
			},
		},
		Edges: []models.WorkflowEdge{},
	}

	err := v.Validate(def)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "requires 'url'")
}

func TestValidator_Validate_HTTPNode_MissingMethod(t *testing.T) {
	v := NewValidator()

	def := models.WorkflowDef{
		Nodes: []models.WorkflowNode{
			{
				ID:   "node1",
				Type: "http",
				Name: "HTTP Request",
				Config: map[string]interface{}{
					"url": "https://api.example.com",
				},
			},
		},
		Edges: []models.WorkflowEdge{},
	}

	err := v.Validate(def)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "requires 'method'")
}

func TestValidator_Validate_DelayNode_MissingConfig(t *testing.T) {
	v := NewValidator()

	def := models.WorkflowDef{
		Nodes: []models.WorkflowNode{
			{
				ID:   "node1",
				Type: "delay",
				Name: "Delay",
			},
		},
		Edges: []models.WorkflowEdge{},
	}

	err := v.Validate(def)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "requires config")
}

func TestValidator_Validate_DelayNode_MissingSeconds(t *testing.T) {
	v := NewValidator()

	def := models.WorkflowDef{
		Nodes: []models.WorkflowNode{
			{
				ID:   "node1",
				Type: "delay",
				Name: "Delay",
				Config: map[string]interface{}{
					"minutes": 5,
				},
			},
		},
		Edges: []models.WorkflowEdge{},
	}

	err := v.Validate(def)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "requires 'seconds'")
}

func TestValidator_Validate_ScriptNode_MissingConfig(t *testing.T) {
	v := NewValidator()

	def := models.WorkflowDef{
		Nodes: []models.WorkflowNode{
			{
				ID:   "node1",
				Type: "script",
				Name: "Script",
			},
		},
		Edges: []models.WorkflowEdge{},
	}

	err := v.Validate(def)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "requires config")
}

func TestValidator_Validate_ConditionNode_MissingConfig(t *testing.T) {
	v := NewValidator()

	def := models.WorkflowDef{
		Nodes: []models.WorkflowNode{
			{
				ID:   "node1",
				Type: "condition",
				Name: "Condition",
			},
		},
		Edges: []models.WorkflowEdge{},
	}

	err := v.Validate(def)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "requires config")
}

func TestValidator_Validate_Edge_MissingID(t *testing.T) {
	v := NewValidator()

	def := models.WorkflowDef{
		Nodes: []models.WorkflowNode{
			{
				ID:   "node1",
				Type: "http",
				Name: "HTTP",
				Config: map[string]interface{}{
					"url":    "https://api.example.com",
					"method": "GET",
				},
			},
		},
		Edges: []models.WorkflowEdge{
			{
				Source: "node1",
				Target: "node1",
			},
		},
	}

	err := v.Validate(def)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing ID")
}

func TestValidator_Validate_Edge_MissingSource(t *testing.T) {
	v := NewValidator()

	def := models.WorkflowDef{
		Nodes: []models.WorkflowNode{
			{
				ID:   "node1",
				Type: "http",
				Name: "HTTP",
				Config: map[string]interface{}{
					"url":    "https://api.example.com",
					"method": "GET",
				},
			},
		},
		Edges: []models.WorkflowEdge{
			{
				ID:     "edge1",
				Target: "node1",
			},
		},
	}

	err := v.Validate(def)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing source")
}

func TestValidator_Validate_Edge_MissingTarget(t *testing.T) {
	v := NewValidator()

	def := models.WorkflowDef{
		Nodes: []models.WorkflowNode{
			{
				ID:   "node1",
				Type: "http",
				Name: "HTTP",
				Config: map[string]interface{}{
					"url":    "https://api.example.com",
					"method": "GET",
				},
			},
		},
		Edges: []models.WorkflowEdge{
			{
				ID:     "edge1",
				Source: "node1",
			},
		},
	}

	err := v.Validate(def)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing target")
}

func TestValidator_Validate_SelfLoop(t *testing.T) {
	v := NewValidator()

	def := models.WorkflowDef{
		Nodes: []models.WorkflowNode{
			{
				ID:   "node1",
				Type: "http",
				Name: "HTTP",
				Config: map[string]interface{}{
					"url":    "https://api.example.com",
					"method": "GET",
				},
			},
		},
		Edges: []models.WorkflowEdge{
			{
				ID:     "edge1",
				Source: "node1",
				Target: "node1",
			},
		},
	}

	err := v.Validate(def)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "self-loop")
}

func TestValidator_Validate_Edge_SourceNotExist(t *testing.T) {
	v := NewValidator()

	def := models.WorkflowDef{
		Nodes: []models.WorkflowNode{
			{
				ID:   "node1",
				Type: "http",
				Name: "HTTP",
				Config: map[string]interface{}{
					"url":    "https://api.example.com",
					"method": "GET",
				},
			},
		},
		Edges: []models.WorkflowEdge{
			{
				ID:     "edge1",
				Source: "nonexistent",
				Target: "node1",
			},
		},
	}

	err := v.Validate(def)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "source node 'nonexistent' does not exist")
}

func TestValidator_Validate_Edge_TargetNotExist(t *testing.T) {
	v := NewValidator()

	def := models.WorkflowDef{
		Nodes: []models.WorkflowNode{
			{
				ID:   "node1",
				Type: "http",
				Name: "HTTP",
				Config: map[string]interface{}{
					"url":    "https://api.example.com",
					"method": "GET",
				},
			},
		},
		Edges: []models.WorkflowEdge{
			{
				ID:     "edge1",
				Source: "node1",
				Target: "nonexistent",
			},
		},
	}

	err := v.Validate(def)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "target node 'nonexistent' does not exist")
}

func TestValidator_Validate_DuplicateEdge(t *testing.T) {
	v := NewValidator()

	def := models.WorkflowDef{
		Nodes: []models.WorkflowNode{
			{
				ID:   "node1",
				Type: "http",
				Name: "HTTP 1",
				Config: map[string]interface{}{
					"url":    "https://api.example.com",
					"method": "GET",
				},
			},
			{
				ID:   "node2",
				Type: "delay",
				Name: "Delay",
				Config: map[string]interface{}{
					"seconds": 5,
				},
			},
		},
		Edges: []models.WorkflowEdge{
			{
				ID:     "edge1",
				Source: "node1",
				Target: "node2",
			},
			{
				ID:     "edge2",
				Source: "node1",
				Target: "node2",
			},
		},
	}

	err := v.Validate(def)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "duplicate edge")
}

func TestValidator_DetectCycles_SimpleCycle(t *testing.T) {
	v := NewValidator()

	def := models.WorkflowDef{
		Nodes: []models.WorkflowNode{
			{ID: "A", Type: "http", Name: "A", Config: map[string]interface{}{"url": "http://a.com", "method": "GET"}},
			{ID: "B", Type: "http", Name: "B", Config: map[string]interface{}{"url": "http://b.com", "method": "GET"}},
			{ID: "C", Type: "http", Name: "C", Config: map[string]interface{}{"url": "http://c.com", "method": "GET"}},
		},
		Edges: []models.WorkflowEdge{
			{ID: "e1", Source: "A", Target: "B"},
			{ID: "e2", Source: "B", Target: "C"},
			{ID: "e3", Source: "C", Target: "A"},
		},
	}

	err := v.Validate(def)
	assert.Error(t, err)
	assert.Equal(t, ErrCycleDetected, err)
}

func TestValidator_DetectCycles_NoCycle(t *testing.T) {
	v := NewValidator()

	def := models.WorkflowDef{
		Nodes: []models.WorkflowNode{
			{ID: "A", Type: "http", Name: "A", Config: map[string]interface{}{"url": "http://a.com", "method": "GET"}},
			{ID: "B", Type: "http", Name: "B", Config: map[string]interface{}{"url": "http://b.com", "method": "GET"}},
			{ID: "C", Type: "http", Name: "C", Config: map[string]interface{}{"url": "http://c.com", "method": "GET"}},
		},
		Edges: []models.WorkflowEdge{
			{ID: "e1", Source: "A", Target: "B"},
			{ID: "e2", Source: "B", Target: "C"},
		},
	}

	err := v.Validate(def)
	assert.NoError(t, err)
}

func TestValidator_TopologicalSort_Simple(t *testing.T) {
	v := NewValidator()

	def := models.WorkflowDef{
		Nodes: []models.WorkflowNode{
			{ID: "A", Type: "http", Name: "A", Config: map[string]interface{}{"url": "http://a.com", "method": "GET"}},
			{ID: "B", Type: "http", Name: "B", Config: map[string]interface{}{"url": "http://b.com", "method": "GET"}},
			{ID: "C", Type: "http", Name: "C", Config: map[string]interface{}{"url": "http://c.com", "method": "GET"}},
		},
		Edges: []models.WorkflowEdge{
			{ID: "e1", Source: "A", Target: "B"},
			{ID: "e2", Source: "B", Target: "C"},
		},
	}

	result, err := v.TopologicalSort(def)
	require.NoError(t, err)

	// A should come before B, B before C
	posA := indexOf(result, "A")
	posB := indexOf(result, "B")
	posC := indexOf(result, "C")

	assert.Less(t, posA, posB, "A should come before B")
	assert.Less(t, posB, posC, "B should come before C")
}

func TestValidator_TopologicalSort_Complex(t *testing.T) {
	v := NewValidator()

	def := models.WorkflowDef{
		Nodes: []models.WorkflowNode{
			{ID: "A", Type: "http", Name: "A", Config: map[string]interface{}{"url": "http://a.com", "method": "GET"}},
			{ID: "B", Type: "http", Name: "B", Config: map[string]interface{}{"url": "http://b.com", "method": "GET"}},
			{ID: "C", Type: "http", Name: "C", Config: map[string]interface{}{"url": "http://c.com", "method": "GET"}},
			{ID: "D", Type: "http", Name: "D", Config: map[string]interface{}{"url": "http://d.com", "method": "GET"}},
		},
		Edges: []models.WorkflowEdge{
			{ID: "e1", Source: "A", Target: "B"},
			{ID: "e2", Source: "A", Target: "C"},
			{ID: "e3", Source: "B", Target: "D"},
			{ID: "e4", Source: "C", Target: "D"},
		},
	}

	result, err := v.TopologicalSort(def)
	require.NoError(t, err)

	// A should come first
	assert.Equal(t, "A", result[0])

	// D should come last
	assert.Equal(t, "D", result[len(result)-1])

	// B and C should come before D
	posB := indexOf(result, "B")
	posC := indexOf(result, "C")
	posD := indexOf(result, "D")

	assert.Less(t, posB, posD, "B should come before D")
	assert.Less(t, posC, posD, "C should come before D")
}

func TestValidator_TopologicalSort_Cycle(t *testing.T) {
	v := NewValidator()

	def := models.WorkflowDef{
		Nodes: []models.WorkflowNode{
			{ID: "A", Type: "http", Name: "A", Config: map[string]interface{}{"url": "http://a.com", "method": "GET"}},
			{ID: "B", Type: "http", Name: "B", Config: map[string]interface{}{"url": "http://b.com", "method": "GET"}},
		},
		Edges: []models.WorkflowEdge{
			{ID: "e1", Source: "A", Target: "B"},
			{ID: "e2", Source: "B", Target: "A"},
		},
	}

	result, err := v.TopologicalSort(def)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, ErrCycleDetected, err)
}

func TestValidator_GetExecutionLevels_Simple(t *testing.T) {
	v := NewValidator()

	def := models.WorkflowDef{
		Nodes: []models.WorkflowNode{
			{ID: "A", Type: "http", Name: "A", Config: map[string]interface{}{"url": "http://a.com", "method": "GET"}},
			{ID: "B", Type: "http", Name: "B", Config: map[string]interface{}{"url": "http://b.com", "method": "GET"}},
			{ID: "C", Type: "http", Name: "C", Config: map[string]interface{}{"url": "http://c.com", "method": "GET"}},
		},
		Edges: []models.WorkflowEdge{
			{ID: "e1", Source: "A", Target: "B"},
			{ID: "e2", Source: "B", Target: "C"},
		},
	}

	levels, err := v.GetExecutionLevels(def)
	require.NoError(t, err)

	assert.Equal(t, 3, len(levels))
	assert.ElementsMatch(t, []string{"A"}, levels[0])
	assert.ElementsMatch(t, []string{"B"}, levels[1])
	assert.ElementsMatch(t, []string{"C"}, levels[2])
}

func TestValidator_GetExecutionLevels_Parallel(t *testing.T) {
	v := NewValidator()

	def := models.WorkflowDef{
		Nodes: []models.WorkflowNode{
			{ID: "A", Type: "http", Name: "A", Config: map[string]interface{}{"url": "http://a.com", "method": "GET"}},
			{ID: "B", Type: "http", Name: "B", Config: map[string]interface{}{"url": "http://b.com", "method": "GET"}},
			{ID: "C", Type: "http", Name: "C", Config: map[string]interface{}{"url": "http://c.com", "method": "GET"}},
			{ID: "D", Type: "http", Name: "D", Config: map[string]interface{}{"url": "http://d.com", "method": "GET"}},
		},
		Edges: []models.WorkflowEdge{
			{ID: "e1", Source: "A", Target: "B"},
			{ID: "e2", Source: "A", Target: "C"},
			{ID: "e3", Source: "B", Target: "D"},
			{ID: "e4", Source: "C", Target: "D"},
		},
	}

	levels, err := v.GetExecutionLevels(def)
	require.NoError(t, err)

	// Level 0: A (no dependencies)
	assert.Equal(t, 1, len(levels[0]))
	assert.Contains(t, levels[0], "A")

	// Level 1: B and C (depend on A, can run in parallel)
	assert.Equal(t, 2, len(levels[1]))
	assert.Contains(t, levels[1], "B")
	assert.Contains(t, levels[1], "C")

	// Level 2: D (depends on B and C)
	assert.Equal(t, 1, len(levels[2]))
	assert.Contains(t, levels[2], "D")
}

func TestValidator_GetExecutionLevels_SingleNode(t *testing.T) {
	v := NewValidator()

	def := models.WorkflowDef{
		Nodes: []models.WorkflowNode{
			{ID: "A", Type: "http", Name: "A", Config: map[string]interface{}{"url": "http://a.com", "method": "GET"}},
		},
		Edges: []models.WorkflowEdge{},
	}

	levels, err := v.GetExecutionLevels(def)
	require.NoError(t, err)

	assert.Equal(t, 1, len(levels))
	assert.ElementsMatch(t, []string{"A"}, levels[0])
}

// Helper function
func indexOf(slice []string, item string) int {
	for i, s := range slice {
		if s == item {
			return i
		}
	}
	return -1
}
