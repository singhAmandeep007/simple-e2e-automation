import React, { useState, useEffect } from "react";
import { GetStatus, GetConfig, StartAgent, StopAgent } from "../../wailsjs/go/main/App";
import { EventsOn } from "../../wailsjs/runtime/runtime";

const STATUS_COLOR = { online: "#68d391", connecting: "#f6e05e", offline: "#fc8181" };
const STATUS_LABEL = { online: "● Online", connecting: "◌ Connecting…", offline: "○ Offline" };

const CARD = {
  background: "linear-gradient(145deg, #1e293b 0%, #0f172a 100%)",
  borderRadius: "16px",
  padding: "28px",
  marginBottom: "24px",
  boxShadow: "0 10px 40px rgba(0,0,0,0.4)",
  border: "1px solid rgba(255,255,255,0.05)",
};
const BTN = (color) => ({
  padding: "12px 32px",
  background: color,
  border: "none",
  borderRadius: "8px",
  color: "#fff",
  fontWeight: "700",
  cursor: "pointer",
  fontSize: "14px",
  marginRight: "12px",
  boxShadow: `0 4px 14px ${color}50`,
  transition: "all 0.2s ease-in-out",
});
const ERR = { marginTop: "16px", color: "#fc8181", fontSize: "14px", fontWeight: "500" };
const INFO = { fontSize: "13px", color: "#a0aec0", marginTop: "8px" };

export default function Status() {
  const [status, setStatus] = useState("offline");
  const [cfg, setCfg] = useState({});
  const [error, setError] = useState(null);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    GetStatus()
      .then(setStatus)
      .catch(() => {});
    GetConfig()
      .then((c) => setCfg(c || {}))
      .catch(() => {});

    // Listen for live status events from Go backend
    const unlisten = EventsOn("agent:status", (newStatus) => {
      setStatus(newStatus);
    });
    return unlisten;
  }, []);

  const handleStart = async () => {
    setError(null);
    setLoading(true);
    try {
      await StartAgent(cfg.agentId || "", cfg.controlPlaneWs || "ws://localhost:4000/ws");
    } catch (err) {
      setError(String(err));
    } finally {
      setLoading(false);
    }
  };

  const handleStop = async () => {
    setError(null);
    try {
      await StopAgent();
    } catch (err) {
      setError(String(err));
    }
  };

  const isRunning = status === "online" || status === "connecting";

  return (
    <div>
      <h2 style={{ fontSize: "20px", fontWeight: "700", marginBottom: "20px", color: "#fff" }}>Agent Status</h2>
      <div style={CARD}>
        <div style={{ display: "flex", alignItems: "center", gap: "16px", marginBottom: "20px" }}>
          <div
            id="statusBadge"
            style={{
              fontSize: "18px",
              fontWeight: "700",
              color: STATUS_COLOR[status] || STATUS_COLOR.offline,
            }}
          >
            {STATUS_LABEL[status] || STATUS_LABEL.offline}
          </div>
        </div>

        <p style={INFO}>
          Agent ID: <strong>{cfg.agentId || "not set"}</strong>
        </p>
        <p style={INFO}>
          Control Plane: <strong>{cfg.controlPlaneWs || "not set"}</strong>
        </p>

        <div style={{ marginTop: "20px" }}>
          {!isRunning && (
            <button
              id="btnStart"
              style={BTN("#e94560")}
              onClick={handleStart}
              disabled={loading}
            >
              {loading ? "Starting…" : "Start Agent"}
            </button>
          )}
          {isRunning && (
            <button
              id="btnStop"
              style={BTN("#718096")}
              onClick={handleStop}
            >
              Stop Agent
            </button>
          )}
        </div>

        {error && <p style={ERR}>⚠ {error}</p>}
      </div>

      <div style={{ ...CARD, fontSize: "13px", color: "#718096" }}>
        <strong style={{ color: "#a0aec0" }}>How it works</strong>
        <ol style={{ marginTop: "8px", paddingLeft: "16px", lineHeight: "1.8" }}>
          <li>
            Configure your Agent ID and Control Plane URL in the <em>Configure</em> tab
          </li>
          <li>
            Click <strong>Start Agent</strong> — the go-agent binary will launch and connect
          </li>
          <li>Status turns green when connected to the control plane</li>
          <li>Use the Web UI to trigger scan commands on this agent</li>
        </ol>
      </div>
    </div>
  );
}
