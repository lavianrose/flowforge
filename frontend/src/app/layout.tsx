import type { Metadata } from "next";
import "./globals.css";
import { AuthProvider } from "@/lib/auth";
import { QueryProvider } from "@/components/QueryProvider";
import { SnackbarProvider } from "@/components/Snackbar";

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
    <html lang="en" className="h-full">
      <head>
        <meta name="color-scheme" content="light" />
        <meta name="theme-color" content="#ffffff" />
      </head>
      <body className="min-h-full flex flex-col font-sans">
        <QueryProvider>
          <SnackbarProvider>
            <AuthProvider>{children}</AuthProvider>
          </SnackbarProvider>
        </QueryProvider>
      </body>
    </html>
  );
}
