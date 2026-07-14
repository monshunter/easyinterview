import type { ContentLimits, RuntimeConfig } from "../api/generated/types";

export const DEFAULT_CONTENT_LIMITS: ContentLimits = {
  resumeUploadBytes: 10 * 1024 * 1024,
  resumePasteTextBytes: 384 * 1024,
  targetJobRawTextBytes: 96 * 1024,
  practiceMessageBytes: 32 * 1024,
  practiceSessionTextBytes: 256 * 1024,
};

export function utf8ByteLength(value: string): number {
  return new TextEncoder().encode(value).byteLength;
}

export function resolveContentLimits(
  runtimeConfig: RuntimeConfig | null | undefined,
): ContentLimits {
  const configured = (runtimeConfig as { contentLimits?: Partial<ContentLimits> } | null | undefined)
    ?.contentLimits;
  return {
    resumeUploadBytes: positiveOrDefault(configured?.resumeUploadBytes, DEFAULT_CONTENT_LIMITS.resumeUploadBytes),
    resumePasteTextBytes: positiveOrDefault(configured?.resumePasteTextBytes, DEFAULT_CONTENT_LIMITS.resumePasteTextBytes),
    targetJobRawTextBytes: positiveOrDefault(configured?.targetJobRawTextBytes, DEFAULT_CONTENT_LIMITS.targetJobRawTextBytes),
    practiceMessageBytes: positiveOrDefault(configured?.practiceMessageBytes, DEFAULT_CONTENT_LIMITS.practiceMessageBytes),
    practiceSessionTextBytes: positiveOrDefault(configured?.practiceSessionTextBytes, DEFAULT_CONTENT_LIMITS.practiceSessionTextBytes),
  };
}

function positiveOrDefault(value: number | undefined, fallback: number): number {
  return Number.isSafeInteger(value) && (value ?? 0) > 0 ? value as number : fallback;
}
