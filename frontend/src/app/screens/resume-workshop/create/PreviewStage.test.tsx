// @vitest-environment jsdom
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { cleanup, fireEvent, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

import { EasyInterviewClient } from "../../../../api/generated/client";
import type {
  ResumeAsset,
  ResumeVersion,
} from "../../../../api/generated/types";
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

const FIXTURES = [
  getRuntimeConfigFixture,
  getMeFixture,
  registerResumeFixture,
  getResumeFixture,
];

const READY_ASSET: ResumeAsset = {
  id: "01918fa0-0000-7000-8000-000000001000",
  title: "粘贴的简历",
  language: "zh",
  parseStatus: "ready",
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

const SAVED_VERSION: ResumeVersion = {
  id: "0195f2d0-0001-7000-8000-000000000201",
  resumeAssetId: READY_ASSET.id,
  versionType: "structured_master",
  displayName: "Structured master",
  parentVersionId: null,
  targetJobId: null,
  seedStrategy: null,
  focusAngle: null,
  structuredProfile: { headline: "Senior Frontend Engineer" },
  suggestions: [],
  provenance: {} as ResumeVersion["provenance"],
  modelId: null,
  promptVersion: null,
  provider: null,
  rubricVersion: null,
  matchScore: null,
  createdAt: "2026-05-17T00:00:00Z",
  updatedAt: "2026-05-17T00:00:00Z",
  deletedAt: null,
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
    resumeAssetId: READY_ASSET.id,
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

describe("PreviewStage — confirm save v1 integration", () => {
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

  it("confirms via confirmResumeStructuredMaster + Idempotency-Key and navigates back to list", async () => {
    const client = buildClient();
    const navigate = vi.fn();
    renderFlow(client, navigate);
    const user = await drivePasteToPreview(client);
    const confirmSpy = vi
      .spyOn(client, "confirmResumeStructuredMaster")
      .mockResolvedValue(SAVED_VERSION);
    await user.click(screen.getByTestId("resume-preview-confirm-save-button"));
    await waitFor(() => expect(confirmSpy).toHaveBeenCalled());
    const call = confirmSpy.mock.calls[0]!;
    expect(call[0]).toBe(READY_ASSET.id);
    expect(call[1]).toMatchObject({
      displayName: "Alice Example",
      language: "en",
    });
    expect(call[1].structuredProfile).toBeTypeOf("object");
    expect(call[2]?.idempotencyKey).toMatch(/^v1\.\d+\.[0-9a-f-]{36}$/);
    // After save, nav back to resume_versions list.
    await waitFor(() => {
      expect(navigate).toHaveBeenCalled();
    });
    expect(navigate.mock.calls.find(([c]) =>
      (c as { name?: string }).name === "resume_versions",
    )).toBeTruthy();
  });

  it("handles 409 already-exists by navigating to the existing master via listResumeVersions", async () => {
    const client = buildClient();
    const navigate = vi.fn();
    renderFlow(client, navigate);
    const user = await drivePasteToPreview(client);
    vi.spyOn(client, "confirmResumeStructuredMaster").mockRejectedValue(
      new Error("HTTP 409 Conflict: RESUME_STRUCTURED_MASTER_ALREADY_EXISTS"),
    );
    vi.spyOn(client, "listResumeVersions").mockResolvedValue({
      items: [
        {
          ...SAVED_VERSION,
          id: "0195f2d0-0001-7000-8000-000000000777",
        },
      ],
			pageInfo: { nextCursor: null, pageSize: 20, hasMore: false },
		});
    await user.click(screen.getByTestId("resume-preview-confirm-save-button"));
    await waitFor(() => {
      expect(navigate).toHaveBeenCalledWith(
        expect.objectContaining({
          name: "resume_versions",
          params: expect.objectContaining({
            versionId: "0195f2d0-0001-7000-8000-000000000777",
            tab: "preview",
          }),
        }),
      );
    });
  });

  it("renders an inline validation error on 422 without navigating", async () => {
    const client = buildClient();
    const navigate = vi.fn();
    renderFlow(client, navigate);
    const user = await drivePasteToPreview(client);
    vi.spyOn(client, "confirmResumeStructuredMaster").mockRejectedValue(
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

  it("does not surface the raw structuredProfile JSON in URL params or localStorage", async () => {
    const client = buildClient();
    const navigate = vi.fn();
    renderFlow(client, navigate);
    const user = await drivePasteToPreview(client);
    vi.spyOn(client, "confirmResumeStructuredMaster").mockResolvedValue(
      SAVED_VERSION,
    );
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

describe("confirmResumeStructuredMaster fixture parity", () => {
  it("default scenario response includes 201 + versionType=structured_master", async () => {
    const fixture = await import(
      "../../../../../../openapi/fixtures/Resumes/confirmResumeStructuredMaster.json"
    );
    const def = fixture.default.scenarios.default;
    expect(def.response.status).toBe(201);
    expect(def.response.body.versionType).toBe("structured_master");
  });
  it("idempotency-replay scenario reuses the same versionId", async () => {
    const fixture = await import(
      "../../../../../../openapi/fixtures/Resumes/confirmResumeStructuredMaster.json"
    );
    const def = fixture.default.scenarios.default;
    const replay = fixture.default.scenarios["idempotency-replay"];
    expect(replay).toBeTruthy();
    expect(replay.response.body.id).toBe(def.response.body.id);
  });
  it("already-exists-409 scenario surfaces RESUME_STRUCTURED_MASTER_ALREADY_EXISTS", async () => {
    const fixture = await import(
      "../../../../../../openapi/fixtures/Resumes/confirmResumeStructuredMaster.json"
    );
    const scenario = fixture.default.scenarios["already-exists-409"];
    expect(scenario.response.status).toBe(409);
    expect(scenario.response.body.error.code).toBe(
      "RESUME_STRUCTURED_MASTER_ALREADY_EXISTS",
    );
  });
  it("validation-422 scenario surfaces VALIDATION_FAILED", async () => {
    const fixture = await import(
      "../../../../../../openapi/fixtures/Resumes/confirmResumeStructuredMaster.json"
    );
    const scenario = fixture.default.scenarios["validation-422"];
    expect(scenario.response.status).toBe(422);
    expect(scenario.response.body.error.code).toBe("VALIDATION_FAILED");
  });
});
