"use client";

import { useParams, useRouter } from "next/navigation";
import { useEffect, useRef, useState } from "react";
import type { WorkflowRun } from "@/lib/api";
import { useAuth } from "@/lib/auth";
import { useRun, useTriggerWorkflow } from "@/lib/hooks";
import { connectSSE } from "@/lib/sse";

interface Step {
  completed_at: string | null;
  created_at: string;
  error: string;
  id: string;
  input: Record<string, unknown>;
  output: Record<string, unknown> | null;
  retry_count: number;
  run_id: string;
  started_at: string | null;
  status: string;
  step_id: string;
}

export default function RunDetailPage() {
  const params = useParams();
  const router = useRouter();
  const { token } = useAuth();
  const { data: runData, isLoading, error } = useRun(params.id as string);
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

      // Always enable live mode — updates arrive in real-time automatically
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
    if (!run) {
      return;
    }

    // Only connect SSE for active runs
    if (run.status !== "running" && run.status !== "pending") {
      return;
    }

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
      token
    );

    return () => {
      if (cleanupRef.current) {
        cleanupRef.current();
      }
    };
  }, [run, params.id, token]);

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
      <div className="flex h-64 items-center justify-center">
        <div className="text-xl">Loading run details...</div>
      </div>
    );
  }

  if (error || !run) {
    return (
      <div className="rounded border border-red-200 bg-red-50 px-4 py-3 text-red-700">
        {error?.message || "Run not found"}
      </div>
    );
  }

  const started = run.started_at ? new Date(run.started_at) : null;
  const completed = run.completed_at ? new Date(run.completed_at) : null;
  let duration: number | null = null;
  if (started && completed) {
    duration = Math.round((completed.getTime() - started.getTime()) / 1000);
  } else if (started) {
    duration = Math.round((now - started.getTime()) / 1000);
  }

  return (
    <div>
      <div className="mb-6 flex items-center justify-between">
        <div>
          <div className="flex items-center gap-3">
            <h2 className="font-bold text-2xl text-gray-900">Run Details</h2>
            {liveMode && (
              <span className="inline-flex items-center gap-1 rounded-full bg-green-100 px-2 py-1 font-medium text-green-800 text-xs">
                <span className="animate-pulse">🔴</span> LIVE
              </span>
            )}
          </div>
          <p className="mt-1 text-gray-600 text-sm">
            Workflow ID:{" "}
            <span className="font-mono">{run.workflow_id.slice(0, 8)}...</span>
          </p>
        </div>
        <div className="flex items-center space-x-2">
          {(run.status === "failed" || run.status === "cancelled") && (
            <button
              className="rounded-md bg-yellow-500 px-4 py-2 text-white hover:bg-yellow-600"
              disabled={retryMutation.isPending}
              onClick={() => {
                setRetryError(null);
                retryMutation.mutate(run.workflow_id, {
                  onSuccess: (newRun) => {
                    router.push(`/dashboard/runs/${newRun.id}`);
                  },
                  onError: (err) => {
                    setRetryError(
                      err instanceof Error ? err.message : "Retry failed"
                    );
                  },
                });
              }}
              type="button"
            >
              {retryMutation.isPending ? "Retrying..." : "Retry Workflow"}
            </button>
          )}
          <button
            className="rounded-md bg-gray-100 px-4 py-2 text-gray-700 hover:bg-gray-200"
            onClick={() => router.back()}
            type="button"
          >
            Back
          </button>
        </div>
      </div>

      {retryError && (
        <div className="mb-4 rounded border border-red-200 bg-red-50 p-4 text-red-700">
          <p className="font-medium">Retry failed</p>
          <p className="mt-1 text-sm">{retryError}</p>
        </div>
      )}

      {/* Run Info */}
      <div className="mb-6 rounded-lg bg-white p-6 shadow">
        <h3 className="mb-4 font-semibold text-lg">Run Information</h3>
        <dl className="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-4">
          <div>
            <dt className="font-medium text-gray-500 text-sm">Status</dt>
            <dd className="mt-1">
              <span
                className={`rounded border px-2 py-1 font-medium text-sm ${getStatusColor(
                  run.status
                )}`}
              >
                {getStepIcon(run.status)} {run.status}
              </span>
            </dd>
          </div>
          <div>
            <dt className="font-medium text-gray-500 text-sm">Triggered By</dt>
            <dd className="mt-1 text-gray-900 text-sm">{run.triggered_by}</dd>
          </div>
          <div>
            <dt className="font-medium text-gray-500 text-sm">Started</dt>
            <dd className="mt-1 text-gray-900 text-sm">
              {started?.toLocaleString() || "-"}
            </dd>
          </div>
          <div>
            <dt className="font-medium text-gray-500 text-sm">Duration</dt>
            <dd className="mt-1 text-gray-900 text-sm">
              {duration ? `${duration}s` : "-"}
            </dd>
          </div>
        </dl>

        {run.error && (
          <div className="mt-4 rounded-md border border-red-200 bg-red-50 p-3">
            <p className="font-medium text-red-800 text-sm">Error</p>
            <p className="mt-1 text-red-700 text-sm">{run.error}</p>
          </div>
        )}
      </div>

      {/* Steps Timeline */}
      <div className="rounded-lg bg-white p-6 shadow">
        <h3 className="mb-4 font-semibold text-lg">
          Execution Steps ({steps.length})
        </h3>

        {steps.length === 0 ? (
          <p className="text-gray-500 text-sm">No steps recorded yet</p>
        ) : (
          <div className="space-y-4">
            {steps.map((step, index) => (
              <div
                className="flex gap-4 rounded-lg border border-gray-200 bg-gray-50 p-4"
                key={step.id}
              >
                <div className="flex-shrink-0">
                  <div
                    className={`flex h-8 w-8 items-center justify-center rounded-full text-sm ${getStatusColor(
                      step.status
                    )} border`}
                  >
                    {index + 1}
                  </div>
                </div>
                <div className="min-w-0 flex-1">
                  <div className="mb-2 flex items-center justify-between">
                    <h4 className="font-medium text-gray-900 text-sm">
                      {step.step_id}
                    </h4>
                    <div className="flex items-center gap-2">
                      <span
                        className={`rounded border px-2 py-1 font-medium text-xs ${getStatusColor(
                          step.status
                        )}`}
                      >
                        {getStepIcon(step.status)} {step.status}
                      </span>
                      {step.retry_count > 0 && (
                        <span className="text-gray-500 text-xs">
                          Retry #{step.retry_count}
                        </span>
                      )}
                    </div>
                  </div>

                  {step.error && (
                    <div className="mt-2 rounded border border-red-200 bg-red-50 p-2 text-red-700 text-sm">
                      {step.error}
                    </div>
                  )}

                  {step.output && (
                    <details className="mt-2">
                      <summary className="cursor-pointer text-gray-600 text-sm hover:text-gray-800">
                        Output
                      </summary>
                      <pre className="mt-2 overflow-x-auto rounded bg-gray-100 p-2 text-xs">
                        {JSON.stringify(step.output, null, 2)}
                      </pre>
                    </details>
                  )}

                  <div className="mt-2 text-gray-500 text-xs">
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
