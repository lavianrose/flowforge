"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { useSnackbar } from "@/components/Snackbar";
import { useAuth } from "@/lib/auth";
import {
  useSchedules,
  useWebhooks,
  useDeleteSchedule,
  useDeleteWebhook,
  useWorkflows,
} from "@/lib/hooks";

export default function SchedulesPage() {
  const router = useRouter();
  const { can } = useAuth();
  const { showSnackbar } = useSnackbar();
  const { data: schedules, isLoading: loadingSchedules } = useSchedules();
  const { data: webhooks, isLoading: loadingWebhooks } = useWebhooks();
  const { data: workflows } = useWorkflows();
  const deleteScheduleMutation = useDeleteSchedule();
  const deleteWebhookMutation = useDeleteWebhook();
  const [tab, setTab] = useState<"schedules" | "webhooks">("schedules");

  const isLoading = loadingSchedules || loadingWebhooks;

  const workflowMap = new Map(
    (workflows || []).map((w) => [w.id, w.name])
  );

  const handleDeleteSchedule = async (id: string) => {
    if (!confirm("Are you sure you want to delete this schedule?")) return;
    try {
      await deleteScheduleMutation.mutateAsync(id);
      showSnackbar("Schedule deleted", "success");
    } catch (err) {
      showSnackbar(
        err instanceof Error ? err.message : "Failed to delete schedule",
        "error"
      );
    }
  };

  const handleDeleteWebhook = async (id: string) => {
    if (!confirm("Are you sure you want to delete this webhook?")) return;
    try {
      await deleteWebhookMutation.mutateAsync(id);
      showSnackbar("Webhook deleted", "success");
    } catch (err) {
      showSnackbar(
        err instanceof Error ? err.message : "Failed to delete webhook",
        "error"
      );
    }
  };

  if (isLoading) {
    return (
      <div className="py-12 text-center">Loading schedules & webhooks...</div>
    );
  }

  return (
    <div>
      <div className="mb-6 flex items-center justify-between">
        <h2 className="font-bold text-2xl text-gray-900">
          Schedules & Webhooks
        </h2>
      </div>

      {/* Tabs */}
      <div className="mb-6 flex space-x-2">
        <button
          className={`rounded-md px-4 py-2 text-sm font-medium ${
            tab === "schedules"
              ? "bg-indigo-600 text-white"
              : "bg-gray-100 text-gray-700 hover:bg-gray-200"
          }`}
          onClick={() => setTab("schedules")}
        >
          Schedules ({(schedules || []).length})
        </button>
        <button
          className={`rounded-md px-4 py-2 text-sm font-medium ${
            tab === "webhooks"
              ? "bg-indigo-600 text-white"
              : "bg-gray-100 text-gray-700 hover:bg-gray-200"
          }`}
          onClick={() => setTab("webhooks")}
        >
          Webhooks ({(webhooks || []).length})
        </button>
      </div>

      {/* Schedules Tab */}
      {tab === "schedules" && (
        <>
          {(schedules || []).length === 0 ? (
            <div className="rounded-lg bg-white py-12 text-center shadow">
              <p className="text-gray-500">No schedules found</p>
              <p className="mt-1 text-gray-400 text-sm">
                Go to a workflow detail page to add a schedule.
              </p>
            </div>
          ) : (
            <div className="overflow-hidden rounded-lg bg-white shadow">
              <table className="min-w-full divide-y divide-gray-200">
                <thead className="bg-gray-50">
                  <tr>
                    <th className="px-6 py-3 text-left font-medium text-gray-500 text-xs uppercase tracking-wider">
                      Workflow
                    </th>
                    <th className="px-6 py-3 text-left font-medium text-gray-500 text-xs uppercase tracking-wider">
                      Cron
                    </th>
                    <th className="px-6 py-3 text-left font-medium text-gray-500 text-xs uppercase tracking-wider">
                      Status
                    </th>
                    <th className="px-6 py-3 text-left font-medium text-gray-500 text-xs uppercase tracking-wider">
                      Next Run
                    </th>
                    <th className="px-6 py-3 text-left font-medium text-gray-500 text-xs uppercase tracking-wider">
                      Last Run
                    </th>
                    <th className="px-6 py-3 text-left font-medium text-gray-500 text-xs uppercase tracking-wider">
                      Actions
                    </th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-gray-200 bg-white">
                  {(schedules || []).map((schedule) => (
                    <tr className="hover:bg-gray-50" key={schedule.id}>
                      <td className="whitespace-nowrap px-6 py-4">
                        <button
                          className="font-medium text-indigo-600 text-sm hover:text-indigo-800"
                          onClick={() =>
                            router.push(
                              `/dashboard/workflows/${schedule.workflow_id}`
                            )
                          }
                        >
                          {workflowMap.get(schedule.workflow_id) ||
                            schedule.workflow_id.slice(0, 8) + "..."}
                        </button>
                      </td>
                      <td className="px-6 py-4">
                        <code className="rounded bg-gray-100 px-2 py-0.5 font-mono text-xs">
                          {schedule.cron_expression}
                        </code>
                      </td>
                      <td className="whitespace-nowrap px-6 py-4">
                        <span
                          className={`rounded px-2 py-1 font-medium text-xs ${
                            schedule.active
                              ? "bg-green-100 text-green-800"
                              : "bg-gray-100 text-gray-600"
                          }`}
                        >
                          {schedule.active ? "Active" : "Inactive"}
                        </span>
                      </td>
                      <td className="whitespace-nowrap px-6 py-4 text-gray-600 text-sm">
                        {new Date(schedule.next_run_at).toLocaleString()}
                      </td>
                      <td className="whitespace-nowrap px-6 py-4 text-gray-600 text-sm">
                        {schedule.last_run_at
                          ? new Date(schedule.last_run_at).toLocaleString()
                          : "-"}
                      </td>
                      <td className="whitespace-nowrap px-6 py-4">
                        {can("delete_schedule") && (
                          <button
                            className="rounded px-2 py-1 text-red-600 text-xs hover:bg-red-50 disabled:opacity-50"
                            disabled={deleteScheduleMutation.isPending}
                            onClick={() => handleDeleteSchedule(schedule.id)}
                          >
                            Delete
                          </button>
                        )}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </>
      )}

      {/* Webhooks Tab */}
      {tab === "webhooks" && (
        <>
          {(webhooks || []).length === 0 ? (
            <div className="rounded-lg bg-white py-12 text-center shadow">
              <p className="text-gray-500">No webhooks found</p>
              <p className="mt-1 text-gray-400 text-sm">
                Go to a workflow detail page to add a webhook.
              </p>
            </div>
          ) : (
            <div className="overflow-hidden rounded-lg bg-white shadow">
              <table className="min-w-full divide-y divide-gray-200">
                <thead className="bg-gray-50">
                  <tr>
                    <th className="px-6 py-3 text-left font-medium text-gray-500 text-xs uppercase tracking-wider">
                      Workflow
                    </th>
                    <th className="px-6 py-3 text-left font-medium text-gray-500 text-xs uppercase tracking-wider">
                      Status
                    </th>
                    <th className="px-6 py-3 text-left font-medium text-gray-500 text-xs uppercase tracking-wider">
                      Path
                    </th>
                    <th className="px-6 py-3 text-left font-medium text-gray-500 text-xs uppercase tracking-wider">
                      Created
                    </th>
                    <th className="px-6 py-3 text-left font-medium text-gray-500 text-xs uppercase tracking-wider">
                      Actions
                    </th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-gray-200 bg-white">
                  {(webhooks || []).map((webhook) => (
                    <tr className="hover:bg-gray-50" key={webhook.id}>
                      <td className="whitespace-nowrap px-6 py-4">
                        <button
                          className="font-medium text-indigo-600 text-sm hover:text-indigo-800"
                          onClick={() =>
                            router.push(
                              `/dashboard/workflows/${webhook.workflow_id}`
                            )
                          }
                        >
                          {workflowMap.get(webhook.workflow_id) ||
                            webhook.workflow_id.slice(0, 8) + "..."}
                        </button>
                      </td>
                      <td className="whitespace-nowrap px-6 py-4">
                        <span
                          className={`rounded px-2 py-1 font-medium text-xs ${
                            webhook.active
                              ? "bg-green-100 text-green-800"
                              : "bg-gray-100 text-gray-600"
                          }`}
                        >
                          {webhook.active ? "Active" : "Inactive"}
                        </span>
                      </td>
                      <td className="px-6 py-4">
                        <code className="rounded bg-gray-100 px-2 py-0.5 font-mono text-xs">
                          {webhook.path}
                        </code>
                      </td>
                      <td className="whitespace-nowrap px-6 py-4 text-gray-600 text-sm">
                        {new Date(webhook.created_at).toLocaleString()}
                      </td>
                      <td className="whitespace-nowrap px-6 py-4">
                        {can("delete_webhook") && (
                          <button
                            className="rounded px-2 py-1 text-red-600 text-xs hover:bg-red-50 disabled:opacity-50"
                            disabled={deleteWebhookMutation.isPending}
                            onClick={() => handleDeleteWebhook(webhook.id)}
                          >
                            Delete
                          </button>
                        )}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </>
      )}
    </div>
  );
}
