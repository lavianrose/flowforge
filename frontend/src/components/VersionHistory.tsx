'use client';

import { useWorkflowVersions, useRollbackWorkflow } from '@/lib/hooks';
import { WorkflowVersion } from '@/lib/api';
import { useSnackbar } from '@/components/Snackbar';

interface VersionHistoryProps {
  workflowId: string;
}

export default function VersionHistory({ workflowId }: VersionHistoryProps) {
  const { data: versions, isLoading, error } = useWorkflowVersions(workflowId);
  const rollbackMutation = useRollbackWorkflow();
  const { showSnackbar } = useSnackbar();

  const handleRollback = async (version: number) => {
    if (!confirm(`Are you sure you want to rollback to version ${version}? This will create a new version with the old definition.`)) {
      return;
    }

    try {
      await rollbackMutation.mutateAsync({ id: workflowId, version });
      showSnackbar(`Successfully rolled back to version ${version}`, 'success');
      window.location.reload();
    } catch (err) {
      showSnackbar(err instanceof Error ? err.message : 'Failed to rollback workflow', 'error');
    }
  };

  if (isLoading) {
    return <div className="text-center py-4 text-gray-500">Loading versions...</div>;
  }

  if (error) {
    return (
      <div className="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded">
        {error.message}
      </div>
    );
  }

  if (!versions || versions.length === 0) {
    return (
      <div className="text-center py-4 text-gray-500">
        No version history available. Create your first version by updating the workflow.
      </div>
    );
  }

  // Sort versions by version number descending (newest first)
  const sortedVersions = [...versions].sort((a, b) => b.version - a.version);

  // The latest version is the current one (highest version number)
  const latestVersion = sortedVersions[0].version;

  return (
    <div className="space-y-3">
      <h3 className="text-lg font-semibold text-gray-900">Version History</h3>
      <div className="space-y-2 max-h-96 overflow-y-auto">
        {sortedVersions.map((version: WorkflowVersion) => (
          <div
            key={version.id}
            className={`border rounded-lg p-4 ${
              version.version === latestVersion
                ? 'border-indigo-300 bg-indigo-50'
                : 'border-gray-200 bg-white'
            }`}
          >
            <div className="flex items-center justify-between">
              <div className="flex-1">
                <div className="flex items-center gap-2">
                  <span className="font-medium text-gray-900">
                    Version {version.version}
                  </span>
                  {version.version === latestVersion && (
                    <span className="px-2 py-1 text-xs font-medium text-indigo-700 bg-indigo-100 rounded-full">
                      Current
                    </span>
                  )}
                </div>
                <div className="mt-1 text-sm text-gray-500">
                  <div className="flex flex-wrap items-center gap-2">
                    <span>
                      {version.definition.nodes.length} nodes
                    </span>
                    <span>•</span>
                    <span>
                      {version.definition.edges.length} connections
                    </span>
                    <span>•</span>
                    <span>
                      by {version.created_by}
                    </span>
                    <span>•</span>
                    <span className="text-xs">
                      {new Date(version.created_at).toLocaleDateString()} {new Date(version.created_at).toLocaleTimeString()}
                    </span>
                  </div>
                </div>
              </div>
              {version.version !== latestVersion && (
                <button
                  onClick={() => handleRollback(version.version)}
                  className="px-3 py-1.5 text-sm font-medium text-indigo-700 bg-indigo-50 rounded-md hover:bg-indigo-100 disabled:opacity-50"
                  disabled={rollbackMutation.isPending}
                >
                  {rollbackMutation.isPending ? 'Rolling back...' : 'Rollback'}
                </button>
              )}
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
