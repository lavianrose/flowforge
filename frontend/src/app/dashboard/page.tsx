"use client";

import { useRouter } from "next/navigation";
import {
  Bar,
  BarChart,
  CartesianGrid,
  Legend,
  Line,
  LineChart,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from "recharts";
import { useHealthStats } from "@/lib/hooks";

export default function DashboardPage() {
  const router = useRouter();
  const { data: stats, isLoading, error, refetch } = useHealthStats();

  // Prepare chart data
  const chartData =
    stats?.hourly_stats?.map((stat) => ({
      hour: `${stat.hour}:00`,
      total: stat.total_runs,
      success: stat.success_runs,
      failed: stat.failed_runs,
      duration: stat.avg_duration,
    })) || [];

  if (isLoading) {
    return (
      <div className="flex h-64 items-center justify-center">
        <div className="text-xl">Loading dashboard...</div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="rounded border border-red-200 bg-red-50 px-4 py-3 text-red-700">
        {error.message || "Failed to load stats"}
      </div>
    );
  }

  return (
    <div>
      <div className="mb-6 flex items-center justify-between">
        <div>
          <h2 className="font-bold text-2xl text-gray-900">Dashboard</h2>
          <p className="mt-1 text-gray-600 text-sm">
            Overview of your workflow performance
          </p>
        </div>
        <div className="flex space-x-3">
          <button
            className="rounded-md bg-green-600 px-4 py-2 text-white hover:bg-green-700"
            onClick={() => router.push("/dashboard/workflows/new")}
          >
            Create Workflow
          </button>
          <button
            className="rounded-md bg-indigo-600 px-4 py-2 text-white hover:bg-indigo-700"
            onClick={() => refetch()}
          >
            Refresh
          </button>
        </div>
      </div>

      {/* Quick Actions */}
      <div className="mb-8 rounded-lg bg-linear-to-r from-indigo-500 to-purple-600 p-6 text-white shadow-lg">
        <h3 className="mb-2 font-bold text-xl">Get Started</h3>
        <p className="mb-4 text-indigo-100">
          Create your first automation workflow to get started
        </p>
        <div className="flex space-x-3">
          <button
            className="rounded-md bg-white px-6 py-3 font-semibold text-indigo-600 hover:bg-indigo-50"
            onClick={() => router.push("/dashboard/workflows/new")}
          >
            Create Workflow
          </button>
          <button
            className="rounded-md bg-indigo-700 px-6 py-3 font-semibold text-white hover:bg-indigo-800"
            onClick={() => router.push("/dashboard/workflows")}
          >
            View All Workflows
          </button>
        </div>
      </div>

      {/* Stats Cards */}
      <div className="mb-8 grid grid-cols-1 gap-6 md:grid-cols-2 lg:grid-cols-4">
        {/* Active Runs */}
        <div className="rounded-lg bg-white p-6 shadow">
          <div className="flex items-center justify-between">
            <div>
              <p className="font-medium text-gray-500 text-sm">Active Runs</p>
              <p className="mt-2 font-bold text-3xl text-gray-900">
                {stats?.active_runs || 0}
              </p>
            </div>
            <div className="flex h-12 w-12 items-center justify-center rounded-full bg-blue-100">
              <span className="text-2xl">⚡</span>
            </div>
          </div>
          <p className="mt-4 text-gray-500 text-xs">
            Currently executing workflows
          </p>
        </div>

        {/* Success Rate */}
        <div className="rounded-lg bg-white p-6 shadow">
          <div className="flex items-center justify-between">
            <div>
              <p className="font-medium text-gray-500 text-sm">Success Rate</p>
              <p className="mt-2 font-bold text-3xl text-gray-900">
                {(stats?.success_rate ?? 0).toFixed(1)}%
              </p>
            </div>
            <div className="flex h-12 w-12 items-center justify-center rounded-full bg-green-100">
              <span className="text-2xl">✅</span>
            </div>
          </div>
          <p className="mt-4 text-gray-500 text-xs">Last 24 hours</p>
        </div>

        {/* Failure Rate */}
        <div className="rounded-lg bg-white p-6 shadow">
          <div className="flex items-center justify-between">
            <div>
              <p className="font-medium text-gray-500 text-sm">Failure Rate</p>
              <p className="mt-2 font-bold text-3xl text-gray-900">
                {(stats?.failure_rate ?? 0).toFixed(1)}%
              </p>
            </div>
            <div className="flex h-12 w-12 items-center justify-center rounded-full bg-red-100">
              <span className="text-2xl">❌</span>
            </div>
          </div>
          <p className="mt-4 text-gray-500 text-xs">Last 24 hours</p>
        </div>

        {/* Avg Duration */}
        <div className="rounded-lg bg-white p-6 shadow">
          <div className="flex items-center justify-between">
            <div>
              <p className="font-medium text-gray-500 text-sm">Avg Duration</p>
              <p className="mt-2 font-bold text-3xl text-gray-900">
                {(stats?.avg_duration_seconds ?? 0).toFixed(1)}s
              </p>
            </div>
            <div className="flex h-12 w-12 items-center justify-center rounded-full bg-purple-100">
              <span className="text-2xl">⏱️</span>
            </div>
          </div>
          <p className="mt-4 text-gray-500 text-xs">Average execution time</p>
        </div>
      </div>

      {/* Charts */}
      <div className="grid grid-cols-1 gap-6 lg:grid-cols-2">
        {/* Runs by Hour Chart */}
        <div className="rounded-lg bg-white p-6 shadow">
          <h3 className="mb-4 font-semibold text-lg">
            Runs by Hour (Last 24h)
          </h3>
          <ResponsiveContainer height={300} width="100%">
            <BarChart data={chartData}>
              <CartesianGrid strokeDasharray="3 3" />
              <XAxis dataKey="hour" />
              <YAxis />
              <Tooltip />
              <Legend />
              <Bar dataKey="success" fill="#10b981" name="Success" />
              <Bar dataKey="failed" fill="#ef4444" name="Failed" />
            </BarChart>
          </ResponsiveContainer>
        </div>

        {/* Average Duration Chart */}
        <div className="rounded-lg bg-white p-6 shadow">
          <h3 className="mb-4 font-semibold text-lg">
            Avg Duration by Hour (seconds)
          </h3>
          <ResponsiveContainer height={300} width="100%">
            <LineChart data={chartData}>
              <CartesianGrid strokeDasharray="3 3" />
              <XAxis dataKey="hour" />
              <YAxis />
              <Tooltip />
              <Legend />
              <Line
                dataKey="duration"
                name="Avg Duration"
                stroke="#8b5cf6"
                strokeWidth={2}
                type="monotone"
              />
            </LineChart>
          </ResponsiveContainer>
        </div>
      </div>

      {/* Summary Stats */}
      <div className="mt-6 rounded-lg bg-white p-6 shadow">
        <h3 className="mb-4 font-semibold text-lg">Summary (Last 24 Hours)</h3>
        <div className="grid grid-cols-1 gap-6 md:grid-cols-3">
          <div className="text-center">
            <p className="text-gray-500 text-sm">Total Runs</p>
            <p className="mt-1 font-bold text-2xl text-gray-900">
              {stats?.total_runs_24h || 0}
            </p>
          </div>
          <div className="text-center">
            <p className="text-gray-500 text-sm">Successful</p>
            <p className="mt-1 font-bold text-2xl text-green-600">
              {stats?.success_runs_24h || 0}
            </p>
          </div>
          <div className="text-center">
            <p className="text-gray-500 text-sm">Failed</p>
            <p className="mt-1 font-bold text-2xl text-red-600">
              {stats?.failed_runs_24h || 0}
            </p>
          </div>
        </div>
      </div>
    </div>
  );
}
