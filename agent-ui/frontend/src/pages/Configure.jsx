import React, { useState, useEffect } from "react";
import { GetConfig, SaveConfig } from "../../wailsjs/go/main/App";

const CARD = {
  background: "linear-gradient(145deg, #1e293b 0%, #0f172a 100%)",
  borderRadius: "16px",
  padding: "28px",
  marginBottom: "24px",
  boxShadow: "0 10px 40px rgba(0,0,0,0.4)",
  border: "1px solid rgba(255,255,255,0.05)",
};
const INPUT = {
  width: "100%",
  background: "rgba(15, 52, 96, 0.4)",
  border: "1px solid rgba(233, 69, 96, 0.3)",
  borderRadius: "8px",
  padding: "12px 16px",
  color: "#fff",
  fontSize: "14px",
  marginTop: "8px",
  outline: "none",
  transition: "border 0.2s",
};
const LABEL = { display: "block", fontSize: "14px", color: "#cbd5e1", fontWeight: "600", marginTop: "16px" };
const BTN = {
  marginTop: "8px",
  padding: "12px 32px",
  background: "#e94560",
  border: "none",
  borderRadius: "8px",
  color: "#fff",
  fontWeight: "700",
  fontSize: "14px",
  cursor: "pointer",
  boxShadow: "0 4px 14px rgba(233, 69, 96, 0.4)",
  transition: "all 0.2s ease-in-out",
};
const MSG = (ok) => ({ marginTop: "12px", color: ok ? "#68d391" : "#fc8181", fontSize: "13px" });

export default function Configure() {
  const [form, setForm] = useState({ agentId: "", controlPlaneWs: "ws://localhost:4000/ws", agentBinPath: "" });
  const [message, setMessage] = useState(null);
  const [ok, setOk] = useState(false);

  useEffect(() => {
    GetConfig()
      .then((cfg) => {
        if (cfg)
          setForm({
            agentId: cfg.agentId || "",
            controlPlaneWs: cfg.controlPlaneWs || "ws://localhost:4000/ws",
            agentBinPath: cfg.agentBinPath || "",
          });
      })
      .catch(() => {});
  }, []);

  const save = async (e) => {
    e.preventDefault();
    try {
      await SaveConfig({ agentId: form.agentId, controlPlaneWs: form.controlPlaneWs, agentBinPath: form.agentBinPath });
      setMessage("Configuration saved successfully");
      setOk(true);
    } catch (err) {
      setMessage("Failed to save: " + err);
      setOk(false);
    }
  };

  const set = (k) => (e) => setForm((f) => ({ ...f, [k]: e.target.value }));

  return (
    <div>
      <h2 style={{ fontSize: "20px", fontWeight: "700", marginBottom: "20px", color: "#fff" }}>Configure Agent</h2>
      <form onSubmit={save}>
        <div style={CARD}>
          <label style={LABEL}>Agent ID</label>
          <input
            id="agentId"
            style={INPUT}
            value={form.agentId}
            onChange={set("agentId")}
            placeholder="e.g. agent-001"
            required
          />

          <label style={LABEL}>Control Plane WebSocket URL</label>
          <input
            id="cpUrl"
            style={INPUT}
            value={form.controlPlaneWs}
            onChange={set("controlPlaneWs")}
            placeholder="ws://localhost:4000/ws"
            required
          />

          <label style={LABEL}>go-agent Binary Path (optional)</label>
          <input
            id="agentBinPath"
            style={INPUT}
            value={form.agentBinPath}
            onChange={set("agentBinPath")}
            placeholder="leave blank to auto-detect"
          />
        </div>
        <button
          type="submit"
          style={BTN}
        >
          Save Configuration
        </button>
        {message && <p style={MSG(ok)}>{message}</p>}
      </form>
    </div>
  );
}
