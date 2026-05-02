'use client';

import React, { createContext, useContext, useEffect, useState } from 'react';
import { api, LoginRequest, LoginResponse } from './api';

interface User {
  id: string;
  email: string;
  role: string;
  tenant_id: string;
}

interface AuthContextType {
  user: User | null;
  token: string | null;
  login: (data: LoginRequest) => Promise<void>;
  logout: () => void;
  loading: boolean;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [token, setToken] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    // Check for stored token on mount
    const storedToken = localStorage.getItem('token');
    if (storedToken) {
      setToken(storedToken);
      api.setToken(storedToken);

      // Verify token and get user
      api.getMe()
        .then((data: any) => {
          setUser(data);
        })
        .catch(() => {
          localStorage.removeItem('token');
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
    localStorage.setItem('token', newToken);
  };

  const logout = () => {
    setToken(null);
    setUser(null);
    api.clearToken();
    localStorage.removeItem('token');
  };

  return (
    <AuthContext.Provider value={{ user, token, login, logout, loading }}>
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
}
