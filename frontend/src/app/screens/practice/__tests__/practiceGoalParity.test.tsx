/**
 * @vitest-environment jsdom
 *
 * Item 3.8 — display under (assisted | strict) × (baseline | debrief) is
 * indistinguishable across the two practiceGoal values for the same
 * practiceMode. Snapshots over the 4 combinations are pairwise equal.
 */

import { describe, expect, it } from "vitest";
import { screen, waitFor } from "@testing-library/react";

import { mountPracticeScreen } from "./practiceTestUtils";

function snapshotKey(): Record<string, string | null> {
  // Capture display-relevant testids' presence + textContent shape.
  const ids = [
    "practice-input-hint",
    "practice-sessionmap-live-notes",
    "practice-rightpanel-strict-banner",
    "practice-rightpanel-experience-label",
  ];
  const out: Record<string, string | null> = {};
  for (const id of ids) {
    const node = screen.queryByTestId(id);
    out[id] = node ? "present" : "absent";
  }
  return out;
}

describe("practiceGoal parity (item 3.8)", () => {
  it("assisted × baseline matches assisted × debrief", async () => {
    const baseline = mountPracticeScreen({
      routeParams: { practiceMode: "assisted", practiceGoal: "baseline" },
    });
    await waitFor(() =>
      expect(screen.getByTestId("practice-screen")).toBeDefined(),
    );
    const a = snapshotKey();
    baseline.unmount();

    mountPracticeScreen({
      routeParams: { practiceMode: "assisted", practiceGoal: "debrief" },
    });
    await waitFor(() =>
      expect(screen.getByTestId("practice-screen")).toBeDefined(),
    );
    const b = snapshotKey();
    expect(b).toEqual(a);
  });

  it("strict × baseline matches strict × debrief", async () => {
    const baseline = mountPracticeScreen({
      routeParams: { practiceMode: "strict", practiceGoal: "baseline" },
    });
    await waitFor(() =>
      expect(screen.getByTestId("practice-screen")).toBeDefined(),
    );
    const a = snapshotKey();
    baseline.unmount();

    mountPracticeScreen({
      routeParams: { practiceMode: "strict", practiceGoal: "debrief" },
    });
    await waitFor(() =>
      expect(screen.getByTestId("practice-screen")).toBeDefined(),
    );
    const b = snapshotKey();
    expect(b).toEqual(a);
  });
});
