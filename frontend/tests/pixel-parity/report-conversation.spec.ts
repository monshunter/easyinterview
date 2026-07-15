import { expect, test, type Page, type Route } from "@playwright/test";
import { readFileSync } from "node:fs";
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

const ROOT = "[data-testid='report-conversation-screen']";
const CONVERSATION_FIXTURE = "openapi/fixtures/Reports/getReportConversation.json";
const REPO_ROOT = resolve(process.cwd(), "..");

interface FixtureResponse {
  status: number;
  headers?: Record<string, string>;
  body: Record<string, unknown>;
}

interface OperationFixture {
  scenarios: Record<string, { response: FixtureResponse }>;
}

test.use({ deviceScaleFactor: 1, locale: "zh-CN", timezoneId: "UTC" });

function conversationFixture(scenario: string): FixtureResponse {
  const fixture = JSON.parse(
    readFileSync(resolve(REPO_ROOT, CONVERSATION_FIXTURE), "utf8"),
  ) as OperationFixture;
  const response = fixture.scenarios[scenario]?.response;
  if (!response) throw new Error(`missing ${CONVERSATION_FIXTURE}#${scenario}`);
  return response;
}

async function fulfillFixture(
  route: Route,
  relativePath: string,
  scenario: string,
) {
  const fixture = JSON.parse(
    readFileSync(resolve(REPO_ROOT, relativePath), "utf8"),
  ) as OperationFixture;
  const response = fixture.scenarios[scenario]?.response;
  if (!response) throw new Error(`missing ${relativePath}#${scenario}`);
  await route.fulfill({
    status: response.status,
    headers: {
      "content-type": "application/json; charset=utf-8",
      ...(response.headers ?? {}),
    },
    body: JSON.stringify(response.body),
  });
}

async function mockFormalConversationApi(page: Page, scenario: string) {
  await page.route("**/api/v1/**", async (route) => {
    const url = new URL(route.request().url());
    if (/\/reports\/[^/]+\/conversation$/.test(url.pathname)) {
      if (scenario === "loading") {
        await new Promise<void>(() => undefined);
        return;
      }
      return fulfillFixture(route, CONVERSATION_FIXTURE, scenario);
    }
    if (url.pathname.endsWith("/runtime/config")) {
      return fulfillFixture(route, "openapi/fixtures/Auth/getRuntimeConfig.json", "default");
    }
    if (url.pathname.endsWith("/me")) {
      return fulfillFixture(route, "openapi/fixtures/Auth/getMe.json", "default");
    }
    return route.fulfill({ status: 204, body: "" });
  });
}

async function injectPrototypeConversationFixture(page: Page, scenario: string) {
  const response = conversationFixture(scenario);
  if (response.status !== 200) {
    throw new Error(`prototype conversation parity requires 200 fixture, got ${response.status}`);
  }
  await injectPrototypeConversationState(page, { state: "ready", ...response.body });
}

async function injectPrototypeConversationState(page: Page, conversation: Record<string, unknown>) {
  await page.route("**/ui-design/src/data.jsx*", async (route) => {
    const sourceResponse = await route.fetch();
    const source = await sourceResponse.text();
    await route.fulfill({
      response: sourceResponse,
      contentType: "application/javascript; charset=utf-8",
      body: `${source}\nwindow.EI_DATA.reportConversation = ${JSON.stringify(conversation)};`,
    });
  });
}

async function expectNoInternalLocators(surface: Page, locators: readonly string[]) {
  const root = surface.locator(ROOT);
  const audit = await root.evaluate((node) => {
    const elements = [node, ...node.querySelectorAll("*")];
    return {
      text: node.textContent ?? "",
      attributes: elements.flatMap((element) =>
        Array.from(element.attributes, ({ name, value }) => `${name}=${value}`),
      ),
    };
  });
  const accessible = await root.ariaSnapshot();
  for (const locator of locators) {
    expect(audit.text).not.toContain(locator);
    expect(audit.attributes.join("\n")).not.toContain(locator);
    expect(accessible).not.toContain(locator);
  }
}

function compactProjectionText(value: string) {
  return value.replace(/\*\*/g, "").replace(/\s+/g, "");
}

async function openConversationSurfaces(
  formal: Page,
  prototype: Page,
  scenario: string,
  reportId: string,
) {
  await Promise.all([
    configureDeterministicPage(formal, "zh"),
    configureDeterministicPage(prototype, "zh"),
  ]);
  await mockFormalConversationApi(formal, scenario);
  await injectPrototypeConversationFixture(prototype, scenario);
  await Promise.all([
    formal.goto(`/report-conversation?reportId=${reportId}&sessionId=must-not-survive`),
    prototype.goto(`/ui-design/#route=report-conversation&reportId=${reportId}&lang=zh&signedIn=1`),
  ]);
  await Promise.all([formal.locator(ROOT).waitFor(), prototype.locator(ROOT).waitFor()]);
  await Promise.all([settleVisualSurface(formal), settleVisualSurface(prototype)]);
}

test.describe("report conversation source, geometry, and screenshot parity", () => {
  test("canonical report-owned transcript mirrors the UI truth and preserves local mobile scroll", async ({
    page,
    context,
  }, testInfo) => {
    const prototype = await context.newPage();
    const response = conversationFixture("prototype-baseline");
    const reportId = String(response.body.reportId);
    const projection = response.body as {
      context: { sourcePlanId: string; resumeId: string };
      messages: unknown[];
    };

    await openConversationSurfaces(page, prototype, "prototype-baseline", reportId);
    expect(new URL(page.url()).search).toBe(`?reportId=${reportId}`);
    await Promise.all([
      expectNoInternalLocators(page, [reportId, projection.context.sourcePlanId, projection.context.resumeId]),
      expectNoInternalLocators(prototype, [reportId, projection.context.sourcePlanId, projection.context.resumeId]),
    ]);
    await Promise.all([
      expect(page.locator("[data-testid='report-conversation-context-strip'] [data-testid='report-context-strip']")).toHaveCount(1),
      expect(prototype.locator("[data-testid='report-conversation-context-strip'] [data-testid='report-context-strip']")).toHaveCount(1),
      expect(page.locator("[data-testid^='report-conversation-message-']")).toHaveCount(projection.messages.length),
      expect(prototype.locator("[data-testid^='report-conversation-message-']")).toHaveCount(projection.messages.length),
      expect(page.locator("textarea, [data-testid='practice-input'], [data-testid*='voice']")).toHaveCount(0),
      expect(prototype.locator("textarea, [data-testid='practice-input'], [data-testid*='voice']")).toHaveCount(0),
    ]);
    expect(compactProjectionText(await normalizedText(page, ROOT))).toBe(
      compactProjectionText(await normalizedText(prototype, ROOT)),
    );
    for (const surface of [page, prototype]) {
      await expect(surface.locator("[data-testid='report-conversation-message-1'] h2")).toHaveText("技术取舍追问");
      await expect(surface.locator("[data-testid='report-conversation-message-1'] p")).toHaveText("请先说明约束、候选方案，以及最终选择的理由。");
      await expect(surface.locator("[data-testid='report-conversation-message-2'] strong")).toHaveCount(1);
      await expect(surface.locator("[data-testid='report-conversation-message-2'] table")).toHaveCount(1);
      await expect(surface.locator("[data-testid='report-conversation-message-2'] pre")).toHaveCount(1);
    }

    const surfaces: ReadonlyArray<{
      label: string;
      selector: string;
      properties: readonly string[];
      relativeTo?: string;
    }> = [
      {
        label: "conversation root absolute viewport geometry",
        selector: ROOT,
        properties: ["max-width", "padding-top", "padding-right", "padding-bottom", "padding-left"],
      },
      {
        label: "frozen context strip",
        selector: "[data-testid='report-conversation-context-strip'] [data-testid='report-context-strip']",
        properties: ["display", "grid-template-columns", "gap", "border-top-width", "border-top-color", "margin-bottom"],
        relativeTo: ROOT,
      },
      {
        label: "readonly transcript boundary",
        selector: "[data-testid='report-conversation-transcript']",
        properties: ["min-width", "overflow-x"],
        relativeTo: ROOT,
      },
      {
        label: "assistant message row",
        selector: "[data-testid='report-conversation-message-1']",
        properties: ["display", "gap", "min-width", "margin-bottom"],
        relativeTo: ROOT,
      },
      {
        label: "assistant identity badge",
        selector: "[data-testid='report-conversation-message-1'] > div:first-child",
        properties: ["width", "height", "border-radius", "flex-shrink", "background-color", "color", "font-size", "font-family", "font-weight"],
        relativeTo: ROOT,
      },
      {
        label: "safe markdown projection",
        selector: "[data-testid='report-conversation-message-2'] .ei-practice-message-body",
        properties: ["min-width", "max-width", "line-height", "overflow-wrap", "word-break"],
        relativeTo: ROOT,
      },
    ];
    for (const surface of surfaces) {
      const [formal, golden] = await Promise.all([
        surfaceSnapshot(page, surface.selector, surface.properties, surface.relativeTo),
        surfaceSnapshot(prototype, surface.selector, surface.properties, surface.relativeTo),
      ]);
      expectSurfaceParity(formal, golden, surface.label);
    }

    if (testInfo.project.name === "mobile") {
      for (const surface of [page, prototype]) {
        expect(await surface.evaluate(() => document.documentElement.scrollWidth)).toBeLessThanOrEqual(390);
        for (const selector of [".ei-practice-message-body pre", ".ei-practice-message-body table"]) {
          const overflow = await surface.locator(selector).evaluate((node) => {
            const element = node as HTMLElement;
            const style = getComputedStyle(element);
            return {
              clientWidth: element.clientWidth,
              scrollWidth: element.scrollWidth,
              overflowX: style.overflowX,
              overscrollBehaviorX: style.overscrollBehaviorX,
            };
          });
          expect(overflow.scrollWidth).toBeGreaterThan(overflow.clientWidth);
          expect(overflow).toMatchObject({ overflowX: "auto", overscrollBehaviorX: "contain" });
        }
      }
    }

    await expectPixelParity(page, prototype, ROOT, testInfo, `report-conversation-${testInfo.project.name}`);
    await expectFullPagePixelParity(page, prototype, testInfo, `report-conversation-full-page-${testInfo.project.name}`);
    await prototype.close();
  });

  test("a legal empty projection keeps the frozen context and matches the prototype empty state", async ({
    page,
    context,
  }, testInfo) => {
    const prototype = await context.newPage();
    const response = conversationFixture("empty-messages");
    const reportId = String(response.body.reportId);

    await openConversationSurfaces(page, prototype, "empty-messages", reportId);
    await Promise.all([
      expect(page.locator("[data-testid='report-conversation-empty']")).toHaveCount(1),
      expect(prototype.locator("[data-testid='report-conversation-empty']")).toHaveCount(1),
      expect(page.locator("[data-testid^='report-conversation-message-']")).toHaveCount(0),
      expect(prototype.locator("[data-testid^='report-conversation-message-']")).toHaveCount(0),
    ]);
    expect(await normalizedText(page, ROOT)).toBe(await normalizedText(prototype, ROOT));

    const [formal, golden] = await Promise.all([
      surfaceSnapshot(
        page,
        "[data-testid='report-conversation-empty']",
        ["padding-top", "padding-bottom", "color", "font-size", "line-height"],
        ROOT,
      ),
      surfaceSnapshot(
        prototype,
        "[data-testid='report-conversation-empty']",
        ["padding-top", "padding-bottom", "color", "font-size", "line-height"],
        ROOT,
      ),
    ]);
    expectSurfaceParity(formal, golden, "legal empty conversation state");
    await expectPixelParity(page, prototype, ROOT, testInfo, `report-conversation-empty-${testInfo.project.name}`);
    await prototype.close();
  });

  test("loading and unavailable states remain recoverable and match the prototype", async ({
    page,
    context,
  }, testInfo) => {
    const reportId = String(conversationFixture("default").body.reportId);
    const loadingPrototype = await context.newPage();
    await Promise.all([
      configureDeterministicPage(page, "zh"),
      configureDeterministicPage(loadingPrototype, "zh"),
    ]);
    await mockFormalConversationApi(page, "loading");
    await injectPrototypeConversationState(loadingPrototype, { state: "loading" });
    await Promise.all([
      page.goto(`/report-conversation?reportId=${reportId}`),
      loadingPrototype.goto(`/ui-design/#route=report-conversation&reportId=${reportId}&lang=zh&signedIn=1`),
    ]);
    const loadingRoot = "[data-testid='report-conversation-loading']";
    await Promise.all([page.locator(loadingRoot).waitFor(), loadingPrototype.locator(loadingRoot).waitFor()]);
    await Promise.all([settleVisualSurface(page), settleVisualSurface(loadingPrototype)]);
    expect(await normalizedText(page, loadingRoot)).toBe(await normalizedText(loadingPrototype, loadingRoot));
    await Promise.all([
      expect(page.locator("[data-testid='report-conversation-loading-back']")).toHaveCount(1),
      expect(loadingPrototype.getByRole("button", { name: "返回面试" })).toHaveCount(1),
    ]);
    const [formalLoading, prototypeLoading] = await Promise.all([
      surfaceSnapshot(page, loadingRoot, ["max-width", "padding", "font-family"]),
      surfaceSnapshot(loadingPrototype, loadingRoot, ["max-width", "padding", "font-family"]),
    ]);
    expectSurfaceParity(formalLoading, prototypeLoading, "conversation loading root");
    await expectPixelParity(page, loadingPrototype, loadingRoot, testInfo, `report-conversation-loading-${testInfo.project.name}`);
    await loadingPrototype.close();

    const unavailableFormal = await context.newPage();
    const unavailablePrototype = await context.newPage();
    await Promise.all([
      configureDeterministicPage(unavailableFormal, "zh"),
      configureDeterministicPage(unavailablePrototype, "zh"),
    ]);
    await mockFormalConversationApi(unavailableFormal, "report-not-found");
    await injectPrototypeConversationState(unavailablePrototype, { state: "network_error" });
    await Promise.all([
      unavailableFormal.goto(`/report-conversation?reportId=${reportId}`),
      unavailablePrototype.goto(`/ui-design/#route=report-conversation&reportId=${reportId}&lang=zh&signedIn=1`),
    ]);
    const unavailableRoot = "[data-testid='report-conversation-unavailable']";
    await Promise.all([
      unavailableFormal.locator(unavailableRoot).waitFor(),
      unavailablePrototype.locator(unavailableRoot).waitFor(),
    ]);
    await Promise.all([settleVisualSurface(unavailableFormal), settleVisualSurface(unavailablePrototype)]);
    expect(await normalizedText(unavailableFormal, unavailableRoot)).toBe(
      await normalizedText(unavailablePrototype, unavailableRoot),
    );
    await Promise.all([
      expect(unavailableFormal.locator("[data-testid='report-conversation-unavailable-back']")).toHaveCount(1),
      expect(unavailablePrototype.getByRole("button", { name: "返回面试" })).toHaveCount(1),
    ]);
    const [formalUnavailable, prototypeUnavailable] = await Promise.all([
      surfaceSnapshot(unavailableFormal, unavailableRoot, ["max-width", "padding", "font-family"]),
      surfaceSnapshot(unavailablePrototype, unavailableRoot, ["max-width", "padding", "font-family"]),
    ]);
    expectSurfaceParity(formalUnavailable, prototypeUnavailable, "conversation unavailable root");
    await expectPixelParity(
      unavailableFormal,
      unavailablePrototype,
      unavailableRoot,
      testInfo,
      `report-conversation-unavailable-${testInfo.project.name}`,
    );
    await unavailableFormal.close();
    await unavailablePrototype.close();
  });
});
