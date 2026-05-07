// @vitest-environment jsdom
/**
 * E2E.P0.004 — App shell language switch scenario.
 *
 * Truth source: docs/spec/frontend-shell/plans/001-app-shell-auth-settings/bdd-plan.md
 *               + bdd-checklist.md.
 *
 * Given a default App shell with runtime config and generated client bootstrap,
 * switching the TopBar language to English must update D1 shell copy while
 * preserving route/test IDs and sending Accept-Language as a display hint.
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
    const languageSelect = screen.getByTestId("topbar-lang-select");
    expect(languageSelect.tagName).toBe("SELECT");
    await user.selectOptions(languageSelect, "en");

    expect(screen.getByTestId("topbar-nav-home")).toHaveTextContent("Home");
    expect(screen.getByTestId("topbar-nav-jd_match")).toHaveTextContent(
      "Job Picks",
    );
    expect(screen.getByTestId("topbar-login")).toHaveTextContent("Sign in");
    expect(screen.getByTestId("topbar-register")).toHaveTextContent("Register");

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
    render(<App client={buildClient(seen, "authenticated")} />);
    await waitFor(() =>
      expect(screen.getByTestId("topbar-user-area")).toHaveAttribute(
        "data-signed-in",
        "true",
      ),
    );
    await user.selectOptions(screen.getByTestId("topbar-lang-select"), "en");
    expect(screen.getByTestId("topbar-user-profile")).toHaveTextContent(
      "User profile",
    );
    expect(screen.getByTestId("topbar-user-settings")).toHaveTextContent(
      "Settings & privacy",
    );
    expect(screen.getByTestId("topbar-user-logout")).toHaveTextContent(
      "Sign out",
    );
    await user.click(screen.getByTestId("topbar-user-profile"));
    expect(screen.getByTestId("route-profile")).toHaveTextContent(
      "User profile",
    );
    await user.click(screen.getByTestId("topbar-user-settings"));
    expect(screen.getByTestId("route-settings")).toHaveTextContent(
      "Settings & privacy",
    );

    expect(screen.queryByTestId("route-welcome")).not.toBeInTheDocument();
    for (const legacy of ["mistakes", "growth", "voice", "drill", "welcome"]) {
      expect(
        screen.queryByTestId(`topbar-nav-${legacy}`),
      ).not.toBeInTheDocument();
    }
    console.log("E2E.P0.004 evidence: language dropdown select Home Job Picks Sign in Register Accept-Language: en");
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
