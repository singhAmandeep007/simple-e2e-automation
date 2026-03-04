/**
 * @tag @agent
 * Agent Creation Tests
 *
 * ARRANGE → ACT → ASSERT pattern testing the full agent creation flow:
 * UI form submission → API validation → DB assertion
 */

describe("Agent Creation", { tags: "@agent" }, () => {
  const agentName = `e2e-agent-${Date.now()}`;
  let createdAgentId;

  before(() => {
    // Ensure control plane is healthy
    cy.request("GET", `${Cypress.env("CP_URL")}/health`)
      .its("status")
      .should("eq", 200);
  });

  it("creates an agent via the UI and verifies it in DB", () => {
    // ── ARRANGE ──────────────────────────────────────────────────────────────
    cy.visit("/agents/create");

    // ── ACT ──────────────────────────────────────────────────────────────────
    cy.get("#agentName").type(agentName);
    // Leave ID blank — let the server generate one
    cy.get("#btnSubmit").click();

    // ── ASSERT UI ─────────────────────────────────────────────────────────────
    // Should redirect to agent list with success banner
    cy.url().should("include", "/?created=");
    cy.url().then((url) => {
      createdAgentId = new URL(url).searchParams.get("created");
      expect(createdAgentId).to.be.a("string").and.not.empty;
    });

    // Banner shown
    cy.get('[data-testid="agent-name"]').should("contain", agentName);
    cy.get('[data-testid="agentStatusBadge"]').first().should("contain", "Offline");

    // ── ASSERT API ────────────────────────────────────────────────────────────
    cy.then(() => {
      cy.apiGetAgent(createdAgentId).then((agent) => {
        expect(agent.name).to.equal(agentName);
        expect(agent.status).to.equal("offline");
      });
    });

    // ── ASSERT DB ─────────────────────────────────────────────────────────────
    cy.then(() => {
      cy.dbAssertAgentStatus(createdAgentId, "offline");
    });
  });

  it("creates an agent with an explicit ID via UI", () => {
    const explicitId = `explicit-${Date.now()}`;

    // ── ARRANGE ──────────────────────────────────────────────────────────────
    cy.visit("/agents/create");

    // ── ACT ──────────────────────────────────────────────────────────────────
    cy.get("#agentName").type(`Explicit ID Agent`);
    cy.get("#agentId").type(explicitId);
    cy.get("#btnSubmit").click();

    // ── ASSERT ────────────────────────────────────────────────────────────────
    cy.url().should("include", `created=${explicitId}`);
    cy.get('[data-testid="agent-id"]').should("contain", explicitId);
    cy.dbAssertAgentStatus(explicitId, "offline");
  });

  it("shows error for duplicate agent ID", () => {
    // ── ARRANGE ──────────────────────────────────────────────────────────────
    const dupeId = `dupe-${Date.now()}`;
    cy.apiCreateAgent({ id: dupeId, name: "First Agent" });

    // ── ACT ──────────────────────────────────────────────────────────────────
    cy.visit("/agents/create");
    cy.get("#agentName").type("Second Agent");
    cy.get("#agentId").type(dupeId);
    cy.get("#btnSubmit").click();

    // ── ASSERT ────────────────────────────────────────────────────────────────
    cy.get(".error-msg")
      .should("be.visible")
      .invoke("text")
      .should("match", /error|UNIQUE/i);
  });
});
