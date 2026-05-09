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
  child.stdout.on("data", (data) => {
    process.stdout.write(`[${label}] ${data}`);
  });
  child.stderr.on("data", (data) => {
    process.stderr.write(`[${label}] ${data}`);
  });
}

/**
 * Build the static assets required before the Go development server starts.
 *
 * @returns {Promise<void>} Resolves after generation succeeds.
 */
async function generateStaticPages() {
  await new Promise((resolve, reject) => {
    const child = spawn(npmCommand(), ["run", "generate"], {
      cwd: repoRoot,
      stdio: ["ignore", "pipe", "pipe"],
    });

    forwardOutput(child, "generate");

    child.on("error", reject);
    child.on("exit", (code, signal) => {
      if (code === 0) {
        resolve();
        return;
      }
      reject(
        new Error(
          `static asset generation failed with code ${code} signal ${signal}`,
        ),
      );
    });
  });
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

  child.on("error", (error) => {
    process.stderr.write(
      `[dev] failed to start Go backend: ${error.message}\n`,
    );
    void shutdown(1);
  });

  child.on("exit", (code, signal) => {
    goProcess = null;
    if (shuttingDown) {
      return;
    }
    process.stderr.write(
      `[dev] Go backend exited unexpectedly with code ${code} signal ${signal}\n`,
    );
    void shutdown(code === 0 || code === null ? 1 : code);
  });

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

  await new Promise((resolve) => {
    let escalationTimer = null;
    const hasExited = () =>
      child.exitCode !== null || child.signalCode !== null;
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

    child.once("exit", (...args) => {
      clearTimeout(escalationTimer);
      resolve(...args);
    });
    sendSignal("SIGINT");
    escalationTimer = setTimeout(() => {
      sendSignal("SIGTERM");
    }, 3000);
    escalationTimer.unref();
  });
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

process.on("SIGINT", () => {
  void shutdown(0);
});

process.on("SIGTERM", () => {
  void shutdown(0);
});

try {
  await generateStaticPages();
  startGo();
  await startVite();
} catch (error) {
  process.stderr.write(`[dev] ${error.message}\n`);
  await shutdown(1);
}
