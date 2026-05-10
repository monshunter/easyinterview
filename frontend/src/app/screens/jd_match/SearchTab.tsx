import type { FC } from "react";

import { useI18n } from "../../i18n/messages";
import type {
  JobMatchRecommendation,
  SavedSearch,
} from "../../../api/generated/types";

import { JobMatchCard } from "./JobMatchCard";

export type SearchResultFilter = "all" | "strong" | "remote" | "unseen";

const FILTER_KEYS: SearchResultFilter[] = ["all", "strong", "remote", "unseen"];

const RESULTS_CAP = 6;

const SOURCE_KEYS = [
  "linkedin",
  "boss",
  "maimai",
  "lagou",
] as const;

export interface SearchTabProps {
  query: string;
  searching: boolean;
  results: JobMatchRecommendation[];
  savedSearches: SavedSearch[];
  resultFilter: SearchResultFilter;
  error: Error | null;
  savedSearchesError: Error | null;
  savedSearchCreating: boolean;
  savedSearchCreateError: Error | null;
  hasRunOnce?: boolean;
  setQuery: (next: string) => void;
  onRun: () => void;
  onSaveCurrent: () => void;
  setResultFilter: (key: SearchResultFilter) => void;
  onOpenJob: (rec: JobMatchRecommendation) => void;
  onCreateSavedSearchRetry: () => void;
}

export const SearchTab: FC<SearchTabProps> = ({
  query,
  searching,
  results,
  savedSearches,
  resultFilter,
  error,
  savedSearchesError,
  savedSearchCreating,
  savedSearchCreateError,
  hasRunOnce,
  setQuery,
  onRun,
  onSaveCurrent,
  setResultFilter,
  onOpenJob,
  onCreateSavedSearchRetry,
}) => {
  const { t } = useI18n();

  const cappedResults = results.slice(0, RESULTS_CAP);

  return (
    <div data-testid="jdmatch-search-tab">
      {/* Search bar */}
      <div
        style={{
          background: "var(--ei-color-bg-card)",
          border: "1px solid var(--ei-color-rule-strong)",
          borderRadius: "var(--ei-radius-sm)",
          padding: "20px 24px",
          marginBottom: 20,
        }}
      >
        <div
          className="ei-label"
          style={{
            color: "var(--ei-color-fg-tertiary)",
            marginBottom: 10,
          }}
        >
          {t("jdMatch.search.dataSourcesHeading")}
        </div>
        <div style={{ display: "flex", gap: 10, alignItems: "center" }}>
          <input
            data-testid="jdmatch-search-input"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            onKeyDown={(e) => {
              if (e.key === "Enter" && !searching && query.trim()) onRun();
            }}
            placeholder={t("jdMatch.search.inputPlaceholder")}
            style={{
              flex: 1,
              padding: "12px 14px",
              fontSize: 14,
              color: "var(--ei-color-fg-primary)",
              background: "var(--ei-color-bg)",
              border: "1px solid var(--ei-color-rule-strong)",
              borderRadius: "var(--ei-radius-sm)",
              fontFamily: "var(--ei-font-sans)",
              outline: "none",
              boxSizing: "border-box",
            }}
          />
          <button
            type="button"
            data-testid="jdmatch-search-run"
            disabled={searching || !query.trim()}
            onClick={onRun}
            style={{
              padding: "10px 18px",
              fontSize: 13,
              fontFamily: "var(--ei-font-sans)",
              background: "var(--ei-color-accent)",
              color: "var(--ei-color-on-accent)",
              border: "none",
              borderRadius: "var(--ei-radius-sm)",
              cursor: searching || !query.trim() ? "not-allowed" : "pointer",
              opacity: searching || !query.trim() ? 0.6 : 1,
            }}
          >
            {searching
              ? t("jdMatch.search.runButtonRunning")
              : t("jdMatch.search.runButton")}
          </button>
        </div>
        {/* Source chips */}
        <div
          data-testid="jdmatch-search-sources"
          style={{
            display: "flex",
            gap: 6,
            marginTop: 14,
            flexWrap: "wrap",
            alignItems: "center",
          }}
        >
          {SOURCE_KEYS.map((k) => (
            <span
              key={k}
              data-testid={`jdmatch-search-source-${k}`}
              style={{
                padding: "2px 8px",
                fontSize: 11,
                fontFamily: "var(--ei-font-mono)",
                background: "var(--ei-color-bg-soft)",
                border: "1px solid var(--ei-color-rule-strong)",
                borderRadius: "var(--ei-radius-pill)",
                color: "var(--ei-color-fg-secondary)",
              }}
            >
              {t(`jdMatch.search.dataSource${capitalize(k)}` as never)}
            </span>
          ))}
        </div>
      </div>

      {/* AGENT scanning panel — searching=true only */}
      {searching ? (
        <div
          data-testid="jdmatch-search-searching-panel"
          style={{
            background: "var(--ei-color-bg-card)",
            border: "1px solid var(--ei-color-accent)",
            borderLeft: "3px solid var(--ei-color-accent)",
            padding: "18px 22px",
            marginBottom: 20,
            borderRadius: "var(--ei-radius-sm)",
          }}
        >
          <div
            className="ei-label"
            style={{ color: "var(--ei-color-accent)", marginBottom: 10 }}
          >
            {t("jdMatch.search.searchingPanelLabel")}
          </div>
          <div
            style={{
              display: "flex",
              flexDirection: "column",
              gap: 6,
              fontSize: 12.5,
              fontFamily: "var(--ei-font-mono)",
              color: "var(--ei-color-fg-secondary)",
            }}
          >
            {[1, 2, 3, 4, 5].map((step) => (
              <div
                key={step}
                data-testid={`jdmatch-search-searching-step-${step}`}
                style={{ opacity: step <= 3 ? 1 : 0.4 }}
              >
                {t(`jdMatch.search.searchingStep${step}` as never)}
              </div>
            ))}
          </div>
        </div>
      ) : null}

      {/* Saved searches */}
      <div style={{ marginBottom: 24 }}>
        <div
          style={{
            display: "flex",
            justifyContent: "space-between",
            alignItems: "center",
            marginBottom: 10,
          }}
        >
          <div
            className="ei-label"
            style={{ color: "var(--ei-color-fg-tertiary)" }}
          >
            {t("jdMatch.search.savedSearchesHeading")}
          </div>
          <button
            type="button"
            data-testid="jdmatch-search-save-current"
            onClick={onSaveCurrent}
            disabled={savedSearchCreating || !query.trim()}
            style={{
              background: "transparent",
              border: "none",
              color: "var(--ei-color-accent)",
              fontSize: 12.5,
              cursor:
                savedSearchCreating || !query.trim() ? "not-allowed" : "pointer",
              opacity: savedSearchCreating || !query.trim() ? 0.5 : 1,
            }}
          >
            +{" "}
            {t("jdMatch.search.savedSearchSaveCurrent")}
          </button>
        </div>
        {savedSearchesError ? (
          <div
            data-testid="jdmatch-search-saved-error"
            style={{
              padding: "12px 16px",
              background: "var(--ei-color-bg-soft)",
              border: "1px dashed var(--ei-color-rule-strong)",
              borderRadius: "var(--ei-radius-sm)",
              color: "var(--ei-color-warn)",
              fontSize: 12.5,
            }}
          >
            {t("jdMatch.search.resultsError")}
          </div>
        ) : (
          <div
            data-testid="jdmatch-search-saved-grid"
            style={{
              display: "grid",
              gridTemplateColumns: "repeat(3, 1fr)",
              gap: 10,
            }}
          >
            {savedSearches.map((s) => (
              <div
                key={s.id}
                data-testid={`jdmatch-search-saved-item-${s.id}`}
                style={{
                  padding: "14px 16px",
                  background: "var(--ei-color-bg-card)",
                  border: "1px solid var(--ei-color-rule-strong)",
                  borderRadius: "var(--ei-radius-sm)",
                }}
              >
                <div
                  style={{
                    display: "flex",
                    justifyContent: "space-between",
                    alignItems: "flex-start",
                    gap: 8,
                    marginBottom: 8,
                  }}
                >
                  <div
                    style={{
                      fontSize: 13.5,
                      color: "var(--ei-color-fg-primary)",
                      fontWeight: 500,
                      lineHeight: 1.3,
                    }}
                  >
                    {s.label}
                  </div>
                  {s.newJobsCount && s.newJobsCount > 0 ? (
                    <div
                      style={{
                        padding: "2px 7px",
                        background: "var(--ei-color-accent-soft)",
                        color: "var(--ei-color-accent)",
                        fontSize: 10.5,
                        fontFamily: "var(--ei-font-mono)",
                        borderRadius: "var(--ei-radius-sm)",
                      }}
                    >
                      +{s.newJobsCount}
                    </div>
                  ) : null}
                </div>
              </div>
            ))}
          </div>
        )}
        {savedSearchCreateError ? (
          <div
            data-testid="jdmatch-search-saved-create-error"
            style={{
              marginTop: 8,
              padding: "10px 14px",
              fontSize: 12,
              color: "var(--ei-color-warn)",
              background: "var(--ei-color-bg-soft)",
              borderRadius: "var(--ei-radius-sm)",
              display: "flex",
              gap: 10,
              alignItems: "center",
            }}
          >
            <span>{t("jdMatch.search.savedSearchCreateError")}</span>
            <button
              type="button"
              data-testid="jdmatch-search-saved-create-retry"
              onClick={onCreateSavedSearchRetry}
              style={{
                background: "transparent",
                border: "1px solid var(--ei-color-accent)",
                color: "var(--ei-color-accent)",
                fontSize: 11.5,
                padding: "3px 8px",
                borderRadius: "var(--ei-radius-sm)",
                cursor: "pointer",
              }}
            >
              {t("jdMatch.search.resultsErrorRetry")}
            </button>
          </div>
        ) : null}
      </div>

      {/* Result header + filter chips */}
      <div
        style={{
          display: "flex",
          justifyContent: "space-between",
          alignItems: "center",
          marginBottom: 12,
        }}
      >
        <div style={{ display: "flex", gap: 6 }}>
          {FILTER_KEYS.map((key) => {
            const active = resultFilter === key;
            return (
              <button
                key={key}
                type="button"
                data-testid={`jdmatch-search-filter-${key}`}
                data-active={active ? "true" : "false"}
                onClick={() => setResultFilter(key)}
                style={{
                  padding: "4px 10px",
                  fontSize: 11.5,
                  borderRadius: 12,
                  cursor: "pointer",
                  border: `1px solid ${
                    active ? "var(--ei-color-accent)" : "var(--ei-color-rule-strong)"
                  }`,
                  background: active
                    ? "var(--ei-color-accent-soft)"
                    : "transparent",
                  color: active
                    ? "var(--ei-color-accent)"
                    : "var(--ei-color-fg-tertiary)",
                  fontFamily: "var(--ei-font-sans)",
                }}
              >
                {t(`jdMatch.search.filter${capitalize(key)}` as never)}
              </button>
            );
          })}
        </div>
      </div>

      {/* Result body */}
      {error ? (
        <div
          data-testid="jdmatch-search-error"
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
          {t("jdMatch.search.resultsError")}
        </div>
      ) : cappedResults.length === 0 ? (
        hasRunOnce && !searching ? (
          <div
            data-testid="jdmatch-search-empty"
            style={{
              padding: "32px 20px",
              textAlign: "center",
              fontSize: 13,
              color: "var(--ei-color-fg-tertiary)",
              fontFamily: "var(--ei-font-mono)",
              background: "var(--ei-color-bg-soft)",
              border: "1px dashed var(--ei-color-rule-strong)",
              borderRadius: "var(--ei-radius-sm)",
            }}
          >
            {t("jdMatch.search.resultsEmpty")}
          </div>
        ) : null
      ) : (
        <div
          data-testid="jdmatch-search-results"
          style={{
            display: "grid",
            gridTemplateColumns: "1fr 1fr",
            gap: 10,
          }}
        >
          {cappedResults.map((rec) => (
            <JobMatchCard
              key={rec.id}
              recommendation={rec}
              active={false}
              onClick={() => onOpenJob(rec)}
            />
          ))}
        </div>
      )}
    </div>
  );
};

function capitalize(s: string): string {
  return s.charAt(0).toUpperCase() + s.slice(1);
}
