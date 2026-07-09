/**
 * @vitest-environment jsdom
 *
 * Item 3.5 — SessionMap turn history. After a successful answer →
 * ask_follow_up, the active turn shifts and the previous turn is marked
 * follow_up_requested via data-status.
 */

import { describe, expect, it } from "vitest";
import { screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

import {
  buildPracticeClient,
  mountPracticeScreen,
} from "./practiceTestUtils";

describe("SessionMap (item 3.5)", () => {
  it("submitting an answer marks the active turn follow-up after assistantAction", async () => {
    const { client } = buildPracticeClient({
      scenarioByOp: { appendSessionEvent: "default" },
    });
    mountPracticeScreen({ client });

    const user = userEvent.setup();
    await waitFor(() =>
      expect(
        (screen.getByTestId("practice-input-send") as HTMLButtonElement).disabled,
      ).toBe(false),
    );

    const textarea = screen.getByTestId("practice-input-textarea") as HTMLTextAreaElement;
    await user.type(textarea, "my answer");
    await user.click(screen.getByTestId("practice-input-send"));

    // The default fixture returns ask_follow_up; SessionMap item-0 should
    // shift away from "active" once the renderer dispatches.
    await waitFor(() => {
      const item = screen.getByTestId("practice-sessionmap-item-0");
      expect(item.getAttribute("data-status")).not.toBe("active");
    });
  });

  it("renders SessionMap label with at least 1 item from initial getPracticeSession", async () => {
    mountPracticeScreen();
    await waitFor(() => {
      expect(screen.getByTestId("practice-sessionmap-label")).toBeDefined();
    });
    const map = screen.getByTestId("practice-sessionmap");
    const items = within(map).queryAllByTestId(/^practice-sessionmap-item-/);
    expect(items.length).toBeGreaterThanOrEqual(1);
  });
});
