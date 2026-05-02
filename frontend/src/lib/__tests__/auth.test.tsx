import { renderHook, act, waitFor } from '@testing-library/react';
import { AuthProvider, useAuth } from '../auth';
import { api } from '../api';

// Mock the API module
jest.mock('../api', () => ({
  api: {
    setToken: jest.fn(),
    clearToken: jest.fn(),
    getMe: jest.fn(),
    login: jest.fn(),
  },
}));

const mockApi = api as jest.Mocked<typeof api>;

describe('AuthProvider', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    localStorage.clear();
  });

  it('should provide auth context', () => {
    const { result } = renderHook(() => useAuth(), {
      wrapper: AuthProvider,
    });

    expect(result.current).toBeDefined();
    expect(result.current.user).toBeNull();
    expect(result.current.token).toBeNull();
    expect(result.current.loading).toBe(false);
  });

  it('should load user from stored token on mount', async () => {
    const mockUser = {
      id: 'user-1',
      email: 'test@example.com',
      role: 'admin',
      tenant_id: 'tenant-1',
    };

    localStorage.setItem('token', 'stored-token');
    mockApi.getMe.mockResolvedValue(mockUser);

    const { result } = renderHook(() => useAuth(), {
      wrapper: AuthProvider,
    });

    expect(result.current.loading).toBe(true);

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
      expect(result.current.user).toEqual(mockUser);
      expect(result.current.token).toBe('stored-token');
      expect(mockApi.setToken).toHaveBeenCalledWith('stored-token');
    });
  });

  it('should handle invalid stored token', async () => {
    localStorage.setItem('token', 'invalid-token');
    mockApi.getMe.mockRejectedValue(new Error('Unauthorized'));

    const { result } = renderHook(() => useAuth(), {
      wrapper: AuthProvider,
    });

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
      expect(result.current.user).toBeNull();
      expect(result.current.token).toBeNull();
      expect(localStorage.getItem('token')).toBeNull();
    });
  });

  it('should login successfully', async () => {
    const mockUser = {
      id: 'user-1',
      email: 'test@example.com',
      role: 'admin',
      tenant_id: 'tenant-1',
    };

    const mockLoginResponse = {
      token: 'new-token',
      user: mockUser,
    };

    mockApi.login.mockResolvedValue(mockLoginResponse);

    const { result } = renderHook(() => useAuth(), {
      wrapper: AuthProvider,
    });

    await act(async () => {
      await result.current.login({
        email: 'test@example.com',
        password: 'password',
      });
    });

    expect(result.current.token).toBe('new-token');
    expect(result.current.user).toEqual(mockUser);
    expect(localStorage.getItem('token')).toBe('new-token');
    expect(mockApi.setToken).toHaveBeenCalledWith('new-token');
  });

  it('should logout successfully', async () => {
    // Start with logged in state
    localStorage.setItem('token', 'test-token');
    const mockUser = {
      id: 'user-1',
      email: 'test@example.com',
      role: 'admin',
      tenant_id: 'tenant-1',
    };
    mockApi.getMe.mockResolvedValue(mockUser);

    const { result } = renderHook(() => useAuth(), {
      wrapper: AuthProvider,
    });

    await waitFor(() => {
      expect(result.current.user).toEqual(mockUser);
    });

    // Logout
    act(() => {
      result.current.logout();
    });

    expect(result.current.user).toBeNull();
    expect(result.current.token).toBeNull();
    expect(localStorage.getItem('token')).toBeNull();
    expect(mockApi.clearToken).toHaveBeenCalled();
  });

  it('should throw error when useAuth is used outside AuthProvider', () => {
    // Suppress console.error for this test
    const consoleError = console.error;
    console.error = jest.fn();

    expect(() => {
      renderHook(() => useAuth());
    }).toThrow('useAuth must be used within an AuthProvider');

    console.error = consoleError;
  });
});

describe('useAuth hook', () => {
  it('should return auth context values', async () => {
    const mockUser = {
      id: 'user-1',
      email: 'test@example.com',
      role: 'admin',
      tenant_id: 'tenant-1',
    };

    mockApi.getMe.mockResolvedValue(mockUser);
    localStorage.setItem('token', 'test-token');

    const { result } = renderHook(() => useAuth(), {
      wrapper: AuthProvider,
    });

    await waitFor(() => {
      expect(result.current.user).toEqual(mockUser);
      expect(result.current.token).toBe('test-token');
      expect(typeof result.current.login).toBe('function');
      expect(typeof result.current.logout).toBe('function');
      expect(typeof result.current.loading).toBe('boolean');
    });
  });
});
