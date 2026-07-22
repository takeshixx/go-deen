import { spawn } from "node:child_process";
import { readFile } from "node:fs/promises";
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
const urlPartsChainHash =
  "#chain=" +
  Buffer.from(
    JSON.stringify({
      version: 1,
      steps: [
        { plugin: "urlparts" },
        { plugin: "urlparts", unprocess: true },
      ],
    }),
  )
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

async function dropSourceFiles(page, files, options = {}) {
  if (options.directory) {
    await page.locator(".source").evaluate((source, sourceFiles) => {
      const event = new Event("drop", { bubbles: true, cancelable: true });
      Object.defineProperty(event, "dataTransfer", {
        value: {
          items: [{ webkitGetAsEntry: () => ({ isDirectory: true }) }],
          files: { length: 1, item: () => new File([""], sourceFiles[0].name, { type: "text/plain" }) },
        },
      });
      source.dispatchEvent(event);
    }, files);
    return;
  }
  await page.locator(".source").dispatchEvent("drop", {
    dataTransfer: await page.evaluateHandle(
      ({ sourceFiles }) => {
        const data = new DataTransfer();
        for (const sourceFile of sourceFiles) {
          data.items.add(new File([sourceFile.content], sourceFile.name, { type: "text/plain" }));
        }
        return data;
      },
      { sourceFiles: files },
    ),
  });
}

async function dropSourceFile(page, name, content) {
  await dropSourceFiles(page, [{ name, content }]);
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

async function assertURLPartsEditorFits(page) {
  const result = await page.getByTestId("urlparts-editor").evaluate((editor) => ({
    ok: editor.scrollWidth <= editor.clientWidth + 1,
    clientWidth: editor.clientWidth,
    scrollWidth: editor.scrollWidth,
  }));
  assert(result.ok, `mobile URL Parts editor should not overflow horizontally: ${JSON.stringify(result)}`);
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
    await dropSourceFiles(page, [
      { name: "one.txt", content: "one" },
      { name: "two.txt", content: "two" },
    ]);
    await page.locator(".source-drop-message.visible", { hasText: "Drop one file at a time." }).waitFor({ timeout: 15000 });
    await page.locator(".meta-source", { hasText: "sample.txt" }).waitFor({ timeout: 15000 });
    await dropSourceFiles(page, [{ name: "folder-entry", content: "" }], { directory: true });
    await page.locator(".source-drop-message.visible", { hasText: "Directories are not supported." }).waitFor({ timeout: 15000 });
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

    await page.goto(`${targetURL}${urlPartsChainHash}`, { waitUntil: "domcontentloaded" });
    const urlSource =
      "https://login-update.example.invalid/a/verify.php?utm_source=mail&redirect=https%3A%2F%2Fportal.example.org%2Fsignin&campaign=retry#continue";
    await page.locator(".source textarea").fill(urlSource);
    const urlStep = page.locator(".card:has(.step-actions)").first();
    const rebuiltStep = page.locator(".card:has(.step-actions)").nth(1);
    const expandURLStep = urlStep.getByRole("button", { name: "Expand step" });
    if (await expandURLStep.count()) {
      await expandURLStep.click();
    }
    const urlEditor = urlStep.getByTestId("urlparts-editor");
    await urlEditor.waitFor({ timeout: 15000 });
    assert(
      await urlStep.getByRole("button", { name: "URL Parts", exact: true }).evaluate((button) => button.classList.contains("active")),
      "forward URL Parts step should select the structured editor",
    );
    assert((await urlEditor.getByRole("textbox", { name: "Rebuilt URL" }).inputValue()) === urlSource, "rebuilt URL should initially match source");
    assert(
      (await urlEditor.getByRole("textbox", { name: "Defanged URL" }).inputValue()).startsWith("hxxps://login-update[.]example[.]invalid/"),
      "analysis should provide a defanged URL",
    );
    const initialAnalysis = await urlEditor.locator(".urlparts-analysis-findings").textContent();
    assert(initialAnalysis.includes("Nested URLs") && initialAnalysis.includes("portal.example.org"), "analysis should expose nested redirect URLs");
    assert(initialAnalysis.includes("Common tracking parameters") && initialAnalysis.includes("utm_source"), "analysis should identify common tracking parameters");
    await urlEditor.getByRole("button", { name: "Copy defanged URL" }).click();
    assert((await page.evaluate(() => navigator.clipboard.readText())).startsWith("hxxps://login-update[.]example[.]invalid/"), "copy defanged URL should use the local analysis value");

    await urlEditor.getByRole("textbox", { name: "URL hostname" }).fill("review.invalid");
    await urlEditor.getByRole("textbox", { name: "Path segment 2" }).fill("checked");
    await urlEditor.getByRole("textbox", { name: "Query value 2" }).fill("https://safe.example.org/result");
    await urlEditor.getByRole("textbox", { name: "URL fragment" }).fill("reviewed");
    const editedURL =
      "https://review.invalid/a/checked?utm_source=mail&redirect=https%3A%2F%2Fsafe.example.org%2Fresult&campaign=retry#reviewed";
    assert((await urlEditor.getByRole("textbox", { name: "Rebuilt URL" }).inputValue()) === editedURL, "structured edits should rebuild the URL");
    assert((await rebuiltStep.locator("textarea.io").first().inputValue()) === editedURL, "structured edits should recompute a downstream reverse step");
    assert((await urlEditor.getByRole("textbox", { name: "Defanged URL" }).inputValue()).startsWith("hxxps://review[.]invalid/"), "defanged URL should update with structured edits");
    assert((await urlEditor.locator(".urlparts-analysis-findings").textContent()).includes("safe.example.org"), "nested URL analysis should update with query edits");

    await urlEditor.getByRole("checkbox", { name: "Show original encoded values" }).check();
    assert((await urlEditor.locator(".urlparts-raw:visible").count()) >= 4, "raw encoding toggle should reveal encoded values");
    await urlEditor.getByRole("button", { name: "Copy URL" }).click();
    assert((await page.evaluate(() => navigator.clipboard.readText())) === editedURL, "copy URL should copy the rebuilt URL");

    await urlEditor.getByRole("textbox", { name: "URL port" }).fill("invalid");
    assert(await urlEditor.getByRole("button", { name: "Copy URL" }).isDisabled(), "invalid URL fields should disable copy");
    assert(await urlEditor.getByRole("button", { name: "Copy defanged URL" }).isDisabled(), "invalid URL fields should disable defanged copy");
    assert((await urlEditor.locator(".urlparts-message.invalid").textContent()).includes("invalid URL port"), "invalid port should show validation feedback");
    await urlEditor.getByRole("textbox", { name: "URL port" }).fill("");

    await urlEditor.getByRole("button", { name: "Add query parameter" }).click();
    await urlStep.getByRole("textbox", { name: "Query key 4" }).waitFor();
    await urlStep.getByRole("textbox", { name: "Query key 4" }).fill("review");
    await urlStep.getByRole("textbox", { name: "Query value 4" }).fill("passed");
    assert((await urlStep.getByRole("textbox", { name: "Rebuilt URL" }).inputValue()).includes("&review=passed#"), "added query row should affect rebuilt URL");
    await urlStep.getByRole("button", { name: "Move query parameter 4 up" }).click();
    await page.waitForFunction(() => document.querySelector('[aria-label="Query key 3"]')?.value === "review");
    assert((await urlStep.getByRole("textbox", { name: "Query key 3" }).inputValue()) === "review", "query move control should preserve ordered parameters");
    await urlStep.getByRole("button", { name: "Duplicate query parameter 1" }).click();
    await urlStep.getByRole("textbox", { name: "Query key 5" }).waitFor();
    await urlStep.getByRole("button", { name: "Remove query parameter 2" }).click();
    await page.waitForFunction(() => document.querySelectorAll('[aria-label^="Query key "]').length === 4);

    await urlStep.getByRole("button", { name: "Raw", exact: true }).click();
    const rawURLParts = urlStep.locator("textarea.io").first();
    const rawDocument = JSON.parse(await rawURLParts.inputValue());
    assert(rawDocument.version === 2 && rawDocument.analysis?.nested_urls?.length === 1, "raw JSON should include versioned machine-readable analysis");
    rawDocument.fragment = "from-raw-json";
    await rawURLParts.fill(JSON.stringify(rawDocument, null, 2));
    await urlStep.getByRole("button", { name: "URL Parts", exact: true }).click();
    assert((await urlStep.getByRole("textbox", { name: "URL fragment" }).inputValue()) === "from-raw-json", "raw JSON edits should refresh structured fields");

    const actionURL =
      "https://login-update.example.invalid/action?utm_source=mail&redirect=https%3A%2F%2Fportal.example.org%2Fsignin&fbclid=abc#continue";
    const selectedTrackingRemovedURL =
      "https://login-update.example.invalid/action?redirect=https%3A%2F%2Fportal.example.org%2Fsignin&fbclid=abc#continue";
    const allTrackingRemovedURL =
      "https://login-update.example.invalid/action?redirect=https%3A%2F%2Fportal.example.org%2Fsignin#continue";
    await page.goto(`${targetURL}${urlPartsChainHash}`, { waitUntil: "domcontentloaded" });
    await page.locator(".source textarea").fill(actionURL);
    const actionStep = page.locator(".card:has(.step-actions)").first();
    const expandActionStep = actionStep.getByRole("button", { name: "Expand step" });
    if (await expandActionStep.count()) {
      await expandActionStep.click();
    }
    await actionStep.getByRole("button", { name: "Remove tracking parameter 1" }).click();
    await page.waitForFunction(
      (expected) => document.querySelector('[aria-label="Rebuilt URL"]')?.value === expected,
      selectedTrackingRemovedURL,
    );
    assert(
      (await page.locator(".card:has(.step-actions)").nth(1).locator("textarea.io").first().inputValue()) === selectedTrackingRemovedURL,
      "selected tracking removal should update downstream reconstruction",
    );
    await page.getByRole("button", { name: "Workflow" }).click();
    await page.getByRole("menuitem", { name: "Undo" }).click();
    await page.waitForFunction((expected) => document.querySelector('[aria-label="Rebuilt URL"]')?.value === expected, actionURL);

    await actionStep.getByRole("button", { name: "Remove all detected tracking parameters" }).click();
    await page.waitForFunction(
      (expected) => document.querySelector('[aria-label="Rebuilt URL"]')?.value === expected,
      allTrackingRemovedURL,
    );
    await actionStep.getByRole("button", { name: "Use nested URL from query parameter 1 as source" }).click();
    await page.waitForFunction(() => document.querySelector(".source textarea")?.value === "https://portal.example.org/signin");
    assert(
      (await page.locator(".card:has(.step-actions)").nth(1).locator("textarea.io").first().inputValue()) === "https://portal.example.org/signin",
      "nested URL promotion should recompute the current chain",
    );
    await page.getByRole("button", { name: "Workflow" }).click();
    await page.getByRole("menuitem", { name: "Undo" }).click();
    await page.waitForFunction((expected) => document.querySelector(".source textarea")?.value === expected, actionURL);
    await page.waitForFunction(
      (expected) => document.querySelector('[aria-label="Rebuilt URL"]')?.value === expected,
      allTrackingRemovedURL,
    );
    await page.getByRole("button", { name: "Workflow" }).click();
    await page.getByRole("menuitem", { name: "Undo" }).click();
    await page.waitForFunction((expected) => document.querySelector('[aria-label="Rebuilt URL"]')?.value === expected, actionURL);

    if (process.env.DEEN_URLPARTS_TEST_FILE) {
      const fixtureURL = (await readFile(process.env.DEEN_URLPARTS_TEST_FILE, "utf8")).replace(/\r?\n$/, "");
      assert(fixtureURL.length > 0 && !fixtureURL.includes("\n"), "URL Parts fixture should contain one non-empty URL");
      await page.goto(`${targetURL}${urlPartsChainHash}`, { waitUntil: "domcontentloaded" });
      await page.locator(".source textarea").fill(fixtureURL);
      const fixtureStep = page.locator(".card:has(.step-actions)").first();
      const expandFixtureStep = fixtureStep.getByRole("button", { name: "Expand step" });
      if (await expandFixtureStep.count()) {
        await expandFixtureStep.click();
      }
      await page.waitForFunction(
        (expected) => document.querySelector('[aria-label="Rebuilt URL"]')?.value === expected,
        fixtureURL,
      );
      const fixtureDocument = JSON.parse(await fixtureStep.locator("textarea.io").first().inputValue());
      assert(
        fixtureDocument.version === 2 && fixtureDocument.analysis?.defanged_url?.length > 0,
        "URL Parts fixture should produce versioned local analysis",
      );
      assert(
        (await page.locator(".card:has(.step-actions)").nth(1).locator("textarea.io").first().inputValue()) === fixtureURL,
        "URL Parts fixture should survive a forward and reverse browser round trip",
      );
      if (fixtureDocument.analysis.tracking_parameters.length > 0) {
        await fixtureStep.getByRole("button", { name: "Remove all detected tracking parameters" }).click();
        await page.waitForFunction(() => !document.querySelector('[aria-label="Remove all detected tracking parameters"]'));
        await page.getByRole("button", { name: "Workflow" }).click();
        await page.getByRole("menuitem", { name: "Undo" }).click();
        await page.waitForFunction(
          (expected) => document.querySelector('[aria-label="Rebuilt URL"]')?.value === expected,
          fixtureURL,
        );
      }
      if (fixtureDocument.analysis.nested_urls.length > 0) {
        const nested = fixtureDocument.analysis.nested_urls[0];
        await fixtureStep
          .getByRole("button", { name: `Use nested URL from query parameter ${nested.parameter_index} as source` })
          .click();
        await page.waitForFunction((expected) => document.querySelector(".source textarea")?.value === expected, nested.url);
        await page.getByRole("button", { name: "Workflow" }).click();
        await page.getByRole("menuitem", { name: "Undo" }).click();
        await page.waitForFunction((expected) => document.querySelector(".source textarea")?.value === expected, fixtureURL);
      }
    }

    await page.setViewportSize({ width: 390, height: 844 });
    await page.goto(`${targetURL}${urlPartsChainHash}`, { waitUntil: "domcontentloaded" });
    await page.locator(".source textarea").fill(urlSource);
    const mobileURLStep = page.locator(".card:has(.step-actions)").first();
    const expandMobileURLStep = mobileURLStep.getByRole("button", { name: "Expand step" });
    if (await expandMobileURLStep.count()) {
      await expandMobileURLStep.click();
    }
    await mobileURLStep.getByTestId("urlparts-editor").waitFor({ timeout: 15000 });
    await assertURLPartsEditorFits(page);

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
