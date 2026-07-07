import type { FC, ReactNode } from "react";

import { useI18n } from "../../../i18n/messages";
import { ResumeWorkshopIcon } from "../components/ResumeWorkshopIcon";

export interface PreviewDraft {
  name: string;
  title?: string;
  location?: string;
  contact: string[];
  summary?: string;
  experience: Array<{
    co: string;
    role: string;
    period: string;
    bullets: string[];
  }>;
  projects: Array<{ name: string; note?: string }>;
  skills: string[];
  education: Array<{ school: string; degree: string }>;
}

export interface ResumePreviewConfirmProps {
  sourceLabel: string;
  draft: PreviewDraft;
  onBack: () => void;
  onConfirm: () => void;
  submitting?: boolean;
  inlineError?: string | null;
}

export const ResumePreviewConfirm: FC<ResumePreviewConfirmProps> = ({
  sourceLabel,
  draft,
  onBack,
  onConfirm,
  submitting,
  inlineError,
}) => {
  const { t } = useI18n();

  const Section = ({ title, children }: { title: string; children: ReactNode }) => (
    <section className="ei-resume-preview-section">
      <h3 className="ei-text-label ei-resume-preview-section-title">{title}</h3>
      {children}
    </section>
  );

  return (
    <div
      className="ei-resume-create-preview"
      data-testid="resume-preview-confirm"
    >
      <button
        type="button"
        className="ei-resume-create-back"
        data-testid="resume-preview-confirm-back-link"
        onClick={onBack}
      >
        <ResumeWorkshopIcon name="arrowLeft" size={14} />
        {t("resumeWorkshop.preview.back")}
      </button>

      <header className="ei-resume-create-preview-header">
        <div>
          <span className="ei-text-label ei-resume-create-preview-eyebrow">
            {t("resumeWorkshop.preview.eyebrow")}
          </span>
          <h1 className="ei-text-display">
            {t("resumeWorkshop.preview.title")}
          </h1>
          <p
            className="ei-resume-create-preview-source"
            data-testid="resume-preview-confirm-source"
          >
            <span>{t("resumeWorkshop.preview.sourcePrefix")}</span>
            {sourceLabel}
            <span className="ei-resume-create-preview-source-divider" aria-hidden="true">
              ·
            </span>
            <span className="ei-resume-create-preview-status">
              {t("resumeWorkshop.preview.statusParsed")}
            </span>
          </p>
        </div>
        <div className="ei-resume-create-preview-actions">
          <button
            type="button"
            className="ei-resume-create-cta-ghost"
            data-testid="resume-preview-confirm-back-button"
            onClick={onBack}
            disabled={submitting}
          >
            {t("resumeWorkshop.preview.backCta")}
          </button>
          <button
            type="button"
            className="ei-resume-create-cta-accent"
            data-testid="resume-preview-confirm-save-button"
            onClick={onConfirm}
            disabled={submitting}
          >
            <ResumeWorkshopIcon name="check" size={14} />
            {t("resumeWorkshop.preview.confirm")}
          </button>
        </div>
      </header>

      {inlineError ? (
        <div
          className="ei-resume-create-error"
          role="alert"
          data-testid="resume-preview-confirm-inline-error"
        >
          {inlineError}
        </div>
      ) : null}

      <div
        className="ei-resume-create-preview-grid"
        data-testid="resume-preview-confirm-content"
      >
        <article className="ei-resume-create-preview-draft">
          <div className="ei-resume-create-preview-identity">
            <div
              className="ei-resume-create-preview-name"
              data-testid="resume-preview-confirm-name"
            >
              {draft.name}
            </div>
            {draft.title || draft.location ? (
              <div className="ei-resume-create-preview-headline">
                {draft.title}
                {draft.title && draft.location ? " · " : ""}
                {draft.location}
              </div>
            ) : null}
            {draft.contact.length > 0 ? (
              <div className="ei-resume-create-preview-contact">
                {draft.contact.join("  ·  ")}
              </div>
            ) : null}
          </div>

          {draft.summary ? (
            <Section title={t("resumeWorkshop.preview.section.summary")}>
              <p className="ei-text-body ei-resume-create-preview-summary">
                {draft.summary}
              </p>
            </Section>
          ) : null}

          {draft.experience.length > 0 ? (
            <Section title={t("resumeWorkshop.preview.section.experience")}>
              <ul className="ei-resume-create-preview-list">
                {draft.experience.map((entry, index) => (
                  <li key={`${entry.co}-${index}`}>
                    <div className="ei-resume-create-preview-list-head">
                      <div>
                        <div className="ei-resume-create-preview-list-role">
                          {entry.role}
                        </div>
                        <div className="ei-resume-create-preview-list-co">
                          {entry.co}
                        </div>
                      </div>
                      <div className="ei-resume-create-preview-list-period">
                        {entry.period}
                      </div>
                    </div>
                    {entry.bullets.length > 0 ? (
                      <ul className="ei-resume-create-preview-bullets">
                        {entry.bullets.map((bullet, bIndex) => (
                          <li key={bIndex}>{bullet}</li>
                        ))}
                      </ul>
                    ) : null}
                  </li>
                ))}
              </ul>
            </Section>
          ) : null}

          {draft.projects.length > 0 ? (
            <Section title={t("resumeWorkshop.preview.section.projects")}>
              <ul className="ei-resume-create-preview-projects">
                {draft.projects.map((project, index) => (
                  <li key={`${project.name}-${index}`}>
                    <span className="ei-resume-create-preview-project-name">
                      {project.name}
                    </span>
                    {project.note ? (
                      <span className="ei-resume-create-preview-project-note">
                        {project.note}
                      </span>
                    ) : null}
                  </li>
                ))}
              </ul>
            </Section>
          ) : null}

          {draft.skills.length > 0 ? (
            <Section title={t("resumeWorkshop.preview.section.skills")}>
              <ul className="ei-resume-create-preview-skills">
                {draft.skills.map((skill, index) => (
                  <li key={`${skill}-${index}`}>{skill}</li>
                ))}
              </ul>
            </Section>
          ) : null}

          {draft.education.length > 0 ? (
            <Section title={t("resumeWorkshop.preview.section.education")}>
              <ul className="ei-resume-create-preview-list">
                {draft.education.map((entry, index) => (
                  <li key={`${entry.school}-${index}`}>
                    <span className="ei-resume-create-preview-list-role">
                      {entry.school}
                    </span>
                    <span className="ei-resume-create-preview-list-period">
                      · {entry.degree}
                    </span>
                  </li>
                ))}
              </ul>
            </Section>
          ) : null}
        </article>

        <aside className="ei-resume-create-preview-sidebar">
          <div
            className="ei-resume-create-sidebar-card"
            data-testid="resume-preview-confirm-sidebar-what-saved"
          >
            <span className="ei-text-label ei-resume-create-sidebar-eyebrow">
              {t("resumeWorkshop.preview.sidebar.whatSavedEyebrow")}
            </span>
            <ul className="ei-resume-create-sidebar-list">
              <li>
                <ResumeWorkshopIcon name="file" size={15} />
                <div>
                  <div className="ei-resume-create-sidebar-item-title">
                    {t(
                      "resumeWorkshop.preview.sidebar.whatSaved.original.title",
                    )}
                  </div>
                  <p className="ei-resume-create-sidebar-item-body">
                    {t(
                      "resumeWorkshop.preview.sidebar.whatSaved.original.body",
                    )}
                  </p>
                </div>
              </li>
              <li>
                <ResumeWorkshopIcon name="resume" size={15} />
                <div>
                  <div className="ei-resume-create-sidebar-item-title">
                    {t(
                      "resumeWorkshop.preview.sidebar.whatSaved.structured.title",
                    )}
                  </div>
                  <p className="ei-resume-create-sidebar-item-body">
                    {t(
                      "resumeWorkshop.preview.sidebar.whatSaved.structured.body",
                    )}
                  </p>
                </div>
              </li>
              <li>
                <ResumeWorkshopIcon name="layers" size={15} />
                <div>
                  <div className="ei-resume-create-sidebar-item-title">
                    {t("resumeWorkshop.preview.sidebar.whatSaved.editable.title")}
                  </div>
                  <p className="ei-resume-create-sidebar-item-body">
                    {t("resumeWorkshop.preview.sidebar.whatSaved.editable.body")}
                  </p>
                </div>
              </li>
            </ul>
          </div>
          <div
            className="ei-resume-create-sidebar-card"
            data-testid="resume-preview-confirm-sidebar-parse-notes"
          >
            <span className="ei-text-label ei-resume-create-sidebar-eyebrow">
              {t("resumeWorkshop.preview.sidebar.notesEyebrow")}
            </span>
            <p className="ei-resume-create-sidebar-item-body">
              {t("resumeWorkshop.preview.sidebar.notesBody")}
            </p>
          </div>
        </aside>
      </div>
    </div>
  );
};
