import { useCallback, useEffect, useMemo, useRef, useState, type FC } from "react";

import { ApiClientError } from "../../../api/generated/client";
import type { PracticeUserMessage } from "../../../api/generated/types";
import { newId } from "../../../lib/ids";
import { resolveContentLimits, utf8ByteLength } from "../../../lib/contentLimits";
import { useI18n, type MessageKey } from "../../i18n/messages";
import { useInterviewContext } from "../../interview-context/InterviewContext";
import { useNavigation } from "../../navigation/NavigationProvider";
import { useAppRuntimeOptional } from "../../runtime/AppRuntimeProvider";
import type { Route } from "../../routes";
import { ErrorState } from "./components/ErrorState";
import { FinishCta } from "./components/FinishCta";
import { InputBar } from "./components/InputBar";
import { PracticeSessionLostState } from "./components/PracticeSessionLostState";
import { TerminalRecovery } from "./components/TerminalRecovery";
import { TopBar } from "./components/TopBar";
import { Transcript, type TranscriptMessage } from "./components/Transcript";
import { useCompletePracticeSession } from "./hooks/useCompletePracticeSession";
import { usePracticeMessages } from "./hooks/usePracticeMessages";
import { usePracticeSessionLoader } from "./hooks/usePracticeSessionLoader";
import { usePracticeTargetDisplay } from "./usePracticeTargetDisplay";

interface PracticeScreenProps { route: Route; }
type PracticeErrorSource = "completion" | "message";
type TransientReplyStatus = "pending" | "retrying" | "retryable_failed" | "terminal_failed";
const PENDING_POLL_DELAYS_MS = [750, 1_500, 3_000, 5_000, 8_000] as const;
const MESSAGE_POST_TIMEOUT_MS = 95_000;
const MESSAGE_POST_TIMEOUT = Symbol("practice-message-post-timeout");

interface TransientUserMessage {
  clientMessageId: string;
  text: string;
  createdAt: string;
  status: TransientReplyStatus;
}

interface ActiveMessageRequest {
  seq: number;
  postController: AbortController;
  readController?: AbortController;
}

export const PracticeScreen: FC<PracticeScreenProps> = ({ route }) => {
  const { ctx } = useInterviewContext();
  const sessionId = route.params.sessionId || ctx.sessionId || "";
  return <PracticeSessionScreen key={sessionId} route={route} sessionId={sessionId} />;
};

const PracticeSessionScreen: FC<PracticeScreenProps & { sessionId: string }> = ({ route, sessionId }) => {
  const { t } = useI18n();
  const { navigate } = useNavigation();
  const { ctx } = useInterviewContext();
  const runtime = useAppRuntimeOptional();
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
  const [transientMessage, setTransientMessage] = useState<TransientUserMessage | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [errorSource, setErrorSource] = useState<PracticeErrorSource | null>(null);
  const [budgetMinutes, setBudgetMinutes] = useState<number | null>(null);
  const messageRequestSeqRef = useRef(0);
  const activeMessageRequestRef = useRef<ActiveMessageRequest | null>(null);
  const transientMessageRef = useRef<TransientUserMessage | null>(null);
  const pendingPollRef = useRef({ clientMessageId: "", attempt: 0 });
  const sessionEpochRef = useRef({ sessionId, active: true });
  const planId = loader.data?.planId || route.params.planId || ctx.planId || "";
  const committedMessages = loader.data?.messages ?? [];
  const contentLimits = resolveContentLimits(
    runtime?.runtime.status === "ready" ? runtime.runtime.config : undefined,
  );
  const hasCommittedCandidateMessage = committedMessages.some((message) => message.role === "user");
  const serverUnresolvedMessage = [...committedMessages].reverse().find(
    (message): message is PracticeUserMessage => message.role === "user" && message.replyStatus !== "complete",
  );
  const activeReplyStatus = transientMessage?.status ?? serverUnresolvedMessage?.replyStatus;
  const isThinking = activeReplyStatus === "pending" || activeReplyStatus === "retrying";
  const hasUnresolvedReply = activeReplyStatus !== undefined && activeReplyStatus !== "complete";
  const terminalRecovery = serverUnresolvedMessage?.replyStatus === "terminal_failed";
  const hasRowLocalRetry = activeReplyStatus === "retryable_failed";
  const loaderFailureNeedsGlobalRecovery = Boolean(loader.error && !terminalRecovery && !hasRowLocalRetry);

  useEffect(() => {
    sessionEpochRef.current.active = true;
    return () => {
      sessionEpochRef.current.active = false;
      messageRequestSeqRef.current += 1;
      activeMessageRequestRef.current?.postController.abort();
      activeMessageRequestRef.current?.readController?.abort();
      activeMessageRequestRef.current = null;
    };
  }, [sessionId]);

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

  const transcript = useMemo<TranscriptMessage[]>(() => {
    const projected = (loader.data?.messages ?? []).map((message): TranscriptMessage => {
      if (message.role === "assistant") {
        return {
          id: message.id,
          role: "assistant",
          text: message.content,
          t: formatMessageTime(message.createdAt),
        };
      }
      const transientStatus = transientMessage?.clientMessageId === message.clientMessageId
        ? transientMessage.status
        : undefined;
      return {
        id: message.id,
        role: "user",
        text: message.content,
        t: formatMessageTime(message.createdAt),
        clientMessageId: message.clientMessageId,
        status: transientStatus ?? message.replyStatus,
      };
    });
    if (
      transientMessage
      && !projected.some((message) => message.clientMessageId === transientMessage.clientMessageId)
    ) {
      projected.push({
        id: `optimistic-${transientMessage.clientMessageId}`,
        role: "user",
        text: transientMessage.text,
        t: formatMessageTime(transientMessage.createdAt),
        clientMessageId: transientMessage.clientMessageId,
        status: transientMessage.status,
      });
    }
    return projected;
  }, [loader.data?.messages, transientMessage]);

  useEffect(() => {
    if (!transientMessage) return;
    const serverMessage = committedMessages.find(
      (message): message is PracticeUserMessage => message.role === "user" && message.clientMessageId === transientMessage.clientMessageId,
    );
    if (
      !serverMessage
      || (transientMessage.status === "retrying" && serverMessage.replyStatus === "retryable_failed")
    ) return;
    transientMessageRef.current = null;
    setTransientMessage(null);
    if (errorSource === "message") {
      setError(null);
      setErrorSource(null);
    }
  }, [committedMessages, errorSource, transientMessage]);

  useEffect(() => {
    const clientMessageId = serverUnresolvedMessage?.replyStatus === "pending"
      ? serverUnresolvedMessage.clientMessageId
      : "";
    if (!clientMessageId) {
      pendingPollRef.current = { clientMessageId: "", attempt: 0 };
      return;
    }
    if (pendingPollRef.current.clientMessageId !== clientMessageId) {
      pendingPollRef.current = { clientMessageId, attempt: 0 };
    }
    if (loader.state !== "data") return;
    const delay = PENDING_POLL_DELAYS_MS[
      Math.min(pendingPollRef.current.attempt, PENDING_POLL_DELAYS_MS.length - 1)
    ];
    const timer = window.setTimeout(() => {
      if (pendingPollRef.current.clientMessageId !== clientMessageId) return;
      pendingPollRef.current = {
        clientMessageId,
        attempt: pendingPollRef.current.attempt + 1,
      };
      loader.refresh();
    }, delay);
    return () => window.clearTimeout(timer);
  }, [loader.refresh, loader.state, serverUnresolvedMessage?.clientMessageId, serverUnresolvedMessage?.replyStatus]);

  const backToWorkspace = useCallback(() => navigate({ name: "workspace", params: { targetJobId: targetDisplay.targetJobId || "" } }), [navigate, targetDisplay.targetJobId]);

  const backToCurrentPlan = useCallback(() => {
    const targetJobId = loader.data?.targetJobId;
    if (!targetJobId) return;
    navigate({ name: "workspace", params: { targetJobId } });
  }, [loader.data?.targetJobId, navigate]);

  const runMessageRequest = useCallback(async (
    submission: { text: string; clientMessageId: string },
    status: "pending" | "retrying",
    createdAt: string,
  ) => {
    const seq = messageRequestSeqRef.current + 1;
    messageRequestSeqRef.current = seq;
    activeMessageRequestRef.current?.postController.abort();
    activeMessageRequestRef.current?.readController?.abort();
    const postController = new AbortController();
    const activeRequest: ActiveMessageRequest = { seq, postController };
    activeMessageRequestRef.current = activeRequest;
    const isCurrent = () => (
      sessionEpochRef.current.active
      && sessionEpochRef.current.sessionId === sessionId
      && messageRequestSeqRef.current === seq
      && activeMessageRequestRef.current?.seq === seq
    );
    setSending(true);
    setError(null);
    setErrorSource(null);
    const nextTransientMessage = { ...submission, createdAt, status };
    transientMessageRef.current = nextTransientMessage;
    setTransientMessage(nextTransientMessage);
    let timeout = 0;
    let timedOut = false;
    try {
      const timeoutFailure = new Promise<never>((_resolve, reject) => {
        timeout = window.setTimeout(() => {
          timedOut = true;
          postController.abort();
          reject(MESSAGE_POST_TIMEOUT);
        }, MESSAGE_POST_TIMEOUT_MS);
      });
      const result = await Promise.race([
        messages.sendMessage(submission, { signal: postController.signal }),
        timeoutFailure,
      ]);
      window.clearTimeout(timeout);
      timeout = 0;
      if (!isCurrent()) return;
      loader.adopt(result.session);
      transientMessageRef.current = null;
      setTransientMessage(null);
    } catch (cause) {
      window.clearTimeout(timeout);
      timeout = 0;
      if (!isCurrent()) return;
      const retryable = timedOut || isRetryableMessageFailure(cause);
      const localFailureMessage = retryable ? null : t(messageFailureFeedbackKey(cause));
      const preserveUnconfirmedFailure = () => {
        const current = transientMessageRef.current;
        if (current?.clientMessageId !== submission.clientMessageId) return;
        if (localFailureMessage) {
          setError(localFailureMessage);
          setErrorSource("message");
        }
        const failedMessage: TransientUserMessage = {
          ...current,
          status: retryable ? "retryable_failed" : "terminal_failed",
        };
        transientMessageRef.current = failedMessage;
        setTransientMessage(failedMessage);
      };
      const readController = new AbortController();
      activeRequest.readController = readController;
      try {
        const pendingRead = loader.read({ signal: readController.signal });
        const session = await pendingRead.result;
        if (!isCurrent()) return;
        const serverMessage = session.messages.find(
          (message) => message.role === "user" && message.clientMessageId === submission.clientMessageId,
        );
        const adopted = loader.adopt(session, { readToken: pendingRead.token });
        if (!adopted) {
          preserveUnconfirmedFailure();
          return;
        }
        if (serverMessage) {
          setError(null);
          setErrorSource(null);
          transientMessageRef.current = null;
          setTransientMessage(null);
        } else {
          preserveUnconfirmedFailure();
        }
      } catch {
        if (!isCurrent()) return;
        preserveUnconfirmedFailure();
      }
    } finally {
      if (timeout) window.clearTimeout(timeout);
      if (isCurrent()) {
        activeMessageRequestRef.current = null;
        setSending(false);
      }
    }
  }, [loader.adopt, loader.read, messages, sessionId, t]);

  const send = useCallback(() => {
    const text = input;
    if (!contentLimits || !text.trim() || sending || paused || hasUnresolvedReply) return;
    const textBytes = utf8ByteLength(text.trim());
    const sessionTextBytes = committedMessages.reduce(
      (total, message) => total + utf8ByteLength(message.content),
      0,
    );
    if (
      textBytes > contentLimits.practiceMessageBytes
      || sessionTextBytes + textBytes > contentLimits.practiceSessionTextBytes
    ) {
      setError(t("practice.errors.textTooLarge"));
      setErrorSource(null);
      return;
    }
    const clientMessageId = newId();
    const createdAt = new Date().toISOString();
    setInput("");
    void runMessageRequest({ text, clientMessageId }, "pending", createdAt);
  }, [committedMessages, contentLimits, hasUnresolvedReply, input, paused, runMessageRequest, sending, t]);

  const retryMessage = useCallback((message: TranscriptMessage) => {
    if (
      sending
      || message.role !== "user"
      || message.status !== "retryable_failed"
      || !message.clientMessageId
    ) return;
    void runMessageRequest(
      { text: message.text, clientMessageId: message.clientMessageId },
      "retrying",
      new Date().toISOString(),
    );
  }, [runMessageRequest, sending]);

  const finish = useCallback(async () => {
    if (!hasCommittedCandidateMessage || hasUnresolvedReply) return;
    const epoch = sessionEpochRef.current;
    const isCurrentSession = () => (
      epoch.active
      && epoch === sessionEpochRef.current
      && epoch.sessionId === sessionId
    );
    setError(null);
    setErrorSource(null);
    try {
      const report = await completion.complete();
      if (!isCurrentSession()) return;
      navigate({ name: "generating", params: { reportId: report.reportId } });
    } catch {
      if (!isCurrentSession()) return;
      setError(t("practice.errors.completionFailed"));
      setErrorSource("completion");
    }
  }, [completion, hasCommittedCandidateMessage, hasUnresolvedReply, navigate, sessionId, t]);

  if (!sessionId || loader.state === "sessionLost") return <PracticeSessionLostState onBack={backToWorkspace} />;

  const loaderBlocksInteraction = loader.state === "loading" || (loader.state === "error" && !hasRowLocalRetry);
  const inputDisabled = !contentLimits || paused || isThinking || terminalRecovery || loaderBlocksInteraction || !messages.ready || loader.data?.status === "completed" || loader.data?.status === "completing";
  const sendDisabled = inputDisabled || hasUnresolvedReply;
  const finishDisabled = paused || sending || loaderBlocksInteraction || !loader.data || !messages.ready || completion.state.kind === "loading" || !hasCommittedCandidateMessage || hasUnresolvedReply || (loader.data.status !== "running" && loader.data.status !== "waiting_user_input");
  const interviewerLabel = route.params.roundName || ctx.roundName || t("practice.toolbar.role.manager");
  const finishDisabledReason = !hasCommittedCandidateMessage && loader.data
    ? t("practice.finishDisabled.zeroAnswer")
    : isThinking
      ? t("practice.finishDisabled.awaitingReply")
      : hasUnresolvedReply
        ? t("practice.finishDisabled.unresolvedReply")
        : undefined;

  return (
    <div data-testid="practice-screen" className="ei-fadein" style={{ height: "100vh", display: "flex", flexDirection: "column", background: "var(--ei-color-bg-canvas)", overflow: "hidden" }}>
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
        finishCta={<FinishCta label={t("practice.finishCta")} disabled={finishDisabled} disabledReason={finishDisabledReason} onFinish={finish} />}
      />
      <main data-testid="practice-conversation" style={{ flex: 1, minHeight: 0, display: "flex", flexDirection: "column", width: "100%" }}>
        <Transcript
          messages={transcript}
          helperText={t("practice.transcript.helper")}
          aiLabel={t("practice.transcript.aiLabel")}
          userLabel={t("practice.transcript.userLabel")}
          thinking={isThinking}
          thinkingLabel={t("practice.transcript.thinking")}
          retryLabel={t("practice.message.retry")}
          onRetry={retryMessage}
        />
        <ErrorState
          message={terminalRecovery ? null : error || (loaderFailureNeedsGlobalRecovery ? t("practice.errors.sessionLoadFailed") : null)}
          retryLabel={t("practice.errors.retry")}
          onRetry={terminalRecovery ? undefined : errorSource === "completion" ? finish : errorSource === "message" ? loader.refresh : loaderFailureNeedsGlobalRecovery ? loader.refresh : undefined}
        />
        <InputBar
          value={input}
          onChange={setInput}
          placeholder={terminalRecovery
            ? t("practice.input.terminalPlaceholder")
            : isThinking
              ? t("practice.input.thinkingPlaceholder")
              : t("practice.input.placeholder")}
          sendLabel={t("practice.input.send")}
          disabled={inputDisabled}
          sendDisabled={sendDisabled}
          recovery={terminalRecovery && loader.data?.targetJobId ? (
            <TerminalRecovery
              title={t("practice.terminal.title")}
              description={t("practice.terminal.description")}
              ctaLabel={t("practice.terminal.backToPlan")}
              onBackToPlan={backToCurrentPlan}
            />
          ) : undefined}
          onSend={send}
        />
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

function isRetryableMessageFailure(cause: unknown): boolean {
  if (!(cause instanceof ApiClientError)) return false;
  if (cause.kind === "abort") return false;
  if (cause.kind === "transport") return true;
  return cause.apiError?.error.retryable === true;
}

function messageFailureFeedbackKey(cause: unknown): MessageKey {
  if (cause instanceof ApiClientError) {
    if (cause.kind === "abort") return "practice.errors.messageAborted";
    if (cause.kind === "http") return "practice.errors.messageRejected";
  }
  return "practice.errors.messageFailed";
}
