import "./commands";
import "@cypress/grep";

before(() => {
  cy.task("cleanDb");
});
