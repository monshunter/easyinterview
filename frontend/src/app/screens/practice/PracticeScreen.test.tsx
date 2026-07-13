/** @vitest-environment jsdom */
import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { createDevMockClient } from "../../../api/devMockClient";
import { App } from "../../App";

afterEach(() => localStorage.removeItem("ei-lang"));

describe("PracticeScreen continuous conversation", () => {
  it("TestE2EP0047RejectsZeroAnswerCompletion: keeps Finish natively disabled with a localized accessible reason until a committed answer is complete", async () => {
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
    vi.spyOn(client, "getPracticeSession").mockResolvedValue({
      ...session,
      messages: session.messages.slice(0, 2),
    });

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
    expect(await screen.findByText("HTTP 503 completion unavailable")).toBeInTheDocument();
    fireEvent.click(screen.getByTestId("practice-error-state-retry"));

    await waitFor(() => expect(complete).toHaveBeenCalledTimes(2));
    expect(send).not.toHaveBeenCalled();
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

    expect(await screen.findByText("HTTP 503 session unavailable")).toBeInTheDocument();
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
});
