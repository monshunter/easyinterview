/** @vitest-environment jsdom */

import { readFileSync } from "node:fs";
import { resolve } from "node:path";

import { act, fireEvent, render, screen, waitFor } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";

import type { EasyInterviewClient } from "../../../../api/generated/client";
import type {
  FeedbackReport,
  ReportConversation,
} from "../../../../api/generated/types";
import { DisplayPreferencesProvider } from "../../../display/DisplayPreferencesProvider";
import { NavigationProvider } from "../../../navigation/NavigationProvider";
import type { Route } from "../../../routes";
import {
  AppRuntimeContext,
  type AppRuntimeValue,
} from "../../../runtime/AppRuntimeProvider";
import { ReportDashboard } from "../../report/components/ReportDashboard";
import { ReportConversationScreen } from "../ReportConversationScreen";

const REPORT_ID = "01918fa0-0070-7000-8000-000000000070";
const NEXT_REPORT_ID = "01918fa0-0070-7000-8000-000000000071";

function conversation(
  overrides: Partial<ReportConversation> = {},
): ReportConversation {
  return {
    reportId: REPORT_ID,
    reportStatus: "ready",
    context: {
      sourcePlanId: "01918fa0-0040-7000-8000-000000000040",
      targetJobTitle: "Platform Engineer",
      targetJobCompany: "Acme",
      resumeId: "01918fa0-0010-7000-8000-000000000010",
      resumeDisplayName: "Platform resume",
      roundId: "round-1-technical",
      roundSequence: 1,
      roundName: "Technical interview",
      roundType: "technical",
      language: "en",
      hasNextRound: true,
    },
    messages: [
      {
        sequence: 1,
        role: "assistant",
        content: "## Trade-off follow-up\n\nExplain the constraints first.",
        createdAt: "2026-07-15T08:00:00Z",
      },
      {
        sequence: 2,
        role: "user",
        content: "I would compare **latency** and cost.",
        createdAt: "2026-07-15T08:00:18Z",
      },
    ],
    ...overrides,
  };
}

function feedbackReport(): FeedbackReport {
  return {
    id: REPORT_ID,
    sessionId: "01918fa0-0050-7000-8000-000000000050",
    targetJobId: "01918fa0-0020-7000-8000-000000000020",
    status: "ready",
    errorCode: null,
    summary: "Grounded report summary.",
    preparednessLevel: "needs_practice",
    context: conversation().context,
    dimensionAssessments: [
      {
        code: "technical_depth",
        label: "Technical depth",
        status: "needs_work",
        confidence: "medium",
      },
    ],
    highlights: [],
    issues: [
      {
        dimensionCode: "technical_depth",
        evidence: "Add a measurable trade-off result.",
        confidence: "medium",
      },
    ],
    nextActions: [
      {
        type: "retry_current_round",
        label: "Practice this round with a measurable result.",
      },
    ],
    retryFocusDimensionCodes: ["technical_depth"],
    provenance: {
      promptVersion: "fixture",
      rubricVersion: "fixture",
      modelId: "fixture",
      language: "en",
      featureFlag: "none",
      dataSourceVersion: "fixture",
    },
    createdAt: "2026-07-15T08:00:00Z",
    updatedAt: "2026-07-15T08:01:00Z",
  };
}

function runtimeValue(client: EasyInterviewClient): AppRuntimeValue {
  return {
    client,
    runtime: { status: "ready", config: {} as never },
    auth: { status: "unauthenticated" },
    refreshAuth: () => undefined,
  };
}

function conversationRoute(reportId: string): Route {
  return {
    name: "report_conversation",
    params: { reportId },
  } as unknown as Route;
}

function viewConversation(
  client: EasyInterviewClient,
  reportId = REPORT_ID,
  navigate = vi.fn(),
) {
  return (
    <DisplayPreferencesProvider>
      <AppRuntimeContext.Provider value={runtimeValue(client)}>
        <NavigationProvider value={{ navigate }}>
          <ReportConversationScreen route={conversationRoute(reportId)} />
        </NavigationProvider>
      </AppRuntimeContext.Provider>
    </DisplayPreferencesProvider>
  );
}

function conversationClient(value: unknown): EasyInterviewClient {
  return {
    getReportConversation: vi.fn(async () => value),
  } as unknown as EasyInterviewClient;
}

function deferred<T>() {
  let resolve!: (value: T) => void;
  const promise = new Promise<T>((res) => {
    resolve = res;
  });
  return { promise, resolve };
}

afterEach(() => {
  localStorage.clear();
});

describe("report-owned readonly conversation", () => {
  it.each([
    ["ready", "report"],
    ["queued", "generating"],
    ["generating", "generating"],
    ["failed", "generating"],
  ] as const)("renders ordered readonly messages for %s and returns to %s", async (reportStatus, parentRoute) => {
    const client = conversationClient(conversation({ reportStatus }));
    const navigate = vi.fn();
    render(viewConversation(client, REPORT_ID, navigate));

    const root = await screen.findByTestId("report-conversation-screen");
    expect(client.getReportConversation).toHaveBeenCalledWith(REPORT_ID);
    expect(screen.getByTestId("report-conversation-message-1")).toHaveAttribute(
      "data-role",
      "assistant",
    );
    expect(screen.getByTestId("report-conversation-message-2")).toHaveAttribute(
      "data-role",
      "user",
    );
    expect(root).toHaveTextContent("Explain the constraints first.");
    expect(root).not.toHaveTextContent(REPORT_ID);
    expect(root).not.toHaveTextContent("01918fa0-0050-7000-8000-000000000050");

    fireEvent.click(screen.getByTestId("report-conversation-back-button"));
    expect(navigate).toHaveBeenCalledWith({
      name: parentRoute,
      params: { reportId: REPORT_ID },
    });
  });

  it("keeps a legal empty transcript readable with its frozen report context", async () => {
    const client = conversationClient(conversation({ messages: [] }));
    render(viewConversation(client));

    const root = await screen.findByTestId("report-conversation-screen");
    expect(root).toHaveTextContent("Acme · Platform Engineer");
    expect(screen.getByTestId("report-conversation-context-strip")).toContainElement(
      screen.getByTestId("report-context-strip"),
    );
    expect(screen.getByTestId("report-conversation-empty")).toHaveTextContent("NO MESSAGES");
    expect(screen.getByTestId("report-conversation-empty")).toHaveTextContent(
      "This verified interview record does not contain any messages.",
    );
    expect(screen.queryByTestId("report-conversation-message-1")).not.toBeInTheDocument();
  });

  it("keeps the prototype loading state recoverable back to the workspace", async () => {
    const pending = deferred<ReportConversation>();
    const client = conversationClient(pending.promise);
    const navigate = vi.fn();
    render(viewConversation(client, REPORT_ID, navigate));

    expect(await screen.findByTestId("report-conversation-loading")).toHaveTextContent(
      "LOADING RECORD",
    );
    expect(screen.getByRole("status")).toHaveTextContent(
      "The conversation will appear only after its report context is verified.",
    );
    fireEvent.click(screen.getByTestId("report-conversation-loading-back"));
    expect(navigate).toHaveBeenCalledWith({ name: "workspace", params: {} });
  });

  it.each([
    ["extra report locator", () => ({ ...conversation(), sessionId: "must-not-leak" })],
    ["empty frozen identity", () => ({
      ...conversation(),
      context: { ...conversation().context, targetJobTitle: "" },
    })],
    ["unknown message role", () => ({
      ...conversation(),
      messages: [{ ...conversation().messages[0]!, role: "system" }],
    })],
    ["blank message content", () => ({
      ...conversation(),
      messages: [{ ...conversation().messages[0]!, content: "   " }],
    })],
    ["non-increasing sequence", () => ({
      ...conversation(),
      messages: [
        { ...conversation().messages[0]!, sequence: 2 },
        { ...conversation().messages[1]!, sequence: 1 },
      ],
    })],
    ["extra message locator", () => ({
      ...conversation(),
      messages: [{ ...conversation().messages[0]!, clientMessageId: "must-not-leak" }],
    })],
  ])("fails closed for %s", async (_label, makeInvalid) => {
    const navigate = vi.fn();
    render(viewConversation(conversationClient(makeInvalid()), REPORT_ID, navigate));

    expect(await screen.findByTestId("report-conversation-unavailable")).toBeInTheDocument();
    expect(screen.queryByTestId("report-conversation-transcript")).not.toBeInTheDocument();
    fireEvent.click(screen.getByTestId("report-conversation-unavailable-back"));
    expect(navigate).toHaveBeenCalledWith({ name: "workspace", params: {} });
  });

  it.each([
    ["not found", new Error("HTTP 404: REPORT_NOT_FOUND")],
    ["network failure", new Error("network unavailable")],
  ])("fails closed without a parent fact after %s", async (_label, failure) => {
    const navigate = vi.fn();
    render(viewConversation(conversationClient(Promise.reject(failure)), REPORT_ID, navigate));

    expect(await screen.findByTestId("report-conversation-unavailable")).toBeInTheDocument();
    fireEvent.click(screen.getByTestId("report-conversation-unavailable-back"));
    expect(navigate).toHaveBeenCalledWith({ name: "workspace", params: {} });
  });

  it("clears the previous transcript when reportId switches and ignores its late response", async () => {
    const first = deferred<ReportConversation>();
    const second = deferred<ReportConversation>();
    const client = {
      getReportConversation: vi.fn((reportId: string) =>
        reportId === REPORT_ID ? first.promise : second.promise,
      ),
    } as unknown as EasyInterviewClient;
    const rendered = render(viewConversation(client));

    await waitFor(() => {
      expect(client.getReportConversation).toHaveBeenCalledWith(REPORT_ID);
    });
    rendered.rerender(viewConversation(client, NEXT_REPORT_ID));
    await waitFor(() => {
      expect(client.getReportConversation).toHaveBeenCalledWith(NEXT_REPORT_ID);
    });

    await act(async () => {
      second.resolve(conversation({
        reportId: NEXT_REPORT_ID,
        context: {
          ...conversation().context,
          targetJobCompany: "Second Co",
          targetJobTitle: "Second Role",
        },
      }));
    });
    expect(await screen.findByRole("heading", { name: "Second Co · Second Role" })).toBeInTheDocument();

    await act(async () => {
      first.resolve(conversation());
    });
    expect(screen.getByRole("heading", { name: "Second Co · Second Role" })).toBeInTheDocument();
    expect(screen.queryByText("Acme · Platform Engineer")).not.toBeInTheDocument();
  });

  it("reuses the safe Practice Markdown body and rejects live/session behavior at the source boundary", () => {
    const source = readFileSync(
      resolve(process.cwd(), "src/app/screens/report-conversation/ReportConversationScreen.tsx"),
      "utf8",
    );
    const loaderSource = readFileSync(
      resolve(process.cwd(), "src/app/screens/report-conversation/hooks/useReportConversation.ts"),
      "utf8",
    );
    expect(source).toContain("PracticeMessageBody");
    expect(loaderSource).toContain("getReportConversation");
    expect(source).toContain("ReportContextStrip");
    for (const forbidden of [
      "getPracticeSession",
      "listPracticeSessions",
      "InputBar",
      "Transcript",
      "createPracticeVoiceTurn",
      "localStorage",
      "sessionStorage",
      "sessionId",
      "clientMessageId",
    ]) {
      expect(source).not.toContain(forbidden);
    }
  });

  it("places the dashboard entry below the frozen Context Strip instead of the report Header", async () => {
    const client = {
      getFeedbackReport: vi.fn(async () => feedbackReport()),
    } as unknown as EasyInterviewClient;
    const navigate = vi.fn();
    render(
      <DisplayPreferencesProvider>
        <AppRuntimeContext.Provider value={runtimeValue(client)}>
          <NavigationProvider value={{ navigate }}>
            <ReportDashboard reportId={REPORT_ID} />
          </NavigationProvider>
        </AppRuntimeContext.Provider>
      </DisplayPreferencesProvider>,
    );

    await screen.findByTestId("report-dashboard");
    const strip = screen.getByTestId("report-context-strip");
    const entry = screen.getByTestId("report-conversation-entry");
    expect(strip.compareDocumentPosition(entry) & Node.DOCUMENT_POSITION_FOLLOWING).not.toBe(0);
    expect(screen.getByTestId("report-header")).not.toContainElement(entry);
    fireEvent.click(entry);
    expect(navigate).toHaveBeenCalledWith({
      name: "report_conversation",
      params: { reportId: REPORT_ID },
    });
  });
});
