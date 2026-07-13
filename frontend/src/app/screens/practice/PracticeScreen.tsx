import { useCallback, useEffect, useMemo, useState, type FC } from "react";

import { useI18n } from "../../i18n/messages";
import { useInterviewContext } from "../../interview-context/InterviewContext";
import { useNavigation } from "../../navigation/NavigationProvider";
import { useAppRuntimeOptional } from "../../runtime/AppRuntimeProvider";
import type { Route } from "../../routes";
import { ErrorState } from "./components/ErrorState";
import { FinishCta } from "./components/FinishCta";
import { InputBar } from "./components/InputBar";
import { PracticeSessionLostState } from "./components/PracticeSessionLostState";
import { TopBar } from "./components/TopBar";
import { Transcript, type TranscriptMessage } from "./components/Transcript";
import { useCompletePracticeSession } from "./hooks/useCompletePracticeSession";
import { usePracticeMessages } from "./hooks/usePracticeMessages";
import { usePracticeSessionLoader } from "./hooks/usePracticeSessionLoader";
import { usePracticeTargetDisplay } from "./usePracticeTargetDisplay";

interface PracticeScreenProps { route: Route; }
type PracticeErrorSource = "message" | "completion";

export const PracticeScreen: FC<PracticeScreenProps> = ({ route }) => {
  const { t } = useI18n();
  const { navigate } = useNavigation();
  const { ctx } = useInterviewContext();
  const runtime = useAppRuntimeOptional();
  const sessionId = route.params.sessionId || ctx.sessionId || "";
  const loader = usePracticeSessionLoader(sessionId);
  const messages = usePracticeMessages(sessionId);
  const completion = useCompletePracticeSession(sessionId);
  const targetDisplay = usePracticeTargetDisplay({
    session: loader.data ? { targetJobId: loader.data.targetJobId } : null,
    routeTargetJobId: route.params.targetJobId,
    contextTargetJobId: ctx.targetJobId,
  });
  const [input, setInput] = useState("");
  const [paused, setPaused] = useState(false);
  const [elapsed, setElapsed] = useState(0);
  const [sending, setSending] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [errorSource, setErrorSource] = useState<PracticeErrorSource | null>(null);
  const [budgetMinutes, setBudgetMinutes] = useState<number | null>(null);
  const planId = loader.data?.planId || route.params.planId || ctx.planId || "";
  const committedMessages = loader.data?.messages ?? [];
  const hasCommittedCandidateMessage = committedMessages.some((message) => message.role === "user");
  const hasPendingAssistantReply = committedMessages.at(-1)?.role === "user";

  useEffect(() => {
    const client = runtime?.client;
    if (!client || !planId) {
      setBudgetMinutes(null);
      return;
    }
    let active = true;
    setBudgetMinutes(null);
    client.getPracticePlan(planId).then((plan) => {
      if (!active) return;
      setBudgetMinutes(
        Number.isInteger(plan.timeBudgetMinutes) && plan.timeBudgetMinutes > 0
          ? plan.timeBudgetMinutes
          : null,
      );
    }).catch(() => {
      if (active) setBudgetMinutes(null);
    });
    return () => {
      active = false;
    };
  }, [planId, runtime?.client]);

  useEffect(() => {
    if (paused) return;
    const timer = window.setInterval(() => setElapsed((value) => value + 1), 1000);
    return () => window.clearInterval(timer);
  }, [paused]);

  const transcript = useMemo<TranscriptMessage[]>(() => (loader.data?.messages ?? []).map((message) => ({
    id: message.id,
    role: message.role === "user" ? "user" : "assistant",
    text: message.content,
    t: formatMessageTime(message.createdAt),
  })), [loader.data?.messages]);

  const backToWorkspace = useCallback(() => navigate({ name: "workspace", params: { targetJobId: targetDisplay.targetJobId || "" } }), [navigate, targetDisplay.targetJobId]);

  const send = useCallback(async () => {
    const text = input.trim();
    if (!text || sending || paused) return;
    setSending(true);
    setError(null);
    setErrorSource(null);
    try {
      const result = await messages.sendMessage(text);
      loader.adopt(result.session);
      setInput("");
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : String(cause));
      setErrorSource("message");
    } finally {
      setSending(false);
    }
  }, [input, loader, messages, paused, sending]);

  const finish = useCallback(async () => {
    if (!hasCommittedCandidateMessage || hasPendingAssistantReply) return;
    setError(null);
    setErrorSource(null);
    try {
      const report = await completion.complete();
      navigate({ name: "generating", params: { reportId: report.reportId } });
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : String(cause));
      setErrorSource("completion");
    }
  }, [completion, hasCommittedCandidateMessage, hasPendingAssistantReply, navigate]);

  if (!sessionId || loader.state === "sessionLost") return <PracticeSessionLostState onBack={backToWorkspace} />;

  const inputDisabled = paused || sending || loader.state === "loading" || !messages.ready || loader.data?.status === "completed" || loader.data?.status === "completing";
  const finishDisabled = paused || sending || loader.state === "loading" || !loader.data || !messages.ready || completion.state.kind === "loading" || !hasCommittedCandidateMessage || hasPendingAssistantReply || (loader.data.status !== "running" && loader.data.status !== "waiting_user_input");
  const interviewerLabel = route.params.roundName || ctx.roundName || t("practice.toolbar.role.manager");

  return (
    <div data-testid="practice-screen" data-session-id={sessionId} data-plan-id={route.params.planId || ctx.planId || ""} data-target-job-id={targetDisplay.targetJobId || ""} className="ei-fadein" style={{ height: "100vh", display: "flex", flexDirection: "column", background: "var(--ei-color-bg-canvas)", overflow: "hidden" }}>
      <TopBar
        company={targetDisplay.companyName ?? t("practice.toolbar.companySkeleton")}
        title={targetDisplay.title ?? t("practice.toolbar.titleSkeleton")}
        elapsed={formatElapsed(elapsed)}
        budget={formatBudget(budgetMinutes)}
        paused={paused}
        pauseLabel={t("practice.toolbar.pause")}
        resumeLabel={t("practice.toolbar.resume")}
        onTogglePause={() => setPaused((value) => !value)}
        interviewerLabel={interviewerLabel}
        phoneDisabledLabel={t("practice.toolbar.phoneDisabled")}
        finishCta={<FinishCta label={t("practice.finishCta")} disabled={finishDisabled} disabledReason={!hasCommittedCandidateMessage && loader.data ? t("practice.finishDisabled.zeroAnswer") : undefined} onFinish={finish} />}
      />
      <main data-testid="practice-conversation" style={{ flex: 1, minHeight: 0, display: "flex", flexDirection: "column", width: "100%" }}>
        <Transcript messages={transcript} helperText={t("practice.transcript.helper")} aiLabel={t("practice.transcript.aiLabel")} userLabel={t("practice.transcript.userLabel")} />
        <ErrorState
          message={error || loader.error?.message || null}
          retryLabel={t("practice.errors.retry")}
          onRetry={errorSource === "message" ? send : errorSource === "completion" ? finish : loader.error ? loader.refresh : undefined}
        />
        <InputBar value={input} onChange={setInput} placeholder={t("practice.input.placeholder")} sendLabel={t("practice.input.send")} disabled={inputDisabled} onSend={send} />
      </main>
    </div>
  );
};

function formatElapsed(seconds: number): string {
  return `${String(Math.floor(seconds / 60)).padStart(2, "0")}:${String(seconds % 60).padStart(2, "0")}`;
}

function formatBudget(minutes: number | null): string {
  if (minutes === null) return "--:--";
  return `${String(minutes).padStart(2, "0")}:00`;
}

function formatMessageTime(raw: string): string {
  const date = new Date(raw);
  if (Number.isNaN(date.getTime())) return "00:00";
  return `${String(date.getHours()).padStart(2, "0")}:${String(date.getMinutes()).padStart(2, "0")}`;
}
