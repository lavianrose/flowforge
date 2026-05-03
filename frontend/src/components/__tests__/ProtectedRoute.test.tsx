import { render, screen, waitFor } from "@testing-library/react";
import ProtectedRoute from "@/components/ProtectedRoute";
import { api } from "@/lib/api";
import { AuthProvider } from "@/lib/auth";

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

function renderWithAuth(ui: React.ReactElement) {
  return render(<AuthProvider>{ui}</AuthProvider>);
}

function seedUser(role: string) {
  const user = {
    id: `user-${role}`,
    email: `${role}@flowforge.local`,
    role,
    tenant_id: "tenant-1",
  };
  localStorage.setItem("token", `${role}-token`);
  mockApi.getMe.mockResolvedValue(user);
  return user;
}

describe("ProtectedRoute", () => {
  beforeEach(() => {
    jest.clearAllMocks();
    localStorage.clear();
  });

  describe("Unauthenticated access", () => {
    it("should redirect to login when not authenticated", async () => {
      renderWithAuth(
        <ProtectedRoute>
          <div>Secret</div>
        </ProtectedRoute>
      );

      await waitFor(() => {
        expect(screen.queryByText("Secret")).not.toBeInTheDocument();
      });
    });
  });

  describe("Permission checks", () => {
    it("should show content when user has required permission", async () => {
      seedUser("admin");

      renderWithAuth(
        <ProtectedRoute requiredPermission="delete">
          <div>Delete Button</div>
        </ProtectedRoute>
      );

      await waitFor(() => {
        expect(screen.getByText("Delete Button")).toBeInTheDocument();
      });
    });

    it("should hide content when user lacks required permission", async () => {
      seedUser("viewer");

      renderWithAuth(
        <ProtectedRoute requiredPermission="delete">
          <div>Delete Button</div>
        </ProtectedRoute>
      );

      await waitFor(() => {
        expect(screen.queryByText("Delete Button")).not.toBeInTheDocument();
      });
    });

    it("should show content without requiredPermission for any authenticated user", async () => {
      seedUser("viewer");

      renderWithAuth(
        <ProtectedRoute>
          <div>Any Auth Content</div>
        </ProtectedRoute>
      );

      await waitFor(() => {
        expect(screen.getByText("Any Auth Content")).toBeInTheDocument();
      });
    });
  });

  describe("Role-based access matrix", () => {
    const cases = [
      { permission: "create", admin: true, editor: true, viewer: false },
      { permission: "edit", admin: true, editor: true, viewer: false },
      { permission: "trigger", admin: true, editor: true, viewer: false },
      { permission: "rollback", admin: true, editor: true, viewer: false },
      { permission: "delete", admin: true, editor: false, viewer: false },
      { permission: "view", admin: true, editor: true, viewer: true },
    ];

    const roles = ["admin", "editor", "viewer"] as const;

    cases.forEach(({ permission, admin, editor, viewer }) => {
      const expected = { admin, editor, viewer };
      roles.forEach((role) => {
        it(`should ${expected[role] ? "show" : "hide"} "${permission}" content for ${role}`, async () => {
          seedUser(role);

          renderWithAuth(
            <ProtectedRoute requiredPermission={permission}>
              <div>{permission} content</div>
            </ProtectedRoute>
          );

          await waitFor(() => {
            if (expected[role]) {
              expect(
                screen.getByText(`${permission} content`)
              ).toBeInTheDocument();
            } else {
              expect(
                screen.queryByText(`${permission} content`)
              ).not.toBeInTheDocument();
            }
          });
        });
      });
    });
  });

  describe("Route protection", () => {
    it("should block viewer from create workflow page", async () => {
      seedUser("viewer");

      renderWithAuth(
        <ProtectedRoute requiredPermission="create">
          <div>New Workflow Page</div>
        </ProtectedRoute>
      );

      await waitFor(() => {
        expect(screen.queryByText("New Workflow Page")).not.toBeInTheDocument();
      });
    });

    it("should allow editor to create workflow page", async () => {
      seedUser("editor");

      renderWithAuth(
        <ProtectedRoute requiredPermission="create">
          <div>New Workflow Page</div>
        </ProtectedRoute>
      );

      await waitFor(() => {
        expect(screen.getByText("New Workflow Page")).toBeInTheDocument();
      });
    });

    it("should block editor from delete actions", async () => {
      seedUser("editor");

      renderWithAuth(
        <ProtectedRoute requiredPermission="delete">
          <div>Delete Workflow</div>
        </ProtectedRoute>
      );

      await waitFor(() => {
        expect(screen.queryByText("Delete Workflow")).not.toBeInTheDocument();
      });
    });

    it("should allow admin full access", async () => {
      seedUser("admin");

      renderWithAuth(
        <ProtectedRoute requiredPermission="delete">
          <div>Delete Workflow</div>
        </ProtectedRoute>
      );

      await waitFor(() => {
        expect(screen.getByText("Delete Workflow")).toBeInTheDocument();
      });
    });
  });

  describe("Error handling", () => {
    it("should handle API errors gracefully", async () => {
      localStorage.setItem("token", "bad-token");
      mockApi.getMe.mockRejectedValue(new Error("Unauthorized"));

      renderWithAuth(
        <ProtectedRoute>
          <div>Protected</div>
        </ProtectedRoute>
      );

      await waitFor(() => {
        expect(screen.queryByText("Protected")).not.toBeInTheDocument();
      });
    });
  });
});
