import { useEffect, useState } from "react";

import type { EasyInterviewClient } from "../../../../api/generated/client";
import { useAppRuntimeOptional } from "../../../runtime/AppRuntimeProvider";
import type { Route } from "../../../routes";
import { resolveReportBackRoute } from "../../report/reportBackRoute";

interface BackRouteOwner {
  client: EasyInterviewClient | null;
  enabled: boolean;
  reportId: string;
  resolution: FailedConversationBackRouteResolution;
}

export type FailedConversationBackRouteResolution =
  | { status: "resolving" }
  | { status: "resolved"; route: Route };

const workspaceRoute = (): Route => ({ name: "workspace", params: {} });
const resolvedWorkspace = (): FailedConversationBackRouteResolution => ({
  status: "resolved",
  route: workspaceRoute(),
});

function initialResolution(
  client: EasyInterviewClient | null,
  reportId: string,
  enabled: boolean,
): FailedConversationBackRouteResolution {
  return enabled && client && reportId
    ? { status: "resolving" }
    : resolvedWorkspace();
}

/** Resolves a failed transcript Back route from the report read, never route state. */
export function useFailedConversationBackRoute(
  reportId: string,
  enabled: boolean,
): FailedConversationBackRouteResolution {
  const runtime = useAppRuntimeOptional();
  const client = runtime?.client ?? null;
  const [owner, setOwner] = useState<BackRouteOwner>(() => ({
    client,
    enabled,
    reportId,
    resolution: initialResolution(client, reportId, enabled),
  }));

  useEffect(() => {
    if (!enabled || !client || !reportId) {
      setOwner({ client, enabled, reportId, resolution: resolvedWorkspace() });
      return;
    }

    setOwner({ client, enabled, reportId, resolution: { status: "resolving" } });

    let active = true;
    client
      .getFeedbackReport(reportId)
      .then((report) => {
        if (!active) return;
        setOwner({
          client,
          enabled,
          reportId,
          resolution: {
            status: "resolved",
            route: resolveReportBackRoute(report, reportId),
          },
        });
      })
      .catch(() => {
        if (!active) return;
        setOwner({ client, enabled, reportId, resolution: resolvedWorkspace() });
      });

    return () => {
      active = false;
    };
  }, [client, enabled, reportId]);

  return owner.client === client &&
    owner.enabled === enabled &&
    owner.reportId === reportId
    ? owner.resolution
    : initialResolution(client, reportId, enabled);
}
