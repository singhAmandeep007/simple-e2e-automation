import React from "react";

const DOT = { width: "8px", height: "8px", borderRadius: "50%", display: "inline-block" };
const STATUS_MAP = {
  online: { color: "#10b981", label: "Online" },
  offline: { color: "#64748b", label: "Offline" },
  connecting: { color: "#f59e0b", label: "Connecting" },
};

export default function AgentStatusBadge({ status }) {
  const s = STATUS_MAP[status] || STATUS_MAP.offline;
  return (
    <span
      data-testid="agentStatusBadge"
      className={`badge badge-${status || "offline"}`}
    >
      <span style={{ ...DOT, background: s.color }} />
      {s.label}
    </span>
  );
}
