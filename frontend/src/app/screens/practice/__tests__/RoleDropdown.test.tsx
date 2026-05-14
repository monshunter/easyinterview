/**
 * @vitest-environment jsdom
 *
 * Item 3.4 — RoleDropdown is a UI-only persona switch. It updates local
 * label state and the AI TRANSPARENCY card hint, but MUST NOT call any
 * generated client method.
 */

import { describe, expect, it, vi } from "vitest";
import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

import { buildPracticeClient, mountPracticeScreen } from "./practiceTestUtils";

describe("RoleDropdown (item 3.4)", () => {
  it("changing the role does not invoke any generated client method besides initial GET", async () => {
    const { client, calls } = buildPracticeClient();
    const spies = {
      append: vi.spyOn(client, "appendSessionEvent"),
      complete: vi.spyOn(client, "completePracticeSession"),
      start: vi.spyOn(client, "startPracticeSession"),
    };
    mountPracticeScreen({ client });

    const user = userEvent.setup();
    await waitFor(() =>
      expect(screen.getByTestId("practice-topbar-role")).toBeDefined(),
    );

    await user.click(screen.getByTestId("practice-topbar-role"));
    await waitFor(() =>
      expect(screen.getByTestId("practice-topbar-role-option-hr")).toBeDefined(),
    );
    await user.click(screen.getByTestId("practice-topbar-role-option-hr"));

    // Right panel AI transparency reflects the selected role.
    expect(
      screen.getByTestId("practice-rightpanel-ai-transparency").textContent,
    ).toMatch(/HR/);

    expect(spies.append).not.toHaveBeenCalled();
    expect(spies.complete).not.toHaveBeenCalled();
    expect(spies.start).not.toHaveBeenCalled();
    // Only the initial getPracticeSession should appear in calls (GET).
    const posts = calls.filter((c) => c.method === "POST");
    expect(posts.length).toBe(0);
  });
});
