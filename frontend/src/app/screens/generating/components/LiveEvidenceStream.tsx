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
        style={{ color: "var(--ei-ink3)", marginBottom: 10 }}
      >
        {t("generating.evidence.streamLabel")}
      </div>
      <div
        data-testid="generating-live-stream-body"
        style={{
          padding: "14px 16px",
          background: "var(--ei-bg-soft, var(--ei-bg))",
          border: "1px solid var(--ei-rule)",
          borderRadius: 2,
          fontFamily: "var(--ei-mono)",
          fontSize: 12,
          lineHeight: 1.75,
          color: "var(--ei-ink2, var(--ei-ink))",
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
            <span style={{ color: "var(--ei-ink4, var(--ei-ink3))" }}>›</span>{" "}
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
              background: "var(--ei-accent)",
              verticalAlign: "text-bottom",
            }}
          />
        ) : null}
      </div>
    </div>
  );
};
