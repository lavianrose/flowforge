import React from 'react';
import { Handle, Position, NodeProps } from 'reactflow';
import 'reactflow/dist/style.css';

const nodeStyles: Record<string, string> = {
  http: 'bg-blue-100 border-blue-500',
  delay: 'bg-yellow-100 border-yellow-500',
  script: 'bg-purple-100 border-purple-500',
  condition: 'bg-green-100 border-green-500',
};

const iconMap: Record<string, string> = {
  http: '🌐',
  delay: '⏱️',
  script: '📜',
  condition: '❓',
};

export default function CustomNode({ data, selected }: NodeProps) {
  const nodeType = data.type as string;
  const bgColor = nodeStyles[nodeType] || 'bg-gray-100 border-gray-500';
  const icon = iconMap[nodeType] || '⚙️';

  return (
    <div
      className={`px-4 py-2 rounded-lg border-2 ${bgColor} ${
        selected ? 'ring-2 ring-indigo-500' : ''
      }`}
      style={{ minWidth: '150px' }}
    >
      <Handle type="target" position={Position.Top} className="w-3 h-3" />
      <div className="flex items-center gap-2">
        <span className="text-xl">{icon}</span>
        <div className="flex-1">
          <div className="font-semibold text-sm">{data.label}</div>
          <div className="text-xs text-gray-600">{nodeType}</div>
        </div>
      </div>
      <Handle type="source" position={Position.Bottom} className="w-3 h-3" />
    </div>
  );
}
