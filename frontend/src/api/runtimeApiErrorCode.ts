import type { ApiErrorCode } from "./generated/types";
import { ALL_ERROR_CODES } from "../lib/conventions";

/**
 * Complete runtime catalog for the OpenAPI ApiErrorCode union.
 * shared/conventions owns the common codes; privacy export is the one
 * API-specific extension currently present in the generated union.
 */
export const ALL_API_ERROR_CODES = defineAllApiErrorCodes([
  ...ALL_ERROR_CODES,
  "PRIVACY_EXPORT_NOT_AVAILABLE",
] as const);

const apiErrorCodes = new Set<string>(ALL_API_ERROR_CODES);

export function isApiErrorCode(value: unknown): value is ApiErrorCode {
  return typeof value === "string" && apiErrorCodes.has(value);
}

function defineAllApiErrorCodes<const TValues extends readonly ApiErrorCode[]>(
  values: TValues & ([ApiErrorCode] extends [TValues[number]] ? unknown : never),
): TValues {
  return values;
}
