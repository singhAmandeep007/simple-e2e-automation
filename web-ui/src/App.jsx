import React from "react";
import { BrowserRouter, Routes, Route, Link, NavLink } from "react-router-dom";
import AgentCreate from "./pages/AgentCreate";
import AgentList from "./pages/AgentList";
import ScanRun from "./pages/ScanRun";
import ScanTree from "./pages/ScanTree";

const NAV_STYLE = {
  background: "#1a1d27",
  borderBottom: "1px solid #2d3748",
  padding: "0 32px",
  display: "flex",
  alignItems: "center",
  gap: "32px",
  height: "56px",
};
const LOGO = { fontWeight: "700", fontSize: "16px", color: "#6366f1", letterSpacing: "-0.5px" };
const NAV_LINK_STYLE = ({ isActive }) => ({
  color: isActive ? "#6366f1" : "#94a3b8",
  fontWeight: isActive ? "600" : "400",
  fontSize: "14px",
  padding: "4px 0",
  borderBottom: isActive ? "2px solid #6366f1" : "2px solid transparent",
  transition: "all 0.2s",
});

export default function App() {
  return (
    <BrowserRouter>
      <nav style={NAV_STYLE}>
        <span style={LOGO}>⚡ Migration Platform</span>
        <NavLink
          to="/"
          end
          style={NAV_LINK_STYLE}
        >
          Agents
        </NavLink>
        <NavLink
          to="/agents/create"
          style={NAV_LINK_STYLE}
        >
          New Agent
        </NavLink>
      </nav>
      <Routes>
        <Route
          path="/"
          element={<AgentList />}
        />
        <Route
          path="/agents/create"
          element={<AgentCreate />}
        />
        <Route
          path="/agents/:agentId/scan"
          element={<ScanRun />}
        />
        <Route
          path="/scans/:scanId/tree"
          element={<ScanTree />}
        />
      </Routes>
    </BrowserRouter>
  );
}
