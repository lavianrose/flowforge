"use client";

import { useParams, useRouter } from "next/navigation";
import ReactFlow, { Background, Controls, MiniMap } from "reactflow";
import { useSnackbar } from "@/components/Snackbar";
import { useAuth } from "@/lib/auth";
import {
  useDeleteWorkflow,
  useTriggerWorkflow,
  useWorkflow,
} from "@/lib/hooks";
import "reactflow/dist/style.css";
import VersionHistory from "@/components/VersionHistory";
import { nodeTypes } from "@/lib/nodeTypes";

export default function WorkflowDetailPage() {
  const params = useParams();
  const router = useRouter();
  const { can } = useAuth();
  const { showSnackbar } = useSnackbar();
  const { data: workflow, isLoading, error } = useWorkflow(params.id as string);
  const deleteMutation = useDeleteWorkflow();
  const triggerMutation = useTriggerWorkflow();

  // Convert workflow definition to ReactFlow format
  const nodes = workflow
    ? workflow.definition.nodes.map((node) => ({
        id: node.id,
        type: "custom",
        position: node.position,
        data: {
          type: node.type,
          label: node.name,
          config: node.config,
        },
      }))
    : [];

  const edges = workflow
    ? workflow.definition.edges.map((edge) => ({
        id: edge.id,
        source: edge.source,
        target: edge.target,
      }))
    : [];

  const handleDelete = async () => {
    if (!workflow) {
      return;
    }
    if (!confirm("Are you sure you want to delete this workflow?")) {
      return;
    }

    try {
      await deleteMutation.mutateAsync(workflow.id);
      showSnackbar("Workflow deleted successfully", "success");
      router.push("/dashboard/workflows");
    } catch (err) {
      showSnackbar(
        err instanceof Error ? err.message : "Failed to delete workflow",
        "error"
      );
    }
  };

  const handleTrigger = async () => {
    if (!workflow) {
      return;
    }

    try {
      await triggerMutation.mutateAsync(workflow.id);
      showSnackbar("Workflow triggered successfully", "success");
    } catch (err) {
      showSnackbar(
        err instanceof Error ? err.message : "Failed to trigger workflow",
        "error"
      );
    }
  };

  if (isLoading) {
    return <div className="py-12 text-center">Loading workflow...</div>;
  }

  if (error || !workflow) {
    return (
      <div className="rounded border border-red-200 bg-red-50 px-4 py-3 text-red-700">
        {error?.message || "Workflow not found"}
      </div>
    );
  }

  return (
    <div>
      <div className="mb-6 flex items-center justify-between">
        <div>
          <h2 className="font-bold text-2xl text-gray-900">{workflow.name}</h2>
          <p className="mt-1 text-gray-600 text-sm">{workflow.description}</p>
        </div>
        <div className="flex space-x-2">
          <button
            className="rounded-md bg-gray-100 px-4 py-2 text-gray-700 hover:bg-gray-200"
            onClick={() => router.push("/dashboard/workflows")}
          >
            Back
          </button>
          {can("trigger") && (
            <button
              className="rounded-md bg-green-600 px-4 py-2 text-white hover:bg-green-700 disabled:opacity-50"
              disabled={triggerMutation.isPending}
              onClick={handleTrigger}
            >
              {triggerMutation.isPending ? "Running..." : "Run Workflow"}
            </button>
          )}
          {can("delete") && (
            <button
              className="rounded-md bg-red-600 px-4 py-2 text-white hover:bg-red-700 disabled:opacity-50"
              disabled={deleteMutation.isPending}
              onClick={handleDelete}
            >
              {deleteMutation.isPending ? "Deleting..." : "Delete"}
            </button>
          )}
        </div>
      </div>

      <div className="grid grid-cols-1 gap-6 lg:grid-cols-2">
        {/* Workflow Info */}
        <div className="rounded-lg bg-white p-6 shadow">
          <h3 className="mb-4 font-semibold text-lg">Workflow Details</h3>
          <dl className="space-y-3">
            <div>
              <dt className="font-medium text-gray-500 text-sm">Status</dt>
              <dd className="mt-1">
                <span
                  className={`rounded px-2 py-1 text-sm ${
                    workflow.active
                      ? "bg-green-100 text-green-800"
                      : "bg-gray-100 text-gray-800"
                  }`}
                >
                  {workflow.active ? "Active" : "Inactive"}
                </span>
              </dd>
            </div>
            <div>
              <dt className="font-medium text-gray-500 text-sm">Timeout</dt>
              <dd className="mt-1 text-gray-900 text-sm">
                {workflow.timeout_seconds} seconds
              </dd>
            </div>
            <div>
              <dt className="font-medium text-gray-500 text-sm">Created</dt>
              <dd className="mt-1 text-gray-900 text-sm">
                {new Date(workflow.created_at).toLocaleString()}
              </dd>
            </div>
            <div>
              <dt className="font-medium text-gray-500 text-sm">
                Last Updated
              </dt>
              <dd className="mt-1 text-gray-900 text-sm">
                {new Date(workflow.updated_at).toLocaleString()}
              </dd>
            </div>
          </dl>
        </div>

        {/* Version History */}
        <div className="rounded-lg bg-white p-6 shadow">
          <VersionHistory workflowId={workflow.id} />
        </div>

        {/* Nodes */}
        <div className="rounded-lg bg-white p-6 shadow lg:col-span-2">
          <h3 className="mb-4 font-semibold text-lg">
            Nodes ({workflow.definition.nodes.length})
          </h3>
          <div className="space-y-2">
            {workflow.definition.nodes.map((node) => (
              <div
                className="rounded-md border border-gray-200 bg-gray-50 p-3"
                key={node.id}
              >
                <div className="flex items-center justify-between">
                  <div>
                    <p className="font-medium text-gray-900">{node.name}</p>
                    <p className="text-gray-500 text-xs">Type: {node.type}</p>
                  </div>
                  <span className="rounded bg-blue-100 px-2 py-1 text-blue-800 text-xs">
                    {node.id}
                  </span>
                </div>
              </div>
            ))}
          </div>
        </div>

        {/* Edges */}
        <div className="rounded-lg bg-white p-6 shadow lg:col-span-2">
          <h3 className="mb-4 font-semibold text-lg">
            Connections ({workflow.definition.edges.length})
          </h3>
          <div className="grid grid-cols-1 gap-2 md:grid-cols-2 lg:grid-cols-3">
            {workflow.definition.edges.map((edge) => (
              <div
                className="rounded-md border border-gray-200 bg-gray-50 p-3 text-sm"
                key={edge.id}
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
      <div className="mt-6 rounded-lg bg-white p-6 shadow">
        <h3 className="mb-4 font-semibold text-lg">Workflow Visualization</h3>
        <div
          className="overflow-hidden rounded-lg border border-gray-200"
          style={{ height: "500px" }}
        >
          <ReactFlow
            edges={edges}
            elementsSelectable={false}
            fitView
            nodes={nodes}
            nodesConnectable={false}
            nodesDraggable={false}
            nodeTypes={nodeTypes}
            panOnScroll={true}
            zoomOnScroll={true}
          >
            <Background />
            <Controls />
            <MiniMap />
          </ReactFlow>
        </div>
        <p className="mt-2 text-gray-500 text-xs">
          💡 This is a read-only view. Use the "Edit Workflow" button to make
          changes.
        </p>
      </div>

      {/* Edit Workflow Button */}
      {can("edit") && (
        <div className="mt-6 flex justify-end">
          <button
            className="rounded-md bg-indigo-600 px-4 py-2 text-white hover:bg-indigo-700"
            onClick={() =>
              router.push(`/dashboard/workflows/${workflow.id}/edit`)
            }
          >
            Edit Workflow
          </button>
        </div>
      )}
    </div>
  );
}
