import { describe, expect, it } from "vitest";

import type { ResumeBranchFormDraft } from "../ResumeBranchFlow";
import {
  mapBranchFormToBranchResumeVersionRequest,
  type BranchSubmitContext,
} from "./mapBranchFormToRequest";

const baseDraft: ResumeBranchFormDraft = {
  name: "v3 ByteDance",
  target: "ByteDance Frontend Platform",
  focus: "platform",
  seed: "copy_master",
};

const baseContext: BranchSubmitContext = {
  parentVersionId: "0195f2d0-0001-7000-8000-000000000201",
  targetJobId: "01918fa0-0000-7000-8000-000000002000",
};

describe("mapBranchFormToBranchResumeVersionRequest", () => {
  it("projects copy_master draft to BranchResumeVersionRequest wire fields", () => {
    const req = mapBranchFormToBranchResumeVersionRequest(
      baseDraft,
      baseContext,
    );
    expect(req).toEqual({
      parentVersionId: baseContext.parentVersionId,
      targetJobId: baseContext.targetJobId,
      seedStrategy: "copy_master",
      displayName: "v3 ByteDance",
      focusAngle: "platform",
    });
  });

  it("preserves blank seedStrategy literal", () => {
    const req = mapBranchFormToBranchResumeVersionRequest(
      { ...baseDraft, seed: "blank", focus: "fullstack" },
      baseContext,
    );
    expect(req.seedStrategy).toBe("blank");
    expect(req.focusAngle).toBe("fullstack");
  });

  it("preserves ai_select seedStrategy literal", () => {
    const req = mapBranchFormToBranchResumeVersionRequest(
      { ...baseDraft, seed: "ai_select", focus: "leadership" },
      baseContext,
    );
    expect(req.seedStrategy).toBe("ai_select");
    expect(req.focusAngle).toBe("leadership");
  });

  it("maps custom focus to `custom:{target}` so backend keeps the bias hint", () => {
    const req = mapBranchFormToBranchResumeVersionRequest(
      {
        ...baseDraft,
        focus: "custom",
        target: " ByteDance Frontend Platform ",
      },
      baseContext,
    );
    expect(req.focusAngle).toBe("custom:ByteDance Frontend Platform");
  });

  it("falls back to literal `custom` when custom focus has empty target", () => {
    const req = mapBranchFormToBranchResumeVersionRequest(
      { ...baseDraft, focus: "custom", target: "" },
      baseContext,
    );
    expect(req.focusAngle).toBe("custom");
  });

  it("trims the display name before sending it on the wire", () => {
    const req = mapBranchFormToBranchResumeVersionRequest(
      { ...baseDraft, name: "  v3 trimmed  " },
      baseContext,
    );
    expect(req.displayName).toBe("v3 trimmed");
  });

  it("throws when parentVersionId is missing in the context", () => {
    expect(() =>
      mapBranchFormToBranchResumeVersionRequest(baseDraft, {
        parentVersionId: "",
        targetJobId: baseContext.targetJobId,
      }),
    ).toThrow(/parentVersionId/);
  });

  it("throws when targetJobId is missing in the context", () => {
    expect(() =>
      mapBranchFormToBranchResumeVersionRequest(baseDraft, {
        parentVersionId: baseContext.parentVersionId,
        targetJobId: "",
      }),
    ).toThrow(/targetJobId/);
  });

  it("throws when displayName is whitespace only", () => {
    expect(() =>
      mapBranchFormToBranchResumeVersionRequest(
        { ...baseDraft, name: "   " },
        baseContext,
      ),
    ).toThrow(/displayName/);
  });

  it("does not leak undeclared fields onto the wire payload", () => {
    const req = mapBranchFormToBranchResumeVersionRequest(
      baseDraft,
      baseContext,
    );
    expect(Object.keys(req).sort()).toEqual(
      [
        "displayName",
        "focusAngle",
        "parentVersionId",
        "seedStrategy",
        "targetJobId",
      ].sort(),
    );
  });
});
