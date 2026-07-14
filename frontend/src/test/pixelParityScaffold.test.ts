import { existsSync, readFileSync } from "node:fs";
import { resolve } from "node:path";

import { describe, expect, it } from "vitest";

const FRONTEND_ROOT = resolve(__dirname, "..", "..");
const PACKAGE_JSON = resolve(FRONTEND_ROOT, "package.json");
const FRONTEND_GITIGNORE = resolve(FRONTEND_ROOT, ".gitignore");
const PLAYWRIGHT_CONFIG = resolve(FRONTEND_ROOT, "playwright.config.ts");
const PIXEL_PARITY_DIR = resolve(FRONTEND_ROOT, "tests", "pixel-parity");
const SERVE_SCRIPT = resolve(FRONTEND_ROOT, "scripts", "serve-pixel-parity.mjs");
const SCREENSHOT_BASELINES = resolve(
  PIXEL_PARITY_DIR,
  "screenshot.spec.ts-snapshots",
);

describe("pixel parity scaffold (Phase 1.1 + 1.2 + 1.3)", () => {
  const pkg = JSON.parse(readFileSync(PACKAGE_JSON, "utf8"));
  const merged = {
    ...(pkg.dependencies ?? {}),
    ...(pkg.devDependencies ?? {}),
  };

  it("declares @playwright/test as devDependency", () => {
    expect(merged["@playwright/test"]).toBeTruthy();
  });

  it("exposes test:pixel-parity and test:pixel-parity:install scripts", () => {
    expect(pkg.scripts?.["test:pixel-parity"]).toBeTruthy();
    expect(pkg.scripts?.["test:pixel-parity:install"]).toBeTruthy();
    // The install script must be the canonical Playwright install command so
    // CI or human operators get a single, copy-paste-able entry point.
    expect(pkg.scripts?.["test:pixel-parity:install"]).toMatch(
      /playwright install.*chromium/,
    );
  });

  it("declares no other browser-rendering frameworks (cypress / puppeteer)", () => {
    expect(merged.cypress).toBeUndefined();
    expect(merged.puppeteer).toBeUndefined();
    expect(merged["@cypress/test"]).toBeUndefined();
  });

  it("does not retain an unused axe adapter", () => {
    expect(merged["@axe-core/playwright"]).toBeUndefined();
  });

  it("has a playwright.config.ts at the frontend root", () => {
    expect(existsSync(PLAYWRIGHT_CONFIG)).toBe(true);
    const cfg = readFileSync(PLAYWRIGHT_CONFIG, "utf8");
    expect(cfg).toMatch(/testDir:\s*["']\.\/tests\/pixel-parity["']/);
    // Two viewport projects are mandatory.
    expect(cfg).toMatch(/name:\s*["']desktop["']/);
    expect(cfg).toMatch(/name:\s*["']mobile["']/);
    expect(cfg).toMatch(/width:\s*1440[\s\S]*height:\s*900/);
    expect(cfg).toMatch(/width:\s*390[\s\S]*height:\s*844/);
    expect(cfg).not.toContain("toHaveScreenshot");
    // webServer must point at the colocated serve script.
    expect(cfg).toMatch(
      /webServer\s*:\s*\{[\s\S]*serve-pixel-parity\.mjs[\s\S]*\}/,
    );
    expect(cfg).toMatch(/url:\s*["']http:\/\/127\.0\.0\.1:4173\/health["']/);
  });

  it("has the pixel-parity test directory shape", () => {
    expect(existsSync(PIXEL_PARITY_DIR)).toBe(true);
  });

  it("keeps buffer-only screenshot smoke free of local baselines", () => {
    expect(existsSync(SCREENSHOT_BASELINES)).toBe(false);
    expect(readFileSync(FRONTEND_GITIGNORE, "utf8")).not.toMatch(
      /tests\/pixel-parity\/.*snapshots/,
    );
  });

  it("ships the serve-pixel-parity.mjs static server fixture", () => {
    expect(existsSync(SERVE_SCRIPT)).toBe(true);
    const src = readFileSync(SERVE_SCRIPT, "utf8");
    // Must use only Node built-ins or local sibling `.mjs` files — no
    // third-party deps. Allowed: `node:*`, bare http/fs/path/url, or any
    // relative path starting with `./` or `../`.
    expect(src).not.toMatch(
      /from\s+["'](?!node:|http$|fs$|path$|url$|\.\/|\.\.\/)/,
    );
    expect(src).toMatch(/createServer/);
    // Health probe path expected by playwright.config.ts.
    expect(src).toMatch(/['"]\/health['"]/);
    // Mounts both frontend dist and ui-design at the documented prefixes.
    expect(src).toMatch(/['"]\/ui-design['"]/);
    // Fail-loud guard for missing dist or ui-design directories.
    expect(src).toMatch(/process\.exit\(1\)/);
  });
});
