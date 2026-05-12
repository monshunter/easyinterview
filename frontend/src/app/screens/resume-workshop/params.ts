export type ResumeWorkshopFlow = "list" | "create" | "branch";
export type ResumeDetailTab = "preview" | "rewrites" | "edit";

export interface ResumeWorkshopParams {
  flow: ResumeWorkshopFlow;
  versionId: string | null;
  tab: ResumeDetailTab | null;
  branchOriginalId: string | null;
}

const DETAIL_TABS: readonly ResumeDetailTab[] = [
  "preview",
  "rewrites",
  "edit",
] as const;

const isDetailTab = (value: string | undefined): value is ResumeDetailTab =>
  typeof value === "string" && (DETAIL_TABS as readonly string[]).includes(value);

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
  return {
    flow: parseFlow(routeParams.flow),
    versionId,
    tab,
    branchOriginalId,
  };
};
