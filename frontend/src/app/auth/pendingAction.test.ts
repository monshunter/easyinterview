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
    resumeId: "frontend-v3",
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
      resumeId: "frontend-v3",
      roundId: "round-manager",
    });
  });

  it("preserves all interview-context keys exposed to PENDING_ACTION_INTERVIEW_KEYS", () => {
    expect(PENDING_ACTION_INTERVIEW_KEYS).toContain("planId");
    expect(PENDING_ACTION_INTERVIEW_KEYS).toContain("targetJobId");
    expect(PENDING_ACTION_INTERVIEW_KEYS).toContain("jdId");
    expect(PENDING_ACTION_INTERVIEW_KEYS).toContain("resumeId");
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
      email: "alice@example.com",
      displayName: "Alice",
      planId: "plan-1",
      targetJobId: "tj-1",
    });
    expect(decoded?.params).toEqual({
      planId: "plan-1",
      targetJobId: "tj-1",
    });
  });

  it("drops raw payload / AI prompt / auth secret keys at encode (plan 004 §3.1)", () => {
    const encoded = encodePendingAction({
      type: "start_practice",
      label: "立即面试",
      route: "practice",
      params: {
        planId: "plan-1",
        rawText: "raw JD body",
        guidedAnswers: "user answer text",
        prompt: "ai prompt",
        token: "secret-token",
        password: "hunter2",
      },
    });
    expect(encoded).toMatchObject({
      pendingRoute: "practice",
      pendingType: "start_practice",
      pendingLabel: "立即面试",
      planId: "plan-1",
    });
    expect(encoded.rawText).toBeUndefined();
    expect(encoded.guidedAnswers).toBeUndefined();
    expect(encoded.prompt).toBeUndefined();
    expect(encoded.token).toBeUndefined();
    expect(encoded.password).toBeUndefined();
  });

  it("drops raw payload / AI prompt / auth secret keys at decode (plan 004 §3.1)", () => {
    const decoded = decodePendingActionRoute({
      pendingRoute: "practice",
      pendingType: "replay_practice",
      pendingLabel: "复练当前轮",
      planId: "plan-1",
      targetJobId: "tj-1",
      rawText: "raw JD body",
      guidedAnswers: "answers",
      prompt: "ai prompt",
      token: "secret-token",
      password: "hunter2",
    });
    expect(decoded).not.toBeNull();
    expect(decoded?.name).toBe("practice");
    expect(decoded?.params).toEqual({
      planId: "plan-1",
      targetJobId: "tj-1",
    });
  });
});
