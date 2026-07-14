import { describe, expect, it } from "vitest";

import { DEFAULT_CONTENT_LIMITS, resolveContentLimits, utf8ByteLength } from "./contentLimits";

describe("content limits", () => {
  it("counts UTF-8 bytes and accepts the exact boundary only", () => {
    expect(utf8ByteLength("你好")).toBe(6);
    expect(utf8ByteLength("你好")).toBeLessThanOrEqual(6);
    expect(utf8ByteLength("你好a")).toBeGreaterThan(6);
  });

  it("uses A4 defaults for missing or invalid public fields", () => {
    expect(resolveContentLimits(undefined)).toEqual(DEFAULT_CONTENT_LIMITS);
    expect(resolveContentLimits({ contentLimits: { resumeUploadBytes: 0 } } as never))
      .toEqual(DEFAULT_CONTENT_LIMITS);
  });
});
