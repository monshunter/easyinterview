// @vitest-environment jsdom
import { describe, expect, it } from "vitest";
import { render, screen } from "@testing-library/react";

import { App } from "./App";

describe("App route normalization", () => {
  const nonCurrentAliases: Array<[string, string]> = [
    ["welcome", "home"],
    ["growth", "home"],
    ["plan", "workspace"],
    ["mistakes", "report"],
    ["drill", "practice"],
    ["followup", "practice"],
    ["experiences", "resume_versions"],
    ["star", "resume_versions"],
    ["resume", "resume_versions"],
    ["onboarding", "resume_versions"],
    ["voice", "home"],
  ];

  it.each(nonCurrentAliases)(
    "renders %s as the normalized %s view (no standalone non-current screen)",
    (nonCurrent, current) => {
      const { unmount } = render(
        <App initialRoute={{ name: nonCurrent, params: {} }} />,
      );
      // workspace now renders WorkspaceScreen; empty params → WorkspacePlanList
      // resume_versions now renders ResumeWorkshopScreen
      // practice now renders PracticeScreen; empty params → PracticeSessionLostState
      // report now renders ReportScreen; empty params → ReportMissingSessionState
      const currentTestId =
        current === "workspace"
          ? "workspace-plan-list"
          : current === "resume_versions"
            ? "resume-workshop-screen"
            : current === "practice"
              ? "practice-session-lost"
              : current === "report"
                ? "report-missing-session"
                : `route-${current}`;
      expect(screen.getByTestId(currentTestId)).toBeInTheDocument();
      expect(screen.queryByTestId(`route-${nonCurrent}`)).not.toBeInTheDocument();
      unmount();
    },
  );

  it("falls back to home when the initial route is unknown", () => {
    render(<App initialRoute={{ name: "totally-bogus", params: {} }} />);
    expect(screen.getByTestId("route-home")).toBeInTheDocument();
    expect(screen.queryByTestId("route-totally-bogus")).not.toBeInTheDocument();
  });
});
