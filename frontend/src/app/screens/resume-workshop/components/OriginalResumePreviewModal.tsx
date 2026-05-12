import { useEffect, useRef, type FC, type KeyboardEvent } from "react";

import { useI18n } from "../../../i18n/messages";

export interface OriginalResumePreviewModalProps {
  open: boolean;
  onClose: () => void;
  /** Plain-text snippet of the original resume body (passed in from preview). */
  originalText: string[];
  /** Title shown in the modal header. */
  title: string;
  /** Loading/error state for the source asset backing the modal body. */
  contentState?: "ready" | "loading" | "error";
  onRetry?: () => void;
}

const FOCUSABLE_SELECTOR =
  "button:not([disabled]), [href], input:not([disabled]), select:not([disabled]), textarea:not([disabled]), [tabindex]:not([tabindex='-1'])";

export const OriginalResumePreviewModal: FC<OriginalResumePreviewModalProps> = ({
  open,
  onClose,
  originalText,
  title,
  contentState = "ready",
  onRetry,
}) => {
  const { t } = useI18n();
  const dialogRef = useRef<HTMLDivElement | null>(null);
  const closeButtonRef = useRef<HTMLButtonElement | null>(null);
  const previouslyFocusedRef = useRef<HTMLElement | null>(null);

  useEffect(() => {
    if (!open) return;
    previouslyFocusedRef.current = document.activeElement as HTMLElement | null;
    closeButtonRef.current?.focus();
    return () => {
      previouslyFocusedRef.current?.focus?.();
    };
  }, [open]);

  if (!open) return null;

  const onKeyDown = (event: KeyboardEvent<HTMLDivElement>) => {
    if (event.key === "Escape") {
      event.preventDefault();
      onClose();
      return;
    }
    if (event.key !== "Tab" || !dialogRef.current) return;
    const focusable = Array.from(
      dialogRef.current.querySelectorAll<HTMLElement>(FOCUSABLE_SELECTOR),
    ).filter((node) => !node.hasAttribute("data-focus-skip"));
    if (focusable.length === 0) return;
    const first = focusable[0]!;
    const last = focusable[focusable.length - 1]!;
    const active = document.activeElement as HTMLElement | null;
    if (event.shiftKey && active === first) {
      event.preventDefault();
      last.focus();
    } else if (!event.shiftKey && active === last) {
      event.preventDefault();
      first.focus();
    }
  };

  return (
    <div
      data-testid="resume-detail-original-modal-overlay"
      className="ei-resume-detail-modal-overlay"
      onClick={onClose}
    >
      <div
        ref={dialogRef}
        data-testid="resume-detail-original-modal"
        className="ei-resume-detail-modal"
        role="dialog"
        aria-modal="true"
        aria-labelledby="resume-detail-original-modal-title"
        aria-describedby="resume-detail-original-modal-desc"
        tabIndex={-1}
        onClick={(event) => event.stopPropagation()}
        onKeyDown={onKeyDown}
      >
        <header className="ei-resume-detail-modal-header">
          <h2
            id="resume-detail-original-modal-title"
            className="ei-text-title"
          >
            {t("resumeWorkshop.detail.modalTitle")}
          </h2>
          <button
            ref={closeButtonRef}
            type="button"
            data-testid="resume-detail-original-modal-close"
            aria-label={t("resumeWorkshop.detail.modalClose")}
            onClick={onClose}
          >
            ×
          </button>
        </header>
        <div
          id="resume-detail-original-modal-desc"
          className="ei-resume-detail-modal-desc"
        >
          <p className="ei-text-body">{title}</p>
          <p className="ei-text-body">
            {t("resumeWorkshop.detail.modalDescription")}
          </p>
        </div>
        <article
          data-testid="resume-detail-original-modal-content"
          className="ei-resume-detail-modal-content"
        >
          {contentState === "loading" ? (
            <p
              className="ei-text-body"
              data-testid="resume-detail-original-modal-loading"
              role="status"
            >
              {t("resumeWorkshop.detail.modalLoading")}
            </p>
          ) : contentState === "error" ? (
            <div
              data-testid="resume-detail-original-modal-error"
              role="alert"
            >
              <p className="ei-text-body">
                {t("resumeWorkshop.detail.modalError")}
              </p>
              {onRetry ? (
                <button
                  type="button"
                  className="ei-cta"
                  data-testid="resume-detail-original-modal-retry"
                  onClick={onRetry}
                >
                  {t("workspace.errors.retry")}
                </button>
              ) : null}
            </div>
          ) : originalText.length === 0 ? (
            <p className="ei-text-body">—</p>
          ) : (
            originalText.map((line, index) => (
              <p key={index} className="ei-text-body">
                {line}
              </p>
            ))
          )}
        </article>
      </div>
    </div>
  );
};
