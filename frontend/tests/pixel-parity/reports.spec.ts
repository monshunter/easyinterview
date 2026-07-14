import { expect, test, type Page } from "@playwright/test";
import { readFileSync } from "node:fs";
import { resolve } from "node:path";

import {
  configureDeterministicPage,
  expectPixelParity,
  expectSurfaceParity,
  normalizedText,
  settleVisualSurface,
  surfaceSnapshot,
} from "./report-parity-helpers";

const ROOT = "[data-testid='reports-screen']";
const TARGET_JOB_ID = "01918fa0-0000-7000-8000-000000002000";
const OTHER_TARGET_JOB_ID = "01918fa0-0000-7000-8000-000000002099";

type ReportsDemoState =
  | "ready"
  | "loading"
  | "empty"
  | "error"
  | "latest-ready"
  | "mismatch";

interface FixtureFile {
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

test.use({ deviceScaleFactor: 1, locale: "zh-CN", timezoneId: "UTC" });

function fixtureResponse(relativePath: string, scenario = "default") {
  const fixture = JSON.parse(
    readFileSync(resolve(process.cwd(), "..", relativePath), "utf8"),
  ) as FixtureFile;
  const response = fixture.scenarios[scenario]?.response;
  if (!response) throw new Error(`missing fixture ${relativePath}#${scenario}`);
  return structuredClone(response);
}

async function fulfillFixture(
  route: import("@playwright/test").Route,
  relativePath: string,
  scenario = "default",
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

function reportsTargetJob(targetJobId = TARGET_JOB_ID) {
  const target = fixtureResponse(
    "openapi/fixtures/TargetJobs/getTargetJob.json",
    "prototype-baseline",
  ).body;
  target.id = targetJobId;
  target.companyName = "星环科技";
  target.title = "资深前端工程师";
  return target;
}

function readyOverview(targetJobId = TARGET_JOB_ID) {
  return {
    targetJobId,
    rounds: [
      {
        round: { roundId: "round-1-hr", roundSequence: 1 },
        currentReport: null,
        latestAttempt: null,
      },
      {
        round: { roundId: "round-2-technical", roundSequence: 2 },
        currentReport: {
          id: "01918fa0-0000-7000-8000-000000007021",
          generatedAt: "2026-07-13T14:20:00Z",
        },
        latestAttempt: {
          id: "01918fa0-0000-7000-8000-000000007022",
          status: "failed",
          errorCode: "AI_PROVIDER_TIMEOUT",
          createdAt: "2026-07-14T09:12:00Z",
        },
      },
      {
        round: { roundId: "round-3-technical", roundSequence: 3 },
        currentReport: null,
        latestAttempt: {
          id: "01918fa0-0000-7000-8000-000000007023",
          status: "generating",
          errorCode: null,
          createdAt: "2026-07-14T09:16:00Z",
        },
      },
      {
        round: { roundId: "round-4-manager", roundSequence: 4 },
        currentReport: {
          id: "01918fa0-0000-7000-8000-000000007024",
          generatedAt: "2026-07-14T09:20:00Z",
        },
        latestAttempt: {
          id: "01918fa0-0000-7000-8000-000000007024",
          status: "ready",
          errorCode: null,
          createdAt: "2026-07-14T09:18:00Z",
        },
      },
    ],
  };
}

function overviewFor(state: ReportsDemoState) {
  const ready = readyOverview();
  if (state === "empty") {
    return {
      targetJobId: TARGET_JOB_ID,
      rounds: ready.rounds.map((item) => ({
        round: item.round,
        currentReport: null,
        latestAttempt: null,
      })),
    };
  }
  if (state === "latest-ready") {
    ready.rounds[1]!.latestAttempt = {
      id: "01918fa0-0000-7000-8000-000000007025",
      status: "ready",
      errorCode: null,
      createdAt: "2026-07-14T10:20:00Z",
    };
    return ready;
  }
  if (state === "mismatch") {
    return { ...ready, targetJobId: OTHER_TARGET_JOB_ID };
  }
  return ready;
}

async function mockFormalReportsApis(page: Page, state: ReportsDemoState) {
  let releaseResponse = () => {};
  const responseGate = new Promise<void>((resolveResponse) => {
    releaseResponse = resolveResponse;
  });
  const reportRequests: string[] = [];

  await page.route("**/api/v1/**", async (route) => {
    const url = new URL(route.request().url());
    const path = url.pathname.replace(/^\/api\/v1/, "");
    const method = route.request().method();
    if (path === "/runtime-config") {
      await fulfillFixture(route, "openapi/fixtures/Auth/getRuntimeConfig.json");
      return;
    }
    if (path === "/me") {
      await fulfillFixture(route, "openapi/fixtures/Auth/getMe.json", "authenticated");
      return;
    }
    if (method === "GET" && /^\/targets\/[^/]+\/reports$/.test(path)) {
      reportRequests.push(path);
      if (state === "loading" || state === "error") await responseGate;
      if (state === "loading") {
        await route.abort("failed");
        return;
      }
      if (state === "error") {
        await route.fulfill({
          status: 503,
          contentType: "application/json; charset=utf-8",
          body: JSON.stringify({
            error: { code: "RESOURCE_NOT_FOUND", message: "unavailable" },
          }),
        });
        return;
      }
      await route.fulfill({
        status: 200,
        contentType: "application/json; charset=utf-8",
        body: JSON.stringify(overviewFor(state)),
      });
      return;
    }
    if (method === "GET" && /^\/targets\/[^/]+$/.test(path)) {
      await route.fulfill({
        status: 200,
        contentType: "application/json; charset=utf-8",
        body: JSON.stringify(reportsTargetJob()),
      });
      return;
    }
    await route.fulfill({
      status: 404,
      contentType: "application/json; charset=utf-8",
      body: JSON.stringify({ error: { code: "NOT_FOUND", message: path } }),
    });
  });

  return { releaseResponse, reportRequests };
}

async function expectNoReportHistory(surface: Page) {
  const root = surface.locator(ROOT);
  await expect(root).not.toContainText(/完整历史|历史版本|全部报告|Report Center/iu);
  const audit = await root.evaluate((node) => {
    const elements = [node, ...node.querySelectorAll("*")];
    return {
      text: node.textContent ?? "",
      attributes: elements.flatMap((element) =>
        Array.from(element.attributes, ({ name, value }) => `${name}=${value}`),
      ),
    };
  });
  const joined = `${audit.text}\n${audit.attributes.join("\n")}`;
  for (const sentinel of [
    "01918fa0-0000-7000-8000-000000007022",
    "01918fa0-0000-7000-8000-000000007023",
    "01918fa0-0000-7000-8000-000000007025",
    OTHER_TARGET_JOB_ID,
  ]) {
    expect(joined).not.toContain(sentinel);
  }
}

async function expectReportsSurfaceParity(
  formal: Page,
  prototype: Page,
  stateSelector: string,
  label: string,
) {
  expect(await normalizedText(formal, ROOT)).toBe(
    await normalizedText(prototype, ROOT),
  );
  const surfaces = [
    {
      label: `${label} root`,
      selector: ROOT,
      properties: ["max-width", "margin-left", "margin-right", "padding", "font-family"],
    },
    {
      label: `${label} back`,
      selector: "[data-testid='reports-back-button']",
      properties: ["border", "background-color", "color", "margin-bottom", "padding", "font-family"],
    },
    {
      label: `${label} header`,
      selector: `${ROOT} > header`,
      properties: ["display", "justify-content", "align-items", "gap", "flex-wrap", "margin-bottom"],
    },
    {
      label: `${label} title`,
      selector: `${ROOT} > header h1`,
      // The formal token keeps Source Serif Pro as an engineering fallback,
      // while the prototype declares Noto Serif SC -> Georgia directly. The
      // rendered primary family is identical; bbox and pixels still gate the
      // actual typography.
      properties: ["font-size", "line-height", "color", "margin", "overflow-wrap"],
    },
    {
      label: `${label} card`,
      selector: `${ROOT} > div`,
      properties: ["background-color", "border", "border-radius", "padding"],
    },
    {
      label: `${label} state`,
      selector: stateSelector,
      properties: ["display", "color", "font-size", "line-height", "padding"],
    },
  ] as const;
  for (const surface of surfaces) {
    const [formalSnapshot, prototypeSnapshot] = await Promise.all([
      surfaceSnapshot(formal, surface.selector, surface.properties),
      surfaceSnapshot(prototype, surface.selector, surface.properties),
    ]);
    expectSurfaceParity(formalSnapshot, prototypeSnapshot, surface.label);
  }
  const [formalPrimarySerif, prototypePrimarySerif] = await Promise.all([
    formal.locator(`${ROOT} > header h1`).evaluate((node) =>
      getComputedStyle(node).fontFamily.split(",")[0]?.trim(),
    ),
    prototype.locator(`${ROOT} > header h1`).evaluate((node) =>
      getComputedStyle(node).fontFamily.split(",")[0]?.trim(),
    ),
  ]);
  expect(formalPrimarySerif).toBe(prototypePrimarySerif);
}

test.describe("current-plan reports source, geometry, and screenshot parity", () => {
  test("current-plan reports ready state matches the UI truth", async ({
    page,
    context,
  }, testInfo) => {
    const prototype = await context.newPage();
    await Promise.all([
      configureDeterministicPage(page, "zh"),
      configureDeterministicPage(prototype, "zh"),
    ]);
    const { reportRequests } = await mockFormalReportsApis(page, "ready");

    await Promise.all([
      page.goto(`/reports?targetJobId=${TARGET_JOB_ID}&section=reports&reportId=route-only`),
      prototype.goto(
        "/ui-design/#route=reports&targetJobId=tj-1&lang=zh&signedIn=1&reportState=ready",
      ),
    ]);
    await Promise.all([
      page.locator("[data-testid='reports-round-4']").waitFor({ timeout: 8_000 }),
      prototype.locator("[data-testid='reports-round-4']").waitFor({ timeout: 8_000 }),
    ]);
    await Promise.all([settleVisualSurface(page), settleVisualSurface(prototype)]);

    expect(new URL(page.url()).search).toBe(`?targetJobId=${TARGET_JOB_ID}`);
    expect(reportRequests).toEqual([`/targets/${TARGET_JOB_ID}/reports`]);
    await expect(page.locator("[data-testid^='reports-round-']")).toHaveCount(4);
    await expect(prototype.locator("[data-testid^='reports-round-']")).toHaveCount(4);
    await expect(page.locator("[data-testid='reports-current']")).toHaveCount(2);
    await expect(page.locator("[data-testid='reports-generating']")).toHaveCount(1);
    await expect(page.locator("[data-testid='reports-failed']")).toHaveCount(1);
    await expect(page.locator("[data-testid='reports-latest-ready']")).toHaveCount(0);
    await expect(page.locator("[data-testid='reports-list'] button")).toHaveCount(3);
    await expect(page.locator("[data-testid='reports-round-4'] button")).toHaveCount(1);
    await expect(page.locator("[data-testid='topbar-nav-reports']")).toHaveCount(0);
    await expectNoReportHistory(page);
    await expectNoReportHistory(prototype);
    await expectReportsSurfaceParity(
      page,
      prototype,
      "[data-testid='reports-round-2']",
      "ready reports",
    );

    const viewport = page.viewportSize();
    expect(viewport).not.toBeNull();
    for (const surface of [page, prototype]) {
      expect(
        await surface.evaluate(() => document.documentElement.scrollWidth),
      ).toBeLessThanOrEqual(viewport!.width);
    }
    const changedRatio = await expectPixelParity(
      page,
      prototype,
      ROOT,
      testInfo,
      `reports-ready-${testInfo.project.name}`,
    );
    const formalScreenshotPath = testInfo.outputPath(
      `reports-ready-formal-${testInfo.project.name}.png`,
    );
    const prototypeScreenshotPath = testInfo.outputPath(
      `reports-ready-prototype-${testInfo.project.name}.png`,
    );
    await Promise.all([
      page.locator(ROOT).screenshot({ path: formalScreenshotPath, animations: "disabled" }),
      prototype.locator(ROOT).screenshot({ path: prototypeScreenshotPath, animations: "disabled" }),
    ]);
    await Promise.all([
      testInfo.attach(`reports-ready-formal-${testInfo.project.name}`, {
        path: formalScreenshotPath,
        contentType: "image/png",
      }),
      testInfo.attach(`reports-ready-prototype-${testInfo.project.name}`, {
        path: prototypeScreenshotPath,
        contentType: "image/png",
      }),
    ]);

    await page.locator("[data-testid='reports-back-button']").click();
    await page.waitForURL(/\/parse\?targetJobId=/);
    expect(new URL(page.url()).pathname + new URL(page.url()).search).toBe(
      `/parse?targetJobId=${TARGET_JOB_ID}`,
    );
    console.log(
      `E2E.P0.059 current-plan reports ready state matches the UI truth project=${testInfo.project.name} viewport=${viewport!.width}x${viewport!.height} rounds=4 current=2 generating=1 failed=1 currentPlanIsolation=true currentLatestOnly=true topbarReportsEntry=0 backTarget=parse changedRatio=${changedRatio.toFixed(6)}`,
    );
    await prototype.close();
  });

  test("reports loading empty error latest-ready and mismatch states match the UI truth", async ({
    context,
  }, testInfo) => {
    for (const state of [
      "loading",
      "empty",
      "error",
      "latest-ready",
      "mismatch",
    ] as const) {
      const formal = await context.newPage();
      const prototype = await context.newPage();
      let releaseResponse = () => {};
      try {
        await Promise.all([
          configureDeterministicPage(formal, "zh"),
          configureDeterministicPage(prototype, "zh"),
        ]);
        const mocked = await mockFormalReportsApis(formal, state);
        releaseResponse = mocked.releaseResponse;
        await Promise.all([
          formal.goto(`/reports?targetJobId=${TARGET_JOB_ID}`),
          prototype.goto(
            `/ui-design/#route=reports&targetJobId=tj-1&lang=zh&signedIn=1&reportState=${state}`,
          ),
        ]);
        await formal
          .locator("[data-testid='reports-target-title']")
          .waitFor({ timeout: 8_000 });
        await expect(formal.locator("[data-testid='reports-target-title']")).toHaveText(
          "星环科技 · 资深前端工程师",
        );
        if (state === "error") releaseResponse();

        const stateSelector =
          state === "latest-ready"
            ? "[data-testid='reports-round-4']"
            : state === "mismatch"
              ? "[data-testid='reports-error']"
              : `[data-testid='reports-${state}']`;
        await Promise.all([
          formal.locator(stateSelector).waitFor({ timeout: 8_000 }),
          prototype.locator(stateSelector).waitFor({ timeout: 8_000 }),
        ]);
        await Promise.all([
          settleVisualSurface(formal),
          settleVisualSurface(prototype),
        ]);

        expect(mocked.reportRequests).toEqual([
          `/targets/${TARGET_JOB_ID}/reports`,
        ]);
        if (state === "mismatch") {
          await expect(formal.locator("[data-testid='reports-list']")).toHaveCount(0);
          await expect(prototype.locator("[data-testid='reports-list']")).toHaveCount(0);
        }
        if (state === "latest-ready") {
          await expect(
            formal.locator("[data-testid='reports-round-2'] [data-testid='reports-latest-ready']"),
          ).toHaveCount(1);
          await expect(
            prototype.locator("[data-testid='reports-round-2'] [data-testid='reports-latest-ready']"),
          ).toHaveCount(1);
          await expect(
            formal.locator("[data-testid='reports-round-2'] [data-testid='reports-generating']"),
          ).toHaveCount(0);
          await expect(formal.locator("[data-testid='reports-generating']")).toHaveCount(1);
          await expect(formal.locator("[data-testid='reports-list'] button")).toHaveCount(3);
        }
        if (state === "empty" || state === "loading" || state === "error" || state === "mismatch") {
          await expect(formal.locator("[data-testid='reports-list']")).toHaveCount(0);
        }
        await expectNoReportHistory(formal);
        await expectNoReportHistory(prototype);
        await expectReportsSurfaceParity(
          formal,
          prototype,
          stateSelector,
          `${state} reports`,
        );

        const viewport = formal.viewportSize();
        expect(viewport).not.toBeNull();
        for (const surface of [formal, prototype]) {
          expect(
            await surface.evaluate(() => document.documentElement.scrollWidth),
          ).toBeLessThanOrEqual(viewport!.width);
        }
        const changedRatio = await expectPixelParity(
          formal,
          prototype,
          ROOT,
          testInfo,
          `reports-${state}-${testInfo.project.name}`,
        );
        console.log(
          `E2E.P0.059 reports state=${state} project=${testInfo.project.name} viewport=${viewport!.width}x${viewport!.height} currentPlanIsolation=${state === "mismatch" ? "true" : "verified"} currentLatestOnly=true changedRatio=${changedRatio.toFixed(6)}`,
        );
      } finally {
        releaseResponse();
        await Promise.all([formal.close(), prototype.close()]);
      }
    }
  });
});
