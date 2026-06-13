import { useCallback, type FC } from "react";

import { useAppRuntimeOptional } from "../../../runtime/AppRuntimeProvider";
import { useI18n } from "../../../i18n/messages";
import { DebriefContextPickerModal, type PickerOption } from "./DebriefContextPickerModal";
import { usePickerOptions } from "../hooks/usePickerOptions";
import type { Resume } from "../types";

interface ResumePickerProps {
  selectedResumeId: string | null;
  onClose: () => void;
  onConfirm: (selection: { resume: Resume | null }) => void;
}

/**
 * D-20 flat resume picker. A single flat list of resumes from `listResumes`;
 * only `parseStatus === 'ready'` active resumes are listed per spec §3.2. The
 * earlier asset→version two-step selection is removed with the version tree.
 */
export const ResumePicker: FC<ResumePickerProps> = ({
  selectedResumeId,
  onClose,
  onConfirm,
}) => {
  const runtime = useAppRuntimeOptional();
  const { t } = useI18n();

  const loadResumes = useCallback(async () => {
    if (!runtime) return { options: [] as PickerOption<Resume>[] };
    const res = await runtime.client.listResumes();
    const items = res.items.filter(
      (resume: Resume) =>
        (resume.status === undefined || resume.status === "active") &&
        resume.parseStatus === "ready",
    );
    const options = items.map<PickerOption<Resume>>((resume) => ({
      id: resume.id,
      title: resume.displayName || resume.title,
      meta: `${resume.language}`,
      value: resume,
    }));
    return { options };
  }, [runtime]);

  const resumeState = usePickerOptions<Resume>({
    enabled: Boolean(runtime),
    load: loadResumes,
  });

  return (
    <DebriefContextPickerModal
      kind="resume"
      options={resumeState.options}
      selectedId={selectedResumeId}
      loading={resumeState.loading}
      errorMessage={resumeState.error}
      banner={
        <span data-testid="debrief-picker-banner-resume-phase">
          {t("debrief.picker.resume.assetPhase")}
        </span>
      }
      onClose={onClose}
      onConfirm={(opt) => {
        onConfirm({ resume: opt ? opt.value : null });
      }}
    />
  );
};
