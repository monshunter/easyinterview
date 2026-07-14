/** @vitest-environment jsdom */
import { readFileSync } from "node:fs";
import { resolve } from "node:path";

import { act, fireEvent, render, screen, waitFor, within } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { ApiClientError } from "../../../api/generated/client";
import type {
  PracticeAssistantMessage,
  PracticeSession,
  PracticeUserMessage,
  ReportWithJob,
  SendPracticeMessageResponse,
} from "../../../api/generated/types";
import { createDevMockClient } from "../../../api/devMockClient";
import { App } from "../../App";

afterEach(() => {
  vi.useRealTimers();
  localStorage.removeItem("ei-lang");
  window.history.replaceState(null, "", "/");
});

describe("PracticeScreen continuous conversation", () => {
  it("traces the terminal CTA to the prototype secondary/sm button interaction contract", () => {
    const prototypeScreen = readFileSync(
      resolve(__dirname, "../../../../../ui-design/src/screen-practice.jsx"),
      "utf8",
    );
    const prototypePrimitives = readFileSync(
      resolve(__dirname, "../../../../../ui-design/src/primitives.jsx"),
      "utf8",
    );
    const formal = readFileSync(
      resolve(__dirname, "components/TerminalRecovery.tsx"),
      "utf8",
    );

    expect(prototypeScreen).toContain('<Btn variant="secondary" size="sm"');
    for (const sourceLiteral of [
      'transition: "transform .08s ease, opacity .15s"',
      'style.transform = "translateY(0.5px)"',
      'style.transform = ""',
    ]) {
      expect(prototypePrimitives).toContain(sourceLiteral);
      expect(formal).toContain(sourceLiteral);
    }
    expect(formal).toContain("opacity: 1");
  });

  it("ZERO_ANSWER_FINISH_DISABLED_PASS: keeps Finish natively disabled with a localized accessible reason until a committed answer is complete", async () => {
    localStorage.setItem("ei-lang", "zh");
    const client = createDevMockClient();
    const session = await client.getPracticeSession("01918fa0-0000-7000-8000-000000005000");
    const complete = vi.spyOn(client, "completePracticeSession");
    vi.spyOn(client, "getPracticeSession").mockResolvedValue({
      ...session,
      messages: session.messages.filter((message) => message.role === "assistant").slice(0, 1),
    });

    render(<App client={client} requestOptions={{ getMe: { headers: { Prefer: "example=authenticated" } } }} initialRoute={{ name: "practice", params: {
      sessionId: "01918fa0-0000-7000-8000-000000005000",
      planId: "01918fa0-0000-7000-8000-000000004000",
      targetJobId: "01918fa0-0000-7000-8000-000000002000",
      reportId: "route-report-must-not-count",
    } }} />);

    await screen.findByText("你好，我们直接开始。先聊聊你最近最有代表性的项目。");
    fireEvent.change(screen.getByTestId("practice-input-textarea"), { target: { value: "尚未提交的草稿" } });
    const finish = screen.getByTestId("practice-finish-cta");
    const reason = screen.getByTestId("practice-finish-disabled-reason");
    expect(finish).toBeDisabled();
    expect(finish).toHaveAttribute("aria-describedby", reason.id);
    expect(reason).toHaveTextContent("请先完成至少一次回答");
    fireEvent.click(finish);
    expect(complete).not.toHaveBeenCalled();
  });

  it("keeps Finish disabled while the latest committed candidate message still awaits an assistant reply", async () => {
    const client = createDevMockClient();
    const session = await client.getPracticeSession("01918fa0-0000-7000-8000-000000005000");
    vi.spyOn(client, "getPracticeSession").mockResolvedValue(
      sessionWithUser(
        openingOnly(session),
        "01918fa0-0000-7000-8000-000000006001",
        "我主导过一次跨团队设计系统迁移。",
        "pending",
      ),
    );

    render(<App client={client} requestOptions={{ getMe: { headers: { Prefer: "example=authenticated" } } }} initialRoute={{ name: "practice", params: {
      sessionId: "01918fa0-0000-7000-8000-000000005000",
      planId: "01918fa0-0000-7000-8000-000000004000",
      targetJobId: "01918fa0-0000-7000-8000-000000002000",
    } }} />);

    await screen.findByText("我主导过一次跨团队设计系统迁移。");
    expect(screen.getByTestId("practice-finish-cta")).toBeDisabled();
  });

  it("renders the persisted practice plan time budget", async () => {
    const client = createDevMockClient();
    const getPlan = vi.spyOn(client, "getPracticePlan").mockResolvedValue({
      id: "01918fa0-0000-7000-8000-000000004000",
      targetJobId: "01918fa0-0000-7000-8000-000000002000",
      goal: "baseline",
      interviewerPersona: "hiring_manager",
      difficulty: "standard",
      language: "zh-CN",
      timeBudgetMinutes: 60,
      resumeId: "01918fa0-0000-7000-8000-000000001000",
      status: "ready",
      createdAt: "2026-07-12T08:00:00Z",
    });

    render(<App client={client} requestOptions={{ getMe: { headers: { Prefer: "example=authenticated" } } }} initialRoute={{ name: "practice", params: {
      sessionId: "01918fa0-0000-7000-8000-000000005000",
      planId: "01918fa0-0000-7000-8000-000000004000",
      targetJobId: "01918fa0-0000-7000-8000-000000002000",
    } }} />);

    await waitFor(() => expect(getPlan).toHaveBeenCalled());
    await waitFor(() => {
      expect(screen.getByTestId("practice-topbar-timer")).toHaveTextContent("/ 60:00");
    });
  });

  it("does not fabricate a budget when the practice plan cannot be read", async () => {
    const client = createDevMockClient();
    vi.spyOn(client, "getPracticePlan").mockRejectedValue(new Error("HTTP 503 plan unavailable"));

    render(<App client={client} requestOptions={{ getMe: { headers: { Prefer: "example=authenticated" } } }} initialRoute={{ name: "practice", params: {
      sessionId: "01918fa0-0000-7000-8000-000000005000",
      planId: "01918fa0-0000-7000-8000-000000004000",
      targetJobId: "01918fa0-0000-7000-8000-000000002000",
    } }} />);

    await screen.findByText("你好，我们直接开始。先聊聊你最近最有代表性的项目。");
    await waitFor(() => {
      expect(screen.getByTestId("practice-topbar-timer")).toHaveTextContent("/ --:--");
    });
    expect(screen.getByTestId("practice-topbar-timer")).not.toHaveTextContent("25:00");
  });

  it("renders ordered messages in one chat window with voice disabled", async () => {
    render(<App client={createDevMockClient()} requestOptions={{ getMe: { headers: { Prefer: "example=authenticated" } } }} initialRoute={{ name: "practice", params: {
      sessionId: "01918fa0-0000-7000-8000-000000005000",
      planId: "01918fa0-0000-7000-8000-000000004000",
      targetJobId: "01918fa0-0000-7000-8000-000000002000",
    } }} />);

    expect(await screen.findByText("你好，我们直接开始。先聊聊你最近最有代表性的项目。")).toBeInTheDocument();
    expect(screen.getByText("我主导过一次跨团队设计系统迁移。")).toBeInTheDocument();
    expect(screen.getByText("当时最难协调的分歧是什么？")).toBeInTheDocument();
    expect(screen.getByTestId("practice-topbar-phone-toggle")).toBeDisabled();
    expect(screen.queryByText(/题\s*\d+\s*\/\s*\d+/)).not.toBeInTheDocument();
    expect(screen.queryByTestId("practice-session-map")).not.toBeInTheDocument();
    expect(screen.queryByTestId("practice-question-card")).not.toBeInTheDocument();
  });

  it("does not expose route or business identifiers as DOM debug metadata", async () => {
    renderPractice(createDevMockClient());

    const practice = await screen.findByTestId("practice-screen");
    expect(practice).not.toHaveAttribute("data-session-id");
    expect(practice).not.toHaveAttribute("data-plan-id");
    expect(practice).not.toHaveAttribute("data-target-job-id");
  });

  it("retries a completion failure through completePracticeSession without sending a draft", async () => {
    const client = createDevMockClient();
    const complete = vi.spyOn(client, "completePracticeSession");
    const originalComplete = complete.getMockImplementation();
    complete.mockRejectedValueOnce(new Error("HTTP 503 completion unavailable"));
    if (originalComplete) complete.mockImplementationOnce(originalComplete);
    const send = vi.spyOn(client, "sendPracticeMessage");

    render(<App client={client} requestOptions={{ getMe: { headers: { Prefer: "example=authenticated" } } }} initialRoute={{ name: "practice", params: {
      sessionId: "01918fa0-0000-7000-8000-000000005000",
      planId: "01918fa0-0000-7000-8000-000000004000",
      targetJobId: "01918fa0-0000-7000-8000-000000002000",
    } }} />);

    await screen.findByText("你好，我们直接开始。先聊聊你最近最有代表性的项目。");
    fireEvent.click(screen.getByTestId("practice-finish-cta"));
    expect(await screen.findByText("We could not finish the interview and start the report. Please try again.")).toBeInTheDocument();
    expect(screen.queryByText("HTTP 503 completion unavailable")).not.toBeInTheDocument();
    fireEvent.click(screen.getByTestId("practice-error-state-retry"));

    await waitFor(() => expect(complete).toHaveBeenCalledTimes(2));
    expect(send).not.toHaveBeenCalled();
  });

  it("hands completion to Generating with reportId as the only URL and history locator", async () => {
    const client = createDevMockClient();
    const reportId = "01918fa0-0000-7000-8000-000000007000";
    const queuedReport = await client.getFeedbackReport(reportId, {
      headers: { Prefer: "example=queued" },
    });
    const getReport = vi.spyOn(client, "getFeedbackReport").mockResolvedValue({
      ...queuedReport,
      id: reportId,
    });
    const complete = vi.spyOn(client, "completePracticeSession");
    window.history.replaceState(
      { stale: "must-be-replaced" },
      "",
      "/practice?sessionId=01918fa0-0000-7000-8000-000000005000&planId=01918fa0-0000-7000-8000-000000004000&targetJobId=01918fa0-0000-7000-8000-000000002000&resumeId=01918fa0-0000-7000-8000-000000001000&roundId=round-1-technical&roundName=Technical",
    );

    render(
      <App
        client={client}
        requestOptions={{
          getMe: { headers: { Prefer: "example=authenticated" } },
        }}
      />,
    );

    await screen.findByText("你好，我们直接开始。先聊聊你最近最有代表性的项目。");
    const finish = screen.getByTestId("practice-finish-cta");
    expect(finish).toBeEnabled();
    fireEvent.click(finish);

    await waitFor(() => expect(complete).toHaveBeenCalledTimes(1));
    await waitFor(() => {
      expect(window.location.pathname).toBe("/generating");
      expect(window.location.search).toBe(`?reportId=${reportId}`);
    });
    await waitFor(() => expect(getReport).toHaveBeenCalledWith(reportId));
    expect(window.history.state).toBeNull();
    const searchParams = new URL(window.location.href).searchParams;
    for (const forbidden of [
      "targetJobId",
      "planId",
      "sessionId",
      "resumeId",
      "roundId",
      "roundName",
      "status",
      "error",
    ]) {
      expect(searchParams.has(forbidden)).toBe(false);
    }
  });

  it("retries an initial loader failure through getPracticeSession", async () => {
    const client = createDevMockClient();
    const getSession = vi.spyOn(client, "getPracticeSession");
    const originalGetSession = getSession.getMockImplementation();
    getSession.mockRejectedValueOnce(new Error("HTTP 503 session unavailable"));
    if (originalGetSession) getSession.mockImplementationOnce(originalGetSession);

    render(<App client={client} requestOptions={{ getMe: { headers: { Prefer: "example=authenticated" } } }} initialRoute={{ name: "practice", params: {
      sessionId: "01918fa0-0000-7000-8000-000000005000",
      planId: "01918fa0-0000-7000-8000-000000004000",
      targetJobId: "01918fa0-0000-7000-8000-000000002000",
    } }} />);

    expect(await screen.findByText("The interview session is temporarily unavailable. Please try again.")).toBeInTheDocument();
    expect(screen.queryByText("HTTP 503 session unavailable")).not.toBeInTheDocument();
    expect(screen.getByTestId("practice-input-textarea")).toBeDisabled();
    expect(screen.getByTestId("practice-input-send")).toBeDisabled();
    fireEvent.click(screen.getByTestId("practice-error-state-retry"));

    expect(await screen.findByText("你好，我们直接开始。先聊聊你最近最有代表性的项目。")).toBeInTheDocument();
    expect(getSession).toHaveBeenCalledTimes(2);
  });

  it("disables Finish while the session is loading or a message is sending", async () => {
    const loadingClient = createDevMockClient();
    vi.spyOn(loadingClient, "getPracticeSession").mockImplementation(() => new Promise(() => undefined));
    const firstRender = render(<App client={loadingClient} requestOptions={{ getMe: { headers: { Prefer: "example=authenticated" } } }} initialRoute={{ name: "practice", params: {
      sessionId: "01918fa0-0000-7000-8000-000000005000",
      planId: "01918fa0-0000-7000-8000-000000004000",
      targetJobId: "01918fa0-0000-7000-8000-000000002000",
    } }} />);

    expect(await screen.findByTestId("practice-finish-cta")).toBeDisabled();
    firstRender.unmount();

    const sendingClient = createDevMockClient();
    vi.spyOn(sendingClient, "sendPracticeMessage").mockImplementation(() => new Promise(() => undefined));
    render(<App client={sendingClient} requestOptions={{ getMe: { headers: { Prefer: "example=authenticated" } } }} initialRoute={{ name: "practice", params: {
      sessionId: "01918fa0-0000-7000-8000-000000005000",
      planId: "01918fa0-0000-7000-8000-000000004000",
      targetJobId: "01918fa0-0000-7000-8000-000000002000",
    } }} />);

    await screen.findByText("你好，我们直接开始。先聊聊你最近最有代表性的项目。");
    fireEvent.change(screen.getByTestId("practice-input-textarea"), { target: { value: "继续" } });
    fireEvent.click(screen.getByTestId("practice-input-send"));
    await waitFor(() => expect(screen.getByTestId("practice-input-textarea")).toBeDisabled());
    expect(screen.getByTestId("practice-finish-cta")).toBeDisabled();
  });

  it("appends the user row and clears/locks the composer synchronously while the interviewer is thinking, then adopts the server pair without duplicates", async () => {
    const client = createDevMockClient();
    const initial = openingOnly(await client.getPracticeSession(SESSION_ID));
    vi.spyOn(client, "getPracticeSession").mockResolvedValue(initial);
    let resolveSend: ((value: SendPracticeMessageResponse) => void) | undefined;
    const send = vi.spyOn(client, "sendPracticeMessage").mockImplementation(
      () => new Promise((resolve) => { resolveSend = resolve; }),
    );

    renderPractice(client);
    await screen.findByText(initial.messages[0]!.content);

    const text = "这是需要立即显示的回答";
    fireEvent.change(screen.getByTestId("practice-input-textarea"), { target: { value: text } });
    fireEvent.click(screen.getByTestId("practice-input-send"));

    expect(send).toHaveBeenCalledTimes(1);
    const request = send.mock.calls[0]![1];
    expect(request.clientMessageId).toMatch(/^[0-9a-f-]{36}$/);
    expect(request.text).toBe(text);
    expect(screen.getByTestId("practice-input-textarea")).toHaveValue("");
    expect(screen.getByTestId("practice-input-textarea")).toBeDisabled();
    expect(screen.getByTestId("practice-input-send")).toBeDisabled();
    expect(screen.getByText(text)).toBeInTheDocument();
    expect(screen.getByTestId("practice-interviewer-thinking")).toHaveAttribute("role", "status");
    expect(screen.getByTestId("practice-interviewer-thinking")).toHaveAttribute("aria-live", "polite");
    expect(screen.queryByTestId("practice-message-retry")).not.toBeInTheDocument();
    expect(screen.getByTestId("practice-finish-cta")).toBeDisabled();

    const response = completedResponse(initial, request.clientMessageId, text);
    await act(async () => resolveSend?.(response));

    await waitFor(() => expect(screen.queryByTestId("practice-interviewer-thinking")).not.toBeInTheDocument());
    expect(screen.getAllByText(text)).toHaveLength(1);
    expect(screen.getAllByText(response.assistantMessage.content)).toHaveLength(1);
  });

  it("isolates all local message state when the mounted practice route changes from session A to B", async () => {
    vi.useFakeTimers();
    const client = createDevMockClient();
    const fixture = openingOnly(await client.getPracticeSession(SESSION_ID));
    const sessionA = sessionFor(fixture, SESSION_ID, "session A opening", TARGET_JOB_ID);
    const sessionB = sessionFor(fixture, SESSION_B_ID, "session B opening", TARGET_JOB_B_ID);
    const getSession = vi.spyOn(client, "getPracticeSession").mockImplementation(async (sessionId) => (
      sessionId === SESSION_B_ID ? sessionB : sessionA
    ));
    let resolveA: ((value: SendPracticeMessageResponse) => void) | undefined;
    let resolveB: ((value: SendPracticeMessageResponse) => void) | undefined;
    let signalA: AbortSignal | undefined;
    const send = vi.spyOn(client, "sendPracticeMessage").mockImplementation((sessionId, _body, options) => new Promise((resolve) => {
      if (sessionId === SESSION_ID) {
        signalA = options?.signal;
        resolveA = resolve;
      } else if (sessionId === SESSION_B_ID) {
        resolveB = resolve;
      }
    }));

    window.history.replaceState(null, "", practiceUrl(SESSION_ID, TARGET_JOB_ID));
    render(<App client={client} requestOptions={{ getMe: { headers: { Prefer: "example=authenticated" } } }} />);
    await act(async () => { await vi.advanceTimersByTimeAsync(0); });
    expect(screen.getByText("session A opening")).toBeInTheDocument();
    await act(async () => { await vi.advanceTimersByTimeAsync(2_000); });
    expect(screen.getByTestId("practice-topbar-timer")).toHaveTextContent("00:02");

    fireEvent.change(screen.getByTestId("practice-input-textarea"), { target: { value: "session A pending answer" } });
    fireEvent.click(screen.getByTestId("practice-input-send"));
    fireEvent.click(screen.getByTestId("practice-topbar-pause"));
    expect(signalA).toBeInstanceOf(AbortSignal);
    expect(screen.getByTestId("practice-interviewer-thinking")).toBeInTheDocument();
    expect(screen.getByTestId("practice-topbar-pause")).toHaveAttribute("aria-pressed", "true");

    window.history.pushState(null, "", practiceUrl(SESSION_B_ID, TARGET_JOB_B_ID));
    act(() => window.dispatchEvent(new PopStateEvent("popstate")));
    await act(async () => { await vi.advanceTimersByTimeAsync(0); });

    expect(signalA?.aborted).toBe(true);
    expect(getSession).toHaveBeenCalledWith(SESSION_B_ID);
    expect(screen.getByText("session B opening")).toBeInTheDocument();
    expect(screen.queryByText("session A pending answer")).not.toBeInTheDocument();
    expect(screen.queryByTestId("practice-interviewer-thinking")).not.toBeInTheDocument();
    expect(screen.queryByTestId("practice-error-state")).not.toBeInTheDocument();
    expect(screen.getByTestId("practice-input-textarea")).toHaveValue("");
    expect(screen.getByTestId("practice-input-textarea")).toBeEnabled();
    expect(screen.getByTestId("practice-topbar-pause")).toHaveAttribute("aria-pressed", "false");
    expect(screen.getByTestId("practice-topbar-timer")).toHaveTextContent("00:00");

    fireEvent.change(screen.getByTestId("practice-input-textarea"), { target: { value: "session B pending answer" } });
    fireEvent.click(screen.getByTestId("practice-input-send"));
    expect(screen.getByTestId("practice-interviewer-thinking")).toBeInTheDocument();
    await act(async () => {
      resolveA?.(completedResponse(sessionA, "late-a-id", "session A pending answer", "late A reply"));
    });
    expect(screen.getByTestId("practice-interviewer-thinking")).toBeInTheDocument();
    expect(screen.getByTestId("practice-input-textarea")).toBeDisabled();
    expect(screen.queryByText("late A reply")).not.toBeInTheDocument();

    const bBody = send.mock.calls.at(-1)![1];
    await act(async () => {
      resolveB?.(completedResponse(sessionB, bBody.clientMessageId, bBody.text, "session B reply"));
    });
    expect(screen.getByText("session B reply")).toBeInTheDocument();
  });

  it("drops session A's local message error, error owner, and draft when the mounted route changes to session B", async () => {
    const client = createDevMockClient();
    const fixture = openingOnly(await client.getPracticeSession(SESSION_ID));
    const sessionA = sessionFor(fixture, SESSION_ID, "error session A", TARGET_JOB_ID);
    const sessionB = sessionFor(fixture, SESSION_B_ID, "clean session B", TARGET_JOB_B_ID);
    vi.spyOn(client, "getPracticeSession").mockImplementation(async (sessionId) => (
      sessionId === SESSION_B_ID ? sessionB : sessionA
    ));
    vi.spyOn(client, "sendPracticeMessage").mockRejectedValue(new Error("session A local failure"));

    window.history.replaceState(null, "", practiceUrl(SESSION_ID, TARGET_JOB_ID));
    render(<App client={client} requestOptions={{ getMe: { headers: { Prefer: "example=authenticated" } } }} />);
    expect(await screen.findByText("error session A")).toBeInTheDocument();
    fireEvent.change(screen.getByTestId("practice-input-textarea"), { target: { value: "failed A answer" } });
    fireEvent.click(screen.getByTestId("practice-input-send"));
    expect(await screen.findByTestId("practice-error-state")).toBeInTheDocument();
    fireEvent.change(screen.getByTestId("practice-input-textarea"), { target: { value: "stale A draft" } });

    window.history.pushState(null, "", practiceUrl(SESSION_B_ID, TARGET_JOB_B_ID));
    act(() => window.dispatchEvent(new PopStateEvent("popstate")));
    expect(await screen.findByText("clean session B")).toBeInTheDocument();
    expect(screen.queryByTestId("practice-error-state")).not.toBeInTheDocument();
    expect(screen.queryByText("failed A answer")).not.toBeInTheDocument();
    expect(screen.getByTestId("practice-input-textarea")).toHaveValue("");
    expect(screen.getByTestId("practice-input-send")).toBeDisabled();
    fireEvent.change(screen.getByTestId("practice-input-textarea"), { target: { value: "fresh B draft" } });
    expect(screen.getByTestId("practice-input-send")).toBeEnabled();
  });

  it("cancels session A's pending poll when the mounted route changes to session B", async () => {
    vi.useFakeTimers();
    const client = createDevMockClient();
    const fixture = openingOnly(await client.getPracticeSession(SESSION_ID));
    const sessionA = sessionWithUser(
      sessionFor(fixture, SESSION_ID, "pending session A", TARGET_JOB_ID),
      "pending-a-message",
      "persisted pending A answer",
      "pending",
    );
    const sessionB = sessionFor(fixture, SESSION_B_ID, "non-pending session B", TARGET_JOB_B_ID);
    const getSession = vi.spyOn(client, "getPracticeSession").mockImplementation(async (sessionId) => (
      sessionId === SESSION_B_ID ? sessionB : sessionA
    ));

    window.history.replaceState(null, "", practiceUrl(SESSION_ID, TARGET_JOB_ID));
    render(<App client={client} requestOptions={{ getMe: { headers: { Prefer: "example=authenticated" } } }} />);
    await act(async () => { await vi.advanceTimersByTimeAsync(0); });
    expect(screen.getByTestId("practice-interviewer-thinking")).toBeInTheDocument();
    expect(getSession.mock.calls.filter(([sessionId]) => sessionId === SESSION_ID)).toHaveLength(1);

    window.history.pushState(null, "", practiceUrl(SESSION_B_ID, TARGET_JOB_B_ID));
    act(() => window.dispatchEvent(new PopStateEvent("popstate")));
    await act(async () => { await vi.advanceTimersByTimeAsync(0); });
    expect(screen.getByText("non-pending session B")).toBeInTheDocument();
    expect(screen.queryByTestId("practice-interviewer-thinking")).not.toBeInTheDocument();
    await act(async () => { await vi.advanceTimersByTimeAsync(10_000); });
    expect(getSession.mock.calls.filter(([sessionId]) => sessionId === SESSION_ID)).toHaveLength(1);
  });

  it.each(["resolve", "reject"] as const)("ignores a late session A completion %s after navigating the mounted practice route to session B", async (outcome) => {
    const client = createDevMockClient();
    const fixture = openingOnly(await client.getPracticeSession(SESSION_ID));
    const sessionA = sessionWithUser(
      sessionFor(fixture, SESSION_ID, "finish session A", TARGET_JOB_ID),
      "complete-a-message",
      "completed A answer",
      "complete",
    );
    const sessionB = sessionWithUser(
      sessionFor(fixture, SESSION_B_ID, "finish session B", TARGET_JOB_B_ID),
      "complete-b-message",
      "completed B answer",
      "complete",
    );
    vi.spyOn(client, "getPracticeSession").mockImplementation(async (sessionId) => (
      sessionId === SESSION_B_ID ? sessionB : sessionA
    ));
    let resolveA: ((value: ReportWithJob) => void) | undefined;
    let rejectA: ((cause: Error) => void) | undefined;
    const complete = vi.spyOn(client, "completePracticeSession").mockImplementation((sessionId) => new Promise((resolve, reject) => {
      if (sessionId === SESSION_ID) {
        resolveA = resolve;
        rejectA = reject;
      }
    }));

    window.history.replaceState(null, "", practiceUrl(SESSION_ID, TARGET_JOB_ID));
    render(<App client={client} requestOptions={{ getMe: { headers: { Prefer: "example=authenticated" } } }} />);
    expect(await screen.findByText("finish session A")).toBeInTheDocument();
    fireEvent.click(screen.getByTestId("practice-finish-cta"));
    expect(complete).toHaveBeenCalledWith(SESSION_ID, expect.anything(), expect.anything());

    window.history.pushState(null, "", practiceUrl(SESSION_B_ID, TARGET_JOB_B_ID));
    act(() => window.dispatchEvent(new PopStateEvent("popstate")));
    expect(await screen.findByText("finish session B")).toBeInTheDocument();
    expect(screen.getByTestId("practice-finish-cta")).toBeEnabled();

    await act(async () => {
      if (outcome === "resolve") resolveA?.(completionReport("report-a"));
      else rejectA?.(new Error("late session A completion failure"));
      await Promise.resolve();
    });
    expect(window.location.pathname).toBe("/practice");
    expect(window.location.search).toContain(`sessionId=${SESSION_B_ID}`);
    expect(screen.getByText("finish session B")).toBeInTheDocument();
    expect(screen.queryByTestId("practice-error-state")).not.toBeInTheDocument();

    fireEvent.click(screen.getByTestId("practice-finish-cta"));
    expect(complete).toHaveBeenLastCalledWith(SESSION_B_ID, expect.anything(), expect.anything());
  });

  it("aborts the POST exactly at 95,000 ms, reconciles the same ID with an independent GET, and ignores a late POST response", async () => {
    const client = createDevMockClient();
    const base = openingOnly(await client.getPracticeSession(SESSION_ID));
    let submission: { text: string; clientMessageId: string } | undefined;
    let postSignal: AbortSignal | undefined;
    let resolveLatePost: ((value: SendPracticeMessageResponse) => void) | undefined;
    const getSession = vi.spyOn(client, "getPracticeSession").mockImplementation(async (_sessionId, options) => {
      if (!submission) return base;
      expect(options?.signal).toBeInstanceOf(AbortSignal);
      expect(options?.signal).not.toBe(postSignal);
      return sessionWithUser(base, submission.clientMessageId, submission.text, "retryable_failed");
    });
    const send = vi.spyOn(client, "sendPracticeMessage").mockImplementation((_sessionId, body, options) => {
      submission = body;
      postSignal = options?.signal;
      return new Promise((resolve) => { resolveLatePost = resolve; });
    });

    const view = renderPractice(client);
    await screen.findByText(base.messages[0]!.content);
    vi.useFakeTimers();
    const text = "95 秒后按同一 ID 对账";
    fireEvent.change(screen.getByTestId("practice-input-textarea"), { target: { value: text } });
    fireEvent.click(screen.getByTestId("practice-input-send"));

    expect(send).toHaveBeenCalledTimes(1);
    expect(postSignal).toBeInstanceOf(AbortSignal);
    expect(postSignal?.aborted).toBe(false);
    const originalId = send.mock.calls[0]![1].clientMessageId;

    await act(async () => { await vi.advanceTimersByTimeAsync(94_999); });
    expect(postSignal?.aborted).toBe(false);
    expect(getSession).toHaveBeenCalledTimes(1);

    await act(async () => { await vi.advanceTimersByTimeAsync(1); });
    await flushPromises();
    expect(postSignal?.aborted).toBe(true);
    expect(getSession).toHaveBeenCalledTimes(2);
    expect(screen.getByTestId("practice-message-retry")).toBeInTheDocument();
    expect(screen.getAllByText(text)).toHaveLength(1);

    await act(async () => {
      resolveLatePost?.(completedResponse(base, originalId, text, "迟到旧 POST 回复"));
    });
    expect(screen.queryByText("迟到旧 POST 回复")).not.toBeInTheDocument();
    expect(screen.getByTestId("practice-message-retry")).toBeInTheDocument();
    view.unmount();
  });

  it("ignores an older timeout reconcile after a newer loader read adopts trusted server truth", async () => {
    const client = createDevMockClient();
    const base = openingOnly(await client.getPracticeSession(SESSION_ID));
    let submission: { text: string; clientMessageId: string } | undefined;
    let resolveOlderReconcile: ((session: PracticeSession) => void) | undefined;
    let readCount = 0;
    const getSession = vi.spyOn(client, "getPracticeSession").mockImplementation(() => {
      readCount += 1;
      if (readCount === 1) return Promise.resolve(base);
      if (readCount === 2) {
        return new Promise((resolve) => { resolveOlderReconcile = resolve; });
      }
      if (!submission) return Promise.resolve(base);
      return Promise.resolve(sessionWithUser(
        base,
        submission.clientMessageId,
        submission.text,
        "terminal_failed",
      ));
    });
    vi.spyOn(client, "sendPracticeMessage").mockImplementation((_sessionId, body) => {
      submission = body;
      return new Promise(() => undefined);
    });

    const view = renderPractice(client);
    await screen.findByText(base.messages[0]!.content);
    vi.useFakeTimers();
    fireEvent.change(screen.getByTestId("practice-input-textarea"), { target: { value: "newer server truth wins" } });
    fireEvent.click(screen.getByTestId("practice-input-send"));
    await act(async () => { await vi.advanceTimersByTimeAsync(95_000); });
    await flushPromises();
    expect(getSession).toHaveBeenCalledTimes(2);

    act(() => window.dispatchEvent(new Event("focus")));
    await flushPromises();
    expect(getSession).toHaveBeenCalledTimes(3);
    expect(screen.getByTestId("practice-terminal-recovery")).toBeInTheDocument();

    await act(async () => { resolveOlderReconcile?.(base); });
    await flushPromises();
    expect(screen.getByTestId("practice-terminal-recovery")).toBeInTheDocument();
    expect(screen.queryByTestId("practice-message-retry")).not.toBeInTheDocument();
    expect(screen.getByTestId("practice-input-send")).toBeDisabled();
    view.unmount();
  });

  it("lets a later-started timeout reconcile win when an older loader refresh resolves first", async () => {
    const client = createDevMockClient();
    const base = openingOnly(await client.getPracticeSession(SESSION_ID));
    let submission: { text: string; clientMessageId: string } | undefined;
    let resolveOlderRefresh: ((session: PracticeSession) => void) | undefined;
    let resolveNewerReconcile: ((session: PracticeSession) => void) | undefined;
    let readCount = 0;
    const getSession = vi.spyOn(client, "getPracticeSession").mockImplementation(() => {
      readCount += 1;
      if (readCount === 1) return Promise.resolve(base);
      if (readCount === 2) return new Promise((resolve) => { resolveOlderRefresh = resolve; });
      return new Promise((resolve) => { resolveNewerReconcile = resolve; });
    });
    vi.spyOn(client, "sendPracticeMessage").mockImplementation((_sessionId, body) => {
      submission = body;
      return new Promise(() => undefined);
    });

    const view = renderPractice(client);
    await screen.findByText(base.messages[0]!.content);
    vi.useFakeTimers();
    fireEvent.change(screen.getByTestId("practice-input-textarea"), { target: { value: "later timeout reconcile wins" } });
    fireEvent.click(screen.getByTestId("practice-input-send"));
    await act(async () => { await vi.advanceTimersByTimeAsync(94_999); });
    act(() => window.dispatchEvent(new Event("focus")));
    await flushPromises();
    expect(getSession).toHaveBeenCalledTimes(2);
    await act(async () => { await vi.advanceTimersByTimeAsync(1); });
    await flushPromises();
    expect(getSession).toHaveBeenCalledTimes(3);
    if (!submission) throw new Error("expected submitted message");

    await act(async () => {
      resolveOlderRefresh?.(sessionWithUser(
        base,
        submission!.clientMessageId,
        submission!.text,
        "terminal_failed",
      ));
    });
    expect(screen.queryByTestId("practice-terminal-recovery")).not.toBeInTheDocument();

    await act(async () => {
      resolveNewerReconcile?.(completedResponse(
        base,
        submission!.clientMessageId,
        submission!.text,
        "newer timeout reconcile reply",
      ).session);
    });
    expect(screen.getByText("newer timeout reconcile reply")).toBeInTheDocument();
    expect(screen.queryByTestId("practice-terminal-recovery")).not.toBeInTheDocument();
    view.unmount();
  });

  it("lets a later-started loader refresh win when the older timeout reconcile resolves first", async () => {
    const client = createDevMockClient();
    const base = openingOnly(await client.getPracticeSession(SESSION_ID));
    let submission: { text: string; clientMessageId: string } | undefined;
    let resolveOlderReconcile: ((session: PracticeSession) => void) | undefined;
    let resolveNewerRefresh: ((session: PracticeSession) => void) | undefined;
    let readCount = 0;
    const getSession = vi.spyOn(client, "getPracticeSession").mockImplementation(() => {
      readCount += 1;
      if (readCount === 1) return Promise.resolve(base);
      if (readCount === 2) return new Promise((resolve) => { resolveOlderReconcile = resolve; });
      return new Promise((resolve) => { resolveNewerRefresh = resolve; });
    });
    vi.spyOn(client, "sendPracticeMessage").mockImplementation((_sessionId, body) => {
      submission = body;
      return new Promise(() => undefined);
    });

    const view = renderPractice(client);
    await screen.findByText(base.messages[0]!.content);
    vi.useFakeTimers();
    fireEvent.change(screen.getByTestId("practice-input-textarea"), { target: { value: "later loader refresh wins" } });
    fireEvent.click(screen.getByTestId("practice-input-send"));
    await act(async () => { await vi.advanceTimersByTimeAsync(95_000); });
    await flushPromises();
    expect(getSession).toHaveBeenCalledTimes(2);
    act(() => window.dispatchEvent(new Event("focus")));
    await flushPromises();
    expect(getSession).toHaveBeenCalledTimes(3);
    if (!submission) throw new Error("expected submitted message");

    await act(async () => {
      resolveOlderReconcile?.(completedResponse(
        base,
        submission!.clientMessageId,
        submission!.text,
        "older reconcile must stay hidden",
      ).session);
    });
    expect(screen.queryByText("older reconcile must stay hidden")).not.toBeInTheDocument();

    await act(async () => {
      resolveNewerRefresh?.(sessionWithUser(
        base,
        submission!.clientMessageId,
        submission!.text,
        "terminal_failed",
      ));
    });
    expect(screen.getByTestId("practice-terminal-recovery")).toBeInTheDocument();
    expect(screen.queryByText("older reconcile must stay hidden")).not.toBeInTheDocument();
    view.unmount();
  });

  it("preserves the timeout failure classification when a newer loader refresh also finishes without the clientMessageId", async () => {
    const client = createDevMockClient();
    const base = openingOnly(await client.getPracticeSession(SESSION_ID));
    let resolveOlderReconcile: ((session: PracticeSession) => void) | undefined;
    let resolveNewerRefresh: ((session: PracticeSession) => void) | undefined;
    let readCount = 0;
    const getSession = vi.spyOn(client, "getPracticeSession").mockImplementation(() => {
      readCount += 1;
      if (readCount === 1) return Promise.resolve(base);
      if (readCount === 2) return new Promise((resolve) => { resolveOlderReconcile = resolve; });
      return new Promise((resolve) => { resolveNewerRefresh = resolve; });
    });
    const send = vi.spyOn(client, "sendPracticeMessage").mockImplementation(
      () => new Promise(() => undefined),
    );

    const view = renderPractice(client);
    await screen.findByText(base.messages[0]!.content);
    vi.useFakeTimers();
    fireEvent.change(screen.getByTestId("practice-input-textarea"), { target: { value: "missing after newer read" } });
    fireEvent.click(screen.getByTestId("practice-input-send"));
    const originalBody = send.mock.calls[0]![1];
    await act(async () => { await vi.advanceTimersByTimeAsync(95_000); });
    await flushPromises();
    expect(getSession).toHaveBeenCalledTimes(2);

    act(() => window.dispatchEvent(new Event("focus")));
    await flushPromises();
    expect(getSession).toHaveBeenCalledTimes(3);
    await act(async () => { resolveOlderReconcile?.(base); });

    expect(screen.getByTestId("practice-message-retry")).toBeInTheDocument();
    expect(screen.queryByTestId("practice-interviewer-thinking")).not.toBeInTheDocument();
    expect(screen.getAllByText(originalBody.text)).toHaveLength(1);

    await act(async () => { resolveNewerRefresh?.(base); });
    expect(screen.getByTestId("practice-message-retry")).toBeInTheDocument();
    expect(screen.queryByTestId("practice-interviewer-thinking")).not.toBeInTheDocument();
    expect(screen.getByTestId("practice-input-send")).toBeDisabled();
    expect(screen.getByTestId("practice-finish-cta")).toBeDisabled();
    expect(send).toHaveBeenCalledTimes(1);
    view.unmount();
  });

  it.each([
    ["complete", "complete"],
    ["pending", "pending"],
    ["terminal", "terminal_failed"],
  ] as const)("adopts authoritative %s server truth after the 95-second timeout", async (_label, replyStatus) => {
    const client = createDevMockClient();
    const base = openingOnly(await client.getPracticeSession(SESSION_ID));
    let submission: { text: string; clientMessageId: string } | undefined;
    vi.spyOn(client, "getPracticeSession").mockImplementation(async () => {
      if (!submission) return base;
      return replyStatus === "complete"
        ? completedResponse(base, submission.clientMessageId, submission.text).session
        : sessionWithUser(base, submission.clientMessageId, submission.text, replyStatus);
    });
    vi.spyOn(client, "sendPracticeMessage").mockImplementation((_sessionId, body) => {
      submission = body;
      return new Promise(() => undefined);
    });

    const view = renderPractice(client);
    await screen.findByText(base.messages[0]!.content);
    vi.useFakeTimers();
    fireEvent.change(screen.getByTestId("practice-input-textarea"), { target: { value: `server ${replyStatus}` } });
    fireEvent.click(screen.getByTestId("practice-input-send"));
    await act(async () => { await vi.advanceTimersByTimeAsync(95_000); });
    await flushPromises();

    if (replyStatus === "complete") {
      expect(screen.getByText("服务端唯一回复")).toBeInTheDocument();
      expect(screen.queryByTestId("practice-interviewer-thinking")).not.toBeInTheDocument();
    } else if (replyStatus === "pending") {
      expect(screen.getByTestId("practice-interviewer-thinking")).toBeInTheDocument();
      expect(screen.queryByTestId("practice-message-retry")).not.toBeInTheDocument();
    } else {
      expect(screen.getByTestId("practice-terminal-recovery")).toBeInTheDocument();
      expect(screen.queryByTestId("practice-message-retry")).not.toBeInTheDocument();
    }
    view.unmount();
  });

  it.each(["missing-id", "read-failure"] as const)("keeps the original row/ID fail-locked when timeout reconciliation ends in %s", async (mode) => {
    const client = createDevMockClient();
    const base = openingOnly(await client.getPracticeSession(SESSION_ID));
    const getSession = vi.spyOn(client, "getPracticeSession")
      .mockResolvedValueOnce(base);
    if (mode === "missing-id") {
      getSession.mockResolvedValue(base);
    } else {
      getSession.mockRejectedValue(new ApiClientError("transport", null, null, new TypeError("offline")));
    }
    const send = vi.spyOn(client, "sendPracticeMessage").mockImplementation(
      () => new Promise(() => undefined),
    );

    const view = renderPractice(client);
    await screen.findByText(base.messages[0]!.content);
    vi.useFakeTimers();
    const text = `uncertain ${mode}`;
    fireEvent.change(screen.getByTestId("practice-input-textarea"), { target: { value: text } });
    fireEvent.click(screen.getByTestId("practice-input-send"));
    const originalBody = send.mock.calls[0]![1];

    await act(async () => { await vi.advanceTimersByTimeAsync(95_000); });
    await flushPromises();

    expect(screen.getAllByText(text)).toHaveLength(1);
    expect(screen.getByTestId("practice-message-retry")).toBeInTheDocument();
    expect(screen.getByTestId("practice-finish-cta")).toBeDisabled();
    fireEvent.change(screen.getByTestId("practice-input-textarea"), { target: { value: "new ID must stay locked" } });
    expect(screen.getByTestId("practice-input-send")).toBeDisabled();
    fireEvent.click(screen.getByTestId("practice-input-send"));
    expect(send).toHaveBeenCalledTimes(1);
    expect(send.mock.calls[0]![1]).toEqual(originalBody);
    view.unmount();
  });

  it("rehydrates a server pending row, shows thinking, polls server truth, and never resends it", async () => {
    const client = createDevMockClient();
    const base = openingOnly(await client.getPracticeSession(SESSION_ID));
    const pendingText = "刷新后仍在等待的回答";
    const pending = sessionWithUser(base, SERVER_MESSAGE_ID, pendingText, "pending");
    const complete = completedResponse(base, SERVER_MESSAGE_ID, pendingText).session;
    const getSession = vi.spyOn(client, "getPracticeSession")
      .mockResolvedValueOnce(pending)
      .mockResolvedValue(complete);
    const send = vi.spyOn(client, "sendPracticeMessage");

    renderPractice(client);

    expect(await screen.findByText(pendingText)).toBeInTheDocument();
    expect(screen.getByTestId("practice-interviewer-thinking")).toBeInTheDocument();
    expect(screen.getByTestId("practice-input-textarea")).toBeDisabled();
    expect(screen.getByTestId("practice-finish-cta")).toBeDisabled();
    await waitFor(() => expect(getSession.mock.calls.length).toBeGreaterThanOrEqual(2), { timeout: 2500 });
    expect(await screen.findByText("服务端唯一回复")).toBeInTheDocument();
    expect(send).not.toHaveBeenCalled();
  });

  it("keeps a reloaded pending row and all mutations locked when its refresh read fails", async () => {
    const client = createDevMockClient();
    const base = openingOnly(await client.getPracticeSession(SESSION_ID));
    const pendingText = "刷新失败时仍须保留的 pending 回答";
    const pending = sessionWithUser(base, SERVER_MESSAGE_ID, pendingText, "pending");
    const getSession = vi.spyOn(client, "getPracticeSession")
      .mockResolvedValueOnce(pending)
      .mockRejectedValueOnce(new ApiClientError("transport", null, null, new TypeError("offline")));
    const send = vi.spyOn(client, "sendPracticeMessage");

    renderPractice(client);
    expect(await screen.findByText(pendingText)).toBeInTheDocument();
    act(() => window.dispatchEvent(new Event("focus")));

    expect(await screen.findByText("The interview session is temporarily unavailable. Please try again.")).toBeInTheDocument();
    expect(getSession).toHaveBeenCalledTimes(2);
    expect(screen.getByText(pendingText)).toBeInTheDocument();
    expect(screen.getByTestId("practice-interviewer-thinking")).toBeInTheDocument();
    expect(screen.getByTestId("practice-input-textarea")).toBeDisabled();
    expect(screen.getByTestId("practice-input-send")).toBeDisabled();
    expect(screen.getByTestId("practice-finish-cta")).toBeDisabled();
    expect(send).not.toHaveBeenCalled();
  });

  it("keeps bounded-backoff pending reconciliation alive until server truth settles without resending", async () => {
    const client = createDevMockClient();
    const base = openingOnly(await client.getPracticeSession(SESSION_ID));
    const pending = sessionWithUser(base, SERVER_MESSAGE_ID, "持续 pending 的回答", "pending");
    const getSession = vi.spyOn(client, "getPracticeSession").mockResolvedValue(pending);
    const send = vi.spyOn(client, "sendPracticeMessage");
    vi.useFakeTimers();

    renderPractice(client);
    await act(async () => {
      await Promise.resolve();
      await Promise.resolve();
    });
    expect(screen.getByText("持续 pending 的回答")).toBeInTheDocument();

    await act(async () => { await vi.advanceTimersByTimeAsync(750); });
    await act(async () => { await vi.advanceTimersByTimeAsync(1_500); });
    await act(async () => { await vi.advanceTimersByTimeAsync(3_000); });
    await act(async () => { await vi.advanceTimersByTimeAsync(60_000); });

    expect(getSession.mock.calls.length).toBeGreaterThan(4);
    expect(send).not.toHaveBeenCalled();
    expect(screen.getByTestId("practice-interviewer-thinking")).toBeInTheDocument();
  });

  it("rehydrates one retryable row and retries with the server text/id while preserving the next draft", async () => {
    const client = createDevMockClient();
    const base = openingOnly(await client.getPracticeSession(SESSION_ID));
    const failedText = "服务端保存的失败回答";
    const failed = sessionWithUser(base, SERVER_MESSAGE_ID, failedText, "retryable_failed");
    vi.spyOn(client, "getPracticeSession").mockResolvedValue(failed);
    let resolveRetry: ((value: SendPracticeMessageResponse) => void) | undefined;
    const send = vi.spyOn(client, "sendPracticeMessage").mockImplementation(
      () => new Promise((resolve) => { resolveRetry = resolve; }),
    );

    renderPractice(client);
    expect(await screen.findByText(failedText)).toBeInTheDocument();
    expect(screen.getAllByTestId("practice-message-retry")).toHaveLength(1);
    expect(screen.getByTestId("practice-finish-cta")).toBeDisabled();

    const draft = "下一条尚未提交的草稿";
    fireEvent.change(screen.getByTestId("practice-input-textarea"), { target: { value: draft } });
    fireEvent.click(screen.getByTestId("practice-message-retry"));

    expect(send).toHaveBeenCalledWith(
      SESSION_ID,
      { clientMessageId: SERVER_MESSAGE_ID, text: failedText },
      expect.objectContaining({ signal: expect.any(AbortSignal) }),
    );
    expect(screen.getByTestId("practice-input-textarea")).toHaveValue(draft);
    expect(screen.getByTestId("practice-input-textarea")).toBeDisabled();
    expect(screen.queryByTestId("practice-message-retry")).not.toBeInTheDocument();
    expect(screen.getByTestId("practice-interviewer-thinking")).toBeInTheDocument();

    await act(async () => resolveRetry?.(completedResponse(base, SERVER_MESSAGE_ID, failedText)));
    await waitFor(() => expect(screen.queryByTestId("practice-interviewer-thinking")).not.toBeInTheDocument());
    expect(screen.getByTestId("practice-input-textarea")).toHaveValue(draft);
    expect(screen.getAllByText(failedText)).toHaveLength(1);
    expect(screen.getAllByText("服务端唯一回复")).toHaveLength(1);
  });

  it("keeps one row-local retry and an editable next draft when online refresh fails after retryable server truth", async () => {
    const client = createDevMockClient();
    const base = openingOnly(await client.getPracticeSession(SESSION_ID));
    const failedText = "服务端已保存的可重试回答";
    const failed = sessionWithUser(base, SERVER_MESSAGE_ID, failedText, "retryable_failed");
    let rejectRefresh: ((cause: Error) => void) | undefined;
    const getSession = vi.spyOn(client, "getPracticeSession")
      .mockResolvedValueOnce(failed)
      .mockImplementationOnce(() => new Promise((_resolve, reject) => { rejectRefresh = reject; }));

    renderPractice(client);
    expect(await screen.findByText(failedText)).toBeInTheDocument();
    expect(screen.getAllByTestId("practice-message-retry")).toHaveLength(1);
    expect(screen.getByTestId("practice-input-textarea")).toBeEnabled();

    act(() => window.dispatchEvent(new Event("online")));
    await waitFor(() => expect(getSession).toHaveBeenCalledTimes(2));
    await act(async () => {
      rejectRefresh?.(new ApiClientError("transport", null, null, new TypeError("offline refresh")));
    });

    expect(screen.getAllByTestId("practice-message-retry")).toHaveLength(1);
    expect(screen.queryByTestId("practice-error-state")).not.toBeInTheDocument();
    expect(screen.queryByTestId("practice-error-state-retry")).not.toBeInTheDocument();
    expect(screen.queryByText(/offline refresh|transport/u)).not.toBeInTheDocument();
    expect(screen.getByTestId("practice-input-textarea")).toBeEnabled();
    expect(screen.getByTestId("practice-input-send")).toBeDisabled();
    expect(screen.getByTestId("practice-finish-cta")).toBeDisabled();
    expect(screen.getAllByText(failedText)).toHaveLength(1);
  });

  it("maps typed HTTP retryable=true to the same row/id retry without duplicate or technical leakage", async () => {
    const client = createDevMockClient();
    const base = openingOnly(await client.getPracticeSession(SESSION_ID));
    vi.spyOn(client, "getPracticeSession").mockResolvedValue(base);
    const send = vi.spyOn(client, "sendPracticeMessage")
      .mockRejectedValueOnce(new ApiClientError("http", 503, {
        error: {
          code: "AI_PROVIDER_TIMEOUT",
          message: "provider timeout with secret diagnostic",
          requestId: "req-typed-retryable",
          retryable: true,
        },
      }))
      .mockImplementationOnce(async (_sessionId, body) => (
        completedResponse(base, body.clientMessageId, body.text, "typed retry completed")
      ));

    renderPractice(client);
    await screen.findByText(base.messages[0]!.content);
    const text = "需要使用同一消息标识重试的回答";
    fireEvent.change(screen.getByTestId("practice-input-textarea"), { target: { value: text } });
    fireEvent.click(screen.getByTestId("practice-input-send"));

    const retry = await screen.findByTestId("practice-message-retry");
    const originalBody = send.mock.calls[0]![1];
    expect(screen.queryByTestId("practice-error-state")).not.toBeInTheDocument();
    expect(screen.queryByText(/AI_PROVIDER_TIMEOUT|provider timeout|secret diagnostic/u)).not.toBeInTheDocument();
    expect(screen.getAllByText(text)).toHaveLength(1);

    const nextDraft = "下一条尚未提交的草稿";
    fireEvent.change(screen.getByTestId("practice-input-textarea"), { target: { value: nextDraft } });
    fireEvent.click(retry);

    expect(await screen.findByText("typed retry completed")).toBeInTheDocument();
    expect(send).toHaveBeenCalledTimes(2);
    expect(send.mock.calls[1]![1]).toEqual(originalBody);
    expect(screen.getAllByText(text)).toHaveLength(1);
    expect(screen.queryByTestId("practice-message-retry")).not.toBeInTheDocument();
    expect(screen.getByTestId("practice-input-textarea")).toHaveValue(nextDraft);
    expect(screen.getByTestId("practice-input-textarea")).toBeEnabled();
  });

  it.each([
    ["zh", "本次回复未能完成。", "请返回当前面试规划，准备好后重新开始一场面试。", "返回当前面试规划"],
    ["en", "This reply could not be completed.", "Return to this interview plan, then start a new session when you are ready.", "Return to this interview plan"],
  ] as const)("renders the %s terminal state with one safe exact current-plan CTA", async (lang, title, description, ctaLabel) => {
    const practiceSource = readFileSync(resolve(__dirname, "PracticeScreen.tsx"), "utf8");
    expect(practiceSource).toContain('navigate({ name: "workspace", params: { targetJobId } })');
    expect(practiceSource).not.toContain('navigate({ name: "parse", params: { targetJobId } })');

    localStorage.setItem("ei-lang", lang);
    const client = createDevMockClient();
    const base = openingOnly(await client.getPracticeSession(SESSION_ID));
    vi.spyOn(client, "getPracticeSession").mockResolvedValue(
      sessionWithUser(base, SERVER_MESSAGE_ID, "raw terminal_failed technical payload", "terminal_failed"),
    );
    const send = vi.spyOn(client, "sendPracticeMessage");

    window.history.replaceState(
      null,
      "",
      `/practice?planId=${PLAN_ID}&sessionId=${SESSION_ID}&targetJobId=route-target-must-not-win`,
    );
    render(<App client={client} requestOptions={{ getMe: { headers: { Prefer: "example=authenticated" } } }} />);
    const recovery = await screen.findByTestId("practice-terminal-recovery");
    expect(recovery).toHaveAttribute("role", "alert");
    expect(within(recovery).getByText(title)).toBeInTheDocument();
    expect(within(recovery).getByText(description)).toBeInTheDocument();
    expect(within(recovery).getAllByRole("button")).toHaveLength(1);
    const cta = within(recovery).getByRole("button", { name: ctaLabel });
    expect(cta).toHaveStyle({
      height: "30px",
      padding: "0 12px",
      fontSize: "13px",
      opacity: "1",
      transition: "transform .08s ease, opacity .15s",
    });
    fireEvent.mouseDown(cta);
    expect(cta).toHaveStyle({ transform: "translateY(0.5px)" });
    fireEvent.mouseUp(cta);
    expect(cta.style.transform).toBe("");
    fireEvent.mouseDown(cta);
    fireEvent.mouseLeave(cta);
    expect(cta.style.transform).toBe("");
    expect(screen.queryByTestId("practice-message-retry")).not.toBeInTheDocument();
    expect(screen.queryByTestId("practice-interviewer-thinking")).not.toBeInTheDocument();
    expect(screen.getByTestId("practice-input-textarea")).toBeDisabled();
    expect(screen.getByTestId("practice-input-send")).toBeDisabled();
    expect(screen.getByTestId("practice-finish-cta")).toBeDisabled();
    expect(screen.queryByText(/HTTP 409|IDEMPOTENCY|provider stack|secret backend diagnostic/u)).not.toBeInTheDocument();
    expect(send).not.toHaveBeenCalled();

    fireEvent.click(cta);
    await waitFor(() => expect(window.location.pathname).toBe("/workspace"));
    expect(window.location.search).toBe(`?targetJobId=${TARGET_JOB_ID}`);
    expect([...new URL(window.location.href).searchParams.keys()]).toEqual(["targetJobId"]);
    expect(window.location.pathname).not.toBe("/parse");
    expect(window.location.search).not.toBe("");
    expect(window.location.href).not.toContain("planId");
  });

  it("keeps trusted terminal recovery as the only prompt when a later loader refresh fails", async () => {
    localStorage.setItem("ei-lang", "en");
    const client = createDevMockClient();
    const base = openingOnly(await client.getPracticeSession(SESSION_ID));
    const terminal = sessionWithUser(
      base,
      SERVER_MESSAGE_ID,
      "trusted terminal row",
      "terminal_failed",
    );
    let rejectRefresh: ((cause: Error) => void) | undefined;
    const getSession = vi.spyOn(client, "getPracticeSession")
      .mockResolvedValueOnce(terminal)
      .mockImplementationOnce(() => new Promise((_resolve, reject) => { rejectRefresh = reject; }));

    renderPractice(client);
    const recovery = await screen.findByTestId("practice-terminal-recovery");
    expect(within(recovery).getAllByRole("button")).toHaveLength(1);
    act(() => window.dispatchEvent(new Event("focus")));
    await waitFor(() => expect(getSession).toHaveBeenCalledTimes(2));
    await act(async () => {
      rejectRefresh?.(new ApiClientError("transport", null, null, new TypeError("offline")));
    });

    expect(screen.getAllByTestId("practice-terminal-recovery")).toHaveLength(1);
    expect(screen.queryByTestId("practice-error-state")).not.toBeInTheDocument();
    expect(screen.queryByTestId("practice-error-state-retry")).not.toBeInTheDocument();
    expect(screen.queryByText(/offline|temporarily unavailable/u)).not.toBeInTheDocument();
    expect(screen.getByTestId("practice-input-send")).toBeDisabled();
    expect(screen.getByTestId("practice-finish-cta")).toBeDisabled();
  });

  it("shows only trusted terminal recovery when reconciliation confirms server terminal_failed", async () => {
    localStorage.setItem("ei-lang", "en");
    const client = createDevMockClient();
    const base = openingOnly(await client.getPracticeSession(SESSION_ID));
    let submission: { text: string; clientMessageId: string } | undefined;
    vi.spyOn(client, "getPracticeSession").mockImplementation(async () => {
      if (!submission) return base;
      return sessionWithUser(base, submission.clientMessageId, submission.text, "terminal_failed");
    });
    vi.spyOn(client, "sendPracticeMessage").mockImplementation(async (_sessionId, body) => {
      submission = body;
      throw new ApiClientError("http", 409, {
        error: {
          code: "IDEMPOTENCY_KEY_MISMATCH",
          message: "technical mismatch detail",
          requestId: "req-trusted-terminal",
          retryable: false,
        },
      });
    });

    renderPractice(client);
    await screen.findByText(base.messages[0]!.content);
    fireEvent.change(screen.getByTestId("practice-input-textarea"), { target: { value: "trusted terminal answer" } });
    fireEvent.click(screen.getByTestId("practice-input-send"));

    expect(await screen.findByTestId("practice-terminal-recovery")).toBeInTheDocument();
    expect(screen.queryByTestId("practice-error-state")).not.toBeInTheDocument();
    expect(screen.queryByText("This message was not accepted. Check it before sending again.")).not.toBeInTheDocument();
    expect(screen.queryByText(/technical mismatch detail|IDEMPOTENCY_KEY_MISMATCH/u)).not.toBeInTheDocument();
    expect(screen.queryByTestId("practice-message-retry")).not.toBeInTheDocument();
  });

  it("reconciles a transport failure with getPracticeSession and adopts completed server truth before offering retry", async () => {
    const client = createDevMockClient();
    const base = openingOnly(await client.getPracticeSession(SESSION_ID));
    let submission: { text: string; clientMessageId: string } | undefined;
    const getSession = vi.spyOn(client, "getPracticeSession").mockImplementation(async () => {
      if (!submission) return base;
      return completedResponse(base, submission.clientMessageId, submission.text).session;
    });
    vi.spyOn(client, "sendPracticeMessage").mockImplementation(async (_sessionId, body) => {
      submission = body;
      throw new ApiClientError("transport", null, null, new TypeError("connection reset"));
    });

    renderPractice(client);
    await screen.findByText(base.messages[0]!.content);
    fireEvent.change(screen.getByTestId("practice-input-textarea"), { target: { value: "已由服务端完成的回答" } });
    fireEvent.click(screen.getByTestId("practice-input-send"));

    expect(await screen.findByText("服务端唯一回复")).toBeInTheDocument();
    expect(getSession.mock.calls.length).toBeGreaterThanOrEqual(2);
    expect(screen.queryByTestId("practice-message-retry")).not.toBeInTheDocument();
    expect(screen.getAllByText("已由服务端完成的回答")).toHaveLength(1);
  });

  it("keeps the same optimistic row retryable when transport reconciliation succeeds without finding its clientMessageId", async () => {
    const client = createDevMockClient();
    const base = openingOnly(await client.getPracticeSession(SESSION_ID));
    const getSession = vi.spyOn(client, "getPracticeSession").mockResolvedValue(base);
    const transport = new ApiClientError("transport", null, null, new TypeError("offline"));
    const send = vi.spyOn(client, "sendPracticeMessage").mockRejectedValue(transport);

    renderPractice(client);
    await screen.findByText(base.messages[0]!.content);
    const text = "对账成功但尚未入库的回答";
    fireEvent.change(screen.getByTestId("practice-input-textarea"), { target: { value: text } });
    fireEvent.click(screen.getByTestId("practice-input-send"));

    const retry = await screen.findByTestId("practice-message-retry");
    expect(getSession.mock.calls.length).toBeGreaterThanOrEqual(2);
    expect(screen.getAllByText(text)).toHaveLength(1);
    expect(screen.getByTestId("practice-input-textarea")).toBeEnabled();
    expect(screen.getByTestId("practice-input-send")).toBeDisabled();
    fireEvent.click(retry);

    await waitFor(() => expect(send).toHaveBeenCalledTimes(2));
    expect(send.mock.calls[1]![1]).toEqual(send.mock.calls[0]![1]);
  });

  it.each([
    [
      "HTTP mismatch",
      new ApiClientError("http", 409, {
        error: {
          code: "IDEMPOTENCY_KEY_MISMATCH",
          message: "clientMessageId was reused with different text",
          requestId: "req-message-mismatch",
          retryable: false,
          details: { field: "clientMessageId" },
        },
      }),
      "这条消息未被接受，请检查内容后重新发送。",
    ],
    ["abort", new ApiClientError("abort", null, null), "请求已中断，请确认当前面试状态后再继续。"],
    ["unknown", new Error("HTTP 503 provider stack: secret backend diagnostic"), "这条消息未能发送，请稍后再试。"],
  ])("localizes non-retryable %s feedback without exposing technical details or rendering a row retry", async (_label, failure, feedback) => {
    localStorage.setItem("ei-lang", "zh");
    const client = createDevMockClient();
    const base = openingOnly(await client.getPracticeSession(SESSION_ID));
    vi.spyOn(client, "getPracticeSession").mockResolvedValue(base);
    const send = vi.spyOn(client, "sendPracticeMessage").mockRejectedValue(failure);

    renderPractice(client);
    await screen.findByText(base.messages[0]!.content);
    const text = "必须恢复到输入框的原提交文本";
    fireEvent.change(screen.getByTestId("practice-input-textarea"), { target: { value: text } });
    fireEvent.click(screen.getByTestId("practice-input-send"));

    expect(await screen.findByText(feedback)).toBeInTheDocument();
    expect(screen.queryByText(/HTTP 409|IDEMPOTENCY_KEY_MISMATCH|clientMessageId was reused|Request aborted|HTTP 503|secret backend diagnostic/u)).not.toBeInTheDocument();
    expect(send).toHaveBeenCalledTimes(1);
    expect(screen.queryByTestId("practice-message-retry")).not.toBeInTheDocument();
    expect(screen.queryByTestId("practice-interviewer-thinking")).not.toBeInTheDocument();
    expect(screen.getAllByText(text)).toHaveLength(1);
    expect(screen.getByTestId("practice-input-textarea")).toHaveValue("");
    expect(screen.getByTestId("practice-input-textarea")).toBeEnabled();
    expect(screen.getByTestId("practice-input-send")).toBeDisabled();
    expect(screen.getByTestId("practice-error-state-retry")).toBeInTheDocument();
    expect(screen.queryByTestId("practice-terminal-recovery")).not.toBeInTheDocument();
  });

  it.each([
    ["terminal", "terminal_failed"],
    ["complete", "complete"],
  ] as const)("clears the local ErrorState when a safe loader refresh finds authoritative %s truth", async (_label, replyStatus) => {
    localStorage.setItem("ei-lang", "en");
    const client = createDevMockClient();
    const base = openingOnly(await client.getPracticeSession(SESSION_ID));
    let submission: { text: string; clientMessageId: string } | undefined;
    let readsAfterFailure = 0;
    const getSession = vi.spyOn(client, "getPracticeSession").mockImplementation(async () => {
      if (!submission) return base;
      readsAfterFailure += 1;
      if (readsAfterFailure === 1) return base;
      return replyStatus === "complete"
        ? completedResponse(base, submission.clientMessageId, submission.text, "authoritative complete reply").session
        : sessionWithUser(base, submission.clientMessageId, submission.text, "terminal_failed");
    });
    const send = vi.spyOn(client, "sendPracticeMessage").mockImplementation(async (_sessionId, body) => {
      submission = body;
      throw new ApiClientError("http", 409, {
        error: {
          code: "IDEMPOTENCY_KEY_MISMATCH",
          message: "technical mismatch detail",
          requestId: "req-safe-refresh",
          retryable: false,
        },
      });
    });

    renderPractice(client);
    await screen.findByText(base.messages[0]!.content);
    fireEvent.change(screen.getByTestId("practice-input-textarea"), { target: { value: "locally uncertain answer" } });
    fireEvent.click(screen.getByTestId("practice-input-send"));

    expect(await screen.findByText("This message was not accepted. Check it before sending again.")).toBeInTheDocument();
    expect(screen.queryByTestId("practice-message-retry")).not.toBeInTheDocument();
    expect(screen.getByTestId("practice-input-send")).toBeDisabled();
    fireEvent.click(screen.getByTestId("practice-error-state-retry"));

    await waitFor(() => expect(getSession).toHaveBeenCalledTimes(3));
    await waitFor(() => expect(screen.queryByTestId("practice-error-state")).not.toBeInTheDocument());
    expect(send).toHaveBeenCalledTimes(1);
    expect(screen.queryByText("technical mismatch detail")).not.toBeInTheDocument();
    if (replyStatus === "terminal_failed") {
      expect(screen.getAllByTestId("practice-terminal-recovery")).toHaveLength(1);
      expect(screen.queryByTestId("practice-message-retry")).not.toBeInTheDocument();
    } else {
      expect(screen.getByText("authoritative complete reply")).toBeInTheDocument();
      expect(screen.queryByTestId("practice-terminal-recovery")).not.toBeInTheDocument();
    }
  });

  it("keeps a missing-ID local terminal failure locked while its safe recovery action only refreshes the loader", async () => {
    localStorage.setItem("ei-lang", "en");
    const client = createDevMockClient();
    const base = openingOnly(await client.getPracticeSession(SESSION_ID));
    const getSession = vi.spyOn(client, "getPracticeSession").mockResolvedValue(base);
    const send = vi.spyOn(client, "sendPracticeMessage").mockRejectedValue(new Error("uncertain provider failure"));

    renderPractice(client);
    await screen.findByText(base.messages[0]!.content);
    fireEvent.change(screen.getByTestId("practice-input-textarea"), { target: { value: "missing ID must stay locked" } });
    fireEvent.click(screen.getByTestId("practice-input-send"));
    expect(await screen.findByText("This message could not be sent. Please try again later.")).toBeInTheDocument();

    fireEvent.click(screen.getByTestId("practice-error-state-retry"));
    await waitFor(() => expect(getSession).toHaveBeenCalledTimes(3));
    expect(send).toHaveBeenCalledTimes(1);
    expect(screen.getByTestId("practice-error-state")).toBeInTheDocument();
    expect(screen.queryByTestId("practice-message-retry")).not.toBeInTheDocument();
    expect(screen.getByTestId("practice-input-send")).toBeDisabled();
    expect(screen.getByTestId("practice-finish-cta")).toBeDisabled();
  });

  it("preserves terminal_failed recovery semantics when a retry receives a non-retryable response", async () => {
    localStorage.setItem("ei-lang", "en");
    const client = createDevMockClient();
    const base = openingOnly(await client.getPracticeSession(SESSION_ID));
    vi.spyOn(client, "getPracticeSession").mockResolvedValue(base);
    const send = vi.spyOn(client, "sendPracticeMessage")
      .mockRejectedValueOnce(new ApiClientError("transport", null, null, new TypeError("offline")))
      .mockRejectedValueOnce(new ApiClientError("http", 409, {
        error: {
          code: "IDEMPOTENCY_KEY_MISMATCH",
          message: "clientMessageId was reused with different text",
          requestId: "req-terminal-message",
          retryable: false,
        },
      }));

    renderPractice(client);
    await screen.findByText(base.messages[0]!.content);
    const text = "Answer that must remain as the terminal row";
    fireEvent.change(screen.getByTestId("practice-input-textarea"), { target: { value: text } });
    fireEvent.click(screen.getByTestId("practice-input-send"));

    const retry = await screen.findByTestId("practice-message-retry");
    const originalBody = send.mock.calls[0]![1];
    fireEvent.change(screen.getByTestId("practice-input-textarea"), { target: { value: "Preserved next draft" } });
    fireEvent.click(retry);

    expect(await screen.findByText("This message was not accepted. Check it before sending again.")).toBeInTheDocument();
    expect(send).toHaveBeenCalledTimes(2);
    expect(send.mock.calls[1]![1]).toEqual(originalBody);
    expect(screen.getAllByText(text)).toHaveLength(1);
    expect(screen.queryByText(/HTTP 409|IDEMPOTENCY_KEY_MISMATCH|clientMessageId was reused/u)).not.toBeInTheDocument();
    expect(screen.queryByTestId("practice-message-retry")).not.toBeInTheDocument();
    expect(screen.queryByTestId("practice-interviewer-thinking")).not.toBeInTheDocument();
    expect(screen.getByTestId("practice-input-textarea")).toHaveValue("Preserved next draft");
    expect(screen.getByTestId("practice-input-textarea")).toBeEnabled();
    expect(screen.getByTestId("practice-input-send")).toBeDisabled();
    expect(screen.getByTestId("practice-finish-cta")).toBeDisabled();
    expect(screen.queryByTestId("practice-terminal-recovery")).not.toBeInTheDocument();
  });

  it("offers one same-ID retry only when both a transport send and reconciliation fail, and preserves the next draft after repeated failure", async () => {
    const client = createDevMockClient();
    const base = openingOnly(await client.getPracticeSession(SESSION_ID));
    const transport = new ApiClientError("transport", null, null, new TypeError("offline"));
    vi.spyOn(client, "getPracticeSession")
      .mockResolvedValueOnce(base)
      .mockRejectedValue(transport);
    const send = vi.spyOn(client, "sendPracticeMessage").mockRejectedValue(transport);

    renderPractice(client);
    await screen.findByText(base.messages[0]!.content);
    fireEvent.change(screen.getByTestId("practice-input-textarea"), { target: { value: "网络未知结果回答" } });
    fireEvent.click(screen.getByTestId("practice-input-send"));

    const retry = await screen.findByTestId("practice-message-retry");
    const originalBody = send.mock.calls[0]![1];
    expect(screen.getByTestId("practice-input-textarea")).toBeEnabled();
    expect(screen.getByTestId("practice-input-send")).toBeDisabled();
    fireEvent.change(screen.getByTestId("practice-input-textarea"), { target: { value: "必须保留的下一条草稿" } });
    fireEvent.click(retry);

    await waitFor(() => expect(send).toHaveBeenCalledTimes(2));
    expect(send.mock.calls[1]![1]).toEqual(originalBody);
    await waitFor(() => expect(screen.getByTestId("practice-message-retry")).toBeInTheDocument());
    expect(screen.getByTestId("practice-input-textarea")).toHaveValue("必须保留的下一条草稿");
    expect(screen.getByTestId("practice-input-textarea")).toBeEnabled();
    expect(screen.getByTestId("practice-input-send")).toBeDisabled();
  });

  it("sends and retries the exact raw Markdown bytes with one clientMessageId without replacing the next draft", async () => {
    const client = createDevMockClient();
    const base = openingOnly(await client.getPracticeSession(SESSION_ID));
    const transport = new ApiClientError("transport", null, null, new TypeError("offline"));
    vi.spyOn(client, "getPracticeSession")
      .mockResolvedValueOnce(base)
      .mockRejectedValue(transport);
    const send = vi.spyOn(client, "sendPracticeMessage").mockRejectedValue(transport);

    renderPractice(client);
    await screen.findByText(base.messages[0]!.content);
    const rawText = "\n\n## 原始回答\n\n- 第一项\n\n<div onclick=\"unsafe()\">保留的原始 HTML</div>\n\n  ";
    fireEvent.change(screen.getByTestId("practice-input-textarea"), { target: { value: rawText } });
    fireEvent.click(screen.getByTestId("practice-input-send"));

    const retry = await screen.findByTestId("practice-message-retry");
    const firstBody = send.mock.calls[0]![1];
    expect(Array.from(new TextEncoder().encode(firstBody.text))).toEqual(
      Array.from(new TextEncoder().encode(rawText)),
    );
    expect(screen.getByRole("heading", { level: 2, name: "原始回答" })).toBeInTheDocument();
    expect(screen.queryByText("保留的原始 HTML")).not.toBeInTheDocument();

    const nextDraft = "下一条草稿不能被重试覆盖";
    fireEvent.change(screen.getByTestId("practice-input-textarea"), { target: { value: nextDraft } });
    fireEvent.click(retry);

    await waitFor(() => expect(send).toHaveBeenCalledTimes(2));
    const retriedBody = send.mock.calls[1]![1];
    expect(retriedBody.clientMessageId).toBe(firstBody.clientMessageId);
    expect(Array.from(new TextEncoder().encode(retriedBody.text))).toEqual(
      Array.from(new TextEncoder().encode(rawText)),
    );
    expect(screen.getByTestId("practice-input-textarea")).toHaveValue(nextDraft);
  });

  it("uses typed ApiClientError for loader 404 instead of parsing an Error.message prefix", async () => {
    const plainClient = createDevMockClient();
    vi.spyOn(plainClient, "getPracticeSession").mockRejectedValue(new Error("HTTP 404 plain text is not a typed status"));
    const first = renderPractice(plainClient);
    expect(await screen.findByText("The interview session is temporarily unavailable. Please try again.")).toBeInTheDocument();
    expect(screen.queryByText("HTTP 404 plain text is not a typed status")).not.toBeInTheDocument();
    expect(screen.queryByTestId("practice-session-lost")).not.toBeInTheDocument();
    first.unmount();

    const typedClient = createDevMockClient();
    vi.spyOn(typedClient, "getPracticeSession").mockRejectedValue(new ApiClientError("http", 404, null));
    renderPractice(typedClient);
    expect(await screen.findByTestId("practice-session-lost")).toBeInTheDocument();
  });
});

const SESSION_ID = "01918fa0-0000-7000-8000-000000005000";
const SESSION_B_ID = "01918fa0-0000-7000-8000-000000005001";
const PLAN_ID = "01918fa0-0000-7000-8000-000000004000";
const TARGET_JOB_ID = "01918fa0-0000-7000-8000-000000002000";
const TARGET_JOB_B_ID = "01918fa0-0000-7000-8000-000000002001";
const SERVER_MESSAGE_ID = "01918fa0-0000-7000-8000-000000006099";

function renderPractice(
  client: ReturnType<typeof createDevMockClient>,
  params: Record<string, string> = {},
) {
  return render(<App client={client} requestOptions={{ getMe: { headers: { Prefer: "example=authenticated" } } }} initialRoute={{ name: "practice", params: {
    sessionId: SESSION_ID,
    planId: PLAN_ID,
    targetJobId: TARGET_JOB_ID,
    ...params,
  } }} />);
}

function openingOnly(session: PracticeSession): PracticeSession {
  return {
    ...session,
    messages: session.messages.filter((message) => message.role === "assistant").slice(0, 1),
  };
}

function sessionFor(
  session: PracticeSession,
  sessionId: string,
  opening: string,
  targetJobId: string,
): PracticeSession {
  const first = session.messages.find((message) => message.role === "assistant");
  if (!first) throw new Error("expected opening assistant message");
  return {
    ...session,
    id: sessionId,
    targetJobId,
    messages: [{ ...first, content: opening }],
  };
}

function practiceUrl(sessionId: string, targetJobId: string): string {
  return `/practice?planId=${PLAN_ID}&sessionId=${sessionId}&targetJobId=${targetJobId}`;
}

function completionReport(reportId: string): ReportWithJob {
  return {
    reportId,
    job: {
      id: `job-${reportId}`,
      jobType: "report_generate",
      status: "queued",
      resourceType: "feedback_report",
      resourceId: reportId,
      errorCode: null,
      createdAt: "2026-07-14T10:00:00Z",
      updatedAt: "2026-07-14T10:00:00Z",
    },
  };
}

function sessionWithUser(
  session: PracticeSession,
  clientMessageId: string,
  content: string,
  replyStatus: PracticeUserMessage["replyStatus"],
): PracticeSession {
  const user: PracticeUserMessage = {
    id: clientMessageId,
    clientMessageId,
    seqNo: 2,
    role: "user",
    content,
    replyStatus,
    createdAt: "2026-07-13T08:03:00Z",
  };
  return { ...session, messages: [...session.messages, user] };
}

function completedResponse(
  session: PracticeSession,
  clientMessageId: string,
  content: string,
  assistantContent = "服务端唯一回复",
): SendPracticeMessageResponse {
  const withUser = sessionWithUser(session, clientMessageId, content, "complete");
  const assistantMessage: PracticeAssistantMessage = {
    id: "01918fa0-0000-7000-8000-000000006100",
    seqNo: 3,
    role: "assistant",
    content: assistantContent,
    createdAt: "2026-07-13T08:03:01Z",
  };
  return {
    acknowledged: true,
    userMessage: withUser.messages.at(-1) as PracticeUserMessage,
    assistantMessage,
    session: { ...withUser, messages: [...withUser.messages, assistantMessage] },
  };
}

async function flushPromises(): Promise<void> {
  await act(async () => {
    await Promise.resolve();
    await Promise.resolve();
    await Promise.resolve();
  });
}
