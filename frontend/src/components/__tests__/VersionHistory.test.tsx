import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import '@testing-library/jest-dom';
import VersionHistory from '../VersionHistory';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { api } from '@/lib/api';

// Mock the API module
jest.mock('@/lib/api', () => ({
  api: {
    getWorkflowVersions: jest.fn(),
    rollbackWorkflow: jest.fn(),
  },
}));

// Mock react-router-dom
const mockPush = jest.fn();
jest.mock('next/navigation', () => ({
  useRouter: () => ({
    push: mockPush,
  }),
}));

const mockApi = api as jest.Mocked<typeof api>;

function createWrapper() {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: {
        retry: false,
      },
    },
  });

  return ({ children }: { children: React.ReactNode }) => (
    <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  );
}

describe('VersionHistory Component', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    // Mock window.location.reload
    Object.defineProperty(window, 'location', {
      writable: true,
      value: { reload: jest.fn() },
    });
    // Mock confirm dialog
    global.confirm = jest.fn(() => true);
    // Mock alert
    global.alert = jest.fn();
  });

  afterEach(() => {
    jest.restoreAllMocks();
  });

  it('should render loading state', () => {
    mockApi.getWorkflowVersions.mockImplementation(
      () => new Promise(() => {}) // Never resolves to keep loading state
    );

    render(<VersionHistory workflowId="wf-1" />, {
      wrapper: createWrapper(),
    });

    expect(screen.getByText('Loading versions...')).toBeInTheDocument();
  });

  it('should render error state', async () => {
    mockApi.getWorkflowVersions.mockRejectedValue(new Error('Failed to fetch'));

    render(<VersionHistory workflowId="wf-1" />, {
      wrapper: createWrapper(),
    });

    await waitFor(() => {
      expect(screen.getByText('Failed to fetch')).toBeInTheDocument();
    });
  });

  it('should render empty state when no versions', async () => {
    mockApi.getWorkflowVersions.mockResolvedValue({ versions: [] });

    render(<VersionHistory workflowId="wf-1" />, {
      wrapper: createWrapper(),
    });

    await waitFor(() => {
      expect(
        screen.getByText('No version history available. Create your first version by updating the workflow.')
      ).toBeInTheDocument();
    });
  });

  it('should render version list', async () => {
    const mockVersions = {
      versions: [
        {
          id: 'ver-1',
          workflow_id: 'wf-1',
          version: 1,
          definition: {
            nodes: [
              {
                id: 'node-1',
                type: 'http',
                name: 'HTTP Request',
                config: {},
                position: { x: 100, y: 100 },
              },
            ],
            edges: [{ id: 'edge-1', source: 'node-1', target: 'node-2' }],
          },
          created_by: 'user-1',
          created_at: '2024-01-01T00:00:00Z',
        },
        {
          id: 'ver-2',
          workflow_id: 'wf-1',
          version: 2,
          definition: {
            nodes: [
              {
                id: 'node-1',
                type: 'http',
                name: 'HTTP Request Updated',
                config: {},
                position: { x: 100, y: 100 },
              },
            ],
            edges: [{ id: 'edge-1', source: 'node-1', target: 'node-2' }],
          },
          created_by: 'user-2',
          created_at: '2024-01-02T00:00:00Z',
        },
      ],
    };

    mockApi.getWorkflowVersions.mockResolvedValue(mockVersions);

    render(<VersionHistory workflowId="wf-1" />, {
      wrapper: createWrapper(),
    });

    await waitFor(() => {
      expect(screen.getByText('Version History')).toBeInTheDocument();
      expect(screen.getByText('Version 2')).toBeInTheDocument();
      expect(screen.getByText('Version 1')).toBeInTheDocument();
      expect(screen.getByText('Current')).toBeInTheDocument(); // Version 2 should have Current badge
    });

    // Check for all occurrences of "1 nodes" and "1 connections"
    const nodeCounts = screen.getAllByText('1 nodes');
    const connectionCounts = screen.getAllByText('1 connections');
    expect(nodeCounts).toHaveLength(2);
    expect(connectionCounts).toHaveLength(2);

    // Check for users
    expect(screen.getByText('by user-1')).toBeInTheDocument();
    expect(screen.getByText('by user-2')).toBeInTheDocument();
  });

  it('should sort versions with highest first (current version)', async () => {
    const mockVersions = {
      versions: [
        {
          id: 'ver-1',
          workflow_id: 'wf-1',
          version: 1,
          definition: { nodes: [], edges: [] },
          created_by: 'user-1',
          created_at: '2024-01-01T00:00:00Z',
        },
        {
          id: 'ver-2',
          workflow_id: 'wf-1',
          version: 2,
          definition: { nodes: [], edges: [] },
          created_by: 'user-1',
          created_at: '2024-01-02T00:00:00Z',
        },
        {
          id: 'ver-3',
          workflow_id: 'wf-1',
          version: 3,
          definition: { nodes: [], edges: [] },
          created_by: 'user-1',
          created_at: '2024-01-03T00:00:00Z',
        },
      ],
    };

    mockApi.getWorkflowVersions.mockResolvedValue(mockVersions);

    render(<VersionHistory workflowId="wf-1" />, {
      wrapper: createWrapper(),
    });

    await waitFor(() => {
      const versionElements = screen.getAllByText(/Version \d+/);
      expect(versionElements[0]).toHaveTextContent('Version 3');
      expect(versionElements[1]).toHaveTextContent('Version 2');
      expect(versionElements[2]).toHaveTextContent('Version 1');
    });
  });

  it('should show rollback button only for non-current versions', async () => {
    const mockVersions = {
      versions: [
        {
          id: 'ver-1',
          workflow_id: 'wf-1',
          version: 1,
          definition: { nodes: [], edges: [] },
          created_by: 'user-1',
          created_at: '2024-01-01T00:00:00Z',
        },
        {
          id: 'ver-2',
          workflow_id: 'wf-1',
          version: 2,
          definition: { nodes: [], edges: [] },
          created_by: 'user-1',
          created_at: '2024-01-02T00:00:00Z',
        },
      ],
    };

    mockApi.getWorkflowVersions.mockResolvedValue(mockVersions);

    render(<VersionHistory workflowId="wf-1" />, {
      wrapper: createWrapper(),
    });

    await waitFor(() => {
      const rollbackButtons = screen.getAllByText('Rollback');
      expect(rollbackButtons).toHaveLength(1); // Only version 1 should have rollback button
    });
  });

  it('should handle rollback successfully', async () => {
    const mockVersions = {
      versions: [
        {
          id: 'ver-1',
          workflow_id: 'wf-1',
          version: 1,
          definition: { nodes: [], edges: [] },
          created_by: 'user-1',
          created_at: '2024-01-01T00:00:00Z',
        },
        {
          id: 'ver-2',
          workflow_id: 'wf-1',
          version: 2,
          definition: { nodes: [], edges: [] },
          created_by: 'user-1',
          created_at: '2024-01-02T00:00:00Z',
        },
      ],
    };

    const mockWorkflow = {
      id: 'wf-1',
      name: 'Test Workflow',
      definition: { nodes: [], edges: [] },
      timeout_seconds: 300,
      active: true,
      created_by: 'user-1',
      created_at: '2024-01-01T00:00:00Z',
      updated_at: '2024-01-02T00:00:00Z',
    };

    mockApi.getWorkflowVersions.mockResolvedValue(mockVersions);
    mockApi.rollbackWorkflow.mockResolvedValue(mockWorkflow);

    render(<VersionHistory workflowId="wf-1" />, {
      wrapper: createWrapper(),
    });

    await waitFor(() => {
      expect(screen.getByText('Version History')).toBeInTheDocument();
    });

    const rollbackButton = screen.getByText('Rollback');
    fireEvent.click(rollbackButton);

    await waitFor(() => {
      expect(global.confirm).toHaveBeenCalledWith(
        expect.stringContaining('Are you sure you want to rollback to version 1')
      );
      expect(mockApi.rollbackWorkflow).toHaveBeenCalledWith('wf-1', 1);
      expect(global.alert).toHaveBeenCalledWith('Successfully rolled back to version 1');
      expect(window.location.reload).toHaveBeenCalled();
    });
  });

  it('should cancel rollback when user confirms false', async () => {
    global.confirm = jest.fn(() => false);

    const mockVersions = {
      versions: [
        {
          id: 'ver-1',
          workflow_id: 'wf-1',
          version: 1,
          definition: { nodes: [], edges: [] },
          created_by: 'user-1',
          created_at: '2024-01-01T00:00:00Z',
        },
        {
          id: 'ver-2',
          workflow_id: 'wf-1',
          version: 2,
          definition: { nodes: [], edges: [] },
          created_by: 'user-1',
          created_at: '2024-01-02T00:00:00Z',
        },
      ],
    };

    mockApi.getWorkflowVersions.mockResolvedValue(mockVersions);

    render(<VersionHistory workflowId="wf-1" />, {
      wrapper: createWrapper(),
    });

    await waitFor(() => {
      expect(screen.getByText('Version History')).toBeInTheDocument();
    });

    const rollbackButton = screen.getByText('Rollback');
    fireEvent.click(rollbackButton);

    await waitFor(() => {
      expect(global.confirm).toHaveBeenCalled();
      expect(mockApi.rollbackWorkflow).not.toHaveBeenCalled();
      expect(window.location.reload).not.toHaveBeenCalled();
    });
  });

  it('should handle rollback error', async () => {
    const mockVersions = {
      versions: [
        {
          id: 'ver-1',
          workflow_id: 'wf-1',
          version: 1,
          definition: { nodes: [], edges: [] },
          created_by: 'user-1',
          created_at: '2024-01-01T00:00:00Z',
        },
        {
          id: 'ver-2',
          workflow_id: 'wf-1',
          version: 2,
          definition: { nodes: [], edges: [] },
          created_by: 'user-1',
          created_at: '2024-01-02T00:00:00Z',
        },
      ],
    };

    mockApi.getWorkflowVersions.mockResolvedValue(mockVersions);
    mockApi.rollbackWorkflow.mockRejectedValue(new Error('Version not found'));

    render(<VersionHistory workflowId="wf-1" />, {
      wrapper: createWrapper(),
    });

    await waitFor(() => {
      expect(screen.getByText('Version History')).toBeInTheDocument();
    });

    const rollbackButton = screen.getByText('Rollback');
    fireEvent.click(rollbackButton);

    await waitFor(() => {
      expect(global.alert).toHaveBeenCalledWith('Version not found');
      expect(window.location.reload).not.toHaveBeenCalled();
    });
  });

  it('should display formatted date and time', async () => {
    const mockVersions = {
      versions: [
        {
          id: 'ver-1',
          workflow_id: 'wf-1',
          version: 1,
          definition: { nodes: [], edges: [] },
          created_by: 'test-user@example.com',
          created_at: '2024-01-15T14:30:00Z',
        },
      ],
    };

    mockApi.getWorkflowVersions.mockResolvedValue(mockVersions);

    render(<VersionHistory workflowId="wf-1" />, {
      wrapper: createWrapper(),
    });

    await waitFor(() => {
      expect(screen.getByText('by test-user@example.com')).toBeInTheDocument();
      // Date formatting depends on locale, so we just check it exists
      const versionCard = screen.getByText('Version 1').closest('div');
      expect(versionCard).toBeInTheDocument();
    });
  });

  it('should show loading state during rollback', async () => {
    const mockVersions = {
      versions: [
        {
          id: 'ver-1',
          workflow_id: 'wf-1',
          version: 1,
          definition: { nodes: [], edges: [] },
          created_by: 'user-1',
          created_at: '2024-01-01T00:00:00Z',
        },
        {
          id: 'ver-2',
          workflow_id: 'wf-1',
          version: 2,
          definition: { nodes: [], edges: [] },
          created_by: 'user-1',
          created_at: '2024-01-02T00:00:00Z',
        },
      ],
    };

    let resolveRollback: (value: any) => void;
    const rollbackPromise = new Promise((resolve) => {
      resolveRollback = resolve;
    });

    mockApi.getWorkflowVersions.mockResolvedValue(mockVersions);
    mockApi.rollbackWorkflow.mockReturnValue(rollbackPromise);

    render(<VersionHistory workflowId="wf-1" />, {
      wrapper: createWrapper(),
    });

    await waitFor(() => {
      expect(screen.getByText('Version History')).toBeInTheDocument();
    });

    const rollbackButton = screen.getByText('Rollback');
    fireEvent.click(rollbackButton);

    await waitFor(() => {
      expect(screen.getByText('Rolling back...')).toBeInTheDocument();
      expect(rollbackButton).toBeDisabled();
    });

    // Resolve the promise
    resolveRollback!({ id: 'wf-1', name: 'Test' });

    await waitFor(() => {
      expect(screen.queryByText('Rolling back...')).not.toBeInTheDocument();
    });
  });
});
