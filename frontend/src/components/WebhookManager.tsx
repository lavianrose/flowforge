"use client";

import { useState } from "react";
import { useSnackbar } from "@/components/Snackbar";
import { useAuth } from "@/lib/auth";
import {
  useCreateWebhook,
  useDeleteWebhook,
  useWebhooks,
} from "@/lib/hooks";

interface WebhookManagerProps {
  workflowId: string;
}

export default function WebhookManager({ workflowId }: WebhookManagerProps) {
  const { can } = useAuth();
  const { showSnackbar } = useSnackbar();
  const { data: webhooks, isLoading } = useWebhooks();
  const createMutation = useCreateWebhook();
  const deleteMutation = useDeleteWebhook();
  const [revealedSecret, setRevealedSecret] = useState<string | null>(null);

  const API_BASE =
    process.env.NEXT_PUBLIC_API_URL?.replace("/api/v1", "") ||
    "http://localhost:3000";

  const workflowWebhooks = (webhooks || []).filter(
    (w) => w.workflow_id === workflowId
  );

  const handleCreate = async () => {
    try {
      await createMutation.mutateAsync(workflowId);
      showSnackbar("Webhook created successfully", "success");
    } catch (err) {
      showSnackbar(
        err instanceof Error ? err.message : "Failed to create webhook",
        "error"
      );
    }
  };

  const handleDelete = async (id: string) => {
    if (!confirm("Are you sure you want to delete this webhook?")) {
      return;
    }

    try {
      await deleteMutation.mutateAsync(id);
      showSnackbar("Webhook deleted", "success");
    } catch (err) {
      showSnackbar(
        err instanceof Error ? err.message : "Failed to delete webhook",
        "error"
      );
    }
  };

  const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text);
    showSnackbar("Copied to clipboard", "success");
  };

  return (
    <div>
      <div className="mb-4 flex items-center justify-between">
        <h3 className="font-semibold text-lg">Webhooks</h3>
        {can("edit") && (
          <button
            className="rounded-md bg-indigo-600 px-3 py-1.5 text-white text-sm hover:bg-indigo-700 disabled:opacity-50"
            disabled={createMutation.isPending}
            onClick={handleCreate}
          >
            {createMutation.isPending ? "Creating..." : "Add Webhook"}
          </button>
        )}
      </div>

      {isLoading ? (
        <p className="text-gray-500 text-sm">Loading webhooks...</p>
      ) : workflowWebhooks.length === 0 ? (
        <p className="text-gray-500 text-sm">
          No webhooks configured. Add one to trigger this workflow via HTTP.
        </p>
      ) : (
        <div className="space-y-2">
          {workflowWebhooks.map((webhook) => (
            <div
              className="rounded-md border border-gray-200 bg-white p-3"
              key={webhook.id}
            >
              <div className="flex items-start justify-between">
                <div className="flex-1">
                  <div className="flex items-center space-x-2">
                    <span
                      className={`rounded px-1.5 py-0.5 text-xs ${
                        webhook.active
                          ? "bg-green-100 text-green-800"
                          : "bg-gray-100 text-gray-600"
                      }`}
                    >
                      {webhook.active ? "Active" : "Inactive"}
                    </span>
                    <span className="text-gray-400 text-xs">
                      Created {new Date(webhook.created_at).toLocaleString()}
                    </span>
                  </div>
                  <div className="mt-2 space-y-1">
                    <div className="flex items-center space-x-2">
                      <span className="text-gray-500 text-xs">URL:</span>
                      <code className="flex-1 truncate rounded bg-gray-100 px-2 py-0.5 font-mono text-xs">
                        {API_BASE}{webhook.path}
                      </code>
                      <button
                        className="rounded px-1.5 py-0.5 text-indigo-600 text-xs hover:bg-indigo-50"
                        onClick={() =>
                          copyToClipboard(`${API_BASE}${webhook.path}`)
                        }
                        type="button"
                      >
                        Copy
                      </button>
                    </div>
                    <div className="flex items-center space-x-2">
                      <span className="text-gray-500 text-xs">Secret:</span>
                      <code className="flex-1 truncate rounded bg-gray-100 px-2 py-0.5 font-mono text-xs">
                        {revealedSecret === webhook.id
                          ? webhook.secret
                          : "••••••••••••••••"}
                      </code>
                      <button
                        className="rounded px-1.5 py-0.5 text-gray-500 text-xs hover:bg-gray-100"
                        onClick={() =>
                          setRevealedSecret(
                            revealedSecret === webhook.id ? null : webhook.id
                          )
                        }
                        type="button"
                      >
                        {revealedSecret === webhook.id ? "Hide" : "Show"}
                      </button>
                      {revealedSecret === webhook.id && (
                        <button
                          className="rounded px-1.5 py-0.5 text-indigo-600 text-xs hover:bg-indigo-50"
                          onClick={() =>
                            copyToClipboard(webhook.secret)
                          }
                          type="button"
                        >
                          Copy
                        </button>
                      )}
                    </div>
                  </div>
                  <p className="mt-1 text-gray-400 text-xs">
                    Send POST with header <code>X-Webhook-Secret</code> to
                    trigger
                  </p>
                </div>
                {can("delete_webhook") && (
                  <button
                    className="ml-2 rounded px-2 py-1 text-red-600 text-sm hover:bg-red-50 disabled:opacity-50"
                    disabled={deleteMutation.isPending}
                    onClick={() => handleDelete(webhook.id)}
                  >
                    Delete
                  </button>
                )}
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
