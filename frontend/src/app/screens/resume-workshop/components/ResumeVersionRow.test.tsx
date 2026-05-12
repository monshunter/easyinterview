// @vitest-environment jsdom
import { describe, expect, it, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

import { DisplayPreferencesProvider } from "../../../display/DisplayPreferencesProvider";
import type { UiResumeVersion } from "../adapters/resume";
import { ResumeVersionRow } from "./ResumeVersionRow";

const TARGETED_VERSION: UiResumeVersion = {
  id: "v-targeted",
  originalId: "asset-1",
  parentVersionId: "v-master",
  name: "Northstar Systems frontend target",
  tag: "TARGETED",
  date: "2026-05-12",
  target: "Northstar",
  bullets: 3,
  accepted: 1,
  match: 84,
  archived: false,
};

const MASTER_VERSION: UiResumeVersion = {
  ...TARGETED_VERSION,
  id: "v-master",
  parentVersionId: null,
  tag: "MASTER",
  match: null,
};

function renderRow(
  version: UiResumeVersion,
  onOpen: (v: UiResumeVersion) => void,
  indent = false,
) {
  return render(
    <DisplayPreferencesProvider>
      <ul>
        <ResumeVersionRow version={version} onOpen={onOpen} indent={indent} />
      </ul>
    </DisplayPreferencesProvider>,
  );
}

describe("ResumeVersionRow rendering and click", () => {
  it("renders tag, date, and match for a TARGETED version", () => {
    renderRow(TARGETED_VERSION, vi.fn());
    const row = screen.getByTestId(`resume-version-row-${TARGETED_VERSION.id}`);
    expect(row).toHaveAttribute("data-tag", "TARGETED");
    expect(row).toHaveTextContent(TARGETED_VERSION.name);
    expect(row).toHaveTextContent(TARGETED_VERSION.date);
    expect(row).toHaveTextContent("84%");
  });

  it("hides the match label for a MASTER version with no matchScore", () => {
    renderRow(MASTER_VERSION, vi.fn());
    const row = screen.getByTestId(`resume-version-row-${MASTER_VERSION.id}`);
    expect(row).toHaveAttribute("data-tag", "MASTER");
    expect(
      screen.queryByTestId(`resume-version-row-${MASTER_VERSION.id}-match`),
    ).not.toBeInTheDocument();
  });

  it("indents row when indent=true", () => {
    renderRow(TARGETED_VERSION, vi.fn(), true);
    const row = screen.getByTestId(`resume-version-row-${TARGETED_VERSION.id}`);
    expect(row).toHaveAttribute("data-indent", "1");
  });

  it("invokes onOpen with the version when the open button is clicked", async () => {
    const onOpen = vi.fn();
    renderRow(TARGETED_VERSION, onOpen);
    const button = screen.getByTestId(
      `resume-version-row-${TARGETED_VERSION.id}-open`,
    );
    await userEvent.setup().click(button);
    expect(onOpen).toHaveBeenCalledWith(TARGETED_VERSION);
  });
});
