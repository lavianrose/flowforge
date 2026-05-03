"use client";

import { useEffect, useState } from "react";

interface NodeConfig {
  [key: string]: unknown;
}

interface NodeConfigPanelProps {
  config: NodeConfig;
  nodeId: string;
  nodeLabel: string;
  nodeType: string;
  onClose: () => void;
  onConfigChange: (nodeId: string, config: NodeConfig) => void;
  onLabelChange: (nodeId: string, label: string) => void;
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
    http: "HTTP Request",
    delay: "Delay",
    script: "Script",
    condition: "Condition",
  };

  const typeColor: Record<string, string> = {
    http: "text-blue-700",
    delay: "text-yellow-700",
    script: "text-purple-700",
    condition: "text-green-700",
  };

  return (
    <div className="h-full overflow-y-auto rounded-lg bg-white p-4 shadow">
      <div className="mb-4 flex items-center justify-between">
        <h3 className="font-semibold text-sm">Node Configuration</h3>
        <button
          className="text-gray-400 text-lg leading-none hover:text-gray-600"
          onClick={onClose}
        >
          &times;
        </button>
      </div>

      <div className="mb-4">
        <label className="mb-1 block font-medium text-gray-700 text-xs">
          Label
        </label>
        <input
          className="w-full rounded border border-gray-300 px-2 py-1.5 text-sm focus:border-indigo-500 focus:outline-none focus:ring-1 focus:ring-indigo-500"
          onChange={(e) => handleLabelChange(e.target.value)}
          type="text"
          value={localLabel}
        />
      </div>

      <div className="mb-3">
        <span
          className={`font-medium text-xs ${typeColor[nodeType] || "text-gray-600"}`}
        >
          Type: {typeLabel[nodeType] || nodeType}
        </span>
      </div>

      {nodeType === "http" && (
        <HttpConfig config={localConfig} onChange={updateConfig} />
      )}
      {nodeType === "delay" && (
        <DelayConfig config={localConfig} onChange={updateConfig} />
      )}
      {nodeType === "script" && (
        <ScriptConfig config={localConfig} onChange={updateConfig} />
      )}
      {nodeType === "condition" && (
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
  const [headerKey, setHeaderKey] = useState("");
  const [headerVal, setHeaderVal] = useState("");

  const headers = (config.headers || {}) as Record<string, string>;

  const addHeader = () => {
    if (!headerKey.trim()) {
      return;
    }
    const updated = { ...headers, [headerKey.trim()]: headerVal };
    onChange("headers", updated);
    setHeaderKey("");
    setHeaderVal("");
  };

  const removeHeader = (key: string) => {
    const updated = { ...headers };
    delete updated[key];
    onChange("headers", updated);
  };

  return (
    <div className="space-y-3">
      <div>
        <label className="mb-1 block font-medium text-gray-700 text-xs">
          URL *
        </label>
        <input
          className="w-full rounded border border-gray-300 px-2 py-1.5 font-mono text-sm focus:border-indigo-500 focus:outline-none focus:ring-1 focus:ring-indigo-500"
          onChange={(e) => onChange("url", e.target.value)}
          placeholder="https://api.example.com/data"
          type="text"
          value={(config.url as string) || ""}
        />
      </div>

      <div>
        <label className="mb-1 block font-medium text-gray-700 text-xs">
          Method *
        </label>
        <select
          className="w-full rounded border border-gray-300 px-2 py-1.5 text-sm focus:border-indigo-500 focus:outline-none focus:ring-1 focus:ring-indigo-500"
          onChange={(e) => onChange("method", e.target.value)}
          value={(config.method as string) || "GET"}
        >
          <option value="GET">GET</option>
          <option value="POST">POST</option>
          <option value="PUT">PUT</option>
          <option value="PATCH">PATCH</option>
          <option value="DELETE">DELETE</option>
        </select>
      </div>

      <div>
        <label className="mb-1 block font-medium text-gray-700 text-xs">
          Body (JSON)
        </label>
        <textarea
          className="w-full rounded border border-gray-300 px-2 py-1.5 font-mono text-sm focus:border-indigo-500 focus:outline-none focus:ring-1 focus:ring-indigo-500"
          onChange={(e) => onChange("body", e.target.value)}
          placeholder='{"key": "value"}'
          rows={4}
          value={
            typeof config.body === "string"
              ? (config.body as string)
              : config.body
                ? JSON.stringify(config.body, null, 2)
                : ""
          }
        />
        <p className="mt-1 text-gray-400 text-xs">
          Supports template: {"{{inputs.node_id.field}}"}
        </p>
      </div>

      <div>
        <label className="mb-1 block font-medium text-gray-700 text-xs">
          Headers
        </label>
        {Object.entries(headers).map(([key, val]) => (
          <div className="mb-1 flex items-center gap-1" key={key}>
            <span className="flex-1 truncate rounded bg-gray-100 px-2 py-0.5 text-xs">
              {key}: {val}
            </span>
            <button
              className="text-red-400 text-xs hover:text-red-600"
              onClick={() => removeHeader(key)}
            >
              &times;
            </button>
          </div>
        ))}
        <div className="mt-1 flex gap-1">
          <input
            className="flex-1 rounded border border-gray-300 px-2 py-1 text-xs focus:outline-none focus:ring-1 focus:ring-indigo-500"
            onChange={(e) => setHeaderKey(e.target.value)}
            placeholder="Key"
            type="text"
            value={headerKey}
          />
          <input
            className="flex-1 rounded border border-gray-300 px-2 py-1 text-xs focus:outline-none focus:ring-1 focus:ring-indigo-500"
            onChange={(e) => setHeaderVal(e.target.value)}
            placeholder="Value"
            type="text"
            value={headerVal}
          />
          <button
            className="rounded bg-blue-500 px-2 py-1 text-white text-xs hover:bg-blue-600"
            onClick={addHeader}
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
        <label className="mb-1 block font-medium text-gray-700 text-xs">
          Delay (seconds) *
        </label>
        <input
          className="w-full rounded border border-gray-300 px-2 py-1.5 text-sm focus:border-indigo-500 focus:outline-none focus:ring-1 focus:ring-indigo-500"
          min={0}
          onChange={(e) =>
            onChange("seconds", Number.parseInt(e.target.value, 10) || 0)
          }
          type="number"
          value={(config.seconds as number) || 5}
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
        <label className="mb-1 block font-medium text-gray-700 text-xs">
          Code *
        </label>
        <textarea
          className="w-full rounded border border-gray-300 px-2 py-1.5 font-mono text-sm focus:border-indigo-500 focus:outline-none focus:ring-1 focus:ring-indigo-500"
          onChange={(e) => onChange("code", e.target.value)}
          placeholder={'return {"result": {{inputs.node_id.field}}}'}
          rows={8}
          value={(config.code as string) || ""}
        />
        <p className="mt-1 text-gray-400 text-xs">
          Supports: <code>return {`{"key": "value"}`}</code> with template
          variables {"{{inputs.node_id.field}}"}
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
        <label className="mb-1 block font-medium text-gray-700 text-xs">
          Expression *
        </label>
        <input
          className="w-full rounded border border-gray-300 px-2 py-1.5 font-mono text-sm focus:border-indigo-500 focus:outline-none focus:ring-1 focus:ring-indigo-500"
          onChange={(e) => onChange("expression", e.target.value)}
          placeholder="{{inputs.node1.status_code}} == 200"
          type="text"
          value={(config.expression as string) || ""}
        />
        <p className="mt-1 text-gray-400 text-xs">
          Supports: ==, !=, &gt;, &lt;, &gt;=, &lt;= operators
        </p>
      </div>
      <div className="rounded bg-gray-50 p-2 text-gray-500 text-xs">
        <p className="mb-1 font-medium">Examples:</p>
        <ul className="list-inside list-disc space-y-0.5">
          <li>{"{{inputs.node1.status_code}} == 200"}</li>
          <li>{"{{inputs.node1.json.count}} > 10"}</li>
          <li>{`{{inputs.node1.body}} != "error"`}</li>
          <li>true</li>
        </ul>
      </div>
    </div>
  );
}
