/**
 * @vitest-environment jsdom
 *
 * Item 4.8 — completion CTA flow:
 *   - clicking the finish CTA fires completePracticeSession
 *   - on 202 success, navigate to `generating` with stable IDs and
 *     PracticeDisplayContext (no raw text in URL)
 *   - StrictMode-style double click dedupes to a single POST + nav
 */

import { describe, expect, it } from "vitest";
import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

import { mountPracticeScreen } from "./practiceTestUtils";
import { EasyInterviewClient } from "../../../../api/generated/client";
import {
  createFixtureBackedFetch,
  createFixtureRegistry,
} from "../../../../api/mockTransport";

import getPracticeSessionFixture from "../../../../../../openapi/fixtures/PracticeSessions/getPracticeSession.json";
import appendSessionEventFixture from "../../../../../../openapi/fixtures/PracticeSessions/appendSessionEvent.json";
import completePracticeSessionFixture from "../../../../../../openapi/fixtures/PracticeSessions/completePracticeSession.json";

interface CapturedRequest {
  url: string;
  method: string;
  bodyText: string | null;
}

function buildFullClient(): {
  client: EasyInterviewClient;
  calls: CapturedRequest[];
} {
  const calls: CapturedRequest[] = [];
  const fixtureFetch = createFixtureBackedFetch(
    createFixtureRegistry([
      getPracticeSessionFixture,
      appendSessionEventFixture,
      completePracticeSessionFixture,
    ]),
    { scenario: "default" },
  );
  const wrappedFetch: typeof fetch = async (input, init) => {
    const url =
      typeof input === "string"
        ? input
        : input instanceof URL
          ? input.href
          : input.url;
    let bodyText: string | null = null;
    if (typeof init?.body === "string") bodyText = init.body;
    calls.push({
      url,
      method: (init?.method ?? "GET").toUpperCase(),
      bodyText,
    });
    return fixtureFetch(input, init);
  };
  return {
    client: new EasyInterviewClient({ fetch: wrappedFetch }),
    calls,
  };
}

function completeCalls(all: CapturedRequest[]): CapturedRequest[] {
  return all.filter(
    (c) =>
      c.method === "POST" &&
      /\/practice\/sessions\/[^/]+\/complete$/.test(
        new URL(c.url, "http://x").pathname,
      ),
  );
}

describe("practice completion CTA (item 4.8)", () => {
  it("clicking finish posts completePracticeSession and navigates to generating", async () => {
    const { client, calls } = buildFullClient();
    const { nav } = mountPracticeScreen({ client });

    const user = userEvent.setup();
    await waitFor(() =>
      expect(screen.getByTestId("practice-finish-cta")).toBeDefined(),
    );

    const cta = screen.getByTestId("practice-finish-cta") as HTMLButtonElement;
    await waitFor(() => expect(cta.disabled).toBe(false));
    await user.click(cta);

    await waitFor(() => {
      expect(completeCalls(calls).length).toBeGreaterThanOrEqual(1);
    });

    await waitFor(() => {
      expect(nav).toHaveBeenCalled();
    });
    const lastNav = nav.mock.calls.at(-1)?.[0];
    expect(lastNav).toBeDefined();
    expect(lastNav.name).toBe("generating");
    expect(lastNav.params.sessionId).toBeTruthy();
    expect(lastNav.params.reportId).toBeTruthy();
    expect(lastNav.params).not.toHaveProperty("answerText");
    expect(lastNav.params).not.toHaveProperty("questionText");
    expect(lastNav.params).not.toHaveProperty("hint");
    expect(lastNav.params).not.toHaveProperty("provenance");
  });

  it("double-clicking finish deduplicates to one POST and one nav", async () => {
    const { client, calls } = buildFullClient();
    const { nav } = mountPracticeScreen({ client });

    const user = userEvent.setup();
    await waitFor(() =>
      expect(screen.getByTestId("practice-finish-cta")).toBeDefined(),
    );
    const cta = screen.getByTestId("practice-finish-cta") as HTMLButtonElement;
    await waitFor(() => expect(cta.disabled).toBe(false));
    await user.click(cta);
    await user.click(cta);

    await waitFor(() => {
      expect(completeCalls(calls).length).toBeGreaterThanOrEqual(1);
    });
    expect(completeCalls(calls).length).toBe(1);
    expect(nav).toHaveBeenCalledTimes(1);
  });
});
