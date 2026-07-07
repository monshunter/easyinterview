import { useCallback } from "react";

import type {
  RegisterResumeRequest,
  ResumeWithJob,
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

// D-20: guided Q&A intake is outside current scope. Resume creation is upload or paste only.
export type RegisterInput = RegisterUploadInput | RegisterPasteInput;

export interface UseResumeRegistrationResult {
  register: (input: RegisterInput) => Promise<ResumeWithJob>;
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
  return {
    sourceType: "paste",
    rawText: input.rawText,
    title: input.title,
    language: input.language,
  };
}

export function useResumeRegistration(): UseResumeRegistrationResult {
  const runtime = useAppRuntimeOptional();

  const register = useCallback(
    async (input: RegisterInput): Promise<ResumeWithJob> => {
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
