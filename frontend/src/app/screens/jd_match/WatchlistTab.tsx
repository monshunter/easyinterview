import type { FC } from "react";

import { useI18n } from "../../i18n/messages";
import type {
  MarketSignal,
  WatchlistItem,
  WatchlistItemTone,
} from "../../../api/generated/types";

import { MarketSignalsGrid } from "./MarketSignalsGrid";

function watchlistToneColor(tone: WatchlistItemTone): string {
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

export interface WatchlistTabProps {
  items: WatchlistItem[];
  loading: boolean;
  error: Error | null;
  signals: MarketSignal[];
  signalsLoading: boolean;
  signalsError: Error | null;
  onChevron: (item: WatchlistItem) => void;
}

export const WatchlistTab: FC<WatchlistTabProps> = ({
  items,
  loading,
  error,
  signals,
  signalsLoading,
  signalsError,
  onChevron,
}) => {
  const { t } = useI18n();

  return (
    <div data-testid="jdmatch-watchlist-tab">
      <div style={{ marginBottom: 28 }}>
        <div
          className="ei-label"
          style={{ color: "var(--ei-color-fg-tertiary)", marginBottom: 10 }}
        >
          {t("jdMatch.watchlist.heading")}
        </div>
        {error ? (
          <div
            data-testid="jdmatch-watchlist-error"
            style={{
              padding: "20px 24px",
              background: "var(--ei-color-bg-soft)",
              border: "1px dashed var(--ei-color-rule-strong)",
              borderRadius: "var(--ei-radius-sm)",
              color: "var(--ei-color-warn)",
              fontSize: 13,
              lineHeight: 1.5,
            }}
          >
            {t("jdMatch.watchlist.error")}
          </div>
        ) : loading && items.length === 0 ? (
          <div
            data-testid="jdmatch-watchlist-loading"
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
        ) : items.length === 0 ? (
          <div
            data-testid="jdmatch-watchlist-empty"
            style={{
              padding: "32px 20px",
              textAlign: "center",
              fontSize: 13,
              color: "var(--ei-color-fg-tertiary)",
              fontFamily: "var(--ei-font-sans)",
              background: "var(--ei-color-bg-soft)",
              border: "1px dashed var(--ei-color-rule-strong)",
              borderRadius: "var(--ei-radius-sm)",
              lineHeight: 1.5,
            }}
          >
            {t("jdMatch.watchlist.empty")}
          </div>
        ) : (
          <div
            style={{ display: "flex", flexDirection: "column", gap: 10 }}
          >
            {items.map((item) => {
              const toneColor = watchlistToneColor(item.tone);
              return (
                <div
                  key={item.id}
                  data-testid={`jdmatch-watchlist-item-${item.id}`}
                  data-tone={item.tone}
                  style={{
                    padding: "16px 20px",
                    background: "var(--ei-color-bg-card)",
                    borderTop: "1px solid var(--ei-color-rule-strong)",
                    borderRight: "1px solid var(--ei-color-rule-strong)",
                    borderBottom: "1px solid var(--ei-color-rule-strong)",
                    borderLeft: `3px solid ${toneColor}`,
                    borderRadius: "var(--ei-radius-sm)",
                    display: "flex",
                    alignItems: "center",
                    gap: 16,
                  }}
                >
                  <div style={{ flex: 1 }}>
                    <div
                      className="ei-serif"
                      style={{
                        fontSize: 16,
                        color: "var(--ei-color-fg-primary)",
                        letterSpacing: "-0.01em",
                        marginBottom: 4,
                      }}
                    >
                      {item.title}
                    </div>
                    <div
                      style={{
                        fontSize: 12.5,
                        color: "var(--ei-color-fg-tertiary)",
                      }}
                    >
                      {item.company}{" "}
                      {`· ${t("jdMatch.watchlist.addedTimeLabel")} `}
                      {formatAddedAt(item.addedAt)}
                    </div>
                  </div>
                  {item.change ? (
                    <div
                      data-testid={`jdmatch-watchlist-item-${item.id}-change`}
                      style={{
                        fontSize: 12.5,
                        color: toneColor,
                        fontFamily: "var(--ei-font-mono)",
                        letterSpacing: "0.04em",
                        display: "flex",
                        alignItems: "center",
                        gap: 8,
                      }}
                    >
                      <span
                        style={{
                          display: "inline-block",
                          width: 6,
                          height: 6,
                          borderRadius: 3,
                          background: toneColor,
                        }}
                      />
                      {item.change}
                    </div>
                  ) : null}
                  <button
                    type="button"
                    data-testid={`jdmatch-watchlist-item-${item.id}-chevron`}
                    onClick={() => onChevron(item)}
                    aria-label={t("jdMatch.watchlist.chevronTooltip")}
                    style={{
                      padding: "6px 10px",
                      background: "transparent",
                      color: "var(--ei-color-fg-tertiary)",
                      border: "1px solid var(--ei-color-rule-strong)",
                      borderRadius: "var(--ei-radius-sm)",
                      cursor: "pointer",
                      fontFamily: "var(--ei-font-mono)",
                      fontSize: 13,
                    }}
                  >
                    ›
                  </button>
                </div>
              );
            })}
          </div>
        )}
      </div>
      <MarketSignalsGrid
        signals={signals}
        loading={signalsLoading}
        error={signalsError}
      />
      <div
        data-testid="jdmatch-watchlist-refresh-footer"
        style={{
          fontSize: 12,
          color: "var(--ei-color-fg-tertiary)",
          fontFamily: "var(--ei-font-mono)",
          textAlign: "center",
          padding: 10,
          background: "var(--ei-color-bg-soft)",
          borderRadius: "var(--ei-radius-sm)",
        }}
      >
        {t("jdMatch.watchlist.refreshFooter")}
      </div>
    </div>
  );
};

function formatAddedAt(addedAt: string): string {
  const t = new Date(addedAt);
  if (Number.isNaN(t.getTime())) return addedAt;
  // Render short ISO date (YYYY-MM-DD) so we never embed locale-specific
  // formatting that could surprise privacy reverse-grep.
  return t.toISOString().slice(0, 10);
}
