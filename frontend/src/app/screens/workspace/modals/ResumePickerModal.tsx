import { useEffect, useMemo, useState, type FC } from "react";

import { useI18n } from "../../../i18n/messages";
import { mapResumeToUiSource } from "../../resume-workshop/adapters/resume";
import { useResumeAssets } from "../../resume-workshop/hooks/useResumeAssets";
import { useModalA11y } from "./useModalA11y";

interface ResumePickerModalProps {
  open: boolean;
  onClose: () => void;
  /** Currently bound resume id (if any) */
  boundResumeId?: string;
  /** Callback when user confirms a selected resume */
  onSelectResume?: (resumeId: string) => void;
}

/**
 * Flat resume picker backed by the current listResumes contract.
 */
export const ResumePickerModal: FC<ResumePickerModalProps> = ({
  open,
  onClose,
  boundResumeId,
  onSelectResume,
}) => {
  const { t } = useI18n();
  const resumesQuery = useResumeAssets();
  const [selectedResumeId, setSelectedResumeId] = useState(boundResumeId ?? "");
  const { modalRef, handleKeyDown, handleBackdropClick } = useModalA11y({
    open,
    onClose,
  });

  const resumes = useMemo(
    () =>
      (resumesQuery.data?.items ?? [])
        .map(mapResumeToUiSource)
        .filter((resume) => resume.status === "active")
        .sort((a, b) => b.updatedAt.localeCompare(a.updatedAt)),
    [resumesQuery.data],
  );
  const activeResumeIds = useMemo(
    () => new Set(resumes.map((resume) => resume.id)),
    [resumes],
  );

  useEffect(() => {
    if (!open) return;
    if (resumesQuery.loading) {
      setSelectedResumeId(boundResumeId ?? "");
      return;
    }
    setSelectedResumeId((current) => {
      if (current && activeResumeIds.has(current)) return current;
      if (boundResumeId && activeResumeIds.has(boundResumeId)) {
        return boundResumeId;
      }
      return "";
    });
  }, [activeResumeIds, boundResumeId, open, resumesQuery.loading]);

  if (!open) return null;

  const selectedResumeIsActive = activeResumeIds.has(selectedResumeId);

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
          {resumesQuery.loading ? (
            <div
              data-testid="workspace-resume-modal-loading"
              role="status"
              style={{ fontSize: 13, color: "var(--ei-color-ink3)" }}
            >
              {t("workspace.resumePicker.loading")}
            </div>
          ) : resumesQuery.error ? (
            <div
              data-testid="workspace-resume-modal-error"
              role="alert"
              style={{ display: "grid", gap: 10 }}
            >
              <div style={{ fontSize: 13, color: "var(--ei-color-danger)" }}>
                {t("workspace.resumePicker.error")}
              </div>
              <button
                type="button"
                onClick={resumesQuery.retry}
                style={{
                  height: 32,
                  padding: "0 12px",
                  justifySelf: "start",
                  fontSize: 13,
                  background: "transparent",
                  color: "var(--ei-color-ink2)",
                  border: "1px solid var(--ei-color-rule)",
                  borderRadius: 2,
                  cursor: "pointer",
                }}
              >
                {t("workspace.errors.retry")}
              </button>
            </div>
          ) : resumes.length === 0 ? (
            <div
              data-testid="workspace-resume-modal-empty"
              style={{ fontSize: 13, color: "var(--ei-color-ink3)" }}
            >
              {t("workspace.resumePicker.empty")}
            </div>
          ) : (
            <div data-testid="workspace-resume-modal-options" style={{ display: "grid", gap: 10 }}>
              {resumes.map((resume) => {
                const selected = selectedResumeIsActive && selectedResumeId === resume.id;
                return (
                  <button
                    key={resume.id}
                    type="button"
                    data-testid={`workspace-resume-modal-option-${resume.id}`}
                    aria-pressed={selected}
                    onClick={() => setSelectedResumeId(resume.id)}
                    style={{
                      display: "grid",
                      gap: 4,
                      width: "100%",
                      padding: 14,
                      textAlign: "left",
                      border: selected
                        ? "1px solid var(--ei-color-accent)"
                        : "1px solid var(--ei-color-rule)",
                      borderRadius: 3,
                      background: selected
                        ? "var(--ei-color-bgSoft)"
                        : "var(--ei-color-bgCard)",
                      cursor: "pointer",
                    }}
                  >
                    <span style={{ fontSize: 13, color: "var(--ei-color-ink)" }}>
                      {resume.name}
                    </span>
                    <span style={{ fontSize: 12, color: "var(--ei-color-ink3)" }}>
                      {resume.sourceName} · {resume.langTag} · {resume.updatedAt}
                    </span>
                    {resume.summary ? (
                      <span style={{ fontSize: 12, color: "var(--ei-color-ink3)" }}>
                        {resume.summary}
                      </span>
                    ) : null}
                  </button>
                );
              })}
            </div>
          )}
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
            disabled={!selectedResumeIsActive}
            onClick={() => {
              if (!selectedResumeIsActive) return;
              onSelectResume?.(selectedResumeId);
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
              cursor: selectedResumeIsActive ? "pointer" : "not-allowed",
              opacity: selectedResumeIsActive ? 1 : 0.5,
            }}
          >
            {t("workspace.resumePicker.use")}
          </button>
        </div>
      </div>
    </div>
  );
};
