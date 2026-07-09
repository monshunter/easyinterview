/**
 * @vitest-environment jsdom
 *
 * Phase 4 error recovery:
 * - append AI timeout keeps the answer in the textarea and retries with the
 *   same clientEventId
 * - complete network/5xx failures retry with the same Idempotency-Key and
 *   expose the back-to-workspace fallback after 3 failed attempts
 */

import { describe, expect, it } from "vitest";
import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

import {
  buildPracticeClient,
  completeCalls,
  eventCalls,
  mountPracticeScreen,
  readBody,
} from "./practiceTestUtils";

describe("practice error recovery (items 4.4 / 4.6)", () => {
  it("append AI timeout surfaces a retryable error and retry reuses clientEventId", async () => {
    const { client, calls } = buildPracticeClient({
      scenarioByOp: { appendSessionEvent: "ai-timeout" },
    });
    mountPracticeScreen({ client });

    const user = userEvent.setup();
    const textarea = screen.getByTestId(
      "practice-input-textarea",
    ) as HTMLTextAreaElement;
    await waitFor(() => expect(textarea.disabled).toBe(false));

    await user.type(textarea, "keep this answer for retry");
    await user.click(screen.getByTestId("practice-input-send"));

    await waitFor(() => {
      expect(screen.getByTestId("practice-error-state-message").textContent).toContain(
        "AI request timed out",
      );
    });
    expect(textarea.value).toBe("keep this answer for retry");

    const first = readBody(eventCalls(calls).at(-1)!);
    await user.click(screen.getByTestId("practice-error-state-retry"));

    await waitFor(() => {
      expect(eventCalls(calls).length).toBeGreaterThanOrEqual(2);
    });
    const second = readBody(eventCalls(calls).at(-1)!);
    expect(second.clientEventId).toBe(first.clientEventId);
  });

  it("complete network failures retry the same key and then expose back-to-workspace", async () => {
    const { client, calls } = buildPracticeClient({
      forceCompleteFailFirstN: 3,
    });
    mountPracticeScreen({ client });

    const user = userEvent.setup();
    const finish = screen.getByTestId("practice-finish-cta") as HTMLButtonElement;
    await waitFor(() => expect(finish.disabled).toBe(false));

    await user.click(finish);
    await waitFor(() =>
      expect(screen.getByTestId("practice-error-state-message").textContent).toContain(
        "Network error",
      ),
    );
    const firstKey = completeCalls(calls).at(-1)!.headers.get("Idempotency-Key");

    await user.click(screen.getByTestId("practice-error-state-retry"));
    await waitFor(() => expect(completeCalls(calls).length).toBe(2));
    expect(completeCalls(calls).at(-1)!.headers.get("Idempotency-Key")).toBe(
      firstKey,
    );

    await user.click(screen.getByTestId("practice-error-state-retry"));
    await waitFor(() => expect(completeCalls(calls).length).toBe(3));
    expect(screen.getByTestId("practice-error-back-to-workspace")).toBeDefined();
  });
});
