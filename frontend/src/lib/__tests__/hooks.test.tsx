import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { act, renderHook, waitFor } from "@testing-library/react";
import { api } from "../api";
import {
  useCreateWorkflow,
  useDeleteWorkflow,
  useRollbackWorkflow,
  useRuns,
  useTriggerWorkflow,
  useUpdateWorkflow,
  useWorkflow,
  useWorkflows,
  useWorkflowVersions,
} from "../hooks";

// Mock the API module
jest.mock("../api", () => ({
  api: {
    getWorkflows: jest.fn(),
    getWorkflow: jest.fn(),
    createWorkflow: jest.fn(),
    updateWorkflow: jest.fn(),
    deleteWorkflow: jest.fn(),
    triggerWorkflow: jest.fn(),
    getWorkflowVersions: jest.fn(),
    rollbackWorkflow: jest.fn(),
    getRuns: jest.fn(),
  },
}));

const mockApi = api as jest.Mocked<typeof api>;

function createWrapper() {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: {
        retry: false,
      },
      mutations: {
        retry: false,
      },
    },
  });

  return ({ children }: { children: React.ReactNode }) => (
    <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  );
}

describe("Workflow Hooks", () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  describe("useWorkflows", () => {
    it("should fetch workflows successfully", async () => {
      const mockWorkflows = {
        workflows: [
          {
            id: "wf-1",
            name: "Test Workflow",
            description: "Test",
            active: true,
            definition: { nodes: [], edges: [] },
            timeout_seconds: 300,
            created_by: "user-1",
            created_at: "2024-01-01T00:00:00Z",
            updated_at: "2024-01-01T00:00:00Z",
          },
        ],
      };

      mockApi.getWorkflows.mockResolvedValue(mockWorkflows);

      const { result } = renderHook(() => useWorkflows(), {
        wrapper: createWrapper(),
      });

      await waitFor(() => {
        expect(result.current.data).toEqual(mockWorkflows.workflows);
        expect(result.current.isLoading).toBe(false);
      });
    });

    it("should handle error state", async () => {
      mockApi.getWorkflows.mockRejectedValue(new Error("Failed to fetch"));

      const { result } = renderHook(() => useWorkflows(), {
        wrapper: createWrapper(),
      });

      await waitFor(() => {
        expect(result.current.error).toBeTruthy();
        expect(result.current.isLoading).toBe(false);
      });
    });
  });

  describe("useWorkflow", () => {
    it("should fetch single workflow", async () => {
      const mockWorkflow = {
        id: "wf-1",
        name: "Test Workflow",
        description: "Test",
        active: true,
        definition: { nodes: [], edges: [] },
        timeout_seconds: 300,
        created_by: "user-1",
        created_at: "2024-01-01T00:00:00Z",
        updated_at: "2024-01-01T00:00:00Z",
      };

      mockApi.getWorkflow.mockResolvedValue(mockWorkflow);

      const { result } = renderHook(() => useWorkflow("wf-1"), {
        wrapper: createWrapper(),
      });

      await waitFor(() => {
        expect(result.current.data).toEqual(mockWorkflow);
      });
    });

    it("should not fetch when id is empty", () => {
      const { result } = renderHook(() => useWorkflow(""), {
        wrapper: createWrapper(),
      });

      expect(mockApi.getWorkflow).not.toHaveBeenCalled();
      expect(result.current.data).toBeUndefined();
    });
  });

  describe("useCreateWorkflow", () => {
    it("should create workflow successfully", async () => {
      const newWorkflow = {
        id: "wf-1",
        name: "New Workflow",
        description: "Description",
        active: true,
        definition: { nodes: [], edges: [] },
        timeout_seconds: 300,
        created_by: "user-1",
        created_at: "2024-01-01T00:00:00Z",
        updated_at: "2024-01-01T00:00:00Z",
      };

      mockApi.createWorkflow.mockResolvedValue(newWorkflow);

      const { result } = renderHook(() => useCreateWorkflow(), {
        wrapper: createWrapper(),
      });

      await act(async () => {
        await result.current.mutateAsync({
          name: "New Workflow",
          description: "Description",
          definition: { nodes: [], edges: [] },
        });
      });

      expect(mockApi.createWorkflow).toHaveBeenCalledWith(
        expect.objectContaining({
          name: "New Workflow",
        })
      );

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });
    });

    it("should handle creation error", async () => {
      mockApi.createWorkflow.mockRejectedValue(new Error("Creation failed"));

      const { result } = renderHook(() => useCreateWorkflow(), {
        wrapper: createWrapper(),
      });

      await act(async () => {
        try {
          await result.current.mutateAsync({
            name: "New Workflow",
            definition: { nodes: [], edges: [] },
          });
        } catch (err) {
          // Expected error
        }
      });

      await waitFor(() => {
        expect(result.current.isError).toBe(true);
      });
    });
  });

  describe("useUpdateWorkflow", () => {
    it("should update workflow successfully", async () => {
      const updatedWorkflow = {
        id: "wf-1",
        name: "Updated Workflow",
        description: "Updated",
        active: true,
        definition: { nodes: [], edges: [] },
        timeout_seconds: 300,
        created_by: "user-1",
        created_at: "2024-01-01T00:00:00Z",
        updated_at: "2024-01-02T00:00:00Z",
      };

      mockApi.updateWorkflow.mockResolvedValue(updatedWorkflow);

      const { result } = renderHook(() => useUpdateWorkflow(), {
        wrapper: createWrapper(),
      });

      await act(async () => {
        await result.current.mutateAsync({
          id: "wf-1",
          data: { name: "Updated Workflow" },
        });
      });

      expect(mockApi.updateWorkflow).toHaveBeenCalledWith("wf-1", {
        name: "Updated Workflow",
      });

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });
    });
  });

  describe("useDeleteWorkflow", () => {
    it("should delete workflow successfully", async () => {
      mockApi.deleteWorkflow.mockResolvedValue(undefined);

      const { result } = renderHook(() => useDeleteWorkflow(), {
        wrapper: createWrapper(),
      });

      await act(async () => {
        await result.current.mutateAsync("wf-1");
      });

      expect(mockApi.deleteWorkflow).toHaveBeenCalledWith("wf-1");

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });
    });
  });

  describe("useTriggerWorkflow", () => {
    it("should trigger workflow successfully", async () => {
      const mockRun = {
        id: "run-1",
        workflow_id: "wf-1",
        tenant_id: "tenant-1",
        status: "pending" as const,
        triggered_by: "user-1",
        created_at: "2024-01-01T00:00:00Z",
      };

      mockApi.triggerWorkflow.mockResolvedValue(mockRun);

      const { result } = renderHook(() => useTriggerWorkflow(), {
        wrapper: createWrapper(),
      });

      await act(async () => {
        await result.current.mutateAsync("wf-1");
      });

      expect(mockApi.triggerWorkflow).toHaveBeenCalledWith("wf-1");

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });
    });
  });

  describe("useRuns", () => {
    it("should fetch runs successfully", async () => {
      const mockRuns = {
        runs: [
          {
            id: "run-1",
            workflow_id: "wf-1",
            tenant_id: "tenant-1",
            status: "success" as const,
            triggered_by: "user-1",
            created_at: "2024-01-01T00:00:00Z",
          },
        ],
      };

      mockApi.getRuns.mockResolvedValue(mockRuns);

      const { result } = renderHook(() => useRuns("wf-1"), {
        wrapper: createWrapper(),
      });

      await waitFor(() => {
        expect(result.current.data).toEqual(mockRuns.runs);
      });
    });

    it("should fetch all runs when no workflowId provided", async () => {
      const mockRuns = {
        runs: [
          {
            id: "run-1",
            workflow_id: "wf-1",
            tenant_id: "tenant-1",
            status: "success" as const,
            triggered_by: "user-1",
            created_at: "2024-01-01T00:00:00Z",
          },
        ],
      };

      mockApi.getRuns.mockResolvedValue(mockRuns);

      const { result } = renderHook(() => useRuns(), {
        wrapper: createWrapper(),
      });

      await waitFor(() => {
        expect(result.current.data).toEqual(mockRuns.runs);
        expect(mockApi.getRuns).toHaveBeenCalledWith(undefined);
      });
    });
  });

  describe("useWorkflowVersions", () => {
    it("should fetch workflow versions successfully", async () => {
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

      mockApi.getWorkflowVersions.mockResolvedValue(mockVersions);

      const { result } = renderHook(() => useWorkflowVersions("wf-1"), {
        wrapper: createWrapper(),
      });

      await waitFor(() => {
        expect(result.current.data).toEqual(mockVersions.versions);
        expect(result.current.data).toHaveLength(2);
        expect(result.current.isLoading).toBe(false);
      });

      expect(mockApi.getWorkflowVersions).toHaveBeenCalledWith("wf-1");
    });

    it("should not fetch when workflowId is empty", () => {
      const { result } = renderHook(() => useWorkflowVersions(""), {
        wrapper: createWrapper(),
      });

      expect(mockApi.getWorkflowVersions).not.toHaveBeenCalled();
      expect(result.current.data).toBeUndefined();
    });

    it("should handle empty versions list", async () => {
      mockApi.getWorkflowVersions.mockResolvedValue({ versions: [] });

      const { result } = renderHook(() => useWorkflowVersions("wf-1"), {
        wrapper: createWrapper(),
      });

      await waitFor(() => {
        expect(result.current.data).toEqual([]);
        expect(result.current.isLoading).toBe(false);
      });
    });

    it("should handle error state", async () => {
      mockApi.getWorkflowVersions.mockRejectedValue(
        new Error("Failed to fetch versions")
      );

      const { result } = renderHook(() => useWorkflowVersions("wf-1"), {
        wrapper: createWrapper(),
      });

      await waitFor(() => {
        expect(result.current.error).toBeTruthy();
        expect(result.current.isLoading).toBe(false);
      });
    });
  });

  describe("useRollbackWorkflow", () => {
    it("should rollback workflow successfully", async () => {
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

      mockApi.rollbackWorkflow.mockResolvedValue(mockWorkflow);

      const { result } = renderHook(() => useRollbackWorkflow(), {
        wrapper: createWrapper(),
      });

      await act(async () => {
        await result.current.mutateAsync({ id: "wf-1", version: 1 });
      });

      expect(mockApi.rollbackWorkflow).toHaveBeenCalledWith("wf-1", 1);

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });
    });

    it("should handle rollback error", async () => {
      mockApi.rollbackWorkflow.mockRejectedValue(
        new Error("Version not found")
      );

      const { result } = renderHook(() => useRollbackWorkflow(), {
        wrapper: createWrapper(),
      });

      await act(async () => {
        try {
          await result.current.mutateAsync({ id: "wf-1", version: 999 });
        } catch (err) {
          // Expected error
        }
      });

      await waitFor(() => {
        expect(result.current.isError).toBe(true);
      });
      expect(mockApi.rollbackWorkflow).toHaveBeenCalledWith("wf-1", 999);
    });

    it("should invalidate queries on successful rollback", async () => {
      const mockWorkflow = {
        id: "wf-1",
        name: "Test Workflow",
        description: "Test",
        active: true,
        definition: { nodes: [], edges: [] },
        timeout_seconds: 300,
        created_by: "user-1",
        created_at: "2024-01-01T00:00:00Z",
        updated_at: "2024-01-02T00:00:00Z",
      };

      mockApi.rollbackWorkflow.mockResolvedValue(mockWorkflow);

      const queryClient = new QueryClient({
        defaultOptions: {
          mutations: {
            retry: false,
          },
        },
      });

      const { result } = renderHook(() => useRollbackWorkflow(), {
        wrapper: ({ children }) => (
          <QueryClientProvider client={queryClient}>
            {children}
          </QueryClientProvider>
        ),
      });

      // Set some initial data
      queryClient.setQueryData(["workflow", "wf-1"], mockWorkflow);
      queryClient.setQueryData(["workflow-versions", "wf-1"], [{ version: 1 }]);

      await act(async () => {
        await result.current.mutateAsync({ id: "wf-1", version: 1 });
      });

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });

      // Verify that queries were invalidated
      expect(
        queryClient.getQueryState(["workflow", "wf-1"])?.isInvalidated
      ).toBe(true);
      expect(
        queryClient.getQueryState(["workflow-versions", "wf-1"])?.isInvalidated
      ).toBe(true);
    });
  });
});
