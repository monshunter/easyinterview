import { useMemo, type FC } from "react";

import { useI18n } from "../../../i18n/messages";

export interface PracticeSessionLostStateProps {
  onBack: () => void;
}

/**
 * Source-level mirror of the spec §4 缺 session/plan 兜底 path. Rendered when
 * `getPracticeSession` / `sendPracticeMessage` / `completePracticeSession`
 * return 404, or when the route lacks a sessionId.
 */
export const PracticeSessionLostState: FC<PracticeSessionLostStateProps> = ({
  onBack,
}) => {
  const { t } = useI18n();
  const labels = useMemo(
    () => ({
      title: t("practice.sessionLost.title"),
      desc: t("practice.sessionLost.desc"),
      cta: t("common.back"),
    }),
    [t],
  );

  return (
    <div
      data-testid="practice-session-lost"
      style={{
        height: "100vh",
        display: "flex",
        alignItems: "center",
        justifyContent: "center",
        background: "var(--ei-color-bg-canvas)",
      }}
    >
      <div
        style={{
          background: "var(--ei-color-bg-card)",
          border: "1px solid var(--ei-color-rule-strong)",
          borderRadius: 3,
          padding: 32,
          textAlign: "center",
          maxWidth: 480,
        }}
      >
        <div
          data-testid="practice-session-lost-title"
          className="ei-serif"
          style={{
            fontSize: 18,
            color: "var(--ei-color-fg-primary)",
            marginBottom: 12,
          }}
        >
          {labels.title}
        </div>
        <div
          data-testid="practice-session-lost-desc"
          style={{
            fontSize: 13,
            color: "var(--ei-color-fg-tertiary)",
            lineHeight: 1.6,
            marginBottom: 20,
          }}
        >
          {labels.desc}
        </div>
        <button
          data-testid="practice-session-lost-cta"
          type="button"
          onClick={onBack}
          style={{
            padding: "8px 16px",
            fontSize: 13,
            background: "var(--ei-color-accent)",
            color: "#fff",
            border: "1px solid var(--ei-color-accent)",
            borderRadius: "var(--ei-radius-control)",
            cursor: "pointer",
          }}
        >
          {labels.cta}
        </button>
      </div>
    </div>
  );
};
