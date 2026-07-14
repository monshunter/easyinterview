import { expect, test } from "@playwright/test";
import { readFileSync, writeFileSync } from "node:fs";
import { resolve } from "node:path";

import {
  configureDeterministicPage,
  expectFullPagePixelParity,
  expectPixelParity,
  expectSurfaceParity,
  normalizedText,
  settleVisualSurface,
  surfaceSnapshot,
} from "./report-parity-helpers";

interface OperationFixture {
  scenarios: Record<string, { response: { status: number; headers?: Record<string, string>; body: unknown } }>;
}

const PLAN_ID = "01918fa0-0000-7000-8000-000000004000";
const TARGET_JOB_ID = "01918fa0-0000-7000-8000-000000002000";
const RESUME_ID = "01918fa0-0000-7000-8000-000000001000";
const PRACTICE_ROOT = "[data-testid='practice-screen']";
type ReplyStateDemo = "immediate-pending" | "persisted-pending" | "retryable-failed" | "terminal-failed";

interface PracticeMockOptions {
  sessionScenario?: string;
  sendScenario?: string;
  beforeSend?: (request: import("@playwright/test").Request) => Promise<void> | void;
  onMessagePost?: (request: import("@playwright/test").Request) => Promise<void> | void;
}

function fixtureResponse(relativePath: string, scenario = "default") {
  const fixture = JSON.parse(readFileSync(resolve(process.cwd(), "..", relativePath), "utf8")) as OperationFixture;
  const response = fixture.scenarios[scenario]?.response;
  if (!response) throw new Error(`missing fixture scenario ${relativePath}#${scenario}`);
  return response;
}

async function fulfillFixture(route: import("@playwright/test").Route, relativePath: string, scenario = "default") {
  const response = fixtureResponse(relativePath, scenario);
  await route.fulfill({
    status: response.status,
    headers: { "content-type": "application/json; charset=utf-8", ...(response.headers ?? {}) },
    body: JSON.stringify(response.body),
  });
}

async function mockPracticeApis(
  page: import("@playwright/test").Page,
  options: PracticeMockOptions = {},
) {
  await page.route("**/api/v1/**", async (route) => {
    const path = new URL(route.request().url()).pathname.replace(/^\/api\/v1/, "");
    if (path === "/runtime-config") return fulfillFixture(route, "openapi/fixtures/Auth/getRuntimeConfig.json");
    if (path === "/me") return fulfillFixture(route, "openapi/fixtures/Auth/getMe.json", "authenticated");
    if (/^\/practice\/sessions\/[^/]+$/.test(path)) return fulfillFixture(route, "openapi/fixtures/PracticeSessions/getPracticeSession.json", options.sessionScenario);
    if (/^\/practice\/plans\/[^/]+$/.test(path)) return fulfillFixture(route, "openapi/fixtures/PracticePlans/getPracticePlan.json");
    if (/^\/targets\/[^/]+$/.test(path)) return fulfillFixture(route, "openapi/fixtures/TargetJobs/getTargetJob.json");
    if (/^\/practice\/sessions\/[^/]+\/messages$/.test(path)) {
      expect(route.request().method()).toBe("POST");
      await options.onMessagePost?.(route.request());
      await options.beforeSend?.(route.request());
      return fulfillFixture(route, "openapi/fixtures/PracticeSessions/sendPracticeMessage.json", options.sendScenario);
    }
    if (/^\/practice\/sessions\/[^/]+\/complete$/.test(path)) return fulfillFixture(route, "openapi/fixtures/PracticeSessions/completePracticeSession.json");
    await route.fulfill({ status: 404, headers: { "content-type": "application/json" }, body: JSON.stringify({ error: { code: "NOT_FOUND", message: path } }) });
  });
}

async function goToPractice(
  page: import("@playwright/test").Page,
  options: PracticeMockOptions = {},
) {
  await configureDeterministicPage(page, "zh");
  await mockPracticeApis(page, options);
  const session = fixtureResponse(
    "openapi/fixtures/PracticeSessions/getPracticeSession.json",
    options.sessionScenario,
  ).body as { id: string };
  await page.addInitScript((route) => {
    (window as Window & { __EASYINTERVIEW_INITIAL_ROUTE__?: { name: string; params: Record<string, string> } }).__EASYINTERVIEW_INITIAL_ROUTE__ = route;
  }, {
    name: "practice",
    params: {
      sessionId: session.id,
      planId: PLAN_ID,
      targetJobId: TARGET_JOB_ID,
      jdId: `jd-${TARGET_JOB_ID}`,
      resumeId: RESUME_ID,
      roundId: "round-technical-1",
      roundName: "用人经理",
      practiceGoal: "baseline",
    },
  });
  await page.goto("/");
  await page.waitForSelector("[data-testid='practice-screen']");
  await expect(page.getByTestId("practice-topbar-company")).toHaveText("Acme");
  await expect(page.getByTestId("practice-topbar-title")).toHaveText("Senior Frontend Engineer");
  await expect(page.getByTestId("practice-topbar-timer")).toContainText("/ 50:00");
}

async function goToPrototypePractice(
  page: import("@playwright/test").Page,
  replyState: ReplyStateDemo,
) {
  await configureDeterministicPage(page, "zh");
  const params = new URLSearchParams({
    route: "practice",
    replyState,
    targetJobId: TARGET_JOB_ID,
    planId: PLAN_ID,
    jdId: `jd-${TARGET_JOB_ID}`,
    resumeId: RESUME_ID,
    roundId: "round-technical-1",
    roundName: "用人经理",
    lang: "zh",
    nochrome: "1",
  });
  await page.goto(`/ui-design/#${params.toString()}`);
  await page.waitForSelector(PRACTICE_ROOT);
  await expect(page.getByTestId("practice-topbar-company")).toHaveText("Acme");
  await expect(page.getByTestId("practice-topbar-title")).toHaveText("Senior Frontend Engineer");
  await expect(page.getByTestId("practice-topbar-timer")).toContainText("/ 50:00");
}

async function expectPracticeStateDomAndViewport(
  page: import("@playwright/test").Page,
  stateSelector: string,
) {
  for (const id of [
    "practice-screen",
    "practice-topbar",
    "practice-conversation",
    "practice-transcript",
    "practice-input",
    "practice-input-textarea",
    "practice-input-send",
  ]) {
    await expect(page.getByTestId(id), id).toHaveCount(1);
  }
  await expect(page.locator(stateSelector), stateSelector).toHaveCount(1);
  const viewport = page.viewportSize();
  expect(viewport).toBeTruthy();
  const geometry = await page.locator(PRACTICE_ROOT).evaluate((node) => {
    const box = node.getBoundingClientRect();
    return {
      width: box.width,
      height: box.height,
      scrollWidth: document.documentElement.scrollWidth,
      innerWidth: window.innerWidth,
      innerHeight: window.innerHeight,
    };
  });
  expect([geometry.width, geometry.height]).toEqual([viewport!.width, viewport!.height]);
  expect([geometry.innerWidth, geometry.innerHeight]).toEqual([viewport!.width, viewport!.height]);
  expect(geometry.scrollWidth).toBeLessThanOrEqual(viewport!.width);
}

async function expectPracticeStateCopyParity(
  formal: import("@playwright/test").Page,
  prototype: import("@playwright/test").Page,
  stateSelector: string,
) {
  for (const selector of [
    "[data-testid='practice-transcript']",
    "[data-testid='practice-finish-disabled-reason']",
    stateSelector,
  ]) {
    expect(await normalizedText(formal, selector), `${selector} copy`).toBe(
      await normalizedText(prototype, selector),
    );
  }
  const [formalPlaceholder, prototypePlaceholder, formalAriaLabel, prototypeAriaLabel] = await Promise.all([
    formal.getByTestId("practice-input-textarea").getAttribute("placeholder"),
    prototype.getByTestId("practice-input-textarea").getAttribute("placeholder"),
    formal.locator(stateSelector).getAttribute("aria-label"),
    prototype.locator(stateSelector).getAttribute("aria-label"),
  ]);
  expect(formalPlaceholder, "practice textarea placeholder").toBe(prototypePlaceholder);
  expect(formalAriaLabel, `${stateSelector} aria-label`).toBe(prototypeAriaLabel);
}

async function practiceDomA11ySnapshot(
  page: import("@playwright/test").Page,
  replyState: ReplyStateDemo,
) {
  const root = page.locator(PRACTICE_ROOT);
  const ariaSnapshot = (await root.ariaSnapshot())
    .replace(/\b\d{2}:\d{2}\b/gu, "<time>")
    .replace(/\s+$/gmu, "");
  const dom = await root.evaluate((rootNode, expectedReplyState) => {
    const normalize = (value: string | null) => (value ?? "")
      .replace(/\b\d{2}:\d{2}\b/gu, "<time>")
      .replace(/\s+/gu, " ")
      .trim();
    const implicitRole = (element: Element) => {
      const explicit = element.getAttribute("role");
      if (explicit) return explicit;
      switch (element.tagName.toLowerCase()) {
        case "button": return "button";
        case "textarea": return "textbox";
        case "main": return "main";
        default: return null;
      }
    };
    const accessibleName = (element: Element, role: string | null) => {
      const ariaLabel = element.getAttribute("aria-label");
      if (ariaLabel) return normalize(ariaLabel);
      const labelledBy = element.getAttribute("aria-labelledby");
      if (labelledBy) {
        return normalize(labelledBy.split(/\s+/u).map((id) => document.getElementById(id)?.textContent ?? "").join(" "));
      }
      if (role === "textbox") return normalize(element.getAttribute("placeholder"));
      return role ? normalize(element.textContent) : "";
    };
    const elements = [rootNode, ...rootNode.querySelectorAll("*")];
    return {
      dataState: {
        replyState: expectedReplyState,
        thinking: Boolean(rootNode.querySelector("[data-testid='practice-interviewer-thinking']")),
        retryable: Boolean(rootNode.querySelector("[data-testid='practice-message-retry']")),
        terminal: Boolean(rootNode.querySelector("[data-testid='practice-terminal-recovery']")),
        composerDisabled: (rootNode.querySelector("[data-testid='practice-input-textarea']") as HTMLTextAreaElement | null)?.disabled ?? null,
        sendDisabled: (rootNode.querySelector("[data-testid='practice-input-send']") as HTMLButtonElement | null)?.disabled ?? null,
        finishDisabled: (rootNode.querySelector("[data-testid='practice-finish-cta']") as HTMLButtonElement | null)?.disabled ?? null,
      },
      elements: elements.map((element) => {
        const role = implicitRole(element);
        const parent = element.parentElement;
        let depth = 0;
        for (let current = parent; current && current !== rootNode.parentElement; current = current.parentElement) depth += 1;
        return {
          tag: element.tagName.toLowerCase(),
          depth,
          childCount: element.children.length,
          className: normalize(element.getAttribute("class")),
          testId: element.getAttribute("data-testid"),
          role,
          accessibleName: accessibleName(element, role),
          ariaLive: element.getAttribute("aria-live"),
          ariaDescribedBy: element.getAttribute("aria-describedby"),
          ariaHidden: element.getAttribute("aria-hidden"),
          disabled: "disabled" in element ? Boolean((element as HTMLButtonElement | HTMLTextAreaElement).disabled) : null,
          placeholder: element.getAttribute("placeholder"),
          value: "value" in element ? String((element as HTMLTextAreaElement | HTMLButtonElement).value ?? "") : null,
          text: normalize(element.textContent),
          dataState: {
            state: element.getAttribute("data-state"),
            replyState: element.getAttribute("data-reply-state"),
            role: element.getAttribute("data-role"),
            ariaPressed: element.getAttribute("aria-pressed"),
            ariaDisabled: element.getAttribute("aria-disabled"),
          },
        };
      }),
    };
  }, replyState);
  return { ariaSnapshot, dom };
}

async function expectPracticeDomA11yParity(
  formal: import("@playwright/test").Page,
  prototype: import("@playwright/test").Page,
  replyState: ReplyStateDemo,
) {
  const [formalSnapshot, prototypeSnapshot] = await Promise.all([
    practiceDomA11ySnapshot(formal, replyState),
    practiceDomA11ySnapshot(prototype, replyState),
  ]);
  expect(formalSnapshot.dom, `${replyState} normalized DOM snapshot`).toEqual(prototypeSnapshot.dom);
  expect(formalSnapshot.ariaSnapshot, `${replyState} accessibility snapshot`).toBe(prototypeSnapshot.ariaSnapshot);
}

async function absoluteSurfaceSnapshot(
  page: import("@playwright/test").Page,
  selector: string,
  properties: readonly string[],
) {
  return surfaceSnapshot(page, selector, properties);
}

async function expectPracticeCoreSurfaceParity(
  formal: import("@playwright/test").Page,
  prototype: import("@playwright/test").Page,
  replyState: ReplyStateDemo,
  stateSelectors: readonly string[],
) {
  const surfaces: Array<{ label: string; selector: string; properties: readonly string[] }> = [
    { label: "root", selector: PRACTICE_ROOT, properties: ["display", "flex-direction", "width", "height", "overflow", "background-color"] },
    { label: "topbar", selector: "[data-testid='practice-topbar']", properties: ["display", "align-items", "flex-wrap", "gap", "padding-top", "padding-right", "padding-bottom", "padding-left", "border-bottom-width", "border-bottom-color", "background-color"] },
    { label: "conversation", selector: "[data-testid='practice-conversation']", properties: ["display", "flex-direction", "width", "min-height"] },
    { label: "transcript", selector: "[data-testid='practice-transcript']", properties: ["overflow-x", "overflow-y", "padding-top", "padding-right", "padding-bottom", "padding-left"] },
    { label: "transcript helper", selector: "[data-testid='practice-transcript-helper']", properties: ["text-align", "margin-top", "font-family", "font-size", "color"] },
    { label: "composer", selector: "[data-testid='practice-input']", properties: ["padding-top", "padding-right", "padding-bottom", "padding-left", "border-top-width", "border-top-color", "background-color"] },
    { label: "textarea", selector: "[data-testid='practice-input-textarea']", properties: ["width", "min-height", "border-top-width", "resize", "font-family", "font-size", "line-height", "background-color", "color"] },
    { label: "send", selector: "[data-testid='practice-input-send']", properties: ["display", "padding-top", "padding-right", "padding-bottom", "padding-left", "border-radius", "font-size", "font-weight", "cursor", "opacity", "background-color", "border-color", "color"] },
    { label: "finish", selector: "[data-testid='practice-finish-cta']", properties: ["display", "padding-top", "padding-right", "padding-bottom", "padding-left", "border-radius", "font-family", "font-size", "font-weight", "cursor", "background-color", "border-color", "color"] },
    { label: "finish reason", selector: "[data-testid='practice-finish-disabled-reason']", properties: ["max-width", "font-size", "line-height", "text-align", "color"] },
  ];
  stateSelectors.forEach((selector, index) => surfaces.push({
    label: `state ${index + 1}`,
    selector,
    properties: ["display", "align-items", "justify-content", "flex-wrap", "gap", "width", "height", "margin-top", "margin-bottom", "padding", "border-radius", "border-color", "background-color", "color", "font-size"],
  }));
  const [formalMessageCount, prototypeMessageCount] = await Promise.all([
    formal.locator("[data-testid^='practice-transcript-message-']").count(),
    prototype.locator("[data-testid^='practice-transcript-message-']").count(),
  ]);
  expect(formalMessageCount, `${replyState} transcript row count`).toBe(prototypeMessageCount);
  for (let index = 0; index < formalMessageCount; index += 1) {
    surfaces.push({
      label: `transcript row ${index}`,
      selector: `[data-testid='practice-transcript-message-${index}']`,
      properties: ["display", "gap", "margin-bottom", "font-size", "color"],
    });
  }
  for (const surface of surfaces) {
    await Promise.all([
      expect(formal.locator(surface.selector), `${replyState} formal ${surface.label}`).toHaveCount(1),
      expect(prototype.locator(surface.selector), `${replyState} prototype ${surface.label}`).toHaveCount(1),
    ]);
    const [formalSurface, prototypeSurface] = await Promise.all([
      absoluteSurfaceSnapshot(formal, surface.selector, surface.properties),
      absoluteSurfaceSnapshot(prototype, surface.selector, surface.properties),
    ]);
    expectSurfaceParity(formalSurface, prototypeSurface, `${replyState} ${surface.label}`);
  }
}

test.use({ locale: "zh-CN", timezoneId: "UTC" });

async function attachStateScreenshot(
  page: import("@playwright/test").Page,
  testInfo: import("@playwright/test").TestInfo,
  label: string,
) {
  await settleVisualSurface(page);
  const screenshotFile = `${label}-${testInfo.project.name}.png`;
  const metadataFile = `${label}-${testInfo.project.name}.metadata.json`;
  const path = testInfo.outputPath(screenshotFile);
  const image = await page.screenshot({ path, fullPage: false, animations: "disabled" });
  expect(image.length, `${label} screenshot`).toBeGreaterThan(10_000);
  const viewport = page.viewportSize();
  if (!viewport) throw new Error(`${label} requires a configured CSS viewport`);
  const deviceScaleFactor = await page.evaluate(() => window.devicePixelRatio);
  const metadataPath = testInfo.outputPath(metadataFile);
  writeFileSync(metadataPath, `${JSON.stringify({
    screenshot_file: screenshotFile,
    css_viewport: [viewport.width, viewport.height],
    device_scale_factor: deviceScaleFactor,
  }, null, 2)}\n`, "utf8");
  await testInfo.attach(`${label}-${testInfo.project.name}`, {
    path,
    contentType: "image/png",
  });
  await testInfo.attach(`${label}-${testInfo.project.name}-metadata`, {
    path: metadataPath,
    contentType: "application/json",
  });
}

test.describe("practice continuous conversation parity", () => {
  test("renders one full-width chat with no structured-question surfaces", async ({ page }) => {
    await goToPractice(page);
    for (const id of [
      "practice-topbar",
      "practice-topbar-phone-toggle",
      "practice-conversation",
      "practice-transcript",
      "practice-input",
      "practice-input-textarea",
      "practice-finish-cta",
    ]) {
      await expect(page.locator(`[data-testid='${id}']`), id).toHaveCount(1);
    }
    for (const stale of ["practice-sessionmap", "practice-question", "practice-question-prompt", "practice-phone-surface"]) {
      await expect(page.locator(`[data-testid='${stale}']`), stale).toHaveCount(0);
    }
    for (const debugAttribute of ["data-session-id", "data-plan-id", "data-target-job-id"]) {
      await expect(page.getByTestId("practice-screen"), debugAttribute).not.toHaveAttribute(debugAttribute);
    }
    await expect(page.locator("[data-testid='practice-topbar-phone-toggle']")).toBeDisabled();
  });

  test("conversation remains inside desktop and mobile viewports", async ({ page }) => {
    await goToPractice(page);
    const viewport = page.viewportSize();
    expect(viewport).toBeTruthy();
    const geometry = await page.locator("[data-testid='practice-conversation']").evaluate((node) => {
      const rect = node.getBoundingClientRect();
      return { left: rect.left, right: rect.right, width: rect.width, scrollWidth: document.documentElement.scrollWidth };
    });
    expect(geometry.left).toBeGreaterThanOrEqual(-1);
    expect(geometry.right).toBeLessThanOrEqual(viewport!.width + 1);
    expect(geometry.width).toBeGreaterThan(viewport!.width * 0.9);
    expect(geometry.scrollWidth).toBeLessThanOrEqual(viewport!.width);
  });

  test("disabled phone control keeps the prototype geometry", async ({ page }) => {
    await goToPractice(page);
    const style = await page.locator("[data-testid='practice-topbar-phone-toggle']").evaluate((node) => {
      const computed = getComputedStyle(node);
      return { width: computed.width, height: computed.height, borderRadius: computed.borderRadius, cursor: computed.cursor };
    });
    expect(style).toEqual({ width: "34px", height: "34px", borderRadius: "17px", cursor: "not-allowed" });
  });

  test("screenshot smoke is non-empty", async ({ page }) => {
    await goToPractice(page);
    const image = await page.screenshot({ fullPage: false });
    expect(image.length).toBeGreaterThan(10_000);
  });

  test("new user input is visible before the reply and locks the composer", async ({ page, context }, testInfo) => {
    const prototype = await context.newPage();
    const requests: Array<{ clientMessageId: string; text: string }> = [];
    let releaseSend: (() => void) | undefined;
    const sendGate = new Promise<void>((resolveSend) => {
      releaseSend = resolveSend;
    });
    await goToPractice(page, {
      beforeSend: async (request) => {
        requests.push(request.postDataJSON() as { clientMessageId: string; text: string });
        await sendGate;
      },
    });

    const answer = "我会先建立可回滚基线，再逐步放量。";
    try {
      await page.getByTestId("practice-input-textarea").fill(answer);
      await page.getByTestId("practice-input-send").click();

      await expect.poll(() => requests.length).toBe(1);
      await expect(page.getByText(answer, { exact: true })).toHaveCount(1);
      await expect(page.getByTestId("practice-input-textarea")).toHaveValue("");
      await expect(page.getByTestId("practice-input-textarea")).toBeDisabled();
      await expect(page.getByTestId("practice-input-textarea")).toHaveAttribute("placeholder", "面试官正在思考…");
      await expect(page.getByTestId("practice-input-send")).toBeDisabled();
      await expect(page.getByTestId("practice-finish-cta")).toBeDisabled();
      await expect(page.getByTestId("practice-interviewer-thinking")).toBeVisible();
      await expect(page.getByTestId("practice-message-retry")).toHaveCount(0);
      expect(requests[0]?.text).toBe(answer);
      expect(requests[0]?.clientMessageId).toMatch(/^[0-9a-f-]{36}$/u);

      await goToPrototypePractice(prototype, "immediate-pending");
      await Promise.all([settleVisualSurface(page), settleVisualSurface(prototype)]);
      const stateSelector = "[data-testid='practice-interviewer-thinking']";
      await Promise.all([
        expectPracticeStateDomAndViewport(page, stateSelector),
        expectPracticeStateDomAndViewport(prototype, stateSelector),
      ]);
      await expectPracticeStateCopyParity(page, prototype, stateSelector);
      await expectPracticeDomA11yParity(page, prototype, "immediate-pending");
      await expectPracticeCoreSurfaceParity(page, prototype, "immediate-pending", [stateSelector]);
      const [formalThinking, prototypeThinking] = await Promise.all([
        surfaceSnapshot(page, stateSelector, ["display", "gap", "margin-bottom", "color", "font-size"], PRACTICE_ROOT),
        surfaceSnapshot(prototype, stateSelector, ["display", "gap", "margin-bottom", "color", "font-size"], PRACTICE_ROOT),
      ]);
      expectSurfaceParity(formalThinking, prototypeThinking, "practice immediate pending thinking");
      await expectPixelParity(page, prototype, stateSelector, testInfo, `practice-immediate-pending-${testInfo.project.name}`);
      await expectFullPagePixelParity(page, prototype, testInfo, `practice-immediate-pending-full-${testInfo.project.name}`);

      await attachStateScreenshot(page, testInfo, "practice-immediate-pending");

      releaseSend?.();
      releaseSend = undefined;
      await expect(page.getByTestId("practice-interviewer-thinking")).toHaveCount(0);
    } finally {
      releaseSend?.();
      await prototype.close();
    }
  });

  test("reloads a persisted pending reply, keeps all actions locked, and sends zero POSTs", async ({ page, context }, testInfo) => {
    const prototype = await context.newPage();
    let messagePostCount = 0;
    await goToPractice(page, {
      sessionScenario: "reply-pending",
      onMessagePost: () => { messagePostCount += 1; },
    });
    await expect(page.getByTestId("practice-interviewer-thinking")).toBeVisible();
    await page.reload();
    await page.waitForSelector(PRACTICE_ROOT);

    const thinking = page.getByTestId("practice-interviewer-thinking");
    await expect(thinking).toBeVisible();
    await expect(thinking).toHaveAttribute("role", "status");
    await expect(thinking).toHaveAttribute("aria-live", "polite");
    await expect(page.getByTestId("practice-input-textarea")).toBeDisabled();
    await expect(page.getByTestId("practice-input-send")).toBeDisabled();
    await expect(page.getByTestId("practice-input-send")).toHaveCSS("cursor", "not-allowed");
    await expect(page.getByTestId("practice-input-send")).toHaveCSS("opacity", "0.5");
    await expect(page.getByTestId("practice-finish-cta")).toBeDisabled();
    await expect(page.getByTestId("practice-message-retry")).toHaveCount(0);
    await expect(page.getByTestId("practice-finish-disabled-reason")).toHaveText("请等待面试官回复。");
    expect(messagePostCount).toBe(0);

    const geometry = await thinking.evaluate((node) => {
      const computed = getComputedStyle(node);
      const box = node.getBoundingClientRect();
      return {
        display: computed.display,
        gap: computed.gap,
        marginBottom: computed.marginBottom,
        height: box.height,
      };
    });
    expect({
      display: geometry.display,
      gap: geometry.gap,
      marginBottom: geometry.marginBottom,
    }).toEqual({ display: "flex", gap: "12px", marginBottom: "18px" });
    expect(geometry.height).toBeCloseTo(28, 3);
    await expect(page.getByTestId("practice-input-textarea")).toHaveCSS("min-height", "74px");

    await goToPrototypePractice(prototype, "persisted-pending");
    await Promise.all([settleVisualSurface(page), settleVisualSurface(prototype)]);
    const stateSelector = "[data-testid='practice-interviewer-thinking']";
    await Promise.all([
      expectPracticeStateDomAndViewport(page, stateSelector),
      expectPracticeStateDomAndViewport(prototype, stateSelector),
    ]);
    await expectPracticeStateCopyParity(page, prototype, stateSelector);
    await expectPracticeDomA11yParity(page, prototype, "persisted-pending");
    await expectPracticeCoreSurfaceParity(page, prototype, "persisted-pending", [stateSelector]);
    const [formalThinking, prototypeThinking] = await Promise.all([
      surfaceSnapshot(page, stateSelector, ["display", "gap", "margin-bottom", "color", "font-size"], PRACTICE_ROOT),
      surfaceSnapshot(prototype, stateSelector, ["display", "gap", "margin-bottom", "color", "font-size"], PRACTICE_ROOT),
    ]);
    expectSurfaceParity(formalThinking, prototypeThinking, "practice persisted pending thinking");
    await expectPixelParity(page, prototype, stateSelector, testInfo, `practice-persisted-pending-${testInfo.project.name}`);
    await expectFullPagePixelParity(page, prototype, testInfo, `practice-persisted-pending-full-${testInfo.project.name}`);

    await attachStateScreenshot(page, testInfo, "practice-persisted-pending");
    expect(messagePostCount).toBe(0);
    await prototype.close();
  });

  test("retryable failure exposes one row-local retry and preserves the next draft", async ({ page, context }, testInfo) => {
    const prototype = await context.newPage();
    const requests: Array<{ clientMessageId: string; text: string }> = [];
    let releaseRetry: (() => void) | undefined;
    const retryGate = new Promise<void>((resolveRetry) => {
      releaseRetry = resolveRetry;
    });
    await goToPractice(page, {
      sessionScenario: "reply-retryable-failed",
      sendScenario: "retry-success-same-client-message",
      beforeSend: async (request) => {
        requests.push(request.postDataJSON() as { clientMessageId: string; text: string });
        await retryGate;
      },
    });

    const retry = page.getByTestId("practice-message-retry");
    const textarea = page.getByTestId("practice-input-textarea");
    const nextDraft = "这是下一条尚未提交的草稿。";
    try {
      await expect(retry).toHaveCount(1);
      await expect(retry).toHaveAttribute("aria-label", "重试这条消息");
      await expect(textarea).toBeEnabled();
      await expect(page.getByTestId("practice-input-send")).toBeDisabled();
      await expect(page.getByTestId("practice-finish-cta")).toBeDisabled();
      await expect(page.getByTestId("practice-interviewer-thinking")).toHaveCount(0);
      await textarea.fill(nextDraft);

      const retryStyle = await retry.evaluate((node) => {
        const computed = getComputedStyle(node);
        return {
          width: computed.width,
          height: computed.height,
          marginTop: computed.marginTop,
          borderRadius: computed.borderRadius,
          padding: computed.padding,
          display: computed.display,
        };
      });
      expect(retryStyle).toEqual({
        width: "28px",
        height: "28px",
        marginTop: "7px",
        borderRadius: "2px",
        padding: "0px",
        display: "inline-flex",
      });
      await goToPrototypePractice(prototype, "retryable-failed");
      await prototype.getByTestId("practice-input-textarea").fill(nextDraft);
      await Promise.all([
        expect(textarea).toHaveValue(nextDraft),
        expect(prototype.getByTestId("practice-input-textarea")).toHaveValue(nextDraft),
      ]);
      await Promise.all([settleVisualSurface(page), settleVisualSurface(prototype)]);
      const stateSelector = "[data-testid='practice-message-retry']";
      await Promise.all([
        expectPracticeStateDomAndViewport(page, stateSelector),
        expectPracticeStateDomAndViewport(prototype, stateSelector),
      ]);
      await expectPracticeStateCopyParity(page, prototype, stateSelector);
      await expectPracticeDomA11yParity(page, prototype, "retryable-failed");
      await expectPracticeCoreSurfaceParity(page, prototype, "retryable-failed", [stateSelector]);
      const [formalRetry, prototypeRetry] = await Promise.all([
        surfaceSnapshot(page, stateSelector, ["display", "width", "height", "margin-top", "padding", "border-radius", "border-color", "background-color", "color"]),
        surfaceSnapshot(prototype, stateSelector, ["display", "width", "height", "margin-top", "padding", "border-radius", "border-color", "background-color", "color"]),
      ]);
      expectSurfaceParity(formalRetry, prototypeRetry, "practice retryable failed row retry");
      await expectPixelParity(page, prototype, stateSelector, testInfo, `practice-retryable-failed-${testInfo.project.name}`);
      await expectFullPagePixelParity(page, prototype, testInfo, `practice-retryable-failed-full-${testInfo.project.name}`);
      await attachStateScreenshot(page, testInfo, "practice-retryable-failed");

      await retry.click();
      await expect.poll(() => requests.length).toBe(1);
      await expect(textarea).toHaveValue(nextDraft);
      await expect(textarea).toBeDisabled();
      await expect(page.getByTestId("practice-interviewer-thinking")).toBeVisible();
      await expect(retry).toHaveCount(0);
      expect(requests[0]).toEqual({
        clientMessageId: "01918fa0-0000-7000-8000-000000007010",
        text: "我先把风险拆成三类。",
      });

      releaseRetry?.();
      releaseRetry = undefined;
      await expect(page.getByTestId("practice-interviewer-thinking")).toHaveCount(0);
      await expect(textarea).toBeEnabled();
      await expect(textarea).toHaveValue(nextDraft);
      await expect(page.getByText("我先把风险拆成三类。", { exact: true })).toHaveCount(1);
      await expect(page.getByText("哪一类风险最容易被团队低估？", { exact: true })).toHaveCount(1);
      await expect(retry).toHaveCount(0);
    } finally {
      releaseRetry?.();
      await prototype.close();
    }
  });

  test("terminal failure has no retry escape hatch and keeps the interview locked", async ({ page, context }, testInfo) => {
    const prototype = await context.newPage();
    await goToPractice(page, { sessionScenario: "reply-terminal-failed" });

    const recovery = page.getByTestId("practice-terminal-recovery");
    const cta = page.getByTestId("practice-terminal-recovery-cta");
    await expect(recovery).toHaveCount(1);
    await expect(recovery).toHaveAttribute("role", "alert");
    await expect(cta).toHaveCount(1);
    await expect(cta).toHaveText("返回当前面试规划");
    await expect(page.getByRole("button", { name: "返回当前面试规划", exact: true })).toHaveCount(1);
    await expect(page.getByText("返回当前面试规划", { exact: true })).toHaveCount(1);
    await expect(page.getByTestId("practice-error-state")).toHaveCount(0);
    await expect(page.getByTestId("practice-message-retry")).toHaveCount(0);
    await expect(page.getByTestId("practice-interviewer-thinking")).toHaveCount(0);
    await expect(page.getByTestId("practice-input-textarea")).toBeDisabled();
    await expect(page.getByTestId("practice-input-send")).toBeDisabled();
    await expect(page.getByTestId("practice-finish-cta")).toBeDisabled();
    await expect(page.getByTestId("practice-finish-disabled-reason")).toHaveText("请先恢复这条未完成回复的消息。");

    const parity = await cta.evaluate((node) => {
      const computed = getComputedStyle(node);
      const box = node.getBoundingClientRect();
      const recoveryBox = node.closest("[data-testid='practice-terminal-recovery']")!.getBoundingClientRect();
      return {
        style: {
          display: computed.display,
          alignItems: computed.alignItems,
          justifyContent: computed.justifyContent,
          gap: computed.gap,
          height: computed.height,
          padding: computed.padding,
          fontSize: computed.fontSize,
          fontWeight: computed.fontWeight,
          backgroundColor: computed.backgroundColor,
          color: computed.color,
          borderColor: computed.borderColor,
          borderRadius: computed.borderRadius,
          cursor: computed.cursor,
          opacity: computed.opacity,
          letterSpacing: computed.letterSpacing,
          transitionDuration: computed.transitionDuration,
          transitionProperty: computed.transitionProperty,
        },
        box: { left: box.left, right: box.right, top: box.top, bottom: box.bottom, width: box.width, height: box.height },
        recoveryBox: { left: recoveryBox.left, right: recoveryBox.right, width: recoveryBox.width },
        viewport: {
          width: window.innerWidth,
          height: window.innerHeight,
          scrollWidth: document.documentElement.scrollWidth,
        },
      };
    });
    expect(parity.style).toEqual({
      display: "flex",
      alignItems: "center",
      justifyContent: "center",
      gap: "8px",
      height: "30px",
      padding: "0px 12px",
      fontSize: "13px",
      fontWeight: "500",
      backgroundColor: "rgb(248, 250, 253)",
      color: "rgb(20, 24, 33)",
      borderColor: "rgb(221, 226, 236)",
      borderRadius: "2px",
      cursor: "pointer",
      opacity: "1",
      letterSpacing: "-0.065px",
      transitionDuration: "0.08s, 0.15s",
      transitionProperty: "transform, opacity",
    });
    expect(parity.box.height).toBeCloseTo(30, 3);
    expect(parity.box.width).toBeGreaterThan(100);
    expect(parity.box.left).toBeGreaterThanOrEqual(-1);
    expect(parity.box.right).toBeLessThanOrEqual(parity.viewport.width + 1);
    expect(parity.box.top).toBeGreaterThanOrEqual(-1);
    expect(parity.box.bottom).toBeLessThanOrEqual(parity.viewport.height + 1);
    expect(parity.recoveryBox.left).toBeGreaterThanOrEqual(-1);
    expect(parity.recoveryBox.right).toBeLessThanOrEqual(parity.viewport.width + 1);
    expect(parity.recoveryBox.width).toBeLessThanOrEqual(parity.viewport.width + 1);
    expect(parity.viewport.scrollWidth).toBeLessThanOrEqual(parity.viewport.width);
    expect([[1440, 900], [390, 844]]).toContainEqual([parity.viewport.width, parity.viewport.height]);

    await goToPrototypePractice(prototype, "terminal-failed");
    await Promise.all([settleVisualSurface(page), settleVisualSurface(prototype)]);
    const stateSelector = "[data-testid='practice-terminal-recovery']";
    await Promise.all([
      expectPracticeStateDomAndViewport(page, stateSelector),
      expectPracticeStateDomAndViewport(prototype, stateSelector),
    ]);
    await expectPracticeStateCopyParity(page, prototype, stateSelector);
    await expectPracticeDomA11yParity(page, prototype, "terminal-failed");
    await expectPracticeCoreSurfaceParity(page, prototype, "terminal-failed", [
      stateSelector,
      "[data-testid='practice-terminal-recovery-cta']",
    ]);
    const [formalRecovery, prototypeRecovery] = await Promise.all([
      surfaceSnapshot(page, stateSelector, ["display", "align-items", "justify-content", "gap", "margin-bottom", "padding", "border-radius", "border-color", "background-color"], PRACTICE_ROOT),
      surfaceSnapshot(prototype, stateSelector, ["display", "align-items", "justify-content", "gap", "margin-bottom", "padding", "border-radius", "border-color", "background-color"], PRACTICE_ROOT),
    ]);
    expectSurfaceParity(formalRecovery, prototypeRecovery, "practice terminal failed recovery");
    await expectPixelParity(page, prototype, stateSelector, testInfo, `practice-terminal-failed-${testInfo.project.name}`);
    await expectFullPagePixelParity(page, prototype, testInfo, `practice-terminal-failed-full-${testInfo.project.name}`);

    await attachStateScreenshot(page, testInfo, "practice-terminal-failed");

    await cta.click();
    await expect.poll(() => new URL(page.url()).pathname).toBe("/parse");
    const url = new URL(page.url());
    expect(url.pathname).toBe("/parse");
    expect(url.pathname).not.toBe("/workspace");
    expect(url.searchParams.get("targetJobId")).toBe(TARGET_JOB_ID);
    expect(url.searchParams.has("planId")).toBe(false);
    expect([...url.searchParams.keys()]).toEqual(["targetJobId"]);
    expect(url.hash).toBe("");
    await prototype.close();
  });
});
