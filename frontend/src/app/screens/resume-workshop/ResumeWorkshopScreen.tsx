import type { FC } from "react";

import { useAppRuntimeOptional } from "../../runtime/AppRuntimeProvider";
import type { Route } from "../../routes";
import { ResumeDetailView } from "./components/ResumeDetailView";
import { ResumeListView } from "./components/ResumeListView";
import { ResumeWorkshopAuthGate } from "./components/ResumeWorkshopAuthGate";
import { ResumeCreateFlow } from "./create/ResumeCreateFlow";
import { parseResumeWorkshopParams } from "./params";

interface ResumeWorkshopScreenProps {
  route: Route;
}

export const ResumeWorkshopScreen: FC<ResumeWorkshopScreenProps> = ({
  route,
}) => {
  const params = parseResumeWorkshopParams(route.params);
  const runtime = useAppRuntimeOptional();
  // Production wires the runtime so `auth.status` reflects `/me`. The screen
  // gates protected Resume APIs strictly on `authenticated`, but routing-only
  // tests (no runtime mounted) bypass the gate so flow / detail dispatch can
  // be verified in isolation per Phase 1.5 contract.
  const isAuthGated = runtime !== null && runtime.auth.status !== "authenticated";

  const rootDataAttributes: Record<string, string> = {
    "data-testid": "resume-workshop-screen",
    "data-route-name": route.name,
    "data-flow": params.flow,
    "data-auth-status": runtime?.auth.status ?? "unmounted",
  };
  if (params.targetJobId) {
    rootDataAttributes["data-target-job-id"] = params.targetJobId;
  }
  if (params.createMode) {
    rootDataAttributes["data-create-mode"] = params.createMode;
  }

  let body;
  if (isAuthGated) {
    body = <ResumeWorkshopAuthGate params={params} />;
  } else if (params.flow === "create") {
    body = <ResumeCreateFlow initialMode={params.createMode ?? undefined} />;
  } else if (params.resumeId) {
    body = (
      <DetailWrapper
        resumeId={params.resumeId}
        tab={params.tab}
        tailorRunId={params.tailorRunId}
        targetJobId={params.targetJobId}
      />
    );
  } else {
    body = <ResumeListView />;
  }

  return (
    <section {...rootDataAttributes} className="ei-screen-shell">
      {body}
    </section>
  );
};

interface DetailWrapperProps {
  resumeId: string;
  tab: import("./params").ResumeDetailTab | null;
  tailorRunId: string | null;
  targetJobId: string | null;
}

const DetailWrapper: FC<DetailWrapperProps> = ({
  resumeId,
  tab,
  tailorRunId,
  targetJobId,
}) => {
  const attrs: Record<string, string> = {
    "data-testid": "resume-workshop-detail",
    "data-resume-id": resumeId,
  };
  if (tab) {
    attrs["data-tab"] = tab;
  }
  if (tailorRunId) {
    attrs["data-tailor-run-id"] = tailorRunId;
  }
  if (targetJobId) {
    attrs["data-target-job-id"] = targetJobId;
  }
  return (
    <div {...attrs}>
      <ResumeDetailView
        resumeId={resumeId}
        initialTab={tab}
        initialTailorRunId={tailorRunId}
        targetJobId={targetJobId}
      />
    </div>
  );
};
