'use client';

import { useHealthStats } from '@/lib/hooks';
import { useRouter } from 'next/navigation';
import {
  BarChart,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
  ResponsiveContainer,
  LineChart,
  Line,
} from 'recharts';

export default function DashboardPage() {
  const router = useRouter();
  const { data: stats, isLoading, error, refetch } = useHealthStats();

  // Prepare chart data
  const chartData = stats?.hourly_stats?.map((stat) => ({
    hour: `${stat.hour}:00`,
    total: stat.total_runs,
    success: stat.success_runs,
    failed: stat.failed_runs,
    duration: stat.avg_duration,
  })) || [];

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="text-xl">Loading dashboard...</div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded">
        {error.message || 'Failed to load stats'}
      </div>
    );
  }

  return (
    <div>
      <div className="flex justify-between items-center mb-6">
        <div>
          <h2 className="text-2xl font-bold text-gray-900">Dashboard</h2>
          <p className="text-sm text-gray-600 mt-1">
            Overview of your workflow performance
          </p>
        </div>
        <div className="flex space-x-3">
          <button
            onClick={() => router.push('/dashboard/workflows/new')}
            className="px-4 py-2 bg-green-600 text-white rounded-md hover:bg-green-700"
          >
            Create Workflow
          </button>
          <button
            onClick={() => refetch()}
            className="px-4 py-2 bg-indigo-600 text-white rounded-md hover:bg-indigo-700"
          >
            Refresh
          </button>
        </div>
      </div>

      {/* Quick Actions */}
      <div className="bg-linear-to-r from-indigo-500 to-purple-600 rounded-lg shadow-lg p-6 mb-8 text-white">
        <h3 className="text-xl font-bold mb-2">Get Started</h3>
        <p className="text-indigo-100 mb-4">Create your first automation workflow to get started</p>
        <div className="flex space-x-3">
          <button
            onClick={() => router.push('/dashboard/workflows/new')}
            className="px-6 py-3 bg-white text-indigo-600 font-semibold rounded-md hover:bg-indigo-50"
          >
            Create Workflow
          </button>
          <button
            onClick={() => router.push('/dashboard/workflows')}
            className="px-6 py-3 bg-indigo-700 text-white font-semibold rounded-md hover:bg-indigo-800"
          >
            View All Workflows
          </button>
        </div>
      </div>

      {/* Stats Cards */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
        {/* Active Runs */}
        <div className="bg-white rounded-lg shadow p-6">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm font-medium text-gray-500">Active Runs</p>
              <p className="text-3xl font-bold text-gray-900 mt-2">
                {stats?.active_runs || 0}
              </p>
            </div>
            <div className="h-12 w-12 bg-blue-100 rounded-full flex items-center justify-center">
              <span className="text-2xl">⚡</span>
            </div>
          </div>
          <p className="text-xs text-gray-500 mt-4">
            Currently executing workflows
          </p>
        </div>

        {/* Success Rate */}
        <div className="bg-white rounded-lg shadow p-6">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm font-medium text-gray-500">Success Rate</p>
              <p className="text-3xl font-bold text-gray-900 mt-2">
                {(stats?.success_rate ?? 0).toFixed(1)}%
              </p>
            </div>
            <div className="h-12 w-12 bg-green-100 rounded-full flex items-center justify-center">
              <span className="text-2xl">✅</span>
            </div>
          </div>
          <p className="text-xs text-gray-500 mt-4">
            Last 24 hours
          </p>
        </div>

        {/* Failure Rate */}
        <div className="bg-white rounded-lg shadow p-6">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm font-medium text-gray-500">Failure Rate</p>
              <p className="text-3xl font-bold text-gray-900 mt-2">
                {(stats?.failure_rate ?? 0).toFixed(1)}%
              </p>
            </div>
            <div className="h-12 w-12 bg-red-100 rounded-full flex items-center justify-center">
              <span className="text-2xl">❌</span>
            </div>
          </div>
          <p className="text-xs text-gray-500 mt-4">
            Last 24 hours
          </p>
        </div>

        {/* Avg Duration */}
        <div className="bg-white rounded-lg shadow p-6">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm font-medium text-gray-500">Avg Duration</p>
              <p className="text-3xl font-bold text-gray-900 mt-2">
                {(stats?.avg_duration_seconds ?? 0).toFixed(1)}s
              </p>
            </div>
            <div className="h-12 w-12 bg-purple-100 rounded-full flex items-center justify-center">
              <span className="text-2xl">⏱️</span>
            </div>
          </div>
          <p className="text-xs text-gray-500 mt-4">
            Average execution time
          </p>
        </div>
      </div>

      {/* Charts */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Runs by Hour Chart */}
        <div className="bg-white rounded-lg shadow p-6">
          <h3 className="text-lg font-semibold mb-4">Runs by Hour (Last 24h)</h3>
          <ResponsiveContainer width="100%" height={300}>
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
        <div className="bg-white rounded-lg shadow p-6">
          <h3 className="text-lg font-semibold mb-4">
            Avg Duration by Hour (seconds)
          </h3>
          <ResponsiveContainer width="100%" height={300}>
            <LineChart data={chartData}>
              <CartesianGrid strokeDasharray="3 3" />
              <XAxis dataKey="hour" />
              <YAxis />
              <Tooltip />
              <Legend />
              <Line
                type="monotone"
                dataKey="duration"
                stroke="#8b5cf6"
                strokeWidth={2}
                name="Avg Duration"
              />
            </LineChart>
          </ResponsiveContainer>
        </div>
      </div>

      {/* Summary Stats */}
      <div className="mt-6 bg-white rounded-lg shadow p-6">
        <h3 className="text-lg font-semibold mb-4">Summary (Last 24 Hours)</h3>
        <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
          <div className="text-center">
            <p className="text-sm text-gray-500">Total Runs</p>
            <p className="text-2xl font-bold text-gray-900 mt-1">
              {stats?.total_runs_24h || 0}
            </p>
          </div>
          <div className="text-center">
            <p className="text-sm text-gray-500">Successful</p>
            <p className="text-2xl font-bold text-green-600 mt-1">
              {stats?.success_runs_24h || 0}
            </p>
          </div>
          <div className="text-center">
            <p className="text-sm text-gray-500">Failed</p>
            <p className="text-2xl font-bold text-red-600 mt-1">
              {stats?.failed_runs_24h || 0}
            </p>
          </div>
        </div>
      </div>
    </div>
  );
}
