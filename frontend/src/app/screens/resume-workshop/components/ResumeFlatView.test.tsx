// @vitest-environment jsdom
import { describe, expect, it, vi } from "vitest";
import { render, screen } from "@testing-library/react";

import { DisplayPreferencesProvider } from "../../../display/DisplayPreferencesProvider";
import { NavigationProvider } from "../../../navigation/NavigationProvider";
import type { UiResumeSource, UiResumeVersion } from "../adapters/resume";
import { ResumeFlatView } from "./ResumeFlatView";

const SOURCE_A: UiResumeSource = {
  id: "asset-a",
  name: "Master Frontend",
  langTag: "中",
  type: "Uploaded",
  createdAt: "2026-04-22",
  status: "active",
  summary: "Frontend platform",
  text: [],
};

const SOURCE_B: UiResumeSource = {
  id: "asset-b",
  name: "Product Platform",
  langTag: "EN",
  type: "Uploaded",
  createdAt: "2026-05-01",
  status: "active",
  summary: "",
  text: [],
};

const HIGH_MATCH: UiResumeVersion = {
  id: "v-high",
  originalId: "asset-a",
  parentVersionId: null,
  name: "High match",
  tag: "TARGETED",
  date: "2026-05-10",
  target: "Northstar",
  bullets: 3,
  accepted: 1,
  match: 84,
  archived: false,
};

const LOW_MATCH: UiResumeVersion = {
  id: "v-low",
  originalId: "asset-b",
  parentVersionId: null,
  name: "Low match",
  tag: "TARGETED",
  date: "2026-05-12",
  target: "Stripe",
  bullets: 2,
  accepted: 0,
  match: 60,
  archived: false,
};

const NO_MATCH_RECENT: UiResumeVersion = {
  id: "v-no-recent",
  originalId: "asset-a",
  parentVersionId: null,
  name: "Master recent",
  tag: "MASTER",
  date: "2026-05-15",
  target: null,
  bullets: 0,
  accepted: 0,
  match: null,
  archived: false,
};

const NO_MATCH_OLDER: UiResumeVersion = {
  id: "v-no-older",
  originalId: "asset-b",
  parentVersionId: null,
  name: "Master older",
  tag: "MASTER",
  date: "2026-04-01",
  target: null,
  bullets: 0,
  accepted: 0,
  match: null,
  archived: false,
};

function renderFlat(versions: UiResumeVersion[]) {
  return render(
    <DisplayPreferencesProvider>
      <NavigationProvider value={{ navigate: vi.fn() }}>
        <ResumeFlatView
          sources={[SOURCE_A, SOURCE_B]}
          versions={versions}
          onOpenVersion={vi.fn()}
        />
      </NavigationProvider>
    </DisplayPreferencesProvider>,
  );
}

describe("ResumeFlatView sorting and rendering", () => {
  it("sorts by match DESC, nulls last, then by updatedAt DESC", () => {
    renderFlat([
      LOW_MATCH,
      NO_MATCH_OLDER,
      HIGH_MATCH,
      NO_MATCH_RECENT,
    ]);

    const ids = Array.from(
      document.querySelectorAll<HTMLElement>(
        "[data-testid^='resume-flat-row-v-'][role='row']",
      ),
    ).map((row) => row.getAttribute("data-testid"));
    expect(ids).toEqual([
      "resume-flat-row-v-high",
      "resume-flat-row-v-low",
      "resume-flat-row-v-no-recent",
      "resume-flat-row-v-no-older",
    ]);
  });

  it("renders the empty placeholder when no versions exist", () => {
    renderFlat([]);
    expect(
      screen.getByTestId("resume-workshop-flat-empty"),
    ).toBeInTheDocument();
  });

  it("renders the source name for each row from the originalId mapping", () => {
    renderFlat([HIGH_MATCH]);
    const row = screen.getByTestId("resume-flat-row-v-high");
    expect(row).toHaveTextContent(SOURCE_A.name);
  });
});
