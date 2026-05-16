import type { FC, MouseEventHandler } from "react";

import { useI18n } from "../../../i18n/messages";

interface SlaHintProps {
  onNotify?: MouseEventHandler<HTMLButtonElement>;
}

/**
 * Source-level mirror of ui-design/src/screens-p0-complete.jsx
 * lines 388-394 (P95 mono hint + "Notify me when ready" ghost button).
 */
export const SlaHint: FC<SlaHintProps> = ({ onNotify }) => {
  const { t } = useI18n();
  return (
    <div
      data-testid="generating-sla-hint"
      style={{
        marginTop: 28,
        paddingTop: 16,
        borderTop: "1px solid var(--ei-rule)",
        display: "flex",
        justifyContent: "space-between",
        alignItems: "center",
      }}
    >
      <div
        data-testid="generating-sla-hint-target"
        style={{
          fontSize: 11,
          color: "var(--ei-ink3)",
          fontFamily: "var(--ei-mono)",
          letterSpacing: "0.04em",
        }}
      >
        {t("generating.sla.target")}
      </div>
      <button
        type="button"
        data-testid="generating-notify-cta"
        onClick={onNotify}
        style={{
          background: "transparent",
          border: "none",
          color: "var(--ei-ink2, var(--ei-ink))",
          fontSize: 13,
          cursor: "pointer",
          fontFamily: "var(--ei-sans)",
        }}
      >
        {t("generating.sla.notifyCta")}
      </button>
    </div>
  );
};
