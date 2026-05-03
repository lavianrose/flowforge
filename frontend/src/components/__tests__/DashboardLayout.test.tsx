import { render, screen, waitFor } from "@testing-library/react";
import DashboardLayout from "@/components/DashboardLayout";
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

describe("DashboardLayout - RBAC", () => {
  beforeEach(() => {
    jest.clearAllMocks();
    localStorage.clear();
  });

  describe("Role badge display", () => {
    it("should display Admin badge", async () => {
      seedUser("admin");

      render(
        <AuthProvider>
          <DashboardLayout>
            <div>Content</div>
          </DashboardLayout>
        </AuthProvider>
      );

      await waitFor(() => {
        expect(screen.getByText("Admin")).toBeInTheDocument();
      });
    });

    it("should display Editor badge", async () => {
      seedUser("editor");

      render(
        <AuthProvider>
          <DashboardLayout>
            <div>Content</div>
          </DashboardLayout>
        </AuthProvider>
      );

      await waitFor(() => {
        expect(screen.getByText("Editor")).toBeInTheDocument();
      });
    });

    it("should display Viewer badge", async () => {
      seedUser("viewer");

      render(
        <AuthProvider>
          <DashboardLayout>
            <div>Content</div>
          </DashboardLayout>
        </AuthProvider>
      );

      await waitFor(() => {
        expect(screen.getByText("Viewer")).toBeInTheDocument();
      });
    });
  });

  describe("User info display", () => {
    it("should show user email in header", async () => {
      seedUser("admin");

      render(
        <AuthProvider>
          <DashboardLayout>
            <div>Content</div>
          </DashboardLayout>
        </AuthProvider>
      );

      await waitFor(() => {
        expect(screen.getByText("admin@flowforge.local")).toBeInTheDocument();
      });
    });
  });

  describe("Navigation", () => {
    it("should show Workflows link for all roles", async () => {
      seedUser("viewer");

      render(
        <AuthProvider>
          <DashboardLayout>
            <div>Content</div>
          </DashboardLayout>
        </AuthProvider>
      );

      await waitFor(() => {
        expect(screen.getByText("Workflows")).toBeInTheDocument();
      });
    });

    it("should show Runs link for all roles", async () => {
      seedUser("viewer");

      render(
        <AuthProvider>
          <DashboardLayout>
            <div>Content</div>
          </DashboardLayout>
        </AuthProvider>
      );

      await waitFor(() => {
        expect(screen.getByText("Runs")).toBeInTheDocument();
      });
    });
  });

  describe("Logout", () => {
    it("should show logout button", async () => {
      seedUser("admin");

      render(
        <AuthProvider>
          <DashboardLayout>
            <div>Content</div>
          </DashboardLayout>
        </AuthProvider>
      );

      await waitFor(() => {
        expect(screen.getByText("Logout")).toBeInTheDocument();
      });
    });

    it("should clear session on logout click", async () => {
      seedUser("admin");

      render(
        <AuthProvider>
          <DashboardLayout>
            <div>Content</div>
          </DashboardLayout>
        </AuthProvider>
      );

      await waitFor(() => {
        expect(screen.getByText("Logout")).toBeInTheDocument();
      });

      screen.getByText("Logout").click();

      expect(localStorage.getItem("token")).toBeNull();
      expect(mockApi.clearToken).toHaveBeenCalled();
    });
  });

  describe("Auth protection", () => {
    it("should redirect to login when not authenticated", async () => {
      const mockPush = jest.fn();
      (require("next/navigation").useRouter as jest.Mock).mockReturnValue({
        push: mockPush,
      });

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

    it("should not render children while checking auth", () => {
      render(
        <AuthProvider>
          <DashboardLayout>
            <div>Dashboard Content</div>
          </DashboardLayout>
        </AuthProvider>
      );

      // Children should not be visible while loading/unauthenticated
      expect(screen.queryByText("Dashboard Content")).not.toBeInTheDocument();
    });

    it("should render children when authenticated", async () => {
      seedUser("admin");

      render(
        <AuthProvider>
          <DashboardLayout>
            <div>Dashboard Content</div>
          </DashboardLayout>
        </AuthProvider>
      );

      await waitFor(() => {
        expect(screen.getByText("Dashboard Content")).toBeInTheDocument();
      });
    });
  });

  describe("Brand", () => {
    it("should display FlowForge brand", async () => {
      seedUser("admin");

      render(
        <AuthProvider>
          <DashboardLayout>
            <div>Content</div>
          </DashboardLayout>
        </AuthProvider>
      );

      await waitFor(() => {
        expect(screen.getByText("FlowForge")).toBeInTheDocument();
      });
    });
  });
});
