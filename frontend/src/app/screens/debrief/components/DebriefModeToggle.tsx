import type { FC } from "react";

import { useI18n } from "../../../i18n/messages";
import type { DebriefInputMode } from "../types";

interface DebriefModeToggleProps {
  inputMode: DebriefInputMode;
  onChange: (next: DebriefInputMode) => void;
}

export const DebriefModeToggle: FC<DebriefModeToggleProps> = ({
  inputMode,
  onChange,
}) => {
  const { t } = useI18n();
  return (
    <div
      className="ei-debrief-mode-toggle"
      data-testid="debrief-mode-toggle"
      data-mode={inputMode}
    >
      <span className="ei-label">{t("debrief.record.mode.label")}</span>
      <div className="ei-debrief-mode-toggle__group" role="tablist">
        <button
          type="button"
          role="tab"
          aria-selected={inputMode === "text"}
          className="ei-debrief-mode-toggle__btn"
          data-testid="debrief-mode-toggle-text"
          data-active={inputMode === "text"}
          onClick={() => onChange("text")}
        >
          {t("debrief.record.mode.text")}
        </button>
        <button
          type="button"
          role="tab"
          aria-selected={inputMode === "voice"}
          className="ei-debrief-mode-toggle__btn"
          data-testid="debrief-mode-toggle-voice"
          data-active={inputMode === "voice"}
          onClick={() => onChange("voice")}
        >
          {t("debrief.record.mode.voice")}
        </button>
      </div>
      <span className="ei-debrief-mode-toggle__hint">
        {inputMode === "voice"
          ? t("debrief.record.mode.hintVoice")
          : t("debrief.record.mode.hintText")}
      </span>
    </div>
  );
};
