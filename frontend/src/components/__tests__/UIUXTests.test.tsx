import { render, screen, waitFor } from "@testing-library/react";
import DashboardLayout from "@/components/DashboardLayout";
import ProtectedRoute from "@/components/ProtectedRoute";
import { api } from "@/lib/api";
import { AuthProvider, useAuth } from "@/lib/auth";

jest.mock("next/navigation", () => ({
  useRouter: jest.fn(() => ({ push: jest.fn() })),
  usePathname: jest.fn(() => "/dashboard"),
}));

jest.mock("@/lib/api", () => ({
  api: {
    setToken: jest.fn(),
    clearToken: jest.fn(),
    getMe: jest.fn(),
    login: jest.fn(),
  },
}));

const mockApi = api as jest.Mocked<typeof api>;

function seedUser(role: string) {
  const user = {
    id: `u-${role}`,
    email: `${role}@test.com`,
    role,
    tenant_id: "t-1",
  };
  localStorage.setItem("token", `${role}-token`);
  mockApi.getMe.mockResolvedValue(user);
  return user;
}

// ─── Role Badge Color Tests ────────────────────────────────────────────
describe("UI/UX: Role Badge Color Coding", () => {
  beforeEach(() => {
    jest.clearAllMocks();
    localStorage.clear();
  });

  it("applies red classes for Admin badge", async () => {
    seedUser("admin");
    render(
      <AuthProvider>
        <DashboardLayout>
          <div>X</div>
        </DashboardLayout>
      </AuthProvider>
    );
    const badge = await screen.findByText("Admin");
    expect(badge.className).toContain("bg-red-100");
    expect(badge.className).toContain("text-red-800");
  });

  it("applies blue classes for Editor badge", async () => {
    seedUser("editor");
    render(
      <AuthProvider>
        <DashboardLayout>
          <div>X</div>
        </DashboardLayout>
      </AuthProvider>
    );
    const badge = await screen.findByText("Editor");
    expect(badge.className).toContain("bg-blue-100");
    expect(badge.className).toContain("text-blue-800");
  });

  it("applies gray classes for Viewer badge", async () => {
    seedUser("viewer");
    render(
      <AuthProvider>
        <DashboardLayout>
          <div>X</div>
        </DashboardLayout>
      </AuthProvider>
    );
    const badge = await screen.findByText("Viewer");
    expect(badge.className).toContain("bg-gray-100");
    expect(badge.className).toContain("text-gray-800");
  });

  it("capitalizes role name in badge", async () => {
    seedUser("admin");
    render(
      <AuthProvider>
        <DashboardLayout>
          <div>X</div>
        </DashboardLayout>
      </AuthProvider>
    );
    const badge = await screen.findByText("Admin");
    expect(badge.textContent).toBe("Admin");
    expect(badge.textContent).not.toBe("admin");
  });

  it("badge has xs font size and medium weight", async () => {
    seedUser("admin");
    render(
      <AuthProvider>
        <DashboardLayout>
          <div>X</div>
        </DashboardLayout>
      </AuthProvider>
    );
    const badge = await screen.findByText("Admin");
    expect(badge.className).toContain("text-xs");
    expect(badge.className).toContain("font-medium");
  });
});

// ─── Smooth Hiding/Showing of Buttons ──────────────────────────────────
describe("UI/UX: Permission-based Button Visibility", () => {
  beforeEach(() => {
    jest.clearAllMocks();
    localStorage.clear();
  });

  it("shows Create button only for editor+ roles", async () => {
    const roles = [
      { role: "admin", visible: true },
      { role: "editor", visible: true },
      { role: "viewer", visible: false },
    ];

    for (const { role, visible } of roles) {
      seedUser(role);
      const { unmount } = render(
        <AuthProvider>
          <ProtectedRoute requiredPermission="create">
            <button>Create Workflow</button>
          </ProtectedRoute>
        </AuthProvider>
      );

      await waitFor(() => {
        if (visible) {
          expect(screen.getByText("Create Workflow")).toBeInTheDocument();
        } else {
          expect(screen.queryByText("Create Workflow")).not.toBeInTheDocument();
        }
      });

      unmount();
      localStorage.clear();
      jest.clearAllMocks();
    }
  });

  it("shows Run button only for editor+ roles", async () => {
    const roles = [
      { role: "admin", visible: true },
      { role: "editor", visible: true },
      { role: "viewer", visible: false },
    ];

    for (const { role, visible } of roles) {
      seedUser(role);
      const { unmount } = render(
        <AuthProvider>
          <ProtectedRoute requiredPermission="trigger">
            <button>Run</button>
          </ProtectedRoute>
        </AuthProvider>
      );

      await waitFor(() => {
        if (visible) {
          expect(screen.getByText("Run")).toBeInTheDocument();
        } else {
          expect(screen.queryByText("Run")).not.toBeInTheDocument();
        }
      });

      unmount();
      localStorage.clear();
      jest.clearAllMocks();
    }
  });

  it("shows Delete button only for admin role", async () => {
    const roles = [
      { role: "admin", visible: true },
      { role: "editor", visible: false },
      { role: "viewer", visible: false },
    ];

    for (const { role, visible } of roles) {
      seedUser(role);
      const { unmount } = render(
        <AuthProvider>
          <ProtectedRoute requiredPermission="delete">
            <button>Delete</button>
          </ProtectedRoute>
        </AuthProvider>
      );

      await waitFor(() => {
        if (visible) {
          expect(screen.getByText("Delete")).toBeInTheDocument();
        } else {
          expect(screen.queryByText("Delete")).not.toBeInTheDocument();
        }
      });

      unmount();
      localStorage.clear();
      jest.clearAllMocks();
    }
  });

  it("shows Edit button only for editor+ roles", async () => {
    const roles = [
      { role: "admin", visible: true },
      { role: "editor", visible: true },
      { role: "viewer", visible: false },
    ];

    for (const { role, visible } of roles) {
      seedUser(role);
      const { unmount } = render(
        <AuthProvider>
          <ProtectedRoute requiredPermission="edit">
            <button>Edit Workflow</button>
          </ProtectedRoute>
        </AuthProvider>
      );

      await waitFor(() => {
        if (visible) {
          expect(screen.getByText("Edit Workflow")).toBeInTheDocument();
        } else {
          expect(screen.queryByText("Edit Workflow")).not.toBeInTheDocument();
        }
      });

      unmount();
      localStorage.clear();
      jest.clearAllMocks();
    }
  });
});

// ─── Console Error Tests ───────────────────────────────────────────────
describe("UI/UX: No Console Errors During Permission Checks", () => {
  let consoleSpy: jest.SpyInstance;

  beforeEach(() => {
    jest.clearAllMocks();
    localStorage.clear();
    consoleSpy = jest.spyOn(console, "error").mockImplementation(() => {});
  });

  afterEach(() => {
    consoleSpy.mockRestore();
  });

  it("should not log errors when checking permissions for admin", async () => {
    seedUser("admin");
    render(
      <AuthProvider>
        <DashboardLayout>
          <ProtectedRoute requiredPermission="delete">
            <button>Delete</button>
          </ProtectedRoute>
        </DashboardLayout>
      </AuthProvider>
    );

    await screen.findByText("Delete");
    expect(consoleSpy).not.toHaveBeenCalled();
  });

  it("should not log errors when checking permissions for viewer", async () => {
    seedUser("viewer");
    render(
      <AuthProvider>
        <DashboardLayout>
          <ProtectedRoute requiredPermission="delete">
            <button>Delete</button>
          </ProtectedRoute>
        </DashboardLayout>
      </AuthProvider>
    );

    await waitFor(() => {
      expect(screen.queryByText("Delete")).not.toBeInTheDocument();
    });
    expect(consoleSpy).not.toHaveBeenCalled();
  });

  it("should not log errors when user is unauthenticated", async () => {
    render(
      <AuthProvider>
        <DashboardLayout>
          <div>Content</div>
        </DashboardLayout>
      </AuthProvider>
    );

    await waitFor(() => {
      expect(screen.queryByText("Content")).not.toBeInTheDocument();
    });
    expect(consoleSpy).not.toHaveBeenCalled();
  });

  it("should not log errors for unknown permission type", async () => {
    seedUser("admin");

    const TestComponent = () => {
      const { can } = useAuth();
      return <span>{can("unknown_permission") ? "yes" : "no"}</span>;
    };

    render(
      <AuthProvider>
        <DashboardLayout>
          <TestComponent />
        </DashboardLayout>
      </AuthProvider>
    );

    await screen.findByText("no");
    expect(consoleSpy).not.toHaveBeenCalled();
  });
});

// ─── Loading State Tests ───────────────────────────────────────────────
describe("UI/UX: Loading States", () => {
  beforeEach(() => {
    jest.clearAllMocks();
    localStorage.clear();
  });

  it("shows loading state while checking auth", () => {
    localStorage.setItem("token", "pending-token");
    // Don't resolve getMe yet
    mockApi.getMe.mockReturnValue(new Promise(() => {}));

    render(
      <AuthProvider>
        <DashboardLayout>
          <div>Protected Content</div>
        </DashboardLayout>
      </AuthProvider>
    );

    expect(screen.getByText("Loading...")).toBeInTheDocument();
    expect(screen.queryByText("Protected Content")).not.toBeInTheDocument();
  });

  it("hides loading state after auth resolves", async () => {
    seedUser("admin");

    render(
      <AuthProvider>
        <DashboardLayout>
          <div>Dashboard Ready</div>
        </DashboardLayout>
      </AuthProvider>
    );

    await screen.findByText("Dashboard Ready");
    expect(screen.queryByText("Loading...")).not.toBeInTheDocument();
  });

  it("shows nothing when auth fails and no stored token", async () => {
    render(
      <AuthProvider>
        <DashboardLayout>
          <div>Should Not Show</div>
        </DashboardLayout>
      </AuthProvider>
    );

    await waitFor(() => {
      expect(screen.queryByText("Should Not Show")).not.toBeInTheDocument();
      expect(screen.queryByText("Loading...")).not.toBeInTheDocument();
    });
  });

  it("shows loading when token exists but getMe has not resolved", () => {
    localStorage.setItem("token", "valid-token");
    mockApi.getMe.mockReturnValue(new Promise(() => {}));

    render(
      <AuthProvider>
        <DashboardLayout>
          <div>Content</div>
        </DashboardLayout>
      </AuthProvider>
    );

    expect(screen.getByText("Loading...")).toBeInTheDocument();
  });
});

// ─── Error Message Display Tests ───────────────────────────────────────
describe("UI/UX: Error Messages Display", () => {
  beforeEach(() => {
    jest.clearAllMocks();
    localStorage.clear();
  });

  it("clears user state when token is expired", async () => {
    localStorage.setItem("token", "expired-token");
    mockApi.getMe.mockRejectedValue(new Error("Unauthorized"));

    render(
      <AuthProvider>
        <DashboardLayout>
          <div>Content</div>
        </DashboardLayout>
      </AuthProvider>
    );

    await waitFor(() => {
      expect(screen.queryByText("Content")).not.toBeInTheDocument();
    });

    expect(localStorage.getItem("token")).toBeNull();
    expect(mockApi.clearToken).toHaveBeenCalled();
  });

  it("redirects to login when getMe fails", async () => {
    const mockPush = jest.fn();
    (require("next/navigation").useRouter as jest.Mock).mockReturnValue({
      push: mockPush,
    });

    localStorage.setItem("token", "bad-token");
    mockApi.getMe.mockRejectedValue(new Error("Server error"));

    render(
      <AuthProvider>
        <DashboardLayout>
          <div>Content</div>
        </DashboardLayout>
      </AuthProvider>
    );

    await waitFor(() => {
      expect(mockPush).toHaveBeenCalledWith("/login");
    });
  });

  it("clears stored token on auth failure", async () => {
    localStorage.setItem("token", "invalid");
    mockApi.getMe.mockRejectedValue(new Error("Unauthorized"));

    render(
      <AuthProvider>
        <div>Test</div>
      </AuthProvider>
    );

    await waitFor(() => {
      expect(localStorage.getItem("token")).toBeNull();
    });
  });
});

// ─── Layout & Structure Tests ──────────────────────────────────────────
describe("UI/UX: Layout Structure", () => {
  beforeEach(() => {
    jest.clearAllMocks();
    localStorage.clear();
  });

  it("renders header with correct structure", async () => {
    seedUser("admin");
    const { container } = render(
      <AuthProvider>
        <DashboardLayout>
          <div>X</div>
        </DashboardLayout>
      </AuthProvider>
    );

    await screen.findByText("FlowForge");
    const header = container.querySelector("header");
    expect(header).toBeInTheDocument();
  });

  it("renders nav with correct structure", async () => {
    seedUser("admin");
    const { container } = render(
      <AuthProvider>
        <DashboardLayout>
          <div>X</div>
        </DashboardLayout>
      </AuthProvider>
    );

    await screen.findByText("FlowForge");
    const nav = container.querySelector("nav");
    expect(nav).toBeInTheDocument();
  });

  it("renders main content area", async () => {
    seedUser("admin");
    const { container } = render(
      <AuthProvider>
        <DashboardLayout>
          <div>Child Content</div>
        </DashboardLayout>
      </AuthProvider>
    );

    await screen.findByText("Child Content");
    const main = container.querySelector("main");
    expect(main).toBeInTheDocument();
    expect(main?.textContent).toContain("Child Content");
  });

  it("applies min-h-screen class to root container", async () => {
    seedUser("admin");
    const { container } = render(
      <AuthProvider>
        <DashboardLayout>
          <div>X</div>
        </DashboardLayout>
      </AuthProvider>
    );

    await screen.findByText("FlowForge");
    const root = container.firstChild as HTMLElement;
    expect(root.className).toContain("min-h-screen");
  });

  it("applies bg-gray-50 background to root container", async () => {
    seedUser("admin");
    const { container } = render(
      <AuthProvider>
        <DashboardLayout>
          <div>X</div>
        </DashboardLayout>
      </AuthProvider>
    );

    await screen.findByText("FlowForge");
    const root = container.firstChild as HTMLElement;
    expect(root.className).toContain("bg-gray-50");
  });
});
