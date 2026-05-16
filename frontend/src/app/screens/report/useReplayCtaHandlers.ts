import { useCallback, useMemo } from "react";

import type { FeedbackReport } from "../../../api/generated/types";
import { useNavigation } from "../../navigation/NavigationProvider";
import { useRequestAuth } from "../../auth/useRequestAuth";
import { useAppRuntimeOptional } from "../../runtime/AppRuntimeProvider";
import type { Route } from "../../routes";
import {
  buildNextRoundPayload,
  buildReplayPayload,
} from "./handoff";

export interface ReplayCtaHandlersInput {
  route: Route;
  report: FeedbackReport | null;
  sessionId: string;
}

export interface ReplayCtaHandlers {
  goReplay: () => void;
  goNextRound: () => void;
}

/**
 * Centralizes the replay / next-round CTA flow so both `ReportHeader` (top
 * CTAs) and `NextTab` (path A/B cards) share one source of truth.
 *
 * - Signed-in users land directly on `practice` with the path's payload.
 * - Signed-out users go through `useRequestAuth({type:'replay_practice', …})`
 *   so login auto-resumes to practice carrying the same payload + the
 *   `autoReplay=1` marker for the AppPendingAction hook.
 */
export function useReplayCtaHandlers(
  input: ReplayCtaHandlersInput,
): ReplayCtaHandlers {
  const { route, report, sessionId } = input;
  const { navigate } = useNavigation();
  const requestAuth = useRequestAuth();
  const runtime = useAppRuntimeOptional();
  const authenticated = runtime?.auth.status === "authenticated";

  const replayParams = useMemo(
    () => buildReplayPayload({ route, report, sessionId }),
    [report, route, sessionId],
  );
  const nextRoundParams = useMemo(
    () => buildNextRoundPayload({ route, report, sessionId }),
    [report, route, sessionId],
  );

  const goReplay = useCallback(() => {
    if (authenticated) {
      navigate({ name: "practice", params: replayParams });
      return;
    }
    requestAuth({
      type: "replay_practice",
      label: "replay",
      route: "practice",
      params: { ...replayParams, autoReplay: "1" },
    });
  }, [authenticated, navigate, replayParams, requestAuth]);

  const goNextRound = useCallback(() => {
    if (authenticated) {
      navigate({ name: "practice", params: nextRoundParams });
      return;
    }
    requestAuth({
      type: "replay_practice",
      label: "next-round",
      route: "practice",
      params: { ...nextRoundParams, autoReplay: "1" },
    });
  }, [authenticated, navigate, nextRoundParams, requestAuth]);

  return { goReplay, goNextRound };
}
