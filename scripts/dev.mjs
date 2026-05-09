// Copyright (C) 2019-2026 Ni Rui <ranqus@gmail.com>
// Copyright (C) 2026 Snuffy2
// SPDX-License-Identifier: AGPL-3.0-only

import { spawn } from "node:child_process";
import path from "node:path";
import process from "node:process";
import { fileURLToPath } from "node:url";
import { createServer } from "vite";

const repoRoot = path.resolve(
  path.dirname(fileURLToPath(import.meta.url)),
  "..",
);
const devConfig = path.join(repoRoot, "sshwifty.conf.example.json");

let shuttingDown = false;
let goProcess = null;
let viteServer = null;

/**
 * Return the platform-specific npm executable name.
 *
 * @returns {string} npm command name for the current platform.
 */
function npmCommand() {
  return process.platform === "win32" ? "npm.cmd" : "npm";
}

/**
 * Prefix a child process's stdout and stderr before forwarding them.
 *
 * @param {import("node:child_process").ChildProcess} child Child process.
 * @param {string} label Output label to add before each chunk.
 */
function forwardOutput(child, label) {
  child.stdout.on(
    "data",
    /**
     * Forward a stdout chunk from the child process.
     *
     * @param {Buffer | string} data Output chunk.
     */
    (data) => {
      process.stdout.write(`[${label}] ${data}`);
    },
  );
  child.stderr.on(
    "data",
    /**
     * Forward a stderr chunk from the child process.
     *
     * @param {Buffer | string} data Error output chunk.
     */
    (data) => {
      process.stderr.write(`[${label}] ${data}`);
    },
  );
}

/**
 * Build the static assets required before the Go development server starts.
 *
 * @returns {Promise<void>} Resolves after generation succeeds.
 */
async function generateStaticPages() {
  await new Promise(
    /**
     * Run the generation command and settle with its process result.
     *
     * @param {() => void} resolve Promise resolver.
     * @param {(reason?: unknown) => void} reject Promise rejecter.
     */
    (resolve, reject) => {
      const child = spawn(npmCommand(), ["run", "generate"], {
        cwd: repoRoot,
        stdio: ["ignore", "pipe", "pipe"],
      });

      forwardOutput(child, "generate");

      child.on("error", reject);
      child.on(
        "exit",
        /**
         * Convert the generator process exit state into a promise outcome.
         *
         * @param {number | null} code Process exit code.
         * @param {NodeJS.Signals | null} signal Signal that ended the process.
         */
        (code, signal) => {
          if (code === 0) {
            resolve();
            return;
          }
          reject(
            new Error(
              `static asset generation failed with code ${code} signal ${signal}`,
            ),
          );
        },
      );
    },
  );
}

/**
 * Start the Go backend with the example development configuration.
 */
function startGo() {
  const child = spawn("go", ["run", "sshwifty.go"], {
    cwd: repoRoot,
    detached: process.platform !== "win32",
    env: {
      ...process.env,
      CGO_ENABLED: "0",
      SSHWIFTY_CONFIG: devConfig,
      SSHWIFTY_DEBUG: "_",
    },
    stdio: ["ignore", "pipe", "pipe"],
  });

  forwardOutput(child, "go");

  child.on(
    "error",
    /**
     * Shut down the supervisor if the Go backend cannot start.
     *
     * @param {Error} error Spawn error.
     */
    (error) => {
      process.stderr.write(
        `[dev] failed to start Go backend: ${error.message}\n`,
      );
      void shutdown(1);
    },
  );

  child.on(
    "exit",
    /**
     * Stop the dev stack if the backend exits before supervisor shutdown.
     *
     * @param {number | null} code Process exit code.
     * @param {NodeJS.Signals | null} signal Signal that ended the process.
     */
    (code, signal) => {
      goProcess = null;
      if (shuttingDown) {
        return;
      }
      process.stderr.write(
        `[dev] Go backend exited unexpectedly with code ${code} signal ${signal}\n`,
      );
      void shutdown(code === 0 || code === null ? 1 : code);
    },
  );

  goProcess = child;
}

/**
 * Start the Vite development server with the repository config.
 *
 * @returns {Promise<void>} Resolves after Vite is listening.
 */
async function startVite() {
  viteServer = await createServer({
    configFile: path.join(repoRoot, "vite.config.js"),
    root: repoRoot,
  });
  await viteServer.listen();
  viteServer.printUrls();
}

/**
 * Stop the Go backend process group, escalating if it does not exit quickly.
 *
 * @returns {Promise<void>} Resolves after the backend has exited.
 */
async function stopGo() {
  if (!goProcess) {
    return;
  }

  const child = goProcess;
  goProcess = null;

  await new Promise(
    /**
     * Send shutdown signals and resolve when the backend process exits.
     *
     * @param {(...args: unknown[]) => void} resolve Promise resolver.
     */
    (resolve) => {
      let escalationTimer = null;
      /**
       * Report whether the child process has already terminated.
       *
       * @returns {boolean} True once exit code or signal is populated.
       */
      const hasExited = () =>
        child.exitCode !== null || child.signalCode !== null;
      /**
       * Send a process signal if the child process is still running.
       *
       * @param {NodeJS.Signals} signal Signal to send.
       */
      const sendSignal = (signal) => {
        if (hasExited()) {
          return;
        }
        try {
          if (process.platform === "win32") {
            child.kill(signal);
          } else {
            process.kill(-child.pid, signal);
          }
        } catch (error) {
          if (error.code !== "ESRCH") {
            throw error;
          }
        }
      };

      child.once(
        "exit",
        /**
         * Clear pending escalation and resolve with the process exit details.
         *
         * @param {unknown[]} args Process exit event arguments.
         */
        (...args) => {
          clearTimeout(escalationTimer);
          resolve(...args);
        },
      );
      sendSignal("SIGINT");
      escalationTimer = setTimeout(
        /**
         * Escalate shutdown if the interrupt signal did not stop the backend.
         */
        () => {
          sendSignal("SIGTERM");
        },
        3000,
      );
      escalationTimer.unref();
    },
  );
}

/**
 * Stop the Vite development server if it is running.
 *
 * @returns {Promise<void>} Resolves after the server closes.
 */
async function stopVite() {
  if (!viteServer) {
    return;
  }
  const server = viteServer;
  viteServer = null;
  await server.close();
}

/**
 * Stop all development processes and exit the supervisor.
 *
 * @param {number} exitCode Process exit code to use.
 * @returns {Promise<void>} Resolves only if process exit is intercepted.
 */
async function shutdown(exitCode) {
  if (shuttingDown) {
    return;
  }
  shuttingDown = true;
  await Promise.allSettled([stopVite(), stopGo()]);
  process.exit(exitCode);
}

process.on(
  "SIGINT",
  /**
   * Stop the dev stack after an interactive interrupt.
   */
  () => {
    void shutdown(0);
  },
);

process.on(
  "SIGTERM",
  /**
   * Stop the dev stack after a termination request.
   */
  () => {
    void shutdown(0);
  },
);

try {
  await generateStaticPages();
  startGo();
  await startVite();
} catch (error) {
  process.stderr.write(`[dev] ${error.message}\n`);
  await shutdown(1);
}
