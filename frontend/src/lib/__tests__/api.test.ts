import { api } from '../api';
import { APIClient } from '../api';

// Mock fetch globally
global.fetch = jest.fn();

describe('API', () => {
  let apiClient: APIClient;

  beforeEach(() => {
    apiClient = new APIClient();
    (global.fetch as jest.Mock).mockClear();
    localStorage.clear();
  });

  describe('setToken and getToken', () => {
    it('should store token', () => {
      apiClient.setToken('test-token');
      // Token is stored privately, we can't directly test retrieval
      // But we can verify it doesn't throw
      expect(() => apiClient.setToken('test-token')).not.toThrow();
    });
  });

  describe('request', () => {
    it('should make request without authentication', async () => {
      (global.fetch as jest.Mock).mockResolvedValue({
        ok: true,
        json: async () => ({ data: 'test' }),
      });

      const result = await apiClient.request('/test');
      expect(result).toEqual({ data: 'test' });
      expect(global.fetch).toHaveBeenCalledWith(
        expect.stringContaining('/test'),
        expect.objectContaining({
          headers: expect.objectContaining({
            'Content-Type': 'application/json',
          }),
        })
      );
    });

    it('should make request with authentication token', async () => {
      apiClient.setToken('test-token');
      (global.fetch as jest.Mock).mockResolvedValue({
        ok: true,
        json: async () => ({ data: 'test' }),
      });

      await apiClient.request('/test');
      expect(global.fetch).toHaveBeenCalledWith(
        expect.any(String),
        expect.objectContaining({
          headers: expect.objectContaining({
            'Authorization': 'Bearer test-token',
          }),
        })
      );
    });

    it('should throw error on non-OK response', async () => {
      (global.fetch as jest.Mock).mockResolvedValue({
        ok: false,
        status: 401,
        json: async () => ({ error: 'Unauthorized' }),
      });

      await expect(apiClient.request('/test')).rejects.toThrow('Unauthorized');
    });

    it('should handle POST requests with body', async () => {
      (global.fetch as jest.Mock).mockResolvedValue({
        ok: true,
        json: async () => ({ success: true }),
      });

      await apiClient.request('/test', {
        method: 'POST',
        body: JSON.stringify({ data: 'test' }),
      });

      expect(global.fetch).toHaveBeenCalledWith(
        expect.any(String),
        expect.objectContaining({
          method: 'POST',
          body: JSON.stringify({ data: 'test' }),
        })
      );
    });
  });

  describe('login', () => {
    it('should login successfully and store token', async () => {
      const mockResponse = {
        token: 'test-token',
        user: {
          id: 'user-1',
          email: 'test@example.com',
          role: 'admin',
        },
      };

      (global.fetch as jest.Mock).mockResolvedValue({
        ok: true,
        json: async () => mockResponse,
      });

      const result = await apiClient.login({
        email: 'test@example.com',
        password: 'password',
      });

      expect(result).toEqual(mockResponse);
    });

    it('should throw error on invalid credentials', async () => {
      (global.fetch as jest.Mock).mockResolvedValue({
        ok: false,
        status: 401,
        json: async () => ({ error: 'Invalid credentials' }),
      });

      await expect(
        apiClient.login({
          email: 'test@example.com',
          password: 'wrong',
        })
      ).rejects.toThrow('Invalid credentials');
    });
  });

  describe('getMe', () => {
    it('should get current user', async () => {
      const mockUser = {
        id: 'user-1',
        email: 'test@example.com',
        role: 'admin',
      };

      (global.fetch as jest.Mock).mockResolvedValue({
        ok: true,
        json: async () => mockUser,
      });

      const result = await apiClient.getMe();
      expect(result).toEqual(mockUser);
    });
  });

  describe('getWorkflows', () => {
    it('should get workflows list', async () => {
      const mockWorkflows = {
        data: [
          {
            id: 'wf-1',
            name: 'Test Workflow',
            description: 'Test',
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
      expect(result.workflows[0].name).toBe('Test Workflow');
    });
  });

  describe('createWorkflow', () => {
    it('should create a new workflow', async () => {
      const newWorkflow = {
        id: 'wf-1',
        name: 'New Workflow',
        description: 'Description',
        active: true,
      };

      (global.fetch as jest.Mock).mockResolvedValue({
        ok: true,
        json: async () => newWorkflow,
      });

      const result = await apiClient.createWorkflow({
        name: 'New Workflow',
        description: 'Description',
        definition: { nodes: [], edges: [] },
        timeout_seconds: 300,
        active: true,
      });

      expect(result.name).toBe('New Workflow');
    });
  });

  describe('triggerWorkflow', () => {
    it('should trigger a workflow', async () => {
      const mockResponse = {
        run_id: 'run-1',
        status: 'pending',
      };

      (global.fetch as jest.Mock).mockResolvedValue({
        ok: true,
        json: async () => mockResponse,
      });

      const result = await apiClient.triggerWorkflow('wf-1');
      expect(result.run_id).toBe('run-1');
    });
  });

  describe('deleteWorkflow', () => {
    it('should delete a workflow', async () => {
      (global.fetch as jest.Mock).mockResolvedValue({
        ok: true,
        json: async () => ({ success: true }),
      });

      await expect(apiClient.deleteWorkflow('wf-1')).resolves.not.toThrow();
    });
  });

  describe('getRuns', () => {
    it('should get workflow runs', async () => {
      const mockRuns = {
        data: [
          {
            id: 'run-1',
            workflow_id: 'wf-1',
            status: 'success',
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
  });

  describe('getHealthStats', () => {
    it('should get health statistics', async () => {
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
});
