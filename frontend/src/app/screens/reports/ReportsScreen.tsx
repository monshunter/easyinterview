import { useEffect, useRef, useState, type FC, type ReactNode } from "react";

import type { TargetJob } from "../../../api/generated/types";
import {
  validateTargetJobReportsOverview,
  type ValidatedTargetJobReportsOverview,
} from "../../../api/targetJobReports";
import {
  buildTargetJobRoundAssumptions,
  type TargetJobRoundAssumption,
} from "../../interview-context/roundAssumptions";
import { useI18n } from "../../i18n/messages";
import { useNavigation } from "../../navigation/NavigationProvider";
import type { Route } from "../../routes";
import { useAppRuntimeOptional } from "../../runtime/AppRuntimeProvider";
import { ReportPageIllustration } from "./ReportPageIllustration";
import { useReportRegeneration } from "./hooks/useReportRegeneration";

interface ReportsScreenProps {
  route: Route;
}

interface LoadingState {
  status: "loading";
  ownerTargetJobId: string;
  targetJob?: TargetJob;
}

interface ErrorState {
  status: "error";
  ownerTargetJobId: string;
  targetJob?: TargetJob;
}

interface ReadyState {
  status: "ready";
  ownerTargetJobId: string;
  targetJob: TargetJob;
  overview: ValidatedTargetJobReportsOverview<TargetJobRoundAssumption>;
}

type ReportsState = LoadingState | ErrorState | ReadyState;

const UUID =
  /^[0-9a-f]{8}-[0-9a-f]{4}-[1-8][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/i;

function routeTargetJobId(route: Route): string {
  if (route.name !== "reports") return "";
  const candidate = route.params.targetJobId?.trim() ?? "";
  return UUID.test(candidate) ? candidate : "";
}

function isExpectedTargetJob(
  value: TargetJob,
  expectedTargetJobId: string,
): value is TargetJob {
  return Boolean(
    value &&
      value.id === expectedTargetJobId &&
      typeof value.title === "string" &&
      value.title.trim() &&
      typeof value.companyName === "string" &&
      value.companyName.trim(),
  );
}

function reportDate(value: string, lang: "en" | "zh"): string {
  return new Intl.DateTimeFormat(lang === "zh" ? "zh-CN" : "en-US", {
    month: "short",
    day: "numeric",
    timeZone: "UTC",
  }).format(new Date(value));
}

const ArrowRightIcon: FC = () => (
  <svg
    aria-hidden="true"
    className="ei-reports-action-icon"
    viewBox="0 0 24 24"
    fill="none"
    stroke="currentColor"
    strokeWidth="1.5"
    strokeLinecap="round"
    strokeLinejoin="round"
  >
    <path d="M5 12h14M13 6l6 6-6 6" />
  </svg>
);

const RecordIcon: FC = () => (
  <svg
    aria-hidden="true"
    className="ei-reports-action-icon"
    viewBox="0 0 24 24"
    fill="none"
    stroke="currentColor"
    strokeWidth="1.6"
    strokeLinecap="round"
    strokeLinejoin="round"
  >
    <path d="M6 3h9l3 3v15H6zM15 3v4h3M9 11h6M9 15h6" />
  </svg>
);

const TargetIcon: FC = () => (
  <svg
    aria-hidden="true"
    viewBox="0 0 24 24"
    width="24"
    height="24"
    fill="none"
    stroke="currentColor"
    strokeWidth="1.7"
    strokeLinecap="round"
    strokeLinejoin="round"
  >
    <path d="M4 21V7l8-4 8 4v14M2 21h20M8 10h2M14 10h2M8 14h2M14 14h2M10 21v-3h4v3" />
  </svg>
);

const ReportAction: FC<{
  onClick: () => void;
  variant?: "primary" | "secondary" | "quiet";
  icon?: "arrow" | "record";
  ariaBusy?: boolean;
  ariaLabel?: string;
  disabled?: boolean;
  testId?: string;
  children: ReactNode;
}> = ({
  onClick,
  variant = "secondary",
  icon,
  ariaBusy,
  ariaLabel,
  disabled = false,
  testId,
  children,
}) => {
  return (
    <button
      type="button"
      onClick={onClick}
      aria-busy={ariaBusy}
      aria-label={ariaLabel}
      data-testid={testId}
      disabled={disabled}
      className={`ei-reports-action ei-reports-action-${variant}`}
    >
      {icon === "record" ? <RecordIcon /> : null}
      {children}
      {icon === "arrow" ? <ArrowRightIcon /> : null}
    </button>
  );
};

const StateCard: FC<{ children: ReactNode }> = ({ children }) => (
  <section className="ei-reports-state-card">{children}</section>
);

export const ReportsScreen: FC<ReportsScreenProps> = ({ route }) => {
  const { lang, t } = useI18n();
  const { navigate, replaceRoute } = useNavigation();
  const runtime = useAppRuntimeOptional();
  const targetJobId = routeTargetJobId(route);
  const requestSequence = useRef(0);
  const [refreshNonce, setRefreshNonce] = useState(0);
  const [state, setState] = useState<ReportsState>(() => ({
    status: "loading",
    ownerTargetJobId: targetJobId,
  }));
  const regeneration = useReportRegeneration({
    client: runtime?.client ?? null,
    targetJobId,
    onAccepted: (reportId) =>
      navigate({ name: "generating", params: { reportId } }),
    onStaleState: () => setRefreshNonce((current) => current + 1),
  });

  useEffect(() => {
    if (targetJobId) return;
    replaceRoute({ name: "workspace", params: {} });
  }, [replaceRoute, targetJobId]);

  useEffect(() => {
    const requestId = requestSequence.current + 1;
    requestSequence.current = requestId;

    if (!targetJobId) return;
    if (!runtime || runtime.auth.status === "loading") {
      setState({ status: "loading", ownerTargetJobId: targetJobId });
      return;
    }
    if (runtime.auth.status !== "authenticated") {
      setState({ status: "error", ownerTargetJobId: targetJobId });
      return;
    }

    let active = true;
    let trustedTargetJob: TargetJob | undefined;
    setState({ status: "loading", ownerTargetJobId: targetJobId });

    const targetPromise = runtime.client.getTargetJob(targetJobId).then((targetJob) => {
      if (!isExpectedTargetJob(targetJob, targetJobId)) {
        throw new Error("Invalid TargetJob reports owner");
      }
      trustedTargetJob = targetJob;
      const rounds = buildTargetJobRoundAssumptions(targetJob);
      if (rounds.length === 0) {
        throw new Error("Invalid TargetJob round catalog");
      }
      if (active && requestSequence.current === requestId) {
        setState((current) => {
          if (current.ownerTargetJobId !== targetJobId) return current;
          if (current.status === "error") {
            return { ...current, targetJob };
          }
          if (current.status === "loading") {
            return {
              status: "loading",
              ownerTargetJobId: targetJobId,
              targetJob,
            };
          }
          return current;
        });
      }
      return { targetJob, rounds };
    });

    Promise.all([
      targetPromise,
      runtime.client.listTargetJobReports(targetJobId),
    ])
      .then(([{ targetJob, rounds }, rawOverview]) => {
        const overview = validateTargetJobReportsOverview(
          rawOverview,
          targetJobId,
          rounds,
        );
        if (!active || requestSequence.current !== requestId) return;
        setState({
          status: "ready",
          ownerTargetJobId: targetJobId,
          targetJob,
          overview,
        });
      })
      .catch(() => {
        if (!active || requestSequence.current !== requestId) return;
        setState({
          status: "error",
          ownerTargetJobId: targetJobId,
          targetJob: trustedTargetJob,
        });
      });

    return () => {
      active = false;
    };
  }, [refreshNonce, runtime?.auth.status, runtime?.client, targetJobId]);

  const renderedState: ReportsState =
    targetJobId && state.ownerTargetJobId === targetJobId
      ? state
      : { status: "loading", ownerTargetJobId: targetJobId };
  const targetJob = renderedState.targetJob;
  const isEmpty =
    renderedState.status === "ready" &&
    renderedState.overview.rounds.every(
      (round) => !round.currentReport && !round.latestAttempt,
    );
  const goBack = () =>
    targetJobId
      ? navigate({ name: "workspace", params: { targetJobId } })
      : navigate({ name: "workspace", params: {} });

  return (
    <main
      className="ei-fadein ei-reports-screen"
      data-testid="reports-screen"
    >
      <button
        type="button"
        data-testid="reports-back-button"
        onClick={goBack}
        className="ei-reports-back"
      >
        ← {targetJobId ? t("reports.backToPlan") : t("reports.backToWorkspace")}
      </button>

      <header className="ei-report-records-header">
        <div className="ei-report-records-header-copy">
          <div className="ei-report-records-eyebrow">
            {t("reports.eyebrow")}
          </div>
          <h1
            data-testid="reports-target-title"
            className="ei-report-records-title"
          >
            {targetJob
              ? `${targetJob.companyName} · ${targetJob.title}`
              : t("reports.currentPlanTitle")}
          </h1>
          <p className="ei-report-records-subtitle">{t("reports.subtitle")}</p>
        </div>
        <ReportPageIllustration testId="reports-header-illustration" />
      </header>

      {renderedState.status === "loading" ? (
        <StateCard>
          <div
            data-testid="reports-loading"
            role="status"
            className="ei-reports-state-copy"
          >
            {t("reports.loading")}
          </div>
        </StateCard>
      ) : renderedState.status === "error" ? (
        <StateCard>
          <div data-testid="reports-error" role="alert">
            <div className="ei-reports-state-eyebrow ei-reports-state-eyebrow-error">
              {t("reports.errorEyebrow")}
            </div>
            <div className="ei-reports-state-title">
              {t("reports.errorTitle")}
            </div>
            <p className="ei-reports-state-description">
              {t("reports.errorDescription")}
            </p>
            <ReportAction onClick={goBack}>
              {targetJobId ? t("reports.returnToPlan") : t("reports.backToWorkspace")}
            </ReportAction>
          </div>
        </StateCard>
      ) : isEmpty ? (
        <StateCard>
          <div data-testid="reports-empty" role="status">
            <div className="ei-reports-state-eyebrow">
              {t("reports.emptyEyebrow")}
            </div>
            <div className="ei-reports-state-title">
              {t("reports.emptyTitle")}
            </div>
            <p className="ei-reports-state-description">
              {t("reports.emptyDescription")}
            </p>
          </div>
        </StateCard>
      ) : (
        <>
          <section
            className="ei-reports-target-summary"
            data-testid="reports-target-summary"
          >
            <span className="ei-reports-target-summary-icon">
              <TargetIcon />
            </span>
            <div className="ei-reports-target-summary-copy">
              <div className="ei-reports-target-summary-label">
                {t("reports.eyebrow")}
              </div>
              <h2>{targetJob!.title}</h2>
              <div className="ei-reports-target-summary-meta">
                <span>{lang === "zh" ? "公司" : "Company"}：{targetJob!.companyName}</span>
                {targetJob!.locationText ? (
                  <span>{lang === "zh" ? "地点" : "Location"}：{targetJob!.locationText}</span>
                ) : null}
                <span>{lang === "zh" ? "轮次" : "Rounds"}：{renderedState.overview.rounds.length}</span>
                <span>{lang === "zh" ? "面试日期" : "Interview date"}：{reportDate(targetJob!.createdAt, lang)}</span>
              </div>
            </div>
          </section>

          <div
            className="ei-reports-timeline"
            data-testid="reports-timeline"
            data-timeline="true"
          >
            <div className="ei-reports-timeline-list" data-testid="reports-list">
            {renderedState.overview.rounds.map((item, index) => {
              const latestStatus = item.latestAttempt?.status;
              const latestIsDifferent = Boolean(
                item.latestAttempt &&
                  item.latestAttempt.id !== item.currentReport?.id,
              );
              const showGenerating = Boolean(
                latestIsDifferent &&
                  (latestStatus === "queued" || latestStatus === "generating"),
              );
              const showFailed = latestIsDifferent && latestStatus === "failed";
              const showLatestReady =
                latestIsDifferent && latestStatus === "ready";
              const failedAttempt = showFailed ? item.latestAttempt : null;
              const latestConversationAttempt = latestIsDifferent
                ? item.latestAttempt
                : null;
              const regenerationState = failedAttempt
                ? regeneration.stateFor(failedAttempt.id)
                : null;
              const canRegenerate = Boolean(
                failedAttempt &&
                  failedAttempt.errorCode !== "REPORT_CONTEXT_TOO_LARGE",
              );

              return (
                <article
                  key={item.displayRound.id}
                  data-testid={`reports-round-${item.displayRound.sequence}`}
                  className="ei-reports-round-card"
                >
                  <div className="ei-reports-round-marker" aria-hidden="true">
                    <span
                      className="ei-reports-round-index"
                      data-testid={`reports-round-index-${item.displayRound.sequence}`}
                    >
                      {String(item.displayRound.sequence).padStart(2, "0")}
                    </span>
                    {index < renderedState.overview.rounds.length - 1 ? (
                      <span className="ei-reports-round-line" />
                    ) : null}
                  </div>
                  <div className="ei-reports-round-copy">
                    <div className="ei-reports-round-heading">
                      <h2>{item.displayRound.name}</h2>
                      {item.currentReport ? (
                        <span className="ei-reports-round-status">
                          {t("reports.currentReport")}
                        </span>
                      ) : null}
                    </div>
                    <div className="ei-reports-round-meta">
                      {item.currentReport ? (
                        <>
                          {t("reports.currentReport")} · {reportDate(item.currentReport.generatedAt, lang)}
                        </>
                      ) : (
                        t("reports.roundEmpty")
                      )}
                      {showGenerating ? (
                        <span> · {t("reports.generating")}</span>
                      ) : null}
                      {showFailed ? (
                        <span data-testid="reports-failed">
                          {" · "}{t("reports.failed")}
                        </span>
                      ) : null}
                      {showLatestReady ? (
                        <span data-testid="reports-latest-ready">
                          {" · "}{t("reports.latestReady")}
                        </span>
                      ) : null}
                    </div>
                  </div>
                  <div className="ei-reports-round-actions">
                    <div className="ei-reports-round-action-row">
                      {item.currentReport ? (
                        <span data-testid="reports-current" className="ei-reports-current-actions">
                          <ReportAction
                            variant="primary"
                            icon="arrow"
                            ariaLabel={t("reports.openCurrentA11y")}
                            onClick={() =>
                              navigate({
                                name: "report",
                                params: { reportId: item.currentReport!.id },
                              })
                            }
                          >
                            {t("reports.openCurrent")}
                          </ReportAction>
                          <ReportAction
                            ariaLabel={t("reports.viewCurrentConversationA11y")}
                            testId="reports-conversation-entry"
                            icon="record"
                            onClick={() =>
                              navigate({
                                name: "report_conversation",
                                params: { reportId: item.currentReport!.id },
                              })
                            }
                          >
                            {t("reports.viewConversation")}
                          </ReportAction>
                        </span>
                      ) : null}
                      {showGenerating && item.latestAttempt ? (
                        <span data-testid="reports-generating">
                          <ReportAction
                            variant="quiet"
                            icon="arrow"
                            onClick={() =>
                              navigate({
                                name: "generating",
                                params: { reportId: item.latestAttempt!.id },
                              })
                            }
                          >
                            {t("reports.viewGeneration")}
                          </ReportAction>
                        </span>
                      ) : null}
                      {failedAttempt ? (
                        <span className="ei-reports-failed-actions">
                          {canRegenerate ? (
                            <ReportAction
                              variant="quiet"
                              ariaBusy={regenerationState?.pending || undefined}
                              ariaLabel={t("reports.regenerateFailedA11y")}
                              disabled={regenerationState?.pending ?? false}
                              testId="reports-failed-regenerate"
                              onClick={() => void regeneration.regenerate(failedAttempt.id)}
                            >
                              {regenerationState?.pending
                                ? t("reports.regenerateFailedPending")
                                : t("reports.regenerateFailed")}
                            </ReportAction>
                          ) : null}
                        </span>
                      ) : null}
                      {latestConversationAttempt ? (
                        <ReportAction
                          ariaLabel={t("reports.viewLatestConversationA11y")}
                          testId="reports-latest-conversation-entry"
                          icon="record"
                          onClick={() =>
                            navigate({
                              name: "report_conversation",
                              params: { reportId: latestConversationAttempt.id },
                            })
                          }
                        >
                          {t("reports.viewConversation")}
                        </ReportAction>
                      ) : null}
                    </div>
                    {failedAttempt && regenerationState?.error ? (
                      <div
                        data-testid="reports-regenerate-error"
                        role="alert"
                        className="ei-reports-regenerate-error"
                      >
                        {t("reports.regenerateError")}
                      </div>
                    ) : null}
                  </div>
                </article>
              );
            })}
            </div>
          </div>
        </>
      )}
    </main>
  );
};
