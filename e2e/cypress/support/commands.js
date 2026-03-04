/**
 * Custom Cypress commands for the Unified E2E POC.
 * Use these in tests to avoid repetition.
 */

const CP_URL = Cypress.env("CP_URL") || "http://localhost:4000";
const CP_WS_URL = Cypress.env("CP_WS_URL") || "ws://localhost:4000/ws";

// ── API helpers ──────────────────────────────────────────────────────────────

/**
 * Create an agent via REST API (not UI).
 * Yields the created agent object.
 */
Cypress.Commands.add("apiCreateAgent", ({ name, id } = {}) => {
  const agentId = id || `test-agent-${Date.now()}`;
  return cy
    .request({
      method: "POST",
      url: `${CP_URL}/agents`,
      body: { name: name || `Test Agent ${agentId}`, id: agentId },
    })
    .its("body");
});

/**
 * Get an agent by ID via REST API.
 */
Cypress.Commands.add("apiGetAgent", (agentId) => {
  return cy.request(`${CP_URL}/agents/${agentId}`).its("body");
});

/**
 * Start a scan via REST API.
 * Yields { scanId, agentId, sourcePath, status }
 */
Cypress.Commands.add("apiStartScan", (agentId, sourcePath) => {
  return cy
    .request({
      method: "POST",
      url: `${CP_URL}/agents/${agentId}/scan`,
      body: { sourcePath },
    })
    .its("body");
});

/**
 * Get a scan by ID via REST API.
 */
Cypress.Commands.add("apiGetScan", (scanId) => {
  return cy.request(`${CP_URL}/scans/${scanId}`).its("body");
});

// ── Agent lifecycle ──────────────────────────────────────────────────────────

/**
 * Spawn the go-agent binary and wait for it to be online in DB.
 * Yields the process pid.
 */
Cypress.Commands.add("startAgent", (agentId) => {
  return cy.task("spawnAgent", { agentId, cpUrl: CP_WS_URL }).then((pid) => {
    // Wait for the agent to register and appear as online in DB
    cy.task("waitForAgentStatus", { agentId, expectedStatus: "online", timeoutMs: 10000 });
    return cy.wrap(pid);
  });
});

/**
 * Kill a running agent and wait for it to go offline.
 */
Cypress.Commands.add("stopAgent", (agentId) => {
  return cy.task("killAgent", { agentId }).then(() => {
    cy.task("waitForAgentStatus", { agentId, expectedStatus: "offline", timeoutMs: 8000 });
  });
});

// ── Assertions ───────────────────────────────────────────────────────────────

/**
 * Assert agent status in DB matches expected.
 */
Cypress.Commands.add("dbAssertAgentStatus", (agentId, expectedStatus) => {
  return cy
    .task("queryDb", {
      sql: "SELECT status FROM agents WHERE id = ?",
      params: [agentId],
    })
    .then((rows) => {
      expect(rows).to.have.length(1);
      expect(rows[0].status).to.equal(expectedStatus);
    });
});

/**
 * Assert scan tree row count in DB.
 */
Cypress.Commands.add("dbAssertScanTreeCount", (scanId, expectedCount) => {
  return cy
    .task("queryDb", {
      sql: "SELECT COUNT(*) as cnt FROM scan_tree WHERE scan_id = ?",
      params: [scanId],
    })
    .then((rows) => {
      expect(rows[0].cnt).to.equal(expectedCount);
    });
});
