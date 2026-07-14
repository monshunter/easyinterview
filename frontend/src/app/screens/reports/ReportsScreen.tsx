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
    width="15"
    height="15"
    viewBox="0 0 24 24"
    fill="none"
    stroke="currentColor"
    strokeWidth="1.5"
    strokeLinecap="round"
    strokeLinejoin="round"
    style={{
      flexShrink: 0,
      display: "inline-block",
      verticalAlign: "middle",
    }}
  >
    <path d="M5 12h14M13 6l6 6-6 6" />
  </svg>
);

const SecondaryButton: FC<{
  onClick: () => void;
  size?: "sm" | "md";
  icon?: boolean;
  children: ReactNode;
}> = ({ onClick, size = "md", icon = false, children }) => {
  const compact = size === "sm";
  return (
    <button
      type="button"
      onClick={onClick}
      onMouseDown={(event) => {
        event.currentTarget.style.transform = "translateY(0.5px)";
      }}
      onMouseUp={(event) => {
        event.currentTarget.style.transform = "";
      }}
      onMouseLeave={(event) => {
        event.currentTarget.style.transform = "";
      }}
      style={{
        display: "inline-flex",
        alignItems: "center",
        justifyContent: "center",
        gap: 8,
        height: compact ? 30 : 38,
        padding: compact ? "0 12px" : "0 16px",
        fontSize: compact ? 13 : 14,
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
      {icon ? <ArrowRightIcon /> : null}
      {children}
    </button>
  );
};

const Card: FC<{ pad?: number; children: ReactNode }> = ({
  pad = 20,
  children,
}) => (
  <div
    style={{
      background: "var(--ei-color-bg-card)",
      border: "1px solid var(--ei-color-rule-strong)",
      borderRadius: 3,
      padding: pad,
      cursor: "default",
      transition: "border-color .15s, transform .15s",
    }}
  >
    {children}
  </div>
);

export const ReportsScreen: FC<ReportsScreenProps> = ({ route }) => {
  const { lang, t } = useI18n();
  const { navigate, replaceRoute } = useNavigation();
  const runtime = useAppRuntimeOptional();
  const targetJobId = routeTargetJobId(route);
  const requestSequence = useRef(0);
  const [state, setState] = useState<ReportsState>(() => ({
    status: "loading",
    ownerTargetJobId: targetJobId,
  }));

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
  }, [runtime?.auth.status, runtime?.client, targetJobId]);

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
      ? navigate({ name: "parse", params: { targetJobId } })
      : navigate({ name: "workspace", params: {} });

  return (
    <main
      className="ei-fadein"
      data-testid="reports-screen"
      style={{
        maxWidth: 1120,
        margin: "0 auto",
        padding: "32px clamp(16px, 5vw, 48px) 96px",
      }}
    >
      <button
        type="button"
        data-testid="reports-back-button"
        onClick={goBack}
        style={{
          border: 0,
          background: "transparent",
          color: "var(--ei-color-fg-tertiary)",
          cursor: "pointer",
          marginBottom: 20,
          padding: 0,
        }}
      >
        ← {targetJobId ? t("reports.backToPlan") : t("reports.backToWorkspace")}
      </button>

      <header
        style={{
          display: "flex",
          justifyContent: "space-between",
          alignItems: "flex-end",
          gap: 20,
          flexWrap: "wrap",
          marginBottom: 24,
        }}
      >
        <div style={{ minWidth: 0, flex: "1 1 520px" }}>
          <div
            className="ei-label"
            style={{
              color: "var(--ei-color-fg-tertiary)",
              marginBottom: 8,
            }}
          >
            {t("reports.eyebrow")}
          </div>
          <h1
            data-testid="reports-target-title"
            className="ei-serif"
            style={{
              margin: 0,
              fontSize: 36,
              color: "var(--ei-color-fg-primary)",
              lineHeight: 1.2,
              overflowWrap: "anywhere",
            }}
          >
            {targetJob
              ? `${targetJob.companyName} · ${targetJob.title}`
              : t("reports.currentPlanTitle")}
          </h1>
          <p
            style={{
              margin: "10px 0 0",
              color: "var(--ei-color-fg-secondary)",
              fontSize: 14,
              lineHeight: 1.65,
            }}
          >
            {t("reports.subtitle")}
          </p>
        </div>
      </header>

      {renderedState.status === "loading" ? (
        <Card>
          <div
            data-testid="reports-loading"
            role="status"
            style={{
              color: "var(--ei-color-fg-tertiary)",
              fontSize: 13,
              lineHeight: 1.65,
            }}
          >
            {t("reports.loading")}
          </div>
        </Card>
      ) : renderedState.status === "error" ? (
        <Card>
          <div data-testid="reports-error" role="alert">
            <div
              className="ei-label"
              style={{
                color: "var(--ei-color-danger)",
                marginBottom: 10,
              }}
            >
              {t("reports.errorEyebrow")}
            </div>
            <div
              className="ei-serif"
              style={{
                color: "var(--ei-color-fg-primary)",
                fontSize: 24,
                marginBottom: 10,
              }}
            >
              {t("reports.errorTitle")}
            </div>
            <p
              style={{
                color: "var(--ei-color-fg-secondary)",
                fontSize: 13,
                lineHeight: 1.65,
                margin: "0 0 18px",
              }}
            >
              {t("reports.errorDescription")}
            </p>
            <SecondaryButton onClick={goBack}>
              {targetJobId ? t("reports.returnToPlan") : t("reports.backToWorkspace")}
            </SecondaryButton>
          </div>
        </Card>
      ) : isEmpty ? (
        <Card>
          <div data-testid="reports-empty" role="status">
            <div
              className="ei-label"
              style={{
                color: "var(--ei-color-fg-tertiary)",
                marginBottom: 10,
              }}
            >
              {t("reports.emptyEyebrow")}
            </div>
            <div
              className="ei-serif"
              style={{
                color: "var(--ei-color-fg-primary)",
                fontSize: 24,
                marginBottom: 10,
              }}
            >
              {t("reports.emptyTitle")}
            </div>
            <p
              style={{
                color: "var(--ei-color-fg-secondary)",
                fontSize: 13,
                lineHeight: 1.65,
                margin: 0,
              }}
            >
              {t("reports.emptyDescription")}
            </p>
          </div>
        </Card>
      ) : (
        <Card pad={0}>
          <div data-testid="reports-list">
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

              return (
                <div
                  key={item.displayRound.id}
                  data-testid={`reports-round-${item.displayRound.sequence}`}
                  style={{
                    padding: "18px 22px",
                    borderBottom:
                      index < renderedState.overview.rounds.length - 1
                        ? "1px dotted var(--ei-color-rule-strong)"
                        : "none",
                    display: "flex",
                    justifyContent: "space-between",
                    alignItems: "center",
                    gap: 16,
                    flexWrap: "wrap",
                  }}
                >
                  <div style={{ minWidth: 0, flex: "1 1 320px" }}>
                    <div
                      style={{
                        color: "var(--ei-color-fg-primary)",
                        fontSize: 14,
                        fontWeight: 500,
                        overflowWrap: "anywhere",
                      }}
                    >
                      {item.displayRound.name}
                    </div>
                    <div
                      style={{
                        color: "var(--ei-color-fg-tertiary)",
                        fontSize: 11.5,
                        lineHeight: 1.55,
                        marginTop: 5,
                      }}
                    >
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
                  <div style={{ display: "flex", gap: 8, flexWrap: "wrap" }}>
                    {item.currentReport ? (
                      <span data-testid="reports-current">
                        <SecondaryButton
                          size="sm"
                          icon
                          onClick={() =>
                            navigate({
                              name: "report",
                              params: { reportId: item.currentReport!.id },
                            })
                          }
                        >
                          {t("reports.openCurrent")}
                        </SecondaryButton>
                      </span>
                    ) : null}
                    {showGenerating && item.latestAttempt ? (
                      <span data-testid="reports-generating">
                        <SecondaryButton
                          size="sm"
                          icon
                          onClick={() =>
                            navigate({
                              name: "generating",
                              params: { reportId: item.latestAttempt!.id },
                            })
                          }
                        >
                          {t("reports.viewGeneration")}
                        </SecondaryButton>
                      </span>
                    ) : null}
                  </div>
                </div>
              );
            })}
          </div>
        </Card>
      )}
    </main>
  );
};
