import { useCallback, useEffect, useMemo, useRef, useState, type FC } from "react";

import type { AssistantAction } from "../../../api/generated/types";
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
import { VoiceSurfaceComingSoon } from "./components/VoiceSurfaceComingSoon";
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
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  type TurnAnnotation = "skipped" | "follow_up_requested" | "done";
  const [turnAnnotations, setTurnAnnotations] = useState<
    Map<number, TurnAnnotation>
  >(() => new Map());

  // Local elapsed timer (UI display only; backend owns server elapsed).
  const [elapsed, setElapsed] = useState(0);
  useEffect(() => {
    if (paused || sessionFlags.inputDisabled) return;
    const id = setInterval(() => setElapsed((v) => v + 1), 1000);
    return () => clearInterval(id);
  }, [paused, sessionFlags.inputDisabled]);

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
    if (sessionFlags.inputDisabled || paused || !input.trim()) return;
    const turnId = loader.data?.currentTurn?.id ?? "";
    if (!turnId) return;
    const answerText = input.trim();
    setTranscript((prev) => [
      ...prev,
      { role: "user", text: answerText, t: fmtElapsed(elapsed) },
    ]);
    setInput("");
    try {
      const result = await events.submitAnswer({ turnId, answerText });
      setActiveAssistantAction(result.assistantAction);
    } catch (err) {
      setErrorMessage(
        err instanceof Error ? err.message : t("practice.errors.unknown"),
      );
    }
  }, [
    elapsed,
    events,
    input,
    loader.data,
    paused,
    sessionFlags.inputDisabled,
    t,
  ]);

  const onHint = useCallback(async () => {
    if (sessionFlags.inputDisabled || paused) return;
    if (showHintBanner) {
      setShowHintBanner(false);
      return;
    }
    const turnId = loader.data?.currentTurn?.id ?? "";
    if (!turnId) return;
    try {
      const result = await events.requestHint({ turnId });
      setActiveAssistantAction(result.assistantAction);
    } catch (err) {
      const msg = err instanceof Error ? err.message : "";
      if (msg.startsWith("HTTP 409 ")) {
        setErrorMessage(t("practice.errors.strictHintConflict"));
      } else {
        setErrorMessage(t("practice.errors.unknown"));
      }
    }
  }, [
    events,
    loader.data,
    paused,
    sessionFlags.inputDisabled,
    showHintBanner,
    t,
  ]);

  const onSkip = useCallback(async () => {
    if (sessionFlags.inputDisabled || paused) return;
    const turn = loader.data?.currentTurn;
    if (!turn) return;
    const turnIndex = turn.turnIndex;
    setTurnAnnotations((prev) => {
      const next = new Map(prev);
      next.set(turnIndex, "skipped");
      return next;
    });
    try {
      const result = await events.skipTurn({ turnId: turn.id });
      setActiveAssistantAction(result.assistantAction);
    } catch (err) {
      setErrorMessage(
        err instanceof Error ? err.message : t("practice.errors.unknown"),
      );
    }
  }, [events, loader.data, paused, sessionFlags.inputDisabled, t]);

  const onTogglePause = useCallback(async () => {
    if (paused) {
      try {
        await events.resumeSession();
        setPaused(false);
      } catch (err) {
        setErrorMessage(
          err instanceof Error ? err.message : t("practice.errors.unknown"),
        );
      }
    } else {
      try {
        await events.pauseSession();
        setPaused(true);
      } catch (err) {
        setErrorMessage(
          err instanceof Error ? err.message : t("practice.errors.unknown"),
        );
      }
    }
  }, [events, paused, t]);

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
    setErrorMessage(null);
  }, []);

  const handleSessionCompleted = useCallback(() => {
    setErrorMessage(null);
  }, []);

  const handoffNavigatedRef = useRef(false);
  const onFinish = useCallback(async () => {
    if (handoffNavigatedRef.current) return;
    try {
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
    } catch (err) {
      setErrorMessage(
        err instanceof Error ? err.message : t("practice.errors.unknown"),
      );
    }
  }, [
    completion,
    ctx,
    mode,
    modality,
    navigate,
    practiceGoal,
    practiceMode,
    sessionId,
    t,
  ]);

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

  if (!sessionId || loader.state === "sessionLost") {
    return <PracticeSessionLostState onBack={handleBackToWorkspace} />;
  }

  const turnTotal = loader.data?.turnCount ?? sessionMapItems.length;
  const currentTurn = loader.data?.currentTurn;
  const hintCount = Number(ctx.hintCount) || 0;

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
        background: "var(--ei-color-bg)",
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
            background: "var(--ei-color-bgCard)",
            border: "1px solid var(--ei-color-rule)",
            borderRadius: 4,
            fontSize: 13,
            color: "var(--ei-color-ink2)",
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
          gridTemplateColumns: "260px 1fr 280px",
          minHeight: 0,
        }}
      >
        <div
          data-testid="practice-sessionmap"
          style={{
            borderRight: "1px solid var(--ei-color-rule)",
            padding: "20px 18px",
            overflowY: "auto",
            background: "var(--ei-color-bgSoft)",
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
            <VoiceSurfaceComingSoon
              title={t("practice.voiceComingSoon.title")}
              desc={t("practice.voiceComingSoon.desc")}
              backLabel={t("practice.voiceComingSoon.backToText")}
              onBackToText={() => handleSwitchMode("text")}
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
                message={errorMessage}
                retryLabel={t("practice.errors.retry")}
                onRetry={errorMessage ? () => setErrorMessage(null) : undefined}
              />
              <InputBar
                value={input}
                onChange={setInput}
                placeholder={t("practice.input.placeholder")}
                hintLabel={t("practice.input.hint")}
                skipLabel={t("practice.input.skip")}
                sendLabel={t("practice.input.send")}
                dictateLabel={t("practice.input.dictateOn")}
                showHintButton={assistance.showHintButton}
                disabled={sessionFlags.inputDisabled || paused}
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

        <RightPanel
          jdLinkLabel={t("practice.rightpanel.jdLink")}
          jdProbesLabel={t("practice.rightpanel.jdProbes")}
          jdProbesText={t("practice.rightpanel.jdProbesSkeleton")}
          experienceLabel={t("practice.rightpanel.experienceLabel")}
          aiTransparencyLabel={t("practice.rightpanel.aiTransparency")}
          aiTransparencyMeta={{
            promptVersion: "v1.0.4",
            rubricVersion: "v0.9",
            modelId: "haiku-4.5",
            language: lang,
            personaLabel: t(PERSONA_LABEL_KEY[persona]),
          }}
          strict={isStrict}
          strictBannerText={t("practice.rightpanel.strictBanner")}
          experiences={[]}
          finishCta={
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
          }
        />
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
