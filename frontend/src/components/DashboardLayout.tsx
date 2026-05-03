"use client";

import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import { useEffect } from "react";
import { useAuth } from "@/lib/auth";

export default function DashboardLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const { user, loading, logout } = useAuth();
  const router = useRouter();
  const pathname = usePathname();

  useEffect(() => {
    if (!(loading || user)) {
      router.push("/login");
    }
  }, [user, loading, router]);

  if (loading) {
    return (
      <div className="flex min-h-screen items-center justify-center">
        <div className="text-xl">Loading...</div>
      </div>
    );
  }

  if (!user) {
    return null;
  }

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Header */}
      <header className="border-gray-200 border-b bg-white shadow-sm">
        <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
          <div className="flex h-16 items-center justify-between">
            <div className="flex items-center">
              <h1 className="font-bold text-2xl text-gray-900">FlowForge</h1>
            </div>
            <div className="flex items-center space-x-4">
              <span className="text-gray-700 text-sm">{user.email}</span>
              <span
                className={`rounded px-2 py-1 font-medium text-xs ${
                  user.role === "admin"
                    ? "bg-red-100 text-red-800"
                    : user.role === "editor"
                      ? "bg-blue-100 text-blue-800"
                      : "bg-gray-100 text-gray-800"
                }`}
              >
                {user.role.charAt(0).toUpperCase() + user.role.slice(1)}
              </span>
              <button
                className="rounded-md px-3 py-2 font-medium text-gray-700 text-sm hover:bg-gray-100 hover:text-gray-900"
                onClick={logout}
              >
                Logout
              </button>
            </div>
          </div>
        </div>
      </header>

      {/* Navigation */}
      <nav className="border-gray-200 border-b bg-white">
        <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
          <div className="flex space-x-8">
            <Link
              className={`inline-flex items-center border-b-2 px-1 pt-1 font-medium text-sm ${
                pathname === "/dashboard" ||
                pathname.startsWith("/dashboard/workflows")
                  ? "border-indigo-500 text-gray-900"
                  : "border-transparent text-gray-500 hover:border-gray-300 hover:text-gray-700"
              }`}
              href="/dashboard"
            >
              Workflows
            </Link>
            <Link
              className={`inline-flex items-center border-b-2 px-1 pt-1 font-medium text-sm ${
                pathname.startsWith("/dashboard/runs")
                  ? "border-indigo-500 text-gray-900"
                  : "border-transparent text-gray-500 hover:border-gray-300 hover:text-gray-700"
              }`}
              href="/dashboard/runs"
            >
              Runs
            </Link>
          </div>
        </div>
      </nav>

      {/* Main content */}
      <main className="mx-auto max-w-7xl px-4 py-8 sm:px-6 lg:px-8">
        {children}
      </main>
    </div>
  );
}
