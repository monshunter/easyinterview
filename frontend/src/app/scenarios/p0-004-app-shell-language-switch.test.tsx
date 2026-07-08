// @vitest-environment jsdom
/**
 * E2E.P0.004 — App shell language switch scenario.
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
  auth: "authenticated" | "unauthenticated",
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
        if (auth === "authenticated") {
          return jsonResponse({
            id: "01918fa0-0000-7000-8000-000000000100",
            emailMasked: "ali***@example.com",
            displayName: "Alice Example",
            uiLanguage: "zh-CN",
            preferredPracticeLanguage: "zh-CN",
            profileCompletionRequired: false,
          });
        }
        return jsonResponse(
          {
            error: {
              code: "AUTH_UNAUTHORIZED",
              message: "not signed in",
              requestId: "req-p0-004",
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

describe("E2E.P0.004 app shell language switch", () => {
  it("switches D1 shell copy to English and sends Accept-Language", async () => {
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
    await user.click(screen.getByTestId("topbar-user-chip"));
    expect(screen.queryByTestId("topbar-user-profile")).not.toBeInTheDocument();
    expect(screen.getByTestId("topbar-user-settings")).toHaveTextContent(
      "Settings & privacy",
    );
    expect(screen.getByTestId("topbar-user-logout")).toHaveTextContent(
      "Sign out",
    );
    await user.click(screen.getByTestId("topbar-user-settings"));
    expect(screen.getByTestId("route-settings")).toHaveTextContent(
      "Settings & privacy",
    );

    expect(screen.queryByTestId("route-welcome")).not.toBeInTheDocument();
    for (const nonCurrent of ["mistakes", "growth", "voice", "drill", "welcome"]) {
      expect(
        screen.queryByTestId(`topbar-nav-${nonCurrent}`),
      ).not.toBeInTheDocument();
    }
    console.log("E2E.P0.004 evidence: language dropdown Home Interview Resume Sign in Accept-Language: en");
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
