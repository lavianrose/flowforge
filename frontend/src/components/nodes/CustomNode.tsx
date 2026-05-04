import { Handle, type NodeProps, Position } from "reactflow";
import "reactflow/dist/style.css";

const nodeStyles: Record<string, string> = {
  http: "bg-blue-100 border-blue-500",
  delay: "bg-yellow-100 border-yellow-500",
  script: "bg-purple-100 border-purple-500",
  condition: "bg-green-100 border-green-500",
};

const iconMap: Record<string, string> = {
  http: "🌐",
  delay: "⏱️",
  script: "📜",
  condition: "❓",
};

function getConfigSummary(
  nodeType: string,
  config: Record<string, unknown>
): string {
  switch (nodeType) {
    case "http": {
      const method = (config.method as string) || "GET";
      const url = (config.url as string) || "(no url)";
      const shortUrl = url.length > 25 ? url.slice(0, 25) + "..." : url;
      return `${method} ${shortUrl}`;
    }
    case "delay": {
      const secs = config.seconds || 0;
      return `${secs}s delay`;
    }
    case "script": {
      const code = (config.code as string) || "";
      const lang = (config.language as string) || "template";
      const langLabel = lang === "template" ? "template" : lang;
      if (!code) {
        return `(no code) [${langLabel}]`;
      }
      const short = code.length > 20 ? code.slice(0, 20) + "..." : code;
      return `${short} [${langLabel}]`;
    }
    case "condition": {
      const expr = (config.expression as string) || "";
      if (!expr) {
        return "(no expression)";
      }
      const short = expr.length > 25 ? expr.slice(0, 25) + "..." : expr;
      return short;
    }
    default:
      return "";
  }
}

export default function CustomNode({ data, selected }: NodeProps) {
  const nodeType = data.type as string;
  const config = (data.config as Record<string, unknown>) || {};
  const bgColor = nodeStyles[nodeType] || "bg-gray-100 border-gray-500";
  const icon = iconMap[nodeType] || "⚙️";
  const summary = getConfigSummary(nodeType, config);

  const hasEmptyRequired =
    (nodeType === "http" && !config.url) ||
    (nodeType === "script" && !config.code) ||
    (nodeType === "condition" && !config.expression);

  return (
    <div
      className={`rounded-lg border-2 px-4 py-2 ${bgColor} ${
        selected ? "ring-2 ring-indigo-500" : ""
      } cursor-pointer`}
      style={{ minWidth: "160px" }}
    >
      <Handle className="h-3 w-3" position={Position.Top} type="target" />
      <div className="flex items-center gap-2">
        <span className="text-xl">{icon}</span>
        <div className="min-w-0 flex-1">
          <div className="truncate font-semibold text-sm">{data.label}</div>
          <div className="text-gray-600 text-xs">{nodeType}</div>
        </div>
      </div>
      {summary && (
        <div className="mt-1 border-gray-300/50 border-t pt-1">
          <div
            className={`truncate font-mono text-xs ${
              hasEmptyRequired ? "text-red-500" : "text-gray-500"
            }`}
            title={summary}
          >
            {summary}
          </div>
        </div>
      )}
      <Handle className="h-3 w-3" position={Position.Bottom} type="source" />
    </div>
  );
}
