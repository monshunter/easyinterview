import type { FC } from "react";

import { useI18n } from "../../../i18n/messages";
import type { MessageKey } from "../../../i18n/locales/zh";

interface DebriefFailureStateProps {
  errorCode: string | null;
  onRetry: () => void;
  onBackToEdit: () => void;
}

function errorMessageKey(code: string | null): MessageKey {
  if (!code) return "debrief.failure.code.UNKNOWN";
  const candidate = `debrief.failure.code.${code}` as MessageKey;
  // We accept any code; renderer falls back to UNKNOWN if not localized.
  return candidate;
}

export const DebriefFailureState: FC<DebriefFailureStateProps> = ({
  errorCode,
  onRetry,
  onBackToEdit,
}) => {
  const { t } = useI18n();
  const key = errorMessageKey(errorCode);
  let message: string;
  try {
    message = t(key);
  } catch {
    message = t("debrief.failure.code.UNKNOWN");
  }
  return (
    <section
      className="ei-debrief-state-card ei-debrief-state-card--failure"
      data-testid="debrief-failure-state"
      data-error-code={errorCode ?? "UNKNOWN"}
    >
      <div className="ei-label">{t("debrief.failure.eyebrow")}</div>
      <h2 className="ei-serif">{t("debrief.failure.title")}</h2>
      <p data-testid="debrief-failure-message">{message}</p>
      <div className="ei-debrief-state-card__actions">
        <button
          type="button"
          className="ei-debrief-btn ei-debrief-btn--ghost"
          data-testid="debrief-failure-back"
          onClick={onBackToEdit}
        >
          {t("debrief.failure.back")}
        </button>
        <button
          type="button"
          className="ei-debrief-btn ei-debrief-btn--accent"
          data-testid="debrief-failure-retry"
          onClick={onRetry}
        >
          {t("debrief.failure.retry")}
        </button>
      </div>
    </section>
  );
};
