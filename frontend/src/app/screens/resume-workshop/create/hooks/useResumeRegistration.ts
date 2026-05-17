import { useCallback } from "react";

import type {
  RegisterResumeRequest,
  ResumeAssetWithJob,
} from "../../../../../api/generated/types";
import { useAppRuntimeOptional } from "../../../../runtime/AppRuntimeProvider";
import { generateIdempotencyKey } from "../../../../../lib/conventions/idempotency";

export type RegisterUploadInput = {
  sourceType: "upload";
  fileObjectId: string;
  title: string;
  language: string;
};

export type RegisterPasteInput = {
  sourceType: "paste";
  rawText: string;
  title: string;
  language: string;
};

export type RegisterGuidedInput = {
  sourceType: "guided";
  guidedAnswers: Record<string, string>;
  title: string;
  language: string;
};

export type RegisterInput =
  | RegisterUploadInput
  | RegisterPasteInput
  | RegisterGuidedInput;

export interface UseResumeRegistrationResult {
  register: (input: RegisterInput) => Promise<ResumeAssetWithJob>;
}

export function buildRegisterPayload(input: RegisterInput): RegisterResumeRequest {
  if (input.sourceType === "upload") {
    return {
      sourceType: "upload",
      fileObjectId: input.fileObjectId,
      title: input.title,
      language: input.language,
    };
  }
  if (input.sourceType === "paste") {
    return {
      sourceType: "paste",
      rawText: input.rawText,
      title: input.title,
      language: input.language,
    };
  }
  return {
    sourceType: "guided",
    guidedAnswers: input.guidedAnswers,
    title: input.title,
    language: input.language,
  };
}

export function useResumeRegistration(): UseResumeRegistrationResult {
  const runtime = useAppRuntimeOptional();

  const register = useCallback(
    async (input: RegisterInput): Promise<ResumeAssetWithJob> => {
      if (!runtime) {
        throw new Error("REGISTER_RUNTIME_UNAVAILABLE");
      }
      const idempotencyKey = generateIdempotencyKey();
      const payload = buildRegisterPayload(input);
      const response = await runtime.client.registerResume(payload, {
        idempotencyKey,
        headers: { "Accept-Language": input.language },
      });
      return response;
    },
    [runtime],
  );

  return { register };
}
