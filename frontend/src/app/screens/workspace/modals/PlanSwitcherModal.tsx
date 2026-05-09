import { useEffect, useState, type FC } from "react";

import { useI18n } from "../../../i18n/messages";
import { useNavigation } from "../../../navigation/NavigationProvider";
import { useWorkspaceTargetJobs } from "../hooks/useWorkspaceTargetJobs";
import { useModalA11y } from "./useModalA11y";

interface PlanSwitcherModalProps {
  open: boolean;
  onClose: () => void;
  onSelectPlan?: (planId: string) => void;
}

/**
 * Phase 3: PlanSwitcherModal — lists candidate plans via listTargetJobs.
 * "Create Plan from New JD" navigates to home.
 * "Use This Plan" updates InterviewContext and triggers refetch.
 */
export const PlanSwitcherModal: FC<PlanSwitcherModalProps> = ({
  open,
  onClose,
  onSelectPlan,
}) => {
  const { t } = useI18n();
  const { navigate } = useNavigation();
  const { loading, jobs } = useWorkspaceTargetJobs();
  const [draftJobId, setDraftJobId] = useState<string | null>(null);
  const { modalRef, handleKeyDown, handleBackdropClick } = useModalA11y({
    open,
    onClose,
  });

  useEffect(() => {
    if (!open) {
      setDraftJobId(null);
    }
  }, [open]);

  if (!open) return null;

  return (
    <div
      data-testid="workspace-plan-modal-overlay"
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
        data-testid="workspace-plan-modal-card"
        role="dialog"
        aria-modal="true"
        aria-label={t("workspace.planSwitcher.title")}
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
              {t("workspace.planSwitcher.title")}
            </div>
            <div style={{ fontSize: 13, color: "var(--ei-color-ink3)", marginTop: 6 }}>
              {t("workspace.planSwitcher.sub")}
            </div>
          </div>
          <button
            data-testid="workspace-plan-modal-close"
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
          {loading ? (
            <div style={{ fontSize: 13, color: "var(--ei-color-ink3)" }}>
              {t("workspace.placeholder")}
            </div>
          ) : (
            jobs.map((job) => {
              const selected = draftJobId === job.id;
              return (
                <button
                  key={job.id}
                  type="button"
                  data-testid={`workspace-plan-modal-card-${job.id}`}
                  style={{
                    display: "block",
                    width: "100%",
                    padding: 14,
                    border: selected
                      ? "1px solid var(--ei-color-accent)"
                      : "1px solid var(--ei-color-rule)",
                    borderRadius: 3,
                    marginBottom: 10,
                    background: selected
                      ? "var(--ei-color-accentSoft)"
                      : "var(--ei-color-bgSoft)",
                    cursor: "pointer",
                    textAlign: "left",
                    fontFamily: "var(--ei-sans)",
                  }}
                  onClick={() => setDraftJobId(job.id)}
                >
                  <div style={{ fontSize: 14, fontWeight: 500, color: "var(--ei-color-ink)" }}>
                    {job.title}
                  </div>
                  <div style={{ fontSize: 12, color: "var(--ei-color-ink3)", marginTop: 4 }}>
                    {job.companyName} · {job.locationText ?? ""}
                  </div>
                </button>
              );
            })
          )}

          <div
            data-testid="workspace-plan-modal-create"
            style={{
              padding: 14,
              border: "1px dashed var(--ei-color-rule)",
              borderRadius: 3,
              fontSize: 13,
              color: "var(--ei-color-ink2)",
              textAlign: "center",
              cursor: "pointer",
            }}
            onClick={() => navigate({ name: "home", params: {} })}
          >
            {t("workspace.planSwitcher.create")}
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
            data-testid="workspace-plan-modal-cancel"
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
            {t("workspace.planSwitcher.cancel")}
          </button>
          <button
            data-testid="workspace-plan-modal-confirm"
            disabled={!draftJobId}
            onClick={() => {
              if (!draftJobId) return;
              onSelectPlan?.(draftJobId);
              onClose();
            }}
            style={{
              height: 32,
              padding: "0 14px",
              fontSize: 13,
              background: draftJobId
                ? "var(--ei-color-accent)"
                : "var(--ei-color-ink4)",
              color: "#fff",
              border: draftJobId
                ? "1px solid var(--ei-color-accent)"
                : "1px solid var(--ei-color-ink4)",
              borderRadius: 2,
              cursor: draftJobId ? "pointer" : "default",
            }}
          >
            {t("workspace.planSwitcher.use")}
          </button>
        </div>
      </div>
    </div>
  );
};
