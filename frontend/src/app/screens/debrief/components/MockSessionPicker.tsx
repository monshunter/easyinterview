import { useCallback, useMemo, type FC } from "react";

import { useAppRuntimeOptional } from "../../../runtime/AppRuntimeProvider";
import { useI18n } from "../../../i18n/messages";
import { DebriefContextPickerModal, type PickerOption } from "./DebriefContextPickerModal";
import { usePickerOptions } from "../hooks/usePickerOptions";
import type { PracticeSession } from "../types";

interface MockSessionPickerProps {
  /** Required to scope the listing. When null, the picker shows an explicit missing-context state. */
  targetJobId: string | null;
  selectedId: string | null;
  onClose: () => void;
  onConfirm: (session: PracticeSession | null) => void;
}

/**
 * Phase 2.3 Mock Session picker. Loads completed sessions via
 * `listPracticeSessions({targetJobId, status:'completed'})`. If the server
 * does not honour the `status` filter, the hook reports a `client-side-status-filter`
 * fallback and we trim the list locally.
 */
export const MockSessionPicker: FC<MockSessionPickerProps> = ({
  targetJobId,
  selectedId,
  onClose,
  onConfirm,
}) => {
  const runtime = useAppRuntimeOptional();
  const { t } = useI18n();

  const load = useCallback(async () => {
    if (!runtime || !targetJobId) {
      return { options: [] as PickerOption<PracticeSession>[] };
    }
    const res = await runtime.client.listPracticeSessions({
      query: { targetJobId, status: "completed" },
    });
    const allCompleted = res.items.every((s) => s.status === "completed");
    const items = allCompleted
      ? res.items
      : res.items.filter((s) => s.status === "completed");
    const options = items.map<PickerOption<PracticeSession>>((session) => ({
      id: session.id,
      title: t("debrief.picker.mockSession.itemTitle").replace(
        "{id}",
        session.id.slice(-6),
      ),
      meta: `${session.language} · ${session.turnCount} ${t(
        "debrief.picker.mockSession.turns",
      )}`,
      value: session,
    }));
    return {
      options,
      fallback: allCompleted ? undefined : ("client-side-status-filter" as const),
    };
  }, [runtime, t, targetJobId]);

  const enabled = useMemo(
    () => Boolean(runtime && targetJobId),
    [runtime, targetJobId],
  );
  const state = usePickerOptions<PracticeSession>({ enabled, load });

  if (!targetJobId) {
    return (
      <DebriefContextPickerModal
        kind="mockSession"
        options={[]}
        selectedId={null}
        emptyCopy={t("debrief.picker.mockSession.needTargetJob")}
        allowEmpty
        noneOptionCopy={t("debrief.picker.mockSession.none")}
        onClose={onClose}
        onConfirm={() => onConfirm(null)}
      />
    );
  }

  return (
    <DebriefContextPickerModal
      kind="mockSession"
      options={state.options}
      selectedId={selectedId}
      loading={state.loading}
      errorMessage={state.error}
      allowEmpty
      noneOptionCopy={t("debrief.picker.mockSession.none")}
      banner={
        state.fallback === "client-side-status-filter"
          ? (
            <span data-testid="debrief-picker-banner-filter-fallback">
              {t("debrief.picker.mockSession.filterFallback")}
            </span>
          )
          : null
      }
      onClose={onClose}
      onConfirm={(opt) => onConfirm(opt?.value ?? null)}
    />
  );
};
