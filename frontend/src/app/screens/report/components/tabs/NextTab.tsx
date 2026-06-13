import type { FC } from "react";

import type { FeedbackReport } from "../../../../../api/generated/types";
import { useI18n } from "../../../../i18n/messages";

interface NextTabProps {
  report: FeedbackReport;
}

/**
 * Source-level mirror of ui-design/src/screen-report.jsx::ReportDetailSurface
 * next branch (lines 470-514). Two-column path A / path B layout carrying
 * path descriptions + the retry checklist only. Per product-scope v2.1 D-19
 * the start CTAs converge to the report Header; this tab no longer renders
 * `report-next-cta-a` / `report-next-cta-b`, pointing users at the Header
 * CTA through a footer line instead.
 */
export const NextTab: FC<NextTabProps> = ({ report }) => {
  const { t } = useI18n();
  const actions = report.nextActions ?? [];
  return (
    <div
      data-testid="report-next-panel"
      style={{
        padding: 24,
        display: "grid",
        gridTemplateColumns: "repeat(auto-fit, minmax(min(240px, 100%), 1fr))",
        gap: 18,
        alignItems: "start",
        minWidth: 0,
      }}
    >
      <div data-testid="report-next-path-a" style={{ minWidth: 0 }}>
        <div
          className="ei-label"
          data-testid="report-next-desc-a"
          style={{ color: "var(--ei-color-accent)", marginBottom: 12 }}
        >
          {t("report.next.pathA.eyebrow")}
        </div>
        <div
          style={{
            fontSize: 13,
            color: "var(--ei-color-fg-tertiary)",
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
              style={{ color: "var(--ei-color-fg-tertiary)" }}
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
                  gridTemplateColumns: "30px minmax(0, 1fr) auto",
                  gap: 12,
                  alignItems: "center",
                  padding: "13px 0",
                  borderBottom:
                    idx < actions.length - 1
                      ? "1px dotted var(--ei-color-rule-soft)"
                      : "none",
                }}
              >
                <span
                  style={{
                    width: 24,
                    height: 24,
                    borderRadius: 12,
                    background:
                      idx === 0 ? "var(--ei-color-accent)" : "var(--ei-color-bg-soft, var(--ei-color-bg-canvas))",
                    color: idx === 0 ? "#fff" : "var(--ei-color-fg-tertiary)",
                    display: "flex",
                    alignItems: "center",
                    justifyContent: "center",
                    fontSize: 11,
                    fontFamily: "var(--ei-font-mono)",
                  }}
                >
                  {idx + 1}
                </span>
                <span
                  style={{
                    fontSize: 14,
                    color: "var(--ei-color-fg-secondary, var(--ei-color-fg-primary))",
                    minWidth: 0,
                    overflowWrap: "anywhere",
                  }}
                >
                  {action.label}
                </span>
                <span
                  style={{
                    fontSize: 11,
                  fontFamily: "var(--ei-font-mono)",
                  color: "var(--ei-color-fg-tertiary)",
                  overflowWrap: "anywhere",
                }}
              >
                  {action.type}
                </span>
              </li>
            ))
          )}
        </ul>
        <div
          data-testid="report-next-path-a-footer"
          style={{
            marginTop: 16,
            fontSize: 12,
            color: "var(--ei-color-fg-tertiary)",
            fontFamily: "var(--ei-font-mono)",
          }}
        >
          {t("report.next.pathA.footer")}
        </div>
      </div>
      <div
        data-testid="report-next-path-b"
        style={{
          padding: 18,
          background: "var(--ei-color-bg-soft, var(--ei-color-bg-canvas))",
          border: "1px solid var(--ei-color-rule-soft)",
          borderRadius: 2,
          minWidth: 0,
        }}
      >
        <div
          className="ei-label"
          style={{ color: "var(--ei-color-fg-tertiary)", marginBottom: 8 }}
        >
          {t("report.next.pathB.eyebrow")}
        </div>
        <div
          className="ei-serif"
          style={{
            fontSize: 20,
            color: "var(--ei-color-fg-primary)",
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
            color: "var(--ei-color-fg-secondary, var(--ei-color-fg-primary))",
            lineHeight: 1.65,
            marginBottom: 14,
          }}
        >
          {t("report.next.pathB.body")}
        </div>
        <div
          data-testid="report-next-path-b-footer"
          style={{
            fontSize: 12,
            color: "var(--ei-color-fg-tertiary)",
            fontFamily: "var(--ei-font-mono)",
          }}
        >
          {t("report.next.pathB.footer")}
        </div>
      </div>
    </div>
  );
};
