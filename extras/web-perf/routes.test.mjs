import { spawn } from "node:child_process";
import process from "node:process";
import { setTimeout as delay } from "node:timers/promises";
import { chromium } from "playwright";

const repoRoot = new URL("../..", import.meta.url);
const assetRoot = new URL("../../internal/web/assets", import.meta.url);
const port = Number(process.env.DEEN_WEB_TEST_PORT || String(19091 + Math.floor(Math.random() * 1000)));
const targetURL = process.env.DEEN_WEB_URL || `http://127.0.0.1:${port}/`;
const basicChainHash =
  "#chain=" +
  Buffer.from(JSON.stringify({ version: 1, steps: [{ plugin: "base64" }] }))
    .toString("base64url")
    .replace(/=+$/, "");
const twoStepChainHash =
  "#chain=" +
  Buffer.from(JSON.stringify({ version: 1, steps: [{ plugin: "base64" }, { plugin: "hex" }] }))
    .toString("base64url")
    .replace(/=+$/, "");

function assert(condition, message) {
  if (!condition) {
    throw new Error(message);
  }
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

async function stopServer(server) {
  if (!server) {
    return;
  }
  try {
    process.kill(-server.pid, "SIGTERM");
  } catch {
    server.kill("SIGTERM");
  }
}

async function afterPaint(page) {
  await page.evaluate(
    () =>
      new Promise((resolve) => {
        requestAnimationFrame(() => requestAnimationFrame(resolve));
      }),
  );
}

async function activeTab(page) {
  return page.locator(".tab.active").textContent();
}

async function dropSourceFile(page, name, content) {
  await page.locator(".source").dispatchEvent("drop", {
    dataTransfer: await page.evaluateHandle(
      ({ fileName, fileContent }) => {
        const data = new DataTransfer();
        data.items.add(new File([fileContent], fileName, { type: "text/plain" }));
        return data;
      },
      { fileName: name, fileContent: content },
    ),
  });
}

async function newTestPage(browser, options = {}) {
  const context = await browser.newContext(options);
  await context.grantPermissions(["clipboard-read", "clipboard-write"], { origin: targetURL });
  const page = await context.newPage();
  return { context, page };
}

async function assertStepActionsFit(page) {
  const result = await page.locator(".card:has(.step-actions)").first().evaluate((card) => {
    const rect = (value) => ({
      left: value.left,
      right: value.right,
      top: value.top,
      bottom: value.bottom,
      width: value.width,
      height: value.height,
    });
    const header = card.querySelector(".card-header");
    const actions = card.querySelector(".step-actions");
    const summary = card.querySelector(".summary");
    if (!header || !actions || !summary) {
      return { ok: false, reason: "missing step header nodes" };
    }
    const cardRect = card.getBoundingClientRect();
    const actionsRect = actions.getBoundingClientRect();
    const summaryRect = summary.getBoundingClientRect();
    const tolerance = 1;
    return {
      ok:
        actionsRect.left >= cardRect.left - tolerance &&
        actionsRect.right <= cardRect.right + tolerance &&
        summaryRect.left >= cardRect.left - tolerance &&
        summaryRect.right <= cardRect.right + tolerance,
      card: rect(cardRect),
      actions: rect(actionsRect),
      summary: rect(summaryRect),
    };
  });
  assert(result.ok, `mobile step actions should fit inside the card: ${JSON.stringify(result)}`);
}

async function main() {
  const server = startServer();
  let browser;
  try {
    await waitForServer(targetURL);
    browser = await chromium.launch({ headless: true });
    const { context, page } = await newTestPage(browser);

    await page.goto(`${targetURL}#about`, { waitUntil: "domcontentloaded" });
    await page.getByRole("button", { name: "About" }).waitFor({ timeout: 15000 });
    assert((await activeTab(page)) === "About", "about route should activate About tab");

    await page.goto(`${targetURL}${basicChainHash}`, { waitUntil: "domcontentloaded" });
    await page.getByRole("button", { name: "Home" }).waitFor({ timeout: 15000 });
    assert((await activeTab(page)) === "Home", "legacy chain route should activate Home tab");
    await page.locator(".card:not(.add)", { hasText: /base64/i }).waitFor({ timeout: 15000 });

    await page.goto(`${targetURL}#examples?search=qr`, { waitUntil: "domcontentloaded" });
    await page.getByRole("textbox", { name: /search examples/i }).waitFor();
    assert((await activeTab(page)) === "Examples", "examples search route should activate Examples tab");
    assert(await page.getByRole("textbox", { name: /search examples/i }).inputValue() === "qr", "examples search should hydrate from URL");

    await page.goto(`${targetURL}#examples/qr-payload-fixture`, { waitUntil: "domcontentloaded" });
    const qrExample = page.locator("details.example-card", { hasText: "QR payload fixture" });
    await qrExample.waitFor();
    await afterPaint(page);
    assert(await qrExample.evaluate((el) => el.open), "example deep link should open matching example");
    await qrExample.getByRole("button", { name: "Preview data" }).click();
    await qrExample.locator("img").waitFor({ timeout: 15000 });
    await qrExample.getByRole("button", { name: "Copy link" }).click();
    assert((await page.evaluate(() => navigator.clipboard.readText())).includes("#examples/qr-payload-fixture"), "example copy link should copy route");

    const fallbackContext = await browser.newContext();
    await fallbackContext.addInitScript(() => {
      Object.defineProperty(navigator, "clipboard", { value: undefined, configurable: true });
      window.prompt = () => "";
    });
    const fallbackPage = await fallbackContext.newPage();
    await fallbackPage.goto(`${targetURL}#examples/qr-payload-fixture`, { waitUntil: "domcontentloaded" });
    const fallbackExample = fallbackPage.locator("details.example-card", { hasText: "QR payload fixture" });
    await fallbackExample.waitFor({ timeout: 15000 });
    await fallbackExample.getByRole("button", { name: "Copy link" }).click();
    assert(
      fallbackPage.url().includes("#examples/qr-payload-fixture"),
      "example copy fallback should keep the route in the address bar",
    );
    await fallbackContext.close();

    await page.goto(`${targetURL}#plugins/base64`, { waitUntil: "domcontentloaded" });
    await page.getByRole("textbox", { name: /search plugins/i }).waitFor();
    assert((await activeTab(page)) === "Plugins", "plugin route should activate Plugins tab");
    const base64Card = page.locator(".plugin-card.route-highlight", { hasText: "Base64" });
    await base64Card.waitFor();
    await base64Card.getByRole("button", { name: "Copy link" }).click();
    assert((await page.evaluate(() => navigator.clipboard.readText())).includes("#plugins/base64"), "plugin copy link should copy route");

    await page.goto(`${targetURL}#home`, { waitUntil: "domcontentloaded" });
    await page.locator(".source").waitFor();
    await dropSourceFile(page, "sample.txt", "deen");
    await page.locator(".meta-source", { hasText: "sample.txt" }).waitFor({ timeout: 15000 });
    await page.evaluate(() => {
      window.__deenDownloads = [];
      if (!HTMLAnchorElement.prototype.__deenDownloadPatched) {
        const originalClick = HTMLAnchorElement.prototype.click;
        HTMLAnchorElement.prototype.click = function patchedClick() {
          if (this.download) {
            window.__deenDownloads.push(this.download);
          }
          return originalClick.call(this);
        };
        HTMLAnchorElement.prototype.__deenDownloadPatched = true;
      }
    });
    await page.getByRole("button", { name: "File" }).click();
    await page.getByRole("menuitem", { name: "Download result" }).click();
    await page.waitForFunction(() => window.__deenDownloads.includes("sample.deen-result.txt"));

    await page.setViewportSize({ width: 390, height: 844 });
    await page.goto(`${targetURL}#examples?search=jwt`, { waitUntil: "domcontentloaded" });
    await page.getByRole("textbox", { name: /search examples/i }).waitFor({ timeout: 15000 });
    assert((await activeTab(page)) === "Examples", "mobile examples route should activate Examples tab");
    await page.goto(`${targetURL}#plugins/base64`, { waitUntil: "domcontentloaded" });
    await page.locator(".plugin-card.route-highlight", { hasText: "Base64" }).waitFor({ timeout: 15000 });
    assert((await activeTab(page)) === "Plugins", "mobile plugin route should activate Plugins tab");
    await page.goto(`${targetURL}${twoStepChainHash}`, { waitUntil: "domcontentloaded" });
    const firstStep = page.locator(".card:has(.step-actions)").first();
    const lastStep = page.locator(".card:has(.step-actions)").nth(1);
    await firstStep.waitFor({ timeout: 15000 });
    await assertStepActionsFit(page);
    assert(await firstStep.getByRole("button", { name: "Move step up" }).isDisabled(), "first step move-up should be disabled");
    assert(await lastStep.getByRole("button", { name: "Move step down" }).isDisabled(), "last step move-down should be disabled");

    await context.close();
    console.log("web route regression tests passed");
  } finally {
    if (browser) {
      await browser.close();
    }
    await stopServer(server);
  }
}

main().catch((err) => {
  console.error(err);
  process.exit(1);
});
