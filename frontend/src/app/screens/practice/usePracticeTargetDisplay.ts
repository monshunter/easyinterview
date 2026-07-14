import { useEffect, useRef, useState } from "react";

import type { EasyInterviewClient } from "../../../api/generated/client";
import type {
  PracticeSession,
  TargetJob,
} from "../../../api/generated/types";
import { useAppRuntimeOptional } from "../../runtime/AppRuntimeProvider";

export interface UsePracticeTargetDisplayOptions {
  /** Loaded server session. While null, route/context may provide a temporary ID. */
  session: Pick<PracticeSession, "targetJobId"> | null;
  routeTargetJobId?: string | null;
  contextTargetJobId?: string | null;
}

export interface PracticeTargetDisplay {
  targetJobId: string | null;
  companyName: string | null;
  title: string | null;
  loading: boolean;
  error: Error | null;
}

interface TargetDisplaySnapshot {
  client: EasyInterviewClient | null;
  targetJobId: string | null;
  companyName: string | null;
  title: string | null;
  loading: boolean;
  error: Error | null;
}

/**
 * Resolves the real TargetJob display identity for Practice.
 *
 * A loaded server session is authoritative. Route/context IDs are only a
 * bootstrap source while that session is absent. Every lookup goes through
 * the generated `getTargetJob` operation. Mount reads stay signal-free so the
 * generated client can share StrictMode's duplicate setup; cleanup and the
 * sequence guard reject responses that are no longer current.
 */
export function usePracticeTargetDisplay(
  options: UsePracticeTargetDisplayOptions,
): PracticeTargetDisplay {
  const runtime = useAppRuntimeOptional();
  const client = runtime?.client ?? null;
  const targetJobId = selectTargetJobId(options);
  const requestSeqRef = useRef(0);
  const [snapshot, setSnapshot] = useState<TargetDisplaySnapshot>(() =>
    emptySnapshot(client, targetJobId, Boolean(client && targetJobId)),
  );

  useEffect(() => {
    const seq = requestSeqRef.current + 1;
    requestSeqRef.current = seq;

    if (!client || !targetJobId) {
      setSnapshot(emptySnapshot(client, targetJobId, false));
      return;
    }

    let active = true;
    setSnapshot(emptySnapshot(client, targetJobId, true));

    client
      .getTargetJob(targetJobId)
      .then((job: TargetJob) => {
        if (!active || requestSeqRef.current !== seq) return;
        setSnapshot({
          client,
          targetJobId,
          companyName: job.companyName,
          title: job.title,
          loading: false,
          error: null,
        });
      })
      .catch((error: unknown) => {
        if (!active || requestSeqRef.current !== seq) return;
        setSnapshot({
          ...emptySnapshot(client, targetJobId, false),
          error: wrapError(error),
        });
      });

    return () => {
      active = false;
    };
  }, [client, targetJobId]);

  if (snapshot.client !== client || snapshot.targetJobId !== targetJobId) {
    return {
      targetJobId,
      companyName: null,
      title: null,
      loading: Boolean(client && targetJobId),
      error: null,
    };
  }

  return {
    targetJobId,
    companyName: snapshot.companyName,
    title: snapshot.title,
    loading: snapshot.loading,
    error: snapshot.error,
  };
}

function selectTargetJobId({
  session,
  routeTargetJobId,
  contextTargetJobId,
}: UsePracticeTargetDisplayOptions): string | null {
  if (session !== null) return nonEmpty(session.targetJobId);
  return nonEmpty(routeTargetJobId) ?? nonEmpty(contextTargetJobId);
}

function nonEmpty(value: string | null | undefined): string | null {
  const normalized = value?.trim();
  return normalized ? normalized : null;
}

function emptySnapshot(
  client: EasyInterviewClient | null,
  targetJobId: string | null,
  loading: boolean,
): TargetDisplaySnapshot {
  return {
    client,
    targetJobId,
    companyName: null,
    title: null,
    loading,
    error: null,
  };
}

function wrapError(error: unknown): Error {
  return error instanceof Error ? error : new Error(String(error));
}
