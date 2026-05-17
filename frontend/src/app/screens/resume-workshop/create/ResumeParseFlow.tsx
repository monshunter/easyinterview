import { useEffect, useState, type FC } from "react";

import { useI18n } from "../../../i18n/messages";
import { ResumeWorkshopIcon } from "../components/ResumeWorkshopIcon";

const PARSE_STEPS = [
  "resumeWorkshop.parsing.step.extract",
  "resumeWorkshop.parsing.step.identity",
  "resumeWorkshop.parsing.step.experience",
  "resumeWorkshop.parsing.step.projects",
  "resumeWorkshop.parsing.step.skills",
  "resumeWorkshop.parsing.step.education",
  "resumeWorkshop.parsing.step.structure",
] as const;

const TICK_INTERVAL_MS = 700;

export type ResumeParseState =
  | { phase: "polling" }
  | {
      phase: "failed";
      errorCode: string;
    }
  | { phase: "ready" };

export interface ResumeParseFlowProps {
  sourceLabel: string;
  parseState: ResumeParseState;
  onCancel: () => void;
  onRetry: () => void;
  /** Test seam: when true, skip the timer that drives the step animation. */
  skipTickerAnimation?: boolean;
}

export const ResumeParseFlow: FC<ResumeParseFlowProps> = ({
  sourceLabel,
  parseState,
  onCancel,
  onRetry,
  skipTickerAnimation,
}) => {
  const { t } = useI18n();
  const [active, setActive] = useState(0);

  useEffect(() => {
    if (skipTickerAnimation) return;
    if (parseState.phase !== "polling") return;
    if (active >= PARSE_STEPS.length - 1) return;
    const timer = window.setTimeout(
      () => setActive((value) => Math.min(value + 1, PARSE_STEPS.length - 1)),
      TICK_INTERVAL_MS,
    );
    return () => window.clearTimeout(timer);
  }, [active, parseState.phase, skipTickerAnimation]);

  if (parseState.phase === "failed") {
    const codeKey = `resumeWorkshop.parsing.failed.code.${parseState.errorCode}`;
    const knownKeys = new Set<string>([
      "resumeWorkshop.parsing.failed.code.AI_TIMEOUT_RETRYABLE",
      "resumeWorkshop.parsing.failed.code.PARSE_TIMEOUT",
      "resumeWorkshop.parsing.failed.code.AI_PROVIDER_TIMEOUT",
      "resumeWorkshop.parsing.failed.code.AI_OUTPUT_INVALID",
    ]);
    const resolvedKey = knownKeys.has(codeKey)
      ? codeKey
      : "resumeWorkshop.parsing.failed.code.UNKNOWN";
    return (
      <div
        className="ei-resume-create-parse"
        data-testid="resume-parse-flow"
        data-phase="failed"
        data-error-code={parseState.errorCode}
      >
        <button
          type="button"
          data-testid="resume-parse-flow-cancel"
          className="ei-resume-create-back"
          onClick={onCancel}
        >
          <ResumeWorkshopIcon name="arrowLeft" size={14} />
          {t("resumeWorkshop.parsing.cancel")}
        </button>
        <div
          className="ei-resume-create-parse-failed"
          data-testid="resume-parse-failed-state"
        >
          <h2 className="ei-text-title">
            {t("resumeWorkshop.parsing.failed.title")}
          </h2>
          <p className="ei-text-body">
            {t(
              resolvedKey as Parameters<typeof t>[0],
            )}
          </p>
          <div className="ei-resume-create-parse-failed-actions">
            <button
              type="button"
              data-testid="resume-parse-flow-retry"
              className="ei-resume-create-cta-accent"
              onClick={onRetry}
            >
              {t("resumeWorkshop.parsing.failed.retry")}
            </button>
            <button
              type="button"
              data-testid="resume-parse-flow-back"
              className="ei-resume-create-cta-ghost"
              onClick={onCancel}
            >
              {t("resumeWorkshop.parsing.failed.back")}
            </button>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div
      className="ei-resume-create-parse"
      data-testid="resume-parse-flow"
      data-phase={parseState.phase}
    >
      <button
        type="button"
        data-testid="resume-parse-flow-cancel"
        className="ei-resume-create-back"
        onClick={onCancel}
      >
        <ResumeWorkshopIcon name="arrowLeft" size={14} />
        {t("resumeWorkshop.parsing.cancel")}
      </button>
      <header className="ei-resume-create-parse-header">
        <span className="ei-text-label ei-resume-create-parse-eyebrow">
          {t("resumeWorkshop.parsing.eyebrow")}
        </span>
        <h1 className="ei-text-display">{t("resumeWorkshop.parsing.title")}</h1>
        <p
          className="ei-resume-create-parse-source"
          data-testid="resume-parse-flow-source"
        >
          <span>{t("resumeWorkshop.parsing.sourcePrefix")}</span>
          {sourceLabel}
        </p>
      </header>
      <section
        className="ei-resume-create-parse-card"
        aria-live="polite"
        aria-busy={parseState.phase === "polling"}
      >
        <div className="ei-text-label ei-resume-create-parse-agent">
          <span className="ei-resume-create-parse-agent-dot" aria-hidden="true" />
          {t("resumeWorkshop.parsing.agent")}
        </div>
        <ul className="ei-resume-create-parse-steps">
          {PARSE_STEPS.map((stepKey, index) => {
            const status =
              index < active ? "done" : index === active ? "active" : "pending";
            return (
              <li
                key={stepKey}
                className="ei-resume-create-parse-step"
                data-testid={`resume-parse-step-${index}`}
                data-status={status}
              >
                <span
                  className="ei-resume-create-parse-step-marker"
                  aria-hidden="true"
                >
                  {status === "done"
                    ? <ResumeWorkshopIcon name="check" size={12} />
                    : status === "active"
                      ? "→"
                      : "·"}
                </span>
                <span className="ei-resume-create-parse-step-label">
                  {t(stepKey as Parameters<typeof t>[0])}
                </span>
              </li>
            );
          })}
        </ul>
      </section>
    </div>
  );
};
