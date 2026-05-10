// @vitest-environment jsdom
import { describe, expect, it, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import type { ReactNode } from "react";

import { DisplayPreferencesProvider } from "../../display/DisplayPreferencesProvider";
import type { JobMatchRecommendation } from "../../../api/generated/types";

import { JDDetail } from "./JDDetail";

function makeRecommendation(
  overrides: Partial<JobMatchRecommendation> = {},
): JobMatchRecommendation {
  return {
    id: "jm-1",
    title: "Senior Frontend Engineer",
    company: "Acme",
    companyTag: "Series C",
    level: "Senior",
    location: "Shanghai · Hybrid",
    comp: "70-95 LPA",
    posted: "2 days ago",
    score: 92,
    fit: { must: 4, total: 5, plus: 3, totalPlus: 4 },
    reasons: ["Reason A", "Reason B"],
    risks: ["Risk A", "Risk B"],
    highlights: ["Highlight A", "Highlight B"],
    seen: true,
    saved: false,
    sourceUrl: "https://acme.example/careers",
    sourceLabel: "acme.example/careers",
    networkNote: null,
    similarInterviewers: null,
    interviewHypotheses: [],
    provenance: {
      promptVersion: "jd_match_recommendation.v1",
      rubricVersion: "jd_match_recommendation_rubric.v1",
      modelId: "model-profile:contract.default",
      language: "zh-CN",
      featureFlag: "none",
      dataSourceVersion: "jd_match.v1",
    },
    ...overrides,
  };
}

function wrap(node: ReactNode, lang: "zh" | "en" = "en") {
  return render(
    <DisplayPreferencesProvider initial={{ lang }}>
      {node}
    </DisplayPreferencesProvider>,
  );
}

describe("JDDetail — view-model parity (item 3.1)", () => {
  it("renders nothing when recommendation is null", () => {
    const { container } = wrap(
      <JDDetail
        recommendation={null}
        onConfirmInterview={() => undefined}
        onToggleSave={() => undefined}
        onOpenSource={() => undefined}
        onMarkNotRelevant={() => undefined}
      />,
    );
    expect(container.firstChild).toBeNull();
  });

  it("renders header with title, score and /100 suffix", () => {
    const rec = makeRecommendation({ score: 92, title: "Senior FE" });
    wrap(
      <JDDetail
        recommendation={rec}
        onConfirmInterview={() => undefined}
        onToggleSave={() => undefined}
        onOpenSource={() => undefined}
        onMarkNotRelevant={() => undefined}
      />,
    );
    const header = screen.getByTestId("jdmatch-detail-header");
    expect(header).toHaveTextContent("Senior FE");
    expect(header).toHaveTextContent("92");
    expect(header).toHaveTextContent("/100");
  });

  it("renders Why-it-matches section with all reasons", () => {
    const rec = makeRecommendation({ reasons: ["Reason A", "Reason B"] });
    wrap(
      <JDDetail
        recommendation={rec}
        onConfirmInterview={() => undefined}
        onToggleSave={() => undefined}
        onOpenSource={() => undefined}
        onMarkNotRelevant={() => undefined}
      />,
    );
    const why = screen.getByTestId("jdmatch-detail-why");
    expect(why).toHaveTextContent("Reason A");
    expect(why).toHaveTextContent("Reason B");
  });

  it("renders Where-it-stretches section with all risks", () => {
    const rec = makeRecommendation({ risks: ["Risk A", "Risk B"] });
    wrap(
      <JDDetail
        recommendation={rec}
        onConfirmInterview={() => undefined}
        onToggleSave={() => undefined}
        onOpenSource={() => undefined}
        onMarkNotRelevant={() => undefined}
      />,
    );
    const risk = screen.getByTestId("jdmatch-detail-risk");
    expect(risk).toHaveTextContent("Risk A");
    expect(risk).toHaveTextContent("Risk B");
  });

  it("renders Role-snapshot section with all highlights as list items", () => {
    const rec = makeRecommendation({
      highlights: ["Highlight A", "Highlight B"],
    });
    wrap(
      <JDDetail
        recommendation={rec}
        onConfirmInterview={() => undefined}
        onToggleSave={() => undefined}
        onOpenSource={() => undefined}
        onMarkNotRelevant={() => undefined}
      />,
    );
    const snapshot = screen.getByTestId("jdmatch-detail-snapshot");
    expect(snapshot).toHaveTextContent("Highlight A");
    expect(snapshot).toHaveTextContent("Highlight B");
    expect(snapshot.querySelectorAll("li").length).toBe(2);
  });

  it("renders INTEL section when networkNote is present", () => {
    const rec = makeRecommendation({
      networkNote: "2 alumni from past mocks",
      similarInterviewers: null,
    });
    wrap(
      <JDDetail
        recommendation={rec}
        onConfirmInterview={() => undefined}
        onToggleSave={() => undefined}
        onOpenSource={() => undefined}
        onMarkNotRelevant={() => undefined}
      />,
    );
    expect(screen.getByTestId("jdmatch-detail-intel")).toBeInTheDocument();
    expect(screen.getByTestId("jdmatch-detail-intel")).toHaveTextContent(
      "2 alumni from past mocks",
    );
  });

  it("renders INTEL section when similarInterviewers > 0", () => {
    const rec = makeRecommendation({
      networkNote: null,
      similarInterviewers: 3,
    });
    wrap(
      <JDDetail
        recommendation={rec}
        onConfirmInterview={() => undefined}
        onToggleSave={() => undefined}
        onOpenSource={() => undefined}
        onMarkNotRelevant={() => undefined}
      />,
    );
    expect(screen.getByTestId("jdmatch-detail-intel")).toBeInTheDocument();
    expect(screen.getByTestId("jdmatch-detail-intel")).toHaveTextContent("3");
  });

  it("hides INTEL section when networkNote and similarInterviewers are both empty", () => {
    const rec = makeRecommendation({
      networkNote: null,
      similarInterviewers: null,
    });
    wrap(
      <JDDetail
        recommendation={rec}
        onConfirmInterview={() => undefined}
        onToggleSave={() => undefined}
        onOpenSource={() => undefined}
        onMarkNotRelevant={() => undefined}
      />,
    );
    expect(screen.queryByTestId("jdmatch-detail-intel")).toBeNull();
  });

  it("renders compliant INTEL disclaimer wording (public interview-review / JD / company-source signals)", () => {
    const rec = makeRecommendation({
      networkNote: null,
      similarInterviewers: 4,
    });
    wrap(
      <JDDetail
        recommendation={rec}
        onConfirmInterview={() => undefined}
        onToggleSave={() => undefined}
        onOpenSource={() => undefined}
        onMarkNotRelevant={() => undefined}
      />,
      "en",
    );
    const intel = screen.getByTestId("jdmatch-detail-intel");
    expect(intel.textContent).toMatch(
      /public interview-review.*JD.*company-source/i,
    );
  });

  it("renders four action buttons with stable testids", () => {
    const rec = makeRecommendation();
    wrap(
      <JDDetail
        recommendation={rec}
        onConfirmInterview={() => undefined}
        onToggleSave={() => undefined}
        onOpenSource={() => undefined}
        onMarkNotRelevant={() => undefined}
      />,
    );
    expect(
      screen.getByTestId("jdmatch-detail-action-confirm"),
    ).toBeInTheDocument();
    expect(screen.getByTestId("jdmatch-detail-action-save")).toBeInTheDocument();
    expect(
      screen.getByTestId("jdmatch-detail-action-source"),
    ).toBeInTheDocument();
    expect(
      screen.getByTestId("jdmatch-detail-action-dismiss"),
    ).toBeInTheDocument();
  });

  it("Confirm interview button invokes onConfirmInterview", () => {
    const rec = makeRecommendation();
    const onConfirmInterview = vi.fn();
    wrap(
      <JDDetail
        recommendation={rec}
        onConfirmInterview={onConfirmInterview}
        onToggleSave={() => undefined}
        onOpenSource={() => undefined}
        onMarkNotRelevant={() => undefined}
      />,
    );
    screen.getByTestId("jdmatch-detail-action-confirm").click();
    expect(onConfirmInterview).toHaveBeenCalledTimes(1);
  });

  it("Save button invokes onToggleSave and label switches by saved state", () => {
    const onToggleSave = vi.fn();
    const recA = makeRecommendation({ saved: false });
    const { rerender } = wrap(
      <JDDetail
        recommendation={recA}
        onConfirmInterview={() => undefined}
        onToggleSave={onToggleSave}
        onOpenSource={() => undefined}
        onMarkNotRelevant={() => undefined}
      />,
      "en",
    );
    const saveBtnA = screen.getByTestId("jdmatch-detail-action-save");
    expect(saveBtnA.textContent?.toLowerCase()).not.toContain("saved");
    saveBtnA.click();
    expect(onToggleSave).toHaveBeenCalledTimes(1);

    rerender(
      <DisplayPreferencesProvider initial={{ lang: "en" }}>
        <JDDetail
          recommendation={makeRecommendation({ saved: true })}
          onConfirmInterview={() => undefined}
          onToggleSave={onToggleSave}
          onOpenSource={() => undefined}
          onMarkNotRelevant={() => undefined}
        />
      </DisplayPreferencesProvider>,
    );
    const saveBtnB = screen.getByTestId("jdmatch-detail-action-save");
    expect(saveBtnB.textContent?.toLowerCase()).toContain("saved");
  });

  it("Source button is disabled when sourceUrl is null", () => {
    const rec = makeRecommendation({ sourceUrl: null });
    wrap(
      <JDDetail
        recommendation={rec}
        onConfirmInterview={() => undefined}
        onToggleSave={() => undefined}
        onOpenSource={() => undefined}
        onMarkNotRelevant={() => undefined}
      />,
    );
    expect(
      (screen.getByTestId("jdmatch-detail-action-source") as HTMLButtonElement)
        .disabled,
    ).toBe(true);
  });

  it("Source button invokes onOpenSource when clicked with non-null url", () => {
    const onOpenSource = vi.fn();
    const rec = makeRecommendation({ sourceUrl: "https://acme.example/careers" });
    wrap(
      <JDDetail
        recommendation={rec}
        onConfirmInterview={() => undefined}
        onToggleSave={() => undefined}
        onOpenSource={onOpenSource}
        onMarkNotRelevant={() => undefined}
      />,
    );
    screen.getByTestId("jdmatch-detail-action-source").click();
    expect(onOpenSource).toHaveBeenCalledTimes(1);
  });

  it("Not relevant button invokes onMarkNotRelevant", () => {
    const rec = makeRecommendation();
    const onMarkNotRelevant = vi.fn();
    wrap(
      <JDDetail
        recommendation={rec}
        onConfirmInterview={() => undefined}
        onToggleSave={() => undefined}
        onOpenSource={() => undefined}
        onMarkNotRelevant={onMarkNotRelevant}
      />,
    );
    screen.getByTestId("jdmatch-detail-action-dismiss").click();
    expect(onMarkNotRelevant).toHaveBeenCalledTimes(1);
  });

  it("renders level / location / comp tags from view-model", () => {
    const rec = makeRecommendation({
      level: "Staff",
      location: "Remote · APAC",
      comp: "85-110 LPA",
    });
    wrap(
      <JDDetail
        recommendation={rec}
        onConfirmInterview={() => undefined}
        onToggleSave={() => undefined}
        onOpenSource={() => undefined}
        onMarkNotRelevant={() => undefined}
      />,
    );
    const header = screen.getByTestId("jdmatch-detail-header");
    expect(header).toHaveTextContent("Staff");
    expect(header).toHaveTextContent("Remote · APAC");
    expect(header).toHaveTextContent("85-110 LPA");
  });
});
