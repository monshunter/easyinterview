import { useRef, useState, type FC } from "react";

import type { TargetJob } from "../../../api/generated/types";
import { generateIdempotencyKey } from "../../../lib/conventions/idempotency";
import { startPracticeFromParams } from "../../interview-context/startPractice";
import { PracticeLaunchTransition } from "../../interview-context/PracticeLaunchTransition";
import { useI18n, type MessageKey } from "../../i18n/messages";
import {
  isTargetJobPracticeStartable,
  targetJobDetailRouteParams,
  targetJobPracticeRouteParams,
} from "../../navigation/interviewContext";
import { useNavigation } from "../../navigation/NavigationProvider";
import { useAppRuntimeOptional } from "../../runtime/AppRuntimeProvider";
import type { Route } from "../../routes";
import { MockInterviewCard } from "../home/MockInterviewCard";
import { DestructiveActionDialog } from "../DestructiveActionDialog";
import { useWorkspaceTargetJobs } from "./hooks/useWorkspaceTargetJobs";

interface WorkspaceScreenProps {
  route: Route;
}

export const WorkspaceScreen: FC<WorkspaceScreenProps> = ({ route: _route }) => {
  return <WorkspacePlanList />;
};

export function formatWorkspaceSavedAt(value: string, lang: "zh" | "en"): string {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return value.slice(0, 10);
  return new Intl.DateTimeFormat(lang === "zh" ? "zh-CN" : "en-US", {
    month: "numeric",
    day: "numeric",
  }).format(date);
}

const WorkspacePlanList: FC = () => {
  const { lang, t } = useI18n();
  const runtime = useAppRuntimeOptional();
  const { navigate } = useNavigation();
  const { loading, jobs, error } = useWorkspaceTargetJobs();
  const [archivedJobIds, setArchivedJobIds] = useState<Set<string>>(() => new Set());
  const [startingJobId, setStartingJobId] = useState<string | null>(null);
  const [startError, setStartError] = useState<MessageKey | null>(null);
  const [deletingJobId, setDeletingJobId] = useState<string | null>(null);
  const [deleteError, setDeleteError] = useState<MessageKey | null>(null);
  const [deleteTarget, setDeleteTarget] = useState<TargetJob | null>(null);
  const deleteKeyRef = useRef<string | null>(null);
  const deletePendingRef = useRef(false);
  const deleteTriggerRef = useRef<HTMLElement | null>(null);

  const visibleJobs = jobs.filter((job) => !archivedJobIds.has(job.id));

  const openPlan = (job: TargetJob) => {
    navigate({
      name: "workspace",
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
    } catch {
      setStartError("workspace.errors.start");
    } finally {
      setStartingJobId(null);
    }
  };

  const openDelete = (job: TargetJob) => {
    deleteTriggerRef.current =
      document.activeElement instanceof HTMLElement ? document.activeElement : null;
    deleteKeyRef.current = generateIdempotencyKey();
    setDeleteError(null);
    setDeleteTarget(job);
  };

  const closeDelete = () => {
    if (deletePendingRef.current) return;
    setDeleteTarget(null);
    setDeleteError(null);
    deleteKeyRef.current = null;
  };

  const deletePlan = async () => {
    if (!deleteTarget || !deleteKeyRef.current || deletePendingRef.current) return;
    if (!runtime || runtime.auth.status !== "authenticated") {
      setDeleteError("workspace.planList.deleteAuthError");
      return;
    }

    const job = deleteTarget;
    deletePendingRef.current = true;
    setDeleteError(null);
    setDeletingJobId(job.id);
    try {
      await runtime.client.archiveTargetJob(job.id, {
        idempotencyKey: deleteKeyRef.current,
      });
      setArchivedJobIds((current) => {
        const next = new Set(current);
        next.add(job.id);
        return next;
      });
      setDeleteTarget(null);
      deleteKeyRef.current = null;
    } catch {
      setDeleteError("workspace.errors.delete");
    } finally {
      deletePendingRef.current = false;
      setDeletingJobId(null);
    }
  };

  return (
    <>
      {startingJobId ? <PracticeLaunchTransition /> : null}
      <main
        data-testid="workspace-plan-list"
        className="ei-workspace-plan-list ei-fadein"
      >
        <div
          data-testid="workspace-plan-inner"
          className="ei-workspace-plan-inner"
        >
          <header className="ei-workspace-plan-header">
            <div className="ei-workspace-plan-heading-copy">
              <div
                data-testid="workspace-plan-list-eyebrow"
                className="ei-workspace-plan-eyebrow"
              >
                {t("workspace.planList.eyebrow")}
              </div>
              <h1
                data-testid="workspace-plan-list-title"
                className="ei-workspace-plan-title"
              >
                {t("workspace.planList.title")}
              </h1>
              <div
                data-testid="workspace-plan-list-subtitle"
                className="ei-workspace-plan-subtitle"
              >
                {t("workspace.planList.subtitle")}
              </div>
            </div>
            <button
              data-testid="workspace-plan-list-create"
              type="button"
              className="ei-workspace-plan-create"
              onClick={() => navigate({ name: "home", params: {} })}
            >
              <svg aria-hidden="true" width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round">
                <circle cx="12" cy="12" r="9" />
                <path d="M12 8v8M8 12h8" />
              </svg>
              {t("workspace.planList.create")}
            </button>
          </header>

          {loading ? (
            <div
              data-testid="workspace-plan-list-loading"
              className="ei-workspace-plan-state"
            >
              {t("workspace.planList.loading")}
            </div>
          ) : error ? (
            <div
              data-testid="workspace-plan-list-error"
              className="ei-workspace-plan-state"
            >
              {t("workspace.planList.error")}
            </div>
          ) : visibleJobs.length === 0 ? (
            <div
              data-testid="workspace-plan-list-empty"
              className="ei-workspace-plan-state ei-workspace-plan-state-empty"
            >
              <div className="ei-workspace-plan-state-title">
                {t("workspace.planList.emptyTitle")}
              </div>
              <div className="ei-workspace-plan-state-copy">
                {t("workspace.planList.emptyDesc")}
              </div>
            </div>
          ) : (
            <div
              data-testid="workspace-plan-list-grid"
              className="ei-workspace-plan-grid"
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
                  footer={(
                    <span className="ei-workspace-card-saved">
                      <svg aria-hidden="true" width="19" height="19" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round">
                        <circle cx="12" cy="12" r="9" />
                        <path d="M12 7v5l3.5 2" />
                      </svg>
                      {t("workspace.planList.lastSaved")}：{formatWorkspaceSavedAt(job.updatedAt, lang)}
                    </span>
                  )}
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
                    onClick: () => openDelete(job),
                    disabled: deletingJobId === job.id,
                  }}
                />
              ))}
            </div>
          )}
          {startError ? (
            <div
              data-testid="workspace-plan-list-start-error"
              className="ei-workspace-plan-error"
            >
              {t(startError)}
            </div>
          ) : null}
        </div>
      </main>
      {deleteTarget ? (
        <DestructiveActionDialog
          eyebrow={t("destructiveDialog.eyebrow")}
          title={t("workspace.planList.deleteQuestion")}
          description={t("workspace.planList.deleteDescription")}
          cancelLabel={t("destructiveDialog.cancel")}
          confirmLabel={t("destructiveDialog.confirm")}
          pendingLabel={t("destructiveDialog.pending")}
          retryLabel={t("destructiveDialog.retry")}
          errorMessage={deleteError ? t(deleteError) : null}
          errorTestId="workspace-plan-list-delete-error"
          pending={deletingJobId === deleteTarget.id}
          restoreFocusRef={deleteTriggerRef}
          onCancel={closeDelete}
          onConfirm={() => void deletePlan()}
        />
      ) : null}
    </>
  );
};
