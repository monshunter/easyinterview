export type ResumeWorkshopFlow = "list" | "create";
export type ResumeCreateMode = "upload" | "paste";

export interface ResumeWorkshopParams {
  flow: ResumeWorkshopFlow;
  resumeId: string | null;
  targetJobId: string | null;
  createMode: ResumeCreateMode | null;
}

const CREATE_MODES: readonly ResumeCreateMode[] = ["upload", "paste"] as const;

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
  const createMode = isCreateMode(routeParams.createMode)
    ? routeParams.createMode
    : null;
  return {
    flow: parseFlow(routeParams.flow),
    resumeId,
    targetJobId,
    createMode,
  };
};
