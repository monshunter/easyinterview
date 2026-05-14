/**
 * @vitest-environment jsdom
 *
 * Item 4.5 — sessionLost fallback when getPracticeSession returns 404.
 * The user lands on PracticeSessionLostState and the CTA navs back to
 * workspace with the carried context IDs.
 */

import { describe, expect, it } from "vitest";
import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

import {
  buildPracticeClient,
  mountPracticeScreen,
} from "./practiceTestUtils";

describe("practice sessionLost (item 4.5)", () => {
  it("404 from getPracticeSession surfaces PracticeSessionLostState", async () => {
    const { client } = buildPracticeClient({
      scenarioByOp: { getPracticeSession: "missing-session" },
    });
    mountPracticeScreen({ client });

    await waitFor(() => {
      expect(screen.getByTestId("practice-session-lost")).toBeDefined();
    });
    expect(screen.queryByTestId("practice-screen")).toBeNull();
  });

  it("clicking back-to-workspace navs with stable context IDs", async () => {
    const { client } = buildPracticeClient({
      scenarioByOp: { getPracticeSession: "missing-session" },
    });
    const { nav } = mountPracticeScreen({ client });

    const user = userEvent.setup();
    await waitFor(() =>
      expect(screen.getByTestId("practice-session-lost-cta")).toBeDefined(),
    );
    await user.click(screen.getByTestId("practice-session-lost-cta"));
    expect(nav).toHaveBeenCalledWith(
      expect.objectContaining({
        name: "workspace",
        params: expect.objectContaining({
          targetJobId: expect.any(String),
        }),
      }),
    );
  });

  it("missing route sessionId also renders PracticeSessionLostState", async () => {
    mountPracticeScreen({ routeParams: { sessionId: "" } });
    await waitFor(() =>
      expect(screen.getByTestId("practice-session-lost")).toBeDefined(),
    );
  });
});
