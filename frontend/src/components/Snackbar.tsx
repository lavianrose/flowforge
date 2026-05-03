"use client";

import type React from "react";
import { createContext, useCallback, useContext, useState } from "react";

type SnackbarType = "success" | "error" | "info" | "warning";

interface SnackbarItem {
  id: number;
  message: string;
  type: SnackbarType;
}

interface SnackbarContextType {
  showSnackbar: (message: string, type?: SnackbarType) => void;
}

const SnackbarContext = createContext<SnackbarContextType | null>(null);

export function useSnackbar() {
  const ctx = useContext(SnackbarContext);
  if (!ctx) {
    throw new Error("useSnackbar must be used within SnackbarProvider");
  }
  return ctx;
}

let nextId = 0;

export function SnackbarProvider({ children }: { children: React.ReactNode }) {
  const [items, setItems] = useState<SnackbarItem[]>([]);

  const showSnackbar = useCallback(
    (message: string, type: SnackbarType = "info") => {
      const id = nextId++;
      setItems((prev) => [...prev, { id, message, type }]);
      setTimeout(() => {
        setItems((prev) => prev.filter((item) => item.id !== id));
      }, 4000);
    },
    []
  );

  const dismiss = useCallback((id: number) => {
    setItems((prev) => prev.filter((item) => item.id !== id));
  }, []);

  const iconMap: Record<SnackbarType, string> = {
    success: "\u2713",
    error: "\u2717",
    warning: "\u26A0",
    info: "\u2139",
  };

  const colorMap: Record<SnackbarType, string> = {
    success: "bg-green-600",
    error: "bg-red-600",
    warning: "bg-yellow-500",
    info: "bg-indigo-600",
  };

  return (
    <SnackbarContext.Provider value={{ showSnackbar }}>
      {children}
      <div className="fixed right-4 bottom-4 z-50 flex max-w-md flex-col gap-2">
        {items.map((item) => (
          <div
            className={`${colorMap[item.type]} flex animate-slide-up items-center gap-3 rounded-lg px-4 py-3 text-white shadow-lg`}
            key={item.id}
            role="alert"
          >
            <span className="flex-shrink-0 font-bold text-lg">
              {iconMap[item.type]}
            </span>
            <span className="flex-1 text-sm">{item.message}</span>
            <button
              className="ml-2 flex-shrink-0 text-white/70 hover:text-white"
              onClick={() => dismiss(item.id)}
            >
              ×
            </button>
          </div>
        ))}
      </div>
    </SnackbarContext.Provider>
  );
}
