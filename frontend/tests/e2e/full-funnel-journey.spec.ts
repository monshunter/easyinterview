import fs from "node:fs/promises";

import { expect, test, type Page, type Request } from "@playwright/test";

interface ScenarioState {
  apiBaseUrl: string;
  frontendOrigin: string;
  userId: string;
  userEmail: string;
  resumeAssetId: string;
  sessionCookieName: string;
  sessionCookieValue: string;
}

interface NetworkCall {
  method: string;
  path: string;
  body: unknown;
}

const ARTIFACT_ROOT =
  "../.test-output/e2e/p0-099-full-funnel-fullstack-ui-journey";
const STATE_PATH = process.env.EI_E2E_STATE_PATH ?? `${ARTIFACT_ROOT}/state.json`;

const PRIVATE_JD = [
  "Scenario confidential JD",
  "for backend platform migrations",
  "with React handoff ownership",
].join(" ");
const PRIVATE_ANSWER = [
  "I split migration risk by owner",
  "then shipped the backend slice behind a measurable rollout gate",
].join(" ");
const PRIVACY_NEEDLES = [PRIVATE_JD, PRIVATE_ANSWER, "add tradeoff"];

test("E2E.P0.099 full funnel import to next-round practice", async ({
  page,
  context,
}) => {
  const state = await waitForScenarioState();
  const consoleMessages: string[] = [];
  const pageErrors: string[] = [];
  const networkCalls: NetworkCall[] = [];

  page.on("console", (msg) => consoleMessages.push(msg.text()));
  page.on("pageerror", (err) => pageErrors.push(err.message));
  page.on("request", (request) => {
    const call = captureAPICall(state, request);
    if (call) networkCalls.push(call);
  });

  await context.addCookies([
    {
      name: state.sessionCookieName,
      value: state.sessionCookieValue,
      domain: new URL(state.apiBaseUrl).hostname,
      path: "/",
      httpOnly: true,
      sameSite: "Lax",
    },
  ]);

  await page.goto(`/?resumeVersionId=${encodeURIComponent(state.resumeAssetId)}`);
  await expect(page.getByTestId("home-jd-textarea")).toBeVisible();
  await page.getByTestId("home-jd-textarea").fill(PRIVATE_JD);

  const importResponse = page.waitForResponse(
    (response) =>
      apiPath(state, response.url()) === "/targets/import" &&
      response.request().method() === "POST" &&
      response.status() === 202,
  );
  await page.getByTestId("home-jd-submit").click();
  await importResponse;

  await expect(page.getByTestId("route-parse")).toBeVisible();
  await expect(page.locator("[data-testid^='parse-loading-step-']")).toHaveCount(4);
  await expect(page.getByTestId("parse-action-confirm")).toBeVisible({
    timeout: 20_000,
  });

  await page.getByTestId("parse-action-confirm").click();
  await expect(page.getByTestId("workspace-cta-start")).toBeVisible({
    timeout: 10_000,
  });

  await page.getByTestId("workspace-cta-start").click();
  const practice = page.getByTestId("practice-screen");
  await expect(practice).toBeVisible({ timeout: 20_000 });
  const firstSessionId = await practice.getAttribute("data-session-id");
  const firstPlanId = await practice.getAttribute("data-plan-id");
  expect(firstSessionId).toBeTruthy();
  expect(firstPlanId).toBeTruthy();

  await page.getByTestId("practice-input-textarea").fill(PRIVATE_ANSWER);
  const answerResponse = page.waitForResponse(
    (response) =>
      apiPath(state, response.url()).endsWith("/events") &&
      response.request().method() === "POST" &&
      response.status() === 200,
  );
  await page.getByTestId("practice-input-send").click();
  await answerResponse;

  const completeResponse = page.waitForResponse(
    (response) =>
      apiPath(state, response.url()).endsWith("/complete") &&
      response.request().method() === "POST" &&
      response.status() === 202,
  );
  await page.getByTestId("practice-rightpanel-cta-finish").click();
  await completeResponse;
  await expect(page.getByTestId("generating-screen")).toBeVisible({
    timeout: 10_000,
  });
  await expect(page.getByTestId("report-dashboard")).toBeVisible({
    timeout: 60_000,
  });

  const reportUrl = new URL(page.url());
  const reportId = reportUrl.searchParams.get("reportId");
  expect(reportUrl.pathname).toBe("/report");
  expect(reportId).toBeTruthy();
  expect(reportUrl.searchParams.get("sessionId")).toBe(firstSessionId);

  const nextPlanRequest = page.waitForRequest((request) => {
    if (apiPath(state, request.url()) !== "/practice/plans") return false;
    if (request.method() !== "POST") return false;
    return (parseJSON(request.postData()) as { goal?: string } | null)?.goal === "next_round";
  });
  const nextStartRequest = page.waitForRequest(
    (request) =>
      apiPath(state, request.url()) === "/practice/sessions" &&
      request.method() === "POST",
  );
  await page.getByTestId("report-next-cta").click();
  const nextPlanBody = parseJSON((await nextPlanRequest).postData()) as {
    goal?: string;
    sourceReportId?: string;
  };
  await nextStartRequest;

  expect(nextPlanBody.goal).toBe("next_round");
  expect(nextPlanBody.sourceReportId).toBe(reportId);

  await expect(page.getByTestId("practice-screen")).toBeVisible({
    timeout: 20_000,
  });
  const nextPractice = page.getByTestId("practice-screen");
  const secondSessionId = await nextPractice.getAttribute("data-session-id");
  const secondPlanId = await nextPractice.getAttribute("data-plan-id");
  expect(secondSessionId).toBeTruthy();
  expect(secondPlanId).toBeTruthy();
  expect(secondSessionId).not.toBe(firstSessionId);
  expect(secondPlanId).not.toBe(firstPlanId);

  const nextPracticeUrl = new URL(page.url());
  expect(nextPracticeUrl.pathname).toBe("/practice");
  expect(nextPracticeUrl.searchParams.get("practiceGoal")).toBe("next_round");
  expect(nextPracticeUrl.searchParams.get("sourceReportId")).toBe(reportId);
  expect(nextPracticeUrl.searchParams.get("sessionId")).toBe(secondSessionId);
  expect(nextPracticeUrl.searchParams.get("planId")).toBe(secondPlanId);

  const createPlanBodies = networkCalls
    .filter((call) => call.method === "POST" && call.path === "/practice/plans")
    .map((call) => call.body as { goal?: string; sourceReportId?: string });
  expect(createPlanBodies.some((body) => body.goal === "baseline")).toBe(true);
  expect(
    createPlanBodies.some(
      (body) =>
        body.goal === "next_round" && body.sourceReportId === reportId,
    ),
  ).toBe(true);
  expect(
    networkCalls.filter(
      (call) => call.method === "POST" && call.path === "/practice/sessions",
    ).length,
  ).toBeGreaterThanOrEqual(2);

  await expectNoBrowserPrivacyLeak(page, consoleMessages, PRIVACY_NEEDLES);
  expect(pageErrors).toEqual([]);
});

async function waitForScenarioState(): Promise<ScenarioState> {
  const deadline = Date.now() + 30_000;
  let lastError: unknown = null;
  while (Date.now() < deadline) {
    try {
      const raw = await fs.readFile(STATE_PATH, "utf8");
      return JSON.parse(raw) as ScenarioState;
    } catch (err) {
      lastError = err;
      await new Promise((resolve) => setTimeout(resolve, 250));
    }
  }
  throw new Error(`state file not ready at ${STATE_PATH}: ${String(lastError)}`);
}

function captureAPICall(state: ScenarioState, request: Request): NetworkCall | null {
  const path = apiPath(state, request.url());
  if (!path) return null;
  return {
    method: request.method(),
    path,
    body: parseJSON(request.postData()),
  };
}

function apiPath(state: ScenarioState, url: string): string {
  const base = new URL(state.apiBaseUrl);
  const target = new URL(url);
  if (target.origin !== base.origin) return "";
  if (!target.pathname.startsWith(base.pathname)) return "";
  return target.pathname.slice(base.pathname.length) || "/";
}

function parseJSON(raw: string | null): unknown {
  if (!raw) return null;
  try {
    return JSON.parse(raw);
  } catch {
    return null;
  }
}

async function expectNoBrowserPrivacyLeak(
  page: Page,
  consoleMessages: string[],
  needles: string[],
): Promise<void> {
  const leaked = await page.evaluate((tokens) => {
    const surfaces: string[] = [window.location.href];
    for (const storage of [window.localStorage, window.sessionStorage]) {
      for (let i = 0; i < storage.length; i += 1) {
        const key = storage.key(i);
        if (!key) continue;
        surfaces.push(key, storage.getItem(key) ?? "");
      }
    }
    return tokens.filter((token) =>
      surfaces.some((surface) =>
        surface.toLowerCase().includes(token.toLowerCase()),
      ),
    );
  }, needles);

  const consoleLeaks = needles.filter((token) =>
    consoleMessages.some((message) =>
      message.toLowerCase().includes(token.toLowerCase()),
    ),
  );

  expect([...new Set([...leaked, ...consoleLeaks])]).toEqual([]);
}
