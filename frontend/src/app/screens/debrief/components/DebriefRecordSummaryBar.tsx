import type { FC } from "react";

import { useI18n } from "../../../i18n/messages";
import type { DebriefEntry } from "../types";

interface DebriefRecordSummaryBarProps {
  entries: DebriefEntry[];
}

/**
 * Source mirror of ui-design/src/screens-p1-depth.jsx lines 182-232. Counts
 * entries by source and renders chips (recorded / text / voice / manual)
 * plus the "shared list" tagline so users know switching mode does not
 * drop data.
 */
export const DebriefRecordSummaryBar: FC<DebriefRecordSummaryBarProps> = ({
  entries,
}) => {
  const { t } = useI18n();
  const counts = entries.reduce(
    (acc, entry) => {
      if (entry.source === "ai_confirmed" || entry.source === "ai_edited")
        acc.text += 1;
      else if (entry.source === "manual") acc.manual += 1;
      else if (entry.source === "voice_extracted") acc.voice += 1;
      return acc;
    },
    { text: 0, voice: 0, manual: 0 },
  );
  const chips: { key: string; labelKey: Parameters<typeof t>[0]; count: number }[] = [];
  if (counts.text > 0)
    chips.push({
      key: "text",
      labelKey: "debrief.record.summary.chipText",
      count: counts.text,
    });
  if (counts.voice > 0)
    chips.push({
      key: "voice",
      labelKey: "debrief.record.summary.chipVoice",
      count: counts.voice,
    });
  if (counts.manual > 0)
    chips.push({
      key: "manual",
      labelKey: "debrief.record.summary.chipManual",
      count: counts.manual,
    });
  return (
    <section
      className="ei-debrief-record-summary"
      data-testid="debrief-record-summary"
    >
      <div className="ei-debrief-record-summary__left">
        <span className="ei-label">{t("debrief.record.summary.eyebrow")}</span>
        <span
          className="ei-debrief-record-summary__count"
          data-testid="debrief-record-summary-count"
        >
          {entries.length}
        </span>
        <span className="ei-debrief-record-summary__unit">
          {t("debrief.record.summary.unit")}
        </span>
        {chips.length > 0 && (
          <ul
            className="ei-debrief-record-summary__chips"
            data-testid="debrief-record-summary-chips"
          >
            {chips.map((chip) => (
              <li key={chip.key} data-testid={`debrief-chip-${chip.key}`}>
                {t(chip.labelKey)} · {chip.count}
              </li>
            ))}
          </ul>
        )}
      </div>
      <div className="ei-debrief-record-summary__right">
        {t("debrief.record.summary.shared")}
      </div>
    </section>
  );
};
