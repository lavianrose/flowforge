package execution

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/lavianrose/flowforge/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// engineWithNilRepos creates an Engine with nil repos for testing node logic
// (node logic methods don't use repos).
func engineWithNilRepos() *Engine {
	return &Engine{}
}

// ---------------------------------------------------------------------------
// resolveTemplate
// ---------------------------------------------------------------------------

func TestResolveTemplate_NoPlaceholders(t *testing.T) {
	e := engineWithNilRepos()
	result := e.resolveTemplate("hello world", nil)
	assert.Equal(t, "hello world", result)
}

func TestResolveTemplate_SinglePlaceholder(t *testing.T) {
	e := engineWithNilRepos()
	inputs := map[string]interface{}{
		"node1": map[string]interface{}{
			"status_code": float64(200),
		},
	}
	result := e.resolveTemplate("{{inputs.node1.status_code}}", inputs)
	assert.Equal(t, "200", result)
}

func TestResolveTemplate_MultiplePlaceholders(t *testing.T) {
	e := engineWithNilRepos()
	inputs := map[string]interface{}{
		"node1": map[string]interface{}{
			"status_code": float64(200),
			"body":        "ok",
		},
	}
	result := e.resolveTemplate("code={{inputs.node1.status_code}} body={{inputs.node1.body}}", inputs)
	assert.Equal(t, "code=200 body=ok", result)
}

func TestResolveTemplate_NestedPath(t *testing.T) {
	e := engineWithNilRepos()
	inputs := map[string]interface{}{
		"node1": map[string]interface{}{
			"json": map[string]interface{}{
				"name": "test",
				"count": float64(5),
			},
		},
	}
	result := e.resolveTemplate("{{inputs.node1.json.name}}", inputs)
	assert.Equal(t, "test", result)
}

func TestResolveTemplate_MissingPath(t *testing.T) {
	e := engineWithNilRepos()
	inputs := map[string]interface{}{
		"node1": map[string]interface{}{
			"status": "ok",
		},
	}
	result := e.resolveTemplate("{{inputs.node1.missing}}", inputs)
	assert.Equal(t, "", result)
}

func TestResolveTemplate_ObjectValue(t *testing.T) {
	e := engineWithNilRepos()
	inputs := map[string]interface{}{
		"node1": map[string]interface{}{
			"data": map[string]interface{}{
				"key": "value",
			},
		},
	}
	result := e.resolveTemplate("{{inputs.node1.data}}", inputs)
	assert.Contains(t, result, "key")
	assert.Contains(t, result, "value")
}

// ---------------------------------------------------------------------------
// evaluateExpression
// ---------------------------------------------------------------------------

func TestEvaluateExpression_True(t *testing.T) {
	result, err := evaluateExpression("true")
	require.NoError(t, err)
	assert.Equal(t, true, result)
}

func TestEvaluateExpression_False(t *testing.T) {
	result, err := evaluateExpression("false")
	require.NoError(t, err)
	assert.Equal(t, false, result)
}

func TestEvaluateExpression_NumberEquals(t *testing.T) {
	result, err := evaluateExpression("200 == 200")
	require.NoError(t, err)
	assert.Equal(t, true, result)
}

func TestEvaluateExpression_NumberNotEquals(t *testing.T) {
	result, err := evaluateExpression("200 != 404")
	require.NoError(t, err)
	assert.Equal(t, true, result)
}

func TestEvaluateExpression_NumberGreaterThan(t *testing.T) {
	result, err := evaluateExpression("10 > 5")
	require.NoError(t, err)
	assert.Equal(t, true, result)
}

func TestEvaluateExpression_NumberLessThan(t *testing.T) {
	result, err := evaluateExpression("5 < 10")
	require.NoError(t, err)
	assert.Equal(t, true, result)
}

func TestEvaluateExpression_NumberGreaterEqual(t *testing.T) {
	result, err := evaluateExpression("10 >= 10")
	require.NoError(t, err)
	assert.Equal(t, true, result)
}

func TestEvaluateExpression_NumberLessEqual(t *testing.T) {
	result, err := evaluateExpression("10 <= 10")
	require.NoError(t, err)
	assert.Equal(t, true, result)
}

func TestEvaluateExpression_StringEquals(t *testing.T) {
	result, err := evaluateExpression(`"hello" == "hello"`)
	require.NoError(t, err)
	assert.Equal(t, true, result)
}

func TestEvaluateExpression_StringNotEquals(t *testing.T) {
	result, err := evaluateExpression(`"hello" != "world"`)
	require.NoError(t, err)
	assert.Equal(t, true, result)
}

func TestEvaluateExpression_MixedComparisonFails(t *testing.T) {
	result, err := evaluateExpression("200 == 200")
	require.NoError(t, err)
	assert.Equal(t, true, result)

	result, err = evaluateExpression("200 == 201")
	require.NoError(t, err)
	assert.Equal(t, false, result)
}

// ---------------------------------------------------------------------------
// parseValue
// ---------------------------------------------------------------------------

func TestParseValue_QuotedString(t *testing.T) {
	val, err := parseValue(`"hello"`)
	require.NoError(t, err)
	assert.Equal(t, "hello", val)
}

func TestParseValue_SingleQuotedString(t *testing.T) {
	val, err := parseValue(`'hello'`)
	require.NoError(t, err)
	assert.Equal(t, "hello", val)
}

func TestParseValue_Number(t *testing.T) {
	val, err := parseValue("42")
	require.NoError(t, err)
	assert.Equal(t, float64(42), val)
}

func TestParseValue_BooleanTrue(t *testing.T) {
	val, err := parseValue("true")
	require.NoError(t, err)
	assert.Equal(t, true, val)
}

func TestParseValue_BooleanFalse(t *testing.T) {
	val, err := parseValue("false")
	require.NoError(t, err)
	assert.Equal(t, false, val)
}

func TestParseValue_PlainString(t *testing.T) {
	val, err := parseValue("some_text")
	require.NoError(t, err)
	assert.Equal(t, "some_text", val)
}

// ---------------------------------------------------------------------------
// toFloat64
// ---------------------------------------------------------------------------

func TestToFloat64_Float64(t *testing.T) {
	val, ok := toFloat64(float64(3.14))
	assert.True(t, ok)
	assert.Equal(t, 3.14, val)
}

func TestToFloat64_Int(t *testing.T) {
	val, ok := toFloat64(42)
	assert.True(t, ok)
	assert.Equal(t, float64(42), val)
}

func TestToFloat64_StringNumber(t *testing.T) {
	val, ok := toFloat64("42.5")
	assert.True(t, ok)
	assert.Equal(t, 42.5, val)
}

func TestToFloat64_InvalidString(t *testing.T) {
	_, ok := toFloat64("notanumber")
	assert.False(t, ok)
}

// ---------------------------------------------------------------------------
// flattenHeaders
// ---------------------------------------------------------------------------

func TestFlattenHeaders(t *testing.T) {
	h := http.Header{}
	h.Set("Content-Type", "application/json")
	h.Add("X-Custom", "value1")

	flat := flattenHeaders(h)
	assert.Equal(t, "application/json", flat["Content-Type"])
	assert.Equal(t, "value1", flat["X-Custom"])
}

// ---------------------------------------------------------------------------
// executeDelay
// ---------------------------------------------------------------------------

func TestExecuteDelay_Success(t *testing.T) {
	e := engineWithNilRepos()
	node := &models.WorkflowNode{
		ID:   "delay1",
		Type: "delay",
		Config: map[string]interface{}{
			"seconds": float64(1),
		},
	}

	start := time.Now()
	output, err := e.executeDelay(context.Background(), node)
	elapsed := time.Since(start)

	require.NoError(t, err)
	assert.Contains(t, output["message"], "Delayed 1 seconds")
	assert.True(t, elapsed >= time.Second, "should have waited at least 1 second")
}

func TestExecuteDelay_StringSeconds(t *testing.T) {
	e := engineWithNilRepos()
	node := &models.WorkflowNode{
		ID:   "delay1",
		Type: "delay",
		Config: map[string]interface{}{
			"seconds": "1",
		},
	}

	output, err := e.executeDelay(context.Background(), node)
	require.NoError(t, err)
	assert.Contains(t, output["message"], "Delayed 1 seconds")
}

func TestExecuteDelay_InvalidSeconds(t *testing.T) {
	e := engineWithNilRepos()
	node := &models.WorkflowNode{
		ID:   "delay1",
		Type: "delay",
		Config: map[string]interface{}{
			"seconds": "notanumber",
		},
	}

	_, err := e.executeDelay(context.Background(), node)
	assert.Error(t, err)
}

func TestExecuteDelay_ContextCancellation(t *testing.T) {
	e := engineWithNilRepos()
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	node := &models.WorkflowNode{
		ID:   "delay1",
		Type: "delay",
		Config: map[string]interface{}{
			"seconds": float64(60),
		},
	}

	_, err := e.executeDelay(ctx, node)
	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err)
}

// ---------------------------------------------------------------------------
// executeHTTP
// ---------------------------------------------------------------------------

func TestExecuteHTTP_GET(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok","count":42}`))
	}))
	defer server.Close()

	e := engineWithNilRepos()
	node := &models.WorkflowNode{
		ID:   "http1",
		Type: "http",
		Config: map[string]interface{}{
			"url":    server.URL,
			"method": "GET",
		},
	}

	output, err := e.executeHTTP(context.Background(), node, nil)
	require.NoError(t, err)

	assert.Equal(t, 200, output["status_code"])
	assert.Contains(t, output["body"], `"status":"ok"`)
	assert.NotNil(t, output["json"])

	jsonData := output["json"].(map[string]interface{})
	assert.Equal(t, "ok", jsonData["status"])
	assert.Equal(t, float64(42), jsonData["count"])
}

func TestExecuteHTTP_POST(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"id":"123"}`))
	}))
	defer server.Close()

	e := engineWithNilRepos()
	node := &models.WorkflowNode{
		ID:   "http1",
		Type: "http",
		Config: map[string]interface{}{
			"url":    server.URL,
			"method": "POST",
			"body":   `{"name":"test"}`,
		},
	}

	output, err := e.executeHTTP(context.Background(), node, nil)
	require.NoError(t, err)
	assert.Equal(t, 201, output["status_code"])
}

func TestExecuteHTTP_WithHeaders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "Bearer my-token", r.Header.Get("Authorization"))
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`ok`))
	}))
	defer server.Close()

	e := engineWithNilRepos()
	node := &models.WorkflowNode{
		ID:   "http1",
		Type: "http",
		Config: map[string]interface{}{
			"url":    server.URL,
			"method": "GET",
			"headers": map[string]interface{}{
				"Authorization": "Bearer my-token",
			},
		},
	}

	output, err := e.executeHTTP(context.Background(), node, nil)
	require.NoError(t, err)
	assert.Equal(t, 200, output["status_code"])
}

func TestExecuteHTTP_TemplateInURL(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"path":"` + r.URL.Path + `"}`))
	}))
	defer server.Close()

	e := engineWithNilRepos()
	node := &models.WorkflowNode{
		ID:   "http1",
		Type: "http",
		Config: map[string]interface{}{
			"url":    server.URL + "/{{inputs.prev.id}}",
			"method": "GET",
		},
	}

	inputs := map[string]interface{}{
		"prev": map[string]interface{}{
			"id": "123",
		},
	}

	output, err := e.executeHTTP(context.Background(), node, inputs)
	require.NoError(t, err)
	assert.Contains(t, output["body"], "/123")
}

func TestExecuteHTTP_ErrorStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"internal"}`))
	}))
	defer server.Close()

	e := engineWithNilRepos()
	node := &models.WorkflowNode{
		ID:   "http1",
		Type: "http",
		Config: map[string]interface{}{
			"url":    server.URL,
			"method": "GET",
		},
	}

	output, err := e.executeHTTP(context.Background(), node, nil)
	require.NoError(t, err) // HTTP errors don't return Go errors
	assert.Equal(t, 500, output["status_code"])
	assert.Contains(t, output["error"], "500")
}

func TestExecuteHTTP_MissingURL(t *testing.T) {
	e := engineWithNilRepos()
	node := &models.WorkflowNode{
		ID:   "http1",
		Type: "http",
		Config: map[string]interface{}{
			"method": "GET",
		},
	}

	_, err := e.executeHTTP(context.Background(), node, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "url is required")
}

// ---------------------------------------------------------------------------
// executeScript
// ---------------------------------------------------------------------------

func TestExecuteScript_ReturnJSON(t *testing.T) {
	e := engineWithNilRepos()
	node := &models.WorkflowNode{
		ID:   "script1",
		Type: "script",
		Config: map[string]interface{}{
			"code": `return {"result": "hello"}`,
		},
	}

	output, err := e.executeScript(context.Background(), node, nil)
	require.NoError(t, err)
	assert.Equal(t, "hello", output["result"])
}

func TestExecuteScript_DirectJSON(t *testing.T) {
	e := engineWithNilRepos()
	node := &models.WorkflowNode{
		ID:   "script1",
		Type: "script",
		Config: map[string]interface{}{
			"code": `{"key": "value", "count": 42}`,
		},
	}

	output, err := e.executeScript(context.Background(), node, nil)
	require.NoError(t, err)
	assert.Equal(t, "value", output["key"])
	assert.Equal(t, float64(42), output["count"])
}

func TestExecuteScript_WithTemplateVariables(t *testing.T) {
	e := engineWithNilRepos()
	node := &models.WorkflowNode{
		ID:   "script1",
		Type: "script",
		Config: map[string]interface{}{
			"code": `return {"status": {{inputs.node1.status_code}}}`,
		},
	}

	inputs := map[string]interface{}{
		"node1": map[string]interface{}{
			"status_code": float64(200),
		},
	}

	output, err := e.executeScript(context.Background(), node, inputs)
	require.NoError(t, err)
	assert.Equal(t, float64(200), output["status"])
}

func TestExecuteScript_EmptyCode(t *testing.T) {
	e := engineWithNilRepos()
	node := &models.WorkflowNode{
		ID:   "script1",
		Type: "script",
		Config: map[string]interface{}{
			"code": "",
		},
	}

	_, err := e.executeScript(context.Background(), node, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "code is required")
}

// ---------------------------------------------------------------------------
// executeCondition
// ---------------------------------------------------------------------------

func TestExecuteCondition_TrueLiteral(t *testing.T) {
	e := engineWithNilRepos()
	node := &models.WorkflowNode{
		ID:   "cond1",
		Type: "condition",
		Config: map[string]interface{}{
			"expression": "true",
		},
	}

	output, err := e.executeCondition(context.Background(), node, nil)
	require.NoError(t, err)
	assert.Equal(t, true, output["passed"])
}

func TestExecuteCondition_FalseLiteral(t *testing.T) {
	e := engineWithNilRepos()
	node := &models.WorkflowNode{
		ID:   "cond1",
		Type: "condition",
		Config: map[string]interface{}{
			"expression": "false",
		},
	}

	output, err := e.executeCondition(context.Background(), node, nil)
	require.NoError(t, err)
	assert.Equal(t, false, output["passed"])
}

func TestExecuteCondition_ComparisonWithTemplate(t *testing.T) {
	e := engineWithNilRepos()
	node := &models.WorkflowNode{
		ID:   "cond1",
		Type: "condition",
		Config: map[string]interface{}{
			"expression": "{{inputs.node1.status_code}} == 200",
		},
	}

	inputs := map[string]interface{}{
		"node1": map[string]interface{}{
			"status_code": float64(200),
		},
	}

	output, err := e.executeCondition(context.Background(), node, inputs)
	require.NoError(t, err)
	assert.Equal(t, true, output["passed"])
}

func TestExecuteCondition_ComparisonFails(t *testing.T) {
	e := engineWithNilRepos()
	node := &models.WorkflowNode{
		ID:   "cond1",
		Type: "condition",
		Config: map[string]interface{}{
			"expression": "{{inputs.node1.status_code}} == 200",
		},
	}

	inputs := map[string]interface{}{
		"node1": map[string]interface{}{
			"status_code": float64(404),
		},
	}

	output, err := e.executeCondition(context.Background(), node, inputs)
	require.NoError(t, err)
	assert.Equal(t, false, output["passed"])
}

func TestExecuteCondition_EmptyExpression(t *testing.T) {
	e := engineWithNilRepos()
	node := &models.WorkflowNode{
		ID:   "cond1",
		Type: "condition",
		Config: map[string]interface{}{
			"expression": "",
		},
	}

	_, err := e.executeCondition(context.Background(), node, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "expression is required")
}

func TestExecuteCondition_GreaterThan(t *testing.T) {
	e := engineWithNilRepos()
	node := &models.WorkflowNode{
		ID:   "cond1",
		Type: "condition",
		Config: map[string]interface{}{
			"expression": "{{inputs.node1.count}} > 10",
		},
	}

	inputs := map[string]interface{}{
		"node1": map[string]interface{}{
			"count": float64(15),
		},
	}

	output, err := e.executeCondition(context.Background(), node, inputs)
	require.NoError(t, err)
	assert.Equal(t, true, output["passed"])
}

// ---------------------------------------------------------------------------
// executeNodeLogic dispatch
// ---------------------------------------------------------------------------

func TestExecuteNodeLogic_UnknownType(t *testing.T) {
	e := engineWithNilRepos()
	node := &models.WorkflowNode{
		ID:   "unk1",
		Type: "unknown",
	}

	_, err := e.executeNodeLogic(context.Background(), node, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown node type")
}

func TestExecuteNodeLogic_DispatchDelay(t *testing.T) {
	e := engineWithNilRepos()
	node := &models.WorkflowNode{
		ID:   "delay1",
		Type: "delay",
		Config: map[string]interface{}{
			"seconds": float64(0),
		},
	}

	output, err := e.executeNodeLogic(context.Background(), node, nil)
	require.NoError(t, err)
	assert.Contains(t, output["message"], "Delayed")
}

func TestExecuteNodeLogic_DispatchHTTP(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`ok`))
	}))
	defer server.Close()

	e := engineWithNilRepos()
	node := &models.WorkflowNode{
		ID:   "http1",
		Type: "http",
		Config: map[string]interface{}{
			"url":    server.URL,
			"method": "GET",
		},
	}

	output, err := e.executeNodeLogic(context.Background(), node, nil)
	require.NoError(t, err)
	assert.Equal(t, 200, output["status_code"])
}

func TestExecuteNodeLogic_DispatchScript(t *testing.T) {
	e := engineWithNilRepos()
	node := &models.WorkflowNode{
		ID:   "script1",
		Type: "script",
		Config: map[string]interface{}{
			"code": `return {"ok": true}`,
		},
	}

	output, err := e.executeNodeLogic(context.Background(), node, nil)
	require.NoError(t, err)
	assert.Equal(t, true, output["ok"])
}

func TestExecuteNodeLogic_DispatchCondition(t *testing.T) {
	e := engineWithNilRepos()
	node := &models.WorkflowNode{
		ID:   "cond1",
		Type: "condition",
		Config: map[string]interface{}{
			"expression": "10 > 5",
		},
	}

	output, err := e.executeNodeLogic(context.Background(), node, nil)
	require.NoError(t, err)
	assert.Equal(t, true, output["passed"])
}
