// @vitest-environment jsdom
import { describe, expect, it } from "vitest";
import { render, screen } from "@testing-library/react";

import { App } from "./App";

describe("App route normalization", () => {
  const legacyAliases: Array<[string, string]> = [
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

  it.each(legacyAliases)(
    "renders %s as the normalized %s view (no standalone legacy screen)",
    (legacy, current) => {
      const { unmount } = render(
        <App initialRoute={{ name: legacy, params: {} }} />,
      );
      // workspace now renders WorkspaceScreen; empty params → WorkspaceEmptyState
      // resume_versions now renders ResumeWorkshopScreen
      // practice now renders PracticeScreen; empty params → PracticeSessionLostState
      // report now renders ReportScreen; empty params → ReportMissingSessionState
      const currentTestId =
        current === "workspace"
          ? "workspace-empty"
          : current === "resume_versions"
            ? "resume-workshop-screen"
            : current === "practice"
              ? "practice-session-lost"
              : current === "report"
                ? "report-missing-session"
                : `route-${current}`;
      expect(screen.getByTestId(currentTestId)).toBeInTheDocument();
      expect(screen.queryByTestId(`route-${legacy}`)).not.toBeInTheDocument();
      unmount();
    },
  );

  it("falls back to home when the initial route is unknown", () => {
    render(<App initialRoute={{ name: "totally-bogus", params: {} }} />);
    expect(screen.getByTestId("route-home")).toBeInTheDocument();
    expect(screen.queryByTestId("route-totally-bogus")).not.toBeInTheDocument();
  });
});
