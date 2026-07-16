// @vitest-environment jsdom
/**
 * Code-level app shell language-switch regression.
 *
 * Truth source: docs/spec/frontend-shell/plans/001-app-shell-auth-settings/bdd-plan.md
 *               + bdd-checklist.md.
 *
 * Given a default App shell with runtime config and generated client bootstrap,
 * selecting English from the TopBar language dropdown must update D1 shell copy
 * while preserving route/test IDs and sending Accept-Language as a display hint.
 */
import { describe, expect, it } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

import { EasyInterviewClient } from "../../api/generated/client";
import { App } from "../App";

function jsonResponse(body: unknown, init?: ResponseInit): Response {
  return new Response(JSON.stringify(body), {
    status: 200,
    headers: { "Content-Type": "application/json" },
    ...init,
  });
}

function buildClient(
  seen: Array<{ url: string; acceptLanguage: string | null }>,
  auth: "authenticated" | "unauthenticated" | "loading" | "error",
): EasyInterviewClient {
  return new EasyInterviewClient({
    fetch: async (input, init) => {
      const url = String(input);
      const headers = new Headers(init?.headers);
      seen.push({ url, acceptLanguage: headers.get("Accept-Language") });
      if (url.endsWith("/runtime-config")) {
        return jsonResponse({
          defaultUiLanguage: "zh-CN",
          featureFlags: {},
          appVersion: "test",
          analyticsEnabled: false,
        });
      }
      if (url.endsWith("/me")) {
        if (auth === "loading") {
          return new Promise<Response>(() => undefined);
        }
        if (auth === "error") {
          return jsonResponse(
            {
              error: {
                code: "VALIDATION_FAILED",
                message: "auth probe unavailable",
                requestId: "req-language-gate-error",
                retryable: false,
              },
            },
            { status: 503, statusText: "Service Unavailable" },
          );
        }
        if (auth === "authenticated") {
          return jsonResponse({
            id: "01918fa0-0000-7000-8000-000000000100",
            email: "alice@example.com",
            displayName: "Alice Example",
            profileCompletionRequired: false,
          });
        }
        return jsonResponse(
          {
            error: {
              code: "AUTH_UNAUTHORIZED",
              message: "not signed in",
              requestId: "req-language-switch",
              retryable: false,
            },
          },
          { status: 401, statusText: "Unauthorized" },
        );
      }
      if (url.endsWith("/auth/email/start")) {
        return jsonResponse({}, { status: 202 });
      }
      return jsonResponse({});
    },
  });
}

describe("app shell language switch", () => {
  it.each([
    ["loading", "正在检查登录状态", "正在验证当前会话，请稍候。"],
    ["error", "需要登录", "请先登录，再打开面试工作区。"],
  ] as const)(
    "localizes the protected-route auth %s gate without mounting the business screen",
    async (auth, expectedTitle, expectedBody) => {
      localStorage.setItem("ei-lang", "zh");
      setNavigatorLanguages("zh-CN", ["zh-CN", "en-US"]);
      const seen: Array<{ url: string; acceptLanguage: string | null }> = [];
      const user = userEvent.setup();
      render(
        <App
          client={buildClient(seen, auth)}
          initialRoute={{ name: "resume_versions", params: {} }}
        />,
      );

      const gate = await screen.findByTestId("auth-route-gate");
      expect(gate).toHaveTextContent("登录状态");
      expect(gate).toHaveTextContent(expectedTitle);
      expect(gate).toHaveTextContent(expectedBody);
      expect(gate).not.toHaveTextContent(/AUTH|Checking sign-in|Sign-in required|Please sign in|Please wait/);
      expect(screen.queryByTestId("resume-workshop-screen")).not.toBeInTheDocument();
      expect(seen.some((request) => request.url.includes("/resumes"))).toBe(false);

      await user.click(screen.getByTestId("topbar-lang-toggle"));
      await user.click(screen.getByTestId("topbar-lang-option-en"));
      expect(gate).toHaveTextContent("Authentication");
      expect(gate).toHaveTextContent(
        auth === "loading" ? "Checking sign-in" : "Sign-in required",
      );
    },
  );

  it("switches D1 shell copy to English and sends Accept-Language", async () => {
    localStorage.setItem("ei-lang", "zh");
    setNavigatorLanguages("zh-CN", ["zh-CN", "en-US"]);
    const seen: Array<{ url: string; acceptLanguage: string | null }> = [];
    const user = userEvent.setup();
    const signedOut = render(
      <App client={buildClient(seen, "unauthenticated")} />,
    );

    await waitFor(() =>
      expect(screen.getByTestId("topbar-login")).toHaveTextContent("登录"),
    );
    const languageToggle = screen.getByTestId("topbar-lang-toggle");
    expect(languageToggle.tagName).toBe("BUTTON");
    expect(languageToggle).toHaveAttribute("aria-expanded", "false");
    expect(languageToggle).toHaveTextContent("中文");
    await user.click(languageToggle);
    expect(screen.getByTestId("topbar-lang-menu")).toBeInTheDocument();
    await user.click(screen.getByTestId("topbar-lang-option-en"));

    expect(screen.getByTestId("topbar-nav-home")).toHaveTextContent(/^Home$/);
    expect(screen.getByTestId("topbar-nav-workspace")).toHaveTextContent(
      /^Interview$/,
    );
    expect(screen.getByTestId("topbar-nav-resume_versions")).toHaveTextContent(
      /^Resume$/,
    );
    expect(screen.getByTestId("topbar-login")).toHaveTextContent("Sign in");
    expect(screen.queryByTestId("topbar-register")).not.toBeInTheDocument();

    await user.click(screen.getByTestId("topbar-login"));
    expect(screen.getByTestId("route-auth_login")).toHaveAttribute(
      "data-route-name",
      "auth_login",
    );
    expect(screen.getByTestId("route-auth_login")).toHaveTextContent("Sign in");
    await user.type(screen.getByTestId("auth-login-email"), "alice@example.com");
    await user.click(screen.getByTestId("auth-login-submit-email"));

    await waitFor(() =>
      expect(
        seen.some(
          (request) =>
            request.url.endsWith("/auth/email/start") &&
            request.acceptLanguage === "en",
        ),
      ).toBe(true),
    );

    signedOut.unmount();
    window.history.replaceState(null, "", "/");
    render(<App client={buildClient(seen, "authenticated")} />);
    await waitFor(() =>
      expect(screen.getByTestId("topbar-user-area")).toHaveAttribute(
        "data-signed-in",
        "true",
      ),
    );
    await user.click(screen.getByTestId("topbar-lang-toggle"));
    await user.click(screen.getByTestId("topbar-lang-option-en"));
    expect(screen.getByTestId("topbar-settings")).toHaveAccessibleName(
      "Settings & privacy",
    );
    expect(screen.queryByTestId("topbar-user-menu")).not.toBeInTheDocument();
    await user.click(screen.getByTestId("topbar-settings"));
    expect(screen.getByTestId("route-settings")).toHaveTextContent(
      "Settings & privacy",
    );

    expect(screen.queryByTestId("route-welcome")).not.toBeInTheDocument();
    for (const outOfScope of ["mistakes", "growth", "voice", "drill", "welcome"]) {
      expect(
        screen.queryByTestId(`topbar-nav-${outOfScope}`),
      ).not.toBeInTheDocument();
    }
    console.log("language switch regression: dropdown Home Interview Resume Sign in Accept-Language: en");
  });
});

function setNavigatorLanguages(language: string, languages = [language]) {
  Object.defineProperty(window.navigator, "language", {
    value: language,
    configurable: true,
  });
  Object.defineProperty(window.navigator, "languages", {
    value: languages,
    configurable: true,
  });
}
