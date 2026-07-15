// @vitest-environment jsdom
import { afterEach, describe, expect, it } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

import { EasyInterviewClient } from "../../api/generated/client";
import { App } from "../App";
import { normalizeLocaleTag } from "./messages";

function jsonResponse(body: unknown, init?: ResponseInit): Response {
  return new Response(JSON.stringify(body), {
    status: 200,
    headers: { "Content-Type": "application/json" },
    ...init,
  });
}

describe("locale bootstrap", () => {
  afterEach(() => {
    setNavigatorLanguages("en-US", ["en-US"]);
    window.localStorage.clear();
  });

  it("normalizes BCP 47 UI locale tags to supported shell languages", () => {
    expect(normalizeLocaleTag("zh-CN")).toBe("zh");
    expect(normalizeLocaleTag("en-US")).toBe("en");
    expect(normalizeLocaleTag("fr-FR")).toBe("en");
    expect(normalizeLocaleTag(undefined)).toBe("en");
  });

  it("uses the browser locale as the initial shell language", async () => {
    setNavigatorLanguages("zh-CN", ["zh-CN", "en-US"]);
    const client = new EasyInterviewClient({
      fetch: async (input) => {
        const url = String(input);
        if (url.endsWith("/runtime-config")) {
          return jsonResponse({
            defaultUiLanguage: "en-US",
            featureFlags: {},
            appVersion: "test",
            analyticsEnabled: false,
          });
        }
        if (url.endsWith("/me")) {
          return jsonResponse(
            {
              error: {
                code: "AUTH_UNAUTHORIZED",
                message: "not signed in",
                requestId: "req-locale-test",
                retryable: false,
              },
            },
            { status: 401, statusText: "Unauthorized" },
          );
        }
        return jsonResponse({});
      },
    });

    render(<App client={client} />);

    await waitFor(() =>
      expect(screen.getByTestId("topbar-nav-home")).toHaveTextContent("首页"),
    );
    expect(screen.getByTestId("topbar-login")).toHaveTextContent("登录");
  });

  it("does not let runtime config or /me locale override frontend preferences", async () => {
    setNavigatorLanguages("en-US", ["en-US"]);
    const seen: Array<{ url: string; acceptLanguage: string | null }> = [];
    const client = new EasyInterviewClient({
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
          return jsonResponse({
            id: "01918fa0-0000-7000-8000-000000000100",
            email: "alice@example.com",
            displayName: "Alice Example",
            profileCompletionRequired: false,
          });
        }
        return jsonResponse({});
      },
    });

    render(<App client={client} />);

    await waitFor(() =>
      expect(screen.getByTestId("topbar-user-area")).toHaveAttribute(
        "data-signed-in",
        "true",
      ),
    );
    expect(screen.getByTestId("topbar-nav-home")).toHaveTextContent("Home");
    expect(screen.getByTestId("topbar-settings")).toHaveAccessibleName(
      "Settings & privacy",
    );
    expect(screen.queryByTestId("topbar-user-menu")).not.toBeInTheDocument();
    expect(
      seen.filter((request) => request.url.endsWith("/me")),
    ).toHaveLength(1);
  });

  it("sends the current UI locale as Accept-Language on D1 auth operations", async () => {
    setNavigatorLanguages("zh-CN", ["zh-CN"]);
    const seen: Array<{ url: string; acceptLanguage: string | null }> = [];
    const client = new EasyInterviewClient({
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
          return jsonResponse(
            {
              error: {
                code: "AUTH_UNAUTHORIZED",
                message: "not signed in",
                requestId: "req-locale-test",
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

    render(<App client={client} />);
    await waitFor(() =>
      expect(screen.getByTestId("topbar-login")).toHaveTextContent("登录"),
    );

    const user = userEvent.setup();
    await user.click(screen.getByTestId("topbar-lang-toggle"));
    await user.click(screen.getByTestId("topbar-lang-option-en"));
    await user.click(screen.getByTestId("topbar-login"));
    await user.type(screen.getByTestId("auth-login-email"), "liuzhe@example.com");
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
    expect(seen.some((request) => request.acceptLanguage === "zh")).toBe(true);
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
