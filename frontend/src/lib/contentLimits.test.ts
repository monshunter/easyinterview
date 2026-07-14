import { describe, expect, it } from "vitest";

import {
  formatBinaryByteLimit,
  utf8ByteLength,
} from "./contentLimits";

describe("content limits", () => {
  it("counts UTF-8 bytes", () => {
    expect(utf8ByteLength("你好")).toBe(6);
    expect(utf8ByteLength("你好a")).toBe(7);
  });

  it("formats binary byte limits without decimal rounding drift", () => {
    expect(formatBinaryByteLimit(2 * 1024 * 1024, true)).toBe("2MiB");
    expect(formatBinaryByteLimit(1536, false)).toBe("1.5 KiB");
    expect(formatBinaryByteLimit(17, false)).toBe("17 bytes");
  });
});
