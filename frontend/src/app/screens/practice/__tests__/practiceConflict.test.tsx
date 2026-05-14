/**
 * @vitest-environment jsdom
 *
 * Defensive strict-hint conflict coverage. The normal strict UI hides the
 * hint button, but the runtime still maps a backend 409 hint policy conflict
 * to a non-retryable inline warning.
 */

import { describe, expect, it } from "vitest";
import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

import {
  buildPracticeClient,
  eventCalls,
  mountPracticeScreen,
} from "./practiceTestUtils";

describe("practice strict-hint conflict (item 4.4)", () => {
  it("maps hint_disabled_in_mode to the strict hint warning without retry", async () => {
    const { client, calls } = buildPracticeClient({
      scenarioByOp: { appendSessionEvent: "hint-strict-conflict" },
    });
    mountPracticeScreen({ client });

    const user = userEvent.setup();
    await waitFor(() =>
      expect(
        (screen.getByTestId("practice-input-hint") as HTMLButtonElement).disabled,
      ).toBe(false),
    );
    await user.click(screen.getByTestId("practice-input-hint"));

    await waitFor(() => {
      expect(screen.getByTestId("practice-error-state-message").textContent).toContain(
        "Hints are disabled",
      );
    });
    expect(screen.queryByTestId("practice-error-state-retry")).toBeNull();
    expect(eventCalls(calls).length).toBe(1);
  });

  it("strict mode still hides the hint button at the DOM boundary", async () => {
    mountPracticeScreen({ routeParams: { practiceMode: "strict" } });
    await waitFor(() => expect(screen.getByTestId("practice-screen")).toBeDefined());
    expect(screen.queryByTestId("practice-input-hint")).toBeNull();
  });
});
