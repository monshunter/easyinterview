import type { FC } from "react";

import { useI18n, type MessageKey } from "../../../i18n/messages";

export interface ReportContextStripProps {
  sessionId: string;
  targetLabel: string | null;
  roundLabel: string | null;
  resumeLabel: string | null;
  modality: string;
  practiceMode: string;
  hintUsed: string;
  hintCount: string;
}

/**
 * Source-level mirror of ui-design/src/screen-report.jsx::ReportContextStrip
 * (lines 266-299). 7 owner / display-knob slots flow in through props so the
 * dashboard never touches raw resume / JD body fields; ContextStrip operates
 * on labels only.
 */
export const ReportContextStrip: FC<ReportContextStripProps> = (props) => {
  const { t, lang } = useI18n();
  const fields: Array<{
    testId: string;
    labelKey: MessageKey;
    value: string;
  }> = [
    {
      testId: "report-context-session",
      labelKey: "report.context.session",
      value: props.sessionId,
    },
    {
      testId: "report-context-job",
      labelKey: "report.context.job",
      value: props.targetLabel ?? "—",
    },
    {
      testId: "report-context-round",
      labelKey: "report.context.round",
      value: props.roundLabel ?? "—",
    },
    {
      testId: "report-context-resume",
      labelKey: "report.context.resume",
      value: props.resumeLabel ?? "—",
    },
    {
      testId: "report-context-modality",
      labelKey: "report.context.modality",
      value: t(
        props.modality === "voice"
          ? "report.context.modality.voice"
          : "report.context.modality.text",
      ),
    },
    {
      testId: "report-context-practice-mode",
      labelKey: "report.context.practiceMode",
      value: t(
        props.practiceMode === "assisted"
          ? "report.context.practiceMode.assisted"
          : "report.context.practiceMode.strict",
      ),
    },
    {
      testId: "report-context-hints",
      labelKey: "report.context.hints",
      value:
        props.hintUsed === "true"
          ? `${t("report.context.hints.used")} · ${props.hintCount}`
          : t("report.context.hints.none"),
    },
  ];
  return (
    <div
      data-testid="report-context-strip"
      data-lang={lang}
      style={{
        display: "grid",
        gridTemplateColumns: "repeat(3, 1fr)",
        border: "1px solid var(--ei-rule)",
        borderRadius: 3,
        background: "var(--ei-bg-card, var(--ei-bg))",
        marginBottom: 18,
        overflow: "hidden",
      }}
    >
      {fields.map((field, i) => (
        <div
          key={field.testId}
          data-testid={field.testId}
          style={{
            padding: "13px 16px",
            borderRight:
              (i + 1) % 3 === 0
                ? "none"
                : "1px dotted var(--ei-rule)",
            borderBottom: i < 3 ? "1px dotted var(--ei-rule)" : "none",
          }}
        >
          <div
            className="ei-label"
            style={{ color: "var(--ei-ink3)", marginBottom: 4 }}
          >
            {t(field.labelKey)}
          </div>
          <div
            style={{
              fontSize: 13.5,
              color: "var(--ei-ink)",
              fontWeight: 500,
              whiteSpace: "nowrap",
              overflow: "hidden",
              textOverflow: "ellipsis",
            }}
          >
            {field.value}
          </div>
        </div>
      ))}
    </div>
  );
};
