import type { FC } from "react";

import { useAppRuntimeOptional } from "../../runtime/AppRuntimeProvider";
import type { Route } from "../../routes";
import { ResumeBranchFlow } from "./branch/ResumeBranchFlow";
import { NotImplementedPlaceholder } from "./components/NotImplementedPlaceholder";
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
  if (params.branchOriginalId) {
    rootDataAttributes["data-branch-original-id"] = params.branchOriginalId;
  }
  if (params.createMode) {
    rootDataAttributes["data-create-mode"] = params.createMode;
  }

  let body;
  if (isAuthGated) {
    body = <ResumeWorkshopAuthGate params={params} />;
  } else if (params.flow === "create") {
    body = <ResumeCreateFlow initialMode={params.createMode ?? undefined} />;
  } else if (params.flow === "branch") {
    body = <ResumeBranchFlow branchOriginalId={params.branchOriginalId} />;
  } else if (params.versionId) {
    body = (
      <DetailWrapper
        versionId={params.versionId}
        tab={params.tab}
        tailorRunId={params.tailorRunId}
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
  versionId: string;
  tab: import("./params").ResumeDetailTab | null;
  tailorRunId: string | null;
}

const DetailWrapper: FC<DetailWrapperProps> = ({
  versionId,
  tab,
  tailorRunId,
}) => {
  const attrs: Record<string, string> = {
    "data-testid": "resume-workshop-detail",
    "data-resume-version-id": versionId,
  };
  if (tab) {
    attrs["data-tab"] = tab;
  }
  if (tailorRunId) {
    attrs["data-tailor-run-id"] = tailorRunId;
  }
  return (
    <div {...attrs}>
      <ResumeDetailView
        versionId={versionId}
        initialTab={tab}
        initialTailorRunId={tailorRunId}
      />
    </div>
  );
};
