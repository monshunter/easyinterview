import { useState, type FC } from "react";

import type { ResumeAsset } from "../../../../api/generated/types";
import { useI18n } from "../../../i18n/messages";
import { useNavigation } from "../../../navigation/NavigationProvider";
import { fireResumeWorkshopToast } from "../components/toast";
import { useResumeStructuredMasterConfirm } from "./hooks/useResumeStructuredMasterConfirm";
import { ResumePreviewConfirm } from "./ResumePreviewConfirm";
import type { PreviewDraft } from "./ResumePreviewConfirm";
import {
  buildStructuredProfilePayload,
  mapParsedSummaryToStructuredProfileDraft,
} from "./adapters/mapParsedSummaryToStructuredProfileDraft";

export interface PreviewStageProps {
  asset: ResumeAsset;
  draft: PreviewDraft;
  sourceLabel: string;
  onBack: () => void;
  onSaved: () => void;
}

export const PreviewStage: FC<PreviewStageProps> = ({
  asset,
  draft,
  sourceLabel,
  onBack,
  onSaved,
}) => {
  const { t, lang } = useI18n();
  const { navigate } = useNavigation();
  const { confirm } = useResumeStructuredMasterConfirm();
  const [submitting, setSubmitting] = useState(false);
  const [inlineError, setInlineError] = useState<string | null>(null);

  // Re-derive in case the asset changed; the prop draft is the cached version.
  const renderedDraft = draft.name
    ? draft
    : mapParsedSummaryToStructuredProfileDraft(asset);

  const handleConfirm = async () => {
    if (submitting) return;
    setSubmitting(true);
    setInlineError(null);
    const displayName = renderedDraft.name || asset.title;
    const structuredProfile = buildStructuredProfilePayload(asset);
    const outcome = await confirm({
      resumeAssetId: asset.id,
      body: { displayName, language: lang, structuredProfile },
    });
    setSubmitting(false);
    if (outcome.kind === "saved") {
      fireResumeWorkshopToast(t("resumeWorkshop.preview.success"), "ok");
      onSaved();
      navigate({ name: "resume_versions", params: {} });
      return;
    }
    if (outcome.kind === "already_exists") {
      fireResumeWorkshopToast(
        t("resumeWorkshop.preview.alreadyExists"),
        "warn",
      );
      if (outcome.existingMasterId) {
        navigate({
          name: "resume_versions",
          params: { versionId: outcome.existingMasterId, tab: "preview" },
        });
      } else {
        navigate({ name: "resume_versions", params: {} });
      }
      onSaved();
      return;
    }
    if (outcome.kind === "validation") {
      setInlineError(t("resumeWorkshop.create.errors.validation"));
      return;
    }
    setInlineError(t("resumeWorkshop.create.errors.confirmFailed"));
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
