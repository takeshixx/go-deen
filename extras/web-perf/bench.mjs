import { spawn } from "node:child_process";
import process from "node:process";
import { setTimeout as delay } from "node:timers/promises";
import { chromium } from "playwright";

const repoRoot = new URL("../..", import.meta.url);
const assetRoot = new URL("../../internal/web/assets", import.meta.url);
const port = Number(process.env.DEEN_WEB_PORT || "19090");
const targetURL = process.env.DEEN_WEB_URL || `http://127.0.0.1:${port}/`;
const iterations = Number(process.env.DEEN_WEB_ITERATIONS || "10");

function percentile(values, p) {
  if (values.length === 0) {
    return 0;
  }
  const sorted = [...values].sort((a, b) => a - b);
  const idx = Math.min(sorted.length - 1, Math.ceil((p / 100) * sorted.length) - 1);
  return sorted[idx];
}

function summarize(label, values) {
  return {
    label,
    min: Math.min(...values),
    p50: percentile(values, 50),
    p95: percentile(values, 95),
    max: Math.max(...values),
  };
}

function summarizeByTab(entries) {
  const grouped = new Map();
  for (const entry of entries) {
    const values = grouped.get(entry.tab) || [];
    values.push(entry.ms);
    grouped.set(entry.tab, values);
  }
  return Object.fromEntries([...grouped.entries()].map(([tab, values]) => [tab, summarize(tab, values)]));
}

async function waitForServer(url, timeoutMs = 10000) {
  const started = Date.now();
  while (Date.now() - started < timeoutMs) {
    try {
      const response = await fetch(url);
      if (response.ok) {
        return;
      }
    } catch {
      // Server is not ready yet.
    }
    await delay(100);
  }
  throw new Error(`timed out waiting for ${url}`);
}

function startServer() {
  if (process.env.DEEN_WEB_URL) {
    return null;
  }
  const child = spawn(
    "go",
    ["run", "./cmd/deen", "serve", "--host", "127.0.0.1", "--port", String(port), "--root", assetRoot.pathname],
    {
      cwd: repoRoot,
      stdio: ["ignore", "pipe", "pipe"],
      env: { ...process.env },
      detached: true,
    },
  );
  child.stdout.on("data", (chunk) => process.stderr.write(chunk));
  child.stderr.on("data", (chunk) => process.stderr.write(chunk));
  return child;
}

async function afterPaint(page) {
  await page.evaluate(
    () =>
      new Promise((resolve) => {
        requestAnimationFrame(() => requestAnimationFrame(resolve));
      }),
  );
}

async function clickTab(page, name) {
  const started = await page.evaluate(() => performance.now());
  await page.getByRole("button", { name }).click();
  await page.waitForFunction(
    (tabName) => {
      const active = document.querySelector(".tab.active");
      return active && active.textContent === tabName;
    },
    name,
  );
  await afterPaint(page);
  const ended = await page.evaluate(() => performance.now());
  return ended - started;
}

async function pageMetrics(page) {
  return page.evaluate(() => ({
    nodes: document.querySelectorAll("*").length,
    heap: performance.memory ? performance.memory.usedJSHeapSize : null,
    cards: document.querySelectorAll(".card, .plugin-card").length,
    textareas: document.querySelectorAll("textarea").length,
  }));
}

async function main() {
  const server = startServer();
  try {
    await waitForServer(targetURL);
    const browser = await chromium.launch({ headless: true });
    const page = await browser.newPage();

    const gotoStart = Date.now();
    await page.goto(targetURL, { waitUntil: "domcontentloaded" });
    await page.getByRole("button", { name: "Home" }).waitFor({ timeout: 15000 });
    await afterPaint(page);
    const startupMs = Date.now() - gotoStart;

    const cold = [];
    for (const tab of ["Examples", "Plugins", "About", "Home"]) {
      cold.push({ tab, ms: await clickTab(page, tab) });
    }

    const warm = [];
    for (let i = 0; i < iterations; i++) {
      for (const tab of ["Examples", "Plugins", "About", "Home"]) {
        warm.push({ tab, ms: await clickTab(page, tab) });
      }
    }

    const metrics = await pageMetrics(page);
    await browser.close();

    const result = {
      url: targetURL,
      iterations,
      startupMs,
      cold,
      warm: summarize("tab click", warm.map((entry) => entry.ms)),
      warmByTab: summarizeByTab(warm),
      metrics,
    };
    console.log(JSON.stringify(result, null, 2));
  } finally {
    if (server) {
      try {
        process.kill(-server.pid, "SIGTERM");
      } catch {
        server.kill("SIGTERM");
      }
    }
  }
}

main().catch((err) => {
  console.error(err);
  process.exit(1);
});
