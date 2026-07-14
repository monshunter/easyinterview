import { useCallback, useMemo } from "react";

import type { RequestOptions } from "../../../../api/generated/client";
import type { SendPracticeMessageResponse } from "../../../../api/generated/types";
import { useInterviewContext } from "../../../interview-context/InterviewContext";
import { useAppRuntimeOptional } from "../../../runtime/AppRuntimeProvider";
import { resolveContentLimits, utf8ByteLength } from "../../../../lib/contentLimits";

export interface PracticeMessageSubmission {
  text: string;
  clientMessageId: string;
}

export interface UsePracticeMessagesResult {
  ready: boolean;
  sendMessage: (
    submission: PracticeMessageSubmission,
    options?: Pick<RequestOptions, "signal">,
  ) => Promise<SendPracticeMessageResponse>;
}

export function usePracticeMessages(explicitSessionId?: string): UsePracticeMessagesResult {
  const runtime = useAppRuntimeOptional();
  const client = runtime?.client;
  const { ctx } = useInterviewContext();
  const sessionId = explicitSessionId ?? ctx.sessionId ?? "";
  const maxMessageBytes = resolveContentLimits(
    runtime?.runtime.status === "ready" ? runtime.runtime.config : undefined,
  ).practiceMessageBytes;

  const sendMessage = useCallback(async (
    submission: PracticeMessageSubmission,
    options?: Pick<RequestOptions, "signal">,
  ) => {
    if (!client) throw new Error("usePracticeMessages: client not mounted");
    if (!sessionId) throw new Error("usePracticeMessages: sessionId missing");
    if (!submission.text.trim()) throw new Error("usePracticeMessages: text missing");
    if (!submission.clientMessageId) throw new Error("usePracticeMessages: clientMessageId missing");
    if (utf8ByteLength(submission.text.trim()) > maxMessageBytes) {
      throw new Error("PRACTICE_MESSAGE_TOO_LARGE");
    }
    return client.sendPracticeMessage(sessionId, submission, options);
  }, [client, maxMessageBytes, sessionId]);

  return useMemo(() => ({ ready: Boolean(client && sessionId), sendMessage }), [client, sessionId, sendMessage]);
}
