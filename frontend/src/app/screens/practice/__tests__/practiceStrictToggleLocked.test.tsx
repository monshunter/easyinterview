/**
 * @vitest-environment jsdom
 *
 * Item 3.7 — Strict toggle visual stays in ui-design parity (role='switch'
 * + aria-checked) but the click handler must not flip the local strict
 * mode and must surface a lock toast. The backend is not contacted.
 */

import { describe, expect, it, vi } from "vitest";
import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

import { buildPracticeClient, mountPracticeScreen } from "./practiceTestUtils";

describe("strict toggle locked at runtime (item 3.7)", () => {
  it("clicking the strict toggle surfaces a lock toast and does not call backend", async () => {
    const { client, calls } = buildPracticeClient();
    const spies = {
      append: vi.spyOn(client, "appendSessionEvent"),
      complete: vi.spyOn(client, "completePracticeSession"),
    };
    mountPracticeScreen({ client });

    const user = userEvent.setup();
    await waitFor(() =>
      expect(screen.getByTestId("practice-topbar-strict")).toBeDefined(),
    );

    const toggleBefore = screen.getByTestId("practice-topbar-strict");
    const ariaBefore = toggleBefore.getAttribute("aria-checked");
    await user.click(toggleBefore);

    // Toast is visible.
    await waitFor(() => {
      expect(screen.getByTestId("practice-strict-locked-toast")).toBeDefined();
    });

    // aria-checked unchanged → strict mode is locked.
    const toggleAfter = screen.getByTestId("practice-topbar-strict");
    expect(toggleAfter.getAttribute("aria-checked")).toBe(ariaBefore);

    expect(spies.append).not.toHaveBeenCalled();
    expect(spies.complete).not.toHaveBeenCalled();
    expect(calls.filter((c) => c.method === "POST").length).toBe(0);
  });
});
