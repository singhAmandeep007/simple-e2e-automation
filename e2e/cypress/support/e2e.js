import "./commands";
import "@cypress/grep";

before(() => {
  // Verify the Control Plane is reachable before running any test.
  // This produces a clear error message if services are not started.
  cy.task("checkServicesReady").then((msg) => {
    cy.log(msg);
  });
  // Clean the DB to ensure each spec starts from a known empty state.
  cy.task("cleanDb");
});
