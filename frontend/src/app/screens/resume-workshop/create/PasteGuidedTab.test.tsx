// @vitest-environment jsdom
import { describe, expect, it, vi } from "vitest";
import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

import { EasyInterviewClient } from "../../../../api/generated/client";
import {
  createFixtureBackedFetch,
  createFixtureRegistry,
} from "../../../../api/mockTransport";
import { AppRuntimeProvider } from "../../../runtime/AppRuntimeProvider";
import { DisplayPreferencesProvider } from "../../../display/DisplayPreferencesProvider";
import { NavigationProvider } from "../../../navigation/NavigationProvider";
import { ResumeCreateFlow } from "./ResumeCreateFlow";

import getRuntimeConfigFixture from "../../../../../../openapi/fixtures/Auth/getRuntimeConfig.json";
import getMeFixture from "../../../../../../openapi/fixtures/Auth/getMe.json";
import registerResumeFixture from "../../../../../../openapi/fixtures/Resumes/registerResume.json";
import getResumeFixture from "../../../../../../openapi/fixtures/Resumes/getResume.json";

const FIXTURES = [
  getRuntimeConfigFixture,
  getMeFixture,
  registerResumeFixture,
  getResumeFixture,
];

function buildClient(scenario: string = "default"): EasyInterviewClient {
  return new EasyInterviewClient({
    fetch: createFixtureBackedFetch(createFixtureRegistry(FIXTURES), {
      scenario,
    }),
  });
}

function renderFlow(client: EasyInterviewClient) {
  return render(
    <DisplayPreferencesProvider>
      <AppRuntimeProvider
        client={client}
        requestOptions={{
          getMe: { headers: { Prefer: "example=authenticated" } },
        }}
      >
        <NavigationProvider value={{ navigate: vi.fn() }}>
          <ResumeCreateFlow />
        </NavigationProvider>
      </AppRuntimeProvider>
    </DisplayPreferencesProvider>,
  );
}

describe("PasteTab — paste-text register + IK", () => {
  it("submits raw text + IK and transitions to parsing stage", async () => {
    const client = buildClient();
    const registerSpy = vi.spyOn(client, "registerResume");
    // Force polling to stay in non-terminal state so parse-flow remains
    // mounted long enough for the assertion irrespective of polling cadence.
    vi.spyOn(client, "getResume").mockResolvedValue({
      id: "01918fa0-0000-7000-8000-000000001000",
      title: "stub",
      language: "zh",
      parseStatus: "processing",
      createdAt: "2026-05-17T00:00:00Z",
      updatedAt: "2026-05-17T00:00:00Z",
    });
    const user = userEvent.setup();

    renderFlow(client);
    await user.click(screen.getByTestId("resume-create-tab-paste"));
    fireEvent.change(screen.getByTestId("resume-create-paste-textarea"), {
      target: { value: "Hello, I am Alice and I lead frontend work." },
    });
    const submit = screen.getByTestId("resume-create-paste-submit");
    expect(submit).not.toBeDisabled();
    await user.click(submit);
    await waitFor(() =>
      expect(screen.getByTestId("resume-parse-flow")).toBeInTheDocument(),
    );
    expect(registerSpy).toHaveBeenCalledTimes(1);
    const call = registerSpy.mock.calls[0]!;
    expect(call[0]).toMatchObject({
      sourceType: "paste",
      rawText: "Hello, I am Alice and I lead frontend work.",
    });
    expect(call[1]?.idempotencyKey).toMatch(/^v1\.\d+\.[0-9a-f-]{36}$/);
  });
});

describe("GuidedTab — 5 step nav + guided register", () => {
  it("walks through 5 steps and submits guided payload with all 5 keys", async () => {
    const client = buildClient();
    const registerSpy = vi.spyOn(client, "registerResume");
    vi.spyOn(client, "getResume").mockResolvedValue({
      id: "01918fa0-0000-7000-8000-000000001000",
      title: "stub",
      language: "zh",
      parseStatus: "processing",
      createdAt: "2026-05-17T00:00:00Z",
      updatedAt: "2026-05-17T00:00:00Z",
    });
    const user = userEvent.setup();

    renderFlow(client);
    await user.click(screen.getByTestId("resume-create-tab-guided"));

    const fillCurrent = (value: string) => {
      fireEvent.change(screen.getByTestId("resume-create-guided-textarea"), {
        target: { value },
      });
    };
    fillCurrent("Senior FE @ Foo");
    await user.click(screen.getByTestId("resume-create-guided-advance"));
    fillCurrent("frontend platform");
    await user.click(screen.getByTestId("resume-create-guided-advance"));
    fillCurrent("RSC migration");
    await user.click(screen.getByTestId("resume-create-guided-advance"));
    fillCurrent("LCP 3.2s → 1.4s");
    await user.click(screen.getByTestId("resume-create-guided-advance"));
    fillCurrent("Staff frontend");
    // Last step — Generate v1
    await user.click(screen.getByTestId("resume-create-guided-advance"));

    await waitFor(() =>
      expect(screen.getByTestId("resume-parse-flow")).toBeInTheDocument(),
    );
    expect(registerSpy).toHaveBeenCalledTimes(1);
    const payload = registerSpy.mock.calls[0]![0];
    expect(payload.sourceType).toBe("guided");
    expect(payload.guidedAnswers).toEqual({
      recentRole: "Senior FE @ Foo",
      direction: "frontend platform",
      proofProject: "RSC migration",
      metrics: "LCP 3.2s → 1.4s",
      target: "Staff frontend",
    });
    expect(registerSpy.mock.calls[0]![1]?.idempotencyKey).toMatch(
      /^v1\.\d+\.[0-9a-f-]{36}$/,
    );
  });
});

describe("registerResume fixture parity (paste-text / guided-answers scenarios)", () => {
  it("paste-text scenario payload shape is byte-compatible with the hook input", () => {
    const paste = registerResumeFixture.scenarios["paste-text"];
    expect(paste).toBeTruthy();
    const body = paste!.request.body as Record<string, unknown>;
    expect(body.sourceType).toBe("paste");
    expect(typeof body.rawText).toBe("string");
    expect(typeof body.title).toBe("string");
    expect(typeof body.language).toBe("string");
    expect(body.fileObjectId).toBeUndefined();
    expect(body.guidedAnswers).toBeUndefined();
  });

  it("guided-answers scenario payload shape is byte-compatible with the hook input", () => {
    const guided = registerResumeFixture.scenarios["guided-answers"];
    expect(guided).toBeTruthy();
    const body = guided!.request.body as Record<string, unknown>;
    expect(body.sourceType).toBe("guided");
    expect(typeof body.title).toBe("string");
    expect(typeof body.language).toBe("string");
    expect(body.guidedAnswers).toBeTypeOf("object");
    expect(body.rawText).toBeUndefined();
    expect(body.fileObjectId).toBeUndefined();
  });

  it("default upload scenario payload shape is byte-compatible with the hook input", () => {
    const def = registerResumeFixture.scenarios.default;
    const body = def.request.body as Record<string, unknown>;
    expect(body.sourceType).toBe("upload");
    expect(typeof body.fileObjectId).toBe("string");
    expect(typeof body.title).toBe("string");
    expect(typeof body.language).toBe("string");
    expect(body.rawText).toBeUndefined();
    expect(body.guidedAnswers).toBeUndefined();
  });
});
