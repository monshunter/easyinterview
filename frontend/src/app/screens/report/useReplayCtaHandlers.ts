import { useCallback, useMemo } from "react";

import type { FeedbackReport } from "../../../api/generated/types";
import { useNavigation } from "../../navigation/NavigationProvider";
import { useRequestAuth } from "../../auth/useRequestAuth";
import { useAppRuntimeOptional } from "../../runtime/AppRuntimeProvider";
import type { Route } from "../../routes";
import { useI18n } from "../../i18n/messages";
import { startPracticeFromParams } from "../../interview-context/startPractice";
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
 * - Signed-in users create/start a fresh practice session from report scope,
 *   then land directly on `practice`.
 * - Signed-out users go through auth and return to report; replay can be
 *   retried there without using `workspace` as a side-effect route.
 */
export function useReplayCtaHandlers(
  input: ReplayCtaHandlersInput,
): ReplayCtaHandlers {
  const { route, report, sessionId } = input;
  const { navigate } = useNavigation();
  const requestAuth = useRequestAuth();
  const runtime = useAppRuntimeOptional();
  const { lang } = useI18n();
  const authenticated = runtime?.auth.status === "authenticated";

  const replayParams = useMemo(
    () => buildReplayPayload({ route, report, sessionId }),
    [report, route, sessionId],
  );
  const nextRoundParams = useMemo(
    () => buildNextRoundPayload({ route, report, sessionId }),
    [report, route, sessionId],
  );
  const startPractice = useCallback(
    async (params: Record<string, string>) => {
      if (!runtime) return;
      const started = await startPracticeFromParams(runtime.client, params, lang);
      navigate({ name: "practice", params: started.params });
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

  return { goReplay, goNextRound };
}
