// @vitest-environment jsdom
import { describe, expect, it, vi } from "vitest";
import { render, screen, fireEvent } from "@testing-library/react";
import type { ReactNode } from "react";

import { DisplayPreferencesProvider } from "../../display/DisplayPreferencesProvider";
import type {
  JobMatchRecommendation,
  SavedSearch,
} from "../../../api/generated/types";

import { SearchTab } from "./SearchTab";

function makeRec(
  overrides: Partial<JobMatchRecommendation> = {},
): JobMatchRecommendation {
  return {
    id: "jm-1",
    title: "Senior FE",
    company: "Acme",
    companyTag: null,
    level: null,
    location: "Remote",
    comp: null,
    posted: "2 days ago",
    score: 92,
    fit: { must: 4, total: 5, plus: 3, totalPlus: 4 },
    reasons: ["Reason A"],
    risks: [],
    highlights: [],
    seen: false,
    saved: false,
    sourceUrl: null,
    sourceLabel: null,
    networkNote: null,
    similarInterviewers: null,
    interviewHypotheses: [],
    provenance: {
      promptVersion: "p.v1",
      rubricVersion: "r.v1",
      modelId: "m",
      language: "zh-CN",
      featureFlag: "none",
      dataSourceVersion: "v1",
    },
    ...overrides,
  };
}

const noopHandlers = {
  setQuery: () => undefined,
  onRun: () => undefined,
  onSaveCurrent: () => undefined,
  setResultFilter: () => undefined,
  onOpenJob: () => undefined,
  onCreateSavedSearchRetry: () => undefined,
};

function wrap(node: ReactNode, lang: "zh" | "en" = "en") {
  return render(
    <DisplayPreferencesProvider initial={{ lang }}>
      {node}
    </DisplayPreferencesProvider>,
  );
}

describe("SearchTab — view-model parity (item 4.1)", () => {
  it("renders root + input + run + sources testids", () => {
    wrap(
      <SearchTab
        query=""
        searching={false}
        results={[]}
        savedSearches={[]}
        resultFilter="all"
        error={null}
        savedSearchesError={null}
        savedSearchCreating={false}
        savedSearchCreateError={null}
        {...noopHandlers}
      />,
    );
    expect(screen.getByTestId("jdmatch-search-tab")).toBeInTheDocument();
    expect(screen.getByTestId("jdmatch-search-input")).toBeInTheDocument();
    expect(screen.getByTestId("jdmatch-search-run")).toBeInTheDocument();
    expect(screen.getByTestId("jdmatch-search-sources")).toBeInTheDocument();
  });

  it("does NOT render searching panel when searching=false", () => {
    wrap(
      <SearchTab
        query=""
        searching={false}
        results={[]}
        savedSearches={[]}
        resultFilter="all"
        error={null}
        savedSearchesError={null}
        savedSearchCreating={false}
        savedSearchCreateError={null}
        {...noopHandlers}
      />,
    );
    expect(
      screen.queryByTestId("jdmatch-search-searching-panel"),
    ).toBeNull();
  });

  it("renders 5-step AGENT panel with i18n step labels and opacity gradient when searching=true", () => {
    wrap(
      <SearchTab
        query="Senior frontend"
        searching
        results={[]}
        savedSearches={[]}
        resultFilter="all"
        error={null}
        savedSearchesError={null}
        savedSearchCreating={false}
        savedSearchCreateError={null}
        {...noopHandlers}
      />,
      "en",
    );
    expect(
      screen.getByTestId("jdmatch-search-searching-panel"),
    ).toBeInTheDocument();
    for (let step = 1; step <= 5; step++) {
      const stepEl = screen.getByTestId(`jdmatch-search-searching-step-${step}`);
      const opacity = (stepEl as HTMLElement).style.opacity;
      // Steps 1-3 are full opacity, 4-5 dimmed to 0.4 (matches ui-design).
      if (step <= 3) {
        expect(opacity).toBe("1");
      } else {
        expect(opacity).toBe("0.4");
      }
    }
  });

  it("renders accent label '● AGENT SCANNING' (en) / '● AGENT 扫描中' (zh)", () => {
    const { unmount } = wrap(
      <SearchTab
        query="x"
        searching
        results={[]}
        savedSearches={[]}
        resultFilter="all"
        error={null}
        savedSearchesError={null}
        savedSearchCreating={false}
        savedSearchCreateError={null}
        {...noopHandlers}
      />,
      "en",
    );
    expect(
      screen.getByTestId("jdmatch-search-searching-panel"),
    ).toHaveTextContent(/agent scanning/i);
    unmount();
    wrap(
      <SearchTab
        query="x"
        searching
        results={[]}
        savedSearches={[]}
        resultFilter="all"
        error={null}
        savedSearchesError={null}
        savedSearchCreating={false}
        savedSearchCreateError={null}
        {...noopHandlers}
      />,
      "zh",
    );
    expect(
      screen.getByTestId("jdmatch-search-searching-panel"),
    ).toHaveTextContent(/AGENT 扫描中/);
  });

  it("does NOT embed any of the forbidden dynamic JD numbers in step labels", () => {
    wrap(
      <SearchTab
        query="x"
        searching
        results={[]}
        savedSearches={[]}
        resultFilter="all"
        error={null}
        savedSearchesError={null}
        savedSearchCreating={false}
        savedSearchCreateError={null}
        {...noopHandlers}
      />,
      "en",
    );
    const panel = screen.getByTestId("jdmatch-search-searching-panel");
    const text = panel.textContent ?? "";
    for (const token of ["248", "87", "unique postings", "唯一岗位"]) {
      expect(text.includes(token), `panel leaked token ${token}`).toBe(false);
    }
  });

  it("renders saved-grid + each saved-item-${id} testid", () => {
    const saved: SavedSearch[] = [
      {
        id: "ss-1",
        label: "Frontend remote",
        query: "frontend remote",
        filters: null,
        newJobsCount: 3,
        lastRunAt: "2026-05-09T10:00:00Z",
        createdAt: "2026-05-09T09:00:00Z",
      },
      {
        id: "ss-2",
        label: "Staff platform",
        query: "staff platform",
        filters: null,
        newJobsCount: 0,
        lastRunAt: "2026-05-09T08:00:00Z",
        createdAt: "2026-05-09T07:00:00Z",
      },
    ];
    wrap(
      <SearchTab
        query=""
        searching={false}
        results={[]}
        savedSearches={saved}
        resultFilter="all"
        error={null}
        savedSearchesError={null}
        savedSearchCreating={false}
        savedSearchCreateError={null}
        {...noopHandlers}
      />,
    );
    expect(screen.getByTestId("jdmatch-search-saved-grid")).toBeInTheDocument();
    expect(
      screen.getByTestId("jdmatch-search-saved-item-ss-1"),
    ).toBeInTheDocument();
    expect(
      screen.getByTestId("jdmatch-search-saved-item-ss-2"),
    ).toBeInTheDocument();
  });

  it("renders 4 filter chip testids (all / strong / remote / unseen)", () => {
    wrap(
      <SearchTab
        query=""
        searching={false}
        results={[]}
        savedSearches={[]}
        resultFilter="all"
        error={null}
        savedSearchesError={null}
        savedSearchCreating={false}
        savedSearchCreateError={null}
        {...noopHandlers}
      />,
    );
    expect(screen.getByTestId("jdmatch-search-filter-all")).toBeInTheDocument();
    expect(screen.getByTestId("jdmatch-search-filter-strong")).toBeInTheDocument();
    expect(screen.getByTestId("jdmatch-search-filter-remote")).toBeInTheDocument();
    expect(screen.getByTestId("jdmatch-search-filter-unseen")).toBeInTheDocument();
  });

  it("renders results grid with at most 6 cards, capped from a larger results array", () => {
    const recs: JobMatchRecommendation[] = Array.from({ length: 9 }, (_, i) =>
      makeRec({ id: `jm-${i + 1}` }),
    );
    wrap(
      <SearchTab
        query="x"
        searching={false}
        results={recs}
        savedSearches={[]}
        resultFilter="all"
        error={null}
        savedSearchesError={null}
        savedSearchCreating={false}
        savedSearchCreateError={null}
        {...noopHandlers}
      />,
    );
    expect(screen.getByTestId("jdmatch-search-results")).toBeInTheDocument();
    for (let i = 1; i <= 6; i++) {
      expect(screen.getByTestId(`jdmatch-card-jm-${i}`)).toBeInTheDocument();
    }
    expect(screen.queryByTestId("jdmatch-card-jm-7")).toBeNull();
  });

  it("renders no-results empty state when results are empty after a search", () => {
    wrap(
      <SearchTab
        query="x"
        searching={false}
        results={[]}
        savedSearches={[]}
        resultFilter="all"
        error={null}
        savedSearchesError={null}
        savedSearchCreating={false}
        savedSearchCreateError={null}
        hasRunOnce
        {...noopHandlers}
      />,
    );
    expect(screen.getByTestId("jdmatch-search-empty")).toBeInTheDocument();
  });

  it("invokes setQuery on input change and onRun on Enter when query is non-empty and not searching", () => {
    const setQuery = vi.fn();
    const onRun = vi.fn();
    wrap(
      <SearchTab
        query="x"
        searching={false}
        results={[]}
        savedSearches={[]}
        resultFilter="all"
        error={null}
        savedSearchesError={null}
        savedSearchCreating={false}
        savedSearchCreateError={null}
        {...noopHandlers}
        setQuery={setQuery}
        onRun={onRun}
      />,
    );
    const input = screen.getByTestId("jdmatch-search-input") as HTMLInputElement;
    fireEvent.change(input, { target: { value: "platform" } });
    expect(setQuery).toHaveBeenCalledWith("platform");
    fireEvent.keyDown(input, { key: "Enter", code: "Enter" });
    expect(onRun).toHaveBeenCalledTimes(1);
  });

  it("disables Run button while searching=true and does not call onRun on Enter", () => {
    const onRun = vi.fn();
    wrap(
      <SearchTab
        query="x"
        searching
        results={[]}
        savedSearches={[]}
        resultFilter="all"
        error={null}
        savedSearchesError={null}
        savedSearchCreating={false}
        savedSearchCreateError={null}
        {...noopHandlers}
        onRun={onRun}
      />,
    );
    const runBtn = screen.getByTestId("jdmatch-search-run") as HTMLButtonElement;
    expect(runBtn.disabled).toBe(true);
    fireEvent.click(runBtn);
    fireEvent.keyDown(screen.getByTestId("jdmatch-search-input"), {
      key: "Enter",
    });
    expect(onRun).not.toHaveBeenCalled();
  });

  it("Save current as watch button invokes onSaveCurrent", () => {
    const onSaveCurrent = vi.fn();
    wrap(
      <SearchTab
        query="frontend"
        searching={false}
        results={[]}
        savedSearches={[]}
        resultFilter="all"
        error={null}
        savedSearchesError={null}
        savedSearchCreating={false}
        savedSearchCreateError={null}
        {...noopHandlers}
        onSaveCurrent={onSaveCurrent}
      />,
    );
    fireEvent.click(screen.getByTestId("jdmatch-search-save-current"));
    expect(onSaveCurrent).toHaveBeenCalledTimes(1);
  });

  it("filter chip click invokes setResultFilter with the chip key", () => {
    const setResultFilter = vi.fn();
    wrap(
      <SearchTab
        query=""
        searching={false}
        results={[]}
        savedSearches={[]}
        resultFilter="all"
        error={null}
        savedSearchesError={null}
        savedSearchCreating={false}
        savedSearchCreateError={null}
        {...noopHandlers}
        setResultFilter={setResultFilter}
      />,
    );
    fireEvent.click(screen.getByTestId("jdmatch-search-filter-strong"));
    expect(setResultFilter).toHaveBeenCalledWith("strong");
  });

  it("renders error inline panel when error is set", () => {
    wrap(
      <SearchTab
        query="x"
        searching={false}
        results={[]}
        savedSearches={[]}
        resultFilter="all"
        error={new Error("HTTP 500 — search down")}
        savedSearchesError={null}
        savedSearchCreating={false}
        savedSearchCreateError={null}
        {...noopHandlers}
      />,
    );
    expect(screen.getByTestId("jdmatch-search-error")).toBeInTheDocument();
  });
});
