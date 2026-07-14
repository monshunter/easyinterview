import { expect, test, type Page } from "@playwright/test";

interface MailpitMessage {
  ID?: string;
  To?: Array<{ Address?: string }>;
}

interface PracticeRoundRef {
  roundId: string;
  roundSequence: number;
}

interface TargetJobProgressResponse {
  currentPracticePlanId?: string | null;
  practiceProgress?: {
    completedRounds: PracticeRoundRef[];
    currentRound: PracticeRoundRef | null;
    status: string;
  };
}

interface PracticePlanResponse {
  id: string;
  resumeId: string;
  roundId?: string | null;
  roundSequence?: number | null;
  status: string;
  targetJobId: string;
  timeBudgetMinutes: number;
}

const FRONTEND_ORIGIN =
  process.env.EI_P0_098_FRONTEND_ORIGIN ?? "http://127.0.0.1:5173";
const API_BASE_URL =
  process.env.EI_P0_098_API_BASE_URL ?? "http://127.0.0.1:8080/api/v1";
const MAILPIT_BASE_URL =
  process.env.EI_P0_098_MAILPIT_BASE_URL ?? "http://127.0.0.1:8025";
const AUTH_EMAIL =
  process.env.EI_P0_098_AUTH_EMAIL ??
  "p0-098-live-round-refresh@example.test";
const RESUME_ID =
  process.env.EI_P0_098_RESUME_ID ??
  "019f6098-0000-7000-8000-000000000002";
const TARGET_JOB_ID =
  process.env.EI_P0_098_TARGET_JOB_ID ??
  "019f6098-0000-7000-8000-000000000003";
const ROUND_ONE_SESSION_ID =
  process.env.EI_P0_098_ROUND_ONE_SESSION_ID ??
  "019f6098-0000-7000-8000-000000000020";
const INTERCEPTED_SESSION_ID = "019f6098-0000-7000-8000-000000000090";
const WORKSPACE_DETAIL_PATH = `/workspace?targetJobId=${TARGET_JOB_ID}`;

interface ReadyDetailRequestCounts {
  getTargetJob: number;
  importTargetJob: number;
  parsePoll: number;
}

test.setTimeout(120_000);

test("E2E.P0.098 completion refreshes Workspace and quick-start posts the backend current round", async ({
  page,
}) => {
  const seenMessageIds = new Set(await listMessageIdsForEmail(AUTH_EMAIL));
  await loginExistingUser(page, AUTH_EMAIL, seenMessageIds);
  await page.evaluate(() => window.localStorage.setItem("ei-lang", "zh"));
  const readyDetailRequests = trackReadyDetailRequests(page);

  await page.goto(`${FRONTEND_ORIGIN}/workspace`, {
    waitUntil: "domcontentloaded",
  });
  const rail = page.getByTestId(`workspace-plan-list-rail-${TARGET_JOB_ID}`);
  await expect(rail).toBeVisible();
  await expectRoundStates(rail, ["current", "pending", "pending"]);
  await expect(rail).toContainText("HR · 30m");
  await expect(rail).toContainText("Technical · 30m");
  await expect(rail).toContainText("Manager · 45m");

  const initialTarget = await getTargetJob(page);
  expect(initialTarget.practiceProgress).toEqual({
    completedRounds: [],
    currentRound: { roundId: "round-1-hr", roundSequence: 1 },
    status: "not_started",
  });
  expect(initialTarget.currentPracticePlanId).toBe(
    "019f6098-0000-7000-8000-000000000010",
  );

  const completion = await page.evaluate(
    async ({ apiBaseUrl, sessionId }) => {
      const response = await fetch(
        `${apiBaseUrl}/practice/sessions/${sessionId}/complete`,
        {
          method: "POST",
          credentials: "include",
          headers: {
            "Content-Type": "application/json",
            "Idempotency-Key": `e2e-p0-098-complete-${Date.now()}`,
          },
          body: JSON.stringify({ clientCompletedAt: new Date().toISOString() }),
        },
      );
      return { status: response.status, body: await response.text() };
    },
    { apiBaseUrl: API_BASE_URL, sessionId: ROUND_ONE_SESSION_ID },
  );
  expect(completion.status, completion.body).toBe(202);
  console.log(
    "E2E.P0.098 live completion API PASS completionStatus=202 persistedFact=session_completed",
  );

  await page.reload({ waitUntil: "domcontentloaded" });
  await expect(rail).toBeVisible();
  await expectRoundStates(rail, ["done", "current", "pending"]);

  const refreshedTarget = await getTargetJob(page);
  expect(refreshedTarget.practiceProgress).toEqual({
    completedRounds: [{ roundId: "round-1-hr", roundSequence: 1 }],
    currentRound: { roundId: "round-2-technical", roundSequence: 2 },
    status: "in_progress",
  });
  expect(refreshedTarget.currentPracticePlanId ?? null).toBeNull();
  console.log(
    "E2E.P0.098 workspace refresh PASS states=done,current,pending currentRound=round-2-technical currentRoundSequence=2",
  );

  const workspaceCardVisit = readyDetailRequests.beginVisit();
  await page
    .getByTestId(`workspace-plan-list-card-body-${TARGET_JOB_ID}`)
    .click();
  await expectReadyWorkspaceDetail(page, workspaceCardVisit);
  await expectWorkspaceDetailRoundPresentation(page);
  readyDetailRequests.endVisit();

  const workspaceRefreshVisit = readyDetailRequests.beginVisit();
  await page.reload({ waitUntil: "domcontentloaded" });
  await expectReadyWorkspaceDetail(page, workspaceRefreshVisit);
  await expectWorkspaceDetailRoundPresentation(page);
  readyDetailRequests.endVisit();

  await page.goto(`${FRONTEND_ORIGIN}/`, { waitUntil: "domcontentloaded" });
  const homeRail = page.getByTestId(`home-recent-mock-rail-${TARGET_JOB_ID}`);
  await expect(homeRail).toBeVisible();
  await expectRoundStates(homeRail, ["done", "current", "pending"]);
  await page.reload({ waitUntil: "domcontentloaded" });
  await expect(homeRail).toBeVisible();
  await expectRoundStates(homeRail, ["done", "current", "pending"]);

  const homeCardVisit = readyDetailRequests.beginVisit();
  await page.getByTestId(`home-recent-mock-card-${TARGET_JOB_ID}`).click();
  await expectReadyWorkspaceDetail(page, homeCardVisit);
  await expectWorkspaceDetailRoundPresentation(page);
  readyDetailRequests.endVisit();

  const detailTarget = await getTargetJob(page);
  expect(detailTarget.practiceProgress?.currentRound).toEqual({
    roundId: "round-2-technical",
    roundSequence: 2,
  });
  console.log(
    "E2E.P0.098 home and workspace detail refresh PASS homeStates=done,current,pending detailCurrentRound=round-2-technical detailCurrentRoundSequence=2",
  );
  console.log(
    `E2E.P0.098 ready cards direct detail PASS sources=workspace,home route=${WORKSPACE_DETAIL_PATH} perVisitGetTargetJob=1 importTargetJob=0 parsePoll=0`,
  );
  console.log(
    "E2E.P0.098 workspace detail refresh PASS states=done,current,pending labels=已进行,即将进行,未进行 visualStyles=distinct",
  );

  await page.goto(`${FRONTEND_ORIGIN}/workspace`, {
    waitUntil: "domcontentloaded",
  });
  await expect(
    page.getByTestId(`workspace-plan-list-start-${TARGET_JOB_ID}`),
  ).toBeVisible();

  let interceptedStartBody: Record<string, unknown> | null = null;
  let interceptedStartIdempotencyKey = "";
  let createdPlanId = "";

  await page.route(
    new RegExp(
      `^${escapeRegExp(API_BASE_URL)}/practice/sessions$`,
    ),
    async (route) => {
      const request = route.request();
      if (request.method() !== "POST") {
        await route.continue();
        return;
      }
      interceptedStartBody = request.postDataJSON() as Record<string, unknown>;
      interceptedStartIdempotencyKey =
        (await request.headerValue("idempotency-key")) ?? "";
      const planId = String(interceptedStartBody.planId ?? "");
      createdPlanId = planId;
      const now = new Date().toISOString();
      await route.fulfill({
        status: 201,
        contentType: "application/json",
        headers: corsHeaders(),
        body: JSON.stringify({
          id: INTERCEPTED_SESSION_ID,
          planId,
          targetJobId: TARGET_JOB_ID,
          status: "waiting_user_input",
          language: "zh-CN",
          messages: [],
          createdAt: now,
          updatedAt: now,
        }),
      });
    },
  );
  await page.route(
    new RegExp(
      `^${escapeRegExp(API_BASE_URL)}/practice/sessions/${INTERCEPTED_SESSION_ID}$`,
    ),
    async (route) => {
      const now = new Date().toISOString();
      await route.fulfill({
        status: 200,
        contentType: "application/json",
        headers: corsHeaders(),
        body: JSON.stringify({
          id: INTERCEPTED_SESSION_ID,
          planId: createdPlanId,
          targetJobId: TARGET_JOB_ID,
          status: "waiting_user_input",
          language: "zh-CN",
          messages: [],
          createdAt: now,
          updatedAt: now,
        }),
      });
    },
  );

  const createPlanResponsePromise = page.waitForResponse(
    (response) =>
      response.request().method() === "POST" &&
      response.url() === `${API_BASE_URL}/practice/plans`,
  );
  const createPlanRequestPromise = page.waitForRequest(
    (request) =>
      request.method() === "POST" &&
      request.url() === `${API_BASE_URL}/practice/plans`,
  );

  await page
    .getByTestId(`workspace-plan-list-start-${TARGET_JOB_ID}`)
    .click();

  const [createPlanRequest, createPlanResponse] = await Promise.all([
    createPlanRequestPromise,
    createPlanResponsePromise,
  ]);
  expect(createPlanResponse.status()).toBe(201);
  const createPlanBody = createPlanRequest.postDataJSON() as Record<
    string,
    unknown
  >;
  const createdPlan = (await createPlanResponse.json()) as PracticePlanResponse;

  expect(createPlanBody).toMatchObject({
    goal: "baseline",
    resumeId: RESUME_ID,
    roundId: "round-2-technical",
    targetJobId: TARGET_JOB_ID,
    timeBudgetMinutes: 30,
  });
  expect(createPlanBody).not.toHaveProperty("roundSequence");
  expect(createdPlan).toMatchObject({
    resumeId: RESUME_ID,
    roundId: "round-2-technical",
    roundSequence: 2,
    status: "ready",
    targetJobId: TARGET_JOB_ID,
    timeBudgetMinutes: 30,
  });
  expect(await createPlanRequest.headerValue("idempotency-key")).not.toBe("");

  await expect.poll(() => interceptedStartBody).not.toBeNull();
  expect(createdPlanId).toBe(createdPlan.id);
  expect(interceptedStartBody).toEqual({ planId: createdPlan.id });
  expect(interceptedStartIdempotencyKey).not.toBe("");
  await expect(page).toHaveURL(/\/practice\?/);
  await expect(page.getByTestId("practice-screen")).toBeVisible();
  expect(new URL(page.url()).searchParams.get("planId")).toBe(createdPlan.id);
  expect(new URL(page.url()).searchParams.get("sessionId")).toBe(INTERCEPTED_SESSION_ID);

  const persistedPlan = await getPracticePlan(page, createdPlan.id);
  expect(persistedPlan.roundId).toBe("round-2-technical");
  expect(persistedPlan.roundSequence).toBe(2);
  console.log(
    "E2E.P0.098 next plan POST PASS requestRoundId=round-2-technical responseRoundId=round-2-technical responseRoundSequence=2 persistedRoundSequence=2",
  );
  console.log(
    "E2E.P0.098 session start interception PASS realPlanCreate=true aiSessionStart=intercepted",
  );
});

async function loginExistingUser(
  page: Page,
  email: string,
  seenMessageIds: Set<string>,
): Promise<void> {
  await page.goto(`${FRONTEND_ORIGIN}/auth/login`, {
    waitUntil: "domcontentloaded",
  });
  await page.getByTestId("auth-login-email").fill(email);
  await page.getByTestId("auth-login-submit-email").click();
  await expect(page.getByTestId("route-auth_verify")).toBeAttached();
  const code = await pollMailCode(email, seenMessageIds);
  await page.getByTestId("auth-verify-code").fill(code);
  await page.getByTestId("auth-verify-submit").click();
  await expect(page.getByTestId("topbar-user-area")).toHaveAttribute(
    "data-signed-in",
    "true",
  );
  await expect(page.getByTestId("route-auth_profile_setup")).toHaveCount(0);
}

async function getTargetJob(page: Page): Promise<TargetJobProgressResponse> {
  return page.evaluate(
    async ({ apiBaseUrl, targetJobId }) => {
      const response = await fetch(`${apiBaseUrl}/targets/${targetJobId}`, {
        credentials: "include",
      });
      const body = await response.text();
      if (!response.ok) {
        throw new Error(`getTargetJob failed: HTTP ${response.status} ${body}`);
      }
      return JSON.parse(body) as TargetJobProgressResponse;
    },
    { apiBaseUrl: API_BASE_URL, targetJobId: TARGET_JOB_ID },
  );
}

async function getPracticePlan(
  page: Page,
  planId: string,
): Promise<PracticePlanResponse> {
  return page.evaluate(
    async ({ apiBaseUrl, requestedPlanId }) => {
      const response = await fetch(
        `${apiBaseUrl}/practice/plans/${requestedPlanId}`,
        { credentials: "include" },
      );
      const body = await response.text();
      if (!response.ok) {
        throw new Error(`getPracticePlan failed: HTTP ${response.status} ${body}`);
      }
      return JSON.parse(body) as PracticePlanResponse;
    },
    { apiBaseUrl: API_BASE_URL, requestedPlanId: planId },
  );
}

async function expectRoundStates(
  rail: ReturnType<Page["getByTestId"]>,
  expected: string[],
): Promise<void> {
  await expect
    .poll(() =>
      rail
        .locator("[data-round-state]")
        .evaluateAll((nodes) =>
          nodes.map((node) => node.getAttribute("data-round-state")),
        ),
    )
    .toEqual(expected);
}

function trackReadyDetailRequests(page: Page): {
  beginVisit: () => ReadyDetailRequestCounts;
  endVisit: () => void;
} {
  let activeVisit: ReadyDetailRequestCounts | null = null;
  const targetPath = new URL(`${API_BASE_URL}/targets/${TARGET_JOB_ID}`)
    .pathname;
  const importPath = new URL(`${API_BASE_URL}/targets/import`).pathname;

  page.on("request", (request) => {
    if (!activeVisit) return;
    const path = new URL(request.url()).pathname;
    if (request.method() === "POST" && path === importPath) {
      activeVisit.importTargetJob += 1;
      return;
    }
    if (request.method() !== "GET" || path !== targetPath) return;
    activeVisit.getTargetJob += 1;
    activeVisit.parsePoll = Math.max(0, activeVisit.getTargetJob - 1);
  });

  return {
    beginVisit: () => {
      if (activeVisit) {
        throw new Error("ready detail request visit is already active");
      }
      activeVisit = {
        getTargetJob: 0,
        importTargetJob: 0,
        parsePoll: 0,
      };
      return activeVisit;
    },
    endVisit: () => {
      activeVisit = null;
    },
  };
}

async function expectReadyWorkspaceDetail(
  page: Page,
  requests: ReadyDetailRequestCounts,
): Promise<void> {
  await expect(page).toHaveURL(`${FRONTEND_ORIGIN}${WORKSPACE_DETAIL_PATH}`);
  await expect(page.getByTestId("route-workspace")).toBeVisible();
  await expect(page.getByTestId("unified-plan-detail")).toBeVisible();
  await expect(page.getByTestId("parse-action-start-interview")).toBeEnabled();
  await expect(page.getByTestId("route-parse")).toHaveCount(0);
  await expect(page.locator('[data-testid^="parse-loading-step-"]')).toHaveCount(
    0,
  );
  await expect.poll(() => requests.getTargetJob).toBe(1);
  await page.waitForTimeout(250);
  expect(requests).toEqual({
    getTargetJob: 1,
    importTargetJob: 0,
    parsePoll: 0,
  });
}

async function expectWorkspaceDetailRoundPresentation(
  page: Page,
): Promise<void> {
  const cards = page
    .getByTestId("parse-rounds")
    .locator("[data-round-state]");
  await expectRoundStates(page.getByTestId("parse-rounds"), [
    "done",
    "current",
    "pending",
  ]);
  await expect(page.getByTestId("parse-round-state-0")).toHaveText("已进行");
  await expect(page.getByTestId("parse-round-state-1")).toHaveText("即将进行");
  await expect(page.getByTestId("parse-round-state-2")).toHaveText("未进行");

  const styles = await cards.evaluateAll((nodes) =>
    nodes.map((node) => {
      const style = window.getComputedStyle(node);
      return {
        backgroundColor: style.backgroundColor,
        borderTopColor: style.borderTopColor,
      };
    }),
  );
  expect(styles).toHaveLength(3);
  expect(new Set(styles.map((style) => style.backgroundColor)).size).toBe(3);
  expect(new Set(styles.map((style) => style.borderTopColor)).size).toBe(3);
}

async function listMessageIdsForEmail(email: string): Promise<string[]> {
  const response = await fetch(`${MAILPIT_BASE_URL}/api/v1/messages?limit=100`);
  if (!response.ok) return [];
  const list = (await response.json()) as { messages?: MailpitMessage[] };
  return (list.messages ?? [])
    .filter((message) =>
      (message.To ?? []).some((recipient) => recipient.Address === email),
    )
    .map((message) => message.ID)
    .filter((id): id is string => Boolean(id));
}

async function pollMailCode(
  email: string,
  seenMessageIds: Set<string>,
): Promise<string> {
  for (let attempt = 0; attempt < 80; attempt += 1) {
    const response = await fetch(
      `${MAILPIT_BASE_URL}/api/v1/messages?limit=100`,
    );
    if (!response.ok) {
      throw new Error(`Mailpit list failed with status ${response.status}`);
    }
    const list = (await response.json()) as { messages?: MailpitMessage[] };
    const message = (list.messages ?? []).find(
      (candidate) =>
        candidate.ID &&
        !seenMessageIds.has(candidate.ID) &&
        (candidate.To ?? []).some((recipient) => recipient.Address === email),
    );
    if (message?.ID) {
      const detailResponse = await fetch(
        `${MAILPIT_BASE_URL}/api/v1/message/${message.ID}`,
      );
      if (!detailResponse.ok) {
        throw new Error(
          `Mailpit message read failed with status ${detailResponse.status}`,
        );
      }
      const detail = (await detailResponse.json()) as {
        Text?: string;
        HTML?: string;
      };
      const match = `${detail.Text ?? ""}\n${detail.HTML ?? ""}`.match(
        /\b\d{6}\b/,
      );
      if (!match) throw new Error("Mailpit message did not contain a code");
      seenMessageIds.add(message.ID);
      return match[0];
    }
    await new Promise((resolve) => setTimeout(resolve, 250));
  }
  throw new Error(`Timed out waiting for Mailpit message to ${email}`);
}

function corsHeaders(): Record<string, string> {
  return {
    "Access-Control-Allow-Credentials": "true",
    "Access-Control-Allow-Origin": FRONTEND_ORIGIN,
  };
}

function escapeRegExp(value: string): string {
  return value.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
}
