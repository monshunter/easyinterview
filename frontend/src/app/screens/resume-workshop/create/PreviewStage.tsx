import { useState, type FC } from "react";

import type { Resume } from "../../../../api/generated/types";
import { useI18n } from "../../../i18n/messages";
import { useNavigation } from "../../../navigation/NavigationProvider";
import { fireResumeWorkshopToast } from "../components/toast";
import { useResumeSave, ResumeSaveError } from "../hooks/useResumeSave";
import { ResumePreviewConfirm } from "./ResumePreviewConfirm";
import type { PreviewDraft } from "./ResumePreviewConfirm";
import {
  buildStructuredProfilePayload,
  mapParsedSummaryToStructuredProfileDraft,
} from "./adapters/mapParsedSummaryToStructuredProfileDraft";

export interface PreviewStageProps {
  resume: Resume;
  draft: PreviewDraft;
  sourceLabel: string;
  onBack: () => void;
  onSaved: () => void;
}

/**
 * D-20 preview-confirm: the resume already exists from `registerResume`. The
 * confirm step writes the parsed structured profile into that flat resume via
 * `updateResume` (no `confirmStructuredMaster` master step), then returns to
 * the workshop list.
 */
export const PreviewStage: FC<PreviewStageProps> = ({
  resume,
  draft,
  sourceLabel,
  onBack,
  onSaved,
}) => {
  const { t } = useI18n();
  const { navigate } = useNavigation();
  const { overwrite } = useResumeSave();
  const [submitting, setSubmitting] = useState(false);
  const [inlineError, setInlineError] = useState<string | null>(null);

  // Re-derive in case the resume changed; the prop draft is the cached version.
  const renderedDraft = draft.name
    ? draft
    : mapParsedSummaryToStructuredProfileDraft(resume);

  const handleConfirm = async () => {
    if (submitting) return;
    setSubmitting(true);
    setInlineError(null);
    const displayName = renderedDraft.name || resume.title;
    const structuredProfile = buildStructuredProfilePayload(resume);
    try {
      await overwrite(resume.id, { displayName, structuredProfile });
      setSubmitting(false);
      fireResumeWorkshopToast(t("resumeWorkshop.preview.success"), "ok");
      onSaved();
      navigate({ name: "resume_versions", params: {} });
    } catch (err) {
      setSubmitting(false);
      if (err instanceof ResumeSaveError && err.kind === "validation") {
        setInlineError(t("resumeWorkshop.create.errors.validation"));
        return;
      }
      setInlineError(t("resumeWorkshop.create.errors.confirmFailed"));
    }
  };

  return (
    <ResumePreviewConfirm
      sourceLabel={sourceLabel}
      draft={renderedDraft}
      onBack={onBack}
      onConfirm={handleConfirm}
      submitting={submitting}
      inlineError={inlineError}
    />
  );
};
