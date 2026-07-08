import type { FC } from "react";
import Markdown from "react-markdown";
import remarkGfm from "remark-gfm";

import { useAppRuntimeOptional } from "../../../runtime/AppRuntimeProvider";
import {
  buildResumeBodyMarkdown,
  getResumeDetailRenderer,
  getResumeSourceUrl,
  mapResumeToUiSource,
} from "../adapters/resume";
import type { Resume } from "../../../../api/generated/types";
import { PdfPageStackPreview } from "./PdfPageStackPreview";

export interface ResumePreviewTabProps {
  resume: Resume;
}

export const ResumePreviewTab: FC<ResumePreviewTabProps> = ({ resume }) => {
  const runtime = useAppRuntimeOptional();
  const uiResume = mapResumeToUiSource(resume);
  const bodyMarkdown = buildResumeBodyMarkdown(resume);
  const renderer = getResumeDetailRenderer(resume);
  const sourceUrl = getResumeSourceUrl(resume, runtime?.client.baseUrl);

  return (
    <div
      data-testid="resume-detail-preview-content"
      className="ei-resume-detail-preview"
    >
      <article className="ei-resume-detail-preview-card">
        {renderer === "pdf" ? (
          <PdfPageStackPreview sourceUrl={sourceUrl} label={uiResume.name} />
        ) : bodyMarkdown.trim() ? (
          <div
            data-testid="resume-detail-markdown-page"
            className="ei-resume-detail-markdown-page"
          >
            <div className="ei-text-body ei-resume-detail-preview-body">
              <Markdown remarkPlugins={[remarkGfm]} skipHtml>
                {bodyMarkdown}
              </Markdown>
            </div>
          </div>
        ) : (
          <div
            data-testid="resume-detail-markdown-page"
            className="ei-resume-detail-markdown-page"
          >
            <p className="ei-text-body ei-resume-detail-preview-empty">
              暂无可读简历正文。
            </p>
          </div>
        )}
      </article>
    </div>
  );
};
