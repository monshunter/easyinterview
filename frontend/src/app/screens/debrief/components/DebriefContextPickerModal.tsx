import { useCallback, useEffect, useState, type FC, type ReactNode } from "react";

import { useI18n } from "../../../i18n/messages";
import type { MessageKey } from "../../../i18n/locales/zh";

export interface PickerOption<T> {
  id: string;
  title: string;
  meta?: string;
  note?: string;
  value: T;
}

interface DebriefContextPickerModalProps<T> {
  /** Picker kind drives copy + a11y label. */
  kind: "targetJob" | "mockSession" | "resume";
  options: PickerOption<T>[];
  selectedId: string | null;
  loading?: boolean;
  errorMessage?: string | null;
  /** Optional banner rendered between the body and the options list. */
  banner?: ReactNode;
  /** Optional empty-state copy when options is empty and not loading. */
  emptyCopy?: string;
  /** Allow confirming with no selection (e.g. mock session is optional). */
  allowEmpty?: boolean;
  /** Copy for the "no selection" sentinel option (mock session "暂不关联"). */
  noneOptionCopy?: string;
  onClose: () => void;
  onConfirm: (selected: PickerOption<T> | null) => void;
}

const META: Record<
  DebriefContextPickerModalProps<unknown>["kind"],
  {
    eyebrow: MessageKey;
    title: MessageKey;
    body: MessageKey;
    confirm: MessageKey;
  }
> = {
  targetJob: {
    eyebrow: "debrief.picker.targetJob.eyebrow",
    title: "debrief.picker.targetJob.title",
    body: "debrief.picker.targetJob.body",
    confirm: "debrief.picker.targetJob.confirm",
  },
  mockSession: {
    eyebrow: "debrief.picker.mockSession.eyebrow",
    title: "debrief.picker.mockSession.title",
    body: "debrief.picker.mockSession.body",
    confirm: "debrief.picker.mockSession.confirm",
  },
  resume: {
    eyebrow: "debrief.picker.resume.eyebrow",
    title: "debrief.picker.resume.title",
    body: "debrief.picker.resume.body",
    confirm: "debrief.picker.resume.confirm",
  },
};

const NONE_SENTINEL = "__debrief_picker_none__";

/**
 * Source mirror of ui-design/src/screens-p1-depth.jsx::DebriefContextPickerModal
 * (lines 434-518). In-page modal that never leaves the debrief route. Esc /
 * outside-click both close. Mobile viewports get a full-bleed sheet via the
 * `data-variant="sheet"` attribute (driven from debrief.css media query).
 */
export function DebriefContextPickerModal<T>({
  kind,
  options,
  selectedId,
  loading,
  errorMessage,
  banner,
  emptyCopy,
  allowEmpty,
  noneOptionCopy,
  onClose,
  onConfirm,
}: DebriefContextPickerModalProps<T>) {
  const { t } = useI18n();
  const meta = META[kind];

  const initialDraft = selectedId ?? (allowEmpty ? NONE_SENTINEL : null);
  const [draftId, setDraftId] = useState<string | null>(initialDraft);

  useEffect(() => {
    const handler = (event: KeyboardEvent) => {
      if (event.key === "Escape") onClose();
    };
    document.addEventListener("keydown", handler);
    return () => document.removeEventListener("keydown", handler);
  }, [onClose]);

  const confirm = useCallback(() => {
    if (draftId === NONE_SENTINEL) {
      onConfirm(null);
      return;
    }
    const picked = options.find((opt) => opt.id === draftId) ?? null;
    if (!picked && !allowEmpty) return;
    onConfirm(picked);
  }, [allowEmpty, draftId, onConfirm, options]);

  const showEmptyState = !loading && !errorMessage && options.length === 0;

  return (
    <div
      role="presentation"
      className="ei-debrief-picker-modal__scrim"
      data-testid="debrief-picker-modal"
      data-kind={kind}
      onClick={onClose}
    >
      <div
        role="dialog"
        aria-modal="true"
        aria-labelledby={`debrief-picker-title-${kind}`}
        className="ei-debrief-picker-modal"
        onClick={(e) => e.stopPropagation()}
      >
        <header className="ei-debrief-picker-modal__head">
          <div>
            <div className="ei-label ei-debrief-picker-modal__eyebrow">
              {t(meta.eyebrow)}
            </div>
            <h2
              id={`debrief-picker-title-${kind}`}
              className="ei-serif ei-debrief-picker-modal__title"
            >
              {t(meta.title)}
            </h2>
            <p className="ei-debrief-picker-modal__body">{t(meta.body)}</p>
          </div>
          <button
            type="button"
            className="ei-debrief-picker-modal__close"
            data-testid="debrief-picker-close"
            onClick={onClose}
            aria-label={t("debrief.picker.close")}
          >
            ×
          </button>
        </header>

        {banner ? (
          <div className="ei-debrief-picker-modal__banner" data-testid="debrief-picker-banner">
            {banner}
          </div>
        ) : null}

        {loading ? (
          <div className="ei-debrief-picker-modal__loading" data-testid="debrief-picker-loading">
            {t("debrief.picker.loading")}
          </div>
        ) : errorMessage ? (
          <div className="ei-debrief-picker-modal__error" data-testid="debrief-picker-error">
            {errorMessage}
          </div>
        ) : showEmptyState ? (
          <div className="ei-debrief-picker-modal__empty" data-testid="debrief-picker-empty">
            {emptyCopy ?? t("debrief.picker.empty")}
          </div>
        ) : (
          <div className="ei-debrief-picker-modal__options" role="radiogroup">
            {allowEmpty && (
              <button
                key={NONE_SENTINEL}
                type="button"
                role="radio"
                aria-checked={draftId === NONE_SENTINEL}
                className="ei-debrief-picker-modal__option"
                data-active={draftId === NONE_SENTINEL}
                data-testid="debrief-picker-option-none"
                onClick={() => setDraftId(NONE_SENTINEL)}
              >
                <span className="ei-debrief-picker-modal__option-mark">
                  {draftId === NONE_SENTINEL ? "●" : "○"}
                </span>
                <span>
                  <span className="ei-debrief-picker-modal__option-title">
                    {noneOptionCopy ?? t("debrief.picker.mockSession.none")}
                  </span>
                </span>
              </button>
            )}
            {options.map((opt) => {
              const active = opt.id === draftId;
              return (
                <button
                  key={opt.id}
                  type="button"
                  role="radio"
                  aria-checked={active}
                  className="ei-debrief-picker-modal__option"
                  data-active={active}
                  data-testid={`debrief-picker-option-${opt.id}`}
                  onClick={() => setDraftId(opt.id)}
                >
                  <span className="ei-debrief-picker-modal__option-mark">
                    {active ? "●" : "○"}
                  </span>
                  <span>
                    <span className="ei-debrief-picker-modal__option-title">
                      {opt.title}
                    </span>
                    {opt.meta ? (
                      <span className="ei-debrief-picker-modal__option-meta">
                        {opt.meta}
                      </span>
                    ) : null}
                    {opt.note ? (
                      <span className="ei-debrief-picker-modal__option-note">
                        {opt.note}
                      </span>
                    ) : null}
                  </span>
                </button>
              );
            })}
          </div>
        )}

        <footer className="ei-debrief-picker-modal__foot">
          <button
            type="button"
            className="ei-debrief-picker-modal__btn ei-debrief-picker-modal__btn--ghost"
            data-testid="debrief-picker-cancel"
            onClick={onClose}
          >
            {t("debrief.picker.cancel")}
          </button>
          <button
            type="button"
            className="ei-debrief-picker-modal__btn ei-debrief-picker-modal__btn--accent"
            data-testid="debrief-picker-confirm"
            disabled={
              loading ||
              (!allowEmpty &&
                (draftId === null || draftId === NONE_SENTINEL))
            }
            onClick={confirm}
          >
            {t(meta.confirm)}
          </button>
        </footer>
      </div>
    </div>
  );
}
