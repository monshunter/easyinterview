/** @vitest-environment jsdom */

import { act, renderHook, waitFor } from "@testing-library/react";
import { StrictMode, type ReactNode } from "react";
import { afterEach, describe, expect, it, vi } from "vitest";

import { ApiClientError, EasyInterviewClient } from "../../../../api/generated/client";
import type { PracticeSession, PracticeUserMessage } from "../../../../api/generated/types";
import { createDevMockClient } from "../../../../api/devMockClient";
import { InterviewContextProvider } from "../../../interview-context/InterviewContext";
import { AppRuntimeContext } from "../../../runtime/AppRuntimeProvider";
import { usePracticeSessionLoader } from "./usePracticeSessionLoader";

const SESSION_ID = "01918fa0-0000-7000-8000-000000005000";
const SERVER_MESSAGE_ID = "01918fa0-0000-7000-8000-000000006099";
const READ_TIMEOUT_MS = 10_000;

afterEach(() => vi.useRealTimers());

describe("usePracticeSessionLoader bounded reads", () => {
  it("keeps the mount read signal-free and reports a logical timeout at the read bound", async () => {
    vi.useFakeTimers();
    const client = new EasyInterviewClient({ fetch: vi.fn<typeof fetch>() });
    const getSession = vi.spyOn(client, "getPracticeSession").mockImplementation(
      () => new Promise(() => undefined),
    );
    const { result } = renderHook(() => usePracticeSessionLoader(SESSION_ID), {
      wrapper: ({ children }) => <Wrapper client={client}>{children}</Wrapper>,
    });

    expect(getSession.mock.calls[0]?.[1]).toBeUndefined();

    await act(async () => { await vi.advanceTimersByTimeAsync(READ_TIMEOUT_MS - 1); });
    expect(result.current.state).toBe("loading");
    await act(async () => { await vi.advanceTimersByTimeAsync(1); });
    expect(result.current.state).toBe("error");
    expect(result.current.error).toMatchObject({ kind: "abort" });
  });

  it("keeps explicit caller cancellation for active reads", async () => {
    const client = new EasyInterviewClient({ fetch: vi.fn<typeof fetch>() });
    const getSession = vi.spyOn(client, "getPracticeSession").mockImplementation(
      () => new Promise(() => undefined),
    );
    const { result } = renderHook(() => usePracticeSessionLoader(SESSION_ID), {
      wrapper: ({ children }) => <Wrapper client={client}>{children}</Wrapper>,
    });
    const controller = new AbortController();
    const pending = result.current.read({ signal: controller.signal });
    const signal = getSession.mock.calls.at(-1)?.[1]?.signal;

    expect(signal).toBeInstanceOf(AbortSignal);
    expect(signal?.aborted).toBe(false);
    const rejected = expect(pending.result).rejects.toMatchObject({ kind: "abort" });
    controller.abort();
    await rejected;
    expect(signal?.aborted).toBe(true);
  });

  it("shares the mount read transport under StrictMode", async () => {
    const session = openingOnly(
      await createDevMockClient().getPracticeSession(SESSION_ID),
    );
    let resolveFetch!: (response: Response) => void;
    const fetch = vi.fn<typeof globalThis.fetch>(
      () => new Promise<Response>((resolve) => { resolveFetch = resolve; }),
    );
    const client = new EasyInterviewClient({ fetch });
    const { result } = renderHook(() => usePracticeSessionLoader(SESSION_ID), {
      wrapper: ({ children }) => (
        <StrictMode>
          <DirectWrapper client={client}>{children}</DirectWrapper>
        </StrictMode>
      ),
    });

    expect(fetch).toHaveBeenCalledTimes(1);
    await act(async () => {
      resolveFetch(new Response(JSON.stringify(session), {
        status: 200,
        headers: { "Content-Type": "application/json" },
      }));
    });
    await waitFor(() => expect(result.current.state).toBe("data"));
    expect(result.current.data?.id).toBe(SESSION_ID);
  });

  it("preserves last same-session unresolved facts when a refresh read fails", async () => {
    const fixtureClient = createDevMockClient();
    const base = openingOnly(await fixtureClient.getPracticeSession(SESSION_ID));
    const pending = sessionWithUser(base, "pending reply");
    const client = new EasyInterviewClient({ fetch: vi.fn<typeof fetch>() });
    const getSession = vi.spyOn(client, "getPracticeSession")
      .mockResolvedValueOnce(pending)
      .mockRejectedValueOnce(new ApiClientError("transport", null, null, new TypeError("offline")));
    const { result } = renderHook(() => usePracticeSessionLoader(SESSION_ID), {
      wrapper: ({ children }) => <Wrapper client={client}>{children}</Wrapper>,
    });

    await waitFor(() => expect(result.current.state).toBe("data"));
    act(() => result.current.refresh());
    await waitFor(() => expect(result.current.state).toBe("error"));

    expect(getSession).toHaveBeenCalledTimes(2);
    expect(result.current.data?.messages.at(-1)).toMatchObject({
      clientMessageId: SERVER_MESSAGE_ID,
      replyStatus: "pending",
      content: "pending reply",
    });
  });

  it("does not let an older read overwrite newer adopted server truth", async () => {
    const fixtureClient = createDevMockClient();
    const base = openingOnly(await fixtureClient.getPracticeSession(SESSION_ID));
    const stale = sessionWithUser(base, "stale pending");
    const current = {
      ...sessionWithUser(base, "current complete", "complete"),
      updatedAt: "2026-07-14T10:00:00Z",
    };
    let resolveOld: ((session: PracticeSession) => void) | undefined;
    const client = new EasyInterviewClient({ fetch: vi.fn<typeof fetch>() });
    vi.spyOn(client, "getPracticeSession").mockImplementation(
      () => new Promise((resolve) => { resolveOld = resolve; }),
    );
    const { result } = renderHook(() => usePracticeSessionLoader(SESSION_ID), {
      wrapper: ({ children }) => <Wrapper client={client}>{children}</Wrapper>,
    });

    act(() => { expect(result.current.adopt(current)).toBe(true); });
    await act(async () => { resolveOld?.(stale); });

    expect(result.current.data?.messages.at(-1)).toMatchObject({
      content: "current complete",
      replyStatus: "complete",
    });
  });
});

function Wrapper({ children, client }: { children: ReactNode; client: EasyInterviewClient }) {
  return <DirectWrapper client={client}>{children}</DirectWrapper>;
}

function DirectWrapper({ children, client }: { children: ReactNode; client: EasyInterviewClient }) {
  return (
    <InterviewContextProvider>
      <AppRuntimeContext.Provider
        value={{
          client,
          runtime: { status: "ready", config: {} as never },
          auth: { status: "authenticated", user: {} as never },
          refreshAuth: () => undefined,
        }}
      >
        {children}
      </AppRuntimeContext.Provider>
    </InterviewContextProvider>
  );
}

function openingOnly(session: PracticeSession): PracticeSession {
  return {
    ...session,
    messages: session.messages.filter((message) => message.role === "assistant").slice(0, 1),
  };
}

function sessionWithUser(
  session: PracticeSession,
  content: string,
  replyStatus: PracticeUserMessage["replyStatus"] = "pending",
): PracticeSession {
  const user: PracticeUserMessage = {
    id: SERVER_MESSAGE_ID,
    clientMessageId: SERVER_MESSAGE_ID,
    seqNo: 2,
    role: "user",
    content,
    replyStatus,
    createdAt: "2026-07-14T09:00:00Z",
  };
  return { ...session, messages: [...session.messages, user] };
}
