export type ResumeWorkshopFlow = "list" | "create";
export type ResumeDetailTab = "preview" | "rewrites" | "edit";
export type ResumeCreateMode = "upload" | "paste";

export interface ResumeWorkshopParams {
  flow: ResumeWorkshopFlow;
  resumeId: string | null;
  tab: ResumeDetailTab | null;
  targetJobId: string | null;
  createMode: ResumeCreateMode | null;
  /**
   * Optional tailor run id carried from the ai_select nav so the Rewrites Tab
   * picks up an in-flight tailor run on first paint and starts polling without
   * a manual rerun. Plan 003 Phase 5.
   */
  tailorRunId: string | null;
}

const DETAIL_TABS: readonly ResumeDetailTab[] = [
  "preview",
  "rewrites",
  "edit",
] as const;

const CREATE_MODES: readonly ResumeCreateMode[] = ["upload", "paste"] as const;

const isDetailTab = (value: string | undefined): value is ResumeDetailTab =>
  typeof value === "string" && (DETAIL_TABS as readonly string[]).includes(value);

const isCreateMode = (
  value: string | undefined,
): value is ResumeCreateMode =>
  typeof value === "string" &&
  (CREATE_MODES as readonly string[]).includes(value);

const parseFlow = (value: string | undefined): ResumeWorkshopFlow => {
  if (value === "create") return "create";
  return "list";
};

export const parseResumeWorkshopParams = (
  routeParams: Record<string, string>,
): ResumeWorkshopParams => {
  const resumeId = routeParams.resumeId ? routeParams.resumeId : null;
  const targetJobId = routeParams.targetJobId ? routeParams.targetJobId : null;
  const tab = isDetailTab(routeParams.tab) ? routeParams.tab : null;
  const createMode = isCreateMode(routeParams.createMode)
    ? routeParams.createMode
    : null;
  const tailorRunId = routeParams.tailorRunId
    ? routeParams.tailorRunId
    : null;
  return {
    flow: parseFlow(routeParams.flow),
    resumeId,
    tab,
    targetJobId,
    createMode,
    tailorRunId,
  };
};
