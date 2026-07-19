/** @vitest-environment jsdom */
import { render, screen, within } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { InputBar } from "./InputBar";

describe("Practice composer helper", () => {
  it("owns the localized helper directly above the input shell", () => {
    render(
      <InputBar
        value=""
        onChange={vi.fn()}
        helperText="像真实面试一样回答，准备好后可以结束面试"
        placeholder="输入你的回答..."
        sendLabel="发送"
        disabled={false}
        onSend={vi.fn()}
      />,
    );

    const composer = screen.getByTestId("practice-input");
    const helper = within(composer).getByTestId("practice-input-helper");
    const shell = within(composer).getByTestId("practice-input-shell");

    expect(helper).toHaveTextContent("像真实面试一样回答，准备好后可以结束面试");
    expect(helper.querySelector("svg")).toBeInTheDocument();
    expect(helper.nextElementSibling).toBe(shell);
    expect(screen.queryByTestId("practice-transcript-helper")).not.toBeInTheDocument();
  });
});
