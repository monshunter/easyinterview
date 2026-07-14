import { useRef, type FC } from "react";

import type { FeedbackReport } from "../../../api/generated/types";
import { useI18n } from "../../i18n/messages";
import { useNavigation } from "../../navigation/NavigationProvider";
import type { Route } from "../../routes";
import { resolveReportBackRoute } from "../report/reportBackRoute";
import { GeneratingErrorState, type GeneratingErrorKind } from "./components/GeneratingErrorState";
import { HeaderHero } from "./components/HeaderHero";
import { useReportGenerationPoll } from "./hooks/useReportGenerationPoll";

interface GeneratingScreenProps {
  route: Route;
}

/** Honest projection of the report resource. No client-simulated progress. */
export const GeneratingScreen: FC<GeneratingScreenProps> = ({ route }) => {
  const { t } = useI18n();
  const { navigate } = useNavigation();
  const reportId = route.params.reportId ?? "";
  const handoffNavigatedRef = useRef(false);

  const handleReady = (report: FeedbackReport) => {
    if (handoffNavigatedRef.current) return;
    if (resolveReportBackRoute(report, reportId).name !== "reports") return;
    handoffNavigatedRef.current = true;
    navigate({ name: "report", params: { reportId: report.id } });
  };

  const poll = useReportGenerationPoll({ reportId, onReady: handleReady });
  const backRoute = resolveReportBackRoute(poll.report, reportId);
  const backDestination = backRoute.name === "reports" ? "reports" : "workspace";
  const goBack = () => navigate(backRoute);

  if (!reportId) {
    return <GeneratingErrorState kind="missingReportId" onBack={goBack} backDestination={backDestination} />;
  }

  if (poll.state === "timeout") {
    return <GeneratingErrorState kind="timeout" onRetry={poll.retry} onBack={goBack} backDestination={backDestination} />;
  }

  if (poll.state === "invalid") {
    return <GeneratingErrorState kind="invalidReport" onBack={goBack} backDestination={backDestination} />;
  }

  if (poll.state === "failed") {
    return (
      <GeneratingErrorState
        kind={failureKind(poll.errorCode)}
        onBack={goBack}
        backDestination={backDestination}
      />
    );
  }

  if (poll.state === "error") {
    return <GeneratingErrorState kind="loadFailed" onRetry={poll.retry} onBack={goBack} backDestination={backDestination} />;
  }

  if (poll.state === "ready") {
    return <GeneratingErrorState kind="invalidReport" onBack={goBack} backDestination={backDestination} />;
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
        <div
          style={{
            marginTop: 28,
            paddingTop: 16,
            borderTop: "1px solid var(--ei-color-rule-strong)",
            display: "flex",
            gap: 10,
            flexWrap: "wrap",
          }}
        >
          <span data-testid="generating-back-button">
            <button
              type="button"
              onClick={goBack}
              style={{
                display: "inline-flex",
                alignItems: "center",
                justifyContent: "center",
                gap: 8,
                height: 30,
                padding: "0 12px",
                fontSize: 13,
                fontWeight: 500,
                background: "var(--ei-color-bg-canvas)",
                color: "var(--ei-color-fg-primary)",
                border: "1px solid var(--ei-color-rule-strong)",
                borderRadius: 2,
                cursor: "pointer",
                fontFamily: "var(--ei-font-sans)",
                letterSpacing: "-0.005em",
                transition: "transform .08s ease, opacity .15s",
              }}
            >
              {t(backDestination === "reports" ? "generating.errors.backToReports" : "generating.errors.backToWorkspace")}
            </button>
          </span>
        </div>
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
