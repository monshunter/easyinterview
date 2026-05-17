import type { FC } from "react";

import { useI18n } from "../../../i18n/messages";
import type { Debrief, DebriefEntry } from "../types";

interface DebriefReplayPlanProps {
  debrief: Debrief | null;
  entries: DebriefEntry[];
  onStart: () => void;
  onBack: () => void;
}

/**
 * Source mirror of ui-design/src/screens-p1-depth.jsx::DebriefReplayPlan
 * (lines 1388-1421). Step 2 launcher previewing the replay-interview plan
 * (real questions + weak-spot probes + ordered playback + resume evidence
 * comparison) and routing the "Start" CTA into the practice flow.
 */
export const DebriefReplayPlan: FC<DebriefReplayPlanProps> = ({
  debrief,
  entries,
  onStart,
  onBack,
}) => {
  const { t } = useI18n();
  const previewQuestions =
    debrief?.questions?.length && debrief.questions.length > 0
      ? debrief.questions.map((q) => q.questionText)
      : entries.map((e) => e.questionText);
  const riskItems = debrief?.riskItems ?? [];
  return (
    <section
      className="ei-debrief-replay"
      data-testid="debrief-replay-plan"
    >
      <div className="ei-label">{t("debrief.replay.eyebrow")}</div>
      <h2 className="ei-serif">{t("debrief.replay.title")}</h2>
      <p>{t("debrief.replay.body")}</p>
      <ul data-testid="debrief-replay-preview-questions">
        {previewQuestions.slice(0, 5).map((q, idx) => (
          <li key={`${idx}-${q}`}>{q}</li>
        ))}
      </ul>
      {riskItems.length > 0 && (
        <ul data-testid="debrief-replay-preview-risks">
          {riskItems.slice(0, 3).map((risk, idx) => (
            <li key={`${idx}-${risk.label}`}>{risk.label}</li>
          ))}
        </ul>
      )}
      <div className="ei-debrief-replay__actions">
        <button
          type="button"
          className="ei-debrief-btn ei-debrief-btn--ghost"
          data-testid="debrief-replay-back"
          onClick={onBack}
        >
          {t("debrief.replay.back")}
        </button>
        <button
          type="button"
          className="ei-debrief-btn ei-debrief-btn--accent"
          data-testid="debrief-start-interview-btn"
          onClick={onStart}
        >
          {t("debrief.replay.cta")}
        </button>
      </div>
    </section>
  );
};
