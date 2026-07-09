// @vitest-environment jsdom
/**
 * E2E.P0.089 — Auth pendingAction + URL privacy redline.
 *
 * Truth source: docs/spec/frontend-shell/plans/004-url-addressable-routing/
 * bdd-plan.md §2 (E2E.P0.089) + bdd-checklist.md.
 *
 * Given an unauthenticated user opens URL-addressable workflow paths with
 * representative raw markers seeded across raw JD text, source URL,
 * resume text, guided answers, parsed summary, structured profile,
 * suggestion text, question/answer text, debrief notes and AI prompt /
 * response tokens, this scenario asserts:
 *   - Restored route and canonical URL contain only route name + safe IDs
 *     / hints; legal handoff params survive allowlist filtering.
 *   - Raw markers have ZERO hits in URL, history.state, pendingAction,
 *     localStorage, sessionStorage and console capture.
 *   - Auth token / secret never enters URL, pendingAction or storage.
 */
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { act, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import type { FC } from "react";

import getMeFixture from "../../../../openapi/fixtures/Auth/getMe.json";
import getRuntimeConfigFixture from "../../../../openapi/fixtures/Auth/getRuntimeConfig.json";
import startAuthEmailChallengeFixture from "../../../../openapi/fixtures/Auth/startAuthEmailChallenge.json";
import verifyAuthEmailChallengeFixture from "../../../../openapi/fixtures/Auth/verifyAuthEmailChallenge.json";
import { EasyInterviewClient } from "../../api/generated/client";
import {
  createFixtureBackedFetch,
  createFixtureRegistry,
} from "../../api/mockTransport";
import { App } from "../App";
import { useRequestAuth, type PendingAction } from "../auth";
import { useNavigation } from "../navigation/NavigationProvider";

/** Representative raw markers — must never appear in URL / history / storage. */
const RAW_MARKERS = {
  rawText: "RAW_JD_TEXT_2c1a",
  rawDescription: "RAW_DESCRIPTION_3d2b",
  sourceUrl: "https://leaked.example.com/jd/4e3c",
  query: "RAW_QUERY_5f4d",
  label: "RAW_LABEL_6a5e",
  guidedAnswers: "RAW_GUIDED_ANSWER_7b6f",
  parsedSummary: "RAW_PARSED_SUMMARY_8c7a",
  structuredProfile: "RAW_STRUCTURED_PROFILE_9d8b",
  suggestion: "RAW_SUGGESTION_aebc",
  originalBullet: "RAW_BULLET_ORIGINAL_bfcd",
  suggestedBullet: "RAW_BULLET_SUGGESTED_c0de",
  questionText: "RAW_QUESTION_TEXT_d1ef",
  answerText: "RAW_ANSWER_TEXT_e2f0",
  notes: "RAW_DEBRIEF_NOTES_f301",
  prompt: "RAW_AI_PROMPT_0412",
  response: "RAW_AI_RESPONSE_1523",
  file: "RAW_BINARY_BLOB_2634",
  token: "AUTH_SECRET_TOKEN_3745",
  password: "AUTH_PASSWORD_4856",
} satisfies Record<string, string>;

function resetWindow(): void {
  delete (window as { __EASYINTERVIEW_INITIAL_ROUTE__?: unknown })
    .__EASYINTERVIEW_INITIAL_ROUTE__;
  window.history.replaceState(null, "", "/");
  window.localStorage.clear();
  window.sessionStorage.clear();
}

let consoleCalls: string[] = [];

beforeEach(() => {
  resetWindow();
  consoleCalls = [];
  vi.spyOn(console, "log").mockImplementation((...args: unknown[]) => {
    consoleCalls.push(args.map(String).join(" "));
  });
  vi.spyOn(console, "warn").mockImplementation((...args: unknown[]) => {
    consoleCalls.push(args.map(String).join(" "));
  });
  vi.spyOn(console, "error").mockImplementation((...args: unknown[]) => {
    consoleCalls.push(args.map(String).join(" "));
  });
});

afterEach(() => {
  resetWindow();
  vi.restoreAllMocks();
});

function captureSurfaces(): Record<string, string> {
  let local = "";
  for (let i = 0; i < window.localStorage.length; i++) {
    const k = window.localStorage.key(i);
    if (k === null) continue;
    local += `${k}=${window.localStorage.getItem(k) ?? ""}\n`;
  }
  let session = "";
  for (let i = 0; i < window.sessionStorage.length; i++) {
    const k = window.sessionStorage.key(i);
    if (k === null) continue;
    session += `${k}=${window.sessionStorage.getItem(k) ?? ""}\n`;
  }
  return {
    url:
      window.location.pathname +
      window.location.search +
      (window.location.hash || ""),
    historyState: JSON.stringify(window.history.state ?? null),
    localStorageDump: local,
    sessionStorageDump: session,
    consoleDump: consoleCalls.join("\n"),
  };
}

function expectNoRawMarkerLeak(label: string): void {
  const surfaces = captureSurfaces();
  for (const [field, marker] of Object.entries(RAW_MARKERS)) {
    for (const [surface, dump] of Object.entries(surfaces)) {
      expect(
        dump.includes(marker),
        `${label} — ${surface} must not contain ${field}=${marker}; got: ${dump}`,
      ).toBe(false);
    }
  }
}

function buildClient(): EasyInterviewClient {
  return new EasyInterviewClient({
    fetch: createFixtureBackedFetch(
      createFixtureRegistry([
        getRuntimeConfigFixture,
        getMeFixture,
        startAuthEmailChallengeFixture,
        verifyAuthEmailChallengeFixture,
      ]),
    ),
  });
}

const PracticePendingTrigger: FC = () => {
  const requestAuth = useRequestAuth();
  return (
    <button
      type="button"
      data-testid="trigger-start-practice-with-raw"
      onClick={() => {
        const action: PendingAction = {
          type: "start_practice",
          label: "立即面试",
          route: "practice",
          params: {
            sessionId: "01918fa0-0000-7000-8000-000000005000",
            planId: "plan-1",
            targetJobId: "tj-1",
            jdId: "jd-1",
            resumeId: "rv-1",
            roundId: "round-1",
            mode: "text",
            modality: "text",
            ...RAW_MARKERS,
          },
        };
        requestAuth(action);
      }}
    >
      立即面试
    </button>
  );
};

describe("E2E.P0.089 auth pendingAction + URL privacy redline", () => {
  it("workspace auto-start pending action: login round-trip restores canonical practice URL with zero raw-marker leak", async () => {
    render(
      <App
        client={buildClient()}
        requestOptions={{
          getMe: { headers: { Prefer: "example=unauthenticated" } },
        }}
      >
        <PracticePendingTrigger />
      </App>,
    );

    await waitFor(() =>
      expect(screen.getByTestId("topbar-user-area")).toHaveAttribute(
        "data-signed-in",
        "false",
      ),
    );

    const user = userEvent.setup();
    await user.click(screen.getByTestId("trigger-start-practice-with-raw"));
    await waitFor(() => screen.getByTestId("route-auth_login"));

    // After redirect to auth_login, the URL must carry only safe pendingAction
    // params + the practice safe params. Raw markers must be absent
    // everywhere.
    expect(window.location.pathname).toBe("/auth/login");
    const beforeLoginSearch = new URLSearchParams(window.location.search);
    expect(beforeLoginSearch.get("pendingRoute")).toBe("practice");
    expect(beforeLoginSearch.get("pendingType")).toBe("start_practice");
    expect(beforeLoginSearch.get("planId")).toBe("plan-1");
    expect(beforeLoginSearch.get("targetJobId")).toBe("tj-1");
    expect(beforeLoginSearch.get("sessionId")).toBe(
      "01918fa0-0000-7000-8000-000000005000",
    );
    expectNoRawMarkerLeak("after redirect to auth_login");

    await user.type(
      screen.getByTestId("auth-login-email"),
      "alice@example.com",
    );
    await user.click(screen.getByTestId("auth-login-submit-email"));
    await waitFor(() => screen.getByTestId("route-auth_verify"));
    await user.type(screen.getByTestId("auth-verify-code"), "654321");
    await user.click(screen.getByTestId("auth-verify-submit"));
    await waitFor(() => screen.getByTestId("practice-screen"));

    // After verify, the canonical practice URL must carry only safe params.
    expect(window.location.pathname).toBe("/practice");
    const afterVerifySearch = new URLSearchParams(window.location.search);
    expect(afterVerifySearch.get("planId")).toBe("plan-1");
    expect(afterVerifySearch.get("targetJobId")).toBe("tj-1");
    expect(afterVerifySearch.get("jdId")).toBe("jd-1");
    expect(afterVerifySearch.get("resumeId")).toBe("rv-1");
    expect(afterVerifySearch.get("roundId")).toBe("round-1");
    expect(afterVerifySearch.get("sessionId")).toBe(
      "01918fa0-0000-7000-8000-000000005000",
    );
    expectNoRawMarkerLeak("after verify restore to practice");
  });

  it("auth/login direct open with hostile raw markers as query keeps only safe params", () => {
    const hostile = new URLSearchParams();
    hostile.set("pendingRoute", "workspace");
    hostile.set("pendingType", "start_practice");
    hostile.set("pendingLabel", "立即面试");
    hostile.set("planId", "plan-1");
    hostile.set("targetJobId", "tj-1");
    for (const [k, v] of Object.entries(RAW_MARKERS)) {
      hostile.set(k, v);
    }
    window.history.replaceState(null, "", `/auth/login?${hostile.toString()}`);
    render(<App client={buildClient()} />);
    expect(window.location.pathname).toBe("/auth/login");
    const safe = new URLSearchParams(window.location.search);
    expect(safe.get("pendingRoute")).toBe("workspace");
    expect(safe.get("pendingType")).toBe("start_practice");
    expect(safe.has("planId")).toBe(false);
    expect(safe.has("targetJobId")).toBe(false);
    expectNoRawMarkerLeak("after hostile direct-open of /auth/login");
  });

  it("browser history popstate from hostile URL scrubs URL and history.state before route restore", async () => {
    render(<App />);

    act(() => {
      window.history.pushState(
        { rawText: RAW_MARKERS.rawText, prompt: RAW_MARKERS.prompt },
        "",
        `/workspace?targetJobId=tj-popstate&rawText=${encodeURIComponent(
          RAW_MARKERS.rawText,
        )}#${encodeURIComponent(RAW_MARKERS.prompt)}`,
      );
      window.dispatchEvent(
        new PopStateEvent("popstate", { state: window.history.state }),
      );
    });

    await waitFor(() => screen.getByTestId("workspace-plan-list"));
    expect(
      window.location.pathname + window.location.search + window.location.hash,
    ).toBe("/workspace");
    expectNoRawMarkerLeak("after hostile popstate route restore");
  });
});
