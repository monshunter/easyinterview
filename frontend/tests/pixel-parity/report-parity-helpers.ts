import { expect, type Page, type TestInfo } from "@playwright/test";
import pixelmatch from "pixelmatch";
import { PNG } from "pngjs";
import { readFileSync } from "node:fs";
import { resolve } from "node:path";

interface OperationFixture {
  scenarios: Record<
    string,
    {
      response: {
        status: number;
        headers?: Record<string, string>;
        body: Record<string, unknown>;
      };
    }
  >;
}

export interface BoxSnapshot {
  x: number;
  y: number;
  width: number;
  height: number;
}

export interface SurfaceSnapshot {
  box: BoxSnapshot;
  style: Record<string, string>;
}

const REPO_ROOT = resolve(process.cwd(), "..");
const REPORT_FIXTURE = "openapi/fixtures/Reports/getFeedbackReport.json";
const FIXED_NOW = new Date("2026-07-12T00:00:00.000Z");

export function reportFixture(scenario: string) {
  return fixtureResponse(REPORT_FIXTURE, scenario);
}

export async function configureDeterministicPage(
  page: Page,
  lang: "zh" | "en",
) {
  await page.clock.setFixedTime(FIXED_NOW);
  await page.addInitScript((uiLanguage) => {
    try {
      localStorage.setItem("ei-lang", uiLanguage);
    } catch {
      // about:blank has no storage origin; the script reruns after navigation.
    }
  }, lang);
  await page.emulateMedia({ reducedMotion: "reduce" });
}

export async function mockFormalReportApis(page: Page, scenario: string) {
  const mutableContextReads: string[] = [];
  await page.route("**/api/v1/**", async (route) => {
    const url = route.request().url();
    if (/\/reports\/[^/]+$/.test(new URL(url).pathname)) {
      return fulfillFixture(route, REPORT_FIXTURE, scenario);
    }
    if (url.endsWith("/runtime/config")) {
      return fulfillFixture(route, "openapi/fixtures/Auth/getRuntimeConfig.json", "default");
    }
    if (url.endsWith("/me")) {
      return fulfillFixture(route, "openapi/fixtures/Auth/getMe.json", "default");
    }
    if (url.includes("/targets/") || url.includes("/resumes/")) {
      mutableContextReads.push(url);
      return route.fulfill({ status: 500, body: "mutable report context read is forbidden" });
    }
    return route.fulfill({ status: 204, body: "" });
  });
  return mutableContextReads;
}

export async function injectPrototypeReportFixture(
  page: Page,
  scenario: string,
  owner: "report" | "reportGeneration",
) {
  const body = reportFixture(scenario).body;
  await page.route("**/ui-design/src/data.jsx*", async (route) => {
    const response = await route.fetch();
    const source = await response.text();
    await route.fulfill({
      response,
      contentType: "application/javascript; charset=utf-8",
      body: `${source}\nwindow.EI_DATA.${owner} = ${JSON.stringify(body)};`,
    });
  });
}

export async function settleVisualSurface(page: Page) {
  await page.addStyleTag({
    content: `
      *, *::before, *::after {
        animation: none !important;
        transition: none !important;
        caret-color: transparent !important;
      }
    `,
  });
  await page.evaluate(async () => {
    await document.fonts.ready;
    await new Promise<void>((resolveFrame) => requestAnimationFrame(() => requestAnimationFrame(() => resolveFrame())));
  });
}

export async function pauseDeterministicClock(page: Page) {
  await page.clock.pauseAt(FIXED_NOW);
}

export async function normalizedText(page: Page, selector: string) {
  return page.locator(selector).evaluate((node) =>
    (node.textContent ?? "").replace(/\s+/g, " ").trim(),
  );
}

export async function surfaceSnapshot(
  page: Page,
  selector: string,
  properties: readonly string[],
  relativeTo?: string,
): Promise<SurfaceSnapshot> {
  return page.locator(selector).evaluate((node, options) => {
    const element = node as HTMLElement;
    const box = element.getBoundingClientRect();
    const origin = options.relativeTo
      ? document.querySelector(options.relativeTo)?.getBoundingClientRect()
      : null;
    const computed = getComputedStyle(element);
    const style: Record<string, string> = {};
    for (const property of options.cssProperties) {
      style[property] = computed.getPropertyValue(property);
    }
    return {
      box: {
        x: box.x - (origin?.x ?? 0),
        y: box.y - (origin?.y ?? 0),
        width: box.width,
        height: box.height,
      },
      style,
    };
  }, { cssProperties: [...properties], relativeTo });
}

export function expectSurfaceParity(
  formal: SurfaceSnapshot,
  prototype: SurfaceSnapshot,
  label: string,
  tolerance = 1,
) {
  expect(formal.style, `${label} computed style`).toEqual(prototype.style);
  for (const key of ["x", "y", "width", "height"] as const) {
    expect(
      Math.abs(formal.box[key] - prototype.box[key]),
      `${label} bbox.${key}: formal=${formal.box[key]}, prototype=${prototype.box[key]}`,
    ).toBeLessThanOrEqual(tolerance);
  }
}

export async function expectPixelParity(
  formalPage: Page,
  prototypePage: Page,
  selector: string,
  testInfo: TestInfo,
  label: string,
  maxChangedRatio = 0.005,
) {
  const [formalBuffer, prototypeBuffer] = await Promise.all([
    formalPage.locator(selector).screenshot({ animations: "disabled" }),
    prototypePage.locator(selector).screenshot({ animations: "disabled" }),
  ]);
  return expectImageBufferParity(
    formalBuffer,
    prototypeBuffer,
    testInfo,
    label,
    maxChangedRatio,
  );
}

export async function expectFullPagePixelParity(
  formalPage: Page,
  prototypePage: Page,
  testInfo: TestInfo,
  label: string,
  maxChangedRatio = 0.005,
) {
  const [formalBuffer, prototypeBuffer] = await Promise.all([
    formalPage.screenshot({ fullPage: true, animations: "disabled" }),
    prototypePage.screenshot({ fullPage: true, animations: "disabled" }),
  ]);
  return expectImageBufferParity(
    formalBuffer,
    prototypeBuffer,
    testInfo,
    label,
    maxChangedRatio,
  );
}

async function expectImageBufferParity(
  formalBuffer: Buffer,
  prototypeBuffer: Buffer,
  testInfo: TestInfo,
  label: string,
  maxChangedRatio: number,
) {
  const formal = PNG.sync.read(formalBuffer);
  const prototype = PNG.sync.read(prototypeBuffer);
  const width = Math.max(formal.width, prototype.width);
  const height = Math.max(formal.height, prototype.height);
  const normalizedFormal = padPng(formal, width, height);
  const normalizedPrototype = padPng(prototype, width, height);
  const diff = new PNG({ width, height });
  const changed = pixelmatch(
    normalizedFormal.data,
    normalizedPrototype.data,
    diff.data,
    width,
    height,
    { threshold: 0.1, includeAA: false },
  );
  const changedRatio = changed / (width * height);
  if (
    changedRatio > maxChangedRatio ||
    formal.width !== prototype.width ||
    formal.height !== prototype.height
  ) {
    await Promise.all([
      testInfo.attach(`${label}-formal.png`, { body: formalBuffer, contentType: "image/png" }),
      testInfo.attach(`${label}-prototype.png`, { body: prototypeBuffer, contentType: "image/png" }),
      testInfo.attach(`${label}-diff.png`, { body: PNG.sync.write(diff), contentType: "image/png" }),
    ]);
  }
  expect(
    { width: formal.width, height: formal.height },
    `${label} screenshot dimensions`,
  ).toEqual({ width: prototype.width, height: prototype.height });
  expect(
    changedRatio,
    `${label} changed-pixel ratio ${changedRatio.toFixed(6)} exceeds ${maxChangedRatio}`,
  ).toBeLessThanOrEqual(maxChangedRatio);
  return changedRatio;
}

function fixtureResponse(relativePath: string, scenario: string) {
  const fixture = JSON.parse(
    readFileSync(resolve(REPO_ROOT, relativePath), "utf8"),
  ) as OperationFixture;
  const response = fixture.scenarios[scenario]?.response;
  if (!response) throw new Error(`missing fixture ${relativePath}#${scenario}`);
  return response;
}

async function fulfillFixture(
  route: import("@playwright/test").Route,
  relativePath: string,
  scenario: string,
) {
  const response = fixtureResponse(relativePath, scenario);
  await route.fulfill({
    status: response.status,
    headers: {
      "content-type": "application/json; charset=utf-8",
      ...(response.headers ?? {}),
    },
    body: JSON.stringify(response.body),
  });
}

function padPng(source: PNG, width: number, height: number) {
  if (source.width === width && source.height === height) return source;
  const padded = new PNG({ width, height, fill: true });
  for (let y = 0; y < source.height; y += 1) {
    source.data.copy(
      padded.data,
      y * width * 4,
      y * source.width * 4,
      (y + 1) * source.width * 4,
    );
  }
  return padded;
}
