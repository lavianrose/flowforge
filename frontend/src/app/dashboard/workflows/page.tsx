"use client";

import { useRouter } from "next/navigation";
import { useSnackbar } from "@/components/Snackbar";
import { useAuth } from "@/lib/auth";
import {
  useDeleteWorkflow,
  useTriggerWorkflow,
  useWorkflows,
} from "@/lib/hooks";

export default function WorkflowsPage() {
  const router = useRouter();
  const { can } = useAuth();
  const { showSnackbar } = useSnackbar();
  const { data: workflows, isLoading, error } = useWorkflows();
  const deleteMutation = useDeleteWorkflow();
  const triggerMutation = useTriggerWorkflow();

  const handleDelete = async (id: string) => {
    if (!confirm("Are you sure you want to delete this workflow?")) {
      return;
    }

    try {
      await deleteMutation.mutateAsync(id);
      showSnackbar("Workflow deleted successfully", "success");
    } catch (err) {
      showSnackbar(
        err instanceof Error ? err.message : "Failed to delete workflow",
        "error"
      );
    }
  };

  const handleTrigger = async (id: string) => {
    try {
      await triggerMutation.mutateAsync(id);
      showSnackbar("Workflow triggered successfully", "success");
    } catch (err) {
      showSnackbar(
        err instanceof Error ? err.message : "Failed to trigger workflow",
        "error"
      );
    }
  };

  if (isLoading) {
    return <div className="py-12 text-center">Loading workflows...</div>;
  }

  if (error) {
    return (
      <div className="rounded border border-red-200 bg-red-50 px-4 py-3 text-red-700">
        {error.message}
      </div>
    );
  }

  return (
    <div>
      <div className="mb-6 flex items-center justify-between">
        <div>
          <h2 className="font-bold text-2xl text-gray-900">Workflows</h2>
          <p className="mt-1 text-gray-600 text-sm">
            Manage and monitor your automation workflows
          </p>
        </div>
        {can("create") && (
          <button
            className="rounded-md bg-indigo-600 px-4 py-2 text-white hover:bg-indigo-700"
            onClick={() => router.push("/dashboard/workflows/new")}
          >
            Create Workflow
          </button>
        )}
      </div>

      {!workflows || workflows.length === 0 ? (
        <div className="rounded-lg bg-white py-12 text-center shadow">
          <p className="mb-4 text-gray-500">No workflows found</p>
          {can("create") && (
            <button
              className="rounded-md bg-indigo-600 px-4 py-2 text-white hover:bg-indigo-700"
              onClick={() => router.push("/dashboard/workflows/new")}
            >
              Create Your First Workflow
            </button>
          )}
        </div>
      ) : (
        <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
          {workflows.map((workflow) => (
            <div className="rounded-lg bg-white p-6 shadow" key={workflow.id}>
              <h3 className="mb-2 font-semibold text-gray-900 text-lg">
                {workflow.name}
              </h3>
              <p className="mb-4 line-clamp-2 h-10 text-gray-600 text-sm">
                {workflow.description || "No description"}
              </p>
              <div className="mb-4 flex items-center justify-between text-gray-500 text-sm">
                <span
                  className={`rounded px-2 py-1 ${
                    workflow.active
                      ? "bg-green-100 text-green-800"
                      : "bg-gray-100 text-gray-800"
                  }`}
                >
                  {workflow.active ? "Active" : "Inactive"}
                </span>
                <span>{workflow.definition.nodes.length} nodes</span>
              </div>
              <div className="flex space-x-2">
                <button
                  className="flex-1 rounded-md bg-indigo-50 px-3 py-2 font-medium text-indigo-600 text-sm hover:bg-indigo-100 disabled:opacity-50"
                  disabled={triggerMutation.isPending}
                  onClick={() =>
                    router.push(`/dashboard/workflows/${workflow.id}`)
                  }
                >
                  View
                </button>
                {can("trigger") && (
                  <button
                    className="flex-1 rounded-md bg-green-600 px-3 py-2 font-medium text-sm text-white hover:bg-green-700 disabled:opacity-50"
                    disabled={triggerMutation.isPending}
                    onClick={() => handleTrigger(workflow.id)}
                  >
                    {triggerMutation.isPending ? "Running..." : "Run"}
                  </button>
                )}
              </div>
              {can("edit") && (
                <div className="mt-3 border-gray-200 border-t pt-3">
                  <button
                    className="w-full rounded-md bg-gray-50 px-3 py-2 font-medium text-gray-700 text-sm hover:bg-gray-100 disabled:opacity-50"
                    disabled={triggerMutation.isPending}
                    onClick={() =>
                      router.push(`/dashboard/workflows/${workflow.id}/edit`)
                    }
                  >
                    Edit Workflow
                  </button>
                </div>
              )}
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
