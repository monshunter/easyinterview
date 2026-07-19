import { useRef, type FC } from "react";

import type { FeedbackReport } from "../../../api/generated/types";
import { useI18n } from "../../i18n/messages";
import { useNavigation } from "../../navigation/NavigationProvider";
import type { Route } from "../../routes";
import { AsyncTransitionScene } from "../../transition/AsyncTransitionScene";
import { resolveReportBackRoute } from "../report/reportBackRoute";
import { GeneratingErrorState, type GeneratingErrorKind } from "./components/GeneratingErrorState";
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
    <AsyncTransitionScene
      variant="report"
      testId="generating-screen"
      data-report-status={status}
      className="ei-fadein"
      card
      contentTestId="generating-transition-card"
      eyebrowTestId="generating-header-eyebrow"
      titleTestId="generating-header-title"
      bodyTestId="generating-header-subtitle"
      eyebrow={t(
        status === "queued"
          ? "generating.status.queued"
          : "generating.status.generating",
      )}
      title={t("generating.header.title")}
      body={t("generating.header.subtitle")}
      showProgress
      action={{
        label: t(
          backDestination === "reports"
            ? "generating.errors.backToReports"
            : "generating.errors.backToWorkspace",
        ),
        onClick: goBack,
        wrapperTestId: "generating-back-button",
      }}
    />
  );
};

function failureKind(code: string | null): GeneratingErrorKind {
  if (code === "REPORT_CONTEXT_TOO_LARGE") return "contextTooLarge";
  if (code === "REPORT_NOT_FOUND") return "notFound";
  if (code === "AI_OUTPUT_INVALID") return "invalidReport";
  return "failed";
}
