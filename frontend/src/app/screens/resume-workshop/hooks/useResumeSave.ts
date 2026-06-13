import { useCallback, useState } from "react";

import type {
  DuplicateResumeRequest,
  Resume,
  UpdateResumeRequest,
} from "../../../../api/generated/types";
import { generateIdempotencyKey } from "../../../../lib/conventions/idempotency";
import { useAppRuntimeOptional } from "../../../runtime/AppRuntimeProvider";

export type ResumeSaveErrorKind =
  | "validation"
  | "cross_user"
  | "idempotency_conflict"
  | "generic";

export class ResumeSaveError extends Error {
  readonly kind: ResumeSaveErrorKind;
  readonly status?: number;
  constructor(kind: ResumeSaveErrorKind, message: string, status?: number) {
    super(message);
    this.name = "ResumeSaveError";
    this.kind = kind;
    this.status = status;
  }
}

function parseError(error: unknown): ResumeSaveError {
  const message = error instanceof Error ? error.message : String(error);
  const statusMatch = /HTTP\s+(\d{3})/.exec(message);
  const status = statusMatch ? Number(statusMatch[1]) : undefined;
  if (status === 422) return new ResumeSaveError("validation", message, 422);
  if (status === 404) return new ResumeSaveError("cross_user", message, 404);
  if (status === 409)
    return new ResumeSaveError("idempotency_conflict", message, 409);
  return new ResumeSaveError("generic", message, status);
}

export interface UseResumeSaveResult {
  /** Overwrite this resume in place via PATCH /resumes/{resumeId}. */
  overwrite: (resumeId: string, body: UpdateResumeRequest) => Promise<Resume>;
  /** Save the accepted rewrites as a new resume via POST /resumes/{resumeId}/duplicate. */
  saveAsNew: (
    resumeId: string,
    body: DuplicateResumeRequest,
  ) => Promise<Resume>;
  pending: boolean;
  lastError: ResumeSaveError | null;
  resetError: () => void;
}

/**
 * D-20 flat save hook. Accepted rewrites + manual edits land through either
 * `updateResume` (overwrite this resume) or `duplicateResume` (save as a new
 * resume). Both carry an idempotency key per the resume-workshop contract.
 */
export function useResumeSave(): UseResumeSaveResult {
  const runtime = useAppRuntimeOptional();
  const client = runtime?.client ?? null;
  const [pending, setPending] = useState(false);
  const [lastError, setLastError] = useState<ResumeSaveError | null>(null);

  const resetError = useCallback(() => setLastError(null), []);

  const overwrite = useCallback(
    async (resumeId: string, body: UpdateResumeRequest): Promise<Resume> => {
      if (!client) {
        const err = new ResumeSaveError("generic", "runtime client not mounted");
        setLastError(err);
        throw err;
      }
      setPending(true);
      setLastError(null);
      try {
        return await client.updateResume(resumeId, body, {
          idempotencyKey: generateIdempotencyKey(),
        });
      } catch (rawErr) {
        const parsed = parseError(rawErr);
        setLastError(parsed);
        throw parsed;
      } finally {
        setPending(false);
      }
    },
    [client],
  );

  const saveAsNew = useCallback(
    async (resumeId: string, body: DuplicateResumeRequest): Promise<Resume> => {
      if (!client) {
        const err = new ResumeSaveError("generic", "runtime client not mounted");
        setLastError(err);
        throw err;
      }
      setPending(true);
      setLastError(null);
      try {
        return await client.duplicateResume(resumeId, body, {
          idempotencyKey: generateIdempotencyKey(),
        });
      } catch (rawErr) {
        const parsed = parseError(rawErr);
        setLastError(parsed);
        throw parsed;
      } finally {
        setPending(false);
      }
    },
    [client],
  );

  return { overwrite, saveAsNew, pending, lastError, resetError };
}
