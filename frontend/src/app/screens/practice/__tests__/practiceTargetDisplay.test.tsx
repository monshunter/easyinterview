/**
 * @vitest-environment jsdom
 */

import { screen, waitFor } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import {
  buildPracticeClient,
  mountPracticeScreen,
} from "./practiceTestUtils";

describe("PracticeScreen target display", () => {
  it("renders the company and role bound to the server session target job", async () => {
    const { client } = buildPracticeClient();
    mountPracticeScreen({ client });

    await waitFor(() =>
      expect(screen.getByTestId("practice-topbar-company")).toHaveTextContent(
        "Acme",
      ),
    );
    expect(screen.getByTestId("practice-topbar-title")).toHaveTextContent(
      "Senior Frontend Engineer",
    );
    expect(screen.queryByText("Target company")).toBeNull();
    expect(screen.queryByText("Target role")).toBeNull();
  });

  it("keeps internal questionIntent out of the user-visible session", async () => {
    const { client } = buildPracticeClient();
    mountPracticeScreen({ client });

    await waitFor(() =>
      expect(screen.getByTestId("practice-question-prompt")).toHaveTextContent(
        "请用 STAR 描述你主导设计系统迁移的项目",
      ),
    );
    expect(
      screen.queryByText("behavioral.leadership.design_system"),
    ).toBeNull();
    expect(screen.getByTestId("practice-question-topic")).toHaveTextContent(
      "Q1",
    );
    expect(screen.getByTestId("practice-sessionmap-item-0")).toHaveTextContent(
      "Q1",
    );
  });
});
