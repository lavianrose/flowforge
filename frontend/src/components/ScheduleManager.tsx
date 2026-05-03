"use client";

import { useState } from "react";
import { useSnackbar } from "@/components/Snackbar";
import { useAuth } from "@/lib/auth";
import {
  useCreateSchedule,
  useDeleteSchedule,
  useSchedules,
} from "@/lib/hooks";

const PRESET_CRONS = [
  { label: "Every 5 minutes", value: "*/5 * * * *" },
  { label: "Every 15 minutes", value: "*/15 * * * *" },
  { label: "Every 30 minutes", value: "*/30 * * * *" },
  { label: "Every hour", value: "0 * * * *" },
  { label: "Every 6 hours", value: "0 */6 * * *" },
  { label: "Every day at midnight", value: "0 0 * * *" },
  { label: "Every Monday at 9 AM", value: "0 9 * * 1" },
  { label: "First day of month", value: "0 0 1 * *" },
];

interface ScheduleManagerProps {
  workflowId: string;
}

export default function ScheduleManager({ workflowId }: ScheduleManagerProps) {
  const { can } = useAuth();
  const { showSnackbar } = useSnackbar();
  const { data: schedules, isLoading } = useSchedules();
  const createMutation = useCreateSchedule();
  const deleteMutation = useDeleteSchedule();
  const [cronExpression, setCronExpression] = useState("");
  const [showForm, setShowForm] = useState(false);

  const workflowSchedules = (schedules || []).filter(
    (s) => s.workflow_id === workflowId
  );

  const handleCreate = async () => {
    if (!cronExpression.trim()) {
      showSnackbar("Please enter a cron expression", "error");
      return;
    }

    try {
      await createMutation.mutateAsync({
        workflowId,
        cronExpression: cronExpression.trim(),
      });
      showSnackbar("Schedule created successfully", "success");
      setCronExpression("");
      setShowForm(false);
    } catch (err) {
      showSnackbar(
        err instanceof Error ? err.message : "Failed to create schedule",
        "error"
      );
    }
  };

  const handleDelete = async (id: string) => {
    if (!confirm("Are you sure you want to delete this schedule?")) {
      return;
    }

    try {
      await deleteMutation.mutateAsync(id);
      showSnackbar("Schedule deleted", "success");
    } catch (err) {
      showSnackbar(
        err instanceof Error ? err.message : "Failed to delete schedule",
        "error"
      );
    }
  };

  return (
    <div>
      <div className="mb-4 flex items-center justify-between">
        <h3 className="font-semibold text-lg">Schedules</h3>
        {can("edit") && !showForm && (
          <button
            className="rounded-md bg-indigo-600 px-3 py-1.5 text-white text-sm hover:bg-indigo-700"
            onClick={() => setShowForm(true)}
          >
            Add Schedule
          </button>
        )}
      </div>

      {showForm && (
        <div className="mb-4 rounded-md border border-gray-200 bg-gray-50 p-4">
          <label className="mb-2 block font-medium text-gray-700 text-sm">
            Cron Expression
          </label>
          <input
            className="mb-3 w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-indigo-500 focus:outline-none focus:ring-1 focus:ring-indigo-500"
            disabled={createMutation.isPending}
            onChange={(e) => setCronExpression(e.target.value)}
            placeholder="* * * * * (min hour day month weekday)"
            type="text"
            value={cronExpression}
          />
          <div className="mb-3 flex flex-wrap gap-1.5">
            {PRESET_CRONS.map((preset) => (
              <button
                className="rounded bg-white px-2 py-1 text-gray-600 text-xs shadow-sm hover:bg-indigo-50 hover:text-indigo-700"
                key={preset.value}
                onClick={() => setCronExpression(preset.value)}
                type="button"
              >
                {preset.label}
              </button>
            ))}
          </div>
          <div className="flex space-x-2">
            <button
              className="rounded-md bg-indigo-600 px-4 py-2 text-white text-sm hover:bg-indigo-700 disabled:opacity-50"
              disabled={createMutation.isPending}
              onClick={handleCreate}
            >
              {createMutation.isPending ? "Creating..." : "Create"}
            </button>
            <button
              className="rounded-md bg-white px-4 py-2 text-gray-700 text-sm hover:bg-gray-100"
              disabled={createMutation.isPending}
              onClick={() => {
                setShowForm(false);
                setCronExpression("");
              }}
            >
              Cancel
            </button>
          </div>
        </div>
      )}

      {isLoading ? (
        <p className="text-gray-500 text-sm">Loading schedules...</p>
      ) : workflowSchedules.length === 0 ? (
        <p className="text-gray-500 text-sm">
          No schedules configured. Add one to run this workflow automatically.
        </p>
      ) : (
        <div className="space-y-2">
          {workflowSchedules.map((schedule) => (
            <div
              className="flex items-center justify-between rounded-md border border-gray-200 bg-white p-3"
              key={schedule.id}
            >
              <div className="flex-1">
                <div className="flex items-center space-x-2">
                  <code className="rounded bg-gray-100 px-2 py-0.5 font-mono text-sm">
                    {schedule.cron_expression}
                  </code>
                  <span
                    className={`rounded px-1.5 py-0.5 text-xs ${
                      schedule.active
                        ? "bg-green-100 text-green-800"
                        : "bg-gray-100 text-gray-600"
                    }`}
                  >
                    {schedule.active ? "Active" : "Inactive"}
                  </span>
                </div>
                <div className="mt-1 space-x-4 text-gray-500 text-xs">
                  <span>
                    Next: {new Date(schedule.next_run_at).toLocaleString()}
                  </span>
                  {schedule.last_run_at && (
                    <span>
                      Last: {new Date(schedule.last_run_at).toLocaleString()}
                    </span>
                  )}
                </div>
              </div>
              {can("delete") && (
                <button
                  className="rounded px-2 py-1 text-red-600 text-sm hover:bg-red-50 disabled:opacity-50"
                  disabled={deleteMutation.isPending}
                  onClick={() => handleDelete(schedule.id)}
                >
                  Delete
                </button>
              )}
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
