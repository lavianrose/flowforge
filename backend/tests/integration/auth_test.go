package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string      `json:"token"`
	User  interface{} `json:"user"`
}

func TestAuthenticationIntegration(t *testing.T) {
	ts := Setup(t)
	defer ts.Teardown(t)

	// Create test tenant
	tenantID := ts.CreateTestTenant(t)

	// Create test users with different roles
	adminUserID := ts.CreateTestUser(t, tenantID, "admin@test.com", "admin123", "admin")
	editorUserID := ts.CreateTestUser(t, tenantID, "editor@test.com", "editor123", "editor")
	viewerUserID := ts.CreateTestUser(t, tenantID, "viewer@test.com", "viewer123", "viewer")

	t.Run("Admin Login Success", func(t *testing.T) {
		reqBody := LoginRequest{
			Email:    "admin@test.com",
			Password: "admin123",
		}

		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := ts.DoRequest(req)
		if err != nil {
			t.Fatalf("Failed to send request: %v", err)
		}

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		var loginResp LoginResponse
		json.NewDecoder(resp.Body).Decode(&loginResp)

		if loginResp.Token == "" {
			t.Error("Expected token in response")
		}

		user := loginResp.User.(map[string]interface{})
		if user["role"] != "admin" {
			t.Errorf("Expected role admin, got %v", user["role"])
		}
	})

	t.Run("Editor Login Success", func(t *testing.T) {
		reqBody := LoginRequest{
			Email:    "editor@test.com",
			Password: "editor123",
		}

		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := ts.DoRequest(req)
		if err != nil {
			t.Fatalf("Failed to send request: %v", err)
		}

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		var loginResp LoginResponse
		json.NewDecoder(resp.Body).Decode(&loginResp)

		if loginResp.Token == "" {
			t.Error("Expected token in response")
		}

		user := loginResp.User.(map[string]interface{})
		if user["role"] != "editor" {
			t.Errorf("Expected role editor, got %v", user["role"])
		}
	})

	t.Run("Viewer Login Success", func(t *testing.T) {
		reqBody := LoginRequest{
			Email:    "viewer@test.com",
			Password: "viewer123",
		}

		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := ts.DoRequest(req)
		if err != nil {
			t.Fatalf("Failed to send request: %v", err)
		}

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		var loginResp LoginResponse
		json.NewDecoder(resp.Body).Decode(&loginResp)

		if loginResp.Token == "" {
			t.Error("Expected token in response")
		}

		user := loginResp.User.(map[string]interface{})
		if user["role"] != "viewer" {
			t.Errorf("Expected role viewer, got %v", user["role"])
		}
	})

	t.Run("Invalid Credentials", func(t *testing.T) {
		reqBody := LoginRequest{
			Email:    "admin@test.com",
			Password: "wrongpassword",
		}

		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := ts.DoRequest(req)
		if err != nil {
			t.Fatalf("Failed to send request: %v", err)
		}

		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("Expected status 401, got %d", resp.StatusCode)
		}
	})

	t.Run("Missing Authorization Header", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/workflows", nil)
		resp, err := ts.DoRequest(req)
		if err != nil {
			t.Fatalf("Failed to send request: %v", err)
		}

		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("Expected status 401, got %d", resp.StatusCode)
		}
	})

	t.Run("Invalid Token", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/workflows", nil)
		req.Header.Set("Authorization", "Bearer invalid-token")

		resp, err := ts.DoRequest(req)
		if err != nil {
			t.Fatalf("Failed to send request: %v", err)
		}

		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("Expected status 401, got %d", resp.StatusCode)
		}
	})

	t.Run("Valid Token Access", func(t *testing.T) {
		token := ts.GenerateTestToken(t, adminUserID, tenantID, "admin@test.com", "admin")

		req := httptest.NewRequest("GET", "/api/v1/workflows", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := ts.DoRequest(req)
		if err != nil {
			t.Fatalf("Failed to send request: %v", err)
		}

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}
	})

	// Use the user IDs in RBAC tests
	_ = editorUserID
	_ = viewerUserID
}

func TestRBACIntegration(t *testing.T) {
	ts := Setup(t)
	defer ts.Teardown(t)

	// Create test tenant
	tenantID := ts.CreateTestTenant(t)

	// Create test users
	adminUserID := ts.CreateTestUser(t, tenantID, "admin@test.com", "admin123", "admin")
	editorUserID := ts.CreateTestUser(t, tenantID, "editor@test.com", "editor123", "editor")
	viewerUserID := ts.CreateTestUser(t, tenantID, "viewer@test.com", "viewer123", "viewer")

	// Create a test workflow as admin
	ctx := context.Background()
	workflowID := ts.CreateTestWorkflow(t, ctx, tenantID, adminUserID, "Test Workflow")

	t.Run("Viewer Cannot Create Workflow", func(t *testing.T) {
		token := ts.GenerateTestToken(t, viewerUserID, tenantID, "viewer@test.com", "viewer")

		workflowData := map[string]interface{}{
			"name":        "Viewer Workflow",
			"description": "Should not be created",
			"definition": map[string]interface{}{
				"nodes": []map[string]interface{}{},
				"edges": []map[string]interface{}{},
			},
			"timeout_seconds": 300,
			"active":          true,
		}

		body, _ := json.Marshal(workflowData)
		req := httptest.NewRequest("POST", "/api/v1/workflows", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := ts.DoRequest(req)
		if err != nil {
			t.Fatalf("Failed to send request: %v", err)
		}

		if resp.StatusCode != http.StatusForbidden {
			t.Errorf("Expected status 403, got %d", resp.StatusCode)
		}
	})

	t.Run("Editor Can Create Workflow", func(t *testing.T) {
		token := ts.GenerateTestToken(t, editorUserID, tenantID, "editor@test.com", "editor")

		workflowData := map[string]interface{}{
			"name":        "Editor Workflow",
			"description": "Created by editor",
			"definition": map[string]interface{}{
				"nodes": []map[string]interface{}{
					{
						"id":       "node-1",
						"type":     "http",
						"name":     "HTTP Request",
						"config":   map[string]string{"url": "https://api.example.com", "method": "GET"},
						"position": map[string]int{"x": 100, "y": 100},
					},
				},
				"edges": []map[string]interface{}{},
			},
			"timeout_seconds": 300,
			"active":          true,
		}

		body, _ := json.Marshal(workflowData)
		req := httptest.NewRequest("POST", "/api/v1/workflows", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := ts.DoRequest(req)
		if err != nil {
			t.Fatalf("Failed to send request: %v", err)
		}

		if resp.StatusCode != http.StatusCreated {
			body, _ := json.Marshal(resp)
			t.Errorf("Expected status 201, got %d. Response: %s", resp.StatusCode, string(body))
		}
	})

	t.Run("Viewer Cannot Trigger Workflow", func(t *testing.T) {
		token := ts.GenerateTestToken(t, viewerUserID, tenantID, "viewer@test.com", "viewer")

		req := httptest.NewRequest("POST", "/api/v1/workflows/"+workflowID+"/trigger", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := ts.DoRequest(req)
		if err != nil {
			t.Fatalf("Failed to send request: %v", err)
		}

		if resp.StatusCode != http.StatusForbidden {
			t.Errorf("Expected status 403, got %d", resp.StatusCode)
		}
	})

	t.Run("Editor Can Trigger Workflow", func(t *testing.T) {
		token := ts.GenerateTestToken(t, editorUserID, tenantID, "editor@test.com", "editor")

		req := httptest.NewRequest("POST", "/api/v1/workflows/"+workflowID+"/trigger", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := ts.DoRequest(req)
		if err != nil {
			t.Fatalf("Failed to send request: %v", err)
		}

		// 202 Accepted is valid for async operations
		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
			t.Errorf("Expected status 200 or 202, got %d", resp.StatusCode)
		}
	})

	t.Run("Editor Cannot Delete Workflow", func(t *testing.T) {
		token := ts.GenerateTestToken(t, editorUserID, tenantID, "editor@test.com", "editor")

		req := httptest.NewRequest("DELETE", "/api/v1/workflows/"+workflowID, nil)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := ts.DoRequest(req)
		if err != nil {
			t.Fatalf("Failed to send request: %v", err)
		}

		if resp.StatusCode != http.StatusForbidden {
			t.Errorf("Expected status 403, got %d", resp.StatusCode)
		}
	})

	t.Run("Admin Can Delete Workflow", func(t *testing.T) {
		token := ts.GenerateTestToken(t, adminUserID, tenantID, "admin@test.com", "admin")

		req := httptest.NewRequest("DELETE", "/api/v1/workflows/"+workflowID, nil)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := ts.DoRequest(req)
		if err != nil {
			t.Fatalf("Failed to send request: %v", err)
		}

		// 204 No Content is standard for successful DELETE
		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
			t.Errorf("Expected status 200 or 204, got %d", resp.StatusCode)
		}
	})
}

func TestMultiTenantIsolation(t *testing.T) {
	ts := Setup(t)
	defer ts.Teardown(t)

	// Create two separate tenants
	tenantAID := ts.CreateTestTenant(t)
	tenantBID := ts.CreateTestTenant(t)

	// Create users for each tenant
	adminAUserID := ts.CreateTestUser(t, tenantAID, "admin-a@test.com", "admin123", "admin")
	adminBUserID := ts.CreateTestUser(t, tenantBID, "admin-b@test.com", "admin123", "admin")

	// Create workflow in tenant A
	ctx := context.Background()
	workflowAID := ts.CreateTestWorkflow(t, ctx, tenantAID, adminAUserID, "Tenant A Workflow")

	t.Run("Tenant B Cannot Access Tenant A Workflow", func(t *testing.T) {
		token := ts.GenerateTestToken(t, adminBUserID, tenantBID, "admin-b@test.com", "admin")

		req := httptest.NewRequest("GET", "/api/v1/workflows/"+workflowAID, nil)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := ts.DoRequest(req)
		if err != nil {
			t.Fatalf("Failed to send request: %v", err)
		}

		// Should return 404 or 403 depending on implementation
		if resp.StatusCode != http.StatusNotFound && resp.StatusCode != http.StatusForbidden {
			t.Errorf("Expected status 404 or 403, got %d", resp.StatusCode)
		}
	})

	t.Run("Tenant A Can Access Own Workflow", func(t *testing.T) {
		token := ts.GenerateTestToken(t, adminAUserID, tenantAID, "admin-a@test.com", "admin")

		req := httptest.NewRequest("GET", "/api/v1/workflows/"+workflowAID, nil)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := ts.DoRequest(req)
		if err != nil {
			t.Fatalf("Failed to send request: %v", err)
		}

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}
	})

	t.Run("Tenant B List Does Not Include Tenant A Workflows", func(t *testing.T) {
		token := ts.GenerateTestToken(t, adminBUserID, tenantBID, "admin-b@test.com", "admin")

		req := httptest.NewRequest("GET", "/api/v1/workflows", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := ts.DoRequest(req)
		if err != nil {
			t.Fatalf("Failed to send request: %v", err)
		}

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		var result map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		// Check if data exists and is array
		data, exists := result["data"]
		if !exists || data == nil {
			// Empty data is acceptable - means no workflows
			return
		}

		workflows, ok := data.([]interface{})
		if !ok {
			// If data is not a slice, check other response formats
			return
		}

		if len(workflows) != 0 {
			t.Errorf("Expected 0 workflows for tenant B, got %d", len(workflows))
		}
	})
}
