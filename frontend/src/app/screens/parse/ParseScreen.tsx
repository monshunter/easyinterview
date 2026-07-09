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
import { useI18n } from "../../i18n/messages";
import { isSelectableInterviewResume } from "../../interview-context/selectableResume";
import { startPracticeFromParams } from "../../interview-context/startPractice";
import { useNavigation } from "../../navigation/NavigationProvider";
import { interviewContextFromTargetJob } from "../../navigation/interviewContext";
import type { Route } from "../../routes";
import type { Resume, TargetJob } from "../../../api/generated/types";

type Stage = "loading" | "preview" | "error" | "failed";

interface ParseScreenProps {
  route: Route;
  _mockStage?: Stage;
  _mockTargetJob?: TargetJob;
}

type HitState = true | "partial" | false;

function nextHit(current: HitState): HitState {
  if (current === true) return "partial";
  if (current === "partial") return false;
  return true;
}

const loadingStepKeys = [
  "parse.loadingStep1",
  "parse.loadingStep2",
  "parse.loadingStep3",
  "parse.loadingStep4",
] as const;

const loadingStepTicks = [600, 900, 800, 700] as const;
const loadingPreviewDelay =
  loadingStepTicks.reduce((total, tick) => total + tick, 0) + 200;

const defaultPracticeParams = {
  mode: "text",
  modality: "text",
  practiceMode: "strict",
  practiceGoal: "baseline",
  hintUsed: "false",
  hintCount: "0",
} as const;

function sortByMostRecentResume(a: Resume, b: Resume): number {
  return Date.parse(b.updatedAt) - Date.parse(a.updatedAt);
}

function buildInterviewParams(
  job: TargetJob,
  resumeId: string,
): Record<string, string> {
  return {
    ...interviewContextFromTargetJob(job, { resumeId }),
    ...defaultPracticeParams,
    language: job.targetLanguage || "",
  };
}

function resumeMeta(resume: Resume): string {
  return [resume.language, resume.sourceType, resume.updatedAt.slice(0, 10)]
    .filter(Boolean)
    .join(" · ");
}

function safeScrollToTop(): void {
  if (navigator.userAgent.toLowerCase().includes("jsdom")) return;
  try {
    window.scrollTo({ top: 0, behavior: "smooth" });
  } catch {
    // jsdom exposes scrollTo but throws because it is not implemented.
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

export const ParseScreen: FC<ParseScreenProps> = ({
  route,
  _mockStage,
  _mockTargetJob,
}) => {
  const { t, lang } = useI18n();
  const { navigate } = useNavigation();
  const runtime = useAppRuntimeOptional();
  const requestAuth = useRequestAuth();
  const [stage, setStage] = useState<Stage>(_mockStage ?? "loading");
  const [step, setStep] = useState(0);
  const [targetJob, setTargetJob] = useState<TargetJob | null>(
    _mockTargetJob ?? null,
  );
  const [editedTitle, setEditedTitle] = useState("");
  const [editedCompany, setEditedCompany] = useState("");
  const [editedLocation, setEditedLocation] = useState("");
  const [editedNotes, setEditedNotes] = useState("");
  const [hitToggles, setHitToggles] = useState<Record<string, HitState>>({});
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const [confirmError, setConfirmError] = useState<string | null>(null);
  const [confirming, setConfirming] = useState(false);
  const [resumesLoading, setResumesLoading] = useState(false);
  const [readyResumes, setReadyResumes] = useState<Resume[]>([]);
  const [selectedResumeId, setSelectedResumeId] = useState("");
  const [resumeError, setResumeError] = useState<string | null>(null);
  const [resumePickerOpen, setResumePickerOpen] = useState(false);
  const [pollNonce, setPollNonce] = useState(0);
  const [pendingReadyJob, setPendingReadyJob] = useState<TargetJob | null>(null);
  const [loadingComplete, setLoadingComplete] = useState(false);
  const pollingRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const resumeRequestSeqRef = useRef(0);

  const steps = loadingStepKeys;
  const targetJobId =
    typeof route.params?.targetJobId === "string"
      ? route.params.targetJobId
      : undefined;
  const isWorkspaceDetail = route.name === "workspace";
  const compactLayout = useParseCompactLayout();
  const routeTestId = isWorkspaceDetail ? "route-workspace" : "route-parse";
  const routeResumeId =
    typeof route.params?.resumeId === "string"
      ? route.params.resumeId
      : undefined;

  const hydrateReadyJob = useCallback((job: TargetJob) => {
    setTargetJob(job);
    setEditedTitle(job.title ?? "");
    setEditedCompany(job.companyName ?? "");
    setEditedLocation(job.locationText ?? "");
  }, []);

  useEffect(() => {
    if (_mockStage || _mockTargetJob) return;

    if (pollingRef.current) {
      clearTimeout(pollingRef.current);
      pollingRef.current = null;
    }
    setTargetJob(null);
    setEditedTitle("");
    setEditedCompany("");
    setEditedLocation("");
    setEditedNotes("");
    setHitToggles({});
    setErrorMessage(null);
    setConfirmError(null);
    setReadyResumes([]);
    setSelectedResumeId("");
    setResumeError(null);
    setResumePickerOpen(false);
    setPendingReadyJob(null);
    setLoadingComplete(false);
    setStep(0);
    setStage("loading");
  }, [targetJobId, _mockStage, _mockTargetJob]);

  useEffect(() => {
    if (stage !== "loading" || _mockStage) return;

    setStep(0);
    setLoadingComplete(false);

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
    timers.push(
      setTimeout(() => {
        setLoadingComplete(true);
      }, loadingPreviewDelay),
    );

    return () => {
      timers.forEach((timer) => clearTimeout(timer));
    };
  }, [stage, _mockStage, pollNonce]);

  useEffect(() => {
    if (stage !== "loading" || !loadingComplete || !pendingReadyJob) return;

    hydrateReadyJob(pendingReadyJob);
    setPendingReadyJob(null);
    setStage("preview");
  }, [stage, loadingComplete, pendingReadyJob, hydrateReadyJob]);

  // Poll getTargetJob when in loading stage
  useEffect(() => {
    if (_mockStage || _mockTargetJob) return;
    if (!runtime) return;

    if (!targetJobId) {
      setStage("error");
      setErrorMessage(lang === "en" ? "Missing target job ID." : "缺少目标岗位 ID。");
      return;
    }

    let cancelled = false;

    const poll = async () => {
      if (cancelled) return;
      try {
        const job = await runtime.client.getTargetJob(targetJobId);
        if (cancelled) return;

        if (job.analysisStatus === "ready") {
          if (isWorkspaceDetail) {
            hydrateReadyJob(job);
            setStage("preview");
          } else {
            setPendingReadyJob(job);
          }
        } else if (job.analysisStatus === "failed") {
          setStage("failed");
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
      setPendingReadyJob(null);
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [
    runtime,
    targetJobId,
    _mockStage,
    _mockTargetJob,
    pollNonce,
    isWorkspaceDetail,
    hydrateReadyJob,
  ]);

  useEffect(() => {
    if (_mockTargetJob) {
      hydrateReadyJob(_mockTargetJob);
    }
  }, [_mockTargetJob, hydrateReadyJob]);

  useEffect(() => {
    const client = runtime?.client;
    const authenticated = runtime?.auth.status === "authenticated";
    if (stage !== "preview" || !targetJob || !client || !authenticated) {
      setResumesLoading(false);
      setReadyResumes([]);
      setSelectedResumeId("");
      setResumeError(null);
      setResumePickerOpen(false);
      return;
    }

    let active = true;
    const requestSeq = resumeRequestSeqRef.current + 1;
    resumeRequestSeqRef.current = requestSeq;

    setResumesLoading(true);
    setResumeError(null);

    client
      .listResumes({ headers: { "Accept-Language": lang } })
      .then((page) => {
        if (!active || resumeRequestSeqRef.current !== requestSeq) return;
        const ready = page.items
          .filter(isSelectableInterviewResume)
          .sort(sortByMostRecentResume);
        setReadyResumes(ready);
        setSelectedResumeId((current) => {
          if (current && ready.some((resume) => resume.id === current)) {
            return current;
          }
          if (
            routeResumeId &&
            ready.some((resume) => resume.id === routeResumeId)
          ) {
            return routeResumeId;
          }
          const targetJobResumeId = targetJob.resumeId?.trim();
          if (
            targetJobResumeId &&
            ready.some((resume) => resume.id === targetJobResumeId)
          ) {
            return targetJobResumeId;
          }
          return "";
        });
      })
      .catch((err: unknown) => {
        if (!active || resumeRequestSeqRef.current !== requestSeq) return;
        setReadyResumes([]);
        setSelectedResumeId("");
        setResumeError(err instanceof Error ? err.message : String(err));
      })
      .finally(() => {
        if (active && resumeRequestSeqRef.current === requestSeq) {
          setResumesLoading(false);
        }
      });

    return () => {
      active = false;
    };
  }, [
    lang,
    runtime?.auth.status,
    runtime?.client,
    routeResumeId,
    stage,
    targetJob,
  ]);

  const toggleHit = useCallback(
    (reqId: string) => {
      setHitToggles((prev) => ({
        ...prev,
        [reqId]: nextHit(prev[reqId] ?? false),
      }));
    },
    [],
  );

  const getHitState = useCallback(
    (reqId: string): HitState => hitToggles[reqId] ?? false,
    [hitToggles],
  );

  const handleCancel = useCallback(() => {
    navigate(isWorkspaceDetail ? { name: "workspace", params: {} } : { name: "home", params: {} });
  }, [isWorkspaceDetail, navigate]);

  const handleReparse = useCallback(() => {
    if (pollingRef.current) {
      clearTimeout(pollingRef.current);
      pollingRef.current = null;
    }
    setPendingReadyJob(null);
    setLoadingComplete(false);
    setErrorMessage(null);
    setStep(0);
    setStage("loading");
    setPollNonce((n) => n + 1);
    safeScrollToTop();
  }, []);

  const selectedResume = useMemo(
    () => readyResumes.find((resume) => resume.id === selectedResumeId) ?? null,
    [readyResumes, selectedResumeId],
  );

  const saveTargetJobEdits = useCallback(async (): Promise<boolean> => {
    if (!targetJob || !runtime) return false;
    setConfirmError(null);
    setConfirming(true);
    const ik = `ik-${crypto.randomUUID()}`;
    try {
      await runtime.client.updateTargetJob(
        targetJob.id,
        {
          titleHint: editedTitle || undefined,
          companyNameHint: editedCompany || undefined,
          locationText: editedLocation || undefined,
          notes: editedNotes || undefined,
        },
        { idempotencyKey: ik },
      );
      return true;
    } catch (err: unknown) {
      setConfirmError(err instanceof Error ? err.message : String(err));
      return false;
    } finally {
      setConfirming(false);
    }
  }, [
    editedCompany,
    editedLocation,
    editedNotes,
    editedTitle,
    runtime,
    targetJob,
  ]);

  const handleSavePlan = useCallback(async () => {
    if (!targetJob || !selectedResume || confirming) return;
    const parseParams = buildInterviewParams(targetJob, selectedResume.id);

    if (runtime?.auth.status !== "authenticated") {
      requestAuth({
        type: "save_interview_plan",
        label: t("parse.savePlanOnly"),
        route: "parse",
        params: parseParams,
      });
      return;
    }

    if (await saveTargetJobEdits()) {
      navigate({
        name: "workspace",
        params: {},
      });
    }
  }, [
    confirming,
    navigate,
    requestAuth,
    runtime?.auth.status,
    saveTargetJobEdits,
    selectedResume,
    t,
    targetJob,
  ]);

  const handleStartInterview = useCallback(async () => {
    if (!targetJob || !selectedResume || confirming || !runtime) return;
    const practiceParams = buildInterviewParams(targetJob, selectedResume.id);

    if (runtime?.auth.status !== "authenticated") {
      requestAuth({
        type: "start_practice",
        label: t("parse.startInterview"),
        route: "parse",
        params: practiceParams,
      });
      return;
    }

    if (await saveTargetJobEdits()) {
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
      } catch (err: unknown) {
        setConfirmError(err instanceof Error ? err.message : String(err));
      } finally {
        setConfirming(false);
      }
    }
  }, [
    confirming,
    lang,
    navigate,
    requestAuth,
    runtime,
    saveTargetJobEdits,
    selectedResume,
    t,
    targetJob,
  ]);

  const handleCreateResume = useCallback(() => {
    navigate({
      name: "resume_versions",
      params: targetJob?.id
        ? { flow: "create", targetJobId: targetJob.id }
        : { flow: "create" },
    });
  }, [navigate, targetJob?.id]);

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
          <div
            data-testid="parse-loading-footer"
            style={{
              marginTop: 40,
              paddingTop: 20,
              borderTop: "1px dotted var(--ei-color-rule-strong)",
              fontSize: 12,
              color: "var(--ei-color-fg-tertiary)",
              fontFamily: "var(--ei-font-mono)",
              lineHeight: 1.6,
            }}
          >
            <div>model · backend-managed · current locale</div>
            <div>rubric · target-job parse · provenance redacted</div>
            <div>typical · 3–6s · this one · slightly richer JD</div>
          </div>
        </div>
      </section>
    );
  }

  const requirements = targetJob?.requirements ?? [];
  const mustHave = requirements.filter((r) => r.kind === "must_have");
  const niceToHave = requirements.filter((r) => r.kind === "nice_to_have");
  const hiddenSignals = targetJob?.summary?.interviewHypotheses ?? [];
  const rounds: { name: string; focus: string }[] = [
    { name: t("parse.round1Name"), focus: t("parse.round1Focus") },
    { name: t("parse.round2Name"), focus: t("parse.round2Focus") },
    { name: t("parse.round3Name"), focus: t("parse.round3Focus") },
    { name: t("parse.round4Name"), focus: t("parse.round4Focus") },
  ];
  const launchDisabled = resumesLoading || !selectedResume || confirming;

  return (
    <section
      data-testid={routeTestId}
      data-route-name={route.name}
      data-route-params={JSON.stringify(route.params)}
      className="ei-fadein"
      style={{
        maxWidth: 1200,
        margin: "0 auto",
        padding: compactLayout ? "24px 16px 72px" : "32px 48px 96px",
      }}
    >
      {/* Header */}
      <div
        data-testid="unified-plan-detail"
        style={{
          display: "flex",
          flexDirection: compactLayout ? "column" : "row",
          justifyContent: "space-between",
          alignItems: "flex-start",
          gap: compactLayout ? 16 : 24,
          marginBottom: 24,
        }}
      >
        <div>
          <div
            className="ei-label"
            style={{
              color: "var(--ei-color-fg-tertiary)",
              marginBottom: 8,
            }}
          >
            {t("parse.stepLabel")}
          </div>
          <h1
            data-testid="unified-plan-detail-title"
            className="ei-serif"
            style={{
              fontSize: compactLayout ? 28 : 32,
              margin: 0,
              color: "var(--ei-color-fg-primary)",
              letterSpacing: "-0.02em",
              lineHeight: 1.2,
            }}
          >
            {t("parse.previewTitle")}
          </h1>
          <div
            style={{
              fontSize: 14,
              color: "var(--ei-color-fg-tertiary)",
              marginTop: 8,
              maxWidth: 620,
              lineHeight: 1.5,
            }}
          >
            {t("parse.previewSub")}
          </div>
        </div>
        <div style={{ textAlign: compactLayout ? "left" : "right" }}>
          <div
            className="ei-label"
            style={{
              color: "var(--ei-color-fg-tertiary)",
              marginBottom: 4,
            }}
          >
            {t("parse.sourceLabel")}
          </div>
          <div
            style={{
              fontSize: 12,
              fontFamily: "var(--ei-font-mono)",
              color: "var(--ei-color-fg-secondary)",
              maxWidth: 280,
              wordBreak: "break-all",
            }}
          >
            {targetJob?.sourceUrl ?? "—"}
          </div>
          <div
            style={{
              fontSize: 11,
              color: "var(--ei-color-fg-tertiary)",
              marginTop: 4,
            }}
          >
            {t("parse.fetchedNow")}
          </div>
        </div>
      </div>

      {/* Basic fields */}
      <div
        className="ei-screen-card"
        style={{ marginBottom: 20, padding: 0 }}
      >
        <div
          style={{
            padding: "16px 24px",
            borderBottom: "1px solid var(--ei-color-rule-strong)",
          }}
        >
          <div
            className="ei-label"
            style={{ color: "var(--ei-color-fg-tertiary)" }}
          >
            {t("parse.basicsSection")}
          </div>
        </div>
        <div
          style={{
            display: "grid",
            gridTemplateColumns: compactLayout ? "1fr" : "repeat(2, 1fr)",
            padding: "6px 24px",
          }}
        >
          {[
            {
              label: t("parse.basicsTitle"),
              value: editedTitle,
              field: "title" as const,
              readOnly: false,
            },
            {
              label: t("parse.basicsCompany"),
              value: editedCompany,
              field: "company" as const,
              readOnly: false,
            },
            {
              label: t("parse.basicsLevel"),
              value: "—",
              field: "level" as const,
              readOnly: true,
            },
            {
              label: t("parse.basicsLocation"),
              value: editedLocation,
              field: "location" as const,
              readOnly: false,
            },
            {
              label: t("parse.basicsLanguage"),
              value: targetJob?.targetLanguage ?? "—",
              field: "language" as const,
              readOnly: true,
            },
          ].map((r, i, arr) => (
            <div
              key={r.field}
              data-testid={`parse-basics-${r.field}`}
              style={{
                display: "flex",
                gap: 14,
                padding: "12px 0",
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
              {r.readOnly ? (
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
              ) : (
                <input
                  value={r.value}
                  onChange={(e) => {
                    if (r.field === "title") setEditedTitle(e.target.value);
                    if (r.field === "company") setEditedCompany(e.target.value);
                    if (r.field === "location")
                      setEditedLocation(e.target.value);
                  }}
                  style={{
                    flex: 1,
                    minWidth: 0,
                    fontSize: 14,
                    color: "var(--ei-color-fg-primary)",
                    background: "transparent",
                    border: "none",
                    outline: "none",
                    borderBottom: "1px dashed transparent",
                    padding: "2px 0",
                    fontFamily: "var(--ei-font-sans)",
                  }}
                  onFocus={(e) =>
                    (e.target.style.borderBottomColor =
                      "var(--ei-color-accent)")
                  }
                  onBlur={(e) =>
                    (e.target.style.borderBottomColor = "transparent")
                  }
                />
              )}
            </div>
          ))}

          {/* Notes field */}
          <div
            data-testid="parse-basics-notes"
            style={{
              display: "flex",
              gap: 14,
              padding: "12px 0",
              gridColumn: "1 / -1",
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
              {t("parse.basicsNotes")}
            </div>
            <input
              value={editedNotes}
              onChange={(e) => setEditedNotes(e.target.value)}
              placeholder={t("parse.basicsNotesPlaceholder")}
              style={{
                flex: 1,
                minWidth: 0,
                fontSize: 14,
                color: "var(--ei-color-fg-primary)",
                background: "transparent",
                border: "none",
                outline: "none",
                borderBottom: "1px dashed transparent",
                padding: "2px 0",
                fontFamily: "var(--ei-font-sans)",
              }}
              onFocus={(e) =>
                (e.target.style.borderBottomColor = "var(--ei-color-accent)")
              }
              onBlur={(e) =>
                (e.target.style.borderBottomColor = "transparent")
              }
            />
          </div>
        </div>
      </div>

      {/* Requirements */}
      <div
        style={{
          display: "grid",
          gridTemplateColumns: compactLayout ? "1fr" : "1fr 1fr",
          gap: 20,
          marginBottom: 20,
        }}
      >
        {/* Must Have */}
        <div className="ei-screen-card" style={{ padding: 0 }}>
          <div
            style={{
              padding: "14px 20px",
              borderBottom: "1px solid var(--ei-color-rule-strong)",
              display: "flex",
              justifyContent: "space-between",
              alignItems: "center",
            }}
          >
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
                  padding: "12px 20px",
                  borderBottom:
                    i < mustHave.length - 1
                      ? "1px dotted var(--ei-color-rule-strong)"
                      : "none",
                  display: "flex",
                  gap: 12,
                  alignItems: "flex-start",
                }}
              >
                <button
                  data-testid={`parse-requirement-must_have-${i}-toggle`}
                  onClick={() => toggleHit(item.id)}
                  style={{
                    background: "transparent",
                    border: "none",
                    padding: 0,
                    cursor: "pointer",
                    marginTop: 2,
                  }}
                >
                  <HitDot hit={getHitState(item.id)} />
                </button>
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
        <div className="ei-screen-card" style={{ padding: 0 }}>
          <div
            style={{
              padding: "14px 20px",
              borderBottom: "1px solid var(--ei-color-rule-strong)",
              display: "flex",
              justifyContent: "space-between",
              alignItems: "center",
            }}
          >
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
                  padding: "12px 20px",
                  borderBottom:
                    i < niceToHave.length - 1
                      ? "1px dotted var(--ei-color-rule-strong)"
                      : "none",
                  display: "flex",
                  gap: 12,
                  alignItems: "flex-start",
                }}
              >
                <button
                  data-testid={`parse-requirement-nice_to_have-${i}-toggle`}
                  onClick={() => toggleHit(item.id)}
                  style={{
                    background: "transparent",
                    border: "none",
                    padding: 0,
                    cursor: "pointer",
                    marginTop: 2,
                  }}
                >
                  <HitDot hit={getHitState(item.id)} />
                </button>
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
        className="ei-screen-card"
        style={{ marginBottom: 20, borderColor: "var(--ei-color-accent)" }}
      >
        <div
          style={{
            display: "flex",
            justifyContent: "space-between",
            alignItems: "center",
            marginBottom: 14,
          }}
        >
          <div>
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
        <div style={{ display: "flex", flexDirection: "column", gap: 8 }}>
          {hiddenSignals.map((h, i) => (
            <div
              key={i}
              data-testid={`parse-hidden-signal-${i}`}
              style={{
                display: "flex",
                gap: 10,
                alignItems: "flex-start",
                padding: "8px 12px",
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
              <button
                style={{
                  background: "transparent",
                  border: "none",
                  color: "var(--ei-color-fg-tertiary)",
                  cursor: "pointer",
                  fontSize: 11,
                  fontFamily: "var(--ei-font-mono)",
                }}
              >
                {t("parse.hiddenRemove")}
              </button>
            </div>
          ))}
        </div>
      </div>

      {/* Round assumptions */}
      <div className="ei-screen-card" style={{ marginBottom: 28 }}>
        <div
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
          style={{
            display: "grid",
            gridTemplateColumns: compactLayout ? "1fr" : "repeat(4, 1fr)",
            gap: 10,
          }}
        >
          {rounds.map((r, i) => (
            <div
              key={i}
              data-testid={`parse-round-${i}`}
              style={{
                padding: "12px 14px",
                background: "var(--ei-color-bg-soft)",
                border: "1px solid var(--ei-color-rule-strong)",
                borderRadius: "var(--ei-radius-sm)",
                position: "relative",
              }}
            >
              <div
                style={{
                  fontFamily: "var(--ei-font-mono)",
                  fontSize: 10.5,
                  color: "var(--ei-color-fg-muted)",
                  marginBottom: 5,
                  letterSpacing: "0.06em",
                }}
              >
                R{i + 1}
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
          ))}
        </div>
      </div>

      {/* Interview launch */}
      <div
        data-testid="parse-launch"
        className="ei-screen-card"
        style={{
          marginBottom: 28,
          borderColor: selectedResume
            ? "var(--ei-color-ok)"
            : "var(--ei-color-warn)",
        }}
      >
        <div
          style={{
            display: "grid",
            gridTemplateColumns: compactLayout
              ? "minmax(0, 1fr)"
              : "1fr minmax(260px, 360px)",
            gap: 24,
            alignItems: "start",
          }}
        >
          <div>
            <div
              className="ei-label"
              style={{
                color: "var(--ei-color-fg-tertiary)",
                marginBottom: 6,
              }}
            >
              {t("parse.launchTitle")}
            </div>
            <div
              className="ei-serif"
              style={{
                fontSize: 22,
                color: "var(--ei-color-fg-primary)",
                letterSpacing: "-0.01em",
                lineHeight: 1.25,
                marginBottom: 8,
              }}
            >
              {t("parse.launchHeading")}
            </div>
            <div
              style={{
                fontSize: 13,
                color: "var(--ei-color-fg-tertiary)",
                lineHeight: 1.5,
                maxWidth: 620,
              }}
            >
              {t("parse.launchSub")}
            </div>
          </div>

          <div
            data-testid="parse-resume-binding"
            style={{
              padding: 14,
              background: "var(--ei-color-bg-soft)",
              border: "1px solid var(--ei-color-rule-strong)",
              borderRadius: "var(--ei-radius-sm)",
              minWidth: 0,
	            }}
          >
            <div
              className="ei-label"
              style={{
                color: selectedResume
                  ? "var(--ei-color-ok)"
                  : "var(--ei-color-warn)",
                marginBottom: 8,
              }}
            >
              {t("parse.resumeBinding")}
            </div>

            {resumesLoading ? (
              <div
                data-testid="parse-resume-loading"
                style={{
                  fontSize: 13,
                  color: "var(--ei-color-fg-tertiary)",
                  lineHeight: 1.5,
                }}
              >
                {t("parse.resumeLoading")}
              </div>
            ) : selectedResume ? (
              <>
                <div
                  style={{
                    fontSize: 14,
                    fontWeight: 600,
                    color: "var(--ei-color-fg-primary)",
                    lineHeight: 1.35,
                    marginBottom: 4,
                  }}
                >
                  {selectedResume.displayName || selectedResume.title}
                </div>
                <div
                  style={{
                    fontSize: 11.5,
                    color: "var(--ei-color-fg-tertiary)",
                    fontFamily: "var(--ei-font-mono)",
                    marginBottom: 12,
                  }}
                >
                  {resumeMeta(selectedResume)}
                </div>
                <button
                  data-testid="parse-resume-picker-toggle"
                  onClick={() => setResumePickerOpen((open) => !open)}
                  style={{
                    padding: "7px 12px",
                    fontSize: 12.5,
                    fontFamily: "var(--ei-font-sans)",
                    background: "transparent",
                    border: "1px solid var(--ei-color-rule-strong)",
                    borderRadius: "var(--ei-radius-sm)",
                    color: "var(--ei-color-fg-primary)",
                    cursor: "pointer",
                  }}
                >
                  {t("parse.resumeChange")}
                </button>
              </>
            ) : readyResumes.length > 0 ? (
              <div data-testid="parse-resume-required">
                <div
                  style={{
                    fontSize: 14,
                    fontWeight: 600,
                    color: "var(--ei-color-fg-primary)",
                    marginBottom: 4,
                  }}
                >
                  {t("parse.resumeRequiredTitle")}
                </div>
                <div
                  style={{
                    fontSize: 12.5,
                    color: "var(--ei-color-fg-tertiary)",
                    lineHeight: 1.5,
                    marginBottom: 12,
                  }}
                >
                  {t("parse.resumeRequiredBody")}
                </div>
              </div>
            ) : (
              <div data-testid="parse-resume-empty">
                <div
                  style={{
                    fontSize: 14,
                    fontWeight: 600,
                    color: "var(--ei-color-fg-primary)",
                    marginBottom: 4,
                  }}
                >
                  {t("parse.resumeEmptyTitle")}
                </div>
                <div
                  style={{
                    fontSize: 12.5,
                    color: "var(--ei-color-fg-tertiary)",
                    lineHeight: 1.5,
                    marginBottom: 12,
                  }}
                >
                  {resumeError ?? t("parse.resumeEmptyBody")}
                </div>
                <button
                  data-testid="parse-resume-create"
                  onClick={handleCreateResume}
                  style={{
                    padding: "7px 12px",
                    fontSize: 12.5,
                    fontFamily: "var(--ei-font-sans)",
                    background: "var(--ei-color-accent)",
                    border: "none",
                    borderRadius: "var(--ei-radius-sm)",
                    color: "#fff",
                    cursor: "pointer",
                  }}
                >
                  {t("parse.resumeCreate")}
                </button>
              </div>
            )}

            {(resumePickerOpen || (!selectedResume && readyResumes.length > 0)) &&
              readyResumes.length > 0 && (
              <div
                data-testid="parse-resume-picker"
                style={{
                  marginTop: 12,
                  display: "flex",
                  flexDirection: "column",
                  gap: 8,
                }}
              >
                {readyResumes.map((resume) => {
                  const active = resume.id === selectedResumeId;
                  return (
                    <button
                      key={resume.id}
                      data-testid={`parse-resume-option-${resume.id}`}
                      onClick={() => {
                        setSelectedResumeId(resume.id);
                        setResumePickerOpen(false);
                      }}
                      style={{
                        textAlign: "left",
                        padding: "9px 10px",
                        background: active
                          ? "var(--ei-color-accent-soft)"
                          : "transparent",
                        border: `1px solid ${
                          active
                            ? "var(--ei-color-accent)"
                            : "var(--ei-color-rule-strong)"
                        }`,
                        borderRadius: "var(--ei-radius-sm)",
                        color: "var(--ei-color-fg-primary)",
                        cursor: "pointer",
                      }}
                    >
                      <div style={{ fontSize: 12.5, fontWeight: 600 }}>
                        {resume.displayName || resume.title}
                      </div>
                      <div
                        style={{
                          fontSize: 11,
                          color: "var(--ei-color-fg-tertiary)",
                          fontFamily: "var(--ei-font-mono)",
                          marginTop: 2,
                        }}
                      >
                        {resumeMeta(resume)}
                      </div>
                    </button>
                  );
                })}
              </div>
            )}
          </div>
        </div>
      </div>

      {/* Footer actions */}
      {confirmError && (
        <div
          data-testid="parse-confirm-error"
          style={{
            padding: "8px 14px",
            background: "var(--ei-color-danger-soft)",
            border: "1px solid var(--ei-color-danger)",
            borderRadius: "var(--ei-radius-sm)",
            fontSize: 13,
            color: "var(--ei-color-danger)",
            marginBottom: 12,
          }}
        >
          {confirmError}
        </div>
      )}
      <div
        style={{
          display: "flex",
          flexDirection: compactLayout ? "column" : "row",
          justifyContent: "space-between",
          alignItems: compactLayout ? "stretch" : "center",
          gap: compactLayout ? 14 : 24,
          padding: "16px 0",
          borderTop: "1px solid var(--ei-color-rule-strong)",
        }}
      >
        <div
          style={{
            fontSize: 12,
            color: "var(--ei-color-fg-tertiary)",
            fontFamily: "var(--ei-font-mono)",
            lineHeight: 1.6,
            maxWidth: 420,
          }}
        >
          {t("parse.footerHint")}
        </div>
        <div
          style={{
            display: "flex",
            gap: 10,
            flexWrap: "wrap",
            justifyContent: compactLayout ? "flex-start" : "flex-end",
          }}
        >
          <button
            data-testid="parse-action-cancel"
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
            {t("parse.cancel")}
          </button>
          {!isWorkspaceDetail && (
            <button
              data-testid="parse-action-reparse"
              onClick={handleReparse}
              style={{
                padding: "8px 18px",
                fontSize: 13.5,
                fontFamily: "var(--ei-font-sans)",
                background: "var(--ei-color-bg-soft)",
                border: "1px solid var(--ei-color-rule-strong)",
                borderRadius: "var(--ei-radius-sm)",
                color: "var(--ei-color-fg-primary)",
                cursor: "pointer",
              }}
            >
              {t("parse.reparse")}
            </button>
          )}
          <button
            data-testid="parse-action-save-plan"
            onClick={handleSavePlan}
            disabled={launchDisabled}
            style={{
              padding: "8px 18px",
              fontSize: 13.5,
              fontFamily: "var(--ei-font-sans)",
              background: "var(--ei-color-bg-soft)",
              border: "1px solid var(--ei-color-rule-strong)",
              borderRadius: "var(--ei-radius-sm)",
              color: "var(--ei-color-fg-primary)",
              cursor: launchDisabled ? "not-allowed" : "pointer",
              opacity: launchDisabled ? 0.5 : 1,
              fontWeight: 500,
            }}
          >
            {t("parse.savePlanOnly")}
          </button>
          <button
            data-testid="parse-action-start-interview"
            onClick={handleStartInterview}
            disabled={launchDisabled}
            style={{
              padding: "8px 18px",
              fontSize: 13.5,
              fontFamily: "var(--ei-font-sans)",
              background: "var(--ei-color-accent)",
              border: "none",
              borderRadius: "var(--ei-radius-sm)",
              color: "#fff",
              cursor: launchDisabled ? "not-allowed" : "pointer",
              opacity: launchDisabled ? 0.5 : 1,
              fontWeight: 500,
            }}
          >
            {t("parse.startInterview")}
          </button>
        </div>
      </div>
    </section>
  );
};
