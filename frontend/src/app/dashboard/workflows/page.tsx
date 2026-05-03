'use client';

import { useWorkflows, useDeleteWorkflow, useTriggerWorkflow } from '@/lib/hooks';
import { useRouter } from 'next/navigation';
import { useAuth } from '@/lib/auth';
import { useSnackbar } from '@/components/Snackbar';

export default function WorkflowsPage() {
  const router = useRouter();
  const { can } = useAuth();
  const { showSnackbar } = useSnackbar();
  const { data: workflows, isLoading, error } = useWorkflows();
  const deleteMutation = useDeleteWorkflow();
  const triggerMutation = useTriggerWorkflow();

  const handleDelete = async (id: string) => {
    if (!confirm('Are you sure you want to delete this workflow?')) {
      return;
    }

    try {
      await deleteMutation.mutateAsync(id);
      showSnackbar('Workflow deleted successfully', 'success');
    } catch (err) {
      showSnackbar(err instanceof Error ? err.message : 'Failed to delete workflow', 'error');
    }
  };

  const handleTrigger = async (id: string) => {
    try {
      await triggerMutation.mutateAsync(id);
      showSnackbar('Workflow triggered successfully', 'success');
    } catch (err) {
      showSnackbar(err instanceof Error ? err.message : 'Failed to trigger workflow', 'error');
    }
  };

  if (isLoading) {
    return <div className="text-center py-12">Loading workflows...</div>;
  }

  if (error) {
    return (
      <div className="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded">
        {error.message}
      </div>
    );
  }

  return (
    <div>
      <div className="flex justify-between items-center mb-6">
        <div>
          <h2 className="text-2xl font-bold text-gray-900">Workflows</h2>
          <p className="text-sm text-gray-600 mt-1">
            Manage and monitor your automation workflows
          </p>
        </div>
        {can('create') && (
          <button
            onClick={() => router.push('/dashboard/workflows/new')}
            className="px-4 py-2 bg-indigo-600 text-white rounded-md hover:bg-indigo-700"
          >
            Create Workflow
          </button>
        )}
      </div>

      {!workflows || workflows.length === 0 ? (
        <div className="text-center py-12 bg-white rounded-lg shadow">
          <p className="text-gray-500 mb-4">No workflows found</p>
          {can('create') && (
            <button
              onClick={() => router.push('/dashboard/workflows/new')}
              className="px-4 py-2 bg-indigo-600 text-white rounded-md hover:bg-indigo-700"
            >
              Create Your First Workflow
            </button>
          )}
        </div>
      ) : (
        <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
          {workflows.map((workflow) => (
            <div key={workflow.id} className="bg-white rounded-lg shadow p-6">
              <h3 className="text-lg font-semibold text-gray-900 mb-2">
                {workflow.name}
              </h3>
              <p className="text-sm text-gray-600 mb-4 h-10 line-clamp-2">
                {workflow.description || 'No description'}
              </p>
              <div className="flex items-center justify-between text-sm text-gray-500 mb-4">
                <span
                  className={`px-2 py-1 rounded ${
                    workflow.active
                      ? 'bg-green-100 text-green-800'
                      : 'bg-gray-100 text-gray-800'
                  }`}
                >
                  {workflow.active ? 'Active' : 'Inactive'}
                </span>
                <span>{workflow.definition.nodes.length} nodes</span>
              </div>
              <div className="flex space-x-2">
                <button
                  onClick={() => router.push(`/dashboard/workflows/${workflow.id}`)}
                  className="flex-1 px-3 py-2 text-sm font-medium text-indigo-600 bg-indigo-50 rounded-md hover:bg-indigo-100 disabled:opacity-50"
                  disabled={triggerMutation.isPending}
                >
                  View
                </button>
                {can('trigger') && (
                  <button
                    onClick={() => handleTrigger(workflow.id)}
                    className="flex-1 px-3 py-2 text-sm font-medium text-white bg-green-600 rounded-md hover:bg-green-700 disabled:opacity-50"
                    disabled={triggerMutation.isPending}
                  >
                    {triggerMutation.isPending ? 'Running...' : 'Run'}
                  </button>
                )}
              </div>
              {can('edit') && (
                <div className="mt-3 pt-3 border-t border-gray-200">
                  <button
                    onClick={() => router.push(`/dashboard/workflows/${workflow.id}/edit`)}
                    className="w-full px-3 py-2 text-sm font-medium text-gray-700 bg-gray-50 rounded-md hover:bg-gray-100 disabled:opacity-50"
                    disabled={triggerMutation.isPending}
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
