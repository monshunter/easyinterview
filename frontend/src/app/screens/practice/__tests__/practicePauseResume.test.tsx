/**
 * @vitest-environment jsdom
 *
 * Item 3.3 (pause-resume slice) — clicking pause posts session_paused;
 * clicking resume posts session_resumed; while paused, submit / hint /
 * skip buttons are disabled and do not POST.
 */

import { describe, expect, it } from "vitest";
import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

import {
  buildPracticeClient,
  eventCalls,
  mountPracticeScreen,
  readBody,
} from "./practiceTestUtils";

describe("practice pause / resume (item 3.3)", () => {
  it("clicking pause posts session_paused; resuming posts session_resumed", async () => {
    const { client, calls } = buildPracticeClient({
      scenarioByOp: { appendSessionEvent: "pause-resume" },
    });
    mountPracticeScreen({ client });

    const user = userEvent.setup();
    await waitFor(() =>
      expect(screen.getByTestId("practice-topbar-pause")).toBeDefined(),
    );

    await user.click(screen.getByTestId("practice-topbar-pause"));
    await waitFor(() => {
      expect(eventCalls(calls).length).toBeGreaterThanOrEqual(1);
    });
    expect(readBody(eventCalls(calls).at(-1)!).kind).toBe("session_paused");

    // Pause → resume button click again
    await user.click(screen.getByTestId("practice-topbar-pause"));
    await waitFor(() => {
      expect(eventCalls(calls).length).toBeGreaterThanOrEqual(2);
    });
    expect(readBody(eventCalls(calls).at(-1)!).kind).toBe("session_resumed");
  });

  it("while paused, submit / hint / skip do not post", async () => {
    const { client, calls } = buildPracticeClient({
      scenarioByOp: { appendSessionEvent: "pause-resume" },
    });
    mountPracticeScreen({ client });

    const user = userEvent.setup();
    await waitFor(() =>
      expect(screen.getByTestId("practice-topbar-pause")).toBeDefined(),
    );
    await user.click(screen.getByTestId("practice-topbar-pause"));
    await waitFor(() => {
      expect(eventCalls(calls).length).toBeGreaterThanOrEqual(1);
    });
    const before = eventCalls(calls).length;

    // Buttons should be disabled
    expect(
      (screen.getByTestId("practice-input-send") as HTMLButtonElement).disabled,
    ).toBe(true);
    expect(
      (screen.getByTestId("practice-input-hint") as HTMLButtonElement).disabled,
    ).toBe(true);
    expect(
      (screen.getByTestId("practice-input-skip") as HTMLButtonElement).disabled,
    ).toBe(true);

    // Clicks while disabled fire no events
    await user.click(screen.getByTestId("practice-input-send"));
    await user.click(screen.getByTestId("practice-input-hint"));
    await user.click(screen.getByTestId("practice-input-skip"));
    expect(eventCalls(calls).length).toBe(before);
  });
});
