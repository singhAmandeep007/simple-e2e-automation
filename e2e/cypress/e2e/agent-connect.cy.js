/**
 * @tag @agent
 * Agent Connection / Disconnect Tests
 *
 * Tests the full lifecycle: spawn agent → wait for online →
 * verify UI + DB → kill agent → verify offline in UI + DB
 */

const FIXTURE_PATH = `/tmp/e2e-connect-${Date.now()}`;

describe("Agent Connection", { tags: "@agent" }, () => {
  let agentId;
  let pid;

  before(() => {
    // ARRANGE: Create agent via API & generate a fixture directory
    cy.apiCreateAgent({ name: "connect-test-agent" }).then((agent) => {
      agentId = agent.id;
    });
    cy.task("generateFixtures", { basePath: FIXTURE_PATH, folders: 1, filesPerFolder: 2 });
  });

  after(() => {
    // TEARDOWN: kill agent if still running, clean fixtures
    if (agentId) {
      cy.task("killAgent", { agentId }).then(() => {
        cy.task("waitForAgentStatus", { agentId, expectedStatus: "offline", timeoutMs: 6000 }); // ignore if already offline
      });
    }
    cy.task("cleanFixtures", { basePath: FIXTURE_PATH });
  });

  it("agent goes online after binary starts", () => {
    // ── ACT ──────────────────────────────────────────────────────────────────
    cy.task("spawnAgent", { agentId, cpUrl: Cypress.env("CP_WS_URL") }).then((p) => {
      pid = p;
    });

    // ── ASSERT DB ─────────────────────────────────────────────────────────────
    cy.task("waitForAgentStatus", { agentId, expectedStatus: "online", timeoutMs: 10000 });
    cy.dbAssertAgentStatus(agentId, "online");

    // ── ASSERT UI ─────────────────────────────────────────────────────────────
    cy.visit("/");
    cy.get(`[data-testid="agent-id"]`)
      .contains(agentId)
      .closest(".card")
      .find('[data-testid="agentStatusBadge"]')
      .should("contain", "Online");
  });

  it("agent shows as connected in API", () => {
    cy.apiGetAgent(agentId).then((agent) => {
      expect(agent.connected).to.be.true;
      expect(agent.status).to.equal("online");
    });
  });

  it("Run Scan button is enabled for a connected agent", () => {
    cy.visit("/");
    cy.get(`#btnScan-${agentId}`).should("be.visible").and("not.be.disabled");
  });

  it("agent goes offline after binary is killed", () => {
    // ── ACT ──────────────────────────────────────────────────────────────────
    cy.task("killAgent", { agentId });

    // ── ASSERT DB ─────────────────────────────────────────────────────────────
    cy.task("waitForAgentStatus", { agentId, expectedStatus: "offline", timeoutMs: 8000 });
    cy.dbAssertAgentStatus(agentId, "offline");

    // ── ASSERT UI ─────────────────────────────────────────────────────────────
    cy.visit("/");
    cy.get(`[data-testid="agent-id"]`)
      .contains(agentId)
      .closest(".card")
      .find('[data-testid="agentStatusBadge"]')
      .should("contain", "Offline");
  });

  it("agent reconnects after restart", () => {
    // ── ACT: restart agent ───────────────────────────────────────────────────
    cy.task("spawnAgent", { agentId, cpUrl: Cypress.env("CP_WS_URL") });

    // ── ASSERT: back online ───────────────────────────────────────────────────
    cy.task("waitForAgentStatus", { agentId, expectedStatus: "online", timeoutMs: 12000 });
    cy.dbAssertAgentStatus(agentId, "online");

    cy.visit("/");
    cy.get(`[data-testid="agent-id"]`)
      .contains(agentId)
      .closest(".card")
      .find('[data-testid="agentStatusBadge"]')
      .should("contain", "Online");
  });
});
