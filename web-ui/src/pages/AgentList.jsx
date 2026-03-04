import React, { useState, useEffect } from "react";
import { Link, useSearchParams } from "react-router-dom";
import { agents } from "../api/client";
import AgentStatusBadge from "../components/AgentStatusBadge";

export default function AgentList() {
  const [agentList, setAgentList] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [searchParams] = useSearchParams();
  const justCreated = searchParams.get("created");

  const fetchAgents = async () => {
    try {
      const res = await agents.list();
      setAgentList(res.data || []);
      setError(null);
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchAgents();
    const interval = setInterval(fetchAgents, 3000); // poll every 3s
    return () => clearInterval(interval);
  }, []);

  return (
    <div className="page">
      <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: "24px" }}>
        <h1
          className="page-title"
          style={{ marginBottom: 0 }}
        >
          Agents
        </h1>
        <Link
          id="btnCreateAgent"
          to="/agents/create"
          className="btn btn-primary"
        >
          + New Agent
        </Link>
      </div>

      {justCreated && (
        <div
          style={{
            background: "rgba(16,185,129,.1)",
            border: "1px solid rgba(16,185,129,.3)",
            borderRadius: "8px",
            padding: "12px 16px",
            marginBottom: "20px",
            fontSize: "14px",
            color: "#10b981",
          }}
        >
          ✓ Agent <strong>{justCreated}</strong> created. Install and start the agent binary to connect.
        </div>
      )}

      {loading && !agentList.length && <p style={{ color: "var(--text-muted)" }}>Loading agents…</p>}

      {error && <p className="error-msg">⚠ {error}</p>}

      {!loading && !error && agentList.length === 0 && (
        <div
          className="card"
          style={{ textAlign: "center", padding: "48px" }}
        >
          <p style={{ color: "var(--text-muted)", marginBottom: "16px" }}>No agents yet.</p>
          <Link
            to="/agents/create"
            className="btn btn-primary"
          >
            Create your first agent
          </Link>
        </div>
      )}

      <div style={{ display: "flex", flexDirection: "column", gap: "12px" }}>
        {agentList.map((agent) => (
          <div
            key={agent.id}
            className="card"
            style={{ display: "flex", alignItems: "center", gap: "16px" }}
          >
            <div style={{ flex: 1 }}>
              <div
                style={{ fontWeight: "600", marginBottom: "4px" }}
                data-testid="agent-name"
              >
                {agent.name}
              </div>
              <div
                style={{ fontSize: "12px", color: "var(--text-muted)", fontFamily: "monospace" }}
                data-testid="agent-id"
              >
                {agent.id}
              </div>
            </div>

            <AgentStatusBadge status={agent.connected ? "online" : agent.status} />

            <Link
              id={`btnScan-${agent.id}`}
              to={`/agents/${agent.id}/scan`}
              className={`btn btn-secondary ${!agent.connected ? "btn-disabled" : ""}`}
              style={{ opacity: agent.connected ? 1 : 0.4, pointerEvents: agent.connected ? "auto" : "none" }}
            >
              Run Scan
            </Link>
          </div>
        ))}
      </div>
    </div>
  );
}
