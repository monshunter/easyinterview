import { describe, expect, it } from "vitest";

import {
  decodePendingActionRoute,
  encodePendingAction,
  PENDING_ACTION_INTERVIEW_KEYS,
  type PendingAction,
} from "./pendingAction";

const SAMPLE: PendingAction = {
  type: "start_practice",
  label: "立即面试",
  route: "practice",
  params: {
    planId: "plan-tj-1",
    targetJobId: "tj-1",
    jdId: "jd-tj-1",
    resumeVersionId: "frontend-v3",
    roundId: "round-manager",
  },
};

describe("pendingAction encode/decode", () => {
  it("encodes the route name + reserved metadata + business params under reserved keys", () => {
    const encoded = encodePendingAction(SAMPLE);
    expect(encoded).toMatchObject({
      pendingRoute: "practice",
      pendingType: "start_practice",
      pendingLabel: "立即面试",
      planId: "plan-tj-1",
      targetJobId: "tj-1",
      jdId: "jd-tj-1",
      resumeVersionId: "frontend-v3",
      roundId: "round-manager",
    });
  });

  it("preserves all interview-context keys exposed to PENDING_ACTION_INTERVIEW_KEYS", () => {
    expect(PENDING_ACTION_INTERVIEW_KEYS).toContain("planId");
    expect(PENDING_ACTION_INTERVIEW_KEYS).toContain("targetJobId");
    expect(PENDING_ACTION_INTERVIEW_KEYS).toContain("jdId");
    expect(PENDING_ACTION_INTERVIEW_KEYS).toContain("resumeVersionId");
    expect(PENDING_ACTION_INTERVIEW_KEYS).toContain("roundId");
  });

  it("decodes back to the original route name + interview-context params", () => {
    const encoded = encodePendingAction(SAMPLE);
    const decoded = decodePendingActionRoute(encoded);
    expect(decoded).not.toBeNull();
    expect(decoded?.name).toBe("practice");
    expect(decoded?.params).toEqual(SAMPLE.params);
  });

  it("returns null when no pendingRoute is present", () => {
    expect(decodePendingActionRoute({})).toBeNull();
    expect(decodePendingActionRoute({ planId: "plan-1" })).toBeNull();
  });

  it("strips reserved keys (pendingRoute / pendingType / pendingLabel / returnTo) from restored params", () => {
    const decoded = decodePendingActionRoute({
      pendingRoute: "practice",
      pendingType: "start_practice",
      pendingLabel: "立即面试",
      returnTo: "/practice",
      planId: "plan-1",
      targetJobId: "tj-1",
    });
    expect(decoded?.params).toEqual({
      planId: "plan-1",
      targetJobId: "tj-1",
    });
  });
});
