import type { Metadata } from "next";
import "./globals.css";
import { QueryProvider } from "@/components/QueryProvider";
import { SnackbarProvider } from "@/components/Snackbar";
import { AuthProvider } from "@/lib/auth";

export const metadata: Metadata = {
  title: "FlowForge - Workflow Automation",
  description: "Real-time multi-tenant workflow orchestration platform",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html className="h-full" lang="en">
      <head>
        <meta content="light" name="color-scheme" />
        <meta content="#ffffff" name="theme-color" />
      </head>
      <body className="flex min-h-full flex-col font-sans">
        <QueryProvider>
          <SnackbarProvider>
            <AuthProvider>{children}</AuthProvider>
          </SnackbarProvider>
        </QueryProvider>
      </body>
    </html>
  );
}
