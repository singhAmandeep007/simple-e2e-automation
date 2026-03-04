import React, { useState } from "react";
import { useParams, useNavigate } from "react-router-dom";
import { scans } from "../api/client";
import ScanProgress from "../components/ScanProgress";

export default function ScanRun() {
  const { agentId } = useParams();
  const navigate = useNavigate();
  const [sourcePath, setSourcePath] = useState("");
  const [scan, setScan] = useState(null);
  const [error, setError] = useState(null);
  const [submitting, setSubmitting] = useState(false);

  const startScan = async (e) => {
    e.preventDefault();
    setSubmitting(true);
    setError(null);
    try {
      const res = await scans.start(agentId, sourcePath);
      setScan({ id: res.data.scanId, status: "pending", totalFiles: 0, totalFolders: 0 });
    } catch (err) {
      setError(err.response?.data?.error || err.message);
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div className="page">
      <div style={{ marginBottom: "24px", fontSize: "13px", color: "var(--text-muted)" }}>
        Agent: <code style={{ color: "var(--accent)" }}>{agentId}</code>
      </div>
      <h1 className="page-title">Run Scan</h1>

      {!scan ? (
        <div
          className="card"
          style={{ maxWidth: "520px" }}
        >
          <form onSubmit={startScan}>
            <div className="form-group">
              <label
                className="label"
                htmlFor="sourcePath"
              >
                Source Path
              </label>
              <input
                id="sourcePath"
                className="input"
                type="text"
                value={sourcePath}
                onChange={(e) => setSourcePath(e.target.value)}
                placeholder="e.g. /Users/you/Documents"
                required
              />
              <p style={{ fontSize: "12px", color: "var(--text-muted)", marginTop: "6px" }}>
                Absolute path on the machine where the agent is running.
              </p>
            </div>
            {error && <p className="error-msg">⚠ {error}</p>}
            <button
              id="btnStartScan"
              className="btn btn-primary"
              type="submit"
              disabled={submitting}
            >
              {submitting ? "Starting…" : "▶ Start Scan"}
            </button>
          </form>
        </div>
      ) : (
        <ScanProgress
          scanId={scan.id}
          onSuccess={(finalScan) => navigate(`/scans/${finalScan.id}/tree`)}
          onError={(err) => setError(err)}
        />
      )}
    </div>
  );
}
