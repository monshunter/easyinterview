import { useEffect, type FC } from "react";

import type { AssistantAction } from "../../../../api/generated/types";

export interface AssistantActionRendererProps {
  action: AssistantAction | null;
  onAskQuestion: (turnId: string, questionText: string) => void;
  onAskFollowUp: (turnId: string, questionText: string) => void;
  onShowHint: (hint: string, turnId: string) => void;
  onSessionWait: () => void;
  onSessionCompleted: () => void;
}

/**
 * Side-effect-only component: dispatches one callback per AssistantAction
 * change, keyed on the action object reference. Provenance MUST stay in
 * RightPanel's AI TRANSPARENCY card; this renderer never reads or echoes
 * provenance fields, so they cannot leak into the main conversation flow.
 *
 * The renderer emits a tiny sentinel `<span>` with no text content so
 * tests can confirm it mounted without it producing any visible UI.
 */
export const AssistantActionRenderer: FC<AssistantActionRendererProps> = ({
  action,
  onAskQuestion,
  onAskFollowUp,
  onShowHint,
  onSessionWait,
  onSessionCompleted,
}) => {
  useEffect(() => {
    if (!action) return;
    switch (action.type) {
      case "ask_question":
        onAskQuestion(action.turnId ?? "", action.questionText ?? "");
        return;
      case "ask_follow_up":
        onAskFollowUp(action.turnId ?? "", action.questionText ?? "");
        return;
      case "show_hint":
        onShowHint(action.hint ?? "", action.turnId ?? "");
        return;
      case "session_wait":
        onSessionWait();
        return;
      case "session_completed":
        onSessionCompleted();
        return;
    }
  }, [
    action,
    onAskQuestion,
    onAskFollowUp,
    onShowHint,
    onSessionWait,
    onSessionCompleted,
  ]);

  if (!action) return null;
  return <span data-testid="practice-assistant-action-render" hidden />;
};
