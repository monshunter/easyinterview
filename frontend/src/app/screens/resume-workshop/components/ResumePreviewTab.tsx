import type { FC } from "react";

import { useI18n } from "../../../i18n/messages";
import {
  buildResumePlainText,
  buildResumePreview,
} from "../adapters/resume";
import type { ResumeVersion } from "../../../../api/generated/types";
import { fireResumeWorkshopToast } from "./toast";

export interface ResumePreviewTabProps {
  version: ResumeVersion;
  onViewOriginal: () => void;
  onExport: () => void;
}

export const ResumePreviewTab: FC<ResumePreviewTabProps> = ({
  version,
  onViewOriginal,
  onExport,
}) => {
  const { t } = useI18n();
  const projection = buildResumePreview(version);

  const onCopy = async () => {
    const text = buildResumePlainText(version);
    if (navigator.clipboard?.writeText) {
      try {
        await navigator.clipboard.writeText(text);
        fireResumeWorkshopToast(t("resumeWorkshop.detail.copySuccess"), "ok");
        return;
      } catch {
        // fall through to warn toast
      }
    }
    fireResumeWorkshopToast(
      t("resumeWorkshop.detail.copyUnavailable"),
      "warn",
    );
  };

  return (
    <div
      data-testid="resume-detail-preview-content"
      className="ei-resume-detail-preview"
      role="tabpanel"
    >
      <article className="ei-resume-detail-preview-card">
        {projection.headline ? (
          <h3 className="ei-text-title">{projection.headline}</h3>
        ) : null}
        {projection.summary ? (
          <p className="ei-text-body">{projection.summary}</p>
        ) : null}
        {projection.sections.map((section) => (
          <section key={section.title} className="ei-resume-detail-preview-section">
            <h4 className="ei-text-label">{section.title}</h4>
            <ul>
              {section.bullets.map((bullet) => (
                <li key={bullet} className="ei-text-body">
                  {bullet}
                </li>
              ))}
            </ul>
          </section>
        ))}
        {projection.skills.length > 0 ? (
          <p className="ei-text-body ei-resume-detail-preview-skills">
            {projection.skills.join(" · ")}
          </p>
        ) : null}
      </article>
      <aside className="ei-resume-detail-preview-actions">
        <button
          type="button"
          data-testid="resume-detail-export-pdf"
          onClick={onExport}
        >
          {t("resumeWorkshop.detail.exportPdf")}
        </button>
        <button
          type="button"
          data-testid="resume-detail-copy-text"
          onClick={onCopy}
        >
          {t("resumeWorkshop.detail.copyText")}
        </button>
        <button
          type="button"
          data-testid="resume-detail-view-original"
          onClick={onViewOriginal}
        >
          {t("resumeWorkshop.detail.viewOriginal")}
        </button>
      </aside>
    </div>
  );
};
