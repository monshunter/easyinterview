// @vitest-environment jsdom
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { cleanup, fireEvent, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

import { EasyInterviewClient } from "../../../../api/generated/client";
import type { Resume } from "../../../../api/generated/types";
import { AppRuntimeProvider } from "../../../runtime/AppRuntimeProvider";
import { DisplayPreferencesProvider } from "../../../display/DisplayPreferencesProvider";
import { NavigationProvider } from "../../../navigation/NavigationProvider";
import {
  createFixtureBackedFetch,
  createFixtureRegistry,
} from "../../../../api/mockTransport";
import { ResumeCreateFlow } from "./ResumeCreateFlow";

import getRuntimeConfigFixture from "../../../../../../openapi/fixtures/Auth/getRuntimeConfig.json";
import getMeFixture from "../../../../../../openapi/fixtures/Auth/getMe.json";
import registerResumeFixture from "../../../../../../openapi/fixtures/Resumes/registerResume.json";
import getResumeFixture from "../../../../../../openapi/fixtures/Resumes/getResume.json";
import updateResumeFixture from "../../../../../../openapi/fixtures/Resumes/updateResume.json";

const FIXTURES = [
  getRuntimeConfigFixture,
  getMeFixture,
  registerResumeFixture,
  getResumeFixture,
  updateResumeFixture,
];

const READY_ASSET: Resume = {
  id: "01918fa0-0000-7000-8000-000000001000",
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
      contact: ["alice@example.com"],
    },
    summary: "Owns the surface · 5y platform work.",
    experience: [],
    projects: [],
    skills: ["React"],
    education: [],
  },
};

// D-20: the confirm step overwrites the existing flat resume via updateResume.
const SAVED_RESUME: Resume = {
  ...READY_ASSET,
  displayName: "Alice Example",
  structuredProfile: { headline: "Senior Frontend Engineer" },
  updatedAt: "2026-05-17T01:00:00Z",
};

function buildClient(): EasyInterviewClient {
  return new EasyInterviewClient({
    fetch: createFixtureBackedFetch(createFixtureRegistry(FIXTURES), {
      scenario: "default",
    }),
  });
}

function renderFlow(client: EasyInterviewClient, navigate = vi.fn()) {
  return {
    navigate,
    ...render(
      <DisplayPreferencesProvider>
        <AppRuntimeProvider
          client={client}
          requestOptions={{
            getMe: { headers: { Prefer: "example=authenticated" } },
          }}
        >
          <NavigationProvider value={{ navigate }}>
            <ResumeCreateFlow />
          </NavigationProvider>
        </AppRuntimeProvider>
      </DisplayPreferencesProvider>,
    ),
  };
}

async function drivePasteToPreview(client: EasyInterviewClient) {
  const user = userEvent.setup();
  vi.spyOn(client, "registerResume").mockResolvedValue({
    resumeId: READY_ASSET.id,
    job: {} as never,
  });
  vi.spyOn(client, "getResume").mockResolvedValue(READY_ASSET);
  await user.click(screen.getByTestId("resume-create-tab-paste"));
  fireEvent.change(screen.getByTestId("resume-create-paste-textarea"), {
    target: { value: "Some pasted resume" },
  });
  await user.click(screen.getByTestId("resume-create-paste-submit"));
  await waitFor(() =>
    expect(screen.getByTestId("resume-preview-confirm")).toBeInTheDocument(),
  );
  return user;
}

describe("PreviewStage — confirm save v1 integration (D-20 flat overwrite)", () => {
  beforeEach(() => {
    window.__EI_RESUME_POLLING_OPTIONS__ = {
      initialDelayMs: 10,
      backoffFactor: 1,
      maxAttempts: 3,
      maxTotalMs: 500,
    };
  });
  afterEach(() => {
    cleanup();
    vi.restoreAllMocks();
    delete window.__EI_RESUME_POLLING_OPTIONS__;
  });

  it("confirms via updateResume + Idempotency-Key and navigates back to the list", async () => {
    const client = buildClient();
    const navigate = vi.fn();
    renderFlow(client, navigate);
    const user = await drivePasteToPreview(client);
    const updateSpy = vi
      .spyOn(client, "updateResume")
      .mockResolvedValue(SAVED_RESUME);
    await user.click(screen.getByTestId("resume-preview-confirm-save-button"));
    await waitFor(() => expect(updateSpy).toHaveBeenCalled());
    const call = updateSpy.mock.calls[0]!;
    expect(call[0]).toBe(READY_ASSET.id);
    expect(call[1]).toMatchObject({ displayName: "Alice Example" });
    expect(call[1].structuredProfile).toBeTypeOf("object");
    expect(call[2]?.idempotencyKey).toMatch(/^v1\.\d+\.[0-9a-f-]{36}$/);
    // After save, nav back to resume_versions list.
    await waitFor(() => {
      expect(navigate).toHaveBeenCalled();
    });
    expect(
      navigate.mock.calls.find(
        ([c]) => (c as { name?: string }).name === "resume_versions",
      ),
    ).toBeTruthy();
  });

  it("renders an inline validation error on 422 without navigating", async () => {
    const client = buildClient();
    const navigate = vi.fn();
    renderFlow(client, navigate);
    const user = await drivePasteToPreview(client);
    vi.spyOn(client, "updateResume").mockRejectedValue(
      new Error("HTTP 422 Unprocessable: VALIDATION_FAILED"),
    );
    await user.click(screen.getByTestId("resume-preview-confirm-save-button"));
    await waitFor(() => {
      expect(
        screen.getByTestId("resume-preview-confirm-inline-error"),
      ).toBeInTheDocument();
    });
    expect(navigate).not.toHaveBeenCalled();
  });

  it("renders an inline error on a generic failure without navigating", async () => {
    const client = buildClient();
    const navigate = vi.fn();
    renderFlow(client, navigate);
    const user = await drivePasteToPreview(client);
    vi.spyOn(client, "updateResume").mockRejectedValue(
      new Error("HTTP 500 Internal Server Error"),
    );
    await user.click(screen.getByTestId("resume-preview-confirm-save-button"));
    await waitFor(() => {
      expect(
        screen.getByTestId("resume-preview-confirm-inline-error"),
      ).toBeInTheDocument();
    });
    expect(navigate).not.toHaveBeenCalled();
  });

  it("does not surface the raw structuredProfile JSON in URL params or localStorage", async () => {
    const client = buildClient();
    const navigate = vi.fn();
    renderFlow(client, navigate);
    const user = await drivePasteToPreview(client);
    vi.spyOn(client, "updateResume").mockResolvedValue(SAVED_RESUME);
    const setLocal = vi.spyOn(window.localStorage.__proto__, "setItem");
    const setSession = vi.spyOn(window.sessionStorage.__proto__, "setItem");
    await user.click(screen.getByTestId("resume-preview-confirm-save-button"));
    await waitFor(() => expect(navigate).toHaveBeenCalled());
    const navParams = navigate.mock.calls
      .map(([c]) => (c as { params?: Record<string, string> }).params ?? {})
      .flatMap((p) => Object.values(p));
    for (const v of navParams) {
      expect(v).not.toContain("structuredProfile");
      expect(v).not.toContain("Owns the surface");
    }
    expect(setLocal).not.toHaveBeenCalled();
    expect(setSession).not.toHaveBeenCalled();
  });
});

describe("updateResume fixture parity (D-20)", () => {
  it("default scenario returns 200 + a flat Resume body with displayName", () => {
    const def = updateResumeFixture.scenarios.default;
    expect(def.response.status).toBe(200);
    expect(typeof def.response.body.displayName).toBe("string");
    expect(def.response.body).not.toHaveProperty("versionType");
    expect(def.response.body).not.toHaveProperty("resumeAssetId");
  });
});
