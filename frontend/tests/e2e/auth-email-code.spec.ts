import { expect, test, type Browser, type Page } from "@playwright/test";

interface MailCode {
  subject: string;
  code: string;
}

interface FlowResult {
  kind: "register" | "login" | "duplicate-register";
  email: string;
  mailSubject: string;
  finalUrl: string;
  meStatus: number;
}

const FRONTEND_ORIGIN =
  process.env.EI_AUTH_EMAIL_CODE_FRONTEND_ORIGIN ?? "http://127.0.0.1:5173";
const API_BASE_URL =
  process.env.EI_AUTH_EMAIL_CODE_API_BASE_URL ?? "http://127.0.0.1:8080/api/v1";
const MAILPIT_BASE_URL =
  process.env.EI_AUTH_EMAIL_CODE_MAILPIT_BASE_URL ?? "http://127.0.0.1:8025";
const AUTH_EMAIL =
  process.env.EI_AUTH_EMAIL_CODE_EMAIL ??
  `auth-email-code-${Date.now()}@example.test`;
const DISPLAY_NAME = "Runtime Verify";
const DUPLICATE_DISPLAY_NAME = "Runtime Duplicate";

test.setTimeout(90_000);

test("E2E.P0.101 auth email-code same-email register/login lifecycle", async ({
  browser,
}) => {
  const { results, consoleErrors, pageErrors, unexpectedHttpFailures } =
    await runLifecycle(browser);

  for (const result of results) {
    console.log(
      [
        `E2E.P0.101 ${result.kind} email-code flow PASS`,
        `email=${result.email}`,
        `mailSubject=${result.mailSubject}`,
        "mailCode=<redacted>",
        `finalUrl=${result.finalUrl}`,
        `meStatus=${result.meStatus}`,
      ].join(" "),
    );
  }
  console.log(
    [
      "E2E.P0.101 auth email-code same-email lifecycle passed",
      `email=${AUTH_EMAIL}`,
      `consoleErrors=${consoleErrors}`,
      `pageErrors=${pageErrors}`,
      `httpFailures=${unexpectedHttpFailures}`,
    ].join(" "),
  );
});

async function runLifecycle(browser: Browser): Promise<{
  results: FlowResult[];
  consoleErrors: number;
  pageErrors: number;
  unexpectedHttpFailures: number;
}> {
  const context = await browser.newContext({ baseURL: FRONTEND_ORIGIN });
  const page = await context.newPage();
  const seenMessageIds = new Set(await listMessageIdsForEmail(AUTH_EMAIL));
  const consoleErrors: string[] = [];
  const pageErrors: string[] = [];
  const httpFailures: Array<{ status: number; url: string }> = [];

  page.on("console", (msg) => {
    const text = msg.text();
    if (
      (msg.type() === "error" || msg.type() === "warning") &&
      !isExpectedAuthNetworkConsoleWarning(text)
    ) {
      consoleErrors.push(redact(text));
    }
  });
  page.on("pageerror", (error) => pageErrors.push(redact(error.message)));
  page.on("response", (response) => {
    if (response.status() >= 400) {
      httpFailures.push({
        status: response.status(),
        url: redact(response.url()),
      });
    }
  });

  try {
    await page.goto(`${FRONTEND_ORIGIN}/auth/register`, {
      waitUntil: "domcontentloaded",
    });
    await startRegister(page, AUTH_EMAIL, DISPLAY_NAME);
    const registerMail = await pollMailCode(AUTH_EMAIL, seenMessageIds);
    await submitCode(page, registerMail.code);
    const registerMeStatus = await assertSignedIn(page, DISPLAY_NAME);
    const registerFinalUrl = page.url();

    await logout(page);

    await page.goto(`${FRONTEND_ORIGIN}/auth/login`, {
      waitUntil: "domcontentloaded",
    });
    await startLogin(page, AUTH_EMAIL);
    const loginMail = await pollMailCode(AUTH_EMAIL, seenMessageIds);
    await submitCode(page, loginMail.code);
    const loginMeStatus = await assertSignedIn(page, DISPLAY_NAME);
    const loginFinalUrl = page.url();

    await logout(page);

    await page.goto(`${FRONTEND_ORIGIN}/auth/register`, {
      waitUntil: "domcontentloaded",
    });
    await fillRegister(page, AUTH_EMAIL, DUPLICATE_DISPLAY_NAME);
    await page.getByTestId("auth-register-submit").click();
    await expect(page.getByTestId("route-auth_register")).toBeAttached();
    await expect(page.getByTestId("auth-register-status")).toBeVisible();
    await expect(page.getByTestId("topbar-user-area")).toHaveAttribute(
      "data-signed-in",
      "false",
    );
    const duplicateMeStatus = await currentUserStatus(page);
    const duplicateFinalUrl = page.url();
    expect(duplicateMeStatus).not.toBe(200);
    expect(new Set(await listMessageIdsForEmail(AUTH_EMAIL))).toEqual(
      seenMessageIds,
    );

    const unexpectedHttpFailures = httpFailures.filter(
      (failure) => !isExpectedAuthLifecycleHttpFailure(failure),
    );
    expect(consoleErrors, "console errors/warnings").toEqual([]);
    expect(pageErrors, "page errors").toEqual([]);
    expect(unexpectedHttpFailures, "unexpected HTTP >=400 failures").toEqual(
      [],
    );

    return {
      results: [
        {
          kind: "register",
          email: AUTH_EMAIL,
          mailSubject: registerMail.subject,
          finalUrl: registerFinalUrl,
          meStatus: registerMeStatus,
        },
        {
          kind: "login",
          email: AUTH_EMAIL,
          mailSubject: loginMail.subject,
          finalUrl: loginFinalUrl,
          meStatus: loginMeStatus,
        },
        {
          kind: "duplicate-register",
          email: AUTH_EMAIL,
          mailSubject: "not-sent",
          finalUrl: duplicateFinalUrl,
          meStatus: duplicateMeStatus,
        },
      ],
      consoleErrors: consoleErrors.length,
      pageErrors: pageErrors.length,
      unexpectedHttpFailures: unexpectedHttpFailures.length,
    };
  } finally {
    await context.close();
  }
}

async function startRegister(
  page: Page,
  email: string,
  displayName: string,
): Promise<void> {
  await fillRegister(page, email, displayName);
  await page.getByTestId("auth-register-submit").click();
  await page.getByTestId("route-auth_verify").waitFor({
    state: "attached",
    timeout: 10_000,
  });
  await expect(page.getByTestId("auth-verify-email-hint")).toContainText(email);
}

async function fillRegister(
  page: Page,
  email: string,
  displayName: string,
): Promise<void> {
  await page.getByTestId("auth-register-name").fill(displayName);
  await page.getByTestId("auth-register-email").fill(email);
  const terms = page.getByTestId("auth-register-terms");
  if (!(await terms.isChecked())) await terms.check();
}

async function startLogin(page: Page, email: string): Promise<void> {
  await page.getByTestId("auth-login-email").fill(email);
  await page.getByTestId("auth-login-submit-email").click();
  await page.getByTestId("route-auth_verify").waitFor({
    state: "attached",
    timeout: 10_000,
  });
  await expect(page.getByTestId("auth-verify-email-hint")).toContainText(email);
}

async function submitCode(page: Page, code: string): Promise<void> {
  await expect(code).toMatch(/^\d{6}$/);
  await page.getByTestId("auth-verify-code").fill(code);
  await page.getByTestId("auth-verify-submit").click();
}

async function assertSignedIn(
  page: Page,
  expectedDisplayName: string,
): Promise<number> {
  await page.waitForFunction(
    () => window.location.pathname === "/",
    null,
    { timeout: 10_000 },
  );
  const meStatus = await currentUserStatus(page);
  expect(meStatus).toBe(200);
  await expect(page.getByTestId("topbar-user-name")).toContainText(
    expectedDisplayName,
  );
  await expect(page.getByTestId("topbar-user-name")).not.toContainText(
    DUPLICATE_DISPLAY_NAME,
  );
  return meStatus;
}

async function logout(page: Page): Promise<void> {
  await page.getByTestId("topbar-user-chip").click();
  await page.getByTestId("topbar-user-logout").click();
  await page.getByTestId("auth-logout-confirm").click();
  await expect(page.getByTestId("topbar-user-area")).toHaveAttribute(
    "data-signed-in",
    "false",
  );
}

async function listMessageIdsForEmail(email: string): Promise<string[]> {
  const listResponse = await fetch(
    `${MAILPIT_BASE_URL}/api/v1/messages?limit=100`,
  );
  if (!listResponse.ok) return [];
  const list = (await listResponse.json()) as {
    messages?: Array<{
      ID?: string;
      To?: Array<{ Address?: string }>;
    }>;
  };
  return (list.messages ?? [])
    .filter((item) => (item.To ?? []).some((to) => to.Address === email))
    .map((item) => item.ID)
    .filter((id): id is string => Boolean(id));
}

async function pollMailCode(
  email: string,
  seenMessageIds: Set<string>,
): Promise<MailCode> {
  for (let attempt = 0; attempt < 80; attempt += 1) {
    const listResponse = await fetch(
      `${MAILPIT_BASE_URL}/api/v1/messages?limit=100`,
    );
    if (!listResponse.ok) {
      throw new Error(`Mailpit list failed with status ${listResponse.status}`);
    }
    const list = (await listResponse.json()) as {
      messages?: Array<{
        ID?: string;
        To?: Array<{ Address?: string }>;
      }>;
    };
    const message = (list.messages ?? []).find(
      (item) =>
        item.ID &&
        !seenMessageIds.has(item.ID) &&
        (item.To ?? []).some((to) => to.Address === email),
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
        Subject?: string;
        Text?: string;
        HTML?: string;
      };
      const body = `${detail.Text ?? ""}\n${detail.HTML ?? ""}`;
      expect(body).not.toContain("/auth/verify?token=");
      expect(body).not.toContain("/api/v1/auth/email/verify");
      const code = findSixDigitCode(body);
      if (!code) {
        throw new Error(`Mailpit message for ${email} did not contain a code`);
      }
      seenMessageIds.add(message.ID);
      return {
        subject: detail.Subject ?? "",
        code,
      };
    }
    await new Promise((resolve) => setTimeout(resolve, 250));
  }
  throw new Error(`Timed out waiting for Mailpit message to ${email}`);
}

function findSixDigitCode(text: string): string | null {
  const match = text.match(/\b\d{6}\b/);
  return match?.[0] ?? null;
}

function isExpectedAuthNetworkConsoleWarning(text: string): boolean {
  return (
    text.includes("Failed to load resource") &&
    (text.includes("status of 401") || text.includes("status of 409"))
  );
}

function isExpectedAuthLifecycleHttpFailure(failure: {
  status: number;
  url: string;
}): boolean {
  if (failure.status === 401 && failure.url.endsWith("/api/v1/me")) return true;
  if (
    failure.status === 401 &&
    failure.url.includes("/api/v1/targets?pageSize=")
  ) {
    return true;
  }
  return (
    failure.status === 409 &&
    failure.url.endsWith("/api/v1/auth/email/start")
  );
}

async function currentUserStatus(page: Page): Promise<number> {
  return page.evaluate(async (apiBaseUrl) => {
    const response = await fetch(`${apiBaseUrl}/me`, {
      credentials: "include",
    });
    return response.status;
  }, API_BASE_URL);
}

function redact(value: string): string {
  return value
    .replace(/token=[^&\s]+/g, "token=<redacted>")
    .replace(/\b\d{6}\b/g, "<redacted-code>")
    .replace(/ei_session=[^;\s]+/g, "ei_session=<redacted>");
}
