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
 * - Signed-in users land on `workspace` with the path's payload so the
 *   workspace owner can create a fresh session before entering practice.
 * - Signed-out users go through `useRequestAuth({type:'replay_practice', …})`
 *   so login auto-resumes to the same workspace auto-start payload.
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
  const replayStartParams = useMemo(
    () => ({ ...replayParams, autoStartPractice: "1" }),
    [replayParams],
  );
  const nextRoundStartParams = useMemo(
    () => ({ ...nextRoundParams, autoStartPractice: "1" }),
    [nextRoundParams],
  );

  const goReplay = useCallback(() => {
    if (authenticated) {
      navigate({ name: "workspace", params: replayStartParams });
      return;
    }
    requestAuth({
      type: "replay_practice",
      label: "replay",
      route: "workspace",
      params: replayStartParams,
    });
  }, [authenticated, navigate, replayStartParams, requestAuth]);

  const goNextRound = useCallback(() => {
    if (authenticated) {
      navigate({ name: "workspace", params: nextRoundStartParams });
      return;
    }
    requestAuth({
      type: "replay_practice",
      label: "next-round",
      route: "workspace",
      params: nextRoundStartParams,
    });
  }, [authenticated, navigate, nextRoundStartParams, requestAuth]);

  return { goReplay, goNextRound };
}
