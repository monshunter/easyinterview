import { useCallback, useMemo, useRef, useState } from "react";

import type { FeedbackReport } from "../../../api/generated/types";
import { useNavigation } from "../../navigation/NavigationProvider";
import { useRequestAuth } from "../../auth/useRequestAuth";
import { useAppRuntimeOptional } from "../../runtime/AppRuntimeProvider";
import type { Route } from "../../routes";
import { useI18n } from "../../i18n/messages";
import { startPracticeFromParams } from "../../interview-context/startPractice";
import type { TargetJobRoundAssumption } from "../../interview-context/roundAssumptions";
import {
  buildNextRoundPayload,
  buildReplayPayload,
} from "./handoff";

export interface ReplayCtaHandlersInput {
  route: Route;
  report: FeedbackReport | null;
  sessionId: string;
  nextRound: TargetJobRoundAssumption | null;
}

export interface ReplayCtaHandlers {
  goReplay: () => void;
  goNextRound: () => void;
  canNextRound: boolean;
  starting: boolean;
}

/**
 * Centralizes the replay / next-round flow for the report Header CTAs.
 *
 * - Signed-in users create/start a fresh practice session from report scope,
 *   then land directly on `practice`.
 * - Signed-out users go through auth and return to report; replay can be
 *   retried there without using `workspace` as a side-effect route.
 */
export function useReplayCtaHandlers(
  input: ReplayCtaHandlersInput,
): ReplayCtaHandlers {
  const { route, report, sessionId, nextRound } = input;
  const { navigate } = useNavigation();
  const requestAuth = useRequestAuth();
  const runtime = useAppRuntimeOptional();
  const { lang } = useI18n();
  const authenticated = runtime?.auth.status === "authenticated";
  const startingRef = useRef(false);
  const [starting, setStarting] = useState(false);

  const replayParams = useMemo(
    () => buildReplayPayload({ route, report, sessionId }),
    [report, route, sessionId],
  );
  const nextRoundParams = useMemo(
    () => nextRound ? buildNextRoundPayload({ route, report, sessionId }, nextRound) : null,
    [nextRound, report, route, sessionId],
  );
  const startPractice = useCallback(
    async (params: Record<string, string>) => {
      if (!runtime || startingRef.current) return;
      startingRef.current = true;
      setStarting(true);
      try {
        const started = await startPracticeFromParams(runtime.client, params, lang);
        navigate({ name: "practice", params: started.params });
      } finally {
        startingRef.current = false;
        setStarting(false);
      }
    },
    [lang, navigate, runtime],
  );

  const goReplay = useCallback(() => {
    if (authenticated) {
      void startPractice(replayParams);
      return;
    }
    requestAuth({
      type: "replay_practice",
      label: "replay",
      route: "report",
      params: route.params,
    });
  }, [authenticated, replayParams, requestAuth, route.params, startPractice]);

  const goNextRound = useCallback(() => {
    if (!nextRoundParams) return;
    if (authenticated) {
      void startPractice(nextRoundParams);
      return;
    }
    requestAuth({
      type: "replay_practice",
      label: "next-round",
      route: "report",
      params: route.params,
    });
  }, [authenticated, nextRoundParams, requestAuth, route.params, startPractice]);

  return {
    goReplay,
    goNextRound,
    canNextRound: nextRoundParams !== null,
    starting,
  };
}
