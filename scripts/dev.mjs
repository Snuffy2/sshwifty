// Copyright (C) 2019-2026 Ni Rui <ranqus@gmail.com>
// Copyright (C) 2026 Snuffy2
// SPDX-License-Identifier: AGPL-3.0-only

import { spawn } from "node:child_process";
import path from "node:path";
import process from "node:process";
import { fileURLToPath } from "node:url";
import { createServer } from "vite";

const repoRoot = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "..");
const devConfig = path.join(repoRoot, "sshwifty.conf.example.json");

let shuttingDown = false;
let goProcess = null;
let viteServer = null;

function forwardOutput(child, label) {
  child.stdout.on("data", (data) => {
    process.stdout.write(`[${label}] ${data}`);
  });
  child.stderr.on("data", (data) => {
    process.stderr.write(`[${label}] ${data}`);
  });
}

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
    process.stderr.write(`[dev] failed to start Go backend: ${error.message}\n`);
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

async function startVite() {
  viteServer = await createServer({
    configFile: path.join(repoRoot, "vite.config.js"),
    root: repoRoot,
  });
  await viteServer.listen();
  viteServer.printUrls();
}

async function stopGo() {
  if (!goProcess) {
    return;
  }

  const child = goProcess;
  goProcess = null;

  await new Promise((resolve) => {
    child.once("exit", resolve);
    if (process.platform === "win32") {
      child.kill("SIGINT");
    } else {
      process.kill(-child.pid, "SIGINT");
    }
    setTimeout(() => {
      if (!child.killed) {
        if (process.platform === "win32") {
          child.kill("SIGTERM");
        } else {
          process.kill(-child.pid, "SIGTERM");
        }
      }
    }, 3000).unref();
  });
}

async function stopVite() {
  if (!viteServer) {
    return;
  }
  const server = viteServer;
  viteServer = null;
  await server.close();
}

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
  startGo();
  await startVite();
} catch (error) {
  process.stderr.write(`[dev] ${error.message}\n`);
  await shutdown(1);
}
