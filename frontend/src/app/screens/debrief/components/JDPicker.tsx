import { useCallback, useMemo, type FC } from "react";

import { useAppRuntimeOptional } from "../../../runtime/AppRuntimeProvider";
import { DebriefContextPickerModal, type PickerOption } from "./DebriefContextPickerModal";
import { usePickerOptions } from "../hooks/usePickerOptions";
import type { TargetJob } from "../types";

interface JDPickerProps {
  selectedId: string | null;
  onClose: () => void;
  onConfirm: (targetJob: TargetJob | null) => void;
}

/**
 * Phase 2.2 JD picker. Loads target jobs via
 * `listTargetJobs({analysisStatus:'ready'})` — only parsed / ready jobs are
 * valid debrief contexts per
 * docs/spec/frontend-debrief/spec.md §3.2.
 */
export const JDPicker: FC<JDPickerProps> = ({
  selectedId,
  onClose,
  onConfirm,
}) => {
  const runtime = useAppRuntimeOptional();
  const load = useCallback(async () => {
    if (!runtime) return { options: [] as PickerOption<TargetJob>[] };
    const res = await runtime.client.listTargetJobs({
      query: { analysisStatus: "ready" },
    });
    const options = res.items.map<PickerOption<TargetJob>>((job) => ({
      id: job.id,
      title: [job.companyName, job.title].filter(Boolean).join(" · "),
      meta: [job.locationText, job.targetLanguage]
        .filter(Boolean)
        .join(" · "),
      note: job.summary?.coreThemes?.[0],
      value: job,
    }));
    return { options };
  }, [runtime]);

  const enabled = useMemo(() => Boolean(runtime), [runtime]);
  const state = usePickerOptions<TargetJob>({ enabled, load });

  return (
    <DebriefContextPickerModal
      kind="targetJob"
      options={state.options}
      selectedId={selectedId}
      loading={state.loading}
      errorMessage={state.error}
      onClose={onClose}
      onConfirm={(opt) => onConfirm(opt?.value ?? null)}
    />
  );
};
