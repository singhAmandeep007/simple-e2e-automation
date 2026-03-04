const { defineConfig } = require("cypress");
const { spawn, execSync } = require("child_process");
const fs = require("fs");
const path = require("path");
const Database = require("better-sqlite3");

// Shared state: running agent processes
const agentProcesses = {};

// Path to SQLite DB — overridable for Docker: CYPRESS_DB_PATH=/cp-data/data.db
const DB_PATH = process.env.CYPRESS_DB_PATH || path.resolve(__dirname, "../control-plane/data/data.db");

// Path to go-agent binary — overridable for Docker: CYPRESS_AGENT_BIN=/usr/local/bin/go-agent
const AGENT_BIN = process.env.CYPRESS_AGENT_BIN || path.resolve(__dirname, "../bin/go-agent");

module.exports = defineConfig({
  e2e: {
    baseUrl: "http://127.0.0.1:5173",
    specPattern: "cypress/e2e/**/*.cy.js",
    supportFile: "cypress/support/e2e.js",
    video: true,
    screenshotOnRunFailure: true,
    defaultCommandTimeout: 10000,
    requestTimeout: 10000,
    responseTimeout: 15000,
    env: {
      CP_URL: "http://localhost:4000",
      CP_WS_URL: "ws://localhost:4000/ws",
      AGENT_BIN: AGENT_BIN,
      DB_PATH: DB_PATH,
    },
    setupNodeEvents(on, config) {
      // ── @cypress/grep ─────────────────────────────────────────────────────
      require("@cypress/grep/src/plugin")(config);

      on("task", {
        // ── Agent lifecycle ─────────────────────────────────────────────────
        /**
         * Spawn the go-agent binary and return the pid.
         * @param {{ agentId: string, cpUrl?: string, configDir?: string }} opts
         */
        spawnAgent({ agentId, cpUrl = "ws://localhost:4000/ws" }) {
          if (!fs.existsSync(AGENT_BIN)) {
            throw new Error(`Agent binary not found at ${AGENT_BIN}. Run: cd agent && go build -o ../bin/go-agent .`);
          }

          const proc = spawn(AGENT_BIN, ["start", "--id", agentId, "--cp-url", cpUrl], {
            detached: false,
            stdio: "pipe",
            cwd: path.resolve(__dirname, "../agent"), // so config.yaml is found
          });

          agentProcesses[agentId] = proc;
          const pid = proc.pid;

          proc.stdout.on("data", (d) => process.stdout.write(`[agent:${agentId}] ${d}`));
          proc.stderr.on("data", (d) => process.stderr.write(`[agent:${agentId}] ${d}`));
          proc.on("exit", (code) => {
            delete agentProcesses[agentId];
            process.stdout.write(`[agent:${agentId}] exited (code=${code})\n`);
          });

          return pid;
        },

        /**
         * Kill a running agent by agentId or pid.
         */
        killAgent({ agentId }) {
          const proc = agentProcesses[agentId];
          if (proc) {
            try {
              proc.kill("SIGTERM");
            } catch (_) {}
            delete agentProcesses[agentId];
          }
          return null;
        },

        // ── Fixture helpers ─────────────────────────────────────────────────
        /**
         * Generate N folders each containing M files at basePath.
         */
        generateFixtures({ basePath, folders = 3, filesPerFolder = 3 }) {
          fs.mkdirSync(basePath, { recursive: true });
          for (let f = 1; f <= folders; f++) {
            const dir = path.join(basePath, `folder-${f}`);
            fs.mkdirSync(dir, { recursive: true });
            for (let i = 1; i <= filesPerFolder; i++) {
              const content = `test file ${i} in folder ${f}\n`.repeat(10);
              fs.writeFileSync(path.join(dir, `file-${i}.txt`), content);
            }
          }
          // Also add a root-level file
          fs.writeFileSync(path.join(basePath, "root-file.txt"), "root level test file\n");
          return { basePath, folders, filesPerFolder };
        },

        /**
         * Remove the fixture directory.
         */
        cleanFixtures({ basePath }) {
          if (fs.existsSync(basePath)) {
            fs.rmSync(basePath, { recursive: true, force: true });
          }
          return null;
        },

        // ── Database helpers ─────────────────────────────────────────────────
        /**
         * Clears all tables in the SQLite database to ensure test isolation.
         */
        cleanDb() {
          if (!fs.existsSync(DB_PATH)) return null;
          const db = new Database(DB_PATH);
          try {
            db.exec("DELETE FROM scan_tree;");
            db.exec("DELETE FROM scans;");
            db.exec("DELETE FROM agents;");
          } catch (e) {
            console.error("[cy:cleanDb] error:", e);
          } finally {
            db.close();
          }
          return null;
        },

        /**
         * Run a raw SQL query against the SQLite DB.
         * Returns all rows.
         */
        queryDb({ sql, params = [] }) {
          if (!fs.existsSync(DB_PATH)) {
            return [];
          }
          const db = new Database(DB_PATH, { readonly: true });
          try {
            const stmt = db.prepare(sql);
            return stmt.all(...params);
          } finally {
            db.close();
          }
        },

        /**
         * Poll DB until agent status matches expected, or timeout.
         */
        waitForAgentStatus({ agentId, expectedStatus, timeoutMs = 10000 }) {
          const start = Date.now();
          const db = new Database(DB_PATH, { readonly: true });
          try {
            while (Date.now() - start < timeoutMs) {
              const row = db.prepare("SELECT status FROM agents WHERE id = ?").get(agentId);
              if (row?.status === expectedStatus) return true;
              // Busy-wait with tiny sleep — acceptable in Node task
              const end = Date.now() + 300;
              while (Date.now() < end) {
                /* spin */
              }
            }
            throw new Error(`Timeout waiting for agent ${agentId} to become "${expectedStatus}"`);
          } finally {
            db.close();
          }
        },

        /**
         * Wait for a scan to reach a terminal status.
         */
        waitForScanStatus({ scanId, expectedStatus, timeoutMs = 20000 }) {
          const start = Date.now();
          const db = new Database(DB_PATH, { readonly: true });
          try {
            while (Date.now() - start < timeoutMs) {
              const row = db.prepare("SELECT status FROM scans WHERE id = ?").get(scanId);
              if (row?.status === expectedStatus) return true;
              if (row?.status === "failed") throw new Error(`Scan ${scanId} failed`);
              const end = Date.now() + 500;
              while (Date.now() < end) {
                /* spin */
              }
            }
            throw new Error(`Timeout waiting for scan ${scanId} to become "${expectedStatus}"`);
          } finally {
            db.close();
          }
        },

        // ── Logging ──────────────────────────────────────────────────────────
        log(msg) {
          console.log(`[cy:task] ${msg}`);
          return null;
        },
      });

      return config;
    },
  },
});
