import type { FC } from "react";

import { useI18n } from "../../../i18n/messages";
import type { DebriefStep } from "../types";

interface DebriefStepperProps {
  step: DebriefStep;
  /**
   * Maximum step the user has reached so far. The stepper allows clicking
   * back to any visited step but never forward-jumping ahead of `maxVisited`.
   */
  maxVisited: DebriefStep;
  onStep: (next: DebriefStep) => void;
}

const STEPS: DebriefStep[] = [0, 1, 2];

const LABEL_KEYS = [
  "debrief.stepper.step0",
  "debrief.stepper.step1",
  "debrief.stepper.step2",
] as const;

/**
 * Source mirror of ui-design/src/screens-p1-depth.jsx::DebriefFullScreen
 * stepper block (lines 148-156). Three steps, current step highlighted via
 * accent underline; steps after `maxVisited` are disabled to enforce the
 * record → analysis → interview ordering specified in
 * docs/spec/frontend-debrief/spec.md §3.1.
 */
export const DebriefStepper: FC<DebriefStepperProps> = ({
  step,
  maxVisited,
  onStep,
}) => {
  const { t } = useI18n();
  return (
    <nav
      className="ei-debrief-stepper"
      data-testid="debrief-stepper"
      aria-label={t("debrief.stepper.ariaLabel")}
    >
      {STEPS.map((s) => {
        const active = s === step;
        const reachable = s <= maxVisited;
        return (
          <button
            key={s}
            type="button"
            className="ei-debrief-stepper__item"
            data-testid={`debrief-stepper-step-${s}`}
            data-active={active ? "true" : "false"}
            data-reachable={reachable ? "true" : "false"}
            aria-current={active ? "step" : undefined}
            disabled={!reachable}
            onClick={() => {
              if (reachable) onStep(s);
            }}
          >
            <span className="ei-debrief-stepper__num">
              {String(s + 1).padStart(2, "0")}
            </span>
            <span className="ei-debrief-stepper__label">
              {t(LABEL_KEYS[s])}
            </span>
          </button>
        );
      })}
    </nav>
  );
};
