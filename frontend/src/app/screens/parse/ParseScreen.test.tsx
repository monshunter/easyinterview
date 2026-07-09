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
    expect(screen.getByTestId("parse-loading-footer")).toBeInTheDocument();
  });

  it("renders loading footer without frontend provider or prompt assumptions", () => {
    render(
      wrap(
        <ParseScreen
          route={{ name: "parse", params: { targetJobId: "tj-1" } }}
        />,
      ),
    );

    const footer = screen.getByTestId("parse-loading-footer");
    expect(footer.textContent).not.toMatch(/claude|haiku|prompt@/i);
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
          route={{ name: "parse", params: { targetJobId: "tj-1" } }}
          _mockStage="preview"
          _mockTargetJob={{
            id: "tj-1",
            title: "Senior Frontend Engineer",
            companyName: "StarRing Tech",
            locationText: "Shanghai · Hybrid",
            analysisStatus: "ready",
            status: "draft",
            sourceType: "manual_text",
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
                kind: "nice_to_have" as const,
                label: "Edge runtime familiarity",
                evidenceLevel: "inferred" as const,
              },
            ],
            summary: {
              coreThemes: ["frontend architecture"],
              interviewHypotheses: ["Cross-team influence"],
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
      "Interview plan detail",
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
    ).toHaveTextContent(/HIT|命中/);
    expect(
      screen.getByTestId("parse-requirement-nice_to_have-0"),
    ).toHaveTextContent(/PARTIAL|部分/);
    expect(
      screen.getByTestId("parse-requirement-nice_to_have-1"),
    ).toHaveTextContent(/PARTIAL|部分/);

    expect(screen.getByTestId("parse-hidden-signal-0")).toBeInTheDocument();
    expect(screen.getByTestId("parse-round-0")).toBeInTheDocument();
    expect(screen.getByTestId("parse-launch")).toBeInTheDocument();
    expect(screen.queryByTestId("parse-action-cancel")).not.toBeInTheDocument();
    expect(screen.queryByTestId("parse-action-reparse")).not.toBeInTheDocument();
    expect(screen.queryByTestId("parse-action-save-plan")).not.toBeInTheDocument();
    expect(screen.getByTestId("parse-action-start-interview")).toBeInTheDocument();
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
            sourceType: "manual_text",
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
            sourceType: "manual_text",
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
