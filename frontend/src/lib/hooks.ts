"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { api, type LoginRequest, type Schedule, type Webhook, type Workflow, type WorkflowRun } from "./api";

// Auth hooks
export function useLogin() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: LoginRequest) => api.login(data),
    onSuccess: (data) => {
      api.setToken(data.token);
      queryClient.invalidateQueries({ queryKey: ["me"] });
    },
  });
}

export function useMe() {
  return useQuery({
    queryKey: ["me"],
    queryFn: () => api.getMe(),
    retry: false,
  });
}

// Workflow hooks
export function useWorkflows() {
  return useQuery({
    queryKey: ["workflows"],
    queryFn: async () => {
      const data = await api.getWorkflows();
      return data.workflows;
    },
    staleTime: 30 * 1000, // 30 seconds
  });
}

export function useWorkflow(id: string) {
  return useQuery({
    queryKey: ["workflow", id],
    queryFn: () => api.getWorkflow(id),
    enabled: !!id,
  });
}

export function useCreateWorkflow() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: Partial<Workflow>) => api.createWorkflow(data),
    onMutate: async (newWorkflow) => {
      // Cancel outgoing refetches
      await queryClient.cancelQueries({ queryKey: ["workflows"] });

      // Snapshot previous value
      const previousWorkflows = queryClient.getQueryData(["workflows"]);

      // Optimistically add workflow to list
      const optimisticWorkflow = {
        ...newWorkflow,
        id: `temp-${Date.now()}`,
        active: false,
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString(),
        timeout_seconds: 300,
        definition: { nodes: [], edges: [] },
      };

      queryClient.setQueryData(["workflows"], (old: any) => [
        optimisticWorkflow,
        ...(old || []),
      ]);

      // Return context with previous value for rollback
      return { previousWorkflows };
    },
    onError: (err, variables, context) => {
      // Rollback to previous value on error
      if (context?.previousWorkflows) {
        queryClient.setQueryData(["workflows"], context.previousWorkflows);
      }
    },
    onSuccess: () => {
      // Refetch to get the actual data from server
      queryClient.invalidateQueries({ queryKey: ["workflows"] });
    },
  });
}

export function useUpdateWorkflow() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: Partial<Workflow> }) =>
      api.updateWorkflow(id, data),
    onMutate: async ({ id, data }) => {
      // Cancel outgoing refetches
      await queryClient.cancelQueries({ queryKey: ["workflow", id] });
      await queryClient.cancelQueries({ queryKey: ["workflows"] });

      // Snapshot previous values
      const previousWorkflow = queryClient.getQueryData(["workflow", id]);
      const previousWorkflows = queryClient.getQueryData(["workflows"]);

      // Optimistically update workflow
      queryClient.setQueryData(["workflow", id], (old: any) => ({
        ...old,
        ...data,
        updated_at: new Date().toISOString(),
      }));

      queryClient.setQueryData(
        ["workflows"],
        (old: any) =>
          old?.map((w: any) =>
            w.id === id
              ? { ...w, ...data, updated_at: new Date().toISOString() }
              : w
          ) || []
      );

      // Return context with previous values for rollback
      return { previousWorkflow, previousWorkflows };
    },
    onError: (err, variables, context) => {
      // Rollback to previous values on error
      if (context?.previousWorkflow) {
        queryClient.setQueryData(
          ["workflow", variables.id],
          context.previousWorkflow
        );
      }
      if (context?.previousWorkflows) {
        queryClient.setQueryData(["workflows"], context.previousWorkflows);
      }
    },
    onSuccess: (_, variables) => {
      // Refetch to get the actual data from server
      queryClient.invalidateQueries({ queryKey: ["workflow", variables.id] });
      queryClient.invalidateQueries({ queryKey: ["workflows"] });
    },
  });
}

export function useDeleteWorkflow() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: string) => api.deleteWorkflow(id),
    onMutate: async (workflowId) => {
      // Cancel outgoing refetches
      await queryClient.cancelQueries({ queryKey: ["workflows"] });

      // Snapshot previous value
      const previousWorkflows = queryClient.getQueryData(["workflows"]);

      // Optimistically remove workflow from list
      queryClient.setQueryData(
        ["workflows"],
        (old: any) => old?.filter((w: any) => w.id !== workflowId) || []
      );

      // Return context with previous value for rollback
      return { previousWorkflows };
    },
    onError: (err, variables, context) => {
      // Rollback to previous value on error
      if (context?.previousWorkflows) {
        queryClient.setQueryData(["workflows"], context.previousWorkflows);
      }
    },
    onSuccess: () => {
      // Refetch to ensure consistency
      queryClient.invalidateQueries({ queryKey: ["workflows"] });
    },
  });
}

export function useTriggerWorkflow() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: string) => api.triggerWorkflow(id),
    onMutate: async (workflowId) => {
      // Cancel outgoing refetches
      await queryClient.cancelQueries({ queryKey: ["runs"] });

      // Snapshot previous value
      const previousRuns = queryClient.getQueryData(["runs"]);

      // Optimistically update runs list
      queryClient.setQueryData(["runs"], (old: any) => {
        const optimisticRun = {
          id: `temp-${Date.now()}`,
          workflow_id: workflowId,
          status: "pending",
          triggered_by: "you",
          started_at: new Date().toISOString(),
          completed_at: null,
          error: null,
        };

        return {
          runs: [optimisticRun, ...(old?.runs || [])],
        };
      });

      // Return context with previous value for rollback
      return { previousRuns };
    },
    onError: (err, variables, context) => {
      // Rollback to previous value on error
      if (context?.previousRuns) {
        queryClient.setQueryData(["runs"], context.previousRuns);
      }
    },
    onSuccess: () => {
      // Refetch to get the actual data from server
      queryClient.invalidateQueries({ queryKey: ["runs"] });
      queryClient.invalidateQueries({ queryKey: ["stats"] });
    },
  });
}

// Run hooks
export function useRuns(workflowId?: string) {
  return useQuery({
    queryKey: ["runs", workflowId],
    queryFn: async () => {
      const data = await api.getRuns(workflowId);
      return data.runs;
    },
    staleTime: 15 * 1000, // 15 seconds
  });
}

export function useRun(id: string) {
  return useQuery<{ run: WorkflowRun; steps: any[] }>({
    queryKey: ["run", id],
    queryFn: () => api.getRun(id),
    enabled: !!id,
    refetchInterval(data: any) {
      // Poll every 1 second if run is pending or running
      if (data?.run?.status === "pending" || data?.run?.status === "running") {
        return 1000;
      }
      return false;
    },
  });
}

// Stats hooks
export function useHealthStats() {
  return useQuery({
    queryKey: ["stats", "health"],
    queryFn: () => api.getHealthStats(),
    refetchInterval: 30 * 1000, // Auto-refresh every 30 seconds
  });
}

// Version hooks
export function useWorkflowVersions(id: string) {
  return useQuery({
    queryKey: ["workflow-versions", id],
    queryFn: async () => {
      const data = await api.getWorkflowVersions(id);
      return data.versions;
    },
    enabled: !!id,
    staleTime: 60 * 1000, // 1 minute
  });
}

export function useRollbackWorkflow() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, version }: { id: string; version: number }) =>
      api.rollbackWorkflow(id, version),
    onSuccess: (_, variables) => {
      // Invalidate workflow and versions queries
      queryClient.invalidateQueries({ queryKey: ["workflow", variables.id] });
      queryClient.invalidateQueries({
        queryKey: ["workflow-versions", variables.id],
      });
      queryClient.invalidateQueries({ queryKey: ["workflows"] });
    },
  });
}

// Schedule hooks
export function useSchedules() {
  return useQuery({
    queryKey: ["schedules"],
    queryFn: async () => {
      const data = await api.getSchedules();
      return data.schedules;
    },
    staleTime: 30 * 1000,
  });
}

export function useCreateSchedule() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({
      workflowId,
      cronExpression,
    }: {
      workflowId: string;
      cronExpression: string;
    }) => api.createSchedule(workflowId, cronExpression),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["schedules"] });
    },
  });
}

export function useDeleteSchedule() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: string) => api.deleteSchedule(id),
    onMutate: async (scheduleId) => {
      await queryClient.cancelQueries({ queryKey: ["schedules"] });
      const previousSchedules = queryClient.getQueryData(["schedules"]);
      queryClient.setQueryData(
        ["schedules"],
        (old: Schedule[] | undefined) =>
          old?.filter((s) => s.id !== scheduleId) || []
      );
      return { previousSchedules };
    },
    onError: (_err, _variables, context) => {
      if (context?.previousSchedules) {
        queryClient.setQueryData(["schedules"], context.previousSchedules);
      }
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["schedules"] });
    },
  });
}

// Webhook hooks
export function useWebhooks() {
  return useQuery({
    queryKey: ["webhooks"],
    queryFn: async () => {
      const data = await api.getWebhooks();
      return data.webhooks;
    },
    staleTime: 30 * 1000,
  });
}

export function useCreateWebhook() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (workflowId: string) => api.createWebhook(workflowId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["webhooks"] });
    },
  });
}

export function useDeleteWebhook() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: string) => api.deleteWebhook(id),
    onMutate: async (webhookId) => {
      await queryClient.cancelQueries({ queryKey: ["webhooks"] });
      const previousWebhooks = queryClient.getQueryData(["webhooks"]);
      queryClient.setQueryData(
        ["webhooks"],
        (old: Webhook[] | undefined) =>
          old?.filter((w) => w.id !== webhookId) || []
      );
      return { previousWebhooks };
    },
    onError: (_err, _variables, context) => {
      if (context?.previousWebhooks) {
        queryClient.setQueryData(["webhooks"], context.previousWebhooks);
      }
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["webhooks"] });
    },
  });
}
