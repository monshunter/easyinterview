import { describe, expect, it } from "vitest";

import type { ApiErrorCode } from "./generated/types";
import { ALL_ERROR_CODES } from "../lib/conventions";
import {
  ALL_API_ERROR_CODES,
  isApiErrorCode,
} from "./runtimeApiErrorCode";

describe("ApiErrorCode runtime catalog", () => {
  it("extends the generated conventions catalog with the OpenAPI privacy code exactly once", () => {
    expect(ALL_API_ERROR_CODES).toEqual([
      ...ALL_ERROR_CODES,
      "PRIVACY_EXPORT_NOT_AVAILABLE",
    ] satisfies readonly ApiErrorCode[]);
    expect(new Set(ALL_API_ERROR_CODES).size).toBe(ALL_API_ERROR_CODES.length);
  });

  it("accepts every runtime code and rejects unknown or non-string values", () => {
    for (const code of ALL_API_ERROR_CODES) expect(isApiErrorCode(code)).toBe(true);
    for (const value of ["REPORT_UNKNOWN_FAILURE", "", null, 404, {}]) {
      expect(isApiErrorCode(value)).toBe(false);
    }
  });
});
