import { useCallback, useMemo, useState, type FC } from "react";

import { generateIdempotencyKey } from "../../../../lib/conventions/idempotency";
import { useDisplayPreferencesOptional } from "../../../display/DisplayPreferencesProvider";
import { useI18n } from "../../../i18n/messages";
import { useNavigation } from "../../../navigation/NavigationProvider";
import { useAppRuntimeOptional } from "../../../runtime/AppRuntimeProvider";
import {
  mapResumeSummaryToUiSource,
  type UiResumeListItem,
} from "../adapters/resume";
import { useResumes } from "../hooks/useResumes";
import { ResumeWorkshopIcon } from "./ResumeWorkshopIcon";

/** D-20 flat resume cards sorted by last edit. */
export const ResumeListView: FC = () => {
  const { t } = useI18n();
  const { navigate } = useNavigation();
  const runtime = useAppRuntimeOptional();
  const lang = useDisplayPreferencesOptional()?.lang ?? "zh";
  const resumesQuery = useResumes();
  const [archivedIds, setArchivedIds] = useState<Set<string>>(() => new Set());
  const [deletingId, setDeletingId] = useState<string | null>(null);
  const [deleteError, setDeleteError] = useState<string | null>(null);

  const resumes = useMemo<UiResumeListItem[]>(
    () =>
      resumesQuery.data?.items
        .filter((resume) => !archivedIds.has(resume.id))
        .map(mapResumeSummaryToUiSource) ?? [],
    [archivedIds, resumesQuery.data],
  );

  const sorted = useMemo(
    () => [...resumes].sort((a, b) => b.updatedAt.localeCompare(a.updatedAt)),
    [resumes],
  );

  const onOpen = useCallback(
    (resumeId: string) => {
      navigate({
        name: "resume_versions",
        params: { resumeId },
      });
    },
    [navigate],
  );
  const onCreate = useCallback(() => {
    navigate({
      name: "resume_versions",
      params: { flow: "create" },
    });
  }, [navigate]);
  const onDelete = useCallback(
    async (resumeId: string) => {
      if (!runtime?.client || runtime.auth.status !== "authenticated") return;
      setDeletingId(resumeId);
      setDeleteError(null);
      try {
        await runtime.client.archiveResume(resumeId, {
          headers: { "Accept-Language": lang },
          idempotencyKey: generateIdempotencyKey(),
        });
        setArchivedIds((current) => {
          const next = new Set(current);
          next.add(resumeId);
          return next;
        });
      } catch {
        setDeleteError(t("resumeWorkshop.list.deleteError"));
      } finally {
        setDeletingId((current) => (current === resumeId ? null : current));
      }
    },
    [lang, runtime, t],
  );

  if (resumesQuery.loading) {
    return (
      <div data-testid="resume-workshop-list" className="ei-screen-card">
        <span className="ei-text-body" role="status">
          {t("resumeWorkshop.list.loading")}
        </span>
      </div>
    );
  }

  if (resumesQuery.error) {
    return (
      <div data-testid="resume-workshop-list" className="ei-screen-card">
        <p className="ei-text-body" role="alert">
          {t("resumeWorkshop.list.error")}
        </p>
        <button
          type="button"
          className="ei-cta"
          data-testid="resume-workshop-list-retry"
          onClick={resumesQuery.retry}
        >
          {t("workspace.errors.retry")}
        </button>
      </div>
    );
  }

  return (
    <div data-testid="resume-workshop-list" className="ei-resume-workshop-list">
      <header className="ei-resume-workshop-list-header">
        <div>
          <span className="ei-text-label">{t("resumeWorkshop.eyebrow")}</span>
          <h1 className="ei-text-display">{t("resumeWorkshop.title")}</h1>
          <p className="ei-text-body">{t("resumeWorkshop.subtitle")}</p>
        </div>
        <button
          type="button"
          className="ei-resume-workshop-create"
          data-testid="resume-workshop-create"
          onClick={onCreate}
        >
          <ResumeWorkshopIcon name="plus" size={22} />
          {t("resumeWorkshop.create")}
        </button>
      </header>

      {sorted.length === 0 ? (
        <div data-testid="resume-workshop-list-empty" className="ei-screen-card">
          <p className="ei-text-body">{t("resumeWorkshop.list.empty")}</p>
        </div>
      ) : (
        <ul
          aria-label={t("resumeWorkshop.title")}
          className="ei-resume-workshop-card-grid"
          data-testid="resume-workshop-card-grid"
        >
          {sorted.map((resume) => (
            <li
              key={resume.id}
              data-testid={`resume-list-card-${resume.id}`}
              className="ei-resume-workshop-card"
            >
              <button
                type="button"
                aria-label={`${t("resumeWorkshop.list.delete")} ${resume.name}`}
                title={t("resumeWorkshop.list.delete")}
                data-testid={`resume-list-delete-${resume.id}`}
                className="ei-resume-workshop-card-delete"
                disabled={deletingId === resume.id}
                onClick={() => void onDelete(resume.id)}
              >
                <ResumeWorkshopIcon name="trash" size={14} />
              </button>

              <div className="ei-resume-workshop-card-heading">
                <span className="ei-resume-workshop-card-icon">
                  <ResumeWorkshopIcon name="resume" size={29} />
                </span>
                <div className="ei-resume-workshop-card-heading-copy">
                  <h2>{resume.name}</h2>
                  {resume.summary ? (
                    <p className="ei-resume-workshop-card-summary">{resume.summary}</p>
                  ) : null}
                </div>
              </div>

              <dl className="ei-resume-workshop-card-meta">
                <div>
                  <dt>{t("resumeWorkshop.list.colSource")}</dt>
                  <dd>{resume.sourceName}</dd>
                </div>
                <div>
                  <dt>{t("resumeWorkshop.list.colLastEdit")}</dt>
                  <dd>{resume.updatedAt}</dd>
                </div>
              </dl>

              <div className="ei-resume-workshop-card-footer">
                <button
                  type="button"
                  aria-label={`${t("resumeWorkshop.list.open")} ${resume.name}`}
                  data-testid={`resume-list-open-${resume.id}`}
                  className="ei-resume-workshop-card-open"
                  onClick={() => onOpen(resume.id)}
                >
                  {t("resumeWorkshop.list.open")}
                </button>
              </div>
            </li>
          ))}
        </ul>
      )}

      {deleteError ? (
        <p
          className="ei-resume-workshop-delete-error"
          data-testid="resume-workshop-delete-error"
          role="alert"
        >
          {deleteError}
        </p>
      ) : null}

      {resumesQuery.data?.pageInfo.hasMore ? (
        <div
          data-testid="resume-workshop-list-paginated"
          className="ei-text-body"
        >
          {t("resumeWorkshop.list.paginated")}
        </div>
      ) : null}
    </div>
  );
};
