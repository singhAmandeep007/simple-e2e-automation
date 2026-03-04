import React, { useEffect, useRef, useState } from "react";
import { scans } from "../api/client";

const POLL_INTERVAL = 1500; // ms

export default function ScanProgress({ scanId, onSuccess, onError }) {
  const [scan, setScan] = useState(null);
  const intervalRef = useRef(null);

  useEffect(() => {
    const poll = async () => {
      try {
        const res = await scans.get(scanId);
        const data = res.data;
        setScan(data);

        if (data.status === "success") {
          clearInterval(intervalRef.current);
          onSuccess?.(data);
        } else if (data.status === "failed") {
          clearInterval(intervalRef.current);
          onError?.(data.error || "Scan failed");
        }
      } catch (err) {
        clearInterval(intervalRef.current);
        onError?.(err.message);
      }
    };

    poll();
    intervalRef.current = setInterval(poll, POLL_INTERVAL);
    return () => clearInterval(intervalRef.current);
  }, [scanId]);

  const statusPct = (s) => ({ pending: 10, running: 55, success: 100, failed: 100 }[s] || 0);

  return (
    <div
      className="card"
      style={{ maxWidth: "520px" }}
    >
      <div style={{ marginBottom: "12px" }}>
        <div style={{ display: "flex", justifyContent: "space-between", marginBottom: "8px" }}>
          <span style={{ fontWeight: "600" }}>Scan in progress</span>
          <span
            className={`badge badge-${scan?.status || "pending"}`}
            data-testid="scanStatus"
          >
            {scan?.status || "pending"}
          </span>
        </div>
        <div style={{ background: "var(--surface2)", borderRadius: "4px", height: "6px", overflow: "hidden" }}>
          <div
            style={{
              height: "100%",
              background: scan?.status === "failed" ? "var(--danger)" : "var(--accent)",
              width: `${statusPct(scan?.status)}%`,
              transition: "width 0.4s ease",
            }}
          />
        </div>
      </div>

      <div style={{ display: "flex", gap: "24px", fontSize: "14px", color: "var(--text-muted)" }}>
        <span>
          Files:{" "}
          <strong
            data-testid="totalFiles"
            style={{ color: "var(--text)" }}
          >
            {scan?.totalFiles ?? "—"}
          </strong>
        </span>
        <span>
          Folders:{" "}
          <strong
            data-testid="totalFolders"
            style={{ color: "var(--text)" }}
          >
            {scan?.totalFolders ?? "—"}
          </strong>
        </span>
      </div>

      <p style={{ fontSize: "12px", color: "var(--text-muted)", marginTop: "12px" }}>
        Scan ID: <code>{scanId}</code>
      </p>
    </div>
  );
}
