/**
 * Phase 4.0 — Pending action `replay_practice` round-trip and allowlist
 * coverage. Report replay resumes to the report owner; report owns the replay
 * CTA and creates a fresh session directly instead of routing through
 * workspace. The base PendingAction.type is a free-form string so the gate is
 * round-trip integrity + privacy red lines (no raw text on URL params).
 */

import { describe, expect, it } from "vitest";

import {
  decodePendingActionRoute,
  encodePendingAction,
  type PendingAction,
} from "../pendingAction";

const REPLAY_ACTION: PendingAction = {
  type: "replay_practice",
  label: "复练当前轮",
  route: "report",
  params: {
    reportId: "report-1",
  },
};

describe("PendingAction replay_practice", () => {
  it("encodes the type / label / route / params on URL-safe keys (TestPendingActionEncodeDecodeReplayPractice)", () => {
    const encoded = encodePendingAction(REPLAY_ACTION);
    expect(encoded).toMatchObject({
      pendingRoute: "report",
      pendingType: "replay_practice",
      pendingLabel: "复练当前轮",
      reportId: "report-1",
    });
  });

  it("decodes back to the same route + params and never spills reserved keys (TestPendingActionEncodeDecodeReplayPractice)", () => {
    const decoded = decodePendingActionRoute(encodePendingAction(REPLAY_ACTION));
    expect(decoded?.name).toBe("report");
    expect(decoded?.params).toEqual(REPLAY_ACTION.params);
    // The 3 reserved keys must not bleed into restored params.
    const params = decoded?.params ?? {};
    expect(params.pendingRoute).toBeUndefined();
    expect(params.pendingType).toBeUndefined();
    expect(params.pendingLabel).toBeUndefined();
  });

  it("the replay_practice type is accepted by the freeform allowlist (TestPendingActionReplayPracticeTypeAllowed)", () => {
    const action: PendingAction = {
      ...REPLAY_ACTION,
      type: "replay_practice",
    };
    // PendingAction.type is a free-form string in the contract; this assert
    // documents that no validator rejects `replay_practice`.
    expect(action.type).toBe("replay_practice");
    const encoded = encodePendingAction(action);
    expect(encoded.pendingType).toBe("replay_practice");
  });

  it("never carries raw answer / question / hint text on URL params (privacy red line)", () => {
    const encoded = encodePendingAction(REPLAY_ACTION);
    for (const value of Object.values(encoded)) {
      expect(value).not.toMatch(/answerText/i);
      expect(value).not.toMatch(/questionText/i);
      expect(value).not.toMatch(/hint:/i);
      expect(value).not.toMatch(/promptHash/i);
    }
  });
});
