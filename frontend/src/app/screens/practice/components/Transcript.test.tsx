/** @vitest-environment jsdom */
import { readFileSync } from "node:fs";
import { resolve } from "node:path";

import { render, screen, within } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";

import { Transcript } from "./Transcript";

const renderTranscript = () =>
  render(
    <Transcript
      messages={[
        {
          id: "assistant-message",
          role: "assistant",
          text: "## 追问\n\n- 背景\n- 结果",
          t: "10:00",
          status: "complete",
        },
        {
          id: "user-message",
          role: "user",
          text: "## 回答\n\n| 指标 | 结果 |\n| --- | --- |\n| 延迟 | 下降 30% |\n\n```ts\nconst safe = true;\n```",
          t: "10:01",
          clientMessageId: "client-message",
          status: "complete",
        },
      ]}
      helperText="helper"
      aiLabel="AI 面试官"
      userLabel="你"
      thinking={false}
      thinkingLabel="思考中"
      retryLabel="重试"
      onRetry={vi.fn()}
    />,
  );

describe("Practice transcript Markdown projection", () => {
  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it("renders persisted assistant and user text through one semantic GFM message body", () => {
    renderTranscript();

    const assistantRow = screen.getByTestId("practice-transcript-message-0");
    const userRow = screen.getByTestId("practice-transcript-message-1");
    const assistantBody = within(assistantRow).getByTestId("practice-message-body");
    const userBody = within(userRow).getByTestId("practice-message-body");

    expect(within(assistantBody).getByRole("heading", { level: 2, name: "追问" })).toBeInTheDocument();
    expect(within(assistantBody).getByRole("list")).toBeInTheDocument();
    expect(within(userBody).getByRole("heading", { level: 2, name: "回答" })).toBeInTheDocument();
    expect(userBody.querySelector("table")).toBeInTheDocument();
    expect(userBody.querySelector("pre code")).toHaveTextContent("const safe = true;");
  });

  it("uses react-markdown with remark-gfm, skipHtml, and no raw-HTML rehype plugin", () => {
    const transcriptSource = readFileSync(resolve(__dirname, "Transcript.tsx"), "utf8");
    const messageBodySource = readFileSync(resolve(__dirname, "PracticeMessageBody.tsx"), "utf8");

    expect(transcriptSource).toContain("<PracticeMessageBody");
    expect(messageBodySource).toContain('from "react-markdown"');
    expect(messageBodySource).toContain('from "remark-gfm"');
    expect(messageBodySource).toContain("remarkPlugins={[remarkGfm]}");
    expect(messageBodySource).toContain("skipHtml");
    expect(messageBodySource).not.toContain("rehypeRaw");
  });

  it("keeps hostile HTML, images, and unsafe links inert while hardening safe external links", () => {
    const fetchSpy = vi.fn();
    vi.stubGlobal("fetch", fetchSpy);
    Object.assign(window, { __practiceMarkdownExecuted: false });

    render(
      <Transcript
        messages={[
          {
            id: "hostile-assistant-message",
            role: "assistant",
            text: [
              '<img src="https://tracker.invalid/raw.png" onerror="window.__practiceMarkdownExecuted=true">',
              '<script>window.__practiceMarkdownExecuted=true</script>',
              '<span onclick="window.__practiceMarkdownExecuted=true">raw event handler</span>',
              "![tracking pixel](https://tracker.invalid/markdown.png)",
              "[unsafe link](javascript:window.__practiceMarkdownExecuted=true)",
              "[safe external link](https://example.com/reference)",
            ].join("\n\n"),
            t: "10:02",
            status: "complete",
          },
        ]}
        helperText="helper"
        aiLabel="AI 面试官"
        userLabel="你"
        thinking={false}
        thinkingLabel="思考中"
        retryLabel="重试"
        onRetry={vi.fn()}
      />,
    );

    const body = screen.getByTestId("practice-message-body");
    expect(body.querySelector("img, script")).toBeNull();
    expect(body.innerHTML).not.toMatch(/onerror|onclick|__practiceMarkdownExecuted/);
    expect(within(body).getByText("unsafe link").closest("a")).toBeNull();

    const safeLink = within(body).getByRole("link", { name: "safe external link" });
    expect(safeLink).toHaveAttribute("href", "https://example.com/reference");
    expect(safeLink).toHaveAttribute("target", "_blank");
    expect(safeLink).toHaveAttribute("rel", "noopener noreferrer");
    expect(fetchSpy).not.toHaveBeenCalled();
    expect((window as unknown as { __practiceMarkdownExecuted: boolean }).__practiceMarkdownExecuted).toBe(false);
  });
});
