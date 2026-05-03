"use client";

import { useRouter } from "next/navigation";
import { useState } from "react";
import { useRuns } from "@/lib/hooks";

export default function RunsPage() {
  const router = useRouter();
  const { data: runs, isLoading, error, refetch } = useRuns();
  const [filter, setFilter] = useState<
    "all" | "running" | "success" | "failed"
  >("all");

  const getStatusColor = (status: string) => {
    switch (status) {
      case "success":
        return "bg-green-100 text-green-800";
      case "failed":
        return "bg-red-100 text-red-800";
      case "running":
        return "bg-blue-100 text-blue-800";
      case "pending":
        return "bg-yellow-100 text-yellow-800";
      default:
        return "bg-gray-100 text-gray-800";
    }
  };

  const filteredRuns = (runs || []).filter((run) => {
    if (filter === "all") {
      return true;
    }
    return run.status === filter;
  });

  if (isLoading) {
    return <div className="py-12 text-center">Loading runs...</div>;
  }

  if (error) {
    return (
      <div className="rounded border border-red-200 bg-red-50 px-4 py-3 text-red-700">
        <p className="font-semibold">Error loading runs</p>
        <p className="text-sm">{error.message}</p>
        <button
          className="mt-2 rounded bg-red-600 px-3 py-1 text-sm text-white hover:bg-red-700"
          onClick={() => refetch()}
        >
          Retry
        </button>
      </div>
    );
  }

  return (
    <div>
      <div className="mb-6 flex items-center justify-between">
        <h2 className="font-bold text-2xl text-gray-900">Workflow Runs</h2>
        <button
          className="rounded-md bg-indigo-600 px-4 py-2 text-white hover:bg-indigo-700"
          onClick={() => refetch()}
        >
          Refresh
        </button>
      </div>

      {/* Filter */}
      <div className="mb-6 flex space-x-2">
        <button
          className={`rounded-md px-3 py-1 text-sm ${
            filter === "all"
              ? "bg-indigo-600 text-white"
              : "bg-gray-100 text-gray-700 hover:bg-gray-200"
          }`}
          onClick={() => setFilter("all")}
        >
          All
        </button>
        <button
          className={`rounded-md px-3 py-1 text-sm ${
            filter === "running"
              ? "bg-indigo-600 text-white"
              : "bg-gray-100 text-gray-700 hover:bg-gray-200"
          }`}
          onClick={() => setFilter("running")}
        >
          Running
        </button>
        <button
          className={`rounded-md px-3 py-1 text-sm ${
            filter === "success"
              ? "bg-indigo-600 text-white"
              : "bg-gray-100 text-gray-700 hover:bg-gray-200"
          }`}
          onClick={() => setFilter("success")}
        >
          Success
        </button>
        <button
          className={`rounded-md px-3 py-1 text-sm ${
            filter === "failed"
              ? "bg-indigo-600 text-white"
              : "bg-gray-100 text-gray-700 hover:bg-gray-200"
          }`}
          onClick={() => setFilter("failed")}
        >
          Failed
        </button>
      </div>

      {filteredRuns.length === 0 ? (
        <div className="rounded-lg bg-white py-12 text-center shadow">
          <p className="text-gray-500">No runs found</p>
        </div>
      ) : (
        <div className="overflow-hidden rounded-lg bg-white shadow">
          <table className="min-w-full divide-y divide-gray-200">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-6 py-3 text-left font-medium text-gray-500 text-xs uppercase tracking-wider">
                  Status
                </th>
                <th className="px-6 py-3 text-left font-medium text-gray-500 text-xs uppercase tracking-wider">
                  Workflow ID
                </th>
                <th className="px-6 py-3 text-left font-medium text-gray-500 text-xs uppercase tracking-wider">
                  Triggered By
                </th>
                <th className="px-6 py-3 text-left font-medium text-gray-500 text-xs uppercase tracking-wider">
                  Started
                </th>
                <th className="px-6 py-3 text-left font-medium text-gray-500 text-xs uppercase tracking-wider">
                  Completed
                </th>
                <th className="px-6 py-3 text-left font-medium text-gray-500 text-xs uppercase tracking-wider">
                  Duration
                </th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-200 bg-white">
              {filteredRuns.map((run) => {
                const started = run.started_at
                  ? new Date(run.started_at)
                  : null;
                const completed = run.completed_at
                  ? new Date(run.completed_at)
                  : null;
                const duration =
                  started && completed
                    ? Math.round(
                        (completed.getTime() - started.getTime()) / 1000
                      )
                    : null;

                return (
                  <tr
                    className="cursor-pointer hover:bg-gray-50"
                    key={run.id}
                    onClick={() => router.push(`/dashboard/runs/${run.id}`)}
                  >
                    <td className="whitespace-nowrap px-6 py-4">
                      <span
                        className={`rounded px-2 py-1 font-medium text-xs ${getStatusColor(
                          run.status
                        )}`}
                      >
                        {run.status}
                      </span>
                    </td>
                    <td className="whitespace-nowrap px-6 py-4 font-mono text-gray-900 text-sm">
                      {run.workflow_id.slice(0, 8)}...
                    </td>
                    <td className="whitespace-nowrap px-6 py-4 text-gray-600 text-sm">
                      {run.triggered_by}
                    </td>
                    <td className="whitespace-nowrap px-6 py-4 text-gray-600 text-sm">
                      {started?.toLocaleString() || "-"}
                    </td>
                    <td className="whitespace-nowrap px-6 py-4 text-gray-600 text-sm">
                      {completed?.toLocaleString() || "-"}
                    </td>
                    <td className="whitespace-nowrap px-6 py-4 text-gray-600 text-sm">
                      {duration === null ? "-" : `${duration}s`}
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}
