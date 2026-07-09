import { useEffect, useState, type FC } from "react";

import type { TargetJob } from "../../../api/generated/types";
import { useI18n, type MessageKey } from "../../i18n/messages";
import { useNavigation } from "../../navigation/NavigationProvider";
import type { Route } from "../../routes";
import { useWorkspaceTargetJobs } from "./hooks/useWorkspaceTargetJobs";

interface WorkspaceScreenProps {
  route: Route;
}

export const WorkspaceScreen: FC<WorkspaceScreenProps> = ({ route: _route }) => {
  const workspaceListCompactLayout = useWorkspaceCompactLayout();
  return <WorkspacePlanList compactLayout={workspaceListCompactLayout} />;
};

interface WorkspacePlanListProps {
  compactLayout: boolean;
}

const WorkspacePlanList: FC<WorkspacePlanListProps> = ({ compactLayout }) => {
  const { t } = useI18n();
  const { navigate } = useNavigation();
  const { loading, jobs, error } = useWorkspaceTargetJobs();

  const openPlan = (job: TargetJob) => {
    const currentPracticePlanId = job.currentPracticePlanId?.trim();
    const resumeId = job.resumeId?.trim();
    navigate({
      name: "parse",
      params: {
        targetJobId: job.id,
        ...(currentPracticePlanId ? { planId: currentPracticePlanId } : {}),
        ...(resumeId ? { resumeId } : {}),
      },
    });
  };

  return (
    <div
      data-testid="workspace-plan-list"
      className="ei-fadein"
      style={{
        maxWidth: 1120,
        margin: "0 auto",
        padding: compactLayout ? "32px 16px 72px" : "48px 48px 96px",
      }}
    >
      <div
        style={{
          display: "flex",
          justifyContent: "space-between",
          alignItems: "flex-start",
          gap: 24,
          flexWrap: "wrap",
          marginBottom: 28,
        }}
      >
        <div style={{ maxWidth: 640 }}>
          <div
            data-testid="workspace-plan-list-eyebrow"
            className="ei-label"
            style={{ color: "var(--ei-color-fg-tertiary)", marginBottom: 8 }}
          >
            {t("workspace.planList.eyebrow")}
          </div>
          <h1
            data-testid="workspace-plan-list-title"
            className="ei-serif"
            style={{
              fontSize: compactLayout ? 30 : 40,
              color: "var(--ei-color-fg-primary)",
              margin: 0,
              lineHeight: 1.14,
            }}
          >
            {t("workspace.planList.title")}
          </h1>
          <div
            data-testid="workspace-plan-list-subtitle"
            style={{
              fontSize: 14,
              color: "var(--ei-color-fg-secondary)",
              marginTop: 10,
              lineHeight: 1.6,
            }}
          >
            {t("workspace.planList.subtitle")}
          </div>
        </div>
        <button
          data-testid="workspace-plan-list-create"
          type="button"
          onClick={() => navigate({ name: "home", params: {} })}
          style={{
            height: 34,
            padding: "0 16px",
            fontSize: 13,
            fontWeight: 500,
            background: "var(--ei-color-accent)",
            color: "#fff",
            border: "1px solid var(--ei-color-accent)",
            borderRadius: 2,
            cursor: "pointer",
            fontFamily: "var(--ei-sans)",
          }}
        >
          {t("workspace.planList.create")}
        </button>
      </div>

      {loading ? (
        <div
          data-testid="workspace-plan-list-loading"
          style={{
            background: "var(--ei-color-bg-card)",
            border: "1px solid var(--ei-color-rule-strong)",
            borderRadius: 3,
            padding: 24,
            color: "var(--ei-color-fg-tertiary)",
            fontSize: 13,
          }}
        >
          {t("workspace.planList.loading")}
        </div>
      ) : error ? (
        <div
          data-testid="workspace-plan-list-error"
          style={{
            background: "var(--ei-color-bg-card)",
            border: "1px solid var(--ei-color-rule-strong)",
            borderRadius: 3,
            padding: 24,
            color: "var(--ei-color-fg-tertiary)",
            fontSize: 13,
          }}
        >
          {t("workspace.planList.error")}
        </div>
      ) : jobs.length === 0 ? (
        <div
          data-testid="workspace-plan-list-empty"
          style={{
            background: "var(--ei-color-bg-card)",
            border: "1px solid var(--ei-color-rule-strong)",
            borderRadius: 3,
            padding: 32,
            textAlign: "center",
          }}
        >
          <div
            className="ei-serif"
            style={{ fontSize: 18, color: "var(--ei-color-fg-primary)", marginBottom: 10 }}
          >
            {t("workspace.planList.emptyTitle")}
          </div>
          <div style={{ fontSize: 13, color: "var(--ei-color-fg-tertiary)", lineHeight: 1.55 }}>
            {t("workspace.planList.emptyDesc")}
          </div>
        </div>
      ) : (
        <div
          data-testid="workspace-plan-list-grid"
          style={{
            display: "grid",
            gridTemplateColumns: compactLayout
              ? "minmax(0, 1fr)"
              : "repeat(auto-fit, minmax(300px, 1fr))",
            gap: 16,
            alignItems: "stretch",
          }}
        >
          {jobs.map((job) => {
            const statusTone = getStatusTone(job.status);
            return (
              <article
                key={job.id}
                data-testid={`workspace-plan-list-card-${job.id}`}
                style={{
                  background: "var(--ei-color-bg-card)",
                  border: "1px solid var(--ei-color-rule-strong)",
                  borderRadius: 3,
                  boxShadow: "var(--ei-shadow-elev2)",
                  minHeight: 178,
                  display: "flex",
                  flexDirection: "column",
                  justifyContent: "space-between",
                  overflow: "hidden",
                }}
              >
                <div
                  data-testid={`workspace-plan-list-card-body-${job.id}`}
                  style={{
                    padding: 20,
                    flex: 1,
                    background: "var(--ei-color-bg-card)",
                  }}
                >
                  <div
                    style={{
                      display: "flex",
                      justifyContent: "space-between",
                      gap: 12,
                      alignItems: "center",
                      marginBottom: 10,
                    }}
                  >
                    <span
                      className="ei-mono"
                      style={{
                        display: "inline-flex",
                        alignItems: "center",
                        padding: "3px 8px",
                        borderRadius: 3,
                        fontSize: 11.5,
                        letterSpacing: "0.04em",
                        background:
                          statusTone === "amber"
                            ? "var(--ei-color-amber-soft)"
                            : statusTone === "muted"
                              ? "var(--ei-color-bg-soft)"
                              : "transparent",
                        color:
                          statusTone === "amber"
                            ? "var(--ei-color-warn)"
                            : statusTone === "muted"
                              ? "var(--ei-color-fg-tertiary)"
                              : "var(--ei-color-fg-secondary)",
                        border:
                          statusTone === "neutral"
                            ? "1px solid var(--ei-color-rule-strong)"
                            : "1px solid transparent",
                        whiteSpace: "nowrap",
                      }}
                    >
                      {formatStatus(job.status, t)}
                    </span>
                    <span
                      className="ei-mono"
                      style={{ fontSize: 12, color: "var(--ei-color-fg-tertiary)" }}
                    >
                      {t("workspace.planList.updated").replace("{date}", formatDate(job.updatedAt))}
                    </span>
                  </div>
                  <div
                    className="ei-serif"
                    style={{
                      fontSize: 20,
                      color: "var(--ei-color-fg-primary)",
                      lineHeight: 1.25,
                      marginBottom: 6,
                    }}
                  >
                    {job.title}
                  </div>
                  <div
                    style={{
                      fontSize: 13,
                      color: "var(--ei-color-fg-secondary)",
                      lineHeight: 1.5,
                    }}
                  >
                    {[job.companyName, job.locationText].filter(Boolean).join(" · ")}
                  </div>
                </div>
                <div
                  data-testid={`workspace-plan-list-card-footer-${job.id}`}
                  style={{
                    borderTop: "1px solid var(--ei-color-rule-strong)",
                    padding: "14px 20px",
                    background: "var(--ei-color-bg-card)",
                    display: "flex",
                    justifyContent: "flex-end",
                    alignItems: "center",
                    gap: 12,
                  }}
                >
                  <button
                    data-testid={`workspace-plan-list-open-${job.id}`}
                    type="button"
                    onClick={() => openPlan(job)}
                    style={{
                      flex: "0 0 auto",
                      height: 32,
                      padding: "0 12px",
                      fontSize: 13,
                      fontWeight: 500,
                      background: "var(--ei-color-accent)",
                      color: "#fff",
                      border: "1px solid var(--ei-color-accent)",
                      borderRadius: 2,
                      cursor: "pointer",
                      fontFamily: "var(--ei-sans)",
                    }}
                  >
                    {t("workspace.planList.open")}
                  </button>
                </div>
              </article>
            );
          })}
        </div>
      )}
    </div>
  );
};

type StatusTone = "amber" | "muted" | "neutral";

function getStatusTone(status: string): StatusTone {
  switch (status) {
    case "applied":
    case "interviewing":
      return "amber";
    case "draft":
    case "preparing":
      return "muted";
    default:
      return "neutral";
  }
}

function formatStatus(status: string, t: (key: MessageKey) => string): string {
  const map: Record<string, MessageKey> = {
    draft: "workspace.status.draft",
    preparing: "workspace.status.preparing",
    applied: "workspace.status.applied",
    interviewing: "workspace.status.interviewing",
    offer: "workspace.status.offer",
    rejected: "workspace.status.rejected",
    archived: "workspace.status.archived",
  };
  const key = map[status];
  return key ? t(key) : status;
}

function formatDate(iso: string): string {
  try {
    const d = new Date(iso);
    return `${d.getMonth() + 1}/${d.getDate()}`;
  } catch {
    return "—";
  }
}

function useWorkspaceCompactLayout(): boolean {
  const [compact, setCompact] = useState(() => {
    if (typeof window === "undefined") return false;
    if (typeof window.matchMedia !== "function") return false;
    return window.matchMedia("(max-width: 720px)").matches;
  });

  useEffect(() => {
    if (typeof window === "undefined") return;
    if (typeof window.matchMedia !== "function") return;
    const query = window.matchMedia("(max-width: 720px)");
    const update = () => setCompact(query.matches);
    update();
    query.addEventListener("change", update);
    return () => query.removeEventListener("change", update);
  }, []);

  return compact;
}
