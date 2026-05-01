'use client';

import { useState, useCallback, useEffect } from 'react';
import { useParams, useRouter } from 'next/navigation';
import ReactFlow, {
  Background,
  Controls,
  MiniMap,
  addEdge,
  Connection,
  Edge,
  Node,
  useNodesState,
  useEdgesState,
} from 'reactflow';
import 'reactflow/dist/style.css';
import { api, Workflow } from '@/lib/api';
import { nodeTypes } from '@/lib/nodeTypes';

export default function EditWorkflowPage() {
  const params = useParams();
  const router = useRouter();
  const [workflow, setWorkflow] = useState<Workflow | null>(null);
  const [nodes, setNodes, onNodesChange] = useNodesState([]);
  const [edges, setEdges, onEdgesChange] = useEdgesState([]);
  const [workflowName, setWorkflowName] = useState('');
  const [workflowDescription, setWorkflowDescription] = useState('');
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    if (params.id) {
      loadWorkflow(params.id as string);
    }
  }, [params.id]);

  const loadWorkflow = async (id: string) => {
    try {
      setLoading(true);
      const data = await api.getWorkflow(id);
      setWorkflow(data);
      setWorkflowName(data.name);
      setWorkflowDescription(data.description);

      // Convert backend format to ReactFlow format
      const flowNodes: Node[] = data.definition.nodes.map((node) => ({
        id: node.id,
        type: 'custom',
        position: node.position,
        data: {
          type: node.type,
          label: node.name,
          config: node.config,
        },
      }));

      const flowEdges: Edge[] = data.definition.edges.map((edge) => ({
        id: edge.id,
        source: edge.source,
        target: edge.target,
      }));

      setNodes(flowNodes);
      setEdges(flowEdges);
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Failed to load workflow');
      router.push('/dashboard');
    } finally {
      setLoading(false);
    }
  };

  const onConnect = useCallback(
    (params: Connection) => setEdges((eds) => addEdge(params, eds)),
    [setEdges]
  );

  const addNode = useCallback((type: string, label: string) => {
    const newNode: Node = {
      id: `${type}-${Date.now()}`,
      type: 'custom',
      position: { x: Math.random() * 400 + 100, y: Math.random() * 300 + 100 },
      data: {
        type,
        label,
        config: getDefaultConfig(type),
      },
    };
    setNodes((nds) => [...nds, newNode]);
  }, [setNodes]);

  const getDefaultConfig = (type: string) => {
    switch (type) {
      case 'http':
        return { url: '', method: 'GET', headers: {} };
      case 'delay':
        return { seconds: 5 };
      case 'script':
        return { code: '' };
      case 'condition':
        return { expression: 'true' };
      default:
        return {};
    }
  };

  const handleSave = async () => {
    if (!workflow) return;

    if (!workflowName.trim()) {
      alert('Please enter a workflow name');
      return;
    }

    if (nodes.length === 0) {
      alert('Please add at least one node');
      return;
    }

    try {
      setSaving(true);

      await api.updateWorkflow(workflow.id, {
        name: workflowName,
        description: workflowDescription,
        definition: {
          nodes: nodes.map((node) => ({
            id: node.id,
            type: node.data.type,
            name: node.data.label,
            config: node.data.config,
            position: node.position,
          })),
          edges: edges.map((edge) => ({
            id: edge.id,
            source: edge.source,
            target: edge.target,
          })),
        },
        timeout_seconds: workflow.timeout_seconds,
        active: workflow.active,
      });

      router.push(`/dashboard/workflows/${workflow.id}`);
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Failed to update workflow');
    } finally {
      setSaving(false);
    }
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center h-[calc(100vh-200px)]">
        <div className="text-xl">Loading workflow...</div>
      </div>
    );
  }

  return (
    <div className="h-[calc(100vh-200px)]">
      <div className="flex justify-between items-center mb-4">
        <div>
          <h2 className="text-2xl font-bold text-gray-900">Edit Workflow</h2>
          <p className="text-sm text-gray-600">Modify your workflow by adding or updating nodes</p>
        </div>
        <div className="flex space-x-2">
          <button
            onClick={() => router.back()}
            className="px-4 py-2 text-gray-700 bg-gray-100 rounded-md hover:bg-gray-200"
          >
            Cancel
          </button>
          <button
            onClick={handleSave}
            disabled={saving}
            className="px-4 py-2 bg-indigo-600 text-white rounded-md hover:bg-indigo-700 disabled:bg-indigo-400"
          >
            {saving ? 'Saving...' : 'Save Changes'}
          </button>
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-4 gap-4 h-[calc(100%-80px)]">
        {/* Sidebar */}
        <div className="lg:col-span-1 bg-white rounded-lg shadow p-4 overflow-y-auto">
          <h3 className="font-semibold mb-4">Workflow Info</h3>
          <div className="space-y-3 mb-6">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Name *
              </label>
              <input
                type="text"
                value={workflowName}
                onChange={(e) => setWorkflowName(e.target.value)}
                className="w-full px-3 py-2 border border-gray-300 rounded-md text-sm focus:outline-none focus:ring-indigo-500 focus:border-indigo-500"
                placeholder="My Workflow"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Description
              </label>
              <textarea
                value={workflowDescription}
                onChange={(e) => setWorkflowDescription(e.target.value)}
                className="w-full px-3 py-2 border border-gray-300 rounded-md text-sm focus:outline-none focus:ring-indigo-500 focus:border-indigo-500"
                rows={3}
                placeholder="What does this workflow do?"
              />
            </div>
          </div>

          <h3 className="font-semibold mb-3">Add Nodes</h3>
          <div className="space-y-2">
            <button
              onClick={() => addNode('http', 'HTTP Request')}
              className="w-full px-3 py-2 text-left text-sm bg-blue-50 text-blue-700 rounded-md hover:bg-blue-100 flex items-center gap-2"
            >
              <span>🌐</span> HTTP Request
            </button>
            <button
              onClick={() => addNode('delay', 'Delay')}
              className="w-full px-3 py-2 text-left text-sm bg-yellow-50 text-yellow-700 rounded-md hover:bg-yellow-100 flex items-center gap-2"
            >
              <span>⏱️</span> Delay
            </button>
            <button
              onClick={() => addNode('script', 'Script')}
              className="w-full px-3 py-2 text-left text-sm bg-purple-50 text-purple-700 rounded-md hover:bg-purple-100 flex items-center gap-2"
            >
              <span>📜</span> Script
            </button>
            <button
              onClick={() => addNode('condition', 'Condition')}
              className="w-full px-3 py-2 text-left text-sm bg-green-50 text-green-700 rounded-md hover:bg-green-100 flex items-center gap-2"
            >
              <span>❓</span> Condition
            </button>
          </div>

          <div className="mt-6 text-xs text-gray-500">
            <p className="mb-2">Tips:</p>
            <ul className="list-disc list-inside space-y-1">
              <li>Drag nodes to reposition</li>
              <li>Connect from bottom to top</li>
              <li>Delete nodes with Backspace</li>
              <li>Select node and Delete to remove</li>
            </ul>
          </div>
        </div>

        {/* Canvas */}
        <div className="lg:col-span-3 bg-white rounded-lg shadow">
          <ReactFlow
            nodes={nodes}
            edges={edges}
            onNodesChange={onNodesChange}
            onEdgesChange={onEdgesChange}
            onConnect={onConnect}
            nodeTypes={nodeTypes}
            fitView
          >
            <Background />
            <Controls />
            <MiniMap />
          </ReactFlow>
        </div>
      </div>
    </div>
  );
}
