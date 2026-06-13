import { useEffect, useMemo, useRef, useState, type FC } from "react";

import type {
  ApiErrorCode,
  FeedbackReport,
} from "../../../api/generated/types";
import { useI18n, type MessageKey } from "../../i18n/messages";
import { useInterviewContext } from "../../interview-context/InterviewContext";
import { useNavigation } from "../../navigation/NavigationProvider";
import type { Route } from "../../routes";
import { GeneratingErrorState } from "./components/GeneratingErrorState";
import { HeaderHero } from "./components/HeaderHero";
import { LiveEvidenceStream } from "./components/LiveEvidenceStream";
import { PhaseList, type PhaseDefinition } from "./components/PhaseList";
import { ProgressBar } from "./components/ProgressBar";
import { SlaHint } from "./components/SlaHint";
import { useReportGenerationPoll } from "./hooks/useReportGenerationPoll";

interface GeneratingScreenProps {
  route: Route;
}

const PHASES: PhaseDefinition[] = [
  { labelKey: "generating.phase.1", hintKey: "generating.phase.1.hint" },
  { labelKey: "generating.phase.2", hintKey: "generating.phase.2.hint" },
  { labelKey: "generating.phase.3", hintKey: "generating.phase.3.hint" },
  { labelKey: "generating.phase.4", hintKey: "generating.phase.4.hint" },
  { labelKey: "generating.phase.5", hintKey: "generating.phase.5.hint" },
];

const EVIDENCE_KEYS: MessageKey[] = [
  "generating.evidence.line.1",
  "generating.evidence.line.2",
  "generating.evidence.line.3",
  "generating.evidence.line.4",
];

const PHASE_TICK_MS = 900;
const EVIDENCE_DRIP_MS = 800;

const HANDOFF_PASSTHROUGH_KEYS = [
  "planId",
  "targetJobId",
  "jdId",
  "resumeId",
  "roundId",
  "roundName",
  "mode",
  "modality",
  "practiceMode",
  "practiceGoal",
  "hintUsed",
  "hintCount",
] as const;

/**
 * Source-level mirror of `ui-design/src/screens-p0-complete.jsx::ReportGeneratingScreen`
 * (lines 269-399). Wires the static composition to the real polling hook:
 * GeneratingScreen drives the visual phase indicator + evidence stream while
 * `useReportGenerationPoll` watches `getFeedbackReport(reportId)` and triggers
 * a single navigation to `report` once status flips to ready / failed.
 *
 * Privacy red lines:
 *   - URL params carry only 7 owner IDs + 6 display knobs + sessionId / reportId.
 *   - Raw answer / question / hint text never reaches route or console.
 *   - getFeedbackReport never receives an Idempotency-Key header.
 */
export const GeneratingScreen: FC<GeneratingScreenProps> = ({ route }) => {
  const { t } = useI18n();
  const { navigate } = useNavigation();
  const { ctx } = useInterviewContext();

  const reportId = route.params.reportId ?? "";
  const sessionId = route.params.sessionId || ctx.sessionId || "";

  // Visual progress is independent of polling cadence: ui-design ticks the 5
  // phases every ~900ms while the AI runs. The hook is the truth source for
  // when we leave generating — visual phase clamps to "complete" once ready.
  const [phaseIndex, setPhaseIndex] = useState(0);
  useEffect(() => {
    if (!reportId) return undefined;
    let cancelled = false;
    let cursor = 0;
    const tick = () => {
      if (cancelled) return;
      cursor += 1;
      setPhaseIndex(Math.min(cursor, PHASES.length));
      if (cursor < PHASES.length) {
        timer = setTimeout(tick, PHASE_TICK_MS);
      }
    };
    let timer = setTimeout(tick, PHASE_TICK_MS);
    return () => {
      cancelled = true;
      clearTimeout(timer);
    };
  }, [reportId]);

  const [visibleEvidence, setVisibleEvidence] = useState<MessageKey[]>([]);
  useEffect(() => {
    if (!reportId) return undefined;
    let cancelled = false;
    const timers: ReturnType<typeof setTimeout>[] = [];
    EVIDENCE_KEYS.forEach((key, i) => {
      timers.push(
        setTimeout(() => {
          if (cancelled) return;
          setVisibleEvidence((prev) => [...prev, key]);
        }, 700 + i * EVIDENCE_DRIP_MS),
      );
    });
    return () => {
      cancelled = true;
      timers.forEach((timer) => clearTimeout(timer));
    };
  }, [reportId]);

  // Debounce nav: even if onReady fires from multiple renders (e.g. fast
  // backoff) only navigate once.
  const handoffNavigatedRef = useRef(false);

  const handoffParams = useMemo(() => {
    const params: Record<string, string> = {};
    for (const key of HANDOFF_PASSTHROUGH_KEYS) {
      const value = route.params[key];
      if (value) params[key] = value;
    }
    if (sessionId) params.sessionId = sessionId;
    if (reportId) params.reportId = reportId;
    return params;
  }, [reportId, route.params, sessionId]);

  const handleReady = (report: FeedbackReport) => {
    if (handoffNavigatedRef.current) return;
    handoffNavigatedRef.current = true;
    navigate({
      name: "report",
      params: {
        ...handoffParams,
        reportId: report.id || handoffParams.reportId || "",
        sessionId: report.sessionId || handoffParams.sessionId || "",
      },
    });
  };

  const handleFailed = (errorCode: ApiErrorCode | string) => {
    if (handoffNavigatedRef.current) return;
    handoffNavigatedRef.current = true;
    navigate({
      name: "report",
      params: {
        ...handoffParams,
        reportStatus: "failed",
        errorCode: String(errorCode),
      },
    });
  };

  const poll = useReportGenerationPoll({
    reportId,
    onReady: handleReady,
    onFailed: handleFailed,
  });

  const goWorkspace = () => {
    navigate({ name: "workspace", params: handoffParams });
  };

  if (!reportId) {
    return (
      <GeneratingErrorState
        kind="missingReportId"
        onBackToWorkspace={goWorkspace}
      />
    );
  }

  if (poll.state === "timeout") {
    return (
      <GeneratingErrorState
        kind="timeout"
        onRetry={poll.retry}
        onBackToWorkspace={goWorkspace}
      />
    );
  }

  const resolve = (key: string) => t(key as MessageKey);
  const visualPhaseIndex = poll.state === "ready" ? PHASES.length : phaseIndex;
  const activePhaseLabel =
    visualPhaseIndex < PHASES.length
      ? resolve(PHASES[visualPhaseIndex]!.labelKey)
      : t("generating.progress.done");
  const evidence = visibleEvidence.map((key) => resolve(key));
  const hasMoreEvidence = visibleEvidence.length < EVIDENCE_KEYS.length;

  return (
    <div
      data-testid="generating-screen"
      className="ei-fadein"
      style={{
        minHeight: "calc(100vh - 58px)",
        background: "var(--ei-color-bg-canvas)",
        display: "flex",
        alignItems: "center",
        justifyContent: "center",
        padding: 48,
      }}
    >
      <div style={{ maxWidth: 780, width: "100%" }}>
        <HeaderHero />
        <ProgressBar
          phaseIndex={visualPhaseIndex}
          totalPhases={PHASES.length}
          activePhaseLabel={activePhaseLabel}
        />
        <PhaseList
          phaseIndex={visualPhaseIndex}
          phases={PHASES}
          resolve={resolve}
        />
        <LiveEvidenceStream evidence={evidence} hasMore={hasMoreEvidence} />
        <SlaHint onNotify={() => navigate({ name: "home", params: {} })} />
      </div>
    </div>
  );
};
