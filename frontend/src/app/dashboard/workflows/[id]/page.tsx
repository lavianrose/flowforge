'use client';

import { useParams, useRouter } from 'next/navigation';
import { useWorkflow, useDeleteWorkflow, useTriggerWorkflow } from '@/lib/hooks';
import ReactFlow, { Background, Controls, MiniMap } from 'reactflow';
import 'reactflow/dist/style.css';
import { nodeTypes } from '@/lib/nodeTypes';

export default function WorkflowDetailPage() {
  const params = useParams();
  const router = useRouter();
  const { data: workflow, isLoading, error } = useWorkflow(params.id as string);
  const deleteMutation = useDeleteWorkflow();
  const triggerMutation = useTriggerWorkflow();

  // Convert workflow definition to ReactFlow format
  const nodes = workflow ? workflow.definition.nodes.map((node) => ({
    id: node.id,
    type: 'custom',
    position: node.position,
    data: {
      type: node.type,
      label: node.name,
      config: node.config,
    },
  })) : [];

  const edges = workflow ? workflow.definition.edges.map((edge) => ({
    id: edge.id,
    source: edge.source,
    target: edge.target,
  })) : [];

  const handleDelete = async () => {
    if (!workflow) return;
    if (!confirm('Are you sure you want to delete this workflow?')) {
      return;
    }

    try {
      await deleteMutation.mutateAsync(workflow.id);
      router.push('/dashboard/workflows');
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Failed to delete workflow');
    }
  };

  const handleTrigger = async () => {
    if (!workflow) return;

    try {
      await triggerMutation.mutateAsync(workflow.id);
      alert('Workflow triggered successfully');
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Failed to trigger workflow');
    }
  };

  if (isLoading) {
    return <div className="text-center py-12">Loading workflow...</div>;
  }

  if (error || !workflow) {
    return (
      <div className="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded">
        {error?.message || 'Workflow not found'}
      </div>
    );
  }

  return (
    <div>
      <div className="flex justify-between items-center mb-6">
        <div>
          <h2 className="text-2xl font-bold text-gray-900">{workflow.name}</h2>
          <p className="text-sm text-gray-600 mt-1">{workflow.description}</p>
        </div>
        <div className="flex space-x-2">
          <button
            onClick={() => router.push('/dashboard/workflows')}
            className="px-4 py-2 text-gray-700 bg-gray-100 rounded-md hover:bg-gray-200"
          >
            Back
          </button>
          <button
            onClick={handleTrigger}
            disabled={triggerMutation.isPending}
            className="px-4 py-2 bg-green-600 text-white rounded-md hover:bg-green-700 disabled:opacity-50"
          >
            {triggerMutation.isPending ? 'Running...' : 'Run Workflow'}
          </button>
          <button
            onClick={handleDelete}
            disabled={deleteMutation.isPending}
            className="px-4 py-2 bg-red-600 text-white rounded-md hover:bg-red-700 disabled:opacity-50"
          >
            {deleteMutation.isPending ? 'Deleting...' : 'Delete'}
          </button>
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Workflow Info */}
        <div className="bg-white rounded-lg shadow p-6">
          <h3 className="text-lg font-semibold mb-4">Workflow Details</h3>
          <dl className="space-y-3">
            <div>
              <dt className="text-sm font-medium text-gray-500">Status</dt>
              <dd className="mt-1">
                <span
                  className={`px-2 py-1 rounded text-sm ${
                    workflow.active
                      ? 'bg-green-100 text-green-800'
                      : 'bg-gray-100 text-gray-800'
                  }`}
                >
                  {workflow.active ? 'Active' : 'Inactive'}
                </span>
              </dd>
            </div>
            <div>
              <dt className="text-sm font-medium text-gray-500">Timeout</dt>
              <dd className="mt-1 text-sm text-gray-900">
                {workflow.timeout_seconds} seconds
              </dd>
            </div>
            <div>
              <dt className="text-sm font-medium text-gray-500">Created</dt>
              <dd className="mt-1 text-sm text-gray-900">
                {new Date(workflow.created_at).toLocaleString()}
              </dd>
            </div>
            <div>
              <dt className="text-sm font-medium text-gray-500">Last Updated</dt>
              <dd className="mt-1 text-sm text-gray-900">
                {new Date(workflow.updated_at).toLocaleString()}
              </dd>
            </div>
          </dl>
        </div>

        {/* Nodes */}
        <div className="bg-white rounded-lg shadow p-6">
          <h3 className="text-lg font-semibold mb-4">Nodes ({workflow.definition.nodes.length})</h3>
          <div className="space-y-2">
            {workflow.definition.nodes.map((node) => (
              <div
                key={node.id}
                className="p-3 bg-gray-50 rounded-md border border-gray-200"
              >
                <div className="flex items-center justify-between">
                  <div>
                    <p className="font-medium text-gray-900">{node.name}</p>
                    <p className="text-xs text-gray-500">Type: {node.type}</p>
                  </div>
                  <span className="text-xs bg-blue-100 text-blue-800 px-2 py-1 rounded">
                    {node.id}
                  </span>
                </div>
              </div>
            ))}
          </div>
        </div>

        {/* Edges */}
        <div className="bg-white rounded-lg shadow p-6 lg:col-span-2">
          <h3 className="text-lg font-semibold mb-4">Connections ({workflow.definition.edges.length})</h3>
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-2">
            {workflow.definition.edges.map((edge) => (
              <div
                key={edge.id}
                className="p-3 bg-gray-50 rounded-md border border-gray-200 text-sm"
              >
                <div className="flex items-center justify-between">
                  <span className="text-gray-600">{edge.source}</span>
                  <span className="text-gray-400">→</span>
                  <span className="text-gray-600">{edge.target}</span>
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>

      {/* Visual DAG Viewer */}
      <div className="mt-6 bg-white rounded-lg shadow p-6">
        <h3 className="text-lg font-semibold mb-4">Workflow Visualization</h3>
        <div className="border border-gray-200 rounded-lg overflow-hidden" style={{ height: '500px' }}>
          <ReactFlow
            nodes={nodes}
            edges={edges}
            nodeTypes={nodeTypes}
            nodesDraggable={false}
            nodesConnectable={false}
            elementsSelectable={false}
            zoomOnScroll={true}
            panOnScroll={true}
            fitView
          >
            <Background />
            <Controls />
            <MiniMap />
          </ReactFlow>
        </div>
        <p className="text-xs text-gray-500 mt-2">
          💡 This is a read-only view. Use the "Edit Workflow" button to make changes.
        </p>
      </div>

      {/* Edit Workflow Button */}
      <div className="mt-6 flex justify-end">
        <button
          onClick={() => router.push(`/dashboard/workflows/${workflow.id}/edit`)}
          className="px-4 py-2 bg-indigo-600 text-white rounded-md hover:bg-indigo-700"
        >
          Edit Workflow
        </button>
      </div>
    </div>
  );
}
