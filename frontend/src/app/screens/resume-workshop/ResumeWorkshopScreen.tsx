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
  // Production wires the runtime so `auth.status` reflects `/me`. When this
  // component is mounted without runtime context, the caller is responsible for
  // applying the auth gate; isolated route tests use that mode for dispatch
  // coverage.
  const isAuthGated = runtime !== null && runtime.auth.status !== "authenticated";

  const rootDataAttributes: Record<string, string> = {
    "data-testid": "resume-workshop-screen",
    "data-route-name": route.name,
    "data-flow": params.resumeId ? "detail" : params.flow,
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
        targetJobId={params.targetJobId}
      />
    );
  } else {
    body = <ResumeListView />;
  }

  return (
    <section
      {...rootDataAttributes}
      className="ei-screen-shell ei-resume-workshop-screen"
    >
      {body}
    </section>
  );
};

interface DetailWrapperProps {
  resumeId: string;
  targetJobId: string | null;
}

const DetailWrapper: FC<DetailWrapperProps> = ({
  resumeId,
  targetJobId,
}) => {
  const attrs: Record<string, string> = {
    "data-testid": "resume-workshop-detail",
    "data-resume-id": resumeId,
  };
  if (targetJobId) {
    attrs["data-target-job-id"] = targetJobId;
  }
  return (
    <div {...attrs}>
      <ResumeDetailView resumeId={resumeId} />
    </div>
  );
};
