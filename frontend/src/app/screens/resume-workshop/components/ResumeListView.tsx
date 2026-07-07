import { useCallback, useMemo, type FC } from "react";

import { useI18n } from "../../../i18n/messages";
import { useNavigation } from "../../../navigation/NavigationProvider";
import { mapResumeToUiSource, type UiResumeSource } from "../adapters/resume";
import { useResumeAssets } from "../hooks/useResumeAssets";
import { ResumeWorkshopIcon } from "./ResumeWorkshopIcon";

/**
 * D-20 flat list view — source-level replica of `ResumeListView` in
 * ui-design/src/screen-resume-workshop.jsx. A single flat table of resumes
 * sorted by last edit.
 */
export const ResumeListView: FC = () => {
  const { t } = useI18n();
  const { navigate } = useNavigation();
  const resumesQuery = useResumeAssets();

  const resumes = useMemo<UiResumeSource[]>(
    () => resumesQuery.data?.items.map(mapResumeToUiSource) ?? [],
    [resumesQuery.data],
  );

  const sorted = useMemo(
    () => [...resumes].sort((a, b) => b.updatedAt.localeCompare(a.updatedAt)),
    [resumes],
  );

  const onOpen = useCallback(
    (resumeId: string) => {
      navigate({
        name: "resume_versions",
        params: { resumeId, tab: "preview" },
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
          <ResumeWorkshopIcon name="plus" size={14} />
          {t("resumeWorkshop.create")}
        </button>
      </header>

      {sorted.length === 0 ? (
        <div data-testid="resume-workshop-list-empty" className="ei-screen-card">
          <p className="ei-text-body">{t("resumeWorkshop.list.empty")}</p>
        </div>
      ) : (
        <div className="ei-resume-workshop-table" data-testid="resume-workshop-table">
          <div className="ei-resume-workshop-table-head" role="row">
            <span role="columnheader">{t("resumeWorkshop.list.colResume")}</span>
            <span role="columnheader">{t("resumeWorkshop.list.colSource")}</span>
            <span role="columnheader">{t("resumeWorkshop.list.colLang")}</span>
            <span role="columnheader">{t("resumeWorkshop.list.colLastEdit")}</span>
            <span role="columnheader" aria-hidden="true" />
          </div>
          {sorted.map((resume) => (
            <div
              key={resume.id}
              role="row"
              data-testid={`resume-list-row-${resume.id}`}
              className="ei-resume-workshop-table-row"
            >
              <div className="ei-resume-workshop-table-name">
                <ResumeWorkshopIcon name="resume" size={13} />
                <div>
                  <div className="ei-resume-workshop-table-name-main">
                    {resume.name}
                  </div>
                  <div className="ei-resume-workshop-table-name-sub">
                    {resume.summary}
                  </div>
                </div>
              </div>
              <div className="ei-resume-workshop-table-source">
                {resume.sourceName}
              </div>
              <div className="ei-resume-workshop-table-lang">
                <span className="ei-resume-workshop-lang-tag">
                  {resume.langTag}
                </span>
              </div>
              <div className="ei-resume-workshop-table-date">
                {resume.updatedAt}
              </div>
              <button
                type="button"
                data-testid={`resume-list-open-${resume.id}`}
                className="ei-resume-workshop-table-open"
                onClick={() => onOpen(resume.id)}
              >
                {t("resumeWorkshop.list.open")}
              </button>
            </div>
          ))}
        </div>
      )}

      <button
        type="button"
        className="ei-resume-workshop-upload-cta"
        data-testid="resume-workshop-upload-cta"
        onClick={onCreate}
      >
        <ResumeWorkshopIcon name="upload" size={14} />
        {t("resumeWorkshop.list.uploadAnother")}
      </button>

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
