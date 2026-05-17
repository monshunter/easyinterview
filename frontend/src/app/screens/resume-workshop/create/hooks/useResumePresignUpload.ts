import { useCallback, useRef } from "react";

import { useAppRuntimeOptional } from "../../../../runtime/AppRuntimeProvider";
import { generateIdempotencyKey } from "../../../../../lib/conventions/idempotency";

export interface PresignedUploadResult {
  fileObjectId: string;
  uploadUrl: string;
  method: "PUT" | "POST";
  headers: Record<string, string>;
  expiresAt: string;
}

export interface UploadFileOptions {
  contentType: string;
}

interface CachedPresign {
  idempotencyKey: string;
  fileSignature: string;
  result: PresignedUploadResult;
}

const TTL_GUARD_MS = 5_000; // refresh if less than 5s remains.

const fileSignature = (file: File): string =>
  `${file.name}|${file.size}|${file.lastModified}`;

export interface UseResumePresignUploadResult {
  uploadFile: (
    file: File,
    options: UploadFileOptions,
  ) => Promise<PresignedUploadResult>;
}

export function useResumePresignUpload(): UseResumePresignUploadResult {
  const runtime = useAppRuntimeOptional();
  const cacheRef = useRef<CachedPresign | null>(null);

  const obtainPresign = useCallback(
    async (
      file: File,
      options: UploadFileOptions,
    ): Promise<PresignedUploadResult> => {
      if (!runtime) {
        throw new Error("UPLOAD_RUNTIME_UNAVAILABLE");
      }
      const { client } = runtime;
      const signature = fileSignature(file);
      const cached = cacheRef.current;
      if (cached && cached.fileSignature === signature) {
        const expiresAtMs = Date.parse(cached.result.expiresAt);
        if (Number.isFinite(expiresAtMs) && expiresAtMs - Date.now() > TTL_GUARD_MS) {
          return cached.result;
        }
      }
      const idempotencyKey = generateIdempotencyKey();
      const presign = await client.createUploadPresign(
        {
          purpose: "resume",
          fileName: file.name,
          contentType: options.contentType,
          byteSize: file.size,
        },
        { idempotencyKey },
      );
      const result: PresignedUploadResult = {
        fileObjectId: presign.fileObjectId,
        uploadUrl: presign.uploadUrl,
        method: presign.method,
        headers: normalizeHeaders(presign.headers),
        expiresAt: presign.expiresAt,
      };
      cacheRef.current = {
        idempotencyKey,
        fileSignature: signature,
        result,
      };
      return result;
    },
    [runtime],
  );

  const uploadBinary = useCallback(
    async (
      file: File,
      presign: PresignedUploadResult,
    ): Promise<void> => {
      const response = await fetch(presign.uploadUrl, {
        method: presign.method,
        headers: presign.headers,
        body: file,
        mode: "cors",
      });
      if (!response.ok) {
        throw new Error(`UPLOAD_PUT_FAILED:${response.status}`);
      }
    },
    [],
  );

  const uploadFile = useCallback(
    async (
      file: File,
      options: UploadFileOptions,
    ): Promise<PresignedUploadResult> => {
      let presign = await obtainPresign(file, options);
      try {
        await uploadBinary(file, presign);
      } catch (error) {
        // If TTL just expired, attempt one re-presign.
        const expiresAtMs = Date.parse(presign.expiresAt);
        if (Number.isFinite(expiresAtMs) && expiresAtMs <= Date.now()) {
          cacheRef.current = null;
          presign = await obtainPresign(file, options);
          await uploadBinary(file, presign);
        } else {
          throw error;
        }
      }
      return presign;
    },
    [obtainPresign, uploadBinary],
  );

  return { uploadFile };
}

function normalizeHeaders(value: Record<string, unknown>): Record<string, string> {
  const out: Record<string, string> = {};
  for (const [key, raw] of Object.entries(value)) {
    if (raw === null || raw === undefined) continue;
    out[key] = String(raw);
  }
  return out;
}
