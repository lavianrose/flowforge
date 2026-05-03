import { fireEvent, render, screen } from "@testing-library/react";
import "@testing-library/jest-dom";

// Mock reactflow
jest.mock("reactflow", () => ({
  Handle: () => <div />,
  Position: { Top: "top", Bottom: "bottom" },
}));

import NodeConfigPanel from "../nodes/NodeConfigPanel";

describe("NodeConfigPanel", () => {
  const mockOnConfigChange = jest.fn();
  const mockOnLabelChange = jest.fn();
  const mockOnClose = jest.fn();

  const defaultProps = {
    nodeId: "node-1",
    nodeType: "http",
    nodeLabel: "HTTP Request",
    config: { url: "https://api.example.com", method: "GET", headers: {} },
    onConfigChange: mockOnConfigChange,
    onLabelChange: mockOnLabelChange,
    onClose: mockOnClose,
  };

  beforeEach(() => {
    jest.clearAllMocks();
  });

  it("should render node configuration header", () => {
    render(<NodeConfigPanel {...defaultProps} />);
    expect(screen.getByText("Node Configuration")).toBeInTheDocument();
  });

  it("should render label input with current value", () => {
    render(<NodeConfigPanel {...defaultProps} />);
    const labelInput = screen.getByDisplayValue("HTTP Request");
    expect(labelInput).toBeInTheDocument();
  });

  it("should call onLabelChange when label is edited", () => {
    render(<NodeConfigPanel {...defaultProps} />);
    const labelInput = screen.getByDisplayValue("HTTP Request");
    fireEvent.change(labelInput, { target: { value: "New Label" } });
    expect(mockOnLabelChange).toHaveBeenCalledWith("node-1", "New Label");
  });

  it("should call onClose when close button is clicked", () => {
    render(<NodeConfigPanel {...defaultProps} />);
    const closeButton = screen.getByText("\u00d7");
    fireEvent.click(closeButton);
    expect(mockOnClose).toHaveBeenCalled();
  });

  // ---- HTTP Config ----
  describe("HTTP Config", () => {
    it("should render URL input", () => {
      render(<NodeConfigPanel {...defaultProps} />);
      expect(
        screen.getByDisplayValue("https://api.example.com")
      ).toBeInTheDocument();
    });

    it("should render method select", () => {
      render(<NodeConfigPanel {...defaultProps} />);
      expect(screen.getByDisplayValue("GET")).toBeInTheDocument();
    });

    it("should call onConfigChange when URL changes", () => {
      render(<NodeConfigPanel {...defaultProps} />);
      const urlInput = screen.getByDisplayValue("https://api.example.com");
      fireEvent.change(urlInput, { target: { value: "https://newurl.com" } });
      expect(mockOnConfigChange).toHaveBeenCalledWith(
        "node-1",
        expect.objectContaining({
          url: "https://newurl.com",
        })
      );
    });

    it("should call onConfigChange when method changes", () => {
      render(<NodeConfigPanel {...defaultProps} />);
      const methodSelect = screen.getByDisplayValue("GET");
      fireEvent.change(methodSelect, { target: { value: "POST" } });
      expect(mockOnConfigChange).toHaveBeenCalledWith(
        "node-1",
        expect.objectContaining({
          method: "POST",
        })
      );
    });

    it("should render body textarea", () => {
      render(<NodeConfigPanel {...defaultProps} />);
      expect(
        screen.getByPlaceholderText('{"key": "value"}')
      ).toBeInTheDocument();
    });
  });

  // ---- Delay Config ----
  describe("Delay Config", () => {
    it("should render seconds input", () => {
      render(
        <NodeConfigPanel
          {...defaultProps}
          config={{ seconds: 10 }}
          nodeType="delay"
        />
      );
      expect(screen.getByDisplayValue("10")).toBeInTheDocument();
    });

    it("should call onConfigChange when seconds changes", () => {
      render(
        <NodeConfigPanel
          {...defaultProps}
          config={{ seconds: 5 }}
          nodeType="delay"
        />
      );
      const secondsInput = screen.getByDisplayValue("5");
      fireEvent.change(secondsInput, { target: { value: "20" } });
      expect(mockOnConfigChange).toHaveBeenCalledWith(
        "node-1",
        expect.objectContaining({
          seconds: 20,
        })
      );
    });
  });

  // ---- Script Config ----
  describe("Script Config", () => {
    it("should render code textarea", () => {
      render(
        <NodeConfigPanel
          {...defaultProps}
          config={{ code: 'return {"ok": true}' }}
          nodeType="script"
        />
      );
      expect(
        screen.getByDisplayValue('return {"ok": true}')
      ).toBeInTheDocument();
    });

    it("should call onConfigChange when code changes", () => {
      render(
        <NodeConfigPanel
          {...defaultProps}
          config={{ code: "" }}
          nodeType="script"
        />
      );
      const codeArea = screen.getByPlaceholderText(
        'return {"result": {{inputs.node_id.field}}}'
      );
      fireEvent.change(codeArea, { target: { value: 'return {"x": 1}' } });
      expect(mockOnConfigChange).toHaveBeenCalledWith(
        "node-1",
        expect.objectContaining({
          code: 'return {"x": 1}',
        })
      );
    });
  });

  // ---- Condition Config ----
  describe("Condition Config", () => {
    it("should render expression input", () => {
      render(
        <NodeConfigPanel
          {...defaultProps}
          config={{ expression: "{{inputs.node1.status_code}} == 200" }}
          nodeType="condition"
        />
      );
      expect(
        screen.getByDisplayValue("{{inputs.node1.status_code}} == 200")
      ).toBeInTheDocument();
    });

    it("should call onConfigChange when expression changes", () => {
      render(
        <NodeConfigPanel
          {...defaultProps}
          config={{ expression: "" }}
          nodeType="condition"
        />
      );
      const exprInput = screen.getByPlaceholderText(
        "{{inputs.node1.status_code}} == 200"
      );
      fireEvent.change(exprInput, { target: { value: "true" } });
      expect(mockOnConfigChange).toHaveBeenCalledWith(
        "node-1",
        expect.objectContaining({
          expression: "true",
        })
      );
    });

    it("should render example expressions", () => {
      render(
        <NodeConfigPanel
          {...defaultProps}
          config={{ expression: "" }}
          nodeType="condition"
        />
      );
      expect(screen.getByText("Examples:")).toBeInTheDocument();
    });
  });
});
