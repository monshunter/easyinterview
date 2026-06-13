// @vitest-environment jsdom
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { cleanup, fireEvent, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

import { EasyInterviewClient } from "../../../../api/generated/client";
import type { Resume } from "../../../../api/generated/types";
import { AppRuntimeProvider } from "../../../runtime/AppRuntimeProvider";
import {
  createFixtureBackedFetch,
  createFixtureRegistry,
} from "../../../../api/mockTransport";
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

function buildClient(): EasyInterviewClient {
  return new EasyInterviewClient({
    fetch: createFixtureBackedFetch(createFixtureRegistry(FIXTURES), {
      scenario: "default",
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

const READY_ASSET: Resume = {
  id: "01918fa0-0000-7000-8000-000000001100",
  title: "粘贴的简历",
  displayName: "粘贴的简历",
  language: "zh",
  parseStatus: "ready",
  status: "active",
  sourceType: "paste",
  createdAt: "2026-05-17T00:00:00Z",
  updatedAt: "2026-05-17T00:00:00Z",
  parsedSummary: {
    identity: {
      name: "Alice Example",
      title: "Senior Frontend Engineer",
      location: "上海",
    },
    summary: "Five years frontend platform work.",
    experience: [
      { co: "Foo", role: "Senior FE", period: "2022 — now", bullets: ["A", "B"] },
    ],
    projects: [],
    skills: ["React", "TS"],
    education: [{ school: "Tongji", degree: "BSc CS" }],
  },
  parsedTextSnapshot: "PRIVATE_PARSED_TEXT_SNAPSHOT_DO_NOT_RENDER",
  originalText: "PRIVATE_ORIGINAL_TEXT_DO_NOT_RENDER",
};

describe("ParsingStage integration", () => {
  beforeEach(() => {
    // Fast polling cadence so tests don't depend on the production 1.5s default.
    window.__EI_RESUME_POLLING_OPTIONS__ = {
      initialDelayMs: 10,
      backoffFactor: 1,
      maxAttempts: 6,
      maxTotalMs: 2_000,
    };
  });
  afterEach(() => {
    cleanup();
    vi.restoreAllMocks();
    vi.useRealTimers();
    delete window.__EI_RESUME_POLLING_OPTIONS__;
  });

  it("polling resolves and transitions to preview without rendering parsedTextSnapshot or originalText in the DOM", { timeout: 10_000 }, async () => {
    const client = buildClient();
    vi.spyOn(client, "registerResume").mockResolvedValue({
      resumeId: "01918fa0-0000-7000-8000-000000001100",
      job: {
        id: "job-1",
        jobType: "resume_parse",
        status: "queued",
        resourceType: "resume_asset",
        resourceId: "01918fa0-0000-7000-8000-000000001100",
        errorCode: null,
        createdAt: "2026-05-17T00:00:00Z",
        updatedAt: "2026-05-17T00:00:00Z",
      } as never,
    });
    const getResumeSpy = vi.spyOn(client, "getResume");
    getResumeSpy.mockResolvedValueOnce({ ...READY_ASSET, parseStatus: "queued" });
    getResumeSpy.mockResolvedValue(READY_ASSET);

    const user = userEvent.setup();
    renderFlow(client);
    await user.click(screen.getByTestId("resume-create-tab-paste"));
    fireEvent.change(screen.getByTestId("resume-create-paste-textarea"), {
      target: { value: "Some pasted resume" },
    });
    await user.click(screen.getByTestId("resume-create-paste-submit"));

		// Eventually transitions to preview, surfacing only structured fields.
		await waitFor(() =>
			expect(screen.getByTestId("resume-preview-confirm")).toBeInTheDocument(),
		);
		const fullDom = document.body.textContent ?? "";
		expect(fullDom).not.toContain("PRIVATE_PARSED_TEXT_SNAPSHOT_DO_NOT_RENDER");
		expect(fullDom).not.toContain("PRIVATE_ORIGINAL_TEXT_DO_NOT_RENDER");
		expect(fullDom).not.toContain("Some pasted resume");
		const previewDom =
			screen.getByTestId("resume-preview-confirm").textContent ?? "";
    expect(previewDom).not.toContain(
      "PRIVATE_PARSED_TEXT_SNAPSHOT_DO_NOT_RENDER",
    );
    expect(previewDom).not.toContain("PRIVATE_ORIGINAL_TEXT_DO_NOT_RENDER");
    expect(previewDom).toContain("Alice Example");
  });

  it("cancel-and-return preserves the paste raw text and does not trigger a new registerResume", async () => {
    const client = buildClient();
    const registerSpy = vi.spyOn(client, "registerResume").mockResolvedValue({
      resumeId: "01918fa0-0000-7000-8000-000000001100",
      job: {} as never,
    });
    vi.spyOn(client, "getResume").mockResolvedValue({
      ...READY_ASSET,
      parseStatus: "processing",
    });

    const user = userEvent.setup();
    renderFlow(client);
    await user.click(screen.getByTestId("resume-create-tab-paste"));
    fireEvent.change(screen.getByTestId("resume-create-paste-textarea"), {
      target: { value: "preserved text content" },
    });
    await user.click(screen.getByTestId("resume-create-paste-submit"));

    await waitFor(() =>
      expect(screen.getByTestId("resume-parse-flow")).toBeInTheDocument(),
    );
    const cancel = screen.getByTestId("resume-parse-flow-cancel");
    await user.click(cancel);

    await waitFor(() =>
      expect(screen.getByTestId("resume-create-flow")).toBeInTheDocument(),
    );
    expect(
      (screen.getByTestId(
        "resume-create-paste-textarea",
      ) as HTMLTextAreaElement).value,
    ).toBe("preserved text content");
    // No new registerResume after cancel.
    expect(registerSpy).toHaveBeenCalledTimes(1);
  });
});
