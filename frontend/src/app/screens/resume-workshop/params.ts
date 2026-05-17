export type ResumeWorkshopFlow = "list" | "create" | "branch";
export type ResumeDetailTab = "preview" | "rewrites" | "edit";
export type ResumeCreateMode = "upload" | "paste" | "guided";

export interface ResumeWorkshopParams {
  flow: ResumeWorkshopFlow;
  versionId: string | null;
  tab: ResumeDetailTab | null;
  branchOriginalId: string | null;
  createMode: ResumeCreateMode | null;
}

const DETAIL_TABS: readonly ResumeDetailTab[] = [
  "preview",
  "rewrites",
  "edit",
] as const;

const CREATE_MODES: readonly ResumeCreateMode[] = [
  "upload",
  "paste",
  "guided",
] as const;

const isDetailTab = (value: string | undefined): value is ResumeDetailTab =>
  typeof value === "string" && (DETAIL_TABS as readonly string[]).includes(value);

const isCreateMode = (
  value: string | undefined,
): value is ResumeCreateMode =>
  typeof value === "string" &&
  (CREATE_MODES as readonly string[]).includes(value);

const parseFlow = (value: string | undefined): ResumeWorkshopFlow => {
  if (value === "create") return "create";
  if (value === "branch") return "branch";
  return "list";
};

export const parseResumeWorkshopParams = (
  routeParams: Record<string, string>,
): ResumeWorkshopParams => {
  const versionId = routeParams.versionId ? routeParams.versionId : null;
  const branchOriginalId = routeParams.branchOriginalId
    ? routeParams.branchOriginalId
    : null;
  const tab = isDetailTab(routeParams.tab) ? routeParams.tab : null;
  const createMode = isCreateMode(routeParams.createMode)
    ? routeParams.createMode
    : null;
  return {
    flow: parseFlow(routeParams.flow),
    versionId,
    tab,
    branchOriginalId,
    createMode,
  };
};
