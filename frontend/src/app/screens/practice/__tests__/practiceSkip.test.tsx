/**
 * @vitest-environment jsdom
 *
 * Item 3.3 (skip slice) — clicking Skip posts turn_skipped via
 * appendSessionEvent and SessionMap reflects the skipped status on the
 * advanced turn.
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

describe("practice skip (item 3.3)", () => {
  it("clicking skip posts turn_skipped with the active turnId", async () => {
    const { client, calls } = buildPracticeClient({
      scenarioByOp: { appendSessionEvent: "turn-skipped" },
    });
    mountPracticeScreen({ client });

    const user = userEvent.setup();
    await waitFor(() =>
      expect(
        (screen.getByTestId("practice-input-skip") as HTMLButtonElement).disabled,
      ).toBe(false),
    );
    await user.click(screen.getByTestId("practice-input-skip"));

    await waitFor(() => {
      expect(eventCalls(calls).length).toBeGreaterThanOrEqual(1);
    });
    const body = readBody(eventCalls(calls).at(-1)!);
    expect(body.kind).toBe("turn_skipped");
    expect(body.payload).toEqual({ turnId: TURN_A });
  });
});
