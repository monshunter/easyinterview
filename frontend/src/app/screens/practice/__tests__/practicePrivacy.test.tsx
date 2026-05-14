/**
 * @vitest-environment jsdom
 *
 * Practice privacy redlines. Raw answer/question/hint/provenance data may be
 * rendered in the active UI where required, but must not leak to navigation
 * params, browser storage, or console output.
 */

import { afterEach, describe, expect, it, vi } from "vitest";
import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

import {
  buildPracticeClient,
  mountPracticeScreen,
} from "./practiceTestUtils";

describe("practice privacy redlines (item 4.10 / Phase 5)", () => {
  afterEach(() => {
    localStorage.clear();
    sessionStorage.clear();
    vi.restoreAllMocks();
  });

  it("does not put raw answer text or provenance into generating route params or storage", async () => {
    const consoleLog = vi.spyOn(console, "log").mockImplementation(() => undefined);
    const rawAnswer = "RAW_ANSWER_DO_NOT_LEAK";
    const { client } = buildPracticeClient();
    const { nav } = mountPracticeScreen({
      client,
      routeParams: { hintCount: "2", hintUsed: "true" },
    });

    const user = userEvent.setup();
    const textarea = screen.getByTestId(
      "practice-input-textarea",
    ) as HTMLTextAreaElement;
    await waitFor(() => expect(textarea.disabled).toBe(false));
    await user.type(textarea, rawAnswer);
    await user.click(screen.getByTestId("practice-input-send"));
    await waitFor(() => expect(textarea.value).toBe(""));

    await user.click(screen.getByTestId("practice-rightpanel-cta-finish"));
    await waitFor(() => expect(nav).toHaveBeenCalled());

    const params = nav.mock.calls.at(-1)?.[0]?.params ?? {};
    expect(JSON.stringify(params)).not.toContain(rawAnswer);
    expect(params).not.toHaveProperty("answerText");
    expect(params).not.toHaveProperty("questionText");
    expect(params).not.toHaveProperty("hint");
    expect(params).not.toHaveProperty("modelId");
    expect(params).not.toHaveProperty("provenance");

    const storageDump = JSON.stringify({
      local: { ...localStorage },
      session: { ...sessionStorage },
    });
    expect(storageDump).not.toContain(rawAnswer);
    expect(storageDump).not.toContain("model-profile:contract.default");
    expect(consoleLog).not.toHaveBeenCalled();
  });
});
