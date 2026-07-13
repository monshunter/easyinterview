import { useRef, type FC } from "react";

import type { FeedbackReport } from "../../../api/generated/types";
import { useNavigation } from "../../navigation/NavigationProvider";
import type { Route } from "../../routes";
import { GeneratingErrorState, type GeneratingErrorKind } from "./components/GeneratingErrorState";
import { HeaderHero } from "./components/HeaderHero";
import { useReportGenerationPoll } from "./hooks/useReportGenerationPoll";

interface GeneratingScreenProps {
  route: Route;
}

/** Honest projection of the report resource. No client-simulated progress. */
export const GeneratingScreen: FC<GeneratingScreenProps> = ({ route }) => {
  const { navigate } = useNavigation();
  const reportId = route.params.reportId ?? "";
  const handoffNavigatedRef = useRef(false);

  const handleReady = (report: FeedbackReport) => {
    if (handoffNavigatedRef.current) return;
    handoffNavigatedRef.current = true;
    const stableReportId = report.id === reportId ? report.id : reportId;
    navigate({ name: "report", params: { reportId: stableReportId } });
  };

  const poll = useReportGenerationPoll({ reportId, onReady: handleReady });
  const goWorkspace = () => navigate({ name: "workspace", params: {} });

  if (!reportId) {
    return <GeneratingErrorState kind="missingReportId" onBackToWorkspace={goWorkspace} />;
  }

  if (poll.state === "timeout") {
    return <GeneratingErrorState kind="timeout" onRetry={poll.retry} onBackToWorkspace={goWorkspace} />;
  }

  if (poll.state === "failed") {
    return (
      <GeneratingErrorState
        kind={failureKind(poll.errorCode)}
        onBackToWorkspace={goWorkspace}
      />
    );
  }

  if (poll.state === "error") {
    return <GeneratingErrorState kind="loadFailed" onRetry={poll.retry} onBackToWorkspace={goWorkspace} />;
  }

  const status = poll.report?.status === "generating" ? "generating" : "queued";
  return (
    <div
      data-testid="generating-screen"
      data-report-status={status}
      aria-live="polite"
      className="ei-fadein"
      style={{
        minHeight: "100vh",
        background: "var(--ei-color-bg-canvas)",
        display: "flex",
        alignItems: "center",
        justifyContent: "center",
        padding: "48px clamp(16px, 6vw, 48px)",
      }}
    >
      <div style={{ maxWidth: 780, width: "100%" }}>
        <HeaderHero status={status} />
      </div>
    </div>
  );
};

function failureKind(code: string | null): GeneratingErrorKind {
  if (code === "REPORT_CONTEXT_TOO_LARGE") return "contextTooLarge";
  if (code === "REPORT_NOT_FOUND") return "notFound";
  if (code === "AI_OUTPUT_INVALID") return "invalidReport";
  return "failed";
}
