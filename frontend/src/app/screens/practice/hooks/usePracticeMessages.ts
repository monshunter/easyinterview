import { useCallback, useMemo, useRef } from "react";

import type { SendPracticeMessageResponse } from "../../../../api/generated/types";
import { newId } from "../../../../lib/ids";
import { useInterviewContext } from "../../../interview-context/InterviewContext";
import { useAppRuntimeOptional } from "../../../runtime/AppRuntimeProvider";

export interface UsePracticeMessagesResult {
  ready: boolean;
  sendMessage: (text: string) => Promise<SendPracticeMessageResponse>;
}

export function usePracticeMessages(explicitSessionId?: string): UsePracticeMessagesResult {
  const runtime = useAppRuntimeOptional();
  const client = runtime?.client;
  const { ctx } = useInterviewContext();
  const sessionId = explicitSessionId ?? ctx.sessionId ?? "";
  const retryRef = useRef<{ text: string; clientMessageId: string } | null>(null);

  const sendMessage = useCallback(async (rawText: string) => {
    if (!client) throw new Error("usePracticeMessages: client not mounted");
    if (!sessionId) throw new Error("usePracticeMessages: sessionId missing");
    const text = rawText.trim();
    if (!text) throw new Error("usePracticeMessages: text missing");
    const clientMessageId = retryRef.current?.text === text ? retryRef.current.clientMessageId : newId();
    retryRef.current = { text, clientMessageId };
    try {
      const result = await client.sendPracticeMessage(sessionId, { clientMessageId, text });
      if (result.acknowledged) retryRef.current = null;
      return result;
    } catch (error) {
      throw error;
    }
  }, [client, sessionId]);

  return useMemo(() => ({ ready: Boolean(client && sessionId), sendMessage }), [client, sessionId, sendMessage]);
}
