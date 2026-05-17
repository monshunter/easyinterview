import {
  useCallback,
  useEffect,
  useMemo,
  useRef,
  useState,
  type Dispatch,
  type FC,
  type SetStateAction,
} from "react";

import type {
  AssistantAction,
  GenerationProvenance,
} from "../../../api/generated/types";
import { useI18n, type MessageKey } from "../../i18n/messages";
import { useInterviewContext } from "../../interview-context/InterviewContext";
import { useNavigation } from "../../navigation/NavigationProvider";
import type { Route } from "../../routes";
import { TopBar } from "./components/TopBar";
import { SessionMap, type SessionMapItem } from "./components/SessionMap";
import { LiveNotes } from "./components/LiveNotes";
import { QuestionCard } from "./components/QuestionCard";
import { Transcript, type TranscriptMessage } from "./components/Transcript";
import { InputBar } from "./components/InputBar";
import { HintBanner } from "./components/HintBanner";
import { RightPanel } from "./components/RightPanel";
import { FinishCta } from "./components/FinishCta";
import { PracticeVoiceSurface } from "./components/PracticeVoiceSurface";
import { PracticeVoiceRightPanel } from "./components/PracticeVoiceRightPanel";
import { PracticeSessionLostState } from "./components/PracticeSessionLostState";
import { ErrorState } from "./components/ErrorState";
import { AssistantActionRenderer } from "./components/AssistantActionRenderer";
import type { InterviewerPersona } from "./components/RoleDropdown";
import { usePracticeSessionLoader } from "./hooks/usePracticeSessionLoader";
import { usePracticeEvents } from "./hooks/usePracticeEvents";
import { usePracticeAssistance } from "./hooks/usePracticeAssistance";
import { usePracticeSession } from "./hooks/usePracticeSession";
import { useCompletePracticeSession } from "./hooks/useCompletePracticeSession";
import { buildPracticeHandoffParams } from "./utils/practiceHandoffParams";

interface PracticeScreenProps {
  route: Route;
}

interface ClassifiedPracticeError {
  messageKey: MessageKey;
  retryable: boolean;
  refreshSession: boolean;
  sessionLost: boolean;
}

interface PracticeErrorState {
  message: string;
  retryable: boolean;
  fallbackBackToWorkspace: boolean;
}

type RetryAction = () => Promise<void>;

const PERSONA_LABEL_KEY: Record<InterviewerPersona, MessageKey> = {
  general: "practice.toolbar.role.general",
  hr: "practice.toolbar.role.hr",
  manager: "practice.toolbar.role.manager",
};

/**
 * PracticeScreen — text-mode mock interview surface.
 *
 * Source-level mirror of `ui-design/src/screen-practice.jsx::PracticeScreen`
 * text branch. Phase 1 landed the static shell; Phase 2 added the event
 * loop hooks; Phase 3 wires them to user interactions: hint / skip /
 * pause-resume / send / role switch / strict-locked toast.
 */
export const PracticeScreen: FC<PracticeScreenProps> = ({ route }) => {
  const { t, lang } = useI18n();
  const { navigate } = useNavigation();
  const { ctx, dispatch } = useInterviewContext();

  const sessionId = route.params.sessionId || ctx.sessionId || "";
  const mode = route.params.mode || ctx.mode || "text";
  const modality = route.params.modality || ctx.modality || mode;
  const practiceMode =
    route.params.practiceMode || ctx.practiceMode || "strict";
  const practiceGoal =
    route.params.practiceGoal || ctx.practiceGoal || "baseline";
  const isStrict = practiceMode === "strict";
  const activeMode = modality === "voice" ? "voice" : "text";

  const loader = usePracticeSessionLoader(sessionId);
  const events = usePracticeEvents(sessionId);
  const completion = useCompletePracticeSession(sessionId);
  const assistance = usePracticeAssistance({
    practiceMode,
    practiceGoal,
  });
  const sessionFlags = usePracticeSession(loader.data?.status ?? null);
  const isNarrow = useNarrowPracticeLayout();

  const [persona, setPersona] = useState<InterviewerPersona>("general");
  const [strictToastOpen, setStrictToastOpen] = useState(false);
  const strictToastTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const [paused, setPaused] = useState(false);
  const [input, setInput] = useState("");
  const [transcript, setTranscript] = useState<TranscriptMessage[]>([]);
  const [showHintBanner, setShowHintBanner] = useState(false);
  const [hintBannerText, setHintBannerText] = useState("");
  const [activeAssistantAction, setActiveAssistantAction] =
    useState<AssistantAction | null>(null);
  const [aiTransparency, setAiTransparency] =
    useState<GenerationProvenance | null>(null);
  const [errorState, setErrorState] = useState<PracticeErrorState | null>(null);
  const retryActionRef = useRef<RetryAction | null>(null);
  const [refreshingAfterConflict, setRefreshingAfterConflict] = useState(false);
  const conflictRefreshStartedRef = useRef(false);
  const [sessionLostByMutation, setSessionLostByMutation] = useState(false);
  type TurnAnnotation = "skipped" | "follow_up_requested" | "done";
  const [turnAnnotations, setTurnAnnotations] = useState<
    Map<number, TurnAnnotation>
  >(() => new Map());
  const inputDisabled =
    sessionFlags.inputDisabled ||
    paused ||
    loader.state === "loading" ||
    refreshingAfterConflict;

  // Local elapsed timer (UI display only; backend owns server elapsed).
  const [elapsed, setElapsed] = useState(0);
  useEffect(() => {
    if (inputDisabled) return;
    const id = setInterval(() => setElapsed((v) => v + 1), 1000);
    return () => clearInterval(id);
  }, [inputDisabled]);

  useEffect(() => {
    if (!refreshingAfterConflict) return;
    if (loader.state === "loading") {
      conflictRefreshStartedRef.current = true;
      return;
    }
    if (conflictRefreshStartedRef.current) {
      setRefreshingAfterConflict(false);
      conflictRefreshStartedRef.current = false;
    }
  }, [loader.state, refreshingAfterConflict]);

  const handleBackToWorkspace = useCallback(() => {
    navigate({
      name: "workspace",
      params: {
        targetJobId: route.params.targetJobId || ctx.targetJobId,
        jdId: route.params.jdId || ctx.jdId || "",
        planId: route.params.planId || ctx.planId || "",
        resumeVersionId:
          route.params.resumeVersionId || ctx.resumeVersionId || "",
      },
    });
  }, [navigate, route.params, ctx]);

  const handleSwitchMode = useCallback(
    (k: "text" | "voice") => {
      navigate({
        name: "practice",
        params: {
          ...route.params,
          sessionId,
          mode: k,
          modality: k,
        },
      });
    },
    [navigate, route.params, sessionId],
  );

  const handleStrictToggleClick = useCallback(() => {
    if (strictToastTimerRef.current) {
      clearTimeout(strictToastTimerRef.current);
    }
    setStrictToastOpen(true);
    strictToastTimerRef.current = setTimeout(() => {
      setStrictToastOpen(false);
    }, 4000);
  }, []);

  useEffect(
    () => () => {
      if (strictToastTimerRef.current) clearTimeout(strictToastTimerRef.current);
    },
    [],
  );

  const applyAssistantAction = useCallback((action: AssistantAction) => {
    setAiTransparency(action.provenance);
    setActiveAssistantAction(action);
  }, []);

  const handleMutationError = useCallback(
    (err: unknown, retryAction: RetryAction | null) => {
      const classified = classifyPracticeError(err);
      if (classified.sessionLost) {
        setSessionLostByMutation(true);
      }
      if (classified.refreshSession) {
        conflictRefreshStartedRef.current = false;
        setRefreshingAfterConflict(true);
        loader.refresh();
      }
      retryActionRef.current =
        classified.retryable && retryAction ? retryAction : null;
      updatePracticeErrorState(setErrorState, {
        message: t(classified.messageKey),
        retryable: classified.retryable && Boolean(retryAction),
        fallbackBackToWorkspace:
          completion.state.kind === "error"
            ? completion.state.fallbackBackToWorkspace
            : false,
      });
    },
    [completion.state, loader, t],
  );

  const runPracticeAction = useCallback(
    async (action: RetryAction, retryAction: RetryAction | null = action) => {
      setErrorState(null);
      try {
        await action();
        retryActionRef.current = null;
      } catch (err) {
        handleMutationError(err, retryAction);
      }
    },
    [handleMutationError],
  );

  const handleRetry = useCallback(() => {
    const retryAction = retryActionRef.current;
    if (!retryAction) {
      setErrorState(null);
      return;
    }
    void runPracticeAction(retryAction, retryAction);
  }, [runPracticeAction]);

  const buildSessionMapItems = useCallback((): SessionMapItem[] => {
    const data = loader.data;
    const turn = data?.currentTurn ?? null;
    const total = Math.max(data?.turnCount ?? 0, turn ? turn.turnIndex : 0, 1);
    const items: SessionMapItem[] = [];
    for (let i = 1; i <= total; i++) {
      let status: SessionMapItem["status"] = "pending";
      if (turn) {
        if (i < turn.turnIndex) {
          status = "done";
        } else if (i === turn.turnIndex) {
          status = "active";
        }
      }
      const annotation = turnAnnotations.get(i);
      if (annotation === "skipped") {
        status = "skipped";
      } else if (annotation === "follow_up_requested") {
        status = "follow_up_requested";
      } else if (annotation === "done") {
        status = "done";
      }
      items.push({
        id: turn && i === turn.turnIndex ? turn.id : `q-skeleton-${i}`,
        topic: turn && i === turn.turnIndex
          ? (turn.questionIntent ?? t("practice.sessionMap.itemTopicSkeleton"))
          : t("practice.sessionMap.itemTopicSkeleton"),
        duration: "—",
        status,
      });
    }
    return items;
  }, [loader.data, turnAnnotations, t]);

  const sessionMapItems = useMemo(buildSessionMapItems, [buildSessionMapItems]);
  const activeIndex = (loader.data?.currentTurn?.turnIndex ?? 1) - 1;

  const fmtElapsed = (sec: number) =>
    `${String(Math.floor(sec / 60)).padStart(2, "0")}:${String(sec % 60).padStart(2, "0")}`;

  // ── handlers ──────────────────────────────────────────────────────────
  const onSend = useCallback(async () => {
    if (inputDisabled || !input.trim()) return;
    const turnId = loader.data?.currentTurn?.id ?? "";
    if (!turnId) return;
    const answerText = input.trim();
    const sentAt = fmtElapsed(elapsed);
    const action = async () => {
      const result = await events.submitAnswer({ turnId, answerText });
      setTranscript((prev) => [
        ...prev,
        { role: "user", text: answerText, t: sentAt },
      ]);
      setInput("");
      applyAssistantAction(result.assistantAction);
    };
    await runPracticeAction(action, action);
  }, [
    applyAssistantAction,
    elapsed,
    events,
    input,
    inputDisabled,
    loader.data,
    runPracticeAction,
  ]);

  const onHint = useCallback(async () => {
    if (inputDisabled) return;
    if (showHintBanner) {
      setShowHintBanner(false);
      return;
    }
    const turnId = loader.data?.currentTurn?.id ?? "";
    if (!turnId) return;
    const action = async () => {
      const result = await events.requestHint({ turnId });
      applyAssistantAction(result.assistantAction);
    };
    await runPracticeAction(action, action);
  }, [
    applyAssistantAction,
    events,
    inputDisabled,
    loader.data,
    runPracticeAction,
    showHintBanner,
  ]);

  const onSkip = useCallback(async () => {
    if (inputDisabled) return;
    const turn = loader.data?.currentTurn;
    if (!turn) return;
    const turnIndex = turn.turnIndex;
    const action = async () => {
      const result = await events.skipTurn({ turnId: turn.id });
      setTurnAnnotations((prev) => {
        const next = new Map(prev);
        next.set(turnIndex, "skipped");
        return next;
      });
      applyAssistantAction(result.assistantAction);
    };
    await runPracticeAction(action, action);
  }, [applyAssistantAction, events, inputDisabled, loader.data, runPracticeAction]);

  const onTogglePause = useCallback(async () => {
    if (paused) {
      const action = async () => {
        await events.resumeSession();
        setPaused(false);
      };
      await runPracticeAction(action, action);
    } else {
      const action = async () => {
        await events.pauseSession();
        setPaused(true);
      };
      await runPracticeAction(action, action);
    }
  }, [events, paused, runPracticeAction]);

  const handleAskQuestion = useCallback(
    (turnId: string, questionText: string) => {
      if (questionText) {
        setTranscript((prev) => [
          ...prev,
          { role: "ai", text: questionText, t: fmtElapsed(elapsed) },
        ]);
      }
      // turn advance is reflected by getPracticeSession refresh; no-op here.
    },
    [elapsed],
  );

  const handleAskFollowUp = useCallback(
    (_turnId: string, questionText: string) => {
      if (questionText) {
        setTranscript((prev) => [
          ...prev,
          {
            role: "ai",
            text: questionText,
            t: fmtElapsed(elapsed),
            followUp: true,
          },
        ]);
      }
      const turnIndex = loader.data?.currentTurn?.turnIndex;
      if (turnIndex) {
        setTurnAnnotations((prev) => {
          const next = new Map(prev);
          next.set(turnIndex, "follow_up_requested");
          return next;
        });
      }
    },
    [elapsed, loader.data?.currentTurn?.turnIndex],
  );

  const handleShowHint = useCallback(
    (hint: string, _turnId: string) => {
      setHintBannerText(hint);
      setShowHintBanner(true);
      dispatch({ type: "INCREMENT_HINT_COUNT" });
    },
    [dispatch],
  );

  const handleSessionWait = useCallback(() => {
    setErrorState(null);
  }, []);

  const handleSessionCompleted = useCallback(() => {
    setErrorState(null);
  }, []);

  const handoffNavigatedRef = useRef(false);
  const onFinish = useCallback(async () => {
    if (handoffNavigatedRef.current) return;
    const action = async () => {
      const report = await completion.complete();
      if (handoffNavigatedRef.current) return;
      handoffNavigatedRef.current = true;
      const handoff = buildPracticeHandoffParams({
        ctx: { ...ctx, sessionId },
        reportId: report.reportId,
        mode,
        modality,
        practiceMode,
        practiceGoal,
        hintCount: Number(ctx.hintCount) || 0,
      });
      navigate({
        name: "generating",
        params: handoff as unknown as Record<string, string>,
      });
    };
    await runPracticeAction(action, action);
  }, [
    completion,
    ctx,
    mode,
    modality,
    navigate,
    practiceGoal,
    practiceMode,
    runPracticeAction,
    sessionId,
  ]);

  useEffect(() => {
    if (completion.state.kind !== "error") return;
    const classified = classifyPracticeError(completion.state.message);
    updatePracticeErrorState(setErrorState, {
      message: t(classified.messageKey),
      retryable:
        completion.state.retryable && Boolean(retryActionRef.current),
      fallbackBackToWorkspace: completion.state.fallbackBackToWorkspace,
    });
  }, [completion.state, t]);

  // Initial transcript seed: first AI question from loader.
  useEffect(() => {
    if (
      loader.state === "data" &&
      loader.data?.currentTurn &&
      transcript.length === 0
    ) {
      setTranscript([
        {
          role: "ai",
          text: loader.data.currentTurn.questionText,
          t: "00:00",
        },
      ]);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [loader.state, loader.data?.currentTurn?.id]);

  if (!sessionId || loader.state === "sessionLost" || sessionLostByMutation) {
    return <PracticeSessionLostState onBack={handleBackToWorkspace} />;
  }

  const turnTotal = loader.data?.turnCount ?? sessionMapItems.length;
  const currentTurn = loader.data?.currentTurn;
  const hintCount = Number(ctx.hintCount) || 0;
  const finishCta = (
    <FinishCta
      label={t("practice.rightpanel.finishCta")}
      hintCount={hintCount}
      hintUsageNote={t("practice.rightpanel.hintUsageNote")}
      disabled={
        sessionFlags.completionCtaDisabled ||
        completion.state.kind === "loading"
      }
      onFinish={onFinish}
    />
  );

  return (
    <div
      data-testid="practice-screen"
      data-session-id={sessionId}
      data-plan-id={route.params.planId || ctx.planId || ""}
      data-target-job-id={route.params.targetJobId || ctx.targetJobId || ""}
      data-jd-id={route.params.jdId || ctx.jdId || ""}
      data-resume-version-id={
        route.params.resumeVersionId || ctx.resumeVersionId || ""
      }
      data-round-id={route.params.roundId || ctx.roundId || ""}
      data-mode={mode}
      data-modality={modality}
      data-practice-mode={practiceMode}
      data-practice-goal={practiceGoal}
      className="ei-fadein"
      style={{
        height: "100vh",
        display: "flex",
        flexDirection: "column",
        background: "var(--ei-color-bg-canvas)",
        overflow: "hidden",
      }}
    >
      <TopBar
        company={t("practice.toolbar.companySkeleton")}
        title={t("practice.toolbar.titleSkeleton")}
        questionIndex={currentTurn?.turnIndex ?? 1}
        questionTotal={Math.max(turnTotal, 5)}
        elapsed={fmtElapsed(elapsed)}
        budget="25:00"
        paused={paused}
        onTogglePause={onTogglePause}
        activeMode={activeMode}
        onSwitchMode={handleSwitchMode}
        strict={isStrict}
        onToggleStrict={handleStrictToggleClick}
        persona={persona}
        onPersonaChange={setPersona}
      />

      {strictToastOpen && (
        <div
          data-testid="practice-strict-locked-toast"
          role="status"
          style={{
            position: "fixed",
            top: 72,
            right: 24,
            zIndex: 50,
            padding: "10px 14px",
            background: "var(--ei-color-bg-card)",
            border: "1px solid var(--ei-color-rule-strong)",
            borderRadius: 4,
            fontSize: 13,
            color: "var(--ei-color-fg-secondary)",
            boxShadow: "0 6px 24px rgba(0,0,0,0.08)",
          }}
        >
          {t("practice.toolbar.strictLockedToast")}
        </div>
      )}

      <div
        data-testid="practice-main"
        style={{
          flex: 1,
          display: "grid",
          gridTemplateColumns: isNarrow
            ? "minmax(0, 1fr)"
            : "260px minmax(0, 1fr) 280px",
          gridAutoRows: isNarrow ? "max-content" : undefined,
          minHeight: 0,
          overflowY: isNarrow ? "auto" : "hidden",
        }}
      >
        <div
          data-testid="practice-sessionmap"
          style={{
            borderRight: "1px solid var(--ei-color-rule-strong)",
            padding: "20px 18px",
            overflowY: "auto",
            background: "var(--ei-color-bg-soft)",
          }}
        >
          <SessionMap
            label={t("practice.sessionMap.label")}
            items={sessionMapItems}
            activeIndex={Math.max(activeIndex, 0)}
          />
          {assistance.showLiveNotes && (
            <LiveNotes
              label={t("practice.sessionMap.liveNotes")}
              okText={t("practice.sessionMap.liveNotesOk")}
              warnText={t("practice.sessionMap.liveNotesWarn")}
              note={t("practice.sessionMap.liveNotesNote")}
            />
          )}
        </div>

        <div
          data-testid="practice-center"
          style={{ display: "flex", flexDirection: "column", minHeight: 0 }}
        >
          {activeMode === "voice" ? (
            <PracticeVoiceSurface
              lang={lang}
              questionBadge={t("practice.question.tagPrefix").replace(
                "{n}",
                String(currentTurn?.turnIndex ?? 1),
              )}
              topic={
                currentTurn?.questionIntent ??
                t("practice.sessionMap.itemTopicSkeleton")
              }
              prompt={
                currentTurn?.questionText ??
                t("practice.question.skeletonPrompt")
              }
              recording={!paused}
              messages={transcript}
            />
          ) : (
            <>
              <QuestionCard
                badgeText={t("practice.question.tagPrefix").replace(
                  "{n}",
                  String(currentTurn?.turnIndex ?? 1),
                )}
                topic={
                  currentTurn?.questionIntent ??
                  t("practice.sessionMap.itemTopicSkeleton")
                }
                tags={[]}
                prompt={
                  currentTurn?.questionText ??
                  t("practice.question.skeletonPrompt")
                }
              />
              <Transcript
                messages={transcript}
                helperText={t("practice.transcript.helper")}
                aiLabel={t("practice.transcript.aiLabel")}
                userLabel={t("practice.transcript.userLabel")}
                followUpLabel={t("practice.transcript.followUp")}
              />
              <ErrorState
                message={errorState?.message ?? null}
                retryLabel={t("practice.errors.retry")}
                onRetry={errorState?.retryable ? handleRetry : undefined}
              />
              {errorState?.fallbackBackToWorkspace ? (
                <button
                  data-testid="practice-error-back-to-workspace"
                  type="button"
                  onClick={handleBackToWorkspace}
                  style={{
                    alignSelf: "flex-start",
                    margin: "0 40px 10px",
                    background: "var(--ei-color-bg-card)",
                    border: "1px solid var(--ei-color-rule-strong)",
                    color: "var(--ei-color-fg-secondary)",
                    padding: "7px 12px",
                    borderRadius: 2,
                    cursor: "pointer",
                    fontSize: 12,
                    fontFamily: "var(--ei-font-sans)",
                  }}
                >
                  {t("practice.errors.backToWorkspace")}
                </button>
              ) : null}
              <InputBar
                value={input}
                onChange={setInput}
                placeholder={t("practice.input.placeholder")}
                hintLabel={t("practice.input.hint")}
                skipLabel={t("practice.input.skip")}
                sendLabel={t("practice.input.send")}
                dictateLabel={t("practice.input.dictateOn")}
                showHintButton={assistance.showHintButton}
                disabled={inputDisabled}
                onHint={onHint}
                onSkip={onSkip}
                onSend={onSend}
                onDictate={() => undefined}
                hintBanner={
                  <HintBanner
                    show={assistance.showHintButton && showHintBanner}
                    prefix={t("practice.hint.prefix")}
                    text={hintBannerText}
                  />
                }
              />
            </>
          )}
        </div>

        {activeMode === "voice" ? (
          <PracticeVoiceRightPanel
            lang={lang}
            strict={isStrict}
            aiTransparencyLabel={t("practice.rightpanel.aiTransparency")}
            aiTransparencyMeta={{
              promptVersion: aiTransparency?.promptVersion ?? "pending",
              rubricVersion: aiTransparency?.rubricVersion ?? "pending",
              modelId: aiTransparency?.modelId ?? "model-profile:pending",
              language: aiTransparency?.language ?? lang,
              personaLabel: t(PERSONA_LABEL_KEY[persona]),
            }}
            finishCta={finishCta}
          />
        ) : (
          <RightPanel
            jdLinkLabel={t("practice.rightpanel.jdLink")}
            jdProbesLabel={t("practice.rightpanel.jdProbes")}
            jdProbesText={t("practice.rightpanel.jdProbesSkeleton")}
            experienceLabel={t("practice.rightpanel.experienceLabel")}
            aiTransparencyLabel={t("practice.rightpanel.aiTransparency")}
            aiTransparencyMeta={{
              promptVersion: aiTransparency?.promptVersion ?? "pending",
              rubricVersion: aiTransparency?.rubricVersion ?? "pending",
              modelId: aiTransparency?.modelId ?? "model-profile:pending",
              language: aiTransparency?.language ?? lang,
              personaLabel: t(PERSONA_LABEL_KEY[persona]),
            }}
            strict={isStrict}
            strictBannerText={t("practice.rightpanel.strictBanner")}
            experiences={[]}
            finishCta={finishCta}
          />
        )}
      </div>

      <AssistantActionRenderer
        action={activeAssistantAction}
        onAskQuestion={handleAskQuestion}
        onAskFollowUp={handleAskFollowUp}
        onShowHint={handleShowHint}
        onSessionWait={handleSessionWait}
        onSessionCompleted={handleSessionCompleted}
      />
    </div>
  );
};

function classifyPracticeError(err: unknown): ClassifiedPracticeError {
  const message = err instanceof Error ? err.message : String(err);
  if (/^HTTP 404\b/.test(message)) {
    return {
      messageKey: "practice.errors.sessionConflict",
      retryable: false,
      refreshSession: false,
      sessionLost: true,
    };
  }
  if (message.includes("AI_PROVIDER_TIMEOUT")) {
    return {
      messageKey: "practice.errors.aiTimeout",
      retryable: true,
      refreshSession: false,
      sessionLost: false,
    };
  }
  if (message.includes("hint_disabled_in_mode")) {
    return {
      messageKey: "practice.errors.strictHintConflict",
      retryable: false,
      refreshSession: false,
      sessionLost: false,
    };
  }
  if (
    message.includes("client_event_id_mismatch") ||
    (message.includes("PRACTICE_SESSION_CONFLICT") &&
      /^HTTP 409\b/.test(message))
  ) {
    return {
      messageKey: "practice.errors.sessionConflict",
      retryable: false,
      refreshSession: true,
      sessionLost: false,
    };
  }
  if (/^HTTP 5\d\d\b/.test(message) || !/^HTTP \d{3}\b/.test(message)) {
    return {
      messageKey: "practice.errors.network",
      retryable: true,
      refreshSession: false,
      sessionLost: false,
    };
  }
  return {
    messageKey: "practice.errors.unknown",
    retryable: false,
    refreshSession: false,
    sessionLost: false,
  };
}

function updatePracticeErrorState(
  setErrorState: Dispatch<SetStateAction<PracticeErrorState | null>>,
  next: PracticeErrorState,
): void {
  setErrorState((prev) =>
    prev &&
      prev.message === next.message &&
      prev.retryable === next.retryable &&
      prev.fallbackBackToWorkspace === next.fallbackBackToWorkspace
      ? prev
      : next,
  );
}

function useNarrowPracticeLayout(): boolean {
  const [isNarrow, setIsNarrow] = useState(() => getNarrowPracticeLayout());

  useEffect(() => {
    const onResize = () => setIsNarrow(getNarrowPracticeLayout());
    window.addEventListener("resize", onResize);
    return () => window.removeEventListener("resize", onResize);
  }, []);

  return isNarrow;
}

function getNarrowPracticeLayout(): boolean {
  return typeof window !== "undefined" && window.innerWidth <= 720;
}
