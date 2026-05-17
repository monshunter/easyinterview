import { useCallback, type FC } from "react";

import { useI18n, type MessageKey } from "../../../i18n/messages";
import { ResumeWorkshopIcon } from "../components/ResumeWorkshopIcon";
import { useResumeRegistration } from "./hooks/useResumeRegistration";
import type { GuidedAnswers } from "./ResumeCreateFlow";
import { deriveDefaultTitle } from "./util/title";

export interface GuidedStepDescriptor {
  key: keyof GuidedAnswers;
  label: MessageKey;
  question: MessageKey;
  placeholder: MessageKey;
}

export interface GuidedTabProps {
  guidedAnswers: GuidedAnswers;
  guidedStep: number;
  steps: GuidedStepDescriptor[];
  submitting: boolean;
  inlineError: string | null;
  onAnswerChange: (key: keyof GuidedAnswers, value: string) => void;
  onSelectStep: (index: number) => void;
  onAdvanceStep: () => void;
  onBackStep: () => void;
  onRegistered: (resumeAssetId: string, sourceLabel: string) => void;
  setSubmitting: (value: boolean) => void;
  setInlineError: (message: string | null) => void;
}

export const GuidedTab: FC<GuidedTabProps> = ({
  guidedAnswers,
  guidedStep,
  steps,
  submitting,
  inlineError,
  onAnswerChange,
  onSelectStep,
  onAdvanceStep,
  onBackStep,
  onRegistered,
  setSubmitting,
  setInlineError,
}) => {
  const { t, lang } = useI18n();
  const register = useResumeRegistration();

  const isLastStep = guidedStep === steps.length - 1;
  const isFirstStep = guidedStep === 0;
  const currentStep = steps[guidedStep]!;

  const handleSubmit = useCallback(async () => {
    setSubmitting(true);
    setInlineError(null);
    try {
      const title = deriveDefaultTitle("guided", lang, null);
      const registered = await register.register({
        sourceType: "guided",
        guidedAnswers: { ...guidedAnswers },
        title,
        language: lang,
      });
      onRegistered(
        registered.resumeAssetId,
        t("resumeWorkshop.create.guided.titleFallback"),
      );
    } catch (error) {
      const message =
        error instanceof Error
          ? mapRegisterError(error.message, t)
          : t("resumeWorkshop.create.errors.registerFailed");
      setInlineError(message);
    } finally {
      setSubmitting(false);
    }
  }, [
    guidedAnswers,
    lang,
    onRegistered,
    register,
    setInlineError,
    setSubmitting,
    t,
  ]);

  return (
    <div
      className="ei-resume-create-guided"
      data-testid="resume-create-guided-panel"
    >
      <nav
        className="ei-resume-create-guided-steps"
        aria-label={t("resumeWorkshop.create.guided.eyebrow")}
      >
        {steps.map((step, index) => {
          const active = index === guidedStep;
          return (
            <button
              key={step.key}
              type="button"
              className="ei-resume-create-guided-step"
              data-testid={`resume-create-guided-step-${index + 1}`}
              data-active={active}
              aria-current={active ? "step" : undefined}
              onClick={() => onSelectStep(index)}
            >
              <span className="ei-resume-create-guided-step-index">
                {index + 1}
              </span>
              <span className="ei-resume-create-guided-step-label">
                {t(step.label)}
              </span>
            </button>
          );
        })}
      </nav>
      <div className="ei-resume-create-guided-body">
        <span className="ei-text-label ei-resume-create-guided-eyebrow">
          {t("resumeWorkshop.create.guided.eyebrow")}
        </span>
        <h2
          className="ei-text-title ei-resume-create-guided-question"
          data-testid="resume-create-guided-question"
        >
          {t(currentStep.question)}
        </h2>
        <textarea
          className="ei-resume-create-guided-textarea"
          data-testid="resume-create-guided-textarea"
          placeholder={t(currentStep.placeholder)}
          value={guidedAnswers[currentStep.key]}
          spellCheck={false}
          onChange={(event) =>
            onAnswerChange(currentStep.key, event.target.value)
          }
        />
        <div className="ei-resume-create-guided-footer">
          <p className="ei-resume-create-guided-helper">
            {t("resumeWorkshop.create.guided.helper")}
          </p>
          <div className="ei-resume-create-guided-controls">
            <button
              type="button"
              className="ei-resume-create-cta-ghost"
              data-testid="resume-create-guided-back"
              disabled={isFirstStep || submitting}
              onClick={onBackStep}
            >
              {t("resumeWorkshop.create.guided.back")}
            </button>
            <button
              type="button"
              className="ei-resume-create-cta-accent"
              data-testid="resume-create-guided-advance"
              disabled={submitting}
              onClick={() => {
                if (isLastStep) {
                  void handleSubmit();
                } else {
                  onAdvanceStep();
                }
              }}
            >
              {isLastStep
                ? t("resumeWorkshop.create.guided.generate")
                : t("resumeWorkshop.create.guided.next")}
              {!isLastStep ? (
                <ResumeWorkshopIcon name="arrowRight" size={14} />
              ) : null}
            </button>
          </div>
        </div>
      </div>
      {inlineError ? (
        <div
          className="ei-resume-create-error"
          role="alert"
          data-testid="resume-create-guided-error"
        >
          {inlineError}
        </div>
      ) : null}
    </div>
  );
};

function mapRegisterError(
  message: string,
  t: (key: Parameters<ReturnType<typeof useI18n>["t"]>[0]) => string,
): string {
  if (/VALIDATION_FAILED/i.test(message)) {
    return t("resumeWorkshop.create.errors.validation");
  }
  return t("resumeWorkshop.create.errors.registerFailed");
}
