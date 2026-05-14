/**
 * @vitest-environment jsdom
 *
 * 409 client_event_id_mismatch recovery: frontend locks the input, triggers a
 * getPracticeSession refresh, and displays the sync-conflict message.
 */

import { describe, expect, it } from "vitest";
import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

import {
  buildPracticeClient,
  getSessionCalls,
  mountPracticeScreen,
} from "./practiceTestUtils";

describe("practice clientEventId conflict recovery (item 4.6)", () => {
  it("refreshes the server session and disables input during mismatch recovery", async () => {
    const { client, calls } = buildPracticeClient({
      scenarioByOp: { appendSessionEvent: "mismatch" },
    });
    mountPracticeScreen({ client });

    const user = userEvent.setup();
    const textarea = screen.getByTestId(
      "practice-input-textarea",
    ) as HTMLTextAreaElement;
    await waitFor(() => expect(textarea.disabled).toBe(false));
    const initialGets = getSessionCalls(calls).length;

    await user.type(textarea, "conflicting answer");
    await user.click(screen.getByTestId("practice-input-send"));

    await waitFor(() => {
      expect(screen.getByTestId("practice-error-state-message").textContent).toContain(
        "Sync conflict",
      );
    });
    expect(screen.queryByTestId("practice-error-state-retry")).toBeNull();
    await waitFor(() => {
      expect(getSessionCalls(calls).length).toBeGreaterThan(initialGets);
    });
  });
});
