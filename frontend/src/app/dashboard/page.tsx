'use client';

import { useState, useEffect } from 'react';
import { api } from '@/lib/api';
import { Workflow } from '@/lib/api';

export default function DashboardPage() {
  const [workflows, setWorkflows] = useState<Workflow[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  useEffect(() => {
    loadWorkflows();
  }, []);

  const loadWorkflows = async () => {
    try {
      setLoading(true);
      const data = await api.getWorkflows();
      setWorkflows(data.workflows);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load workflows');
    } finally {
      setLoading(false);
    }
  };

  const handleTrigger = async (id: string) => {
    try {
      await api.triggerWorkflow(id);
      alert('Workflow triggered successfully');
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Failed to trigger workflow');
    }
  };

  if (loading) {
    return <div className="text-center py-12">Loading workflows...</div>;
  }

  if (error) {
    return (
      <div className="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded">
        {error}
      </div>
    );
  }

  return (
    <div>
      <div className="flex justify-between items-center mb-6">
        <h2 className="text-2xl font-bold text-gray-900">Workflows</h2>
        <button
          onClick={() => (window.location.href = '/workflows/new')}
          className="px-4 py-2 bg-indigo-600 text-white rounded-md hover:bg-indigo-700"
        >
          Create Workflow
        </button>
      </div>

      {workflows.length === 0 ? (
        <div className="text-center py-12 bg-white rounded-lg shadow">
          <p className="text-gray-500 mb-4">No workflows found</p>
          <button
            onClick={() => (window.location.href = '/workflows/new')}
            className="px-4 py-2 bg-indigo-600 text-white rounded-md hover:bg-indigo-700"
          >
            Create Your First Workflow
          </button>
        </div>
      ) : (
        <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
          {workflows.map((workflow) => (
            <div key={workflow.id} className="bg-white rounded-lg shadow p-6">
              <h3 className="text-lg font-semibold text-gray-900 mb-2">
                {workflow.name}
              </h3>
              <p className="text-sm text-gray-600 mb-4">
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
                  onClick={() => (window.location.href = `/workflows/${workflow.id}`)}
                  className="flex-1 px-3 py-2 text-sm font-medium text-indigo-600 bg-indigo-50 rounded-md hover:bg-indigo-100"
                >
                  View
                </button>
                <button
                  onClick={() => handleTrigger(workflow.id)}
                  className="flex-1 px-3 py-2 text-sm font-medium text-white bg-green-600 rounded-md hover:bg-green-700"
                >
                  Run
                </button>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
