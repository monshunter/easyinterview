import { useEffect, useRef, useState } from "react";

import type { EasyInterviewClient } from "../../../../api/generated/client";
import type { ReportConversation } from "../../../../api/generated/types";
import { useAppRuntimeOptional } from "../../../runtime/AppRuntimeProvider";

export type UseReportConversationState =
  | "idle"
  | "loading"
  | "data"
  | "error"
  | "notFound";

export interface UseReportConversationResult {
  state: UseReportConversationState;
  data: ReportConversation | null;
}

interface ConversationOwner {
  client: EasyInterviewClient | null;
  reportId: string;
}

interface OwnedConversation {
  client: EasyInterviewClient;
  reportId: string;
  value: ReportConversation;
}

const HTTP_NOT_FOUND_MARKER = "HTTP 404";

/**
 * Single-shot report-owned transcript loader. Its owner fence prevents a
 * late response for one reportId from rendering after a route switch.
 */
export function useReportConversation(
  reportId: string,
): UseReportConversationResult {
  const runtime = useAppRuntimeOptional();
  const client = runtime?.client ?? null;
  const initial: UseReportConversationState = !reportId
    ? "error"
    : client
      ? "loading"
      : "idle";
  const [state, setState] = useState<UseReportConversationState>(initial);
  const [stateOwner, setStateOwner] = useState<ConversationOwner>(() => ({
    client,
    reportId,
  }));
  const [ownedData, setOwnedData] = useState<OwnedConversation | null>(null);
  const runSequence = useRef(0);

  useEffect(() => {
    setStateOwner({ client, reportId });
    setOwnedData(null);
    if (!reportId) {
      setState("error");
      return;
    }
    if (!client) {
      setState("idle");
      return;
    }
    setState("loading");

    const sequence = runSequence.current + 1;
    runSequence.current = sequence;
    let cancelled = false;

    client
      .getReportConversation(reportId)
      .then((value) => {
        if (cancelled || runSequence.current !== sequence) return;
        setOwnedData({ client, reportId, value });
        setState("data");
      })
      .catch((error: unknown) => {
        if (cancelled || runSequence.current !== sequence) return;
        const message = error instanceof Error ? error.message : String(error);
        setState(message.startsWith(HTTP_NOT_FOUND_MARKER) ? "notFound" : "error");
      });

    return () => {
      cancelled = true;
    };
  }, [client, reportId]);

  const ownerMatches = stateOwner.client === client && stateOwner.reportId === reportId;
  const data =
    ownedData?.client === client && ownedData.reportId === reportId
      ? ownedData.value
      : null;

  return { state: ownerMatches ? state : initial, data };
}
