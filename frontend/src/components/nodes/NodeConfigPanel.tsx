'use client';

import { useState, useEffect } from 'react';

interface NodeConfig {
  [key: string]: unknown;
}

interface NodeConfigPanelProps {
  nodeId: string;
  nodeType: string;
  nodeLabel: string;
  config: NodeConfig;
  onConfigChange: (nodeId: string, config: NodeConfig) => void;
  onLabelChange: (nodeId: string, label: string) => void;
  onClose: () => void;
}

export default function NodeConfigPanel({
  nodeId,
  nodeType,
  nodeLabel,
  config,
  onConfigChange,
  onLabelChange,
  onClose,
}: NodeConfigPanelProps) {
  const [localConfig, setLocalConfig] = useState<NodeConfig>({ ...config });
  const [localLabel, setLocalLabel] = useState(nodeLabel);

  // Reset local state when node changes
  useEffect(() => {
    setLocalConfig({ ...config });
    setLocalLabel(nodeLabel);
  }, [nodeId, nodeType]);

  const updateConfig = (key: string, value: unknown) => {
    const updated = { ...localConfig, [key]: value };
    setLocalConfig(updated);
    onConfigChange(nodeId, updated);
  };

  const handleLabelChange = (value: string) => {
    setLocalLabel(value);
    onLabelChange(nodeId, value);
  };

  const typeLabel: Record<string, string> = {
    http: 'HTTP Request',
    delay: 'Delay',
    script: 'Script',
    condition: 'Condition',
  };

  const typeColor: Record<string, string> = {
    http: 'text-blue-700',
    delay: 'text-yellow-700',
    script: 'text-purple-700',
    condition: 'text-green-700',
  };

  return (
    <div className="bg-white rounded-lg shadow p-4 overflow-y-auto h-full">
      <div className="flex items-center justify-between mb-4">
        <h3 className="font-semibold text-sm">Node Configuration</h3>
        <button
          onClick={onClose}
          className="text-gray-400 hover:text-gray-600 text-lg leading-none"
        >
          &times;
        </button>
      </div>

      <div className="mb-4">
        <label className="block text-xs font-medium text-gray-700 mb-1">
          Label
        </label>
        <input
          type="text"
          value={localLabel}
          onChange={(e) => handleLabelChange(e.target.value)}
          className="w-full px-2 py-1.5 border border-gray-300 rounded text-sm focus:outline-none focus:ring-1 focus:ring-indigo-500 focus:border-indigo-500"
        />
      </div>

      <div className="mb-3">
        <span className={`text-xs font-medium ${typeColor[nodeType] || 'text-gray-600'}`}>
          Type: {typeLabel[nodeType] || nodeType}
        </span>
      </div>

      {nodeType === 'http' && (
        <HttpConfig config={localConfig} onChange={updateConfig} />
      )}
      {nodeType === 'delay' && (
        <DelayConfig config={localConfig} onChange={updateConfig} />
      )}
      {nodeType === 'script' && (
        <ScriptConfig config={localConfig} onChange={updateConfig} />
      )}
      {nodeType === 'condition' && (
        <ConditionConfig config={localConfig} onChange={updateConfig} />
      )}
    </div>
  );
}

// ---- HTTP Config ----
function HttpConfig({
  config,
  onChange,
}: {
  config: NodeConfig;
  onChange: (key: string, value: unknown) => void;
}) {
  const [headerKey, setHeaderKey] = useState('');
  const [headerVal, setHeaderVal] = useState('');

  const headers = (config.headers || {}) as Record<string, string>;

  const addHeader = () => {
    if (!headerKey.trim()) return;
    const updated = { ...headers, [headerKey.trim()]: headerVal };
    onChange('headers', updated);
    setHeaderKey('');
    setHeaderVal('');
  };

  const removeHeader = (key: string) => {
    const updated = { ...headers };
    delete updated[key];
    onChange('headers', updated);
  };

  return (
    <div className="space-y-3">
      <div>
        <label className="block text-xs font-medium text-gray-700 mb-1">
          URL *
        </label>
        <input
          type="text"
          value={(config.url as string) || ''}
          onChange={(e) => onChange('url', e.target.value)}
          className="w-full px-2 py-1.5 border border-gray-300 rounded text-sm font-mono focus:outline-none focus:ring-1 focus:ring-indigo-500 focus:border-indigo-500"
          placeholder="https://api.example.com/data"
        />
      </div>

      <div>
        <label className="block text-xs font-medium text-gray-700 mb-1">
          Method *
        </label>
        <select
          value={(config.method as string) || 'GET'}
          onChange={(e) => onChange('method', e.target.value)}
          className="w-full px-2 py-1.5 border border-gray-300 rounded text-sm focus:outline-none focus:ring-1 focus:ring-indigo-500 focus:border-indigo-500"
        >
          <option value="GET">GET</option>
          <option value="POST">POST</option>
          <option value="PUT">PUT</option>
          <option value="PATCH">PATCH</option>
          <option value="DELETE">DELETE</option>
        </select>
      </div>

      <div>
        <label className="block text-xs font-medium text-gray-700 mb-1">
          Body (JSON)
        </label>
        <textarea
          value={typeof config.body === 'string' ? (config.body as string) : config.body ? JSON.stringify(config.body, null, 2) : ''}
          onChange={(e) => onChange('body', e.target.value)}
          className="w-full px-2 py-1.5 border border-gray-300 rounded text-sm font-mono focus:outline-none focus:ring-1 focus:ring-indigo-500 focus:border-indigo-500"
          rows={4}
          placeholder='{"key": "value"}'
        />
        <p className="text-xs text-gray-400 mt-1">
          Supports template: {'{{inputs.node_id.field}}'}
        </p>
      </div>

      <div>
        <label className="block text-xs font-medium text-gray-700 mb-1">
          Headers
        </label>
        {Object.entries(headers).map(([key, val]) => (
          <div key={key} className="flex items-center gap-1 mb-1">
            <span className="text-xs bg-gray-100 px-2 py-0.5 rounded flex-1 truncate">
              {key}: {val}
            </span>
            <button
              onClick={() => removeHeader(key)}
              className="text-red-400 hover:text-red-600 text-xs"
            >
              &times;
            </button>
          </div>
        ))}
        <div className="flex gap-1 mt-1">
          <input
            type="text"
            value={headerKey}
            onChange={(e) => setHeaderKey(e.target.value)}
            placeholder="Key"
            className="flex-1 px-2 py-1 border border-gray-300 rounded text-xs focus:outline-none focus:ring-1 focus:ring-indigo-500"
          />
          <input
            type="text"
            value={headerVal}
            onChange={(e) => setHeaderVal(e.target.value)}
            placeholder="Value"
            className="flex-1 px-2 py-1 border border-gray-300 rounded text-xs focus:outline-none focus:ring-1 focus:ring-indigo-500"
          />
          <button
            onClick={addHeader}
            className="px-2 py-1 bg-blue-500 text-white rounded text-xs hover:bg-blue-600"
          >
            +
          </button>
        </div>
      </div>
    </div>
  );
}

// ---- Delay Config ----
function DelayConfig({
  config,
  onChange,
}: {
  config: NodeConfig;
  onChange: (key: string, value: unknown) => void;
}) {
  return (
    <div className="space-y-3">
      <div>
        <label className="block text-xs font-medium text-gray-700 mb-1">
          Delay (seconds) *
        </label>
        <input
          type="number"
          value={(config.seconds as number) || 5}
          onChange={(e) => onChange('seconds', parseInt(e.target.value, 10) || 0)}
          min={0}
          className="w-full px-2 py-1.5 border border-gray-300 rounded text-sm focus:outline-none focus:ring-1 focus:ring-indigo-500 focus:border-indigo-500"
        />
      </div>
    </div>
  );
}

// ---- Script Config ----
function ScriptConfig({
  config,
  onChange,
}: {
  config: NodeConfig;
  onChange: (key: string, value: unknown) => void;
}) {
  return (
    <div className="space-y-3">
      <div>
        <label className="block text-xs font-medium text-gray-700 mb-1">
          Code *
        </label>
        <textarea
          value={(config.code as string) || ''}
          onChange={(e) => onChange('code', e.target.value)}
          className="w-full px-2 py-1.5 border border-gray-300 rounded text-sm font-mono focus:outline-none focus:ring-1 focus:ring-indigo-500 focus:border-indigo-500"
          rows={8}
          placeholder={'return {"result": {{inputs.node_id.field}}}'}
        />
        <p className="text-xs text-gray-400 mt-1">
          Supports: <code>return {`{"key": "value"}`}</code> with template variables {'{{inputs.node_id.field}}'}
        </p>
      </div>
    </div>
  );
}

// ---- Condition Config ----
function ConditionConfig({
  config,
  onChange,
}: {
  config: NodeConfig;
  onChange: (key: string, value: unknown) => void;
}) {
  return (
    <div className="space-y-3">
      <div>
        <label className="block text-xs font-medium text-gray-700 mb-1">
          Expression *
        </label>
        <input
          type="text"
          value={(config.expression as string) || ''}
          onChange={(e) => onChange('expression', e.target.value)}
          className="w-full px-2 py-1.5 border border-gray-300 rounded text-sm font-mono focus:outline-none focus:ring-1 focus:ring-indigo-500 focus:border-indigo-500"
          placeholder='{{inputs.node1.status_code}} == 200'
        />
        <p className="text-xs text-gray-400 mt-1">
          Supports: ==, !=, &gt;, &lt;, &gt;=, &lt;= operators
        </p>
      </div>
      <div className="text-xs text-gray-500 bg-gray-50 p-2 rounded">
        <p className="font-medium mb-1">Examples:</p>
        <ul className="space-y-0.5 list-disc list-inside">
          <li>{`{{inputs.node1.status_code}} == 200`}</li>
          <li>{`{{inputs.node1.json.count}} > 10`}</li>
          <li>{`{{inputs.node1.body}} != "error"`}</li>
          <li>true</li>
        </ul>
      </div>
    </div>
  );
}
