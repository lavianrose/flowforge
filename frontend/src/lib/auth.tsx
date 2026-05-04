"use client";

import type React from "react";
import { createContext, useContext, useEffect, useState } from "react";
import { api, type LoginRequest, type LoginResponse } from "./api";

interface User {
  email: string;
  id: string;
  role: string;
  tenant_id: string;
}

interface AuthContextType {
  can: (action: string) => boolean;
  loading: boolean;
  login: (data: LoginRequest) => Promise<void>;
  logout: () => void;
  token: string | null;
  user: User | null;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [token, setToken] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    // Check for stored token on mount
    const storedToken = localStorage.getItem("token");
    if (storedToken) {
      setToken(storedToken);
      api.setToken(storedToken);

      // Verify token and get user
      api
        .getMe()
        .then((data: any) => {
          setUser(data);
        })
        .catch(() => {
          localStorage.removeItem("token");
          setToken(null);
          api.clearToken();
        })
        .finally(() => {
          setLoading(false);
        });
    } else {
      setLoading(false);
    }
  }, []);

  const login = async (data: LoginRequest) => {
    const response: LoginResponse = await api.login(data);
    const { token: newToken, user: newUser } = response;

    setToken(newToken);
    setUser(newUser);
    api.setToken(newToken);
    localStorage.setItem("token", newToken);
  };

  const logout = () => {
    setToken(null);
    setUser(null);
    api.clearToken();
    localStorage.removeItem("token");
  };

  const can = (action: string): boolean => {
    if (!user) {
      return false;
    }

    const role = user.role;

    switch (action) {
      case "view":
        return true;
      case "create":
      case "edit":
      case "trigger":
      case "rollback":
      case "delete_schedule":
      case "delete_webhook":
        return role === "editor" || role === "admin";
      case "delete":
        return role === "admin";
      default:
        return false;
    }
  };

  return (
    <AuthContext.Provider value={{ user, token, login, logout, loading, can }}>
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error("useAuth must be used within an AuthProvider");
  }
  return context;
}
