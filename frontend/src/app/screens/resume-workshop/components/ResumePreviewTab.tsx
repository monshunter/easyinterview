import type { FC } from "react";

import { buildResumeBodyLines, mapResumeToUiSource } from "../adapters/resume";
import type { Resume } from "../../../../api/generated/types";

export interface ResumePreviewTabProps {
  resume: Resume;
}

export const ResumePreviewTab: FC<ResumePreviewTabProps> = ({ resume }) => {
  const uiResume = mapResumeToUiSource(resume);
  const bodyLines = buildResumeBodyLines(resume);

  return (
    <div
      data-testid="resume-detail-preview-content"
      className="ei-resume-detail-preview"
    >
      <article className="ei-resume-detail-preview-card">
        <h3 className="ei-text-title">{uiResume.name}</h3>
        {bodyLines.length > 0 ? (
          <div className="ei-text-body ei-resume-detail-preview-body">
            {bodyLines.map((line, index) => (
              <p key={`${index}-${line}`}>{line}</p>
            ))}
          </div>
        ) : (
          <p className="ei-text-body ei-resume-detail-preview-empty">
            {uiResume.summary}
          </p>
        )}
      </article>
    </div>
  );
};
