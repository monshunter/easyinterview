import { useEffect, useState, type FC } from "react";

import type { TargetJob } from "../../../api/generated/types";
import { generateIdempotencyKey } from "../../../lib/conventions/idempotency";
import { startPracticeFromParams } from "../../interview-context/startPractice";
import { useI18n } from "../../i18n/messages";
import {
  isTargetJobPracticeStartable,
  targetJobDetailRouteParams,
  targetJobPracticeRouteParams,
} from "../../navigation/interviewContext";
import { useNavigation } from "../../navigation/NavigationProvider";
import { useAppRuntimeOptional } from "../../runtime/AppRuntimeProvider";
import type { Route } from "../../routes";
import { MockInterviewCard } from "../home/MockInterviewCard";
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
  const { lang, t } = useI18n();
  const runtime = useAppRuntimeOptional();
  const { navigate } = useNavigation();
  const { loading, jobs, error } = useWorkspaceTargetJobs();
  const [archivedJobIds, setArchivedJobIds] = useState<Set<string>>(() => new Set());
  const [startingJobId, setStartingJobId] = useState<string | null>(null);
  const [startError, setStartError] = useState<string | null>(null);
  const [deletingJobId, setDeletingJobId] = useState<string | null>(null);
  const [deleteError, setDeleteError] = useState<string | null>(null);

  const visibleJobs = jobs.filter((job) => !archivedJobIds.has(job.id));

  const openPlan = (job: TargetJob) => {
    navigate({
      name: "parse",
      params: targetJobDetailRouteParams(job),
    });
  };

  const startInterview = async (job: TargetJob) => {
    const params = targetJobPracticeRouteParams(job);
    if (
      !runtime ||
      runtime.auth.status !== "authenticated" ||
      !params.resumeId ||
      !params.roundId
    ) {
      openPlan(job);
      return;
    }

    setStartError(null);
    setStartingJobId(job.id);
    try {
      const started = await startPracticeFromParams(runtime.client, params, lang);
      navigate({ name: "practice", params: started.params });
    } catch (err: unknown) {
      setStartError(err instanceof Error ? err.message : String(err));
    } finally {
      setStartingJobId(null);
    }
  };

  const deletePlan = async (job: TargetJob) => {
    if (!runtime || runtime.auth.status !== "authenticated") {
      setDeleteError(t("workspace.planList.deleteAuthError"));
      return;
    }

    setDeleteError(null);
    setDeletingJobId(job.id);
    try {
      await runtime.client.archiveTargetJob(job.id, {
        idempotencyKey: generateIdempotencyKey(),
      });
      setArchivedJobIds((current) => {
        const next = new Set(current);
        next.add(job.id);
        return next;
      });
    } catch (err: unknown) {
      setDeleteError(err instanceof Error ? err.message : String(err));
    } finally {
      setDeletingJobId(null);
    }
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
      ) : visibleJobs.length === 0 ? (
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
              : "repeat(auto-fill, minmax(300px, 360px))",
            justifyContent: "start",
            gap: 16,
            alignItems: "stretch",
          }}
        >
          {visibleJobs.map((job) => (
            <MockInterviewCard
              key={job.id}
              job={job}
              onClick={() => openPlan(job)}
              cardTestId={`workspace-plan-list-card-${job.id}`}
              bodyTestId={`workspace-plan-list-card-body-${job.id}`}
              railTestId={`workspace-plan-list-rail-${job.id}`}
              footerTestId={`workspace-plan-list-card-footer-${job.id}`}
              primaryAction={{
                label: t("workspace.planList.start"),
                testId: `workspace-plan-list-start-${job.id}`,
                onClick: () => startInterview(job),
                disabled:
                  startingJobId === job.id ||
                  !isTargetJobPracticeStartable(job),
              }}
              deleteAction={{
                label: t("workspace.planList.delete"),
                testId: `workspace-plan-list-delete-${job.id}`,
                onClick: () => deletePlan(job),
                disabled: deletingJobId === job.id,
              }}
            />
          ))}
        </div>
      )}
      {deleteError ? (
        <div
          data-testid="workspace-plan-list-delete-error"
          style={{
            color: "var(--ei-color-danger)",
            fontSize: 13,
            marginTop: 12,
          }}
        >
          {deleteError}
        </div>
      ) : null}
      {startError ? (
        <div
          data-testid="workspace-plan-list-start-error"
          style={{
            color: "var(--ei-color-danger)",
            fontSize: 13,
            marginTop: 12,
          }}
        >
          {startError}
        </div>
      ) : null}
    </div>
  );
};

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
