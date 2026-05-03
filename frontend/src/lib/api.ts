const API_BASE =
  process.env.NEXT_PUBLIC_API_URL || "http://localhost:3000/api/v1";

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
  active: boolean;
  created_at: string;
  created_by: string;
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
  description: string;
  id: string;
  name: string;
  tenant_id: string;
  timeout_seconds: number;
  updated_at: string;
}

export interface WorkflowRun {
  completed_at?: string;
  created_at: string;
  created_by?: string;
  error?: string;
  id: string;
  started_at?: string;
  status: "pending" | "running" | "success" | "failed" | "cancelled";
  tenant_id: string;
  triggered_by: string;
  workflow_id: string;
}

export interface WorkflowVersion {
  created_at: string;
  created_by: string;
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
  id: string;
  version: number;
  workflow_id: string;
}

export interface Schedule {
  id: string;
  workflow_id: string;
  tenant_id: string;
  cron_expression: string;
  active: boolean;
  next_run_at: string;
  last_run_at?: string;
  created_by: string;
  created_at: string;
  updated_at: string;
}

export interface Webhook {
  id: string;
  workflow_id: string;
  tenant_id: string;
  path: string;
  secret: string;
  active: boolean;
  created_by: string;
  created_at: string;
  updated_at: string;
}

export interface HealthStats {
  active_runs: number;
  avg_duration_seconds: number;
  failed_runs_24h: number;
  failure_rate: number;
  hourly_stats: Array<{
    hour: number;
    total_runs: number;
    success_runs: number;
    failed_runs: number;
    avg_duration: number;
  }>;
  success_rate: number;
  success_runs_24h: number;
  total_runs_24h: number;
}

export class APIClient {
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
    const headers: Record<string, string> = {
      "Content-Type": "application/json",
      ...(options.headers as Record<string, string>),
    };

    if (this.token) {
      headers["Authorization"] = `Bearer ${this.token}`;
    }

    const response = await fetch(`${API_BASE}${endpoint}`, {
      ...options,
      headers,
    });

    if (!response.ok) {
      const error = await response
        .json()
        .catch(() => ({ error: "Request failed" }));
      throw new Error(error.error || error.message || "Request failed");
    }

    return response.json();
  }

  // Auth
  async login(data: LoginRequest): Promise<LoginResponse> {
    return this.request<LoginResponse>("/auth/login", {
      method: "POST",
      body: JSON.stringify(data),
    });
  }

  async getMe() {
    return this.request("/auth/me");
  }

  // Workflows
  async getWorkflows(): Promise<{ workflows: Workflow[] }> {
    const response = await this.request<{ data: Workflow[]; pagination: any }>(
      "/workflows"
    );
    return { workflows: response.data || [] };
  }

  async getWorkflow(id: string): Promise<Workflow> {
    return this.request(`/workflows/${id}`);
  }

  async createWorkflow(data: Partial<Workflow>): Promise<Workflow> {
    return this.request("/workflows", {
      method: "POST",
      body: JSON.stringify(data),
    });
  }

  async updateWorkflow(id: string, data: Partial<Workflow>): Promise<Workflow> {
    return this.request(`/workflows/${id}`, {
      method: "PUT",
      body: JSON.stringify(data),
    });
  }

  async deleteWorkflow(id: string): Promise<void> {
    return this.request(`/workflows/${id}`, {
      method: "DELETE",
    });
  }

  async triggerWorkflow(id: string): Promise<WorkflowRun> {
    return this.request(`/workflows/${id}/trigger`, {
      method: "POST",
    });
  }

  async getWorkflowVersions(
    id: string
  ): Promise<{ versions: WorkflowVersion[] }> {
    return this.request(`/workflows/${id}/versions`);
  }

  async rollbackWorkflow(id: string, version: number): Promise<Workflow> {
    return this.request(`/workflows/${id}/rollback/${version}`, {
      method: "POST",
    });
  }

  // Runs
  async getRun(id: string): Promise<{ run: WorkflowRun; steps: any[] }> {
    return this.request(`/runs/${id}`);
  }

  async getRuns(workflowId?: string): Promise<{ runs: WorkflowRun[] }> {
    const params = workflowId ? `?workflow_id=${workflowId}` : "";
    const response = await this.request<{
      data: WorkflowRun[];
      pagination: any;
    }>(`/runs${params}`);
    return { runs: response.data || [] };
  }

  // Stats
  async getHealthStats(): Promise<HealthStats> {
    return this.request("/stats/health");
  }

  // Schedules
  async getSchedules(): Promise<{ schedules: Schedule[] }> {
    const response = await this.request<{ schedules: Schedule[] }>("/schedules");
    return { schedules: response.schedules || [] };
  }

  async createSchedule(
    workflowId: string,
    cronExpression: string
  ): Promise<Schedule> {
    return this.request(`/workflows/${workflowId}/schedule`, {
      method: "POST",
      body: JSON.stringify({
        workflow_id: workflowId,
        cron_expression: cronExpression,
      }),
    });
  }

  async deleteSchedule(id: string): Promise<void> {
    return this.request(`/schedules/${id}`, {
      method: "DELETE",
    });
  }

  // Webhooks
  async getWebhooks(): Promise<{ webhooks: Webhook[] }> {
    const response = await this.request<{ webhooks: Webhook[] }>("/webhooks");
    return { webhooks: response.webhooks || [] };
  }

  async createWebhook(workflowId: string): Promise<Webhook> {
    return this.request(`/workflows/${workflowId}/webhook`, {
      method: "POST",
      body: JSON.stringify({ workflow_id: workflowId }),
    });
  }

  async deleteWebhook(id: string): Promise<void> {
    return this.request(`/webhooks/${id}`, {
      method: "DELETE",
    });
  }
}

export const api = new APIClient();
