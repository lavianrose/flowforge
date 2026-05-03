'use client';

import React, { createContext, useContext, useState, useCallback } from 'react';

type SnackbarType = 'success' | 'error' | 'info' | 'warning';

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
  if (!ctx) throw new Error('useSnackbar must be used within SnackbarProvider');
  return ctx;
}

let nextId = 0;

export function SnackbarProvider({ children }: { children: React.ReactNode }) {
  const [items, setItems] = useState<SnackbarItem[]>([]);

  const showSnackbar = useCallback((message: string, type: SnackbarType = 'info') => {
    const id = nextId++;
    setItems((prev) => [...prev, { id, message, type }]);
    setTimeout(() => {
      setItems((prev) => prev.filter((item) => item.id !== id));
    }, 4000);
  }, []);

  const dismiss = useCallback((id: number) => {
    setItems((prev) => prev.filter((item) => item.id !== id));
  }, []);

  const iconMap: Record<SnackbarType, string> = {
    success: '\u2713',
    error: '\u2717',
    warning: '\u26A0',
    info: '\u2139',
  };

  const colorMap: Record<SnackbarType, string> = {
    success: 'bg-green-600',
    error: 'bg-red-600',
    warning: 'bg-yellow-500',
    info: 'bg-indigo-600',
  };

  return (
    <SnackbarContext.Provider value={{ showSnackbar }}>
      {children}
      <div className="fixed bottom-4 right-4 z-50 flex flex-col gap-2 max-w-md">
        {items.map((item) => (
          <div
            key={item.id}
            className={`${colorMap[item.type]} text-white px-4 py-3 rounded-lg shadow-lg flex items-center gap-3 animate-slide-up`}
            role="alert"
          >
            <span className="text-lg font-bold flex-shrink-0">{iconMap[item.type]}</span>
            <span className="text-sm flex-1">{item.message}</span>
            <button
              onClick={() => dismiss(item.id)}
              className="text-white/70 hover:text-white ml-2 flex-shrink-0"
            >
              \u00d7
            </button>
          </div>
        ))}
      </div>
    </SnackbarContext.Provider>
  );
}
