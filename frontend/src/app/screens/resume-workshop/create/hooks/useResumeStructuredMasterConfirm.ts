import { useCallback, useRef } from "react";

import type {
  ConfirmResumeStructuredMasterRequest,
  ResumeVersion,
} from "../../../../../api/generated/types";
import { useAppRuntimeOptional } from "../../../../runtime/AppRuntimeProvider";
import { generateIdempotencyKey } from "../../../../../lib/conventions/idempotency";

export type ConfirmOutcome =
  | { kind: "saved"; version: ResumeVersion }
  | { kind: "already_exists"; existingMasterId: string | null }
  | { kind: "validation"; details?: Record<string, unknown> }
  | { kind: "error"; message: string };

export interface ConfirmStructuredMasterInput {
  resumeAssetId: string;
  body: ConfirmResumeStructuredMasterRequest;
}

export interface UseResumeStructuredMasterConfirmResult {
  confirm: (input: ConfirmStructuredMasterInput) => Promise<ConfirmOutcome>;
}

interface ConfirmCacheEntry {
  resumeAssetId: string;
  idempotencyKey: string;
}

export function useResumeStructuredMasterConfirm(): UseResumeStructuredMasterConfirmResult {
  const runtime = useAppRuntimeOptional();
  const cacheRef = useRef<ConfirmCacheEntry | null>(null);

  const confirm = useCallback(
    async (input: ConfirmStructuredMasterInput): Promise<ConfirmOutcome> => {
      if (!runtime) {
        return { kind: "error", message: "CONFIRM_RUNTIME_UNAVAILABLE" };
      }
      const { client } = runtime;
      const cached = cacheRef.current;
      const idempotencyKey =
        cached && cached.resumeAssetId === input.resumeAssetId
          ? cached.idempotencyKey
          : generateIdempotencyKey();
      cacheRef.current = {
        resumeAssetId: input.resumeAssetId,
        idempotencyKey,
      };
      try {
        const version = await client.confirmResumeStructuredMaster(
          input.resumeAssetId,
          input.body,
          {
            idempotencyKey,
            headers: input.body.language
              ? { "Accept-Language": input.body.language }
              : undefined,
          },
        );
        return { kind: "saved", version };
      } catch (error) {
        const parsed = parseConfirmError(error);
        if (parsed.kind === "already_exists") {
          // Force a new IK for the next user-initiated retry attempt.
          cacheRef.current = null;
          try {
            const list = await client.listResumeVersions(input.resumeAssetId, {
              headers: input.body.language
              ? { "Accept-Language": input.body.language }
              : undefined,
            });
            const master = list.items.find(
              (item) => item.versionType === "structured_master",
            );
            return {
              kind: "already_exists",
              existingMasterId: master ? master.id : null,
            };
          } catch {
            return { kind: "already_exists", existingMasterId: null };
          }
        }
        if (parsed.kind === "validation") {
          cacheRef.current = null;
          return parsed;
        }
        cacheRef.current = null;
        return parsed;
      }
    },
    [runtime],
  );

  return { confirm };
}

interface ParsedError {
  status?: number;
  code?: string;
  message: string;
  details?: Record<string, unknown>;
}

function extractErrorMetadata(error: unknown): ParsedError {
  if (error instanceof Error) {
    const anyErr = error as Error & {
      status?: number;
      code?: string;
      body?: { code?: string; details?: Record<string, unknown> };
    };
    return {
      status: anyErr.status,
      code: anyErr.code ?? anyErr.body?.code,
      message: error.message,
      details: anyErr.body?.details,
    };
  }
  return { message: String(error) };
}

export function parseConfirmError(error: unknown): ConfirmOutcome {
  const meta = extractErrorMetadata(error);
  const code = meta.code ?? "";
  if (
    meta.status === 409 ||
    code === "RESUME_STRUCTURED_MASTER_ALREADY_EXISTS" ||
    /RESUME_STRUCTURED_MASTER_ALREADY_EXISTS/i.test(meta.message) ||
    /\b409\b/.test(meta.message)
  ) {
    return { kind: "already_exists", existingMasterId: null };
  }
  if (
    meta.status === 422 ||
    code === "VALIDATION_FAILED" ||
    /VALIDATION_FAILED/i.test(meta.message) ||
    /\b422\b/.test(meta.message)
  ) {
    return { kind: "validation", details: meta.details };
  }
  return { kind: "error", message: meta.message };
}
