import { type FC } from "react";

import { useI18n } from "../../../i18n/messages";
import { useModalA11y } from "./useModalA11y";

interface ResumePickerModalProps {
  open: boolean;
  onClose: () => void;
  /** Currently bound resume id (if any) */
  boundResumeId?: string;
  /** Callback when user confirms the same (only) resume */
  onUseResume?: () => void;
}

/**
 * Phase 3: ResumePickerModal in disabled-list mode.
 * Only the currently bound resume is enabled + selected;
 * remaining slots render disabled placeholder cards.
 * listResumes is NOT called (missing OpenAPI operation).
 */
export const ResumePickerModal: FC<ResumePickerModalProps> = ({
  open,
  onClose,
  boundResumeId,
  onUseResume,
}) => {
  const { t } = useI18n();
  const { modalRef, handleKeyDown, handleBackdropClick } = useModalA11y({
    open,
    onClose,
  });

  if (!open) return null;

  return (
    <div
      data-testid="workspace-resume-modal-overlay"
      role="presentation"
      style={{
        position: "fixed",
        inset: 0,
        zIndex: 1000,
        background: "rgba(0,0,0,0.4)",
        display: "flex",
        alignItems: "center",
        justifyContent: "center",
      }}
      onClick={handleBackdropClick}
    >
      <div
        ref={modalRef}
        data-testid="workspace-resume-modal-card"
        role="dialog"
        aria-modal="true"
        aria-label={t("workspace.resumePicker.title")}
        style={{
          background: "var(--ei-color-bgCard)",
          border: "1px solid var(--ei-color-rule)",
          borderRadius: 4,
          width: 560,
          maxWidth: "calc(100vw - 48px)",
          maxHeight: "calc(100vh - 48px)",
          overflow: "auto",
          display: "flex",
          flexDirection: "column",
        }}
        onKeyDown={handleKeyDown}
      >
        <div
          style={{
            padding: "20px 24px 16px",
            borderBottom: "1px solid var(--ei-color-rule)",
            display: "flex",
            justifyContent: "space-between",
            alignItems: "flex-start",
          }}
        >
          <div>
            <div className="ei-serif" style={{ fontSize: 18, color: "var(--ei-color-ink)" }}>
              {t("workspace.resumePicker.title")}
            </div>
            <div style={{ fontSize: 13, color: "var(--ei-color-ink3)", marginTop: 6 }}>
              {t("workspace.resumePicker.sub")}
            </div>
          </div>
          <button
            data-testid="workspace-resume-modal-close"
            onClick={onClose}
            aria-label="Close"
            style={{
              background: "transparent",
              border: "none",
              fontSize: 18,
              cursor: "pointer",
              color: "var(--ei-color-ink3)",
              padding: "2px 6px",
            }}
          >
            ×
          </button>
        </div>

        <div style={{ padding: 20, flex: 1 }}>
          {/* Current bound resume card */}
          <div
            data-testid={`workspace-resume-modal-card-${boundResumeId ?? "none"}`}
            style={{
              padding: 14,
              border: "1px solid var(--ei-color-rule)",
              borderRadius: 3,
              marginBottom: 10,
              background: "var(--ei-color-bgSoft)",
            }}
          >
            <div style={{ fontSize: 13, color: "var(--ei-color-ink)" }}>
              {boundResumeId
                ? `${t("workspace.resumePicker.selected")} ${boundResumeId}`
                : t("workspace.resumePicker.selected")}
            </div>
          </div>

          {/* Disabled placeholder cards */}
          <div
            data-testid="workspace-resume-modal-disabled-note"
            style={{
              padding: 14,
              border: "1px dashed var(--ei-color-rule)",
              borderRadius: 3,
              fontSize: 12.5,
              color: "var(--ei-color-ink3)",
              textAlign: "center",
            }}
          >
            {t("workspace.resumePicker.disabledNote")}
          </div>
        </div>

        <div
          style={{
            padding: "16px 24px",
            borderTop: "1px solid var(--ei-color-rule)",
            display: "flex",
            justifyContent: "flex-end",
            gap: 10,
          }}
        >
          <button
            data-testid="workspace-resume-modal-cancel"
            onClick={onClose}
            style={{
              height: 32,
              padding: "0 14px",
              fontSize: 13,
              background: "transparent",
              color: "var(--ei-color-ink2)",
              border: "1px solid var(--ei-color-rule)",
              borderRadius: 2,
              cursor: "pointer",
            }}
          >
            {t("workspace.resumePicker.cancel")}
          </button>
          <button
            data-testid="workspace-resume-modal-confirm"
            onClick={() => {
              onUseResume?.();
              onClose();
            }}
            style={{
              height: 32,
              padding: "0 14px",
              fontSize: 13,
              background: "var(--ei-color-accent)",
              color: "#fff",
              border: "1px solid var(--ei-color-accent)",
              borderRadius: 2,
              cursor: "pointer",
            }}
          >
            {t("workspace.resumePicker.use")}
          </button>
        </div>
      </div>
    </div>
  );
};
