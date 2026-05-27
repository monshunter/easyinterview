import { expect, test, type Browser, type Page } from "@playwright/test";

type FlowKind = "login" | "register";

interface MailLink {
  subject: string;
  link: string;
}

interface FlowResult {
  kind: FlowKind;
  email: string;
  mailSubject: string;
  finalUrl: string;
  meStatus: number;
  consoleErrors: number;
  pageErrors: number;
  httpFailures: number;
}

const FRONTEND_ORIGIN =
  process.env.EI_AUTH_MAIL_LINK_FRONTEND_ORIGIN ?? "http://127.0.0.1:5173";
const API_BASE_URL =
  process.env.EI_AUTH_MAIL_LINK_API_BASE_URL ?? "http://127.0.0.1:8080/api/v1";
const MAILPIT_BASE_URL =
  process.env.EI_AUTH_MAIL_LINK_MAILPIT_BASE_URL ?? "http://127.0.0.1:8025";
const LOGIN_EMAIL =
  process.env.EI_AUTH_MAIL_LINK_LOGIN_EMAIL ??
  `auth-mail-link-login-${Date.now()}@example.test`;
const REGISTER_EMAIL =
  process.env.EI_AUTH_MAIL_LINK_REGISTER_EMAIL ??
  `auth-mail-link-register-${Date.now()}@example.test`;

test.setTimeout(60_000);

test("E2E.P0.101 auth mail-link login/register callback", async ({
  browser,
}) => {
  const results = [
    await runFlow(browser, "login", LOGIN_EMAIL),
    await runFlow(browser, "register", REGISTER_EMAIL),
  ];

  for (const result of results) {
    console.log(
      [
        `E2E.P0.101 ${result.kind} mail-link flow PASS`,
        `email=${result.email}`,
        `mailSubject=${result.mailSubject}`,
        `mailLink=${FRONTEND_ORIGIN}/auth/verify?token=<redacted>`,
        `finalUrl=${result.finalUrl}`,
        `meStatus=${result.meStatus}`,
        `consoleErrors=${result.consoleErrors}`,
        `pageErrors=${result.pageErrors}`,
        `httpFailures=${result.httpFailures}`,
      ].join(" "),
    );
  }
  console.log("E2E.P0.101 auth mail-link login/register 2 flows passed");
});

async function runFlow(
  browser: Browser,
  kind: FlowKind,
  email: string,
): Promise<FlowResult> {
  const context = await browser.newContext({ baseURL: FRONTEND_ORIGIN });
  const page = await context.newPage();
  const consoleErrors: string[] = [];
  const pageErrors: string[] = [];
  const httpFailures: Array<{ status: number; url: string }> = [];

  page.on("console", (msg) => {
    if (msg.type() === "error" || msg.type() === "warning") {
      consoleErrors.push(redact(msg.text()));
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
    await page.goto(`${FRONTEND_ORIGIN}/auth/${kind}`, {
      waitUntil: "domcontentloaded",
    });

    if (kind === "register") {
      await page.getByTestId("auth-register-name").fill("Runtime Verify");
      await page.getByTestId("auth-register-email").fill(email);
      const terms = page.getByTestId("auth-register-terms");
      if (!(await terms.isChecked())) await terms.check();
      await page.getByTestId("auth-register-submit").click();
    } else {
      await page.getByTestId("auth-login-email").fill(email);
      await page.getByTestId("auth-login-submit-email").click();
    }

    await page.getByTestId("route-auth_verify").waitFor({
      state: "attached",
      timeout: 10_000,
    });
    await expect(page.getByTestId("auth-verify-email-hint")).toContainText(
      email,
    );

    const mail = await pollMailLink(email);
    const url = new URL(mail.link);
    expect(url.origin).toBe(FRONTEND_ORIGIN);
    expect(url.pathname).toBe("/auth/verify");
    expect(url.searchParams.get("token")).toBeTruthy();

    await page.goto(mail.link, { waitUntil: "domcontentloaded" });
    await page.waitForFunction(
      () => !window.location.href.includes("token="),
      null,
      { timeout: 10_000 },
    );
    await page.waitForFunction(
      () => window.location.pathname === "/",
      null,
      { timeout: 10_000 },
    );

    const meStatus = await currentUserStatus(page);
    expect(meStatus).toBe(200);
    expect(consoleErrors, "console errors/warnings").toEqual([]);
    expect(pageErrors, "page errors").toEqual([]);
    expect(httpFailures, "HTTP >=400 failures").toEqual([]);

    return {
      kind,
      email,
      mailSubject: mail.subject,
      finalUrl: page.url(),
      meStatus,
      consoleErrors: consoleErrors.length,
      pageErrors: pageErrors.length,
      httpFailures: httpFailures.length,
    };
  } finally {
    await context.close();
  }
}

async function pollMailLink(email: string): Promise<MailLink> {
  for (let attempt = 0; attempt < 40; attempt += 1) {
    const listResponse = await fetch(`${MAILPIT_BASE_URL}/api/v1/messages?limit=50`);
    if (!listResponse.ok) {
      throw new Error(`Mailpit list failed with status ${listResponse.status}`);
    }
    const list = (await listResponse.json()) as {
      messages?: Array<{
        ID?: string;
        To?: Array<{ Address?: string }>;
      }>;
    };
    const message = (list.messages ?? []).find((item) =>
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
      };
      const text = detail.Text ?? "";
      const link = findFrontendMailLink(text);
      if (!link) {
        throw new Error(
          `Mailpit message for ${email} did not contain a frontend auth callback`,
        );
      }
      return { subject: detail.Subject ?? "", link };
    }
    await new Promise((resolve) => setTimeout(resolve, 250));
  }
  throw new Error(`Timed out waiting for Mailpit message to ${email}`);
}

function findFrontendMailLink(text: string): string | null {
  const escaped = escapeRegExp(FRONTEND_ORIGIN);
  const match = text.match(
    new RegExp(`${escaped}/auth/verify\\?token=[^\\s]+`),
  );
  return match?.[0] ?? null;
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
    .replace(/ei_session=[^;\s]+/g, "ei_session=<redacted>");
}

function escapeRegExp(value: string): string {
  return value.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
}
