// @vitest-environment jsdom
import { describe, expect, it, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import type { ReactNode } from "react";

import { DisplayPreferencesProvider } from "../../display/DisplayPreferencesProvider";
import type { JobMatchRecommendation } from "../../../api/generated/types";

import { RecommendedTab } from "./RecommendedTab";

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
    reasons: ["Reason A"],
    risks: ["Risk A"],
    highlights: ["Highlight A"],
    seen: false,
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

const noopHandlers = {
  onSelect: () => undefined,
  onConfirmInterview: () => undefined,
  onToggleSave: () => undefined,
  onOpenSource: () => undefined,
  onMarkNotRelevant: () => undefined,
};

describe("RecommendedTab — list + sticky detail (item 3.1)", () => {
  it("renders root testid", () => {
    const recs = [makeRecommendation()];
    wrap(
      <RecommendedTab
        recommendations={recs}
        loading={false}
        error={null}
        selectedId={recs[0]!.id}
        {...noopHandlers}
      />,
    );
    expect(screen.getByTestId("jdmatch-recommended-tab")).toBeInTheDocument();
  });

  it("renders one card per recommendation in given order", () => {
    const recs = [
      makeRecommendation({ id: "jm-a", title: "First" }),
      makeRecommendation({ id: "jm-b", title: "Second" }),
      makeRecommendation({ id: "jm-c", title: "Third" }),
    ];
    wrap(
      <RecommendedTab
        recommendations={recs}
        loading={false}
        error={null}
        selectedId={recs[0]!.id}
        {...noopHandlers}
      />,
    );
    expect(screen.getByTestId("jdmatch-card-jm-a")).toBeInTheDocument();
    expect(screen.getByTestId("jdmatch-card-jm-b")).toBeInTheDocument();
    expect(screen.getByTestId("jdmatch-card-jm-c")).toBeInTheDocument();
  });

  it("renders sticky JDDetail showing the selected recommendation", () => {
    const recs = [
      makeRecommendation({ id: "jm-a", title: "First" }),
      makeRecommendation({ id: "jm-b", title: "Second" }),
    ];
    wrap(
      <RecommendedTab
        recommendations={recs}
        loading={false}
        error={null}
        selectedId="jm-b"
        {...noopHandlers}
      />,
    );
    const header = screen.getByTestId("jdmatch-detail-header");
    expect(header).toHaveTextContent("Second");
    expect(header).not.toHaveTextContent("First");
  });

  it("falls back to first recommendation when selectedId does not match", () => {
    const recs = [
      makeRecommendation({ id: "jm-a", title: "First" }),
      makeRecommendation({ id: "jm-b", title: "Second" }),
    ];
    wrap(
      <RecommendedTab
        recommendations={recs}
        loading={false}
        error={null}
        selectedId="missing"
        {...noopHandlers}
      />,
    );
    expect(screen.getByTestId("jdmatch-detail-header")).toHaveTextContent(
      "First",
    );
  });

  it("invokes onSelect with the clicked recommendation id", () => {
    const recs = [
      makeRecommendation({ id: "jm-a" }),
      makeRecommendation({ id: "jm-b" }),
    ];
    const onSelect = vi.fn();
    wrap(
      <RecommendedTab
        recommendations={recs}
        loading={false}
        error={null}
        selectedId="jm-a"
        {...noopHandlers}
        onSelect={onSelect}
      />,
    );
    screen.getByTestId("jdmatch-card-jm-b").click();
    expect(onSelect).toHaveBeenCalledWith("jm-b");
  });

  it("renders empty state when recommendations is [] and loading=false", () => {
    wrap(
      <RecommendedTab
        recommendations={[]}
        loading={false}
        error={null}
        selectedId={null}
        {...noopHandlers}
      />,
    );
    expect(screen.getByTestId("jdmatch-recommended-empty")).toBeInTheDocument();
    expect(screen.queryByTestId("jdmatch-detail-header")).toBeNull();
  });

  it("renders loading state when loading=true and no recommendations yet", () => {
    wrap(
      <RecommendedTab
        recommendations={[]}
        loading
        error={null}
        selectedId={null}
        {...noopHandlers}
      />,
    );
    expect(screen.getByTestId("jdmatch-recommended-loading")).toBeInTheDocument();
  });

  it("renders error state with retry button when error is set", () => {
    const onRetry = vi.fn();
    wrap(
      <RecommendedTab
        recommendations={[]}
        loading={false}
        error={new Error("boom")}
        selectedId={null}
        onRetry={onRetry}
        {...noopHandlers}
      />,
    );
    expect(screen.getByTestId("jdmatch-recommended-error")).toBeInTheDocument();
    const retry = screen.getByTestId("jdmatch-recommended-error-retry");
    retry.click();
    expect(onRetry).toHaveBeenCalledTimes(1);
  });

  it("propagates Confirm interview action to onConfirmInterview with the selected recommendation", () => {
    const recs = [makeRecommendation({ id: "jm-a" })];
    const onConfirmInterview = vi.fn();
    wrap(
      <RecommendedTab
        recommendations={recs}
        loading={false}
        error={null}
        selectedId="jm-a"
        {...noopHandlers}
        onConfirmInterview={onConfirmInterview}
      />,
    );
    screen.getByTestId("jdmatch-detail-action-confirm").click();
    expect(onConfirmInterview).toHaveBeenCalledTimes(1);
    expect(onConfirmInterview.mock.calls[0]![0]).toEqual(
      expect.objectContaining({ id: "jm-a" }),
    );
  });

  it("propagates Save toggle to onToggleSave with the selected recommendation", () => {
    const recs = [makeRecommendation({ id: "jm-a" })];
    const onToggleSave = vi.fn();
    wrap(
      <RecommendedTab
        recommendations={recs}
        loading={false}
        error={null}
        selectedId="jm-a"
        {...noopHandlers}
        onToggleSave={onToggleSave}
      />,
    );
    screen.getByTestId("jdmatch-detail-action-save").click();
    expect(onToggleSave).toHaveBeenCalledTimes(1);
    expect(onToggleSave.mock.calls[0]![0]).toEqual(
      expect.objectContaining({ id: "jm-a" }),
    );
  });

  it("propagates Open source to onOpenSource with the selected recommendation", () => {
    const recs = [
      makeRecommendation({ id: "jm-a", sourceUrl: "https://x.example" }),
    ];
    const onOpenSource = vi.fn();
    wrap(
      <RecommendedTab
        recommendations={recs}
        loading={false}
        error={null}
        selectedId="jm-a"
        {...noopHandlers}
        onOpenSource={onOpenSource}
      />,
    );
    screen.getByTestId("jdmatch-detail-action-source").click();
    expect(onOpenSource).toHaveBeenCalledTimes(1);
  });

  it("propagates Mark not relevant to onMarkNotRelevant with the selected recommendation", () => {
    const recs = [makeRecommendation({ id: "jm-a" })];
    const onMarkNotRelevant = vi.fn();
    wrap(
      <RecommendedTab
        recommendations={recs}
        loading={false}
        error={null}
        selectedId="jm-a"
        {...noopHandlers}
        onMarkNotRelevant={onMarkNotRelevant}
      />,
    );
    screen.getByTestId("jdmatch-detail-action-dismiss").click();
    expect(onMarkNotRelevant).toHaveBeenCalledTimes(1);
  });
});
