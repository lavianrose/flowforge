"use client";

import { useParams, useRouter } from "next/navigation";
import { useCallback, useEffect, useState } from "react";
import ReactFlow, {
  addEdge,
  Background,
  type Connection,
  Controls,
  type Edge,
  MiniMap,
  type Node,
  type NodeMouseHandler,
  useEdgesState,
  useNodesState,
} from "reactflow";
import "reactflow/dist/style.css";
import NodeConfigPanel from "@/components/nodes/NodeConfigPanel";
import { useSnackbar } from "@/components/Snackbar";
import { api, type Workflow } from "@/lib/api";
import { nodeTypes } from "@/lib/nodeTypes";

const getDefaultConfig = (type: string) => {
  switch (type) {
    case "http":
      return { url: "", method: "GET", headers: {} };
    case "delay":
      return { seconds: 5 };
    case "script":
      return { code: "" };
    case "condition":
      return { expression: "" };
    default:
      return {};
  }
};

export default function EditWorkflowPage() {
  const params = useParams();
  const router = useRouter();
  const [workflow, setWorkflow] = useState<Workflow | null>(null);
  const [nodes, setNodes, onNodesChange] = useNodesState([]);
  const [edges, setEdges, onEdgesChange] = useEdgesState([]);
  const [workflowName, setWorkflowName] = useState("");
  const [workflowDescription, setWorkflowDescription] = useState("");
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [selectedNodeId, setSelectedNodeId] = useState<string | null>(null);
  const { showSnackbar } = useSnackbar();

  const selectedNode = nodes.find((n) => n.id === selectedNodeId) || null;

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

      const flowNodes: Node[] = data.definition.nodes.map((node) => ({
        id: node.id,
        type: "custom" as const,
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
      showSnackbar(
        err instanceof Error ? err.message : "Failed to load workflow",
        "error",
      );
      router.push("/dashboard");
    } finally {
      setLoading(false);
    }
  };

  const onConnect = useCallback(
    (params: Connection) => setEdges((eds) => addEdge(params, eds)),
    [setEdges],
  );

  const onNodeClick: NodeMouseHandler = useCallback((_event, node) => {
    setSelectedNodeId(node.id);
  }, []);

  const onPaneClick = useCallback(() => {
    setSelectedNodeId(null);
  }, []);

  const addNode = useCallback(
    (type: string, label: string) => {
      const newNode: Node = {
        id: `${type}-${Date.now()}`,
        type: "custom",
        position: {
          x: Math.random() * 400 + 100,
          y: Math.random() * 300 + 100,
        },
        data: {
          type,
          label,
          config: getDefaultConfig(type),
        },
      };
      setNodes((nds) => [...nds, newNode]);
      setSelectedNodeId(newNode.id);
    },
    [setNodes],
  );

  const handleConfigChange = useCallback(
    (nodeId: string, config: Record<string, unknown>) => {
      setNodes((nds) =>
        nds.map((n) =>
          n.id === nodeId ? { ...n, data: { ...n.data, config } } : n,
        ),
      );
    },
    [setNodes],
  );

  const handleLabelChange = useCallback(
    (nodeId: string, label: string) => {
      setNodes((nds) =>
        nds.map((n) =>
          n.id === nodeId ? { ...n, data: { ...n.data, label } } : n,
        ),
      );
    },
    [setNodes],
  );

  const handleSave = async () => {
    if (!workflow) {
      return;
    }

    if (!workflowName.trim()) {
      showSnackbar("Please enter a workflow name", "warning");
      return;
    }

    if (nodes.length === 0) {
      showSnackbar("Please add at least one node", "warning");
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
      showSnackbar(
        err instanceof Error ? err.message : "Failed to update workflow",
        "error",
      );
    } finally {
      setSaving(false);
    }
  };

  if (loading) {
    return (
      <div className="flex h-[calc(100vh-200px)] items-center justify-center">
        <div className="text-xl">Loading workflow...</div>
      </div>
    );
  }

  return (
    <div className="h-[calc(100vh-200px)]">
      <div className="mb-4 flex items-center justify-between">
        <div>
          <h2 className="font-bold text-2xl text-gray-900">Edit Workflow</h2>
          <p className="text-gray-600 text-sm">
            Modify your workflow by adding or updating nodes
          </p>
        </div>
        <div className="flex space-x-2">
          <button
            className="rounded-md bg-gray-100 px-4 py-2 text-gray-700 hover:bg-gray-200"
            onClick={() => router.back()}
          >
            Cancel
          </button>
          <button
            className="rounded-md bg-indigo-600 px-4 py-2 text-white hover:bg-indigo-700 disabled:bg-indigo-400"
            disabled={saving}
            onClick={handleSave}
          >
            {saving ? "Saving..." : "Save Changes"}
          </button>
        </div>
      </div>

      <div
        className={`grid h-[calc(100%-80px)] gap-4 ${selectedNode ? "grid-cols-1 lg:grid-cols-[280px_1fr_300px]" : "grid-cols-1 lg:grid-cols-[280px_1fr]"}`}
      >
        {/* Sidebar */}
        <div className="overflow-y-auto rounded-lg bg-white p-4 shadow">
          <h3 className="mb-4 font-semibold">Workflow Info</h3>
          <div className="mb-6 space-y-3">
            <div>
              <label className="mb-1 block font-medium text-gray-700 text-sm">
                Name *
              </label>
              <input
                className="w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-indigo-500 focus:outline-none focus:ring-indigo-500"
                onChange={(e) => setWorkflowName(e.target.value)}
                placeholder="My Workflow"
                type="text"
                value={workflowName}
              />
            </div>
            <div>
              <label className="mb-1 block font-medium text-gray-700 text-sm">
                Description
              </label>
              <textarea
                className="w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-indigo-500 focus:outline-none focus:ring-indigo-500"
                onChange={(e) => setWorkflowDescription(e.target.value)}
                placeholder="What does this workflow do?"
                rows={3}
                value={workflowDescription}
              />
            </div>
          </div>

          <h3 className="mb-3 font-semibold">Add Nodes</h3>
          <div className="space-y-2">
            <button
              className="flex w-full items-center gap-2 rounded-md bg-blue-50 px-3 py-2 text-left text-blue-700 text-sm hover:bg-blue-100"
              onClick={() => addNode("http", "HTTP Request")}
            >
              <span>🌐</span> HTTP Request
            </button>
            <button
              className="flex w-full items-center gap-2 rounded-md bg-yellow-50 px-3 py-2 text-left text-sm text-yellow-700 hover:bg-yellow-100"
              onClick={() => addNode("delay", "Delay")}
            >
              <span>⏱️</span> Delay
            </button>
            <button
              className="flex w-full items-center gap-2 rounded-md bg-purple-50 px-3 py-2 text-left text-purple-700 text-sm hover:bg-purple-100"
              onClick={() => addNode("script", "Script")}
            >
              <span>📜</span> Script
            </button>
            <button
              className="flex w-full items-center gap-2 rounded-md bg-green-50 px-3 py-2 text-left text-green-700 text-sm hover:bg-green-100"
              onClick={() => addNode("condition", "Condition")}
            >
              <span>❓</span> Condition
            </button>
          </div>

          <div className="mt-6 text-gray-500 text-xs">
            <p className="mb-2">Tips:</p>
            <ul className="list-inside list-disc space-y-1">
              <li>Click a node to configure it</li>
              <li>Drag nodes to reposition</li>
              <li>Connect from bottom to top</li>
              <li>Delete nodes with Backspace</li>
            </ul>
          </div>
        </div>

        {/* Canvas */}
        <div className="rounded-lg bg-white shadow">
          <ReactFlow
            edges={edges}
            fitView
            nodes={nodes}
            nodeTypes={nodeTypes}
            onConnect={onConnect}
            onEdgesChange={onEdgesChange}
            onNodeClick={onNodeClick}
            onNodesChange={onNodesChange}
            onPaneClick={onPaneClick}
          >
            <Background />
            <Controls />
            <MiniMap />
          </ReactFlow>
        </div>

        {/* Config Panel */}
        {selectedNode && (
          <NodeConfigPanel
            config={(selectedNode.data.config as Record<string, unknown>) || {}}
            nodeId={selectedNode.id}
            nodeLabel={selectedNode.data.label as string}
            nodeType={selectedNode.data.type as string}
            onClose={() => setSelectedNodeId(null)}
            onConfigChange={handleConfigChange}
            onLabelChange={handleLabelChange}
          />
        )}
      </div>
    </div>
  );
}
