import type { FC } from "react";

import type { FeedbackReport } from "../../../../../api/generated/types";
import { useI18n } from "../../../../i18n/messages";

interface NextTabProps {
  report: FeedbackReport;
  onReplay: () => void;
  onNextRound: () => void;
}

/**
 * Source-level mirror of ui-design/src/screen-report.jsx::ReportDetailSurface
 * next branch (lines 470-514). Two-column path A / path B layout with their
 * own action CTAs.
 */
export const NextTab: FC<NextTabProps> = ({ report, onReplay, onNextRound }) => {
  const { t } = useI18n();
  const actions = report.nextActions ?? [];
  return (
    <div
      data-testid="report-next-panel"
      style={{
        padding: 24,
        display: "grid",
        gridTemplateColumns: "1.15fr .85fr",
        gap: 18,
        alignItems: "start",
      }}
    >
      <div data-testid="report-next-path-a">
        <div
          className="ei-label"
          data-testid="report-next-desc-a"
          style={{ color: "var(--ei-accent)", marginBottom: 12 }}
        >
          {t("report.next.pathA.eyebrow")}
        </div>
        <div
          style={{
            fontSize: 13,
            color: "var(--ei-ink3)",
            lineHeight: 1.65,
            marginBottom: 12,
          }}
        >
          {t("report.next.pathA.body")}
        </div>
        <ul
          data-testid="report-next-actions-list"
          style={{ listStyle: "none", padding: 0, margin: 0 }}
        >
          {actions.length === 0 ? (
            <li
              data-testid="report-next-actions-empty"
              style={{ color: "var(--ei-ink3)" }}
            >
              {t("report.next.empty")}
            </li>
          ) : (
            actions.map((action, idx) => (
              <li
                key={idx}
                data-testid={`report-next-action-${idx}`}
                style={{
                  display: "grid",
                  gridTemplateColumns: "30px 1fr auto",
                  gap: 12,
                  alignItems: "center",
                  padding: "13px 0",
                  borderBottom:
                    idx < actions.length - 1
                      ? "1px dotted var(--ei-rule)"
                      : "none",
                }}
              >
                <span
                  style={{
                    width: 24,
                    height: 24,
                    borderRadius: 12,
                    background:
                      idx === 0 ? "var(--ei-accent)" : "var(--ei-bg-soft, var(--ei-bg))",
                    color: idx === 0 ? "#fff" : "var(--ei-ink3)",
                    display: "flex",
                    alignItems: "center",
                    justifyContent: "center",
                    fontSize: 11,
                    fontFamily: "var(--ei-mono)",
                  }}
                >
                  {idx + 1}
                </span>
                <span style={{ fontSize: 14, color: "var(--ei-ink2, var(--ei-ink))" }}>
                  {action.label}
                </span>
                <span
                  style={{
                    fontSize: 11,
                    fontFamily: "var(--ei-mono)",
                    color: "var(--ei-ink3)",
                  }}
                >
                  {action.type}
                </span>
              </li>
            ))
          )}
        </ul>
        <div style={{ marginTop: 16 }}>
          <button
            type="button"
            data-testid="report-next-cta-a"
            onClick={onReplay}
            style={{
              padding: "10px 16px",
              background: "var(--ei-accent)",
              color: "#fff",
              border: "1px solid var(--ei-accent)",
              borderRadius: 2,
              cursor: "pointer",
              fontFamily: "var(--ei-sans)",
              fontSize: 13,
            }}
          >
            {t("report.next.pathA.cta")}
          </button>
        </div>
      </div>
      <div
        data-testid="report-next-path-b"
        style={{
          padding: 18,
          background: "var(--ei-bg-soft, var(--ei-bg))",
          border: "1px solid var(--ei-rule)",
          borderRadius: 2,
        }}
      >
        <div
          className="ei-label"
          style={{ color: "var(--ei-ink3)", marginBottom: 8 }}
        >
          {t("report.next.pathB.eyebrow")}
        </div>
        <div
          className="ei-serif"
          style={{
            fontSize: 20,
            color: "var(--ei-ink)",
            lineHeight: 1.35,
            marginBottom: 10,
          }}
        >
          {t("report.next.pathB.title")}
        </div>
        <div
          data-testid="report-next-desc-b"
          style={{
            fontSize: 13,
            color: "var(--ei-ink2, var(--ei-ink))",
            lineHeight: 1.65,
            marginBottom: 14,
          }}
        >
          {t("report.next.pathB.body")}
        </div>
        <button
          type="button"
          data-testid="report-next-cta-b"
          onClick={onNextRound}
          style={{
            padding: "10px 16px",
            background: "transparent",
            color: "var(--ei-ink2, var(--ei-ink))",
            border: "1px solid var(--ei-rule)",
            borderRadius: 2,
            cursor: "pointer",
            fontFamily: "var(--ei-sans)",
            fontSize: 13,
          }}
        >
          {t("report.next.pathB.cta")}
        </button>
      </div>
    </div>
  );
};
