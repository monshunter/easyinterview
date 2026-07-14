/** @vitest-environment jsdom */

import { act, renderHook } from "@testing-library/react";
import { type ReactNode } from "react";
import { describe, expect, it, vi } from "vitest";

import { EasyInterviewClient } from "../../../../api/generated/client";
import type { SendPracticeMessageResponse } from "../../../../api/generated/types";
import { InterviewContextProvider } from "../../../interview-context/InterviewContext";
import { AppRuntimeProvider } from "../../../runtime/AppRuntimeProvider";
import { usePracticeMessages } from "./usePracticeMessages";

const SESSION_ID = "01918fa0-0000-7000-8000-000000005000";
const FIRST_ID = "01918fa0-0000-7000-8000-000000006091";
const SECOND_ID = "01918fa0-0000-7000-8000-000000006092";

describe("usePracticeMessages", () => {
  it("sends the caller-owned text/clientMessageId exactly and never guesses retry identity from equal text", async () => {
    const client = new EasyInterviewClient({ fetch: vi.fn<typeof fetch>() });
    const send = vi.spyOn(client, "sendPracticeMessage").mockResolvedValue({} as SendPracticeMessageResponse);
    const { result } = renderHook(() => usePracticeMessages(SESSION_ID), {
      wrapper: ({ children }) => <Wrapper client={client}>{children}</Wrapper>,
    });

    await act(async () => {
      await result.current.sendMessage({ text: "same text", clientMessageId: FIRST_ID });
      await result.current.sendMessage({ text: "same text", clientMessageId: SECOND_ID });
    });

    expect(send.mock.calls.map((call) => call[1])).toEqual([
      { text: "same text", clientMessageId: FIRST_ID },
      { text: "same text", clientMessageId: SECOND_ID },
    ]);
  });

  it("forwards the caller AbortSignal through generated request options", async () => {
    const client = new EasyInterviewClient({ fetch: vi.fn<typeof fetch>() });
    const send = vi.spyOn(client, "sendPracticeMessage").mockResolvedValue({} as SendPracticeMessageResponse);
    const { result } = renderHook(() => usePracticeMessages(SESSION_ID), {
      wrapper: ({ children }) => <Wrapper client={client}>{children}</Wrapper>,
    });
    const controller = new AbortController();
    const submission = { text: "bounded request", clientMessageId: FIRST_ID };

    await act(async () => {
      await result.current.sendMessage(submission, { signal: controller.signal });
    });

    expect(send).toHaveBeenCalledWith(
      SESSION_ID,
      submission,
      { signal: controller.signal },
    );
  });

  it("accepts exact UTF-8 bytes and rejects limit+1 before the generated client", async () => {
		const client = new EasyInterviewClient({ fetch: vi.fn<typeof fetch>() });
		const send = vi.spyOn(client, "sendPracticeMessage").mockResolvedValue({} as SendPracticeMessageResponse);
		const { result } = renderHook(() => usePracticeMessages(SESSION_ID), {
			wrapper: ({ children }) => <Wrapper client={client}>{children}</Wrapper>,
		});
		await act(async () => {
			await result.current.sendMessage({ text: "a".repeat(32_768), clientMessageId: FIRST_ID });
		});
		await expect(result.current.sendMessage({ text: `${"a".repeat(32_768)}b`, clientMessageId: SECOND_ID }))
			.rejects.toThrow("PRACTICE_MESSAGE_TOO_LARGE");
		expect(send).toHaveBeenCalledTimes(1);
	});
});

function Wrapper({ children, client }: { children: ReactNode; client: EasyInterviewClient }) {
  return (
    <InterviewContextProvider>
      <AppRuntimeProvider client={client}>{children}</AppRuntimeProvider>
    </InterviewContextProvider>
  );
}
