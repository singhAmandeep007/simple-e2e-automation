import React, { useEffect, useState } from "react";
import { useParams, Link } from "react-router-dom";
import { scans } from "../api/client";
import FolderTree from "../components/FolderTree";

export default function ScanTree() {
  const { scanId } = useParams();
  const [treeData, setTreeData] = useState(null);
  const [scanInfo, setScanInfo] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  useEffect(() => {
    const load = async () => {
      try {
        const [infoRes, treeRes] = await Promise.all([scans.get(scanId), scans.tree(scanId)]);
        setScanInfo(infoRes.data);
        setTreeData(treeRes.data);
      } catch (err) {
        setError(err.response?.data?.error || err.message);
      } finally {
        setLoading(false);
      }
    };
    load();
  }, [scanId]);

  if (loading)
    return (
      <div className="page">
        <p style={{ color: "var(--text-muted)" }}>Loading scan results…</p>
      </div>
    );
  if (error)
    return (
      <div className="page">
        <p className="error-msg">⚠ {error}</p>
      </div>
    );

  return (
    <div className="page">
      <h1 className="page-title">Scan Results</h1>

      {scanInfo && (
        <div
          className="card"
          style={{ marginBottom: "20px", display: "flex", gap: "32px", flexWrap: "wrap" }}
        >
          <div>
            <div className="label">Source Path</div>
            <code
              data-testid="sourcePath"
              style={{ color: "var(--accent)" }}
            >
              {scanInfo.sourcePath}
            </code>
          </div>
          <div>
            <div className="label">Files</div>
            <strong
              data-testid="totalFiles"
              style={{ fontSize: "18px" }}
            >
              {scanInfo.totalFiles}
            </strong>
          </div>
          <div>
            <div className="label">Folders</div>
            <strong
              data-testid="totalFolders"
              style={{ fontSize: "18px" }}
            >
              {scanInfo.totalFolders}
            </strong>
          </div>
          <div>
            <div className="label">Status</div>
            <span className={`badge badge-${scanInfo.status}`}>{scanInfo.status}</span>
          </div>
        </div>
      )}

      <div className="card">
        <div className="section-title">Folder Tree</div>
        {treeData?.tree?.length > 0 ? (
          <FolderTree nodes={treeData.tree} />
        ) : (
          <p style={{ color: "var(--text-muted)", fontSize: "14px" }}>No entries found.</p>
        )}
      </div>

      <Link
        to="/"
        className="btn btn-secondary"
        style={{ marginTop: "20px" }}
      >
        ← Back to Agents
      </Link>
    </div>
  );
}
