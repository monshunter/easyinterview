import { describe, expect, it } from "vitest";

import type { Resume } from "../../api/generated/types";
import {
  hasReadableResumeEvidence,
  isSelectableInterviewResume,
} from "./selectableResume";

const baseResume: Resume = {
  id: "01918fa0-0000-7000-8000-000000001000",
  title: "Resume",
  displayName: "Resume",
  language: "zh-CN",
  parseStatus: "queued",
  status: "active",
  sourceType: "paste",
  originalText: null,
  parsedTextSnapshot: null,
  createdAt: "2026-05-12T08:00:00Z",
  updatedAt: "2026-05-12T08:00:00Z",
  deletedAt: null,
};

describe("selectable interview resumes", () => {
  it("keeps ready non-archived resumes selectable", () => {
    expect(
      isSelectableInterviewResume({
        ...baseResume,
        parseStatus: "ready",
      }),
    ).toBe(true);
  });

  it("keeps non-ready resumes selectable when readable evidence exists", () => {
    expect(
      isSelectableInterviewResume({
        ...baseResume,
        parseStatus: "failed",
        parsedTextSnapshot: "# Resume body",
      }),
    ).toBe(true);
    expect(
      isSelectableInterviewResume({
        ...baseResume,
        parseStatus: "queued",
        originalText: "Pasted resume body",
      }),
    ).toBe(true);
    expect(
      isSelectableInterviewResume({
        ...baseResume,
        parseStatus: "processing",
        structuredProfile: { headline: "Platform engineer" },
      }),
    ).toBe(true);
  });

  it("does not select archived or unreadable pending resumes", () => {
    expect(
      isSelectableInterviewResume({
        ...baseResume,
        parseStatus: "queued",
      }),
    ).toBe(false);
    expect(
      isSelectableInterviewResume({
        ...baseResume,
        parseStatus: "ready",
        status: "archived",
      }),
    ).toBe(false);
    expect(
      hasReadableResumeEvidence({
        ...baseResume,
        structuredProfile: {},
      }),
    ).toBe(false);
  });
});
