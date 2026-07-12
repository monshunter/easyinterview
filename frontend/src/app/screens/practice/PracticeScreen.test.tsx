/** @vitest-environment jsdom */
import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { createDevMockClient } from "../../../api/devMockClient";
import { App } from "../../App";

describe("PracticeScreen continuous conversation", () => {
  it("renders ordered messages in one chat window with voice disabled", async () => {
    render(<App client={createDevMockClient()} requestOptions={{ getMe: { headers: { Prefer: "example=authenticated" } } }} initialRoute={{ name: "practice", params: {
      sessionId: "01918fa0-0000-7000-8000-000000005000",
      planId: "01918fa0-0000-7000-8000-000000004000",
      targetJobId: "01918fa0-0000-7000-8000-000000002000",
    } }} />);

    expect(await screen.findByText("你好，我们直接开始。先聊聊你最近最有代表性的项目。")).toBeInTheDocument();
    expect(screen.getByText("我主导过一次跨团队设计系统迁移。")).toBeInTheDocument();
    expect(screen.getByText("当时最难协调的分歧是什么？")).toBeInTheDocument();
    expect(screen.getByTestId("practice-topbar-phone-toggle")).toBeDisabled();
    expect(screen.queryByText(/题\s*\d+\s*\/\s*\d+/)).not.toBeInTheDocument();
    expect(screen.queryByTestId("practice-session-map")).not.toBeInTheDocument();
    expect(screen.queryByTestId("practice-question-card")).not.toBeInTheDocument();
  });
});
