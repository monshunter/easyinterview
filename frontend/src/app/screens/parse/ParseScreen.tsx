import {
  useState,
  useEffect,
  useCallback,
  useMemo,
  useRef,
  type FC,
} from "react";

import { useAppRuntimeOptional } from "../../runtime/AppRuntimeProvider";
import { useRequestAuth } from "../../auth/useRequestAuth";
import { useI18n, type MessageKey } from "../../i18n/messages";
import {
  buildTargetJobRoundAssumptions,
  resolveTargetJobPracticeProgress,
} from "../../interview-context/roundAssumptions";
import { startPracticeFromParams } from "../../interview-context/startPractice";
import { PracticeLaunchTransition } from "../../interview-context/PracticeLaunchTransition";
import { useNavigation } from "../../navigation/NavigationProvider";
import { targetJobPracticeRouteParams } from "../../navigation/interviewContext";
import type { Route } from "../../routes";
import type { TargetJob } from "../../../api/generated/types";

type Stage = "loading" | "preview" | "error" | "failed";

interface ParseScreenProps {
  route: Route;
  _mockStage?: Stage;
  _mockTargetJob?: TargetJob;
}

type HitState = true | "partial" | false;

const loadingStepKeys = [
  "parse.loadingStep1",
  "parse.loadingStep2",
  "parse.loadingStep3",
  "parse.loadingStep4",
] as const;

const loadingStepTicks = [600, 900, 800, 700] as const;

function buildInterviewParams(
  job: TargetJob,
  resumeId: string,
): Record<string, string> {
  return targetJobPracticeRouteParams(job, { resumeId });
}

function hitStateFromEvidence(level: string | undefined): HitState {
  if (level === "explicit") return true;
  if (level === "implicit" || level === "inferred") return "partial";
  return false;
}

function safeScrollToTop(): void {
  if (navigator.userAgent.toLowerCase().includes("jsdom")) return;
  try {
    window.scrollTo({ top: 0, behavior: "smooth" });
  } catch {
    // jsdom exposes scrollTo but still throws inside the test environment.
  }
}

function useParseCompactLayout(): boolean {
  const [compact, setCompact] = useState(() => {
    if (typeof window === "undefined") return false;
    if (typeof window.matchMedia !== "function") return false;
    return window.matchMedia("(max-width: 720px)").matches;
  });

  useEffect(() => {
    if (typeof window === "undefined") return;
    if (typeof window.matchMedia !== "function") return;
    const query = window.matchMedia("(max-width: 720px)");
    const update = () => setCompact(query.matches);
    update();
    query.addEventListener("change", update);
    return () => query.removeEventListener("change", update);
  }, []);

  return compact;
}

const PlanSectionIcon: FC<{ variant: "basics" | "must" | "nice" | "hidden" | "rounds" }> = ({ variant }) => (
  <span className={`ei-plan-detail-section-icon ei-plan-detail-section-icon--${variant}`} aria-hidden="true">
    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round">
      {variant === "basics" ? <><path d="M5 7h14v12H5z" /><path d="M9 7V4h6v3M8 11h8M8 15h5" /></> : null}
      {variant === "must" ? <><circle cx="12" cy="12" r="9" /><path d="m8 12 2.5 2.5L16 9" /></> : null}
      {variant === "nice" ? <><path d="m12 3 2.7 5.5 6.1.9-4.4 4.3 1 6.1-5.4-2.9-5.4 2.9 1-6.1-4.4-4.3 6.1-.9L12 3Z" /></> : null}
      {variant === "hidden" ? <><path d="M3 12s3.4-6 9-6 9 6 9 6-3.4 6-9 6-9-6-9-6Z" /><circle cx="12" cy="12" r="2.5" /></> : null}
      {variant === "rounds" ? <><circle cx="12" cy="12" r="9" /><path d="M12 7v5l3 2" /></> : null}
    </svg>
  </span>
);

export const ParseScreen: FC<ParseScreenProps> = ({
  route,
  _mockStage,
  _mockTargetJob,
}) => {
  const { t, lang } = useI18n();
  const { navigate, replaceRoute } = useNavigation();
  const runtime = useAppRuntimeOptional();
  const requestAuth = useRequestAuth();
  const [stage, setStage] = useState<Stage>(_mockStage ?? "loading");
  const [step, setStep] = useState(0);
  const [loadedTargetJob, setLoadedTargetJob] = useState<TargetJob | null>(null);
  const targetJob = _mockTargetJob ?? loadedTargetJob;
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const [confirmError, setConfirmError] = useState<MessageKey | null>(null);
  const [confirming, setConfirming] = useState(false);
  const [pollNonce, setPollNonce] = useState(0);
  const pollingRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const loadedTargetJobRef = useRef<TargetJob | null>(null);

  const steps = loadingStepKeys;
  const targetJobId =
    typeof route.params?.targetJobId === "string"
      ? route.params.targetJobId
      : undefined;
  const isWorkspaceDetail = route.name === "workspace";
  const compactLayout = useParseCompactLayout();
  const routeTestId = isWorkspaceDetail ? "route-workspace" : "route-parse";
  const rounds = useMemo(
    () => buildTargetJobRoundAssumptions(targetJob),
    [targetJob],
  );
  const hydrateReadyJob = useCallback((job: TargetJob) => {
    loadedTargetJobRef.current = job;
    setLoadedTargetJob(job);
  }, []);

  useEffect(() => {
    if (_mockStage || _mockTargetJob) return;

    if (pollingRef.current) {
      clearTimeout(pollingRef.current);
      pollingRef.current = null;
    }
    loadedTargetJobRef.current = null;
    setLoadedTargetJob(null);
    setErrorMessage(null);
    setConfirmError(null);
    setStep(0);
    setStage("loading");
  }, [targetJobId, _mockStage, _mockTargetJob]);

  useEffect(() => {
    if (stage !== "loading" || _mockStage) return;

    setStep(0);
    const timers: Array<ReturnType<typeof setTimeout>> = [];
    let elapsed = 0;
    loadingStepTicks.forEach((tick, i) => {
      elapsed += tick;
      timers.push(
        setTimeout(() => {
          setStep(i + 1);
        }, elapsed),
      );
    });
    return () => {
      timers.forEach((timer) => clearTimeout(timer));
    };
  }, [stage, _mockStage, pollNonce]);

  // Poll getTargetJob when in loading stage
  useEffect(() => {
    if (_mockStage || _mockTargetJob) return;
    const client = runtime?.client;
    if (!client) return;

    if (!targetJobId) {
      setStage("error");
      setErrorMessage(lang === "en" ? "Missing target job ID." : "缺少目标岗位 ID。");
      return;
    }

    if (
      isWorkspaceDetail &&
      loadedTargetJobRef.current?.id === targetJobId &&
      loadedTargetJobRef.current.analysisStatus === "ready"
    ) {
      setStage("preview");
      return;
    }

    let cancelled = false;

    const poll = async () => {
      if (cancelled) return;
      try {
        const job = await client.getTargetJob(targetJobId);
        if (cancelled) return;

        if (job.id !== targetJobId) {
          setStage("error");
          setErrorMessage(
            lang === "en"
              ? "The requested interview plan could not be verified."
              : "无法确认请求的面试规划。",
          );
          return;
        }

        if (job.analysisStatus === "ready") {
          hydrateReadyJob(job);
          if (isWorkspaceDetail) {
            setStage("preview");
          } else {
            replaceRoute({
              name: "workspace",
              params: { targetJobId: job.id },
            });
          }
        } else if (job.analysisStatus === "failed") {
          setStage("failed");
        } else if (isWorkspaceDetail) {
          setStage("error");
          setErrorMessage(
            lang === "en"
              ? "This interview plan is not ready yet."
              : "这份面试规划尚未准备完成。",
          );
        } else {
          // queued or processing — keep polling
          pollingRef.current = setTimeout(poll, 600);
        }
      } catch {
        if (!cancelled) {
          setStage("error");
          setErrorMessage(
            lang === "en"
              ? "Failed to fetch parse status."
              : "获取解析状态失败。",
          );
        }
      }
    };

    poll();

    return () => {
      cancelled = true;
      if (pollingRef.current) {
        clearTimeout(pollingRef.current);
        pollingRef.current = null;
      }
    };
  }, [
    runtime?.client,
    targetJobId,
    _mockStage,
    _mockTargetJob,
    pollNonce,
    isWorkspaceDetail,
    hydrateReadyJob,
    lang,
    replaceRoute,
  ]);

  useEffect(() => {
    if (_mockTargetJob) {
      hydrateReadyJob(_mockTargetJob);
    }
  }, [_mockTargetJob, hydrateReadyJob]);

  const handleCancel = useCallback(() => {
    navigate(isWorkspaceDetail ? { name: "workspace", params: {} } : { name: "home", params: {} });
  }, [isWorkspaceDetail, navigate]);

  const handleOpenReports = useCallback(() => {
    if (
      stage !== "preview" ||
      route.name !== "workspace" ||
      !targetJobId ||
      !targetJob ||
      targetJob.id !== targetJobId
    ) {
      return;
    }
    navigate({ name: "reports", params: { targetJobId: targetJob.id } });
  }, [navigate, route.name, stage, targetJob, targetJobId]);

  const handleReparse = useCallback(() => {
    if (pollingRef.current) {
      clearTimeout(pollingRef.current);
      pollingRef.current = null;
    }
    setErrorMessage(null);
    setStep(0);
    setStage("loading");
    setPollNonce((n) => n + 1);
    safeScrollToTop();
  }, []);

  const boundResumeId = targetJob?.resumeId?.trim() ?? "";

  const handleOpenBoundResume = useCallback(() => {
    if (
      stage !== "preview" ||
      route.name !== "workspace" ||
      !targetJobId ||
      !targetJob ||
      targetJob.id !== targetJobId ||
      !boundResumeId
    ) {
      return;
    }
    navigate({
      name: "resume_versions",
      params: { resumeId: boundResumeId },
    });
  }, [boundResumeId, navigate, route.name, stage, targetJob, targetJobId]);

  const handleStartInterview = useCallback(async () => {
    if (!targetJob || !boundResumeId || confirming || !runtime) return;
    const practiceParams = buildInterviewParams(targetJob, boundResumeId);

    if (runtime?.auth.status !== "authenticated") {
      requestAuth({
        type: "start_practice",
        label: t("parse.startInterview"),
        route: isWorkspaceDetail ? "workspace" : "parse",
        params: practiceParams,
      });
      return;
    }

    setConfirmError(null);
    setConfirming(true);
    try {
      const started = await startPracticeFromParams(
        runtime.client,
        practiceParams,
        lang,
      );
      navigate({
        name: "practice",
        params: started.params,
      });
    } catch {
      setConfirmError("parse.errors.start");
    } finally {
      setConfirming(false);
    }
  }, [
    confirming,
    lang,
    navigate,
    requestAuth,
    runtime,
    boundResumeId,
    t,
    targetJob,
    isWorkspaceDetail,
  ]);

  const HitDot: FC<{ hit: HitState }> = ({ hit }) => {
    const color =
      hit === true
        ? "var(--ei-color-ok)"
        : hit === "partial"
          ? "var(--ei-color-warn)"
          : "var(--ei-color-fg-muted)";
    const label =
      hit === true
        ? t("parse.hit")
        : hit === "partial"
          ? t("parse.partial")
          : t("parse.gap");
    const bg =
      hit === true
        ? "var(--ei-color-ok-soft)"
        : hit === "partial"
          ? "var(--ei-color-warn-soft)"
          : "transparent";
    const border = hit === false ? "1px dashed var(--ei-color-rule-strong)" : "none";

    return (
      <div
        style={{
          display: "flex",
          gap: 5,
          alignItems: "center",
          padding: "2px 7px",
          background: bg,
          border,
          borderRadius: "var(--ei-radius-sm)",
        }}
      >
        <div
          style={{
            width: 5,
            height: 5,
            borderRadius: 3,
            background: color,
          }}
        />
        <span
          style={{
            fontSize: 10.5,
            color,
            fontFamily: "var(--ei-font-mono)",
            letterSpacing: "0.04em",
            textTransform: "uppercase",
          }}
        >
          {label}
        </span>
      </div>
    );
  };

  if (stage === "failed") {
    return (
      <section
        data-testid={routeTestId}
        data-route-name={route.name}
        data-route-params={JSON.stringify(route.params)}
        className="ei-fadein"
        style={{
          minHeight: "calc(100vh - 58px)",
          display: "flex",
          alignItems: "center",
          justifyContent: "center",
          padding: 48,
        }}
      >
        <div
          data-testid="parse-error"
          style={{ maxWidth: 520, width: "100%", textAlign: "center" }}
        >
          <div
            data-testid="parse-failed-title"
            className="ei-serif"
            style={{
              fontSize: 28,
              color: "var(--ei-color-fg-primary)",
              letterSpacing: "-0.015em",
              marginBottom: 12,
            }}
          >
            {t("parse.failedTitle")}
          </div>
          <div
            data-testid="parse-failed-message"
            style={{
              fontSize: 14,
              color: "var(--ei-color-fg-tertiary)",
              marginBottom: 28,
              lineHeight: 1.5,
            }}
          >
            {t("parse.failedMessage")}
          </div>
          <div style={{ display: "flex", gap: 10, justifyContent: "center" }}>
            <button
              data-testid="parse-failed-reparse"
              onClick={handleReparse}
              style={{
                padding: "8px 18px",
                fontSize: 13.5,
                fontFamily: "var(--ei-font-sans)",
                background: "var(--ei-color-accent)",
                border: "none",
                borderRadius: "var(--ei-radius-sm)",
                color: "#fff",
                cursor: "pointer",
              }}
            >
              {t("parse.failedReparse")}
            </button>
            <button
              data-testid="parse-failed-home"
              onClick={handleCancel}
              style={{
                padding: "8px 18px",
                fontSize: 13.5,
                fontFamily: "var(--ei-font-sans)",
                background: "transparent",
                border: "1px solid var(--ei-color-rule-strong)",
                borderRadius: "var(--ei-radius-sm)",
                color: "var(--ei-color-fg-primary)",
                cursor: "pointer",
              }}
            >
              {t("parse.failedHome")}
            </button>
          </div>
        </div>
      </section>
    );
  }

  if (stage === "error") {
    return (
      <section
        data-testid={routeTestId}
        data-route-name={route.name}
        data-route-params={JSON.stringify(route.params)}
        className="ei-fadein"
        style={{
          minHeight: "calc(100vh - 58px)",
          display: "flex",
          alignItems: "center",
          justifyContent: "center",
          padding: 48,
        }}
      >
        <div
          data-testid="parse-error"
          style={{ maxWidth: 520, width: "100%", textAlign: "center" }}
        >
          <div
            className="ei-serif"
            style={{
              fontSize: 28,
              color: "var(--ei-color-fg-primary)",
              letterSpacing: "-0.015em",
              marginBottom: 12,
            }}
          >
            {t("parse.errorTitle")}
          </div>
          <div
            style={{
              fontSize: 14,
              color: "var(--ei-color-fg-tertiary)",
              marginBottom: 28,
            }}
          >
            {errorMessage ?? t("parse.errorMessage")}
          </div>
          <button
            onClick={handleCancel}
            style={{
              padding: "8px 18px",
              fontSize: 13.5,
              fontFamily: "var(--ei-font-sans)",
              background: "var(--ei-color-accent)",
              border: "none",
              borderRadius: "var(--ei-radius-sm)",
              color: "#fff",
              cursor: "pointer",
            }}
          >
            {t("parse.errorHome")}
          </button>
        </div>
      </section>
    );
  }

  if (stage === "loading") {
    if (isWorkspaceDetail) {
      return (
        <section
          data-testid={routeTestId}
          data-route-name={route.name}
          data-route-params={JSON.stringify(route.params)}
          className="ei-fadein ei-plan-detail-screen"
        >
          <div
            data-testid="workspace-detail-loading"
            className="ei-screen-card"
            style={{ color: "var(--ei-color-fg-tertiary)", fontSize: 13 }}
          >
            {t("workspace.detail.loading")}
          </div>
        </section>
      );
    }
    return (
      <section
        data-testid={routeTestId}
        data-route-name={route.name}
        data-route-params={JSON.stringify(route.params)}
        className="ei-fadein"
        style={{
          minHeight: "calc(100vh - 58px)",
          display: "flex",
          alignItems: "center",
          justifyContent: "center",
          padding: 48,
        }}
      >
        <div style={{ maxWidth: 520, width: "100%" }}>
          <div
            className="ei-label"
            style={{
              color: "var(--ei-color-fg-tertiary)",
              marginBottom: 12,
            }}
          >
            {t("parse.loadingLabel")}
          </div>
          <div
            className="ei-serif"
            style={{
              fontSize: 28,
              color: "var(--ei-color-fg-primary)",
              letterSpacing: "-0.015em",
              lineHeight: 1.3,
              marginBottom: 32,
            }}
          >
            {t("parse.loadingTitle")}
          </div>
          <div style={{ display: "flex", flexDirection: "column", gap: 14 }}>
            {steps.map((s, i) => {
              const done = i < step;
              const active = i === step;
              return (
                <div
                  key={i}
                  data-testid={`parse-loading-step-${i}`}
                  style={{
                    display: "flex",
                    gap: 14,
                    alignItems: "center",
                  }}
                >
                  <div
                    style={{
                      width: 22,
                      height: 22,
                      borderRadius: 11,
                      border: `1.5px solid ${
                        done
                          ? "var(--ei-color-ok)"
                          : active
                            ? "var(--ei-color-accent)"
                            : "var(--ei-color-rule-strong)"
                      }`,
                      background: done
                        ? "var(--ei-color-ok)"
                        : "transparent",
                      display: "flex",
                      alignItems: "center",
                      justifyContent: "center",
                      flexShrink: 0,
                    }}
                  >
                    {done && (
                      <svg
                        width="12"
                        height="12"
                        viewBox="0 0 12 12"
                        fill="none"
                        stroke="#fff"
                        strokeWidth="2.5"
                      >
                        <path d="M2 6l3 3 5-5" />
                      </svg>
                    )}
                    {active && (
                      <div
                        className="ei-pulse"
                        style={{
                          width: 6,
                          height: 6,
                          borderRadius: 3,
                          background: "var(--ei-color-accent)",
                        }}
                      />
                    )}
                  </div>
                  <div
                    style={{
                      fontSize: 14,
                      color: done
                        ? "var(--ei-color-fg-tertiary)"
                        : active
                          ? "var(--ei-color-fg-primary)"
                          : "var(--ei-color-fg-muted)",
                      textDecoration: done ? "line-through" : "none",
                    }}
                  >
                    {t(s as "parse.loadingStep1")}
                  </div>
                  {active && (
                    <div
                      style={{
                        fontFamily: "var(--ei-font-mono)",
                        fontSize: 11,
                        color: "var(--ei-color-fg-muted)",
                        marginLeft: "auto",
                      }}
                    >
                      <span className="ei-pulse">●</span>{" "}
                      {t("parse.loadingWorking")}
                    </div>
                  )}
                </div>
              );
            })}
          </div>
        </div>
      </section>
    );
  }

  const requirements = targetJob?.requirements ?? [];
  const mustHave = requirements.filter((r) => r.kind === "must_have");
  const niceToHave = requirements.filter((r) => r.kind === "nice_to_have");
  const hiddenSignals = requirements
    .filter((r) => r.kind === "hidden_signal")
    .map((r) => r.label);
  const progress = resolveTargetJobPracticeProgress(targetJob);
  const launchDisabled =
    !boundResumeId || confirming || !progress.currentRound;

  return (
    <>
      {confirming ? <PracticeLaunchTransition /> : null}
      <section
      data-testid={routeTestId}
      data-route-name={route.name}
      data-route-params={JSON.stringify(route.params)}
      className="ei-fadein ei-plan-detail-screen"
    >
      {/* Header */}
      <div
        data-testid="unified-plan-detail"
        className="ei-plan-detail-header"
      >
        <div className="ei-plan-detail-heading">
          <div className="ei-label ei-plan-detail-step">
            {t("parse.stepLabel")}
          </div>
          <div className="ei-plan-detail-title-row">
            <h1
              data-testid="unified-plan-detail-title"
              className="ei-serif ei-plan-detail-title"
            >
              {t("parse.previewTitle")}
            </h1>
            {route.name === "workspace" && targetJobId && targetJob?.id === targetJobId ? (
              boundResumeId ? (
                <button
                  type="button"
                  data-testid="parse-resume-link"
                  onClick={handleOpenBoundResume}
                  className="ei-plan-detail-resume-link"
                >
                  {t("parse.resumeBinding")}
                </button>
              ) : (
                <span
                  data-testid="parse-resume-missing"
                  className="ei-plan-detail-resume-missing"
                >
                  {t("parse.resumeEmptyTitle")}
                </span>
              )
            ) : null}
          </div>
          <div className="ei-plan-detail-subtitle">
            {t("parse.previewSub")}
          </div>
        </div>

        {route.name === "workspace" && targetJobId && targetJob?.id === targetJobId ? (
          <div
            data-testid="parse-leading-actions"
            className="ei-plan-detail-actions"
          >
            <button
              data-testid="parse-action-start-interview"
              onClick={handleStartInterview}
              disabled={launchDisabled}
              className="ei-plan-detail-primary-action"
            >
              {t("parse.startInterview")}
            </button>
            <span data-testid="parse-reports-entry">
              <button
                type="button"
                onClick={handleOpenReports}
                onMouseDown={(event) => {
                  event.currentTarget.style.transform = "translateY(0.5px)";
                }}
                onMouseUp={(event) => {
                  event.currentTarget.style.transform = "";
                }}
                onMouseLeave={(event) => {
                  event.currentTarget.style.transform = "";
                }}
                className="ei-plan-detail-secondary-action"
              >
                <svg
                  aria-hidden="true"
                  width="16"
                  height="16"
                  viewBox="0 0 24 24"
                  fill="none"
                  stroke="currentColor"
                  strokeWidth="1.5"
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  style={{ flexShrink: 0, display: "inline-block", verticalAlign: "middle" }}
                >
                  <path d="M5 12h14M13 6l6 6-6 6" />
                </svg>
                {t("parse.reports.label")}
              </button>
            </span>
          </div>
        ) : null}
      </div>

      {confirmError ? (
        <div data-testid="parse-confirm-error" className="ei-plan-detail-error">
          {t(confirmError)}
        </div>
      ) : null}

      {/* Basic fields */}
      <div
        className="ei-screen-card ei-plan-detail-card ei-plan-detail-basics"
        style={{ marginBottom: 12, padding: 0 }}
      >
        <div className="ei-plan-detail-card-heading">
          <PlanSectionIcon variant="basics" />
          <div
            className="ei-label"
            style={{ color: "var(--ei-color-fg-tertiary)" }}
          >
            {t("parse.basicsSection")}
          </div>
        </div>
        <div
          className="ei-plan-detail-basics-grid"
          style={{
            display: "grid",
            gridTemplateColumns: compactLayout ? "1fr" : "repeat(3, 1fr)",
            padding: "2px 24px",
          }}
        >
          {[
            {
              label: t("parse.basicsTitle"),
              value: targetJob?.title ?? "—",
              field: "title" as const,
            },
            {
              label: t("parse.basicsCompany"),
              value: targetJob?.companyName ?? "—",
              field: "company" as const,
            },
            {
              label: t("parse.basicsLevel"),
              value: "—",
              field: "level" as const,
            },
            {
              label: t("parse.basicsLocation"),
              value: targetJob?.locationText ?? "—",
              field: "location" as const,
            },
            {
              label: t("parse.basicsLanguage"),
              value: targetJob?.targetLanguage ?? "—",
              field: "language" as const,
            },
          ].map((r, i, arr) => (
            <div
              key={r.field}
              data-testid={`parse-basics-${r.field}`}
              style={{
                display: "flex",
                gap: 14,
                padding: "7px 0",
                borderBottom:
                  i < arr.length - 1
                    ? "1px dotted var(--ei-color-rule-strong)"
                    : "none",
                alignItems: "baseline",
              }}
            >
              <div
                className="ei-label"
                style={{
                  color: "var(--ei-color-fg-tertiary)",
                  minWidth: 68,
                  fontSize: 10.5,
                }}
              >
                {r.label}
              </div>
              <div
                style={{
                  flex: 1,
                  minWidth: 0,
                  fontSize: 14,
                  color: "var(--ei-color-fg-primary)",
                  padding: "2px 0",
                  fontFamily: "var(--ei-font-sans)",
                }}
              >
                {r.value}
              </div>
            </div>
          ))}
        </div>
      </div>

      {/* Requirements */}
      <div
        className="ei-plan-detail-requirements"
        style={{
          display: "grid",
          gridTemplateColumns: compactLayout ? "1fr" : "1fr 1fr",
          gap: 20,
          marginBottom: 12,
        }}
      >
        {/* Must Have */}
        <div
          className="ei-plan-detail-requirement-card"
          style={{
            background: "var(--ei-color-bg-card)",
            border: "1px solid var(--ei-color-rule-strong)",
            borderRadius: "var(--ei-radius-md)",
            padding: 0,
            cursor: "default",
            transition: "border-color .15s, transform .15s",
          }}
        >
          <div
            className="ei-plan-detail-card-heading"
            style={{
              padding: "14px 20px",
              borderBottom: "1px solid var(--ei-color-rule-strong)",
              display: "flex",
              justifyContent: "space-between",
              alignItems: "center",
            }}
          >
            <PlanSectionIcon variant="must" />
            <div
              className="ei-label"
              style={{ color: "var(--ei-color-fg-tertiary)" }}
            >
              {t("parse.mustHave")}
            </div>
            <div
              style={{
                fontSize: 11,
                color: "var(--ei-color-fg-tertiary)",
                fontFamily: "var(--ei-font-mono)",
              }}
            >
              {mustHave.length}
            </div>
          </div>
          <div>
            {mustHave.map((item, i) => (
              <div
                key={item.id}
                data-testid={`parse-requirement-must_have-${i}`}
                style={{
                  padding: "4px 20px",
                  borderBottom:
                    i < mustHave.length - 1
                      ? "1px dotted var(--ei-color-rule-strong)"
                      : "none",
                  display: "flex",
                  gap: 12,
                  alignItems: "flex-start",
                }}
              >
                <div style={{ marginTop: 2 }}>
                  <HitDot hit={hitStateFromEvidence(item.evidenceLevel)} />
                </div>
                <div style={{ flex: 1, minWidth: 0 }}>
                  <div
                    style={{
                      fontSize: 13.5,
                      color: "var(--ei-color-fg-primary)",
                      lineHeight: 1.45,
                    }}
                  >
                    {item.label}
                  </div>
                </div>
              </div>
            ))}
          </div>
        </div>

        {/* Nice to Have */}
        <div className="ei-screen-card ei-plan-detail-requirement-card" style={{ padding: 0, display: "block", gap: 0 }}>
          <div
            className="ei-plan-detail-card-heading"
            style={{
              padding: "14px 20px",
              borderBottom: "1px solid var(--ei-color-rule-strong)",
              display: "flex",
              justifyContent: "space-between",
              alignItems: "center",
            }}
          >
            <PlanSectionIcon variant="nice" />
            <div
              className="ei-label"
              style={{ color: "var(--ei-color-fg-tertiary)" }}
            >
              {t("parse.niceToHave")}
            </div>
            <div
              style={{
                fontSize: 11,
                color: "var(--ei-color-fg-tertiary)",
                fontFamily: "var(--ei-font-mono)",
              }}
            >
              {niceToHave.length}
            </div>
          </div>
          <div>
            {niceToHave.map((item, i) => (
              <div
                key={item.id}
                data-testid={`parse-requirement-nice_to_have-${i}`}
                style={{
                  padding: "4px 20px",
                  borderBottom:
                    i < niceToHave.length - 1
                      ? "1px dotted var(--ei-color-rule-strong)"
                      : "none",
                  display: "flex",
                  gap: 12,
                  alignItems: "flex-start",
                }}
              >
                <div style={{ marginTop: 2 }}>
                  <HitDot hit={hitStateFromEvidence(item.evidenceLevel)} />
                </div>
                <div style={{ flex: 1, minWidth: 0 }}>
                  <div
                    style={{
                      fontSize: 13.5,
                      color: "var(--ei-color-fg-primary)",
                      lineHeight: 1.45,
                    }}
                  >
                    {item.label}
                  </div>
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>

      {/* Hidden signals */}
      <div
        className="ei-screen-card ei-plan-detail-card ei-plan-detail-hidden"
        style={{ marginBottom: 12, borderColor: "var(--ei-color-accent)" }}
      >
        <div
          className="ei-plan-detail-hidden-heading"
          style={{
            display: "flex",
            justifyContent: "space-between",
            alignItems: "center",
            marginBottom: 14,
          }}
        >
          <div>
            <PlanSectionIcon variant="hidden" />
            <div
              className="ei-label"
              style={{
                color: "var(--ei-color-accent)",
                marginBottom: 4,
              }}
            >
              {t("parse.hiddenSignalsTitle")}
            </div>
            <div
              style={{ fontSize: 13, color: "var(--ei-color-fg-tertiary)" }}
            >
              {t("parse.hiddenSignalsSub")}
            </div>
          </div>
          <div
            style={{
              display: "flex",
              gap: 4,
              alignItems: "center",
              padding: "3px 10px",
              background: "var(--ei-color-accent-soft)",
              borderRadius: "var(--ei-radius-pill)",
              fontSize: 10.5,
              color: "var(--ei-color-accent)",
              fontFamily: "var(--ei-font-mono)",
              letterSpacing: "0.04em",
            }}
          >
            <svg
              width="10"
              height="10"
              viewBox="0 0 10 10"
              fill="currentColor"
            >
              <path d="M5 0l1.5 3 3.5.5-2.5 2.5.5 3.5L5 7.5 2 9.5l.5-3.5L0 3.5l3.5-.5L5 0z" />
            </svg>
            {t("parse.hiddenConfidence")}
          </div>
        </div>
        <div className="ei-plan-detail-hidden-items">
          {hiddenSignals.map((h, i) => (
            <div
              key={i}
              data-testid={`parse-hidden-signal-${i}`}
              style={{
                display: "flex",
                gap: 10,
                alignItems: "flex-start",
                padding: "4px 10px",
                background: "var(--ei-color-bg-soft)",
                borderRadius: "var(--ei-radius-sm)",
              }}
            >
              <svg
                width="12"
                height="12"
                viewBox="0 0 10 10"
                fill="var(--ei-color-accent)"
                style={{ marginTop: 3, flexShrink: 0 }}
              >
                <path d="M5 0l1.5 3 3.5.5-2.5 2.5.5 3.5L5 7.5 2 9.5l.5-3.5L0 3.5l3.5-.5L5 0z" />
              </svg>
              <div
                style={{
                  fontSize: 13.5,
                  color: "var(--ei-color-fg-primary)",
                  lineHeight: 1.5,
                  flex: 1,
                }}
              >
                {h}
              </div>
            </div>
          ))}
        </div>
      </div>

      {/* Round assumptions */}
      <div className="ei-screen-card ei-plan-detail-card ei-plan-detail-rounds" style={{ marginBottom: 28 }}>
        <div
          className="ei-plan-detail-round-heading"
          style={{
            display: "flex",
            justifyContent: "space-between",
            alignItems: "center",
            marginBottom: 14,
          }}
        >
          <div
            className="ei-label"
            style={{ color: "var(--ei-color-fg-tertiary)" }}
          >
            <PlanSectionIcon variant="rounds" />
            {t("parse.roundsTitle")}
          </div>
          <div
            style={{
              fontSize: 11,
              color: "var(--ei-color-fg-tertiary)",
              fontFamily: "var(--ei-font-mono)",
            }}
          >
            {t("parse.roundsHint")}
          </div>
        </div>
        <div
          data-testid="parse-rounds"
          className="ei-plan-detail-round-grid"
          style={{
            display: "grid",
            gridTemplateColumns: compactLayout
              ? "1fr"
              : `repeat(${Math.min(Math.max(rounds.length, 1), 4)}, 1fr)`,
            gap: 10,
          }}
        >
          {rounds.map((r, i) => {
            const roundState = !progress.valid
              ? null
              : i < progress.completedCount
                ? "done"
                : i === progress.currentIndex
                  ? "current"
                  : "pending";
            const roundStateLabel =
              roundState === "done"
                ? t("parse.roundState.done")
                : roundState === "current"
                  ? t("parse.roundState.current")
                  : roundState === "pending"
                    ? t("parse.roundState.pending")
                    : null;
            const roundBackground =
              roundState === "done"
                ? "var(--ei-color-ok-soft)"
                : roundState === "current"
                  ? "var(--ei-color-accent-soft)"
                  : "var(--ei-color-bg-soft)";
            const roundBorder =
              roundState === "done"
                ? "var(--ei-color-ok)"
                : roundState === "current"
                  ? "var(--ei-color-accent)"
                  : "var(--ei-color-rule-strong)";
            const roundLabelColor =
              roundState === "done"
                ? "var(--ei-color-ok)"
                : roundState === "current"
                  ? "var(--ei-color-accent)"
                  : "var(--ei-color-fg-tertiary)";
            return (
              <div
                key={r.id}
                data-testid={`parse-round-${i}`}
                data-round-state={roundState ?? undefined}
                style={{
                  padding: "8px 12px",
                  background: roundBackground,
                  border: `1px solid ${roundBorder}`,
                  borderRadius: "var(--ei-radius-sm)",
                  position: "relative",
                }}
              >
                <div
                  style={{
                    display: "flex",
                    justifyContent: "space-between",
                    alignItems: "center",
                    gap: 8,
                    marginBottom: 5,
                  }}
                >
                  <span
                    style={{
                      fontFamily: "var(--ei-font-mono)",
                      fontSize: 10.5,
                      color: "var(--ei-color-fg-muted)",
                      letterSpacing: "0.06em",
                    }}
                  >
                    R{i + 1}
                  </span>
                  {roundStateLabel && (
                    <span
                      data-testid={`parse-round-state-${i}`}
                      style={{
                        fontFamily: "var(--ei-font-mono)",
                        fontSize: 10.5,
                        color: roundLabelColor,
                        letterSpacing: "0.04em",
                      }}
                    >
                      {roundStateLabel}
                    </span>
                  )}
                </div>
              <div
                style={{
                  fontSize: 13,
                  color: "var(--ei-color-fg-primary)",
                  fontWeight: 500,
                  marginBottom: 4,
                }}
              >
                {r.name}
              </div>
              <div
                style={{
                  fontSize: 11.5,
                  color: "var(--ei-color-fg-tertiary)",
                  lineHeight: 1.45,
                }}
              >
                {r.focus}
              </div>
              </div>
            );
          })}
        </div>
      </div>

      </section>
    </>
  );
};
