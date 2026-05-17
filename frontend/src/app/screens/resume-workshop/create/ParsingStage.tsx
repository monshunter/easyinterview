import { useEffect, type FC } from "react";

import type { ResumeAsset } from "../../../../api/generated/types";
import {
  useResumeParsingPolling,
  type UseResumeParsingPollingOptions,
} from "./hooks/useResumeParsingPolling";
import { ResumeParseFlow, type ResumeParseState } from "./ResumeParseFlow";
import type { PreviewDraft } from "./ResumePreviewConfirm";
import { mapParsedSummaryToStructuredProfileDraft } from "./adapters/mapParsedSummaryToStructuredProfileDraft";

declare global {
  interface Window {
    __EI_RESUME_POLLING_OPTIONS__?: UseResumeParsingPollingOptions;
  }
}

export interface ParsingStageProps {
  resumeAssetId: string;
  sourceLabel: string;
  onReady: (asset: ResumeAsset, draft: PreviewDraft) => void;
  onCancel: () => void;
  /** Optional polling overrides for tests. */
  pollingOptions?: UseResumeParsingPollingOptions;
}

export const ParsingStage: FC<ParsingStageProps> = ({
  resumeAssetId,
  sourceLabel,
  onReady,
  onCancel,
  pollingOptions,
}) => {
  // Tests may install a fast polling cadence via a window-scoped seam.
  const testOverrides =
    typeof window !== "undefined"
      ? window.__EI_RESUME_POLLING_OPTIONS__
      : undefined;
  const { snapshot, retry, cancel } = useResumeParsingPolling(
    resumeAssetId,
    pollingOptions ?? testOverrides,
  );

  useEffect(() => {
    if (snapshot.status === "ready" && snapshot.asset) {
      const draft = mapParsedSummaryToStructuredProfileDraft(snapshot.asset);
      onReady(snapshot.asset, draft);
    }
  }, [snapshot, onReady]);

  let parseState: ResumeParseState;
  if (snapshot.status === "failed") {
    parseState = {
      phase: "failed",
      errorCode: snapshot.errorCode ?? "UNKNOWN",
    };
  } else if (snapshot.status === "ready") {
    parseState = { phase: "ready" };
  } else {
    parseState = { phase: "polling" };
  }

  return (
    <ResumeParseFlow
      sourceLabel={sourceLabel}
      parseState={parseState}
      onCancel={() => {
        cancel();
        onCancel();
      }}
      onRetry={retry}
    />
  );
};
