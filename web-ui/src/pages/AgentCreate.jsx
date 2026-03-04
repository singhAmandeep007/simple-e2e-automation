import React, { useState } from "react";
import { useNavigate } from "react-router-dom";
import { agents } from "../api/client";

export default function AgentCreate() {
  const [name, setName] = useState("");
  const [id, setId] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const navigate = useNavigate();

  const handleSubmit = async (e) => {
    e.preventDefault();
    setLoading(true);
    setError(null);
    try {
      const res = await agents.create({ name, id: id || undefined });
      navigate(`/?created=${res.data.id}`);
    } catch (err) {
      setError(err.response?.data?.error || err.message);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="page">
      <h1 className="page-title">Create Agent</h1>
      <div
        className="card"
        style={{ maxWidth: "480px" }}
      >
        <form onSubmit={handleSubmit}>
          <div className="form-group">
            <label
              className="label"
              htmlFor="agentName"
            >
              Agent Name *
            </label>
            <input
              id="agentName"
              className="input"
              type="text"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="e.g. macbook-pro-agent"
              required
            />
          </div>
          <div className="form-group">
            <label
              className="label"
              htmlFor="agentId"
            >
              Agent ID <span style={{ color: "#64748b" }}>(optional — auto-generated if blank)</span>
            </label>
            <input
              id="agentId"
              className="input"
              type="text"
              value={id}
              onChange={(e) => setId(e.target.value)}
              placeholder="e.g. agent-001"
            />
          </div>
          {error && <p className="error-msg">⚠ {error}</p>}
          <button
            id="btnSubmit"
            className="btn btn-primary"
            type="submit"
            disabled={loading}
            style={{ marginTop: "20px" }}
          >
            {loading ? "Creating…" : "+ Create Agent"}
          </button>
        </form>
      </div>

      <div
        className="card"
        style={{ maxWidth: "480px", marginTop: "20px", fontSize: "13px", color: "var(--text-muted)" }}
      >
        <strong style={{ color: "var(--text)" }}>Next steps after creating:</strong>
        <ol style={{ marginTop: "8px", paddingLeft: "16px", lineHeight: "1.9" }}>
          <li>Open Agent Manager desktop app</li>
          <li>Enter your Agent ID and control plane URL</li>
          <li>
            Click <strong>Start Agent</strong> — it will connect automatically
          </li>
          <li>Come back here and run a scan</li>
        </ol>
      </div>
    </div>
  );
}
