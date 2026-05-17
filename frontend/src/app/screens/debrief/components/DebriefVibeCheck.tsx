import type { FC } from "react";

import { useI18n } from "../../../i18n/messages";

export const DebriefVibeCheck: FC = () => {
  const { t } = useI18n();
  return (
    <aside className="ei-debrief-vibe" data-testid="debrief-vibe-check">
      <div className="ei-label">{t("debrief.record.vibe.eyebrow")}</div>
      <div className="ei-debrief-vibe__group">
        <div className="ei-debrief-vibe__label">
          {t("debrief.record.vibe.overall")}
        </div>
        <div
          className="ei-debrief-vibe__mood"
          data-testid="debrief-vibe-mood"
          aria-label={t("debrief.record.vibe.moodAria")}
        >
          {["🙁", "😐", "🙂", "😊"].map((mood, index) => (
            <button
              key={mood}
              type="button"
              data-active={index === 2}
              aria-pressed={index === 2}
            >
              {mood}
            </button>
          ))}
        </div>
      </div>
      <label className="ei-debrief-vibe__group">
        <span className="ei-debrief-vibe__label">
          {t("debrief.record.vibe.liked")}
        </span>
        <textarea
          rows={2}
          defaultValue={t("debrief.record.vibe.likedDefault")}
        />
      </label>
      <label className="ei-debrief-vibe__group">
        <span className="ei-debrief-vibe__label">
          {t("debrief.record.vibe.stumbled")}
        </span>
        <textarea
          rows={2}
          defaultValue={t("debrief.record.vibe.stumbledDefault")}
        />
      </label>
    </aside>
  );
};
