"use client";

import { useState, useEffect, useRef } from "react";
import { useParams, useRouter } from "next/navigation";
import { useRun, useTriggerWorkflow } from "@/lib/hooks";
import { connectSSE } from "@/lib/sse";
import { WorkflowRun } from "@/lib/api";

interface Step {
  id: string;
  run_id: string;
  step_id: string;
  status: string;
  input: Record<string, unknown>;
  output: Record<string, unknown> | null;
  error: string;
  retry_count: number;
  started_at: string | null;
  completed_at: string | null;
  created_at: string;
}

export default function RunDetailPage() {
  const params = useParams();
  const router = useRouter();
  const {
    data: runData,
    isLoading,
    error,
    refetch,
  } = useRun(params.id as string);
  const retryMutation = useTriggerWorkflow();
  const [retryError, setRetryError] = useState<string | null>(null);
  const [run, setRun] = useState(runData?.run || null);
  const [steps, setSteps] = useState<Step[]>(runData?.steps || []);
  const [liveMode, setLiveMode] = useState(false);
  const [now, setNow] = useState<number>(() => Date.now());
  const cleanupRef = useRef<(() => void) | null>(null);

  // Update local state when query data changes
  /* eslint-disable react-hooks/set-state-in-effect */
  useEffect(() => {
    if (runData) {
      setRun(runData.run);
      setSteps(runData.steps || []);

      // Auto-enable live mode if run is still running
      if (
        runData.run.status === "running" ||
        runData.run.status === "pending"
      ) {
        setLiveMode(true);
      }
    }
  }, [runData]);
  /* eslint-enable react-hooks/set-state-in-effect */

  useEffect(() => {
    if (!liveMode || !run) return;

    // Connect to SSE stream
    const apiUrl =
      process.env.NEXT_PUBLIC_API_URL || "http://localhost:3000/api/v1";
    const sseUrl = `${apiUrl}/runs/${params.id}/stream`;

    cleanupRef.current = connectSSE(
      sseUrl,
      (message) => {
        if (message.event === "run_state") {
          const updatedRun = message.data as WorkflowRun;
          setRun(updatedRun);

          // Stop live mode when run completes
          if (
            updatedRun.status === "success" ||
            updatedRun.status === "failed" ||
            updatedRun.status === "cancelled"
          ) {
            setLiveMode(false);
          }
        } else if (message.event === "steps_state") {
          const updatedSteps = message.data as Step[];
          setSteps(updatedSteps);
        }
      },
      (err) => {
        console.error("SSE error:", err);
        setLiveMode(false);
      },
    );

    return () => {
      if (cleanupRef.current) {
        cleanupRef.current();
      }
    };
  }, [liveMode, run, params.id]);

  useEffect(() => {
    if (run?.status !== "running" && run?.status !== "pending") {
      return;
    }

    const interval = window.setInterval(() => {
      setNow(Date.now());
    }, 1000);

    return () => window.clearInterval(interval);
  }, [run?.status]);

  const getStatusColor = (status: string) => {
    switch (status) {
      case "success":
        return "bg-green-100 text-green-800 border-green-300";
      case "failed":
        return "bg-red-100 text-red-800 border-red-300";
      case "running":
        return "bg-blue-100 text-blue-800 border-blue-300";
      case "pending":
        return "bg-yellow-100 text-yellow-800 border-yellow-300";
      default:
        return "bg-gray-100 text-gray-800 border-gray-300";
    }
  };

  const getStepIcon = (status: string) => {
    switch (status) {
      case "success":
        return "✅";
      case "failed":
        return "❌";
      case "running":
        return "⏳";
      case "pending":
        return "⏸️";
      default:
        return "❓";
    }
  };

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="text-xl">Loading run details...</div>
      </div>
    );
  }

  if (error || !run) {
    return (
      <div className="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded">
        {error?.message || "Run not found"}
      </div>
    );
  }

  const started = run.started_at ? new Date(run.started_at) : null;
  const completed = run.completed_at ? new Date(run.completed_at) : null;
  const duration =
    started && completed
      ? Math.round((completed.getTime() - started.getTime()) / 1000)
      : started
        ? Math.round((now - started.getTime()) / 1000)
        : null;

  return (
    <div>
      <div className="flex justify-between items-center mb-6">
        <div>
          <div className="flex items-center gap-3">
            <h2 className="text-2xl font-bold text-gray-900">Run Details</h2>
            {liveMode && (
              <span className="inline-flex items-center gap-1 px-2 py-1 bg-green-100 text-green-800 text-xs font-medium rounded-full">
                <span className="animate-pulse">🔴</span> LIVE
              </span>
            )}
          </div>
          <p className="text-sm text-gray-600 mt-1">
            Workflow ID:{" "}
            <span className="font-mono">{run.workflow_id.slice(0, 8)}...</span>
          </p>
        </div>
        <div className="flex items-center space-x-2">
          {(run.status === "failed" || run.status === "cancelled") && (
            <button
              onClick={() => {
                setRetryError(null);
                retryMutation.mutate(run.workflow_id, {
                  onSuccess: (newRun) => {
                    router.push(`/dashboard/runs/${newRun.id}`);
                  },
                  onError: (err) => {
                    setRetryError(
                      err instanceof Error ? err.message : "Retry failed",
                    );
                  },
                });
              }}
              className="px-4 py-2 bg-yellow-500 text-white rounded-md hover:bg-yellow-600"
              disabled={retryMutation.isPending}
            >
              {retryMutation.isPending ? "Retrying..." : "Retry Workflow"}
            </button>
          )}
          <button
            onClick={() => {
              setLiveMode(!liveMode);
              if (!liveMode) {
                refetch();
              }
            }}
            className={`px-4 py-2 rounded-md ${
              liveMode
                ? "bg-green-600 text-white hover:bg-green-700"
                : "bg-gray-100 text-gray-700 hover:bg-gray-200"
            }`}
            disabled={run.status !== "running" && run.status !== "pending"}
          >
            {liveMode ? "Live" : "Live Updates"}
          </button>
          <button
            onClick={() => refetch()}
            className="px-4 py-2 bg-indigo-600 text-white rounded-md hover:bg-indigo-700"
          >
            Refresh
          </button>
          <button
            onClick={() => router.back()}
            className="px-4 py-2 text-gray-700 bg-gray-100 rounded-md hover:bg-gray-200"
          >
            Back
          </button>
        </div>
      </div>

      {retryError && (
        <div className="mb-4 p-4 bg-red-50 border border-red-200 text-red-700 rounded">
          <p className="font-medium">Retry failed</p>
          <p className="text-sm mt-1">{retryError}</p>
        </div>
      )}

      {/* Run Info */}
      <div className="bg-white rounded-lg shadow p-6 mb-6">
        <h3 className="text-lg font-semibold mb-4">Run Information</h3>
        <dl className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
          <div>
            <dt className="text-sm font-medium text-gray-500">Status</dt>
            <dd className="mt-1">
              <span
                className={`px-2 py-1 text-sm font-medium rounded border ${getStatusColor(
                  run.status,
                )}`}
              >
                {getStepIcon(run.status)} {run.status}
              </span>
            </dd>
          </div>
          <div>
            <dt className="text-sm font-medium text-gray-500">Triggered By</dt>
            <dd className="mt-1 text-sm text-gray-900">{run.triggered_by}</dd>
          </div>
          <div>
            <dt className="text-sm font-medium text-gray-500">Started</dt>
            <dd className="mt-1 text-sm text-gray-900">
              {started?.toLocaleString() || "-"}
            </dd>
          </div>
          <div>
            <dt className="text-sm font-medium text-gray-500">Duration</dt>
            <dd className="mt-1 text-sm text-gray-900">
              {duration ? `${duration}s` : "-"}
            </dd>
          </div>
        </dl>

        {run.error && (
          <div className="mt-4 p-3 bg-red-50 border border-red-200 rounded-md">
            <p className="text-sm text-red-800 font-medium">Error</p>
            <p className="text-sm text-red-700 mt-1">{run.error}</p>
          </div>
        )}
      </div>

      {/* Steps Timeline */}
      <div className="bg-white rounded-lg shadow p-6">
        <h3 className="text-lg font-semibold mb-4">
          Execution Steps ({steps.length})
        </h3>

        {steps.length === 0 ? (
          <p className="text-sm text-gray-500">No steps recorded yet</p>
        ) : (
          <div className="space-y-4">
            {steps.map((step, index) => (
              <div
                key={step.id}
                className="flex gap-4 p-4 bg-gray-50 rounded-lg border border-gray-200"
              >
                <div className="flex-shrink-0">
                  <div
                    className={`w-8 h-8 rounded-full flex items-center justify-center text-sm ${getStatusColor(
                      step.status,
                    )} border`}
                  >
                    {index + 1}
                  </div>
                </div>
                <div className="flex-1 min-w-0">
                  <div className="flex items-center justify-between mb-2">
                    <h4 className="text-sm font-medium text-gray-900">
                      {step.step_id}
                    </h4>
                    <div className="flex items-center gap-2">
                      <span
                        className={`px-2 py-1 text-xs font-medium rounded border ${getStatusColor(
                          step.status,
                        )}`}
                      >
                        {getStepIcon(step.status)} {step.status}
                      </span>
                      {step.retry_count > 0 && (
                        <span className="text-xs text-gray-500">
                          Retry #{step.retry_count}
                        </span>
                      )}
                    </div>
                  </div>

                  {step.error && (
                    <div className="mt-2 p-2 bg-red-50 border border-red-200 rounded text-sm text-red-700">
                      {step.error}
                    </div>
                  )}

                  {step.output && (
                    <details className="mt-2">
                      <summary className="text-sm text-gray-600 cursor-pointer hover:text-gray-800">
                        Output
                      </summary>
                      <pre className="mt-2 text-xs bg-gray-100 p-2 rounded overflow-x-auto">
                        {JSON.stringify(step.output, null, 2)}
                      </pre>
                    </details>
                  )}

                  <div className="mt-2 text-xs text-gray-500">
                    {step.started_at && (
                      <span>
                        Started: {new Date(step.started_at).toLocaleString()}
                      </span>
                    )}
                    {step.completed_at && (
                      <span className="ml-4">
                        Completed:{" "}
                        {new Date(step.completed_at).toLocaleString()}
                      </span>
                    )}
                  </div>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
