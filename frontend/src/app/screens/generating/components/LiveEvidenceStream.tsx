import type { FC } from "react";

import { useI18n } from "../../../i18n/messages";

interface LiveEvidenceStreamProps {
  evidence: string[];
  hasMore: boolean;
}

/**
 * Source-level mirror of ui-design/src/screens-p0-complete.jsx
 * lines 374-386. Renders the LIVE OBSERVATIONS label and a mono-stream area
 * with fade-in fragments and a blinking cursor while more evidence pending.
 */
export const LiveEvidenceStream: FC<LiveEvidenceStreamProps> = ({
  evidence,
  hasMore,
}) => {
  const { t } = useI18n();
  return (
    <div data-testid="generating-live-stream">
      <div
        className="ei-label"
        data-testid="generating-live-stream-label"
        style={{ color: "var(--ei-color-fg-tertiary)", marginBottom: 10 }}
      >
        {t("generating.evidence.streamLabel")}
      </div>
      <div
        data-testid="generating-live-stream-body"
        style={{
          padding: "14px 16px",
          background: "var(--ei-color-bg-soft, var(--ei-color-bg-canvas))",
          border: "1px solid var(--ei-color-rule-soft)",
          borderRadius: 2,
          fontFamily: "var(--ei-font-mono)",
          fontSize: 12,
          lineHeight: 1.75,
          color: "var(--ei-color-fg-secondary, var(--ei-color-fg-primary))",
          minHeight: 100,
        }}
      >
        {evidence.map((line, i) => (
          <div
            key={`${i}-${line.slice(0, 8)}`}
            data-testid={`generating-evidence-${i}`}
            className="ei-fadein"
            style={{ marginBottom: 4 }}
          >
            <span style={{ color: "var(--ei-ink4, var(--ei-color-fg-tertiary))" }}>›</span>{" "}
            {line}
          </div>
        ))}
        {hasMore ? (
          <span
            data-testid="generating-live-stream-cursor"
            className="ei-pulse"
            style={{
              display: "inline-block",
              width: 8,
              height: 12,
              background: "var(--ei-color-accent)",
              verticalAlign: "text-bottom",
            }}
          />
        ) : null}
      </div>
    </div>
  );
};
