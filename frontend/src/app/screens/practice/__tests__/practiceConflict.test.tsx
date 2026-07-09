/**
 * @vitest-environment jsdom
 *
 * Hint conflict coverage. Hints are optional in-session assistance, so the
 * frontend keeps the hint control visible and maps backend conflicts to the
 * normal session-conflict recovery path.
 */

import { describe, expect, it } from "vitest";
import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

import {
  buildPracticeClient,
  eventCalls,
  mountPracticeScreen,
} from "./practiceTestUtils";

describe("practice hint conflict (item 4.4)", () => {
  it("maps backend hint conflicts to session recovery without strict-mode UI", async () => {
    const { client, calls } = buildPracticeClient({
      scenarioByOp: { appendSessionEvent: "hint-conflict" },
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
        "Sync conflict",
      );
    });
    expect(screen.queryByTestId("practice-topbar-strict")).toBeNull();
    expect(screen.queryByTestId("practice-error-state-retry")).toBeNull();
    expect(eventCalls(calls).length).toBe(1);
  });

  it("legacy strict route params keep the hint control available", async () => {
    mountPracticeScreen({ routeParams: { practiceMode: "strict" } });
    await waitFor(() => expect(screen.getByTestId("practice-screen")).toBeDefined());
    expect(screen.getByTestId("practice-input-hint")).toBeDefined();
    expect(screen.queryByTestId("practice-topbar-strict")).toBeNull();
  });
});
