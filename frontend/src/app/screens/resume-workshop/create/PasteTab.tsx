import { useCallback, type FC } from "react";

import { useI18n } from "../../../i18n/messages";
import { ResumeWorkshopIcon } from "../components/ResumeWorkshopIcon";
import { useResumeRegistration } from "./hooks/useResumeRegistration";
import { deriveDefaultTitle } from "./util/title";

export interface PasteTabProps {
  rawText: string;
  submitting: boolean;
  inlineError: string | null;
  onRawTextChange: (text: string) => void;
  onRegistered: (resumeId: string, sourceLabel: string) => void;
  setSubmitting: (value: boolean) => void;
  setInlineError: (message: string | null) => void;
}

export const PasteTab: FC<PasteTabProps> = ({
  rawText,
  submitting,
  inlineError,
  onRawTextChange,
  onRegistered,
  setSubmitting,
  setInlineError,
}) => {
  const { t, lang } = useI18n();
  const register = useResumeRegistration();

  const trimmed = rawText.trim();
  const canSubmit = trimmed.length > 0 && !submitting;

  const handleSubmit = useCallback(async () => {
    if (!canSubmit) return;
    setSubmitting(true);
    setInlineError(null);
    try {
      const title = deriveDefaultTitle("paste", lang, null);
      const registered = await register.register({
        sourceType: "paste",
        rawText,
        title,
        language: lang,
      });
      onRegistered(registered.resumeId, title);
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
    canSubmit,
    lang,
    onRegistered,
    rawText,
    register,
    setInlineError,
    setSubmitting,
    t,
  ]);

  return (
    <div
      className="ei-resume-create-paste"
      data-testid="resume-create-paste-panel"
    >
      <textarea
        className="ei-resume-create-paste-textarea"
        data-testid="resume-create-paste-textarea"
        placeholder={t("resumeWorkshop.create.paste.placeholder")}
        value={rawText}
        onChange={(event) => onRawTextChange(event.target.value)}
        spellCheck={false}
      />
      <div className="ei-resume-create-paste-footer">
        <p className="ei-resume-create-paste-helper">
          {t("resumeWorkshop.create.paste.helper")}
        </p>
        <button
          type="button"
          className="ei-resume-create-cta-accent"
          data-testid="resume-create-paste-submit"
          disabled={!canSubmit}
          onClick={handleSubmit}
        >
          <ResumeWorkshopIcon name="sparkle" size={14} />
          {t("resumeWorkshop.create.paste.submit")}
        </button>
      </div>
      {inlineError ? (
        <div
          className="ei-resume-create-error"
          role="alert"
          data-testid="resume-create-paste-error"
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
