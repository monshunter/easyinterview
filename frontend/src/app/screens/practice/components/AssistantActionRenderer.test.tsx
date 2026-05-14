/**
 * @vitest-environment jsdom
 *
 * Item 2.2 — AssistantActionRenderer dispatches the 5 assistantAction.type
 * branches via callbacks, and never leaks `provenance` into the main
 * conversation surface. Provenance must only be rendered by RightPanel's
 * AI TRANSPARENCY card.
 */

import { describe, expect, it, vi } from "vitest";
import { render, screen } from "@testing-library/react";

import type { AssistantAction } from "../../../../api/generated/types";
import { AssistantActionRenderer } from "./AssistantActionRenderer";

const SESSION_RUNNING = "running" as AssistantAction["sessionStatus"];
const SESSION_WAITING = "waiting_user_input" as AssistantAction["sessionStatus"];
const SESSION_COMPLETED = "completed" as AssistantAction["sessionStatus"];

function makeAction(
  overrides: Partial<AssistantAction>,
): AssistantAction {
  return {
    type: "ask_question",
    sessionStatus: SESSION_RUNNING,
    turnId: "turn-1",
    questionText: null,
    hint: null,
    provenance: {
      promptVersion: "v1.0.4",
      rubricVersion: "v0.9",
      modelId: "haiku-4.5",
      language: "zh-CN",
      featureFlag: "follow_up_v3",
      dataSourceVersion: "practice_session.v9",
    },
    ...overrides,
  };
}

function makeCallbacks() {
  return {
    onAskQuestion: vi.fn(),
    onAskFollowUp: vi.fn(),
    onShowHint: vi.fn(),
    onSessionWait: vi.fn(),
    onSessionCompleted: vi.fn(),
  };
}

describe("AssistantActionRenderer", () => {
  it("ask_question: invokes onAskQuestion(turnId, questionText) exactly once", () => {
    const cbs = makeCallbacks();
    render(
      <AssistantActionRenderer
        action={makeAction({
          type: "ask_question",
          turnId: "turn-2",
          questionText: "Q2 prompt",
        })}
        {...cbs}
      />,
    );
    expect(cbs.onAskQuestion).toHaveBeenCalledTimes(1);
    expect(cbs.onAskQuestion).toHaveBeenCalledWith("turn-2", "Q2 prompt");
    expect(cbs.onAskFollowUp).not.toHaveBeenCalled();
    expect(cbs.onShowHint).not.toHaveBeenCalled();
    expect(cbs.onSessionWait).not.toHaveBeenCalled();
    expect(cbs.onSessionCompleted).not.toHaveBeenCalled();
  });

  it("ask_follow_up: invokes onAskFollowUp(turnId, questionText) exactly once", () => {
    const cbs = makeCallbacks();
    render(
      <AssistantActionRenderer
        action={makeAction({
          type: "ask_follow_up",
          turnId: "turn-1",
          questionText: "Follow up?",
        })}
        {...cbs}
      />,
    );
    expect(cbs.onAskFollowUp).toHaveBeenCalledTimes(1);
    expect(cbs.onAskFollowUp).toHaveBeenCalledWith("turn-1", "Follow up?");
    expect(cbs.onAskQuestion).not.toHaveBeenCalled();
  });

  it("show_hint: invokes onShowHint(hint, turnId) exactly once", () => {
    const cbs = makeCallbacks();
    render(
      <AssistantActionRenderer
        action={makeAction({
          type: "show_hint",
          turnId: "turn-3",
          hint: "Try STAR + numbers",
        })}
        {...cbs}
      />,
    );
    expect(cbs.onShowHint).toHaveBeenCalledTimes(1);
    expect(cbs.onShowHint).toHaveBeenCalledWith("Try STAR + numbers", "turn-3");
  });

  it("session_wait: invokes onSessionWait() exactly once", () => {
    const cbs = makeCallbacks();
    render(
      <AssistantActionRenderer
        action={makeAction({
          type: "session_wait",
          sessionStatus: SESSION_WAITING,
          questionText: null,
        })}
        {...cbs}
      />,
    );
    expect(cbs.onSessionWait).toHaveBeenCalledTimes(1);
  });

  it("session_completed: invokes onSessionCompleted() exactly once", () => {
    const cbs = makeCallbacks();
    render(
      <AssistantActionRenderer
        action={makeAction({
          type: "session_completed",
          sessionStatus: SESSION_COMPLETED,
        })}
        {...cbs}
      />,
    );
    expect(cbs.onSessionCompleted).toHaveBeenCalledTimes(1);
  });

  it("provenance never leaks into the rendered DOM", () => {
    const cbs = makeCallbacks();
    const { container } = render(
      <AssistantActionRenderer
        action={makeAction({
          type: "ask_follow_up",
          questionText: "Follow up?",
          provenance: {
            promptVersion: "v9.9.9",
            rubricVersion: "v8.8",
            modelId: "secret-model-id",
            language: "zh-CN",
            featureFlag: "secret-flag",
            dataSourceVersion: "v42",
          },
        })}
        {...cbs}
      />,
    );
    expect(container.textContent).not.toContain("v9.9.9");
    expect(container.textContent).not.toContain("secret-model-id");
    expect(container.textContent).not.toContain("secret-flag");
    expect(container.textContent).not.toContain("v42");
    // The renderer should be a side-effect-only sentinel; no visible UI.
    expect(screen.queryByTestId("practice-assistant-action-render")).not.toBeNull();
    const sentinel = screen.getByTestId("practice-assistant-action-render");
    expect(sentinel.textContent).toBe("");
  });

  it("re-renders with the same action object do not re-fire callbacks", () => {
    const cbs = makeCallbacks();
    const action = makeAction({
      type: "ask_question",
      questionText: "Q1",
    });
    const { rerender } = render(
      <AssistantActionRenderer action={action} {...cbs} />,
    );
    rerender(<AssistantActionRenderer action={action} {...cbs} />);
    rerender(<AssistantActionRenderer action={action} {...cbs} />);
    expect(cbs.onAskQuestion).toHaveBeenCalledTimes(1);
  });

  it("a new action object (different reference) re-fires the matching callback", () => {
    const cbs = makeCallbacks();
    const { rerender } = render(
      <AssistantActionRenderer
        action={makeAction({ type: "ask_question", questionText: "Q1" })}
        {...cbs}
      />,
    );
    rerender(
      <AssistantActionRenderer
        action={makeAction({
          type: "ask_question",
          questionText: "Q2",
          turnId: "turn-99",
        })}
        {...cbs}
      />,
    );
    expect(cbs.onAskQuestion).toHaveBeenCalledTimes(2);
    expect(cbs.onAskQuestion).toHaveBeenLastCalledWith("turn-99", "Q2");
  });

  it("null action renders nothing and fires no callbacks", () => {
    const cbs = makeCallbacks();
    const { container } = render(
      <AssistantActionRenderer action={null} {...cbs} />,
    );
    expect(container.textContent).toBe("");
    expect(cbs.onAskQuestion).not.toHaveBeenCalled();
    expect(cbs.onAskFollowUp).not.toHaveBeenCalled();
    expect(cbs.onShowHint).not.toHaveBeenCalled();
    expect(cbs.onSessionWait).not.toHaveBeenCalled();
    expect(cbs.onSessionCompleted).not.toHaveBeenCalled();
  });
});
