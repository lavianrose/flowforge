import React from 'react';
import { render } from '@testing-library/react';
import '@testing-library/jest-dom';

// Mock reactflow before importing CustomNode
jest.mock('reactflow', () => ({
  Handle: ({ type, position }: { type: string; position: string }) => (
    <div data-testid={`handle-${type}`} data-position={position} />
  ),
  Position: { Top: 'top', Bottom: 'bottom' },
}));

import CustomNode from '../nodes/CustomNode';

describe('CustomNode', () => {
  const baseProps = {
    id: 'node-1',
    data: {
      type: 'http',
      label: 'HTTP Request',
      config: {
        url: 'https://api.example.com/data',
        method: 'GET',
        headers: {},
      },
    },
    type: 'custom',
    selected: false,
    isConnectable: true,
    xPos: 0,
    yPos: 0,
    dragging: false,
    zIndex: 0,
  };

  it('should render node label', () => {
    const { container } = render(
      <CustomNode {...baseProps} />
    );
    expect(container.textContent).toContain('HTTP Request');
  });

  it('should render node type', () => {
    const { container } = render(
      <CustomNode {...baseProps} />
    );
    expect(container.textContent).toContain('http');
  });

  it('should show HTTP config summary with method and url', () => {
    const { container } = render(
      <CustomNode {...baseProps} />
    );
    expect(container.textContent).toContain('GET');
    expect(container.textContent).toContain('api.example.com');
  });

  it('should show delay config summary', () => {
    const { container } = render(
      <CustomNode
        {...baseProps}
        data={{
          type: 'delay',
          label: 'Wait',
          config: { seconds: 10 },
        }}
      />
    );
    expect(container.textContent).toContain('10s delay');
  });

  it('should show script config summary with code preview', () => {
    const { container } = render(
      <CustomNode
        {...baseProps}
        data={{
          type: 'script',
          label: 'Transform',
          config: { code: 'return {"key": "value"}' },
        }}
      />
    );
    expect(container.textContent).toContain('return {"key": "value"}');
  });

  it('should show condition config summary', () => {
    const { container } = render(
      <CustomNode
        {...baseProps}
        data={{
          type: 'condition',
          label: 'Check',
          config: { expression: '10 > 5' },
        }}
      />
    );
    expect(container.textContent).toContain('10 > 5');
  });

  it('should show (no code) when script code is empty', () => {
    const { container } = render(
      <CustomNode
        {...baseProps}
        data={{
          type: 'script',
          label: 'Script',
          config: { code: '' },
        }}
      />
    );
    expect(container.textContent).toContain('(no code)');
  });

  it('should show (no expression) when condition expression is empty', () => {
    const { container } = render(
      <CustomNode
        {...baseProps}
        data={{
          type: 'condition',
          label: 'Condition',
          config: { expression: '' },
        }}
      />
    );
    expect(container.textContent).toContain('(no expression)');
  });

  it('should show (no url) when HTTP url is empty', () => {
    const { container } = render(
      <CustomNode
        {...baseProps}
        data={{
          type: 'http',
          label: 'HTTP',
          config: { url: '', method: 'POST', headers: {} },
        }}
      />
    );
    expect(container.textContent).toContain('(no url)');
  });

  it('should show ring when selected', () => {
    const { container } = render(
      <CustomNode {...baseProps} selected={true} />
    );
    const nodeDiv = container.firstChild as HTMLElement;
    expect(nodeDiv.className).toContain('ring-2');
  });

  it('should not show ring when not selected', () => {
    const { container } = render(
      <CustomNode {...baseProps} selected={false} />
    );
    const nodeDiv = container.firstChild as HTMLElement;
    expect(nodeDiv.className).not.toContain('ring-2');
  });

  it('should render handles', () => {
    const { getByTestId } = render(
      <CustomNode {...baseProps} />
    );
    expect(getByTestId('handle-target')).toBeInTheDocument();
    expect(getByTestId('handle-source')).toBeInTheDocument();
  });
});
