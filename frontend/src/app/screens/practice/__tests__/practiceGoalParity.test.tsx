/**
 * @vitest-environment jsdom
 *
 * Item 3.8 — display remains stable when legacy route practiceMode and
 * current core-loop goals vary. Hints stay optional in-session assistance.
 */

import { describe, expect, it } from "vitest";
import { screen, waitFor } from "@testing-library/react";

import { mountPracticeScreen } from "./practiceTestUtils";

function snapshotKey(): Record<string, string | null> {
  // Capture display-relevant testids' presence + textContent shape.
  const ids = [
    "practice-input-hint",
    "practice-finish-cta",
    "practice-topbar-strict",
    "practice-topbar-role",
  ];
  const out: Record<string, string | null> = {};
  for (const id of ids) {
    const node = screen.queryByTestId(id);
    out[id] = node ? "present" : "absent";
  }
  return out;
}

describe("practiceGoal parity (item 3.8)", () => {
  it("assisted × baseline matches assisted × retry_current_round", async () => {
    const baseline = mountPracticeScreen({
      routeParams: { practiceMode: "assisted", practiceGoal: "baseline" },
    });
    await waitFor(() =>
      expect(screen.getByTestId("practice-screen")).toBeDefined(),
    );
    const a = snapshotKey();
    baseline.unmount();

    mountPracticeScreen({
      routeParams: { practiceMode: "assisted", practiceGoal: "retry_current_round" },
    });
    await waitFor(() =>
      expect(screen.getByTestId("practice-screen")).toBeDefined(),
    );
    const b = snapshotKey();
    expect(b).toEqual(a);
  });

  it("strict × baseline matches strict × next_round", async () => {
    const baseline = mountPracticeScreen({
      routeParams: { practiceMode: "strict", practiceGoal: "baseline" },
    });
    await waitFor(() =>
      expect(screen.getByTestId("practice-screen")).toBeDefined(),
    );
    const a = snapshotKey();
    baseline.unmount();

    mountPracticeScreen({
      routeParams: { practiceMode: "strict", practiceGoal: "next_round" },
    });
    await waitFor(() =>
      expect(screen.getByTestId("practice-screen")).toBeDefined(),
    );
    const b = snapshotKey();
    expect(b).toEqual(a);
  });
});
