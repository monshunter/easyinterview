import { expect, test, type Browser, type Page } from "@playwright/test";

interface MailCode {
  subject: string;
  code: string;
}

interface FlowResult {
  kind:
    | "first-login-profile-setup"
    | "cross-browser-relogin-profile-setup"
    | "logout-relogin-profile-setup"
    | "existing-email-login";
  email: string;
  mailSubject: string;
  finalUrl: string;
  meStatus: number;
  profileCompletionRequired: boolean;
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
const AUTH_RATE_LIMIT_WINDOW_MS = 65_000;

test.setTimeout(180_000);

test("E2E.P0.101 auth email-code same-email login/profile lifecycle", async ({
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
        `profileCompletionRequired=${String(result.profileCompletionRequired)}`,
      ].join(" "),
    );
  }
  console.log(
    [
      "E2E.P0.101 profile-required gates PASS",
      "refresh=profile-setup",
      "deepLink=profile-setup",
      "crossBrowser=profile-setup",
      "logoutRelogin=profile-setup",
      "authStartBodyKeys=email",
      "authRegisterLivePage=absent",
      "topbarRegister=absent",
    ].join(" "),
  );
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
  const seenMessageIds = new Set(await listMessageIdsForEmail(AUTH_EMAIL));
  const consoleErrors: string[] = [];
  const pageErrors: string[] = [];
  const httpFailures: Array<{ status: number; url: string }> = [];
  const authStartBodies: Array<Record<string, unknown>> = [];
  const results: FlowResult[] = [];

  const firstContext = await browser.newContext({ baseURL: FRONTEND_ORIGIN });
  const firstPage = await firstContext.newPage();
  attachDiagnostics(firstPage, {
    consoleErrors,
    pageErrors,
    httpFailures,
    authStartBodies,
  });

  try {
    await assertNoRegisterEntry(firstPage);
    await firstPage.goto(`${FRONTEND_ORIGIN}/auth/login`, {
      waitUntil: "domcontentloaded",
    });
    await startLogin(firstPage, AUTH_EMAIL);
    const firstLoginMail = await pollMailCode(AUTH_EMAIL, seenMessageIds);
    await submitCode(firstPage, firstLoginMail.code);
    await expectProfileSetup(firstPage);
    const firstLoginUser = await currentUserContext(firstPage);
    expect(firstLoginUser.status).toBe(200);
    expect(firstLoginUser.profileCompletionRequired).toBe(true);

    await firstPage.reload({ waitUntil: "domcontentloaded" });
    await expectProfileSetup(firstPage);
    await firstPage.goto(`${FRONTEND_ORIGIN}/practice?planId=p0-101-deep-link`, {
      waitUntil: "domcontentloaded",
    });
    await expectProfileSetup(firstPage);
    await expect(firstPage).toHaveURL(/\/auth\/profile/);
    await expect(firstPage).toHaveURL(/pendingRoute=practice/);
    results.push({
      kind: "first-login-profile-setup",
      email: AUTH_EMAIL,
      mailSubject: firstLoginMail.subject,
      finalUrl: firstPage.url(),
      meStatus: firstLoginUser.status,
      profileCompletionRequired: true,
    });
  } finally {
    await firstContext.close();
  }

  const secondContext = await browser.newContext({ baseURL: FRONTEND_ORIGIN });
  const secondPage = await secondContext.newPage();
  attachDiagnostics(secondPage, {
    consoleErrors,
    pageErrors,
    httpFailures,
    authStartBodies,
  });

  try {
    await secondPage.goto(`${FRONTEND_ORIGIN}/auth/login`, {
      waitUntil: "domcontentloaded",
    });
    await startLogin(secondPage, AUTH_EMAIL);
    const crossBrowserMail = await pollMailCode(AUTH_EMAIL, seenMessageIds);
    await submitCode(secondPage, crossBrowserMail.code);
    await expectProfileSetup(secondPage);
    const crossBrowserUser = await currentUserContext(secondPage);
    expect(crossBrowserUser.status).toBe(200);
    expect(crossBrowserUser.profileCompletionRequired).toBe(true);
    results.push({
      kind: "cross-browser-relogin-profile-setup",
      email: AUTH_EMAIL,
      mailSubject: crossBrowserMail.subject,
      finalUrl: secondPage.url(),
      meStatus: crossBrowserUser.status,
      profileCompletionRequired: true,
    });

    await secondPage.goto(`${FRONTEND_ORIGIN}/auth/logout`, {
      waitUntil: "domcontentloaded",
    });
    await secondPage.getByTestId("auth-logout-confirm").click();
    await expect(secondPage.getByTestId("topbar-user-area")).toHaveAttribute(
      "data-signed-in",
      "false",
    );

    await secondPage.goto(`${FRONTEND_ORIGIN}/auth/login`, {
      waitUntil: "domcontentloaded",
    });
    await waitForAuthRateLimitWindow();
    await startLogin(secondPage, AUTH_EMAIL);
    const logoutReloginMail = await pollMailCode(AUTH_EMAIL, seenMessageIds);
    await submitCode(secondPage, logoutReloginMail.code);
    await expectProfileSetup(secondPage);
    const logoutReloginUser = await currentUserContext(secondPage);
    expect(logoutReloginUser.status).toBe(200);
    expect(logoutReloginUser.profileCompletionRequired).toBe(true);
    results.push({
      kind: "logout-relogin-profile-setup",
      email: AUTH_EMAIL,
      mailSubject: logoutReloginMail.subject,
      finalUrl: secondPage.url(),
      meStatus: logoutReloginUser.status,
      profileCompletionRequired: true,
    });

    await completeProfile(secondPage, DISPLAY_NAME);
    const completedUser = await assertSignedIn(secondPage, DISPLAY_NAME);
    expect(completedUser.profileCompletionRequired).toBe(false);

    await logout(secondPage);

    await secondPage.goto(`${FRONTEND_ORIGIN}/auth/login`, {
      waitUntil: "domcontentloaded",
    });
    await startLogin(secondPage, AUTH_EMAIL);
    const completedLoginMail = await pollMailCode(AUTH_EMAIL, seenMessageIds);
    await submitCode(secondPage, completedLoginMail.code);
    const completedLoginUser = await assertSignedIn(secondPage, DISPLAY_NAME);
    await expect(secondPage.getByTestId("route-auth_profile_setup")).toHaveCount(0);

    results.push({
      kind: "existing-email-login",
      email: AUTH_EMAIL,
      mailSubject: completedLoginMail.subject,
      finalUrl: secondPage.url(),
      meStatus: completedLoginUser.status,
      profileCompletionRequired:
        completedLoginUser.profileCompletionRequired ?? false,
    });

    expect(authStartBodies).toHaveLength(4);
    for (const body of authStartBodies) {
      expect(Object.keys(body).sort()).toEqual(["email"]);
      expect(body).not.toHaveProperty("purpose");
      expect(body).not.toHaveProperty("displayName");
    }

    const unexpectedHttpFailures = httpFailures.filter(
      (failure) => !isExpectedAuthLifecycleHttpFailure(failure),
    );
    expect(consoleErrors, "console errors/warnings").toEqual([]);
    expect(pageErrors, "page errors").toEqual([]);
    expect(unexpectedHttpFailures, "unexpected HTTP >=400 failures").toEqual(
      [],
    );

    return {
      results,
      consoleErrors: consoleErrors.length,
      pageErrors: pageErrors.length,
      unexpectedHttpFailures: unexpectedHttpFailures.length,
    };
  } finally {
    await secondContext.close();
  }
}

async function waitForAuthRateLimitWindow(): Promise<void> {
  // The real auth API silently dedupes the third same-email/IP challenge within
  // one minute. Wait instead of mutating DB state so this remains a real-stack
  // browser flow.
  await new Promise((resolve) =>
    setTimeout(resolve, AUTH_RATE_LIMIT_WINDOW_MS),
  );
}

function attachDiagnostics(
  page: Page,
  sinks: {
    consoleErrors: string[];
    pageErrors: string[];
    httpFailures: Array<{ status: number; url: string }>;
    authStartBodies: Array<Record<string, unknown>>;
  },
): void {
  page.on("console", (msg) => {
    const text = msg.text();
    if (
      (msg.type() === "error" || msg.type() === "warning") &&
      !isExpectedAuthNetworkConsoleWarning(text)
    ) {
      sinks.consoleErrors.push(redact(text));
    }
  });
  page.on("pageerror", (error) => sinks.pageErrors.push(redact(error.message)));
  page.on("response", (response) => {
    if (response.status() >= 400) {
      sinks.httpFailures.push({
        status: response.status(),
        url: redact(response.url()),
      });
    }
  });
  page.on("request", (request) => {
    if (
      request.method() === "POST" &&
      request.url().endsWith("/api/v1/auth/email/start")
    ) {
      const payload = request.postDataJSON() as Record<string, unknown>;
      sinks.authStartBodies.push(payload);
    }
  });
}

async function assertNoRegisterEntry(page: Page): Promise<void> {
  await page.goto(`${FRONTEND_ORIGIN}/auth/register`, {
    waitUntil: "domcontentloaded",
  });
  await expect(page.getByTestId("route-auth_register")).toHaveCount(0);
  await expect(page.getByTestId("topbar-register")).toHaveCount(0);
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

async function expectProfileSetup(page: Page): Promise<void> {
  await page.getByTestId("route-auth_profile_setup").waitFor({
    state: "attached",
    timeout: 10_000,
  });
  await expect(page.getByTestId("auth-profile-name")).toBeVisible();
}

async function completeProfile(
  page: Page,
  displayName: string,
): Promise<void> {
  await page.getByTestId("route-auth_profile_setup").waitFor({
    state: "attached",
    timeout: 10_000,
  });
  await page.getByTestId("auth-profile-name").fill(displayName);
  const terms = page.getByTestId("auth-profile-terms");
  if (!(await terms.isChecked())) await terms.check();
  await page.getByTestId("auth-profile-submit").click();
}

async function assertSignedIn(
  page: Page,
  expectedDisplayName: string,
): Promise<{
  status: number;
  profileCompletionRequired?: boolean;
  displayName?: string;
}> {
  await page.waitForFunction(
    () => window.location.pathname === "/",
    null,
    { timeout: 10_000 },
  );
  const currentUser = await currentUserContext(page);
  expect(currentUser.status).toBe(200);
  expect(currentUser.displayName).toBe(expectedDisplayName);
  await expect(page.getByTestId("topbar-user-name")).toContainText(
    expectedDisplayName,
  );
  await expect(page.getByTestId("topbar-user-name")).not.toContainText(
    "Runtime Duplicate",
  );
  return currentUser;
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
    text.includes("status of 401")
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
  return false;
}

async function currentUserContext(page: Page): Promise<{
  status: number;
  profileCompletionRequired?: boolean;
  displayName?: string;
}> {
  return page.evaluate(async (apiBaseUrl) => {
    const response = await fetch(`${apiBaseUrl}/me`, {
      credentials: "include",
    });
    const body = response.ok ? await response.json() : {};
    return {
      status: response.status,
      profileCompletionRequired: body.profileCompletionRequired,
      displayName: body.displayName,
    };
  }, API_BASE_URL);
}

function redact(value: string): string {
  return value
    .replace(/token=[^&\s]+/g, "token=<redacted>")
    .replace(/\b\d{6}\b/g, "<redacted-code>")
    .replace(/ei_session=[^;\s]+/g, "ei_session=<redacted>");
}
