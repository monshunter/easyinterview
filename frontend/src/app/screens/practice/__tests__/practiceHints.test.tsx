/**
 * @vitest-environment jsdom
 *
 * Item 3.2 — assisted mode hint flow:
 *  - clicking the hint button posts hint_requested via appendSessionEvent
 *  - server returns assistantAction.show_hint
 *  - HintBanner becomes visible; hintCount increments via INCREMENT_HINT_COUNT
 *  - clicking hint again hides the banner without re-posting
 *  - strict mode does NOT render the hint button (DOM negative)
 */

import { describe, expect, it } from "vitest";
import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

import {
  buildPracticeClient,
  eventCalls,
  mountPracticeScreen,
  readBody,
  TURN_A,
} from "./practiceTestUtils";

describe("practice hints (item 3.2)", () => {
  it("assisted mode: clicking hint posts hint_requested and renders HintBanner", async () => {
    const { client, calls } = buildPracticeClient({
      scenarioByOp: { appendSessionEvent: "show-hint" },
    });
    mountPracticeScreen({ client });

    // wait for loader to settle (button enabled)
    await waitFor(() => {
      expect(
        (screen.getByTestId("practice-input-hint") as HTMLButtonElement).disabled,
      ).toBe(false);
    });

    const user = userEvent.setup();
    await user.click(screen.getByTestId("practice-input-hint"));

    await waitFor(() => {
      expect(eventCalls(calls).length).toBeGreaterThanOrEqual(1);
    });
    const body = readBody(eventCalls(calls).at(-1)!);
    expect(body.kind).toBe("hint_requested");
    expect(body.payload).toEqual({ turnId: TURN_A });

    await waitFor(() => {
      expect(screen.getByTestId("practice-hint-banner")).toBeDefined();
    });
    // hintCount written into right panel CTA usage note
    await waitFor(() => {
      expect(screen.getByTestId("practice-rightpanel-hint-count")).toBeDefined();
    });
  });

  it("clicking hint again hides the banner without an extra POST", async () => {
    const { client, calls } = buildPracticeClient({
      scenarioByOp: { appendSessionEvent: "show-hint" },
    });
    mountPracticeScreen({ client });

    const user = userEvent.setup();
    await waitFor(() =>
      expect(
        (screen.getByTestId("practice-input-hint") as HTMLButtonElement).disabled,
      ).toBe(false),
    );
    await user.click(screen.getByTestId("practice-input-hint"));
    await waitFor(() =>
      expect(screen.getByTestId("practice-hint-banner")).toBeDefined(),
    );
    const before = eventCalls(calls).length;
    await user.click(screen.getByTestId("practice-input-hint"));
    // Banner hidden, no new request
    await waitFor(() => {
      expect(screen.queryByTestId("practice-hint-banner")).toBeNull();
    });
    expect(eventCalls(calls).length).toBe(before);
  });

  it("strict mode: hint button DOM is not rendered", async () => {
    mountPracticeScreen({
      routeParams: { practiceMode: "strict" },
    });
    await waitFor(() => {
      expect(screen.getByTestId("practice-screen")).toBeDefined();
    });
    expect(screen.queryByTestId("practice-input-hint")).toBeNull();
    expect(screen.getByTestId("practice-rightpanel-strict-banner")).toBeDefined();
  });
});
