/**
 * @vitest-environment jsdom
 *
 * Item 4.2 — buildPracticeHandoffParams must propagate all stable IDs +
 * the PracticeDisplayContext to the generating route, and never include
 * raw answer / question / hint / prompt / provenance fields.
 */

import { describe, expect, it } from "vitest";

import { DEFAULT_INTERVIEW_CONTEXT } from "../../../interview-context/InterviewContext";
import {
  buildPracticeHandoffParams,
  findForbiddenHandoffKeys,
} from "./practiceHandoffParams";

const REPORT_ID = "01918fa0-0000-7000-8000-00000000a000";

describe("buildPracticeHandoffParams", () => {
  it("emits all stable owner IDs + PracticeDisplayContext fields", () => {
    const params = buildPracticeHandoffParams({
      ctx: {
        ...DEFAULT_INTERVIEW_CONTEXT,
        planId: "plan-1",
        targetJobId: "tj-1",
        jdId: "jd-1",
        resumeId: "rv-1",
        roundId: "round-1",
        sessionId: "sess-1",
      },
      reportId: REPORT_ID,
      mode: "text",
      modality: "text",
      practiceMode: "assisted",
      practiceGoal: "baseline",
      hintCount: 3,
    });

    expect(params.planId).toBe("plan-1");
    expect(params.targetJobId).toBe("tj-1");
    expect(params.jdId).toBe("jd-1");
    expect(params.resumeId).toBe("rv-1");
    expect(params.roundId).toBe("round-1");
    expect(params.sessionId).toBe("sess-1");
    expect(params.reportId).toBe(REPORT_ID);
    expect(params.mode).toBe("text");
    expect(params.modality).toBe("text");
    expect(params.practiceMode).toBe("assisted");
    expect(params.practiceGoal).toBe("baseline");
    expect(params.hintUsed).toBe("true");
    expect(params.hintCount).toBe("3");
  });

  it("hintUsed is 'false' when hintCount is 0", () => {
    const params = buildPracticeHandoffParams({
      ctx: {
        ...DEFAULT_INTERVIEW_CONTEXT,
        targetJobId: "tj-1",
        sessionId: "sess-1",
      },
      reportId: REPORT_ID,
      mode: "text",
      modality: "text",
      practiceMode: "strict",
      practiceGoal: "baseline",
      hintCount: 0,
    });
    expect(params.hintUsed).toBe("false");
    expect(params.hintCount).toBe("0");
  });

  it("does not contain forbidden raw text or provenance fields", () => {
    const params = buildPracticeHandoffParams({
      ctx: {
        ...DEFAULT_INTERVIEW_CONTEXT,
        sessionId: "sess-1",
      },
      reportId: REPORT_ID,
      mode: "text",
      modality: "text",
      practiceMode: "strict",
      practiceGoal: "baseline",
      hintCount: 1,
    });
    expect(findForbiddenHandoffKeys(params as unknown as Record<string, unknown>)).toBeNull();
    expect(params).not.toHaveProperty("answerText");
    expect(params).not.toHaveProperty("questionText");
    expect(params).not.toHaveProperty("hint");
    expect(params).not.toHaveProperty("prompt");
    expect(params).not.toHaveProperty("provenance");
    expect(params).not.toHaveProperty("modelId");
  });

  it("findForbiddenHandoffKeys catches injected raw fields", () => {
    expect(
      findForbiddenHandoffKeys({
        sessionId: "sess-1",
        answerText: "leak",
      }),
    ).toEqual(["answerText"]);
    expect(
      findForbiddenHandoffKeys({
        sessionId: "sess-1",
        modelId: "leak",
        provenance: "leak",
      }),
    ).toEqual(["modelId", "provenance"]);
  });
});
