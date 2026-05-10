// @vitest-environment jsdom
import { describe, expect, it, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import type { ReactNode } from "react";

import { DisplayPreferencesProvider } from "../../display/DisplayPreferencesProvider";
import type { JobMatchRecommendation } from "../../../api/generated/types";

import { JobMatchCard } from "./JobMatchCard";

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
    reasons: ["Drove design-system migration across 12 teams"],
    risks: ["Edge runtime experience light"],
    highlights: ["Design system migration"],
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

describe("JobMatchCard — view-model parity (item 3.1)", () => {
  it("renders root testid keyed on recommendation id", () => {
    const rec = makeRecommendation();
    wrap(
      <JobMatchCard
        recommendation={rec}
        active={false}
        onClick={() => undefined}
      />,
    );
    expect(screen.getByTestId(`jdmatch-card-${rec.id}`)).toBeInTheDocument();
  });

  it("renders score testid with the numeric value", () => {
    const rec = makeRecommendation({ score: 92 });
    wrap(
      <JobMatchCard
        recommendation={rec}
        active={false}
        onClick={() => undefined}
      />,
    );
    expect(screen.getByTestId(`jdmatch-card-${rec.id}-score`)).toHaveTextContent(
      "92",
    );
  });

  it("renders STRONG FIT label when score ≥ 85", () => {
    const rec = makeRecommendation({ score: 92 });
    wrap(
      <JobMatchCard
        recommendation={rec}
        active={false}
        onClick={() => undefined}
      />,
      "en",
    );
    expect(screen.getByTestId(`jdmatch-card-${rec.id}-label`)).toHaveTextContent(
      /strong/i,
    );
  });

  it("renders GOOD FIT label when 70 ≤ score < 85", () => {
    const rec = makeRecommendation({ score: 78 });
    wrap(
      <JobMatchCard
        recommendation={rec}
        active={false}
        onClick={() => undefined}
      />,
      "en",
    );
    expect(screen.getByTestId(`jdmatch-card-${rec.id}-label`)).toHaveTextContent(
      /good/i,
    );
  });

  it("renders STRETCH label when score < 70", () => {
    const rec = makeRecommendation({ score: 64 });
    wrap(
      <JobMatchCard
        recommendation={rec}
        active={false}
        onClick={() => undefined}
      />,
      "en",
    );
    expect(screen.getByTestId(`jdmatch-card-${rec.id}-label`)).toHaveTextContent(
      /stretch/i,
    );
  });

  it("renders unseen dot when recommendation.seen is false", () => {
    const rec = makeRecommendation({ seen: false });
    wrap(
      <JobMatchCard
        recommendation={rec}
        active={false}
        onClick={() => undefined}
      />,
    );
    expect(
      screen.getByTestId(`jdmatch-card-${rec.id}-unseen-dot`),
    ).toBeInTheDocument();
  });

  it("hides unseen dot when recommendation.seen is true", () => {
    const rec = makeRecommendation({ seen: true });
    wrap(
      <JobMatchCard
        recommendation={rec}
        active={false}
        onClick={() => undefined}
      />,
    );
    expect(
      screen.queryByTestId(`jdmatch-card-${rec.id}-unseen-dot`),
    ).toBeNull();
  });

  it("renders saved pin when recommendation.saved is true", () => {
    const rec = makeRecommendation({ saved: true });
    wrap(
      <JobMatchCard
        recommendation={rec}
        active={false}
        onClick={() => undefined}
      />,
    );
    expect(
      screen.getByTestId(`jdmatch-card-${rec.id}-saved-pin`),
    ).toBeInTheDocument();
  });

  it("renders top reason from reasons[0]", () => {
    const rec = makeRecommendation({
      reasons: ["First reason wins", "Second reason ignored"],
    });
    wrap(
      <JobMatchCard
        recommendation={rec}
        active={false}
        onClick={() => undefined}
      />,
    );
    const topReason = screen.getByTestId(`jdmatch-card-${rec.id}-top-reason`);
    expect(topReason).toHaveTextContent("First reason wins");
    expect(topReason).not.toHaveTextContent("Second reason ignored");
  });

  it("renders fit footer with must X/Y, plus X/Y, posted, and sourceLabel", () => {
    const rec = makeRecommendation({
      fit: { must: 4, total: 5, plus: 3, totalPlus: 4 },
      posted: "2 days ago",
      sourceLabel: "acme.example/careers",
    });
    wrap(
      <JobMatchCard
        recommendation={rec}
        active={false}
        onClick={() => undefined}
      />,
    );
    const footer = screen.getByTestId(`jdmatch-card-${rec.id}-fit-footer`);
    expect(footer).toHaveTextContent("4/5");
    expect(footer).toHaveTextContent("3/4");
    expect(footer).toHaveTextContent("2 days ago");
    expect(footer).toHaveTextContent("acme.example/careers");
  });

  it("renders company, companyTag, location and comp", () => {
    const rec = makeRecommendation({
      company: "Acme",
      companyTag: "Series C",
      location: "Shanghai · Hybrid",
      comp: "70-95 LPA",
    });
    wrap(
      <JobMatchCard
        recommendation={rec}
        active={false}
        onClick={() => undefined}
      />,
    );
    expect(
      screen.getByTestId(`jdmatch-card-${rec.id}-company`),
    ).toHaveTextContent("Acme");
    expect(
      screen.getByTestId(`jdmatch-card-${rec.id}-company`),
    ).toHaveTextContent("Series C");
    expect(
      screen.getByTestId(`jdmatch-card-${rec.id}-meta`),
    ).toHaveTextContent("Shanghai · Hybrid");
    expect(
      screen.getByTestId(`jdmatch-card-${rec.id}-meta`),
    ).toHaveTextContent("70-95 LPA");
  });

  it("invokes onClick when the card root is clicked", () => {
    const rec = makeRecommendation();
    const onClick = vi.fn();
    wrap(
      <JobMatchCard
        recommendation={rec}
        active={false}
        onClick={onClick}
      />,
    );
    screen.getByTestId(`jdmatch-card-${rec.id}`).click();
    expect(onClick).toHaveBeenCalledTimes(1);
  });

  it("flags active state via data-active attribute", () => {
    const rec = makeRecommendation();
    wrap(
      <JobMatchCard
        recommendation={rec}
        active
        onClick={() => undefined}
      />,
    );
    expect(
      screen.getByTestId(`jdmatch-card-${rec.id}`).getAttribute("data-active"),
    ).toBe("true");
  });

  it("encodes score band on root via data-score-band (strong/good/stretch)", () => {
    const cases = [
      { score: 92, band: "strong" },
      { score: 78, band: "good" },
      { score: 64, band: "stretch" },
    ];
    for (const { score, band } of cases) {
      const rec = makeRecommendation({ id: `jm-${score}`, score });
      wrap(
        <JobMatchCard
          recommendation={rec}
          active={false}
          onClick={() => undefined}
        />,
      );
      expect(
        screen
          .getByTestId(`jdmatch-card-${rec.id}`)
          .getAttribute("data-score-band"),
      ).toBe(band);
    }
  });
});
