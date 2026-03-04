import React, { useState } from "react";
import Configure from "./pages/Configure";
import Status from "./pages/Status";

const NAV_STYLE = {
  display: "flex",
  gap: "1px",
  background: "#0f3460",
  padding: "12px 0 0 80px", // Push right to clear macOS traffic lights
  borderBottom: "1px solid #16213e",
  "--wails-draggable": "drag",
};
const TAB = (active) => ({
  padding: "12px 24px",
  cursor: "pointer",
  border: "none",
  background: active ? "#e94560" : "transparent",
  color: "#fff",
  fontWeight: active ? "700" : "400",
  fontSize: "14px",
  transition: "background 0.2s",
  "--wails-draggable": "no-drag",
});
const CONTAINER = { minHeight: "100vh", background: "#1a1a2e", color: "#e0e0e0" };

export default function App() {
  const [tab, setTab] = useState("status");

  return (
    <div style={CONTAINER}>
      <nav style={NAV_STYLE}>
        <button
          style={TAB(tab === "status")}
          onClick={() => setTab("status")}
        >
          Status
        </button>
        <button
          style={TAB(tab === "configure")}
          onClick={() => setTab("configure")}
        >
          Configure
        </button>
      </nav>
      <div style={{ padding: "24px" }}>{tab === "status" ? <Status /> : <Configure />}</div>
    </div>
  );
}
