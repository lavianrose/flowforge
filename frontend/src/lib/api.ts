const API_BASE = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:3000/api/v1';

export interface LoginRequest {
  email: string;
  password: string;
}

export interface LoginResponse {
  token: string;
  user: {
    id: string;
    email: string;
    role: string;
    tenant_id: string;
  };
}

export interface Workflow {
  id: string;
  tenant_id: string;
  name: string;
  description: string;
  definition: {
    nodes: Array<{
      id: string;
      type: string;
      name: string;
      config: Record<string, unknown>;
      position: { x: number; y: number };
    }>;
    edges: Array<{
      id: string;
      source: string;
      target: string;
    }>;
  };
  timeout_seconds: number;
  active: boolean;
  created_by: string;
  created_at: string;
  updated_at: string;
}

export interface WorkflowRun {
  id: string;
  workflow_id: string;
  tenant_id: string;
  status: 'pending' | 'running' | 'success' | 'failed' | 'cancelled';
  error?: string;
  started_at?: string;
  completed_at?: string;
  created_by?: string;
  created_at: string;
  triggered_by: string;
}

export interface HealthStats {
  active_runs: number;
  success_rate: number;
  failure_rate: number;
  avg_duration_seconds: number;
  total_runs_24h: number;
  success_runs_24h: number;
  failed_runs_24h: number;
  hourly_stats: Array<{
    hour: number;
    total_runs: number;
    success_runs: number;
    failed_runs: number;
    avg_duration: number;
  }>;
}

class APIClient {
  private token: string | null = null;

  setToken(token: string) {
    this.token = token;
  }

  clearToken() {
    this.token = null;
  }

  private async request<T>(
    endpoint: string,
    options: RequestInit = {}
  ): Promise<T> {
    const headers: HeadersInit = {
      'Content-Type': 'application/json',
      ...options.headers,
    };

    if (this.token) {
      headers['Authorization'] = `Bearer ${this.token}`;
    }

    const response = await fetch(`${API_BASE}${endpoint}`, {
      ...options,
      headers,
    });

    if (!response.ok) {
      const error = await response.json().catch(() => ({ error: 'Request failed' }));
      throw new Error(error.error || error.message || 'Request failed');
    }

    return response.json();
  }

  // Auth
  async login(data: LoginRequest): Promise<LoginResponse> {
    return this.request<LoginResponse>('/auth/login', {
      method: 'POST',
      body: JSON.stringify(data),
    });
  }

  async getMe() {
    return this.request('/auth/me');
  }

  // Workflows
  async getWorkflows(): Promise<{ workflows: Workflow[] }> {
    return this.request('/workflows');
  }

  async getWorkflow(id: string): Promise<Workflow> {
    return this.request(`/workflows/${id}`);
  }

  async createWorkflow(data: Partial<Workflow>): Promise<Workflow> {
    return this.request('/workflows', {
      method: 'POST',
      body: JSON.stringify(data),
    });
  }

  async updateWorkflow(id: string, data: Partial<Workflow>): Promise<Workflow> {
    return this.request(`/workflows/${id}`, {
      method: 'PUT',
      body: JSON.stringify(data),
    });
  }

  async deleteWorkflow(id: string): Promise<void> {
    return this.request(`/workflows/${id}`, {
      method: 'DELETE',
    });
  }

  async triggerWorkflow(id: string): Promise<WorkflowRun> {
    return this.request(`/workflows/${id}/trigger`, {
      method: 'POST',
    });
  }

  // Runs
  async getRun(id: string): Promise<{ run: WorkflowRun; steps: any[] }> {
    return this.request(`/runs/${id}`);
  }

  async getRuns(workflowId?: string): Promise<{ runs: WorkflowRun[] }> {
    const params = workflowId ? `?workflow_id=${workflowId}` : '';
    return this.request(`/runs${params}`);
  }

  // Stats
  async getHealthStats(): Promise<HealthStats> {
    return this.request('/stats/health');
  }
}

export const api = new APIClient();
