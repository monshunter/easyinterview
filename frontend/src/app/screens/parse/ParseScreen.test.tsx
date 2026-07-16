// @vitest-environment jsdom
import { describe, expect, it, vi } from "vitest";
import { render, screen } from "@testing-library/react";

import { DisplayPreferencesProvider } from "../../display/DisplayPreferencesProvider";
import { NavigationProvider } from "../../navigation/NavigationProvider";
import { ParseScreen } from "./ParseScreen";

function wrap(ui: React.ReactElement, navigate = vi.fn()) {
  return (
    <NavigationProvider value={{ navigate }}>
      <DisplayPreferencesProvider>{ui}</DisplayPreferencesProvider>
    </NavigationProvider>
  );
}

describe("ParseScreen", () => {
  it("renders loading state with 4 progress steps and expected testids", () => {
    render(
      wrap(
        <ParseScreen
          route={{ name: "parse", params: { targetJobId: "tj-1" } }}
        />,
      ),
    );

    expect(screen.getByTestId("parse-loading-step-0")).toBeInTheDocument();
    expect(screen.getByTestId("parse-loading-step-1")).toBeInTheDocument();
    expect(screen.getByTestId("parse-loading-step-2")).toBeInTheDocument();
    expect(screen.getByTestId("parse-loading-step-3")).toBeInTheDocument();
    expect(screen.queryByTestId("parse-loading-footer")).not.toBeInTheDocument();
  });

  it("does not render internal parse metadata while loading", () => {
    render(
      wrap(
        <ParseScreen
          route={{ name: "parse", params: { targetJobId: "tj-1" } }}
        />,
      ),
    );

    expect(screen.queryByTestId("parse-loading-footer")).not.toBeInTheDocument();
    expect(document.body.textContent).not.toMatch(
      /model|provider|rubric|prompt@|provenance|typical|latency/i,
    );
  });

  it("renders shell data attributes", () => {
    render(
      wrap(
        <ParseScreen
          route={{ name: "parse", params: { targetJobId: "tj-1" } }}
        />,
      ),
    );

    const root = screen.getByTestId("route-parse");
    expect(root).toBeInTheDocument();
    expect(root.getAttribute("data-route-name")).toBe("parse");
  });

  it("renders preview state with all required testids", () => {
    render(
      wrap(
        <ParseScreen
          route={{ name: "workspace", params: { targetJobId: "tj-1" } }}
          _mockStage="preview"
          _mockTargetJob={{
            id: "tj-1",
            title: "Senior Frontend Engineer",
            companyName: "StarRing Tech",
            locationText: "Shanghai · Hybrid",
            analysisStatus: "ready",
            status: "draft",
            targetLanguage: "zh-CN",
            createdAt: new Date().toISOString(),
            updatedAt: new Date().toISOString(),
            openQuestionIssueCount: 0,
            requirements: [
              {
                id: "req-1",
                kind: "must_have" as const,
                label: "React 18 + TypeScript",
                evidenceLevel: "explicit" as const,
              },
              {
                id: "req-2",
                kind: "nice_to_have" as const,
                label: "WAI-ARIA experience",
                evidenceLevel: "implicit" as const,
              },
              {
                id: "req-3",
                kind: "hidden_signal" as const,
                label: "Hiring team values architecture influence",
                evidenceLevel: "inferred" as const,
              },
            ],
            summary: {
              coreThemes: ["frontend architecture"],
              interviewRounds: [
                {
                  sequence: 1,
                  type: "hr",
                  name: "Recruiter screen",
                  durationMinutes: 30,
                  focus: "LLM HR screen probes motivation fit",
                },
                {
                  sequence: 2,
                  type: "technical",
                  name: "Frontend architecture interview",
                  durationMinutes: 55,
                  focus: "LLM technical round probes frontend architecture",
                },
              ],
              provenance: {
                modelId: "claude-haiku-4.5",
                promptVersion: "prompt@a8f2e1",
                rubricVersion: "jd-parse-v1.3",
                dataSourceVersion: "2026-05-08",
                featureFlag: "jd-parse-default",
                language: "zh-CN",
              },
            },
            fitSummary: {
              riskSignals: ["Startup ambiguity"],
              strengths: ["Strong React"],
              provenance: {
                modelId: "claude-haiku-4.5",
                promptVersion: "prompt@a8f2e1",
                rubricVersion: "jd-parse-v1.3",
                dataSourceVersion: "2026-05-08",
                featureFlag: "jd-parse-default",
                language: "zh-CN",
              },
            },
          }}
        />,
      ),
    );

    expect(screen.getByTestId("parse-basics-title")).toBeInTheDocument();
    expect(screen.getByTestId("parse-basics-title").querySelector("input")).toBeNull();
    expect(screen.getByTestId("unified-plan-detail")).toBeInTheDocument();
    expect(screen.getByTestId("unified-plan-detail-title")).toHaveTextContent(
      "Review your interview plan",
    );
    expect(document.body).not.toHaveTextContent("这是我从 JD 里读出来的内容");
    expect(document.body).not.toHaveTextContent("Here's what I read from the JD");
    expect(screen.getByTestId("parse-basics-company")).toBeInTheDocument();
    expect(screen.getByTestId("parse-basics-location")).toBeInTheDocument();
    expect(screen.queryByTestId("parse-basics-notes")).not.toBeInTheDocument();
    expect(screen.getByTestId("parse-basics-level")).toBeInTheDocument();
    expect(screen.getByTestId("parse-basics-language")).toBeInTheDocument();

    expect(
      screen.getByTestId("parse-requirement-must_have-0"),
    ).toBeInTheDocument();
    expect(
      screen.queryByTestId("parse-requirement-must_have-0-toggle"),
    ).not.toBeInTheDocument();
    expect(
      screen.getByTestId("parse-requirement-nice_to_have-0"),
    ).toBeInTheDocument();
    expect(
      screen.getByTestId("parse-requirement-must_have-0"),
    ).toHaveTextContent(/Match|命中/);
    expect(
      screen.getByTestId("parse-requirement-nice_to_have-0"),
    ).toHaveTextContent(/Partial match|部分/);
    expect(screen.getByTestId("parse-hidden-signal-0")).toHaveTextContent(
      "Hiring team values architecture influence",
    );
    expect(screen.getByTestId("parse-round-0")).toHaveTextContent(
      "Recruiter screen · 30m",
    );
    expect(screen.getByTestId("parse-round-0")).toHaveTextContent(
      "LLM HR screen probes motivation fit",
    );
    expect(screen.getByTestId("parse-round-1")).toHaveTextContent(
      "Frontend architecture interview · 55m",
    );
    expect(screen.getByTestId("parse-round-1")).toHaveTextContent(
      "LLM technical round probes frontend architecture",
    );
    expect(screen.getByTestId("parse-round-0")).not.toHaveTextContent(
      /Motivation, timing|动机/,
    );
    expect(screen.queryByTestId("parse-launch")).not.toBeInTheDocument();
    expect(screen.queryByTestId("parse-resume-binding")).not.toBeInTheDocument();
    expect(screen.getByTestId("parse-leading-actions")).toBeInTheDocument();
    expect(screen.queryByTestId("parse-action-cancel")).not.toBeInTheDocument();
    expect(screen.queryByTestId("parse-action-reparse")).not.toBeInTheDocument();
    expect(screen.queryByTestId("parse-action-save-plan")).not.toBeInTheDocument();
    expect(screen.getByTestId("parse-action-start-interview")).toBeInTheDocument();
    expect(screen.getByTestId("parse-reports-entry")).toBeInTheDocument();
    expect(screen.queryByTestId("parse-action-confirm")).not.toBeInTheDocument();
  });

  it("level and language fields are read-only", () => {
    render(
      wrap(
        <ParseScreen
          route={{ name: "parse", params: { targetJobId: "tj-1" } }}
          _mockStage="preview"
          _mockTargetJob={{
            id: "tj-1",
            title: "Engineer",
            companyName: "Acme",
            analysisStatus: "ready",
            status: "draft",
            targetLanguage: "zh-CN",
            createdAt: new Date().toISOString(),
            updatedAt: new Date().toISOString(),
            openQuestionIssueCount: 0,
            requirements: [],
          }}
        />,
      ),
    );

    const levelEl = screen.getByTestId("parse-basics-level");
    const langEl = screen.getByTestId("parse-basics-language");
    expect(levelEl.tagName).not.toBe("INPUT");
    expect(langEl.tagName).not.toBe("INPUT");
  });

  it("success preview does not expose a cancel action", () => {
    const navigate = vi.fn();
    render(
      wrap(
        <ParseScreen
          route={{ name: "parse", params: { targetJobId: "tj-1" } }}
          _mockStage="preview"
          _mockTargetJob={{
            id: "tj-1",
            title: "Engineer",
            companyName: "Acme",
            analysisStatus: "ready",
            status: "draft",
            targetLanguage: "zh-CN",
            createdAt: new Date().toISOString(),
            updatedAt: new Date().toISOString(),
            openQuestionIssueCount: 0,
            requirements: [],
          }}
        />,
        navigate,
      ),
    );

    expect(screen.queryByTestId("parse-action-cancel")).not.toBeInTheDocument();
    expect(navigate).not.toHaveBeenCalled();
  });
});
