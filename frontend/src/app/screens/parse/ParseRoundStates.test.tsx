// @vitest-environment jsdom
import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import type { TargetJob } from "../../../api/generated/types";
import { DisplayPreferencesProvider } from "../../display/DisplayPreferencesProvider";
import { NavigationProvider } from "../../navigation/NavigationProvider";
import { ParseScreen } from "./ParseScreen";

const rounds: NonNullable<TargetJob["summary"]>["interviewRounds"] = [
  {
    sequence: 1,
    type: "technical",
    name: "Frontend architecture screen",
    durationMinutes: 45,
    focus: "Scale a design system across teams.",
  },
  {
    sequence: 2,
    type: "manager",
    name: "Hiring manager interview",
    durationMinutes: 50,
    focus: "Demonstrate ownership and influence.",
  },
  {
    sequence: 3,
    type: "culture",
    name: "Collaboration interview",
    durationMinutes: 40,
    focus: "Explain collaboration and operating style.",
  },
];

function targetJob(
  practiceProgress: TargetJob["practiceProgress"],
  status: TargetJob["status"] = "interviewing",
): TargetJob {
  return {
    id: "target-round-states",
    status,
    analysisStatus: "ready",
    title: "Senior Frontend Engineer",
    companyName: "Acme",
    locationText: "Shanghai · Hybrid",
    targetLanguage: "zh-CN",
    summary: {
      coreThemes: ["Frontend architecture"],
      interviewRounds: rounds,
      provenance: {
        promptVersion: "fixture",
        rubricVersion: "fixture",
        modelId: "fixture",
        language: "zh-CN",
        featureFlag: "none",
        dataSourceVersion: "fixture",
      },
    },
    requirements: [],
    openQuestionIssueCount: 0,
    resumeId: "resume-bound",
    practiceProgress,
    createdAt: "2026-07-14T00:00:00Z",
    updatedAt: "2026-07-14T00:00:00Z",
  };
}

function renderDetail(job: TargetJob, lang: "zh" | "en" = "zh") {
  return render(
    <NavigationProvider value={{ navigate: () => undefined }}>
      <DisplayPreferencesProvider initial={{ lang }}>
        <ParseScreen
          route={{ name: "workspace", params: { targetJobId: job.id } }}
          _mockStage="preview"
          _mockTargetJob={job}
        />
      </DisplayPreferencesProvider>
    </NavigationProvider>,
  );
}

describe("Workspace detail round states", () => {
  it("renders the persisted completed, current and pending sequence with distinct treatments", () => {
    renderDetail(
      targetJob({
        status: "in_progress",
        completedRounds: [
          { roundId: "round-1-technical", roundSequence: 1 },
        ],
        currentRound: { roundId: "round-2-manager", roundSequence: 2 },
      }),
    );

    const done = screen.getByTestId("parse-round-0");
    const current = screen.getByTestId("parse-round-1");
    const pending = screen.getByTestId("parse-round-2");

    expect(done).toHaveAttribute("data-round-state", "done");
    expect(current).toHaveAttribute("data-round-state", "current");
    expect(pending).toHaveAttribute("data-round-state", "pending");
    expect(done).toHaveTextContent("已进行");
    expect(current).toHaveTextContent("即将进行");
    expect(pending).toHaveTextContent("未进行");
    expect(done.style.background).toBe("var(--ei-color-ok-soft)");
    expect(done.style.border).toBe("1px solid var(--ei-color-ok)");
    expect(current.style.background).toBe("var(--ei-color-accent-soft)");
    expect(current.style.border).toBe("1px solid var(--ei-color-accent)");
    expect(pending.style.background).toBe("var(--ei-color-bg-soft)");
    expect(pending.style.border).toBe(
      "1px solid var(--ei-color-rule-strong)",
    );
  });

  it("renders every round as completed when the backend projection is complete", () => {
    renderDetail(
      targetJob({
        status: "completed",
        completedRounds: [
          { roundId: "round-1-technical", roundSequence: 1 },
          { roundId: "round-2-manager", roundSequence: 2 },
          { roundId: "round-3-culture", roundSequence: 3 },
        ],
        currentRound: null,
      }),
      "en",
    );

    expect(screen.getAllByText("Completed")).toHaveLength(3);
    expect(document.querySelectorAll('[data-round-state="done"]')).toHaveLength(3);
    expect(document.querySelector('[data-round-state="current"]')).toBeNull();
    expect(screen.getByTestId("parse-action-start-interview")).toBeDisabled();
  });

  it("fails closed with neutral cards when the projection is missing or invalid", () => {
    const { rerender } = renderDetail(targetJob(undefined));

    for (const card of screen.getAllByTestId(/parse-round-\d/)) {
      expect(card).not.toHaveAttribute("data-round-state");
      expect(card.style.background).toBe("var(--ei-color-bg-soft)");
      expect(card.style.border).toBe(
        "1px solid var(--ei-color-rule-strong)",
      );
    }
    expect(document.body).not.toHaveTextContent(/已进行|即将进行|未进行/);
    expect(screen.getByTestId("parse-action-start-interview")).toBeDisabled();

    const invalid = targetJob({
      status: "in_progress",
      completedRounds: [
        { roundId: "round-2-manager", roundSequence: 2 },
      ],
      currentRound: { roundId: "round-1-technical", roundSequence: 1 },
    });
    rerender(
      <NavigationProvider value={{ navigate: () => undefined }}>
        <DisplayPreferencesProvider initial={{ lang: "zh" }}>
          <ParseScreen
            route={{ name: "workspace", params: { targetJobId: invalid.id } }}
            _mockStage="preview"
            _mockTargetJob={invalid}
          />
        </DisplayPreferencesProvider>
      </NavigationProvider>,
    );

    expect(document.querySelector("[data-round-state]")).toBeNull();
    expect(screen.getByTestId("parse-action-start-interview")).toBeDisabled();
  });

  it("does not derive round state from the TargetJob lifecycle status", () => {
    const progress: NonNullable<TargetJob["practiceProgress"]> = {
      status: "in_progress",
      completedRounds: [
        { roundId: "round-1-technical", roundSequence: 1 },
      ],
      currentRound: { roundId: "round-2-manager", roundSequence: 2 },
    };
    const { rerender } = renderDetail(targetJob(progress, "draft"));
    const statesBefore = screen
      .getAllByTestId(/parse-round-\d/)
      .map((card) => card.getAttribute("data-round-state"));

    rerender(
      <NavigationProvider value={{ navigate: () => undefined }}>
        <DisplayPreferencesProvider initial={{ lang: "zh" }}>
          <ParseScreen
            route={{
              name: "workspace",
              params: { targetJobId: "target-round-states" },
            }}
            _mockStage="preview"
            _mockTargetJob={targetJob(progress, "offer")}
          />
        </DisplayPreferencesProvider>
      </NavigationProvider>,
    );

    expect(
      screen
        .getAllByTestId(/parse-round-\d/)
        .map((card) => card.getAttribute("data-round-state")),
    ).toEqual(statesBefore);
  });
});
