/**
 * @vitest-environment jsdom
 */

import { describe, expect, it } from "vitest";
import { act, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

import type {
  AssistantAction,
  PracticeSession,
  SessionEventResult,
} from "../../../../api/generated/types";
import {
  SESSION_A,
  TARGET_JOB_A,
  TURN_A,
  buildPracticeClient,
  eventCalls,
  mountPracticeScreen,
  readBody,
} from "./practiceTestUtils";

const TURN_B = "01918fa0-0000-7000-8000-000000006001";
const SESSION_B = "01918fa0-0000-7000-8000-000000005001";

describe("PracticeScreen server session continuity", () => {
  it("adopts each successful mutation snapshot so the next answer targets the new turn", async () => {
    const { client, calls } = buildPracticeClient({
      appendResults: [
        eventResult(
          practiceSession(TURN_B, 2, "请说明这次迁移最终产生了什么量化结果？"),
          assistantAction("ask_question", TURN_B, "请说明这次迁移最终产生了什么量化结果？"),
        ),
        eventResult(
          practiceSession(TURN_B, 2, "这个结果与最初目标相比如何？"),
          assistantAction("ask_follow_up", TURN_B, "这个结果与最初目标相比如何？"),
        ),
      ],
    });
    mountPracticeScreen({ client });
    const user = userEvent.setup();
    const textarea = screen.getByTestId(
      "practice-input-textarea",
    ) as HTMLTextAreaElement;

    await waitFor(() => expect(textarea.disabled).toBe(false));
    await user.type(textarea, "第一次回答");
    await user.click(screen.getByTestId("practice-input-send"));

    await waitFor(() =>
      expect(screen.getByTestId("practice-question-prompt")).toHaveTextContent(
        "请说明这次迁移最终产生了什么量化结果？",
      ),
    );
    expect(textarea.value).toBe("");

    await user.type(textarea, "第二次回答");
    await user.click(screen.getByTestId("practice-input-send"));
    await waitFor(() => expect(eventCalls(calls)).toHaveLength(2));

    const firstPayload = readBody(eventCalls(calls)[0]!).payload as Record<
      string,
      unknown
    >;
    const secondPayload = readBody(eventCalls(calls)[1]!).payload as Record<
      string,
      unknown
    >;
    expect(firstPayload.turnId).toBe(TURN_A);
    expect(secondPayload.turnId).toBe(TURN_B);
  });

  it("resets local session state before rendering a different practice session", async () => {
    const questionA = "A 会话首题：请介绍当前项目。";
    const questionB = "B 会话首题：请说明下一阶段目标。";
    const sessionA = practiceSession(TURN_A, 1, questionA);
    const sessionB = practiceSession(TURN_B, 1, questionB, SESSION_B);
    const { client } = buildPracticeClient({
      scenarioByOp: { appendSessionEvent: "show-hint" },
      sessionResults: {
        [SESSION_A]: sessionA,
        [SESSION_B]: sessionB,
      },
    });
    const harness = mountPracticeScreen({ client });
    const user = userEvent.setup();

    expect(await screen.findByText(questionA)).toBeDefined();
    const textareaA = screen.getByTestId(
      "practice-input-textarea",
    ) as HTMLTextAreaElement;
    await user.type(textareaA, "只属于 A 的未提交草稿");
    await user.click(screen.getByTestId("practice-input-hint"));
    expect(await screen.findByTestId("practice-hint-banner")).toBeDefined();
    await user.click(screen.getByTestId("practice-topbar-pause"));
    await waitFor(() =>
      expect(screen.getByTestId("practice-topbar-pause")).toHaveTextContent(
        "Resume",
      ),
    );

    act(() => {
      harness.rerenderPractice({
        sessionId: SESSION_B,
        targetJobId: TARGET_JOB_A,
      });
    });

    const textareaB = screen.getByTestId(
      "practice-input-textarea",
    ) as HTMLTextAreaElement;
    expect(textareaB.value).toBe("");
    expect(screen.getByTestId("practice-topbar-pause")).toHaveTextContent(
      "Pause",
    );
    expect(screen.queryByTestId("practice-hint-banner")).toBeNull();
    expect(screen.queryByText(questionA)).toBeNull();
    expect(screen.queryByText("只属于 A 的未提交草稿")).toBeNull();
    expect(await screen.findByText(questionB)).toBeDefined();
    expect(screen.getAllByTestId(/practice-transcript-message-/)).toHaveLength(
      1,
    );
  });
});

function practiceSession(
  turnId: string,
  turnIndex: number,
  questionText: string,
  sessionId = SESSION_A,
): PracticeSession {
  return {
    id: sessionId,
    planId: "01918fa0-0000-7000-8000-000000004000",
    targetJobId: TARGET_JOB_A,
    status: "running",
    language: "zh-CN",
    hintsEnabled: true,
    turnCount: turnIndex,
    currentTurn: {
      id: turnId,
      turnIndex,
      questionText,
      status: "asked",
    },
    createdAt: "2026-07-11T00:00:00Z",
    updatedAt: "2026-07-11T00:00:01Z",
  };
}

function assistantAction(
  type: AssistantAction["type"],
  turnId: string,
  questionText: string,
): AssistantAction {
  return {
    type,
    turnId,
    questionText,
    sessionStatus: "running",
    provenance: {
      promptVersion: "practice.session.follow_up.v1",
      rubricVersion: "not_applicable",
      modelId: "model-profile:test",
      language: "zh-CN",
      featureFlag: "none",
      dataSourceVersion: "test",
    },
  };
}

function eventResult(
  session: PracticeSession,
  assistantActionValue: AssistantAction,
): SessionEventResult {
  return {
    acknowledged: true,
    session,
    assistantAction: assistantActionValue,
  };
}
