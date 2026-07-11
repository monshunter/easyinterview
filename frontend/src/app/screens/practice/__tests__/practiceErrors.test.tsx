/**
 * @vitest-environment jsdom
 *
 * Phase 4 error recovery:
 * - generated-question failures degrade to session_wait, keep the answer in
 *   the textarea, and a new submit mints a new clientEventId
 * - replay fixtures preserve the original successful assistant snapshot
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
  it("session_wait retains the answer without a transcript duplicate and a new submit mints a new clientEventId", async () => {
    const { client, calls } = buildPracticeClient({
      scenarioByOp: { appendSessionEvent: "ai-timeout" },
    });
    mountPracticeScreen({ client });

    const user = userEvent.setup();
    const textarea = screen.getByTestId(
      "practice-input-textarea",
    ) as HTMLTextAreaElement;
    await waitFor(() => expect(textarea.disabled).toBe(false));
    const answer = "保留这段回答，等待服务端重新生成问题";
    await user.type(textarea, answer);

    await user.click(screen.getByTestId("practice-input-send"));
    await waitFor(() => expect(eventCalls(calls)).toHaveLength(1));
    expect(textarea.value).toBe(answer);
    expect(screen.getByTestId("practice-transcript")).not.toHaveTextContent(answer);
    const first = readBody(eventCalls(calls)[0]!);

    await user.click(screen.getByTestId("practice-input-send"));
    await waitFor(() => expect(eventCalls(calls)).toHaveLength(2));
    const second = readBody(eventCalls(calls)[1]!);
    expect(second.clientEventId).not.toBe(first.clientEventId);
    expect(textarea.value).toBe(answer);
    expect(screen.getByTestId("practice-transcript")).not.toHaveTextContent(answer);
  });

  it("append replay renders the original successful snapshot once", async () => {
    const { client, calls } = buildPracticeClient({
      scenarioByOp: { appendSessionEvent: "replay" },
    });
    mountPracticeScreen({ client });

    const user = userEvent.setup();
    const textarea = screen.getByTestId(
      "practice-input-textarea",
    ) as HTMLTextAreaElement;
    await waitFor(() => expect(textarea.disabled).toBe(false));

    const answer = "这次回答应采用原始成功快照";
    await user.type(textarea, answer);
    await user.click(screen.getByTestId("practice-input-send"));

    await waitFor(() => expect(eventCalls(calls)).toHaveLength(1));
    expect(textarea.value).toBe("");
    const transcript = screen.getByTestId("practice-transcript");
    expect(transcript).toHaveTextContent(answer);
    expect(transcript).toHaveTextContent(
      "在 12 个团队里推动迁移时，最大的反对意见是什么？你怎么处理的？",
    );
    expect(screen.queryByTestId("practice-error-state-message")).toBeNull();
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
