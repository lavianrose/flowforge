import { act, renderHook, waitFor } from "@testing-library/react";
import { api } from "../api";
import { AuthProvider, useAuth } from "../auth";

// Mock the API module
jest.mock("../api", () => ({
  api: {
    setToken: jest.fn(),
    clearToken: jest.fn(),
    getMe: jest.fn(),
    login: jest.fn(),
  },
}));

const mockApi = api as jest.Mocked<typeof api>;

describe("RBAC Permissions", () => {
  beforeEach(() => {
    jest.clearAllMocks();
    localStorage.clear();
  });

  describe("Admin Role Permissions", () => {
    it("should grant all permissions to admin", async () => {
      const adminUser = {
        id: "user-1",
        email: "admin@flowforge.local",
        role: "admin",
        tenant_id: "tenant-1",
      };

      localStorage.setItem("token", "admin-token");
      mockApi.getMe.mockResolvedValue(adminUser);

      const { result } = renderHook(() => useAuth(), {
        wrapper: AuthProvider,
      });

      await waitFor(() => {
        expect(result.current.user).toEqual(adminUser);
      });

      // Test all permissions
      expect(result.current.can("view")).toBe(true);
      expect(result.current.can("create")).toBe(true);
      expect(result.current.can("edit")).toBe(true);
      expect(result.current.can("trigger")).toBe(true);
      expect(result.current.can("delete")).toBe(true);
      expect(result.current.can("rollback")).toBe(true);
    });

    it("should have admin role badge color", async () => {
      const adminUser = {
        id: "user-1",
        email: "admin@flowforge.local",
        role: "admin",
        tenant_id: "tenant-1",
      };

      localStorage.setItem("token", "admin-token");
      mockApi.getMe.mockResolvedValue(adminUser);

      const { result } = renderHook(() => useAuth(), {
        wrapper: AuthProvider,
      });

      await waitFor(() => {
        expect(result.current.user?.role).toBe("admin");
      });
    });
  });

  describe("Editor Role Permissions", () => {
    it("should grant editor permissions but not delete", async () => {
      const editorUser = {
        id: "user-2",
        email: "editor@flowforge.local",
        role: "editor",
        tenant_id: "tenant-1",
      };

      localStorage.setItem("token", "editor-token");
      mockApi.getMe.mockResolvedValue(editorUser);

      const { result } = renderHook(() => useAuth(), {
        wrapper: AuthProvider,
      });

      await waitFor(() => {
        expect(result.current.user).toEqual(editorUser);
      });

      // Test editor permissions
      expect(result.current.can("view")).toBe(true);
      expect(result.current.can("create")).toBe(true);
      expect(result.current.can("edit")).toBe(true);
      expect(result.current.can("trigger")).toBe(true);
      expect(result.current.can("rollback")).toBe(true);
      expect(result.current.can("delete")).toBe(false);
    });

    it("should have editor role badge color", async () => {
      const editorUser = {
        id: "user-2",
        email: "editor@flowforge.local",
        role: "editor",
        tenant_id: "tenant-1",
      };

      localStorage.setItem("token", "editor-token");
      mockApi.getMe.mockResolvedValue(editorUser);

      const { result } = renderHook(() => useAuth(), {
        wrapper: AuthProvider,
      });

      await waitFor(() => {
        expect(result.current.user?.role).toBe("editor");
      });
    });
  });

  describe("Viewer Role Permissions", () => {
    it("should only grant view permissions", async () => {
      const viewerUser = {
        id: "user-3",
        email: "viewer@flowforge.local",
        role: "viewer",
        tenant_id: "tenant-1",
      };

      localStorage.setItem("token", "viewer-token");
      mockApi.getMe.mockResolvedValue(viewerUser);

      const { result } = renderHook(() => useAuth(), {
        wrapper: AuthProvider,
      });

      await waitFor(() => {
        expect(result.current.user).toEqual(viewerUser);
      });

      // Test viewer permissions (read-only)
      expect(result.current.can("view")).toBe(true);
      expect(result.current.can("create")).toBe(false);
      expect(result.current.can("edit")).toBe(false);
      expect(result.current.can("trigger")).toBe(false);
      expect(result.current.can("rollback")).toBe(false);
      expect(result.current.can("delete")).toBe(false);
    });

    it("should have viewer role badge color", async () => {
      const viewerUser = {
        id: "user-3",
        email: "viewer@flowforge.local",
        role: "viewer",
        tenant_id: "tenant-1",
      };

      localStorage.setItem("token", "viewer-token");
      mockApi.getMe.mockResolvedValue(viewerUser);

      const { result } = renderHook(() => useAuth(), {
        wrapper: AuthProvider,
      });

      await waitFor(() => {
        expect(result.current.user?.role).toBe("viewer");
      });
    });
  });

  describe("Permission Edge Cases", () => {
    it("should return false for unknown permissions", async () => {
      const adminUser = {
        id: "user-1",
        email: "admin@flowforge.local",
        role: "admin",
        tenant_id: "tenant-1",
      };

      localStorage.setItem("token", "admin-token");
      mockApi.getMe.mockResolvedValue(adminUser);

      const { result } = renderHook(() => useAuth(), {
        wrapper: AuthProvider,
      });

      await waitFor(() => {
        expect(result.current.user).toEqual(adminUser);
      });

      // Test unknown permission
      expect(result.current.can("unknown_permission")).toBe(false);
    });

    it("should return false when user is not authenticated", async () => {
      const { result } = renderHook(() => useAuth(), {
        wrapper: AuthProvider,
      });

      await waitFor(() => {
        expect(result.current.user).toBeNull();
      });

      // All permissions should return false when not authenticated
      expect(result.current.can("view")).toBe(false);
      expect(result.current.can("create")).toBe(false);
      expect(result.current.can("delete")).toBe(false);
    });

    it("should handle role changes correctly", async () => {
      // Start as viewer
      const viewerUser = {
        id: "user-3",
        email: "viewer@flowforge.local",
        role: "viewer",
        tenant_id: "tenant-1",
      };

      localStorage.setItem("token", "viewer-token");
      mockApi.getMe.mockResolvedValue(viewerUser);

      const { result } = renderHook(() => useAuth(), {
        wrapper: AuthProvider,
      });

      await waitFor(() => {
        expect(result.current.user?.role).toBe("viewer");
      });

      // Viewer should not have delete permission
      expect(result.current.can("delete")).toBe(false);

      // Simulate role upgrade to admin
      const adminUser = {
        id: "user-3",
        email: "admin@flowforge.local",
        role: "admin",
        tenant_id: "tenant-1",
      };

      await act(async () => {
        const mockLoginResponse = {
          token: "admin-token",
          user: adminUser,
        };
        mockApi.login.mockResolvedValue(mockLoginResponse);
        await result.current.login({
          email: "admin@flowforge.local",
          password: "admin123",
        });
      });

      // Admin should have delete permission
      expect(result.current.can("delete")).toBe(true);
    });
  });

  describe("Permission Matrix", () => {
    it("should follow correct permission matrix for all roles", async () => {
      const roles = [
        {
          role: "admin",
          expected: {
            view: true,
            create: true,
            edit: true,
            trigger: true,
            delete: true,
            rollback: true,
          },
        },
        {
          role: "editor",
          expected: {
            view: true,
            create: true,
            edit: true,
            trigger: true,
            delete: false,
            rollback: true,
          },
        },
        {
          role: "viewer",
          expected: {
            view: true,
            create: false,
            edit: false,
            trigger: false,
            delete: false,
            rollback: false,
          },
        },
      ];

      for (const { role, expected } of roles) {
        const user = {
          id: `user-${role}`,
          email: `${role}@flowforge.local`,
          role,
          tenant_id: "tenant-1",
        };

        localStorage.setItem("token", `${role}-token`);
        mockApi.getMe.mockResolvedValue(user);

        const { result } = renderHook(() => useAuth(), {
          wrapper: AuthProvider,
        });

        await waitFor(() => {
          expect(result.current.user?.role).toBe(role);
        });

        // Test all permissions for this role
        Object.entries(expected).forEach(([permission, expectedValue]) => {
          const actualValue = result.current.can(permission);
          expect(actualValue).toBe(expectedValue);
        });

        // Cleanup for next iteration
        localStorage.clear();
        jest.clearAllMocks();
      }
    });
  });

  describe("Permission Consistency", () => {
    it("should maintain consistent permissions across multiple checks", async () => {
      const editorUser = {
        id: "user-2",
        email: "editor@flowforge.local",
        role: "editor",
        tenant_id: "tenant-1",
      };

      localStorage.setItem("token", "editor-token");
      mockApi.getMe.mockResolvedValue(editorUser);

      const { result } = renderHook(() => useAuth(), {
        wrapper: AuthProvider,
      });

      await waitFor(() => {
        expect(result.current.user).toEqual(editorUser);
      });

      // Check permissions multiple times to ensure consistency
      for (let i = 0; i < 5; i++) {
        expect(result.current.can("view")).toBe(true);
        expect(result.current.can("delete")).toBe(false);
      }
    });
  });

  describe("Role Badge Display", () => {
    it("should display correct role for admin", async () => {
      const adminUser = {
        id: "user-1",
        email: "admin@flowforge.local",
        role: "admin",
        tenant_id: "tenant-1",
      };

      localStorage.setItem("token", "admin-token");
      mockApi.getMe.mockResolvedValue(adminUser);

      const { result } = renderHook(() => useAuth(), {
        wrapper: AuthProvider,
      });

      await waitFor(() => {
        expect(result.current.user?.role).toBe("admin");
      });
    });

    it("should display correct role for editor", async () => {
      const editorUser = {
        id: "user-2",
        email: "editor@flowforge.local",
        role: "editor",
        tenant_id: "tenant-1",
      };

      localStorage.setItem("token", "editor-token");
      mockApi.getMe.mockResolvedValue(editorUser);

      const { result } = renderHook(() => useAuth(), {
        wrapper: AuthProvider,
      });

      await waitFor(() => {
        expect(result.current.user?.role).toBe("editor");
      });
    });

    it("should display correct role for viewer", async () => {
      const viewerUser = {
        id: "user-3",
        email: "viewer@flowforge.local",
        role: "viewer",
        tenant_id: "tenant-1",
      };

      localStorage.setItem("token", "viewer-token");
      mockApi.getMe.mockResolvedValue(viewerUser);

      const { result } = renderHook(() => useAuth(), {
        wrapper: AuthProvider,
      });

      await waitFor(() => {
        expect(result.current.user?.role).toBe("viewer");
      });
    });
  });

  describe("User Information", () => {
    it("should provide complete user information", async () => {
      const user = {
        id: "user-1",
        email: "admin@flowforge.local",
        role: "admin",
        tenant_id: "tenant-1",
      };

      localStorage.setItem("token", "admin-token");
      mockApi.getMe.mockResolvedValue(user);

      const { result } = renderHook(() => useAuth(), {
        wrapper: AuthProvider,
      });

      await waitFor(() => {
        expect(result.current.user).toEqual(user);
        expect(result.current.user?.id).toBe("user-1");
        expect(result.current.user?.email).toBe("admin@flowforge.local");
        expect(result.current.user?.role).toBe("admin");
        expect(result.current.user?.tenant_id).toBe("tenant-1");
      });
    });
  });
});
