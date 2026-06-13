import type { FC } from "react";

import { useI18n } from "../../../i18n/messages";
import type { MessageKey } from "../../../i18n/locales/zh";
import type { DebriefPickerKind, DebriefSelectedContext } from "../types";

interface DebriefContextStripProps {
  selectedContext: DebriefSelectedContext;
  onOpenPicker: (kind: DebriefPickerKind) => void;
}

interface StripCardConfig {
  kind: DebriefPickerKind;
  labelKey: MessageKey;
  metaKey: MessageKey;
  actionKey: MessageKey;
  icon: string;
}

const CARDS: StripCardConfig[] = [
  {
    kind: "targetJob",
    labelKey: "debrief.contextStrip.targetJobLabel",
    metaKey: "debrief.contextStrip.targetJobMeta",
    actionKey: "debrief.contextStrip.actionChange",
    icon: "💼",
  },
  {
    kind: "mockSession",
    labelKey: "debrief.contextStrip.mockSessionLabel",
    metaKey: "debrief.contextStrip.mockSessionMeta",
    actionKey: "debrief.contextStrip.actionSelect",
    icon: "📊",
  },
  {
    kind: "resume",
    labelKey: "debrief.contextStrip.resumeLabel",
    metaKey: "debrief.contextStrip.resumeMeta",
    actionKey: "debrief.contextStrip.actionChange",
    icon: "📄",
  },
];

function pickTitle(
  ctx: DebriefSelectedContext,
  kind: DebriefPickerKind,
  unsetCopy: string,
): string {
  switch (kind) {
    case "targetJob": {
      const tj = ctx.targetJob;
      if (!tj) return unsetCopy;
      const company = tj.companyName ?? "";
      const title = tj.title ?? "";
      const composed = [company, title].filter(Boolean).join(" · ");
      return composed || tj.id;
    }
    case "mockSession": {
      const ms = ctx.mockSession;
      if (!ms) return unsetCopy;
      return ms.id;
    }
    case "resume": {
      const resume = ctx.resume;
      if (!resume) return unsetCopy;
      return resume.displayName || resume.title || resume.id;
    }
  }
}

/**
 * Source mirror of ui-design/src/screens-p1-depth.jsx::DebriefContextStrip
 * (lines 412-432). Three cards (target job, mock session, resume) — each
 * shows the currently-selected display value or the "未选择" fallback and
 * surfaces a button that opens the in-page picker modal.
 *
 * The strip itself is presentational: option fetching, modal mounting, and
 * `SET_DEBRIEF_CONTEXT` reducer writes happen one level up in
 * `<DebriefScreen>` (Phase 1.1) and the picker modals (Phase 2).
 */
export const DebriefContextStrip: FC<DebriefContextStripProps> = ({
  selectedContext,
  onOpenPicker,
}) => {
  const { t } = useI18n();
  const unsetCopy = t("debrief.contextStrip.unset");

  return (
    <section
      className="ei-debrief-context-strip"
      data-testid="debrief-context-strip"
    >
      {CARDS.map((card) => {
        const title = pickTitle(selectedContext, card.kind, unsetCopy);
        return (
          <article
            key={card.kind}
            className="ei-debrief-context-strip__card"
            data-testid={`debrief-context-card-${card.kind}`}
          >
            <span
              className="ei-debrief-context-strip__icon"
              aria-hidden="true"
            >
              {card.icon}
            </span>
            <div className="ei-debrief-context-strip__body">
              <div
                className="ei-label ei-debrief-context-strip__label"
                data-testid={`debrief-context-card-${card.kind}-label`}
              >
                {t(card.labelKey)}
              </div>
              <div
                className="ei-debrief-context-strip__title"
                data-testid={`debrief-context-card-${card.kind}-title`}
              >
                {title}
              </div>
              <div
                className="ei-debrief-context-strip__meta"
                data-testid={`debrief-context-card-${card.kind}-meta`}
              >
                {t(card.metaKey)}
              </div>
            </div>
            <button
              type="button"
              className="ei-debrief-context-strip__action"
              data-testid={`debrief-context-card-${card.kind}-open`}
              onClick={() => onOpenPicker(card.kind)}
            >
              {t(card.actionKey)}
            </button>
          </article>
        );
      })}
    </section>
  );
};
