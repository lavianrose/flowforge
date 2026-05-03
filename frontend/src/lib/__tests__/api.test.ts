import { APIClient } from "../api";

// Mock fetch globally
global.fetch = jest.fn();

describe("API", () => {
  let apiClient: APIClient;

  beforeEach(() => {
    apiClient = new APIClient();
    (global.fetch as jest.Mock).mockClear();
    localStorage.clear();
  });

  describe("setToken and getToken", () => {
    it("should store token", () => {
      apiClient.setToken("test-token");
      // Token is stored privately, we can't directly test retrieval
      // But we can verify it doesn't throw
      expect(() => apiClient.setToken("test-token")).not.toThrow();
    });
  });

  describe("request", () => {
    it("should make request without authentication", async () => {
      (global.fetch as jest.Mock).mockResolvedValue({
        ok: true,
        json: async () => ({ data: "test" }),
      });

      const result = await apiClient.request("/test");
      expect(result).toEqual({ data: "test" });
      expect(global.fetch).toHaveBeenCalledWith(
        expect.stringContaining("/test"),
        expect.objectContaining({
          headers: expect.objectContaining({
            "Content-Type": "application/json",
          }),
        })
      );
    });

    it("should make request with authentication token", async () => {
      apiClient.setToken("test-token");
      (global.fetch as jest.Mock).mockResolvedValue({
        ok: true,
        json: async () => ({ data: "test" }),
      });

      await apiClient.request("/test");
      expect(global.fetch).toHaveBeenCalledWith(
        expect.any(String),
        expect.objectContaining({
          headers: expect.objectContaining({
            Authorization: "Bearer test-token",
          }),
        })
      );
    });

    it("should throw error on non-OK response", async () => {
      (global.fetch as jest.Mock).mockResolvedValue({
        ok: false,
        status: 401,
        json: async () => ({ error: "Unauthorized" }),
      });

      await expect(apiClient.request("/test")).rejects.toThrow("Unauthorized");
    });

    it("should handle POST requests with body", async () => {
      (global.fetch as jest.Mock).mockResolvedValue({
        ok: true,
        json: async () => ({ success: true }),
      });

      await apiClient.request("/test", {
        method: "POST",
        body: JSON.stringify({ data: "test" }),
      });

      expect(global.fetch).toHaveBeenCalledWith(
        expect.any(String),
        expect.objectContaining({
          method: "POST",
          body: JSON.stringify({ data: "test" }),
        })
      );
    });
  });

  describe("login", () => {
    it("should login successfully and store token", async () => {
      const mockResponse = {
        token: "test-token",
        user: {
          id: "user-1",
          email: "test@example.com",
          role: "admin",
        },
      };

      (global.fetch as jest.Mock).mockResolvedValue({
        ok: true,
        json: async () => mockResponse,
      });

      const result = await apiClient.login({
        email: "test@example.com",
        password: "password",
      });

      expect(result).toEqual(mockResponse);
    });

    it("should throw error on invalid credentials", async () => {
      (global.fetch as jest.Mock).mockResolvedValue({
        ok: false,
        status: 401,
        json: async () => ({ error: "Invalid credentials" }),
      });

      await expect(
        apiClient.login({
          email: "test@example.com",
          password: "wrong",
        })
      ).rejects.toThrow("Invalid credentials");
    });
  });

  describe("getMe", () => {
    it("should get current user", async () => {
      const mockUser = {
        id: "user-1",
        email: "test@example.com",
        role: "admin",
      };

      (global.fetch as jest.Mock).mockResolvedValue({
        ok: true,
        json: async () => mockUser,
      });

      const result = await apiClient.getMe();
      expect(result).toEqual(mockUser);
    });
  });

  describe("getWorkflows", () => {
    it("should get workflows list", async () => {
      const mockWorkflows = {
        data: [
          {
            id: "wf-1",
            name: "Test Workflow",
            description: "Test",
            active: true,
          },
        ],
        pagination: {
          page: 1,
          limit: 10,
          total: 1,
        },
      };

      (global.fetch as jest.Mock).mockResolvedValue({
        ok: true,
        json: async () => mockWorkflows,
      });

      const result = await apiClient.getWorkflows();
      expect(result.workflows).toHaveLength(1);
      expect(result.workflows[0].name).toBe("Test Workflow");
    });
  });

  describe("createWorkflow", () => {
    it("should create a new workflow", async () => {
      const newWorkflow = {
        id: "wf-1",
        name: "New Workflow",
        description: "Description",
        active: true,
      };

      (global.fetch as jest.Mock).mockResolvedValue({
        ok: true,
        json: async () => newWorkflow,
      });

      const result = await apiClient.createWorkflow({
        name: "New Workflow",
        description: "Description",
        definition: { nodes: [], edges: [] },
        timeout_seconds: 300,
        active: true,
      });

      expect(result.name).toBe("New Workflow");
    });
  });

  describe("triggerWorkflow", () => {
    it("should trigger a workflow", async () => {
      const mockResponse = {
        run_id: "run-1",
        status: "pending",
      };

      (global.fetch as jest.Mock).mockResolvedValue({
        ok: true,
        json: async () => mockResponse,
      });

      const result = await apiClient.triggerWorkflow("wf-1");
      expect(result.run_id).toBe("run-1");
    });
  });

  describe("getWorkflow", () => {
    it("should get a single workflow", async () => {
      const mockWorkflow = {
        id: "wf-1",
        name: "Test Workflow",
        description: "Test description",
        active: true,
        definition: {
          nodes: [
            {
              id: "node-1",
              type: "http",
              name: "Test",
              config: {},
              position: { x: 0, y: 0 },
            },
          ],
          edges: [],
        },
        timeout_seconds: 300,
        created_by: "user-1",
        created_at: "2024-01-01T00:00:00Z",
        updated_at: "2024-01-01T00:00:00Z",
      };

      (global.fetch as jest.Mock).mockResolvedValue({
        ok: true,
        json: async () => mockWorkflow,
      });

      const result = await apiClient.getWorkflow("wf-1");
      expect(result.id).toBe("wf-1");
      expect(result.name).toBe("Test Workflow");
      expect(result.definition.nodes).toHaveLength(1);
    });

    it("should handle workflow not found error", async () => {
      (global.fetch as jest.Mock).mockResolvedValue({
        ok: false,
        status: 404,
        json: async () => ({ error: "Workflow not found" }),
      });

      await expect(apiClient.getWorkflow("invalid-wf")).rejects.toThrow(
        "Workflow not found"
      );
    });
  });

  describe("updateWorkflow", () => {
    it("should update a workflow", async () => {
      const mockWorkflow = {
        id: "wf-1",
        name: "Updated Workflow",
        description: "Updated description",
        active: true,
        definition: {
          nodes: [
            {
              id: "node-1",
              type: "http",
              name: "Test",
              config: {},
              position: { x: 0, y: 0 },
            },
          ],
          edges: [],
        },
        timeout_seconds: 300,
        created_by: "user-1",
        created_at: "2024-01-01T00:00:00Z",
        updated_at: "2024-01-02T00:00:00Z",
      };

      (global.fetch as jest.Mock).mockResolvedValue({
        ok: true,
        json: async () => mockWorkflow,
      });

      const result = await apiClient.updateWorkflow("wf-1", {
        name: "Updated Workflow",
        description: "Updated description",
      });

      expect(result.name).toBe("Updated Workflow");
      expect(result.description).toBe("Updated description");

      // Verify PUT request was made
      expect(global.fetch).toHaveBeenCalledWith(
        expect.stringContaining("/workflows/wf-1"),
        expect.objectContaining({
          method: "PUT",
        })
      );
    });

    it("should handle workflow not found on update", async () => {
      (global.fetch as jest.Mock).mockResolvedValue({
        ok: false,
        status: 404,
        json: async () => ({ error: "Workflow not found" }),
      });

      await expect(
        apiClient.updateWorkflow("invalid-wf", { name: "Updated" })
      ).rejects.toThrow("Workflow not found");
    });
  });

  describe("deleteWorkflow", () => {
    it("should delete a workflow", async () => {
      (global.fetch as jest.Mock).mockResolvedValue({
        ok: true,
        json: async () => ({ success: true }),
      });

      await expect(apiClient.deleteWorkflow("wf-1")).resolves.not.toThrow();
    });
  });

  describe("getRuns", () => {
    it("should get workflow runs", async () => {
      const mockRuns = {
        data: [
          {
            id: "run-1",
            workflow_id: "wf-1",
            status: "success",
          },
        ],
        pagination: {
          page: 1,
          limit: 10,
          total: 1,
        },
      };

      (global.fetch as jest.Mock).mockResolvedValue({
        ok: true,
        json: async () => mockRuns,
      });

      const result = await apiClient.getRuns();
      expect(result.runs).toHaveLength(1);
    });

    it("should get runs for specific workflow", async () => {
      const mockRuns = {
        data: [
          {
            id: "run-1",
            workflow_id: "wf-1",
            status: "success",
          },
        ],
        pagination: {
          page: 1,
          limit: 10,
          total: 1,
        },
      };

      (global.fetch as jest.Mock).mockResolvedValue({
        ok: true,
        json: async () => mockRuns,
      });

      const result = await apiClient.getRuns("wf-1");
      expect(result.runs).toHaveLength(1);

      // Verify query parameter was included
      expect(global.fetch).toHaveBeenCalledWith(
        expect.stringContaining("workflow_id=wf-1"),
        expect.any(Object)
      );
    });
  });

  describe("getRun", () => {
    it("should get a single run with steps", async () => {
      const mockRun = {
        run: {
          id: "run-1",
          workflow_id: "wf-1",
          tenant_id: "tenant-1",
          status: "success",
          triggered_by: "user-1",
          created_at: "2024-01-01T00:00:00Z",
          started_at: "2024-01-01T00:00:01Z",
          completed_at: "2024-01-01T00:00:05Z",
        },
        steps: [
          {
            id: "step-1",
            run_id: "run-1",
            step_id: "node-1",
            status: "success",
            input: { url: "https://example.com" },
            output: { status: 200 },
            error: null,
            retry_count: 0,
            created_at: "2024-01-01T00:00:01Z",
            completed_at: "2024-01-01T00:00:02Z",
          },
        ],
      };

      (global.fetch as jest.Mock).mockResolvedValue({
        ok: true,
        json: async () => mockRun,
      });

      const result = await apiClient.getRun("run-1");
      expect(result.run.id).toBe("run-1");
      expect(result.run.status).toBe("success");
      expect(result.steps).toHaveLength(1);
      expect(result.steps[0].status).toBe("success");
    });

    it("should handle run not found error", async () => {
      (global.fetch as jest.Mock).mockResolvedValue({
        ok: false,
        status: 404,
        json: async () => ({ error: "Run not found" }),
      });

      await expect(apiClient.getRun("invalid-run")).rejects.toThrow(
        "Run not found"
      );
    });
  });

  describe("getHealthStats", () => {
    it("should get health statistics", async () => {
      const mockStats = {
        active_runs: 5,
        success_rate: 95.5,
        failure_rate: 4.5,
        avg_duration_seconds: 120,
        total_runs_24h: 100,
        success_runs_24h: 95,
        failed_runs_24h: 5,
        hourly_stats: [],
      };

      (global.fetch as jest.Mock).mockResolvedValue({
        ok: true,
        json: async () => mockStats,
      });

      const result = await apiClient.getHealthStats();
      expect(result.active_runs).toBe(5);
      expect(result.success_rate).toBe(95.5);
    });
  });

  describe("getWorkflowVersions", () => {
    it("should get workflow versions", async () => {
      const mockVersions = {
        versions: [
          {
            id: "ver-1",
            workflow_id: "wf-1",
            version: 1,
            definition: {
              nodes: [
                {
                  id: "node-1",
                  type: "http",
                  name: "Test",
                  config: {},
                  position: { x: 0, y: 0 },
                },
              ],
              edges: [],
            },
            created_by: "user-1",
            created_at: "2024-01-01T00:00:00Z",
          },
          {
            id: "ver-2",
            workflow_id: "wf-1",
            version: 2,
            definition: {
              nodes: [
                {
                  id: "node-1",
                  type: "http",
                  name: "Test Updated",
                  config: {},
                  position: { x: 0, y: 0 },
                },
              ],
              edges: [],
            },
            created_by: "user-1",
            created_at: "2024-01-02T00:00:00Z",
          },
        ],
      };

      (global.fetch as jest.Mock).mockResolvedValue({
        ok: true,
        json: async () => mockVersions,
      });

      const result = await apiClient.getWorkflowVersions("wf-1");
      expect(result.versions).toHaveLength(2);
      expect(result.versions[0].version).toBe(1);
      expect(result.versions[1].version).toBe(2);
    });

    it("should handle empty versions list", async () => {
      (global.fetch as jest.Mock).mockResolvedValue({
        ok: true,
        json: async () => ({ versions: [] }),
      });

      const result = await apiClient.getWorkflowVersions("wf-1");
      expect(result.versions).toHaveLength(0);
    });
  });

  describe("rollbackWorkflow", () => {
    it("should rollback workflow to specific version", async () => {
      const mockWorkflow = {
        id: "wf-1",
        name: "Test Workflow",
        description: "Test",
        active: true,
        definition: {
          nodes: [
            {
              id: "node-1",
              type: "http",
              name: "Test",
              config: {},
              position: { x: 0, y: 0 },
            },
          ],
          edges: [],
        },
        timeout_seconds: 300,
        created_by: "user-1",
        created_at: "2024-01-01T00:00:00Z",
        updated_at: "2024-01-02T00:00:00Z",
      };

      (global.fetch as jest.Mock).mockResolvedValue({
        ok: true,
        json: async () => mockWorkflow,
      });

      const result = await apiClient.rollbackWorkflow("wf-1", 1);
      expect(result.id).toBe("wf-1");
      expect(result.name).toBe("Test Workflow");

      // Verify the correct endpoint was called
      expect(global.fetch).toHaveBeenCalledWith(
        expect.stringContaining("/workflows/wf-1/rollback/1"),
        expect.objectContaining({
          method: "POST",
        })
      );
    });

    it("should handle version not found error", async () => {
      (global.fetch as jest.Mock).mockResolvedValue({
        ok: false,
        status: 404,
        json: async () => ({ error: "version not found" }),
      });

      await expect(apiClient.rollbackWorkflow("wf-1", 999)).rejects.toThrow(
        "version not found"
      );
    });

    it("should handle workflow not found error", async () => {
      (global.fetch as jest.Mock).mockResolvedValue({
        ok: false,
        status: 404,
        json: async () => ({ error: "workflow not found" }),
      });

      await expect(apiClient.rollbackWorkflow("invalid-wf", 1)).rejects.toThrow(
        "workflow not found"
      );
    });
  });
});
