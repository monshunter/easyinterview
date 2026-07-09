import { useCallback, useEffect, useRef, useState, type FC } from "react";

import type { TargetJob } from "../../../api/generated/types";
import { useI18n, type MessageKey } from "../../i18n/messages";
import { useNavigation } from "../../navigation/NavigationProvider";
import { useInterviewContext } from "../../interview-context/InterviewContext";
import { normalizeServerBoundId } from "../../interview-context/apiIds";
import type { Route } from "../../routes";
import { useWorkspaceTargetJob } from "./hooks/useWorkspaceTargetJob";
import { useWorkspaceTargetJobs } from "./hooks/useWorkspaceTargetJobs";
import { useWorkspaceResume } from "./hooks/useWorkspaceResume";
import { useStartPractice } from "./hooks/useStartPractice";
import { useWorkspacePracticePlan } from "./hooks/useWorkspacePracticePlan";
import { WorkspaceInsightCard } from "./WorkspaceInsightCard";
import { PlanSwitcherModal } from "./modals/PlanSwitcherModal";
import { ResumePickerModal } from "./modals/ResumePickerModal";
import { ParseScreen } from "../parse/ParseScreen";
import { useAppRuntimeOptional } from "../../runtime/AppRuntimeProvider";
import { useRequestAuth } from "../../auth/useRequestAuth";

interface WorkspaceScreenProps {
  route: Route;
}

export const WorkspaceScreen: FC<WorkspaceScreenProps> = ({ route }) => {
  const { t, lang } = useI18n();
  const { navigate } = useNavigation();
  const {
    loading,
    data: tj,
    error: targetError,
    empty: targetEmpty,
    notFound: targetNotFound,
    retry: retryTargetJob,
  } = useWorkspaceTargetJob();
  const { data: resume, empty: resumeEmpty } = useWorkspaceResume();
  useWorkspacePracticePlan();
  const { ctx, dispatch } = useInterviewContext();
  const [plannerOpen, setPlannerOpen] = useState(false);
  const [resumePickerOpen, setResumePickerOpen] = useState(false);
  const autoStartRef = useRef(false);
  const compactLayout = useWorkspaceCompactLayout();
  const hasCurrentPlanContext =
    Boolean(route.params.targetJobId) ||
    Boolean(route.params.jobId) ||
    Boolean(route.params.planId) ||
    Boolean(ctx.targetJobId) ||
    Boolean(ctx.jobId) ||
    Boolean(ctx.planId);
  const autoStartRequested =
    route.params.autoStartPractice === "1" || ctx.autoStartPractice === "1";

  // ── Empty / missing states ──
  const hasBoundResume = !!normalizeServerBoundId(ctx.resumeId);
  const showEmptyState = (targetEmpty || targetNotFound) && !loading && !tj;
  const showTargetError = !!targetError && !targetNotFound && !loading && !tj;
  const showMissingResume = !targetEmpty && !loading && tj && (resumeEmpty || !hasBoundResume);

  // ── Derived display values per plan §3.7 mapping ──

  const planEyebrowTitle = tj?.title
    ? `${tj.companyName} · ${tj.title}`
    : t("workspace.planEyebrowTitle");

  const planEyebrowStatus = tj
    ? formatStatus(tj.status, t)
    : t("workspace.planEyebrowStatus");

  const planEyebrowSub = tj
    ? `${roundLabel(route.params.roundId, t) ?? t("workspace.planEyebrowSub")} · ${t("workspace.resumeBound")}`
    : t("workspace.planEyebrowSub");

  const headerTitle = tj?.title ?? t("workspace.headerTitle");
  const headerSubtitle = tj
    ? [tj.companyName, tj.locationText, tj.sourceType ? formatSourceType(tj.sourceType, t) : null]
        .filter(Boolean)
        .join(" · ")
    : t("workspace.headerSubtitle");
  const headerPrepStatus = tj ? derivePrepStatus(tj, t) : t("workspace.headerPrepStatus");
  const headerUpdated = tj
    ? t("workspace.headerUpdated").replace("4/20", formatDate(tj.updatedAt))
    : t("workspace.headerUpdated");
  const headerLevel = tj?.targetLanguage
    ? tj.targetLanguage.toUpperCase()
    : t("workspace.headerLevel");

  const jdTitle = tj?.title ?? t("workspace.jdTitle");
  const jdMeta = tj
    ? [tj.companyName, tj.locationText, tj.sourceType ? formatSourceType(tj.sourceType, t) : null]
        .filter(Boolean)
        .join(" · ")
    : t("workspace.jdMeta");

  const resumeTitle = resume?.title ?? t("workspace.resumeTitle");
  const resumeMeta = resume
    ? readResumeSummary(resume.parsedSummary ?? null)
    : t("workspace.resumeMeta");

  const statusTone = tj ? getStatusTone(tj.status) : "amber";

  const { state: startState, start: doStart } = useStartPractice();
  const runtime = useAppRuntimeOptional();
  const requestAuth = useRequestAuth();

  const navigateToPractice = useCallback(
    (result: { sessionId: string; planId: string }) => {
      navigate({
        name: "practice",
        params: {
          ...route.params,
          sessionId: result.sessionId,
          planId: result.planId,
          targetJobId: ctx.targetJobId,
          jdId: ctx.jdId ?? "",
          resumeId: ctx.resumeId ?? "",
          roundId: ctx.roundId ?? "",
          mode: ctx.mode,
          modality: ctx.modality,
          practiceMode: ctx.practiceMode,
          practiceGoal: ctx.practiceGoal,
          hintUsed: ctx.hintUsed,
          hintCount: ctx.hintCount,
        },
      });
    },
    [ctx, navigate, route.params],
  );

  useEffect(() => {
    if (ctx.autoStartPractice !== "1") return;
    if (autoStartRef.current) return;
    if (loading || !tj) return;
    if (runtime?.auth.status !== "authenticated") return;
    if (!normalizeServerBoundId(ctx.targetJobId) || !normalizeServerBoundId(ctx.resumeId)) return;

    autoStartRef.current = true;
    dispatch({ type: "CLEAR_AUTO_START" });
    void doStart().then((result) => {
      if (result.kind === "success") {
        navigateToPractice(result);
      }
    });
  }, [
    ctx.autoStartPractice,
    ctx.targetJobId,
    ctx.resumeId,
    dispatch,
    doStart,
    loading,
    navigateToPractice,
    runtime?.auth.status,
    tj,
  ]);

  const handleStart = async () => {
    if (runtime?.auth.status !== "authenticated") {
      requestAuth({
        type: "start_practice",
        label: t("workspace.startCore"),
        route: "workspace",
        params: {
          targetJobId: ctx.targetJobId,
          jdId: ctx.jdId ?? "",
          resumeId: ctx.resumeId ?? "",
          roundId: ctx.roundId ?? "",
          planId: ctx.planId ?? "",
          mode: ctx.mode,
          modality: ctx.modality,
          practiceMode: ctx.practiceMode,
          practiceGoal: ctx.practiceGoal,
          hintUsed: ctx.hintUsed,
          hintCount: ctx.hintCount,
          autoStartPractice: "1",
        },
      });
      return;
    }

    const result = await doStart();
    if (result.kind === "success") {
      navigateToPractice(result);
    }
  };

  if (!hasCurrentPlanContext) {
    return <WorkspacePlanList compactLayout={compactLayout} />;
  }

  if (!autoStartRequested) {
    return <ParseScreen route={route} />;
  }

  if (showTargetError) {
    return (
      <div
        className="ei-fadein"
        style={{
          width: "100%",
          maxWidth: 560,
          margin: compactLayout ? "48px auto" : "80px auto",
          padding: compactLayout ? "0 16px" : "0 24px",
        }}
      >
        <div
          data-testid="workspace-target-error"
          style={{
            background: "var(--ei-color-bgCard)",
            border: "1px solid var(--ei-color-rule)",
            borderRadius: 3,
            padding: 32,
            textAlign: "center",
          }}
        >
          <div
            data-testid="workspace-target-error-eyebrow"
            className="ei-label"
            style={{ color: "var(--ei-color-danger)", marginBottom: 8 }}
          >
            {t("workspace.targetError.eyebrow")}
          </div>
          <div
            data-testid="workspace-target-error-title"
            className="ei-serif"
            style={{ fontSize: 18, color: "var(--ei-color-ink)", marginBottom: 12 }}
          >
            {t("workspace.targetError.title")}
          </div>
          <div
            data-testid="workspace-target-error-desc"
            style={{ fontSize: 13, color: "var(--ei-color-ink3)", marginBottom: 20, lineHeight: 1.55 }}
          >
            {t("workspace.targetError.desc")}
          </div>
          <button
            data-testid="workspace-target-error-retry"
            onClick={retryTargetJob}
            style={{
              height: 34,
              padding: "0 16px",
              fontSize: 13,
              fontWeight: 500,
              background: "var(--ei-color-accent)",
              color: "#fff",
              border: "1px solid var(--ei-color-accent)",
              borderRadius: 2,
              cursor: "pointer",
            }}
          >
            {t("workspace.errors.retry")}
          </button>
        </div>
      </div>
    );
  }

  if (showEmptyState) {
    return (
      <div
        className="ei-fadein"
        style={{
          width: "100%",
          maxWidth: 560,
          margin: compactLayout ? "48px auto" : "80px auto",
          padding: compactLayout ? "0 16px" : "0 24px",
        }}
      >
        <div
          data-testid="workspace-empty"
          style={{
            background: "var(--ei-color-bgCard)",
            border: "1px solid var(--ei-color-rule)",
            borderRadius: 3,
            padding: 32,
            textAlign: "center",
          }}
        >
          <div
            data-testid="workspace-empty-eyebrow"
            className="ei-label"
            style={{ color: "var(--ei-color-ink3)", marginBottom: 8 }}
          >
            {t("workspace.empty.eyebrow")}
          </div>
          <div
            data-testid="workspace-empty-title"
            className="ei-serif"
            style={{ fontSize: 18, color: "var(--ei-color-ink)", marginBottom: 12 }}
          >
            {t("workspace.empty.title")}
          </div>
          <div
            data-testid="workspace-empty-desc"
            style={{ fontSize: 13, color: "var(--ei-color-ink3)", marginBottom: 20, lineHeight: 1.55 }}
          >
            {t("workspace.empty.desc")}
          </div>
          <button
            data-testid="workspace-empty-cta"
            onClick={() => navigate({ name: "home", params: {} })}
            style={{
              height: 34,
              padding: "0 16px",
              fontSize: 13,
              fontWeight: 500,
              background: "var(--ei-color-accent)",
              color: "#fff",
              border: "1px solid var(--ei-color-accent)",
              borderRadius: 2,
              cursor: "pointer",
            }}
          >
            {t("workspace.empty.cta")}
          </button>
        </div>
      </div>
    );
  }

  if (showMissingResume) {
    return (
      <div
        className="ei-fadein"
        style={{
          width: "100%",
          maxWidth: 560,
          margin: compactLayout ? "48px auto" : "80px auto",
          padding: compactLayout ? "0 16px" : "0 24px",
        }}
      >
        <div
          data-testid="workspace-missing-resume"
          style={{
            background: "var(--ei-color-bgCard)",
            border: "1px solid var(--ei-color-rule)",
            borderRadius: 3,
            padding: 32,
            textAlign: "center",
          }}
        >
          <div
            data-testid="workspace-missing-resume-eyebrow"
            className="ei-label"
            style={{ color: "var(--ei-color-ink3)", marginBottom: 8 }}
          >
            {t("workspace.missingResume.eyebrow")}
          </div>
          <div
            data-testid="workspace-missing-resume-title"
            className="ei-serif"
            style={{ fontSize: 18, color: "var(--ei-color-ink)", marginBottom: 12 }}
          >
            {t("workspace.missingResume.title")}
          </div>
          <div
            data-testid="workspace-missing-resume-desc"
            style={{ fontSize: 13, color: "var(--ei-color-ink3)", marginBottom: 20, lineHeight: 1.55 }}
          >
            {t("workspace.missingResume.desc")}
          </div>
          <button
            data-testid="workspace-missing-resume-cta"
            onClick={() => navigate({ name: "resume_versions", params: { flow: "create" } })}
            style={{
              height: 34,
              padding: "0 16px",
              fontSize: 13,
              fontWeight: 500,
              background: "var(--ei-color-accent)",
              color: "#fff",
              border: "1px solid var(--ei-color-accent)",
              borderRadius: 2,
              cursor: "pointer",
            }}
          >
            {t("workspace.missingResume.cta")}
          </button>
        </div>
      </div>
    );
  }

  return (
    <div
      className="ei-fadein"
      style={{
        maxWidth: 1280,
        margin: "0 auto",
        padding: compactLayout ? "24px 16px 72px" : "32px 48px 96px",
      }}
    >
      {/* crumbs */}
      <button
        data-testid="workspace-crumbs"
        onClick={() => navigate({ name: "home", params: {} })}
        style={{
          background: "transparent",
          border: "none",
          color: "var(--ei-color-ink3)",
          fontSize: 13,
          display: "flex",
          alignItems: "center",
          gap: 6,
          padding: 0,
          marginBottom: 20,
          cursor: "pointer",
        }}
      >
        <svg
          width={14}
          height={14}
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          strokeWidth={1.5}
        >
          <path d="M19 12H5M11 18l-6-6 6-6" />
        </svg>{" "}
        {t("workspace.crumbs")}
      </button>

      {/* Plan eyebrow */}
      <div
        data-testid="workspace-plan-eyebrow"
        style={{
          background: "var(--ei-color-bgCard)",
          border: "1px solid var(--ei-color-rule)",
          borderRadius: 3,
          padding: "14px 16px",
          marginBottom: 24,
          display: "flex",
          justifyContent: "space-between",
          alignItems: "center",
          gap: 18,
          flexWrap: "wrap",
        }}
      >
        <div style={{ minWidth: compactLayout ? 0 : 280 }}>
          <div
            data-testid="workspace-plan-eyebrow-label"
            className="ei-label"
            style={{
              color: "var(--ei-color-ink3)",
              marginBottom: 5,
            }}
          >
            {t("workspace.planEyebrow")}
          </div>
          <div
            style={{
              display: "flex",
              alignItems: "center",
              gap: 10,
              flexWrap: "wrap",
            }}
          >
            <div
              data-testid="workspace-plan-eyebrow-title"
              className="ei-serif"
              style={{ fontSize: 18, color: "var(--ei-color-ink)" }}
            >
              {planEyebrowTitle}
            </div>
            <span
              data-testid="workspace-plan-eyebrow-status"
              className="ei-mono"
              style={{
                display: "inline-flex",
                alignItems: "center",
                gap: 4,
                padding: "3px 8px",
                borderRadius: 3,
                fontSize: 11.5,
                letterSpacing: "0.04em",
                background:
                  statusTone === "amber"
                    ? "var(--ei-color-amberSoft)"
                    : statusTone === "muted"
                      ? "var(--ei-color-bgSoft)"
                      : "transparent",
                color:
                  statusTone === "amber"
                    ? "var(--ei-color-warn)"
                    : statusTone === "muted"
                      ? "var(--ei-color-ink3)"
                      : "var(--ei-color-ink2)",
                border:
                  statusTone === "neutral"
                    ? "1px solid var(--ei-color-rule)"
                    : "1px solid transparent",
                whiteSpace: "nowrap",
              }}
            >
              {planEyebrowStatus}
            </span>
          </div>
          <div
            data-testid="workspace-plan-eyebrow-sub"
            style={{
              fontSize: 12.5,
              color: "var(--ei-color-ink3)",
              marginTop: 5,
              lineHeight: 1.55,
            }}
          >
            {planEyebrowSub}
          </div>
        </div>
        <div style={{ display: "flex", gap: 10, alignItems: "center" }}>
          <button
            data-testid="workspace-plan-action-switch"
            onClick={() => setPlannerOpen(true)}
            style={{
              display: "inline-flex",
              alignItems: "center",
              justifyContent: "center",
              gap: 8,
              height: 30,
              padding: "0 12px",
              fontSize: 13,
              fontWeight: 500,
              background: "var(--ei-color-bg)",
              color: "var(--ei-color-ink)",
              border: "1px solid var(--ei-color-rule)",
              borderRadius: 2,
              cursor: "pointer",
              fontFamily: "var(--ei-sans)",
            }}
          >
            {t("workspace.switchPlan")}
          </button>
          <button
            data-testid="workspace-plan-action-create"
            onClick={() => navigate({ name: "home", params: {} })}
            style={{
              display: "inline-flex",
              alignItems: "center",
              justifyContent: "center",
              gap: 8,
              height: 30,
              padding: "0 12px",
              fontSize: 13,
              fontWeight: 500,
              background: "transparent",
              color: "var(--ei-color-ink2)",
              border: "1px solid transparent",
              borderRadius: 2,
              cursor: "pointer",
              fontFamily: "var(--ei-sans)",
            }}
          >
            {t("workspace.createPlan")}
          </button>
        </div>
      </div>

      {/* Header summary */}
      <div
        data-testid="workspace-header"
        style={{
          display: "flex",
          justifyContent: "space-between",
          alignItems: "flex-start",
          gap: 24,
          flexWrap: "wrap",
          marginBottom: 32,
        }}
      >
        <div style={{ flex: 1, minWidth: compactLayout ? 0 : 320 }}>
          <div
            style={{
              display: "flex",
              gap: 8,
              alignItems: "center",
              marginBottom: 10,
            }}
          >
            <span
              data-testid="workspace-header-tag"
              className="ei-mono"
              style={{
                display: "inline-flex",
                alignItems: "center",
                gap: 4,
                padding: "3px 8px",
                borderRadius: 3,
                fontSize: 11.5,
                letterSpacing: "0.04em",
                background:
                  statusTone === "amber"
                    ? "var(--ei-color-amberSoft)"
                    : statusTone === "muted"
                      ? "var(--ei-color-bgSoft)"
                      : "transparent",
                color:
                  statusTone === "amber"
                    ? "var(--ei-color-warn)"
                    : statusTone === "muted"
                      ? "var(--ei-color-ink3)"
                      : "var(--ei-color-ink2)",
                border:
                  statusTone === "neutral"
                    ? "1px solid var(--ei-color-rule)"
                    : "1px solid transparent",
                whiteSpace: "nowrap",
              }}
            >
              {planEyebrowStatus}
            </span>
            <span
              data-testid="workspace-header-level"
              className="ei-mono"
              style={{
                display: "inline-flex",
                alignItems: "center",
                gap: 4,
                padding: "3px 8px",
                borderRadius: 3,
                fontSize: 11.5,
                letterSpacing: "0.04em",
                background: "transparent",
                color: "var(--ei-color-ink3)",
                border: "1px solid var(--ei-color-rule)",
                whiteSpace: "nowrap",
              }}
            >
              {headerLevel}
            </span>
            <span
              data-testid="workspace-header-updated"
              className="ei-mono"
              style={{
                fontSize: 12,
                color: "var(--ei-color-ink3)",
              }}
            >
              {headerUpdated}
            </span>
          </div>
          <h1
            data-testid="workspace-header-title"
            className="ei-serif"
            style={{
              fontSize: 38,
              color: "var(--ei-color-ink)",
              margin: 0,
              letterSpacing: "-0.02em",
              lineHeight: 1.15,
            }}
          >
            {headerTitle}
          </h1>
          <div
            data-testid="workspace-header-subtitle"
            style={{
              fontSize: 15,
              color: "var(--ei-color-ink2)",
              marginTop: 6,
            }}
          >
            {headerSubtitle}
          </div>
        </div>
        <div
          style={{
            minWidth: compactLayout ? 0 : 168,
            textAlign: compactLayout ? "left" : "right",
            paddingTop: 4,
          }}
        >
          <div
            className="ei-label"
            style={{
              color: "var(--ei-color-ink3)",
              marginBottom: 6,
            }}
          >
            {t("workspace.prepStatus")}
          </div>
          <div
            data-testid="workspace-header-prep"
            className="ei-serif"
            style={{
              fontSize: 22,
              color: "var(--ei-color-ink)",
              marginBottom: 8,
            }}
          >
            {headerPrepStatus}
          </div>
        </div>
      </div>

      {/* Interview Launcher */}
      <div
        data-testid="workspace-launcher"
        style={{
          background: "var(--ei-color-bgCard)",
          border: "1px solid var(--ei-color-rule)",
          borderRadius: 3,
          padding: 22,
          marginBottom: 32,
        }}
      >
        {/* Round Rail */}
        <div data-testid="workspace-round-rail">
          <div
            style={{
              display: "flex",
              justifyContent: "space-between",
              gap: 16,
              alignItems: "baseline",
              marginBottom: 12,
            }}
          >
            <div
              className="ei-label"
              style={{ color: "var(--ei-color-ink3)" }}
            >
              {t("workspace.flow")}
            </div>
          </div>
          <div style={{ position: "relative" }}>
            <div
              style={{
                position: "absolute",
                top: 13,
                left: 13,
                right: 13,
                height: 1,
                background: "var(--ei-color-rule)",
              }}
            />
            <div
              style={{
                display: "grid",
                gridTemplateColumns: `repeat(${ROUND_FALLBACK_KEYS.length}, 1fr)`,
                alignItems: "start",
              }}
            >
              {ROUND_FALLBACK_KEYS.map((nameKey, i) => (
                <div
                  key={i}
                  style={{
                    position: "relative",
                    display: "flex",
                    flexDirection: "column",
                    alignItems:
                      i === 0
                        ? "flex-start"
                        : i === ROUND_FALLBACK_KEYS.length - 1
                          ? "flex-end"
                          : "center",
                    minHeight: 72,
                  }}
                >
                  <div
                    style={{
                      width: 26,
                      height: 26,
                      borderRadius: 13,
                      border:
                        "1px solid var(--ei-color-rule)",
                      background: "var(--ei-color-bgCard)",
                      color: "var(--ei-color-ink3)",
                      display: "flex",
                      alignItems: "center",
                      justifyContent: "center",
                      zIndex: 1,
                    }}
                  >
                    <span
                      className="ei-mono"
                      style={{ fontSize: 11 }}
                    >
                      {i + 1}
                    </span>
                  </div>
                  <div
                    style={{
                      fontSize: 12.5,
                      color: "var(--ei-color-ink3)",
                      marginTop: 8,
                      textAlign:
                        i === 0
                          ? "left"
                          : i === ROUND_FALLBACK_KEYS.length - 1
                            ? "right"
                            : "center",
                      maxWidth: 140,
                    }}
                  >
                    {t(nameKey)}
                  </div>
                  <div
                    style={{
                      fontSize: 11,
                      color: "var(--ei-color-ink4)",
                      marginTop: 3,
                      textAlign:
                        i === 0
                          ? "left"
                          : i === ROUND_FALLBACK_KEYS.length - 1
                            ? "right"
                            : "center",
                      maxWidth: 140,
                      lineHeight: 1.35,
                    }}
                  >
                    {t(`workspace.roundState${i + 1}` as any)}
                  </div>
                </div>
              ))}
            </div>
          </div>
        </div>

        <div
          style={{
            display: "flex",
            justifyContent: "space-between",
            alignItems: "flex-start",
            marginTop: 22,
            marginBottom: 18,
            gap: 20,
            flexWrap: "wrap",
          }}
        >
          <div>
            <div
              className="ei-label"
              style={{
                color: "var(--ei-color-ink3)",
                marginBottom: 4,
              }}
            >
              {t("workspace.launchLabel")}
            </div>
            <div
              className="ei-serif"
              style={{
                fontSize: 21,
                color: "var(--ei-color-ink)",
              }}
            >
              {t("workspace.launchTitle")}
            </div>
            <div
              style={{
                fontSize: 13.5,
                color: "var(--ei-color-ink2)",
                marginTop: 6,
              }}
            >
              {t("workspace.roundStatus")}
            </div>
            <div
              style={{
                fontSize: 13.5,
                color: "var(--ei-color-ink3)",
                marginTop: 6,
                lineHeight: 1.6,
                maxWidth: 680,
              }}
            >
              {t("workspace.launchSub")}
            </div>
          </div>
          <button
            data-testid="workspace-cta-start"
            onClick={handleStart}
            disabled={startState.kind === "loading"}
            style={{
              display: "inline-flex",
              alignItems: "center",
              justifyContent: "center",
              gap: 8,
              height: 38,
              padding: "0 16px",
              fontSize: 14,
              fontWeight: 500,
              background: startState.kind === "loading"
                ? "var(--ei-color-ink4)"
                : "var(--ei-color-accent)",
              color: "#fff",
              border: "1px solid var(--ei-color-accent)",
              borderRadius: 2,
              cursor: startState.kind === "loading" ? "default" : "pointer",
              fontFamily: "var(--ei-sans)",
              opacity: startState.kind === "loading" ? 0.7 : 1,
            }}
          >
            {startState.kind === "loading" ? (
              <>{t("workspace.placeholder")}</>
            ) : (
              <>
                <svg
                  width={16}
                  height={16}
                  viewBox="0 0 24 24"
                  fill="#fff"
                  stroke="none"
                >
                  <path d="M7 5l12 7-12 7V5z" />
                </svg>
                {t("workspace.startCore")}
              </>
            )}
          </button>

          {startState.kind === "error" && (
            <div
              data-testid="workspace-cta-error"
              style={{
                marginTop: 12,
                fontSize: 13,
                color: "var(--ei-color-danger)",
              }}
            >
              {startState.message}
              {startState.retryable && (
                <button
                  data-testid="workspace-cta-retry"
                  onClick={handleStart}
                  style={{
                    marginLeft: 10,
                    height: 28,
                    padding: "0 10px",
                    fontSize: 12,
                    background: "transparent",
                    color: "var(--ei-color-ink2)",
                    border: "1px solid var(--ei-color-rule)",
                    borderRadius: 2,
                    cursor: "pointer",
                  }}
                >
                  {t("workspace.errors.retry")}
                </button>
              )}
            </div>
          )}

          {startState.kind === "error" && !startState.retryable && (
            <button
              data-testid="workspace-cta-back-home"
              onClick={() => navigate({ name: "home", params: {} })}
              style={{
                marginTop: 12,
                height: 28,
                padding: "0 12px",
                fontSize: 12,
                background: "transparent",
                color: "var(--ei-color-ink2)",
                border: "1px solid var(--ei-color-rule)",
                borderRadius: 2,
                cursor: "pointer",
              }}
            >
              {t("workspace.errors.backHome")}
            </button>
          )}
        </div>

        <div
          style={{
            display: "grid",
            gridTemplateColumns: compactLayout
              ? "1fr"
              : "minmax(0, 1fr) minmax(0, 1fr)",
            gap: 12,
          }}
        >
          {/* JD BindingPill */}
          <div
            data-testid="workspace-binding-jd"
            style={{
              padding: "14px 16px",
              background: "var(--ei-color-bgSoft)",
              border: "1px solid var(--ei-color-rule)",
              borderRadius: 2,
              display: "grid",
              gridTemplateColumns: "32px minmax(0, 1fr)",
              gap: 12,
              alignItems: "center",
            }}
          >
            <div
              style={{
                width: 32,
                height: 32,
                borderRadius: 16,
                background: "var(--ei-color-bgCard)",
                border: "1px solid var(--ei-color-rule)",
                color: "var(--ei-color-accent)",
                display: "flex",
                alignItems: "center",
                justifyContent: "center",
              }}
            >
              <svg
                width={15}
                height={15}
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                strokeWidth={1.5}
              >
                <rect x="3" y="7" width="18" height="13" rx="1" />
                <path d="M8 7V4h8v3M3 12h18" />
              </svg>
            </div>
            <div style={{ minWidth: 0 }}>
              <div
                className="ei-label"
                style={{
                  color: "var(--ei-color-ink3)",
                  marginBottom: 3,
                }}
              >
                {t("workspace.jdBound")}
              </div>
              <div
                style={{
                  fontSize: 14,
                  color: "var(--ei-color-ink)",
                  fontWeight: 500,
                  whiteSpace: "nowrap",
                  overflow: "hidden",
                  textOverflow: "ellipsis",
                }}
              >
                {jdTitle}
              </div>
              <div
                style={{
                  fontSize: 12,
                  color: "var(--ei-color-ink3)",
                  marginTop: 2,
                  whiteSpace: "nowrap",
                  overflow: "hidden",
                  textOverflow: "ellipsis",
                }}
              >
                {jdMeta}
              </div>
            </div>
          </div>

          {/* Resume BindingPill */}
          <div
            data-testid="workspace-binding-resume"
            style={{
              padding: "14px 16px",
              background: "var(--ei-color-bgSoft)",
              border: "1px solid var(--ei-color-rule)",
              borderRadius: 2,
              display: "grid",
              gridTemplateColumns: compactLayout
                ? "32px minmax(0, 1fr)"
                : "32px minmax(0, 1fr) auto",
              gap: 12,
              alignItems: "center",
            }}
          >
            <div
              style={{
                width: 32,
                height: 32,
                borderRadius: 16,
                background: "var(--ei-color-bgCard)",
                border: "1px solid var(--ei-color-rule)",
                color: "var(--ei-color-accent)",
                display: "flex",
                alignItems: "center",
                justifyContent: "center",
              }}
            >
              <svg
                width={15}
                height={15}
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                strokeWidth={1.5}
              >
                <path d="M7 3h8l4 4v14H7V3zM15 3v5h5M9 12h8M9 16h6M9 8h3" />
              </svg>
            </div>
              <div style={{ minWidth: 0 }}>
                <div
                  className="ei-label"
                  style={{
                    color: "var(--ei-color-ink3)",
                    marginBottom: 3,
                  }}
                >
                  {t("workspace.resumeBound")}
                </div>
                <div
                  style={{
                    fontSize: 14,
                    color: "var(--ei-color-ink)",
                    fontWeight: 500,
                    whiteSpace: "nowrap",
                    overflow: "hidden",
                    textOverflow: "ellipsis",
                  }}
                >
                  {resumeTitle}
                </div>
                <div
                  style={{
                    fontSize: 12,
                    color: "var(--ei-color-ink3)",
                    marginTop: 2,
                    whiteSpace: "nowrap",
                    overflow: "hidden",
                    textOverflow: "ellipsis",
                  }}
                >
                  {resumeMeta}
                </div>
              </div>
            <button
              data-testid="workspace-binding-resume-change"
              onClick={() => setResumePickerOpen(true)}
              style={{
                gridColumn: compactLayout ? "1 / -1" : undefined,
                justifySelf: compactLayout ? "start" : undefined,
                background: "transparent",
                border: "1px solid var(--ei-color-rule)",
                borderRadius: 2,
                color: "var(--ei-color-ink2)",
                padding: "5px 10px",
                fontSize: 12,
                cursor: "pointer",
              }}
            >
              {t("workspace.changeResume")}
            </button>
          </div>
        </div>

        <div
          data-testid="workspace-note-practice"
          style={{
            fontSize: 12,
            color: "var(--ei-color-ink3)",
            marginTop: 12,
            display: "flex",
            gap: 6,
            alignItems: "center",
          }}
        >
          <svg
            width={12}
            height={12}
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            strokeWidth={1.5}
          >
            <circle cx="12" cy="12" r="9" />
            <path d="M12 8v.01M11 12h1v5h1" />
          </svg>{" "}
          {t("workspace.notePractice")}
        </div>
      </div>

      {/* 2-column Main */}
      <div
        style={{
          display: "grid",
          gridTemplateColumns: compactLayout
            ? "1fr"
            : "minmax(0, 1.4fr) minmax(0, 1fr)",
          gap: 24,
        }}
      >
        {/* Left column */}
        <div
          style={{
            display: "flex",
            flexDirection: "column",
            gap: 24,
            minWidth: 0,
          }}
        >
          {/* WorkspaceInsightCard */}
          <WorkspaceInsightCard
            companyName={tj?.companyName}
            locationText={tj?.locationText}
            sourceType={tj?.sourceType}
            summary={typeof tj?.summary?.coreThemes?.[0] === "string" ? tj.summary.coreThemes[0] : undefined}
            targetJobId={ctx.targetJobId || (route.params.targetJobId as string)}
            jdId={ctx.jdId || (route.params.jdId as string)}
          />

          {/* JD breakdown card */}
          <div
            data-testid="workspace-jd-card"
            style={{
              background: "var(--ei-color-bgCard)",
              border: "1px solid var(--ei-color-rule)",
              borderRadius: 3,
            }}
          >
            <div
              style={{
                padding: "16px 20px",
                borderBottom: "1px solid var(--ei-color-rule)",
                display: "flex",
                justifyContent: "space-between",
                alignItems: "center",
              }}
            >
              <div>
                <div
                  className="ei-label"
                  style={{
                    color: "var(--ei-color-ink3)",
                    marginBottom: 2,
                  }}
                >
                  {t("workspace.jdCardLabel")}
                </div>
                <div
                  className="ei-serif"
                  style={{
                    fontSize: 17,
                    color: "var(--ei-color-ink)",
                  }}
                >
                  {t("workspace.requirements")}
                </div>
              </div>
            </div>
            <div style={{ padding: 20 }}>
              {JD_BLOCKS.map((block) => {
                const items = tj
                  ? tj.requirements.filter((r) => r.kind === block.kind)
                  : [];
                return (
                  <div
                    key={block.kind}
                    data-testid={`workspace-jd-block-${block.testSuffix}`}
                    style={{ marginTop: block.kind === "must_have" ? 0 : 18 }}
                  >
                    <div
                      style={{
                        display: "flex",
                        alignItems: "center",
                        gap: 8,
                        marginBottom: 8,
                      }}
                    >
                      <span
                        className="ei-mono"
                        style={{
                          display: "inline-flex",
                          alignItems: "center",
                          gap: 4,
                          padding: "3px 8px",
                          borderRadius: 3,
                          fontSize: 11.5,
                          letterSpacing: "0.04em",
                          background: block.tagBg,
                          color: block.tagColor,
                          whiteSpace: "nowrap",
                        }}
                      >
                        {t(block.labelKey)}
                      </span>
                    </div>
                    {items.length === 0 ? (
                      <div
                        style={{
                          fontSize: 13.5,
                          color: "var(--ei-color-ink2)",
                        }}
                      >
                        <span>○</span> {t("workspace.placeholder")}
                      </div>
                    ) : (
                      items.map((r) => (
                        <div
                          key={r.id}
                          style={{
                            fontSize: 13.5,
                            color: "var(--ei-color-ink2)",
                            padding: "3px 0",
                          }}
                        >
                          <span>○</span> {r.label}
                        </div>
                      ))
                    )}
                  </div>
                );
              })}
            </div>
          </div>
        </div>

        {/* Right column */}
        <div
          style={{
            display: "flex",
            flexDirection: "column",
            gap: 24,
            minWidth: 0,
          }}
        >
          {/* Risks & strengths */}
          <div
            data-testid="workspace-prep-card"
            style={{
              background: "var(--ei-color-bgCard)",
              border: "1px solid var(--ei-color-rule)",
              borderRadius: 3,
              padding: 20,
            }}
          >
            <div
              className="ei-label"
              style={{
                color: "var(--ei-color-ink3)",
                marginBottom: 14,
              }}
            >
              {t("workspace.prep")}
            </div>
            <div
              data-testid="workspace-prep-strongs"
              style={{ marginBottom: 14 }}
            >
              <div
                style={{
                  fontSize: 12.5,
                  color: "var(--ei-color-ok)",
                  fontWeight: 500,
                  marginBottom: 6,
                }}
              >
                ● {t("workspace.strongs")}
              </div>
              {tj?.fitSummary?.strengths?.length ? (
                tj.fitSummary.strengths.map((s, i) => (
                  <div
                    key={i}
                    data-testid={`workspace-prep-strong-${i}`}
                    style={{
                      fontSize: 13,
                      color: "var(--ei-color-ink2)",
                      padding: "4px 0",
                    }}
                  >
                    {s}
                  </div>
                ))
              ) : (
                <div
                  style={{
                    fontSize: 13,
                    color: "var(--ei-color-ink2)",
                    padding: "4px 0",
                  }}
                >
                  {t("workspace.placeholder")}
                </div>
              )}
            </div>
            <div data-testid="workspace-prep-risks">
              <div
                style={{
                  fontSize: 12.5,
                  color: "var(--ei-color-danger)",
                  fontWeight: 500,
                  marginBottom: 6,
                }}
              >
                ● {t("workspace.risks")}
              </div>
              {tj?.fitSummary &&
              (tj.fitSummary.riskSignals?.length || tj.fitSummary.gaps?.length) ? (
                [
                  ...(tj.fitSummary.riskSignals ?? []),
                  ...(tj.fitSummary.gaps ?? []),
                ].map((r, i) => (
                  <div
                    key={i}
                    data-testid={`workspace-prep-risk-${i}`}
                    style={{
                      fontSize: 13,
                      color: "var(--ei-color-ink2)",
                      padding: "4px 0",
                    }}
                  >
                    {r}
                  </div>
                ))
              ) : (
                <div
                  style={{
                    fontSize: 13,
                    color: "var(--ei-color-ink2)",
                    padding: "4px 0",
                  }}
                >
                  {t("workspace.placeholder")}
                </div>
              )}
            </div>
          </div>

          {/* Current-plan records placeholder */}
          <div
            data-testid="workspace-history-card"
            style={{
              background: "var(--ei-color-bgCard)",
              border: "1px solid var(--ei-color-rule)",
              borderRadius: 3,
            }}
          >
            <div
              style={{
                padding: "16px 20px",
                borderBottom:
                  "1px solid var(--ei-color-rule)",
                display: "flex",
                justifyContent: "space-between",
                alignItems: "center",
              }}
            >
              <div>
                <div
                  className="ei-label"
                  style={{
                    color: "var(--ei-color-ink3)",
                    marginBottom: 2,
                  }}
                >
                  {t("workspace.historyLabel")}
                </div>
                <div
                  className="ei-serif"
                  style={{
                    fontSize: 17,
                    color: "var(--ei-color-ink)",
                  }}
                >
                  {t("workspace.practices")}
                </div>
              </div>
            </div>
            <div
              data-testid="workspace-history-empty"
              style={{
                padding: "24px 20px",
                fontSize: 13,
                color: "var(--ei-color-ink3)",
                textAlign: "center",
              }}
            >
              {t("workspace.historyEmpty")}
            </div>
          </div>
        </div>
      </div>
      <PlanSwitcherModal
        open={plannerOpen}
        onClose={() => setPlannerOpen(false)}
        onSelectPlan={(targetJobId) => {
          dispatch({
            type: "HYDRATE_FROM_ROUTE",
            params: {
              targetJobId,
              jobId: targetJobId,
              jdId: `jd-${targetJobId}`,
              planId: "",
              resumeId: ctx.resumeId ?? "",
              roundId: ctx.roundId ?? "round-technical-1",
              roundName: ctx.roundName ?? "",
              practiceMode: ctx.practiceMode,
              practiceGoal: ctx.practiceGoal,
              mode: ctx.mode,
              modality: ctx.modality,
              hintUsed: ctx.hintUsed,
              hintCount: ctx.hintCount,
            },
          });
          setPlannerOpen(false);
        }}
      />
      <ResumePickerModal
        open={resumePickerOpen}
        onClose={() => setResumePickerOpen(false)}
        boundResumeId={ctx.resumeId}
        onSelectResume={(resumeId) => {
          dispatch({ type: "MERGE_RESUME", resume: { id: resumeId } });
        }}
      />
    </div>
  );
};

interface WorkspacePlanListProps {
  compactLayout: boolean;
}

const WorkspacePlanList: FC<WorkspacePlanListProps> = ({ compactLayout }) => {
  const { t } = useI18n();
  const { navigate } = useNavigation();
  const { loading, jobs, error } = useWorkspaceTargetJobs();

  const openPlan = (job: TargetJob) => {
    const currentPracticePlanId = job.currentPracticePlanId?.trim();
    const resumeId = job.resumeId?.trim();
    navigate({
      name: "workspace",
      params: {
        targetJobId: job.id,
        jobId: job.id,
        jdId: `jd-${job.id}`,
        ...(currentPracticePlanId ? { planId: currentPracticePlanId } : {}),
        ...(resumeId ? { resumeId } : {}),
      },
    });
  };

  return (
    <div
      data-testid="workspace-plan-list"
      className="ei-fadein"
      style={{
        maxWidth: 1120,
        margin: "0 auto",
        padding: compactLayout ? "32px 16px 72px" : "48px 48px 96px",
      }}
    >
      <div
        style={{
          display: "flex",
          justifyContent: "space-between",
          alignItems: "flex-start",
          gap: 24,
          flexWrap: "wrap",
          marginBottom: 28,
        }}
      >
        <div style={{ maxWidth: 640 }}>
          <div
            data-testid="workspace-plan-list-eyebrow"
            className="ei-label"
            style={{ color: "var(--ei-color-fg-tertiary)", marginBottom: 8 }}
          >
            {t("workspace.planList.eyebrow")}
          </div>
          <h1
            data-testid="workspace-plan-list-title"
            className="ei-serif"
            style={{
              fontSize: compactLayout ? 30 : 40,
              color: "var(--ei-color-fg-primary)",
              margin: 0,
              lineHeight: 1.14,
            }}
          >
            {t("workspace.planList.title")}
          </h1>
          <div
            data-testid="workspace-plan-list-subtitle"
            style={{
              fontSize: 14,
              color: "var(--ei-color-fg-secondary)",
              marginTop: 10,
              lineHeight: 1.6,
            }}
          >
            {t("workspace.planList.subtitle")}
          </div>
        </div>
        <button
          data-testid="workspace-plan-list-create"
          type="button"
          onClick={() => navigate({ name: "home", params: {} })}
          style={{
            height: 34,
            padding: "0 16px",
            fontSize: 13,
            fontWeight: 500,
            background: "var(--ei-color-accent)",
            color: "#fff",
            border: "1px solid var(--ei-color-accent)",
            borderRadius: 2,
            cursor: "pointer",
            fontFamily: "var(--ei-sans)",
          }}
        >
          {t("workspace.planList.create")}
        </button>
      </div>

      {loading ? (
        <div
          data-testid="workspace-plan-list-loading"
          style={{
            background: "var(--ei-color-bg-card)",
            border: "1px solid var(--ei-color-rule-strong)",
            borderRadius: 3,
            padding: 24,
            color: "var(--ei-color-fg-tertiary)",
            fontSize: 13,
          }}
        >
          {t("workspace.planList.loading")}
        </div>
      ) : error ? (
        <div
          data-testid="workspace-plan-list-error"
          style={{
            background: "var(--ei-color-bg-card)",
            border: "1px solid var(--ei-color-rule-strong)",
            borderRadius: 3,
            padding: 24,
            color: "var(--ei-color-fg-tertiary)",
            fontSize: 13,
          }}
        >
          {t("workspace.planList.error")}
        </div>
      ) : jobs.length === 0 ? (
        <div
          data-testid="workspace-plan-list-empty"
          style={{
            background: "var(--ei-color-bg-card)",
            border: "1px solid var(--ei-color-rule-strong)",
            borderRadius: 3,
            padding: 32,
            textAlign: "center",
          }}
        >
          <div
            className="ei-serif"
            style={{ fontSize: 18, color: "var(--ei-color-fg-primary)", marginBottom: 10 }}
          >
            {t("workspace.planList.emptyTitle")}
          </div>
          <div style={{ fontSize: 13, color: "var(--ei-color-fg-tertiary)", lineHeight: 1.55 }}>
            {t("workspace.planList.emptyDesc")}
          </div>
        </div>
      ) : (
        <div
          data-testid="workspace-plan-list-grid"
          style={{
            display: "grid",
            gridTemplateColumns: compactLayout
              ? "minmax(0, 1fr)"
              : "repeat(auto-fit, minmax(300px, 1fr))",
            gap: 16,
            alignItems: "stretch",
          }}
        >
          {jobs.map((job) => {
            const statusTone = getStatusTone(job.status);
            return (
              <article
                key={job.id}
                data-testid={`workspace-plan-list-card-${job.id}`}
                style={{
                  background: "var(--ei-color-bg-card)",
                  border: "1px solid var(--ei-color-rule-strong)",
                  borderRadius: 3,
                  boxShadow: "var(--ei-shadow-elev2)",
                  minHeight: 178,
                  display: "flex",
                  flexDirection: "column",
                  justifyContent: "space-between",
                  overflow: "hidden",
                }}
              >
                <div
                  data-testid={`workspace-plan-list-card-body-${job.id}`}
                  style={{
                    padding: 20,
                    flex: 1,
                    background: "var(--ei-color-bg-card)",
                  }}
                >
                  <div
                    style={{
                      display: "flex",
                      justifyContent: "space-between",
                      gap: 12,
                      alignItems: "center",
                      marginBottom: 10,
                    }}
                  >
                    <span
                      className="ei-mono"
                      style={{
                        display: "inline-flex",
                        alignItems: "center",
                        padding: "3px 8px",
                        borderRadius: 3,
                        fontSize: 11.5,
                        letterSpacing: "0.04em",
                        background:
                          statusTone === "amber"
                            ? "var(--ei-color-amber-soft)"
                            : statusTone === "muted"
                              ? "var(--ei-color-bg-soft)"
                              : "transparent",
                        color:
                          statusTone === "amber"
                            ? "var(--ei-color-warn)"
                            : statusTone === "muted"
                              ? "var(--ei-color-fg-tertiary)"
                              : "var(--ei-color-fg-secondary)",
                        border:
                          statusTone === "neutral"
                            ? "1px solid var(--ei-color-rule-strong)"
                            : "1px solid transparent",
                        whiteSpace: "nowrap",
                      }}
                    >
                      {formatStatus(job.status, t)}
                    </span>
                    <span
                      className="ei-mono"
                      style={{ fontSize: 12, color: "var(--ei-color-fg-tertiary)" }}
                    >
                      {t("workspace.planList.updated").replace("{date}", formatDate(job.updatedAt))}
                    </span>
                  </div>
                  <div
                    className="ei-serif"
                    style={{
                      fontSize: 20,
                      color: "var(--ei-color-fg-primary)",
                      lineHeight: 1.25,
                      marginBottom: 6,
                    }}
                  >
                    {job.title}
                  </div>
                  <div
                    style={{
                      fontSize: 13,
                      color: "var(--ei-color-fg-secondary)",
                      lineHeight: 1.5,
                    }}
                  >
                    {[job.companyName, job.locationText].filter(Boolean).join(" · ")}
                  </div>
                </div>
                <div
                  data-testid={`workspace-plan-list-card-footer-${job.id}`}
                  style={{
                    borderTop: "1px solid var(--ei-color-rule-strong)",
                    padding: "14px 20px",
                    background: "var(--ei-color-bg-card)",
                    display: "flex",
                    justifyContent: "flex-end",
                    alignItems: "center",
                    gap: 12,
                  }}
                >
                  <button
                    data-testid={`workspace-plan-list-open-${job.id}`}
                    type="button"
                    onClick={() => openPlan(job)}
                    style={{
                      flex: "0 0 auto",
                      height: 32,
                      padding: "0 12px",
                      fontSize: 13,
                      fontWeight: 500,
                      background: "var(--ei-color-accent)",
                      color: "#fff",
                      border: "1px solid var(--ei-color-accent)",
                      borderRadius: 2,
                      cursor: "pointer",
                      fontFamily: "var(--ei-sans)",
                    }}
                  >
                    {t("workspace.planList.open")}
                  </button>
                </div>
              </article>
            );
          })}
        </div>
      )}
    </div>
  );
};

// ── Phase 2 helpers (colocated) ──

const JD_BLOCKS: {
  kind: "must_have" | "nice_to_have" | "hidden_signal";
  labelKey: MessageKey;
  testSuffix: string;
  tagBg: string;
  tagColor: string;
}[] = [
  {
    kind: "must_have",
    labelKey: "workspace.must",
    testSuffix: "must",
    tagBg: "var(--ei-color-accentSoft)",
    tagColor: "var(--ei-color-accent)",
  },
  {
    kind: "nice_to_have",
    labelKey: "workspace.nice",
    testSuffix: "nice",
    tagBg: "var(--ei-color-amberSoft)",
    tagColor: "var(--ei-color-warn)",
  },
  {
    kind: "hidden_signal",
    labelKey: "workspace.hidden",
    testSuffix: "hidden",
    tagBg: "var(--ei-color-coolSoft)",
    tagColor: "var(--ei-color-cool)",
  },
];

const ROUND_FALLBACK_KEYS = [
  "workspace.roundName1",
  "workspace.roundName2",
  "workspace.roundName3",
  "workspace.roundName4",
] as const satisfies readonly MessageKey[];
const ROUND_IDS = ["round-hr", "round-tech1", "round-tech2", "round-manager"] as const;

type StatusTone = "amber" | "muted" | "neutral";

function getStatusTone(status: string): StatusTone {
  switch (status) {
    case "applied":
    case "interviewing":
      return "amber";
    case "draft":
    case "preparing":
      return "muted";
    default:
      return "neutral";
  }
}

function formatStatus(status: string, t: (key: MessageKey) => string): string {
  const map: Record<string, MessageKey> = {
    draft: "workspace.status.draft",
    preparing: "workspace.status.preparing",
    applied: "workspace.status.applied",
    interviewing: "workspace.status.interviewing",
    offer: "workspace.status.offer",
    rejected: "workspace.status.rejected",
    archived: "workspace.status.archived",
  };
  const key = map[status];
  return key ? t(key) : status;
}

function formatDate(iso: string): string {
  try {
    const d = new Date(iso);
    return `${d.getMonth() + 1}/${d.getDate()}`;
  } catch {
    return "—";
  }
}

function formatSourceType(s: string, t: (key: MessageKey) => string): string {
  const map: Record<string, MessageKey> = {
    manual_text: "workspace.source.manualText",
    url: "workspace.source.url",
    file: "workspace.source.file",
    manual_form: "workspace.source.manualForm",
  };
  const key = map[s];
  return key ? t(key) : s;
}

function roundLabel(roundId: string | undefined, t: (key: MessageKey) => string): string | null {
  if (!roundId) return null;
  const idx = ROUND_IDS.indexOf(roundId as typeof ROUND_IDS[number]);
  if (idx >= 0) return t(ROUND_FALLBACK_KEYS[idx]!);
  return null;
}

function readResumeSummary(parsedSummary: Record<string, unknown> | null): string {
  if (!parsedSummary) return "—";
  const parts: string[] = [];
  const headline = typeof parsedSummary.headline === "string" ? parsedSummary.headline : null;
  const yoe = typeof parsedSummary.yearsOfExperience === "number"
    ? `${parsedSummary.yearsOfExperience}y`
    : null;
  if (headline) parts.push(headline);
  if (yoe) parts.push(yoe);
  return parts.length > 0 ? parts.join(" · ") : "—";
}

function derivePrepStatus(
  tj: { fitSummary?: { strengths?: unknown[]; gaps?: unknown[]; riskSignals?: unknown[] } | null; openQuestionIssueCount: number },
  t: (key: MessageKey) => string,
): string {
  const fs = tj.fitSummary;
  if (!fs) return "—";
  const strongs = (fs.strengths?.length ?? 0);
  const risks = (fs.riskSignals?.length ?? 0);
  const gaps = (fs.gaps?.length ?? 0);
  if (strongs > 0) {
    return t("workspace.prepStatus.hitCount").replace("{count}", String(strongs));
  }
  if (risks > 0 || gaps > 0 || tj.openQuestionIssueCount > 0) {
    return t("workspace.prepStatus.needsWork");
  }
  return "—";
}

function useWorkspaceCompactLayout(): boolean {
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
