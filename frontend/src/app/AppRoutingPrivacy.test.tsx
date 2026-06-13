// @vitest-environment jsdom
/**
 * Plan 004 Phase 3.2 — URL / privacy redline.
 *
 * Asserts that representative raw JD, resume text, guided answers, parsed
 * summary, suggestion text, AI prompt / response and auth secret markers
 * have ZERO hits in:
 *   - URL (window.location.pathname + search + hash)
 *   - History state (window.history.state)
 *   - pendingAction encoded params
 *   - localStorage / sessionStorage
 *
 * Driven by the App + route store + pendingAction encode/decode safe-param
 * allowlist. The test deliberately seeds raw markers as route params /
 * pendingAction params to prove the allowlist actually drops them.
 */
import { afterEach, beforeEach, describe, expect, it } from "vitest";
import { act, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import type { FC } from "react";

import { App } from "./App";
import { useNavigation } from "./navigation/NavigationProvider";
import { encodePendingAction } from "./auth/pendingAction";

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

beforeEach(resetWindow);
afterEach(resetWindow);

const NavTrigger: FC<{
  testid: string;
  to: { name: string; params: Record<string, string> };
}> = ({ testid, to }) => {
  const { navigate } = useNavigation();
  return (
    <button type="button" data-testid={testid} onClick={() => navigate(to)}>
      go
    </button>
  );
};

function captureSurfaces(): {
  url: string;
  historyState: string;
  localStorageDump: string;
  sessionStorageDump: string;
} {
  const url =
    window.location.pathname +
    window.location.search +
    (window.location.hash || "");
  const historyState = JSON.stringify(window.history.state ?? null);
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
    url,
    historyState,
    localStorageDump: local,
    sessionStorageDump: session,
  };
}

function expectNoRawMarkerLeak(): void {
  const surfaces = captureSurfaces();
  for (const [field, marker] of Object.entries(RAW_MARKERS)) {
    expect(
      surfaces.url.includes(marker),
      `URL must not contain ${field}=${marker} — URL is ${surfaces.url}`,
    ).toBe(false);
    expect(
      surfaces.historyState.includes(marker),
      `history.state must not contain ${field}=${marker} — state is ${surfaces.historyState}`,
    ).toBe(false);
    expect(
      surfaces.localStorageDump.includes(marker),
      `localStorage must not contain ${field}=${marker}`,
    ).toBe(false);
    expect(
      surfaces.sessionStorageDump.includes(marker),
      `sessionStorage must not contain ${field}=${marker}`,
    ).toBe(false);
  }
}

describe("Plan 004 Phase 3.2 — URL / privacy redline", () => {
  it("navigate(workspace) with raw markers drops every marker from URL / history / storage", async () => {
    render(
      <App>
        <NavTrigger
          testid="go-workspace-raw"
          to={{
            name: "workspace",
            params: {
              targetJobId: "tj-redline",
              resumeId: "rv-redline",
              planId: "plan-redline",
              ...RAW_MARKERS,
            },
          }}
        />
      </App>,
    );
    const user = userEvent.setup();
    await user.click(screen.getByTestId("go-workspace-raw"));
    await waitFor(() => screen.getByTestId("workspace-empty"));
    expectNoRawMarkerLeak();
    // Legitimate handoff keys must still survive.
    expect(window.location.search).toContain("targetJobId=tj-redline");
    expect(window.location.search).toContain("planId=plan-redline");
    expect(window.location.search).toContain("resumeId=rv-redline");
  });

  it("navigate(report) with reportStatus + raw markers keeps handoff keys but drops raw markers", async () => {
    render(
      <App>
        <NavTrigger
          testid="go-report-raw"
          to={{
            name: "report",
            params: {
              sessionId: "s-1",
              reportId: "rpt-1",
              reportStatus: "failed",
              errorCode: "AI_PROVIDER_TIMEOUT",
              ...RAW_MARKERS,
            },
          }}
        />
      </App>,
    );
    const user = userEvent.setup();
    await user.click(screen.getByTestId("go-report-raw"));
    await waitFor(() => screen.getByTestId("report-failure-state"));
    expectNoRawMarkerLeak();
    expect(window.location.search).toContain("reportId=rpt-1");
    expect(window.location.search).toContain("reportStatus=failed");
    expect(window.location.search).toContain("errorCode=AI_PROVIDER_TIMEOUT");
  });

  it("navigate(jd_match) normalizes to home and drops raw query/label (D-17)", async () => {
    render(
      <App>
        <NavTrigger
          testid="go-jdmatch-raw"
          to={{
            name: "jd_match",
            params: {
              tab: "search",
              query: "RAW_SEARCH_QUERY",
              label: "RAW_SAVED_SEARCH_LABEL",
            },
          }}
        />
      </App>,
    );
    const user = userEvent.setup();
    await user.click(screen.getByTestId("go-jdmatch-raw"));
    await waitFor(() =>
      expect(screen.getByTestId("topbar-nav-home")).toHaveAttribute(
        "aria-current",
        "page",
      ),
    );
    const url = window.location.pathname + window.location.search;
    expect(url).toBe("/");
    expect(url.includes("RAW_")).toBe(false);
  });

  it("encodePendingAction never carries raw markers even when caller passes them", () => {
    const encoded = encodePendingAction({
      type: "start_practice",
      label: "立即面试",
      route: "practice",
      params: {
        planId: "plan-1",
        targetJobId: "tj-1",
        ...RAW_MARKERS,
      },
    });
    const dump = JSON.stringify(encoded);
    for (const [field, marker] of Object.entries(RAW_MARKERS)) {
      expect(
        dump.includes(marker),
        `encodePendingAction leaked ${field}=${marker}: ${dump}`,
      ).toBe(false);
    }
    expect(encoded.planId).toBe("plan-1");
    expect(encoded.targetJobId).toBe("tj-1");
  });

  it("navigating to auth_login with pendingAction never leaks raw markers into URL or history", async () => {
    render(
      <App>
        <NavTrigger
          testid="go-auth-with-pending"
          to={{
            name: "auth_login",
            params: encodePendingAction({
              type: "start_practice",
              label: "立即面试",
              route: "practice",
              params: {
                planId: "plan-1",
                targetJobId: "tj-1",
                ...RAW_MARKERS,
              },
            }),
          }}
        />
      </App>,
    );
    const user = userEvent.setup();
    await user.click(screen.getByTestId("go-auth-with-pending"));
    await waitFor(() => screen.getByTestId("route-auth_login"));
    expectNoRawMarkerLeak();
    expect(window.location.pathname).toBe("/auth/login");
    expect(window.location.search).toContain("pendingRoute=practice");
    expect(window.location.search).toContain("pendingType=start_practice");
    expect(window.location.search).toContain("planId=plan-1");
  });

  it("popstate from hostile history entry canonicalizes URL and clears raw history.state markers", async () => {
    render(<App />);

    act(() => {
      window.history.pushState(
        { rawText: RAW_MARKERS.rawText },
        "",
        `/workspace?targetJobId=tj-popstate&rawText=${encodeURIComponent(
          RAW_MARKERS.rawText,
        )}#${encodeURIComponent(RAW_MARKERS.prompt)}`,
      );
      window.dispatchEvent(
        new PopStateEvent("popstate", { state: window.history.state }),
      );
    });

    await waitFor(() => screen.getByTestId("workspace-empty"));
    expect(
      window.location.pathname + window.location.search + window.location.hash,
    ).toBe("/workspace?targetJobId=tj-popstate");
    expectNoRawMarkerLeak();
  });
});
