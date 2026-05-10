import type { FC } from "react";

import { useI18n } from "../../i18n/messages";
import type { MarketSignal } from "../../../api/generated/types";

function toneColor(tone: MarketSignal["tone"]): string {
  switch (tone) {
    case "ok":
      return "var(--ei-color-ok)";
    case "warn":
      return "var(--ei-color-warn)";
    case "muted":
    default:
      return "var(--ei-color-fg-tertiary)";
  }
}

export interface MarketSignalsGridProps {
  signals: MarketSignal[];
  loading: boolean;
  error: Error | null;
}

export const MarketSignalsGrid: FC<MarketSignalsGridProps> = ({
  signals,
  loading,
  error,
}) => {
  const { t } = useI18n();
  return (
    <div data-testid="jdmatch-market-signals-grid">
      <div
        className="ei-label"
        style={{ color: "var(--ei-color-fg-tertiary)", marginBottom: 10 }}
      >
        {t("jdMatch.watchlist.marketSignalsHeading")}
      </div>
      {error ? (
        <div
          data-testid="jdmatch-market-signals-error"
          style={{
            padding: "20px 24px",
            background: "var(--ei-color-bg-soft)",
            border: "1px dashed var(--ei-color-rule-strong)",
            borderRadius: "var(--ei-radius-sm)",
            color: "var(--ei-color-warn)",
            fontSize: 13,
          }}
        >
          {t("jdMatch.watchlist.marketSignalUnavailable")}
        </div>
      ) : loading && signals.length === 0 ? (
        <div
          data-testid="jdmatch-market-signals-loading"
          style={{
            padding: "20px 24px",
            background: "var(--ei-color-bg-soft)",
            border: "1px dashed var(--ei-color-rule-strong)",
            borderRadius: "var(--ei-radius-sm)",
            color: "var(--ei-color-fg-tertiary)",
            fontSize: 13,
            textAlign: "center",
            fontFamily: "var(--ei-font-mono)",
          }}
        >
          …
        </div>
      ) : (
        <div
          data-testid="jdmatch-market-signals-inner"
          className="jdmatch-market-signals-inner"
          style={{
            display: "grid",
            gridTemplateColumns: "repeat(4, 1fr)",
            gap: 10,
            marginBottom: 20,
          }}
        >
          {signals.map((signal, i) => {
            const dColor = toneColor(signal.tone);
            return (
              <div
                key={`${signal.k}-${i}`}
                data-testid={`jdmatch-market-signal-${i}`}
                data-signal-key={signal.k}
                data-signal-tone={signal.tone}
                style={{
                  padding: "14px 16px",
                  background: "var(--ei-color-bg-card)",
                  border: "1px solid var(--ei-color-rule-strong)",
                  borderRadius: "var(--ei-radius-sm)",
                }}
              >
                <div
                  className="ei-label"
                  style={{
                    color: "var(--ei-color-fg-tertiary)",
                    marginBottom: 6,
                  }}
                >
                  {signal.k}
                </div>
                <div
                  style={{
                    display: "flex",
                    alignItems: "baseline",
                    gap: 8,
                  }}
                >
                  <div
                    className="ei-serif"
                    style={{
                      fontSize: 24,
                      color: "var(--ei-color-fg-primary)",
                      letterSpacing: "-0.015em",
                    }}
                  >
                    {signal.v}
                  </div>
                  {signal.d ? (
                    <div
                      style={{
                        fontSize: 11.5,
                        color: dColor,
                        fontFamily: "var(--ei-font-mono)",
                      }}
                    >
                      {signal.d}
                    </div>
                  ) : (
                    <div
                      data-testid={`jdmatch-market-signal-${i}-fallback`}
                      style={{
                        fontSize: 11.5,
                        color: "var(--ei-color-fg-muted)",
                        fontFamily: "var(--ei-font-mono)",
                      }}
                    >
                      —
                    </div>
                  )}
                </div>
              </div>
            );
          })}
        </div>
      )}
    </div>
  );
};
