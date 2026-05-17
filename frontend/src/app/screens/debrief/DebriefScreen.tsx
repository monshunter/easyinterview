import { useCallback, useEffect, useMemo, useState, type FC } from "react";

import { useNavigation } from "../../navigation/NavigationProvider";
import { useRequestAuth } from "../../auth/useRequestAuth";
import { useInterviewContext } from "../../interview-context/InterviewContext";
import type { Route } from "../../routes";
import { DebriefAnalysisStep } from "./components/DebriefAnalysisStep";
import { DebriefContextStrip } from "./components/DebriefContextStrip";
import { DebriefFailureState } from "./components/DebriefFailureState";
import { DebriefHeader } from "./components/DebriefHeader";
import { DebriefMissingContextState } from "./components/DebriefMissingContextState";
import { DebriefModeToggle } from "./components/DebriefModeToggle";
import { DebriefRecordSummaryBar } from "./components/DebriefRecordSummaryBar";
import { DebriefReplayPlan } from "./components/DebriefReplayPlan";
import { DebriefStepper } from "./components/DebriefStepper";
import { DebriefSubmitCTA } from "./components/DebriefSubmitCTA";
import { DebriefTimeoutState } from "./components/DebriefTimeoutState";
import { GuidedDebriefRecord } from "./components/GuidedDebriefRecord";
import { JDPicker } from "./components/JDPicker";
import { MockSessionPicker } from "./components/MockSessionPicker";
import { ResumePicker } from "./components/ResumePicker";
import { VoiceDebriefRecord } from "./components/VoiceDebriefRecord";
import { useDebriefPolling } from "./hooks/useDebriefPolling";
import { useSubmitDebrief } from "./hooks/useSubmitDebrief";
import { useSuggestDebriefQuestions } from "./hooks/useSuggestDebriefQuestions";
import {
  EMPTY_SELECTED_CONTEXT,
  type DebriefEntry,
  type DebriefInputMode,
  type DebriefPickerKind,
  type DebriefSelectedContext,
  type DebriefStep,
} from "./types";
import "./debrief.css";

interface DebriefScreenProps {
  route: Route;
}

/**
 * Source mirror of ui-design/src/screens-p1-depth.jsx::DebriefFullScreen.
 *
 * Composes Phase 2-6 surfaces into one route: 3 picker modals + Step 0
 * record (text + voice shell) + Step 1 analysis + Step 2 replay launcher.
 * Wires Phase 4 suggestions and Phase 5 createDebrief + dual-track polling
 * + failure/missing/timeout states, plus the Phase 5.4 InterviewContext
 * reducer extension.
 */
export const DebriefScreen: FC<DebriefScreenProps> = ({ route }) => {
  const { navigate } = useNavigation();
  const requestAuth = useRequestAuth();
  const { ctx, dispatch } = useInterviewContext();

  const [step, setStep] = useState<DebriefStep>(0);
  const [maxVisited, setMaxVisited] = useState<DebriefStep>(0);
  const [inputMode, setInputMode] = useState<DebriefInputMode>("text");
  const [selectedContext, setSelectedContext] =
    useState<DebriefSelectedContext>(EMPTY_SELECTED_CONTEXT);
  const [pickerKind, setPickerKind] = useState<DebriefPickerKind | null>(null);
  const [entries, setEntries] = useState<DebriefEntry[]>([]);
  const [activeGuide, setActiveGuide] = useState(0);

  const advanceStep = useCallback((next: DebriefStep) => {
    setStep(next);
    setMaxVisited((prev) => (next > prev ? next : prev));
  }, []);

  const handleBack = useCallback(() => {
    navigate({ name: "home" });
  }, [navigate]);

  const handleOpenPicker = useCallback((kind: DebriefPickerKind) => {
    setPickerKind(kind);
  }, []);

  // Phase 2.5 — auto-trigger suggestions once targetJob + resume are both set.
  const suggestionsEnabled =
    Boolean(selectedContext.targetJob) &&
    Boolean(selectedContext.resumeVersion);
  const language = ctx.modality === "en" ? "en-US" : "zh-CN";
  const suggestions = useSuggestDebriefQuestions({
    targetJobId: selectedContext.targetJob?.id,
    sessionId: selectedContext.mockSession?.id,
    resumeVersionId: selectedContext.resumeVersion?.id,
    language,
    enabled: suggestionsEnabled,
  });

  const submit = useSubmitDebrief();
  const polling = useDebriefPolling({
    debriefJobId: submit.result?.job?.id ?? null,
    debriefId: submit.result?.debriefId ?? null,
    enabled: submit.status === "succeeded",
  });

  // When polling completes, advance to step 1; on failed/timeout, hold step 0
  // and let the state cards drive the next action.
  useEffect(() => {
    if (polling.state === "succeeded" && polling.debrief) {
      advanceStep(1);
    }
  }, [advanceStep, polling.debrief, polling.state]);

  const handleSubmit = useCallback(async () => {
    if (!selectedContext.targetJob) return;
    await submit.submit({
      targetJobId: selectedContext.targetJob.id,
      roundType: "technical",
      language,
      entries,
    });
  }, [entries, language, selectedContext.targetJob, submit]);

  const handleStartReplay = useCallback(() => {
    if (!selectedContext.targetJob || !submit.result) return;
    const params: Record<string, string> = {
      practiceGoal: "debrief",
      mode: "text",
      modality: "text",
      targetJobId: selectedContext.targetJob.id,
    };
    if (selectedContext.resumeVersion?.id) {
      params.resumeVersionId = selectedContext.resumeVersion.id;
    }
    if (selectedContext.mockSession?.id) {
      params.sessionId = selectedContext.mockSession.id;
    }
    if (submit.result.debriefId) {
      params.debriefId = submit.result.debriefId;
    }
    requestAuth({
      type: "start_debrief_interview",
      label: "开始复盘面试",
      route: "practice",
      params,
    });
  }, [requestAuth, selectedContext, submit.result]);

  const failureCard = useMemo(() => {
    if (submit.status === "validation_failed" || submit.status === "failed") {
      return (
        <DebriefFailureState
          errorCode={submit.error?.code ?? null}
          onRetry={() => {
            submit.reset();
            void handleSubmit();
          }}
          onBackToEdit={() => {
            submit.reset();
            setStep(0);
          }}
        />
      );
    }
    if (polling.state === "failed") {
      return (
        <DebriefFailureState
          errorCode={polling.errorCode}
          onRetry={() => {
            polling.restart();
          }}
          onBackToEdit={() => {
            polling.restart();
            submit.reset();
            setStep(0);
          }}
        />
      );
    }
    if (polling.state === "timeout") {
      return (
        <DebriefTimeoutState
          onRetry={() => polling.restart()}
          onBackToEdit={() => {
            polling.restart();
            submit.reset();
            setStep(0);
          }}
        />
      );
    }
    return null;
  }, [handleSubmit, polling, submit]);

  const stepPanel = useMemo(() => {
    if (failureCard) return failureCard;
    if (step === 0) {
      if (!selectedContext.targetJob) {
        return (
          <DebriefMissingContextState
            onPickTargetJob={() => setPickerKind("targetJob")}
          />
        );
      }
      return (
        <>
          <DebriefRecordSummaryBar entries={entries} />
          <DebriefModeToggle inputMode={inputMode} onChange={setInputMode} />
          <div data-mode={inputMode} className="ei-debrief-step0-panels">
            <div hidden={inputMode !== "text"}>
              <GuidedDebriefRecord
                suggestions={suggestions.suggestions}
                loading={suggestions.loading}
                errorCode={suggestions.error?.code ?? null}
                entries={entries}
                setEntries={setEntries}
                activeGuide={activeGuide}
                setActiveGuide={setActiveGuide}
                onRegenerate={suggestions.refetch}
              />
            </div>
            <div hidden={inputMode !== "voice"}>
              <VoiceDebriefRecord entries={entries} />
            </div>
          </div>
          <DebriefSubmitCTA
            entriesCount={entries.length}
            targetJobSelected={Boolean(selectedContext.targetJob)}
            submitting={submit.status === "submitting"}
            onSubmit={handleSubmit}
          />
        </>
      );
    }
    if (step === 1) {
      if (!polling.debrief) {
        return (
          <div data-testid="debrief-analysis-pending">
            {polling.state === "running"
              ? "AI 分析中…"
              : "等待复盘提交…"}
          </div>
        );
      }
      return (
        <DebriefAnalysisStep
          debrief={polling.debrief}
          selectedContext={selectedContext}
          onAdvance={() => advanceStep(2)}
        />
      );
    }
    return (
      <DebriefReplayPlan
        debrief={polling.debrief}
        entries={entries}
        onStart={handleStartReplay}
        onBack={() => advanceStep(1)}
      />
    );
  }, [
    activeGuide,
    advanceStep,
    entries,
    failureCard,
    handleStartReplay,
    handleSubmit,
    inputMode,
    polling.debrief,
    polling.state,
    selectedContext,
    step,
    submit.status,
    suggestions,
  ]);

  // Hydrate context from route params on mount and propagate context choices
  // into the InterviewContext (mainly debriefId once Phase 5 completes).
  useEffect(() => {
    dispatch({ type: "HYDRATE_FROM_ROUTE", params: route.params });
  }, [dispatch, route.params]);

  return (
    <section
      className="ei-screen-shell ei-debrief-screen"
      data-testid="route-debrief"
      data-route-name={route.name}
      data-route-params={JSON.stringify(route.params)}
      data-step={String(step)}
      data-input-mode={inputMode}
      data-picker-kind={pickerKind ?? "none"}
      data-polling-state={polling.state}
      data-submit-status={submit.status}
    >
      <DebriefHeader
        selectedContext={selectedContext}
        onBack={handleBack}
      />
      <DebriefContextStrip
        selectedContext={selectedContext}
        onOpenPicker={handleOpenPicker}
      />
      <DebriefStepper
        step={step}
        maxVisited={maxVisited}
        onStep={(next) => {
          if (next <= maxVisited) setStep(next);
        }}
      />
      <div
        className="ei-debrief-step-panel"
        data-testid={`debrief-step-panel-${step}`}
      >
        {stepPanel}
      </div>
      {pickerKind === "targetJob" && (
        <JDPicker
          selectedId={selectedContext.targetJob?.id ?? null}
          onClose={() => setPickerKind(null)}
          onConfirm={(targetJob) => {
            setSelectedContext((prev) => ({ ...prev, targetJob }));
            setPickerKind(null);
          }}
        />
      )}
      {pickerKind === "mockSession" && (
        <MockSessionPicker
          targetJobId={selectedContext.targetJob?.id ?? null}
          selectedId={selectedContext.mockSession?.id ?? null}
          onClose={() => setPickerKind(null)}
          onConfirm={(mockSession) => {
            setSelectedContext((prev) => ({ ...prev, mockSession }));
            setPickerKind(null);
          }}
        />
      )}
      {pickerKind === "resume" && (
        <ResumePicker
          selectedAssetId={selectedContext.resumeAsset?.id ?? null}
          selectedVersionId={selectedContext.resumeVersion?.id ?? null}
          onClose={() => setPickerKind(null)}
          onConfirm={(selection) => {
            setSelectedContext((prev) => ({
              ...prev,
              resumeAsset: selection.asset,
              resumeVersion: selection.version,
            }));
            setPickerKind(null);
          }}
        />
      )}
    </section>
  );
};
