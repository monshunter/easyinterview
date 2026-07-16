import {
  useCallback,
  useEffect,
  useRef,
  useState,
  type Dispatch,
  type SetStateAction,
} from "react";

import {
  ApiClientError,
  type EasyInterviewClient,
} from "../../../../api/generated/client";
import type { ReportWithJob } from "../../../../api/generated/types";
import { generateIdempotencyKey } from "../../../../lib/conventions/idempotency";
import { JOB_TYPE_REPORT_GENERATE } from "../../../../lib/jobs/jobs";

export interface ReportRegenerationRowState {
  error: boolean;
  pending: boolean;
}

interface UseReportRegenerationOptions {
  client: EasyInterviewClient | null;
  onAccepted: (reportId: string) => void;
  onStaleState: () => void;
  targetJobId: string;
}

interface RegenerationScope {
  client: EasyInterviewClient;
  idempotencyKey: string | null;
  inFlight: Promise<void> | null;
  reportId: string;
  targetJobId: string;
}

const IDLE_ROW: ReportRegenerationRowState = {
  error: false,
  pending: false,
};

export function useReportRegeneration({
  client,
  onAccepted,
  onStaleState,
  targetJobId,
}: UseReportRegenerationOptions): {
  regenerate: (reportId: string) => Promise<void>;
  stateFor: (reportId: string) => ReportRegenerationRowState;
} {
  const scopesRef = useRef(new Map<string, RegenerationScope>());
  const ownerRef = useRef({ client, targetJobId });
  ownerRef.current = { client, targetJobId };

  const onAcceptedRef = useRef(onAccepted);
  const onStaleStateRef = useRef(onStaleState);
  onAcceptedRef.current = onAccepted;
  onStaleStateRef.current = onStaleState;

  const [rowStates, setRowStates] = useState<
    Record<string, ReportRegenerationRowState>
  >({});

  useEffect(() => {
    const ownedScopes = scopesRef.current;
    return () => {
      for (const [key, scope] of ownedScopes) {
        if (scope.client === client && scope.targetJobId === targetJobId) {
          ownedScopes.delete(key);
        }
      }
    };
  }, [client, targetJobId]);

  const stateFor = useCallback(
    (reportId: string): ReportRegenerationRowState =>
      rowStates[scopeKey(targetJobId, reportId)] ?? IDLE_ROW,
    [rowStates, targetJobId],
  );

  const regenerate = useCallback(
    async (reportId: string): Promise<void> => {
      if (!client || !targetJobId || !reportId) return;

      const key = scopeKey(targetJobId, reportId);
      let scope = scopesRef.current.get(key);
      if (
        !scope ||
        scope.client !== client ||
        scope.targetJobId !== targetJobId ||
        scope.reportId !== reportId
      ) {
        scope = {
          client,
          idempotencyKey: null,
          inFlight: null,
          reportId,
          targetJobId,
        };
        scopesRef.current.set(key, scope);
      }
      if (scope.inFlight) return scope.inFlight;
      if (!scope.idempotencyKey) {
        scope.idempotencyKey = generateIdempotencyKey();
      }

      updateRow(setRowStates, key, { error: false, pending: true });
      const request = client
        .regenerateFeedbackReport(reportId, {
          idempotencyKey: scope.idempotencyKey,
        })
        .then((result) => {
          scope.inFlight = null;
          if (!isCurrentScope(scopesRef.current, ownerRef.current, key, scope)) {
            return;
          }
          if (!isMatchingQueuedRegeneration(result, reportId)) {
            updateRow(setRowStates, key, { error: true, pending: false });
            return;
          }

          scope.idempotencyKey = null;
          updateRow(setRowStates, key, IDLE_ROW);
          onAcceptedRef.current(reportId);
        })
        .catch((error: unknown) => {
          scope.inFlight = null;
          if (!isCurrentScope(scopesRef.current, ownerRef.current, key, scope)) {
            return;
          }

          if (isExplicitClientRejection(error)) {
            scope.idempotencyKey = null;
          }
          updateRow(setRowStates, key, { error: true, pending: false });
          if (isStaleStateRejection(error)) {
            onStaleStateRef.current();
          }
        });
      scope.inFlight = request;
      return request;
    },
    [client, targetJobId],
  );

  return { regenerate, stateFor };
}

function scopeKey(targetJobId: string, reportId: string): string {
  return `${targetJobId}:${reportId}`;
}

function updateRow(
  setRows: Dispatch<SetStateAction<Record<string, ReportRegenerationRowState>>>,
  key: string,
  state: ReportRegenerationRowState,
): void {
  setRows((current) => ({ ...current, [key]: state }));
}

function isCurrentScope(
  scopes: Map<string, RegenerationScope>,
  owner: { client: EasyInterviewClient | null; targetJobId: string },
  key: string,
  scope: RegenerationScope,
): boolean {
  return Boolean(
    owner.client === scope.client &&
      owner.targetJobId === scope.targetJobId &&
      scopes.get(key) === scope,
  );
}

function isMatchingQueuedRegeneration(
  result: ReportWithJob,
  expectedReportId: string,
): boolean {
  return Boolean(
    result &&
      result.reportId === expectedReportId &&
      result.job &&
      result.job.jobType === JOB_TYPE_REPORT_GENERATE &&
      result.job.status === "queued" &&
      result.job.resourceType === "feedback_report" &&
      result.job.resourceId === expectedReportId,
  );
}

function isExplicitClientRejection(error: unknown): boolean {
  return Boolean(
    error instanceof ApiClientError &&
      error.kind === "http" &&
      error.status !== null &&
      error.status >= 400 &&
      error.status < 500,
  );
}

function isStaleStateRejection(error: unknown): boolean {
  if (!(error instanceof ApiClientError)) return false;
  const code = error.apiError?.error.code;
  return (
    code === "REPORT_INVALID_STATE_TRANSITION" || code === "REPORT_NOT_READY"
  );
}
