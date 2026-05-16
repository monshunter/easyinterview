import type { FC } from "react";

import { useNavigation } from "../../navigation/NavigationProvider";
import type { Route } from "../../routes";
import { ReportDashboard } from "./components/ReportDashboard";
import { ReportFailureState } from "./components/ReportFailureState";
import { ReportMissingSessionState } from "./components/ReportMissingSessionState";

interface ReportScreenProps {
  route: Route;
}

/**
 * Source-level mirror of `ui-design/src/screen-report.jsx::ReportScreen`
 * (lines 1-44). Three-way dispatch based on route params:
 *   - reportStatus === 'failed' → ReportFailureState (errorCode-driven copy).
 *   - missing sessionId         → ReportMissingSessionState.
 *   - otherwise                 → ReportDashboard (real dashboard with hooks).
 *
 * The dispatch is deliberately stateless — InterviewContext is hydrated by
 * App.tsx::InterviewContextRouteSync, and ReportDashboard owns the real
 * `useFeedbackReport` read path with its own loading / data / error / notFound
 * branches.
 */
export const ReportScreen: FC<ReportScreenProps> = ({ route }) => {
  const { navigate } = useNavigation();
  const params = route.params;

  if (params.reportStatus === "failed") {
    return (
      <ReportFailureState
        errorCode={params.errorCode ?? null}
        onRetry={() =>
          navigate({
            name: "generating",
            params: rebuildHandoffParams(params),
          })
        }
        onBackToWorkspace={() =>
          navigate({
            name: "workspace",
            params: rebuildHandoffParams(params),
          })
        }
      />
    );
  }

  if (!params.sessionId) {
    return (
      <ReportMissingSessionState
        onBackToWorkspace={() =>
          navigate({
            name: "workspace",
            params: rebuildHandoffParams(params),
          })
        }
      />
    );
  }

  if (!params.reportId) {
    return (
      <ReportMissingSessionState
        kind="missingReport"
        onBackToWorkspace={() =>
          navigate({
            name: "workspace",
            params: rebuildHandoffParams(params),
          })
        }
      />
    );
  }

  return <ReportDashboard route={route} />;
};

const HANDOFF_KEYS = [
  "planId",
  "targetJobId",
  "jdId",
  "resumeVersionId",
  "roundId",
  "sessionId",
  "reportId",
] as const;

function rebuildHandoffParams(
  params: Record<string, string>,
): Record<string, string> {
  const next: Record<string, string> = {};
  for (const key of HANDOFF_KEYS) {
    if (params[key]) next[key] = params[key]!;
  }
  return next;
}
