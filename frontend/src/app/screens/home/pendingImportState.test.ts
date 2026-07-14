import { afterEach, describe, expect, it, vi } from "vitest";

import {
  consumePendingImportIntent,
  storePendingImportIntent,
} from "./pendingImportState";

const intent = {
  rawText: "Senior Frontend Engineer needed",
  targetLanguage: "zh-CN",
  resumeId: "01918fa0-0000-7000-8000-000000001000",
  idempotencyKey: "ik-pending-import",
};

afterEach(() => {
  vi.restoreAllMocks();
});

describe("pending import one-shot vault", () => {
  it("atomically consumes the exact intent once", () => {
    const id = storePendingImportIntent(intent);

    expect(consumePendingImportIntent(id)).toMatchObject(intent);
    expect(consumePendingImportIntent(id)).toBeNull();
  });

  it("deletes an expired intent without returning business data", () => {
    const now = vi.spyOn(Date, "now").mockReturnValue(0);
    const id = storePendingImportIntent(intent);
    now.mockReturnValue(Number.MAX_SAFE_INTEGER);

    expect(consumePendingImportIntent(id)).toBeNull();
    expect(consumePendingImportIntent(id)).toBeNull();
  });

  it("returns null for a vault entry lost across refresh or process restart", () => {
    expect(consumePendingImportIntent("pending-import-missing")).toBeNull();
  });
});
