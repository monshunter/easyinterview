// @vitest-environment jsdom
import { describe, expect, it, vi } from "vitest";
import { fireEvent, render, screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

import { EasyInterviewClient } from "../../../../api/generated/client";
import type { Resume } from "../../../../api/generated/types";
import {
  createFixtureBackedFetch,
  createFixtureRegistry,
} from "../../../../api/mockTransport";
import { DisplayPreferencesProvider } from "../../../display/DisplayPreferencesProvider";
import { NavigationProvider } from "../../../navigation/NavigationProvider";
import { AppRuntimeProvider } from "../../../runtime/AppRuntimeProvider";
import type { Route } from "../../../routes";
import { ResumeWorkshopScreen } from "../ResumeWorkshopScreen";

import getRuntimeConfigFixture from "../../../../../../openapi/fixtures/Auth/getRuntimeConfig.json";
import getMeFixture from "../../../../../../openapi/fixtures/Auth/getMe.json";
import getResumeFixture from "../../../../../../openapi/fixtures/Resumes/getResume.json";
import exportResumeFixture from "../../../../../../openapi/fixtures/Resumes/exportResume.json";
import requestResumeTailorFixture from "../../../../../../openapi/fixtures/ResumeTailor/requestResumeTailor.json";
import getResumeTailorRunFixture from "../../../../../../openapi/fixtures/ResumeTailor/getResumeTailorRun.json";

const FIXTURES = [
  getRuntimeConfigFixture,
  getMeFixture,
  getResumeFixture,
  exportResumeFixture,
  requestResumeTailorFixture,
  getResumeTailorRunFixture,
];

function buildClient(scenario: string): EasyInterviewClient {
  return new EasyInterviewClient({
    fetch: createFixtureBackedFetch(createFixtureRegistry(FIXTURES), {
      scenario,
    }),
  });
}

function renderDetail(
  scenario: string,
  route: Route,
  nav: ReturnType<typeof vi.fn> = vi.fn(),
) {
  return render(
    <DisplayPreferencesProvider>
      <AppRuntimeProvider
        client={buildClient(scenario)}
        requestOptions={{
          getMe: { headers: { Prefer: "example=authenticated" } },
        }}
      >
        <NavigationProvider value={{ navigate: nav }}>
          <ResumeWorkshopScreen route={route} />
        </NavigationProvider>
      </AppRuntimeProvider>
    </DisplayPreferencesProvider>,
  );
}

function renderDetailWithClient(
  client: EasyInterviewClient,
  route: Route,
  nav: ReturnType<typeof vi.fn> = vi.fn(),
) {
  return render(
    <DisplayPreferencesProvider>
      <AppRuntimeProvider
        client={client}
        requestOptions={{
          getMe: { headers: { Prefer: "example=authenticated" } },
        }}
      >
        <NavigationProvider value={{ navigate: nav }}>
          <ResumeWorkshopScreen route={route} />
        </NavigationProvider>
      </AppRuntimeProvider>
    </DisplayPreferencesProvider>,
  );
}

const RESUME_ID = getResumeFixture.scenarios.default.response.body.id;

describe("ResumeDetailView container (Phase 3.1)", () => {
  it("renders the crumb and three tabs once the resume loads", async () => {
    renderDetail("default", {
      name: "resume_versions",
      params: { resumeId: RESUME_ID, tab: "preview" },
    });

    await waitFor(() => {
      expect(screen.getByTestId("resume-detail-crumb")).toBeInTheDocument();
    });
    expect(screen.getByTestId("resume-detail-tab-preview")).toBeInTheDocument();
    expect(screen.getByTestId("resume-detail-tab-rewrites")).toBeInTheDocument();
    expect(screen.getByTestId("resume-detail-tab-edit")).toBeInTheDocument();
  });

  it("defaults the active tab to preview when no tab is supplied", async () => {
    renderDetail("default", {
      name: "resume_versions",
      params: { resumeId: RESUME_ID },
    });

    await waitFor(() => {
      expect(screen.getByTestId("resume-detail-tab-preview")).toHaveAttribute(
        "aria-selected",
        "true",
      );
    });
    expect(
      screen.getByTestId("resume-detail-preview-content"),
    ).toBeInTheDocument();
  });

  it("keeps Rewrites active when tab=rewrites is in the URL and renders ResumeRewritesTab", async () => {
    renderDetail("default", {
      name: "resume_versions",
      params: { resumeId: RESUME_ID, tab: "rewrites" },
    });

    await waitFor(() => {
      expect(
        screen.getByTestId("resume-detail-tab-rewrites"),
      ).toBeInTheDocument();
    });
    expect(screen.getByTestId("resume-detail-tab-rewrites")).toHaveAttribute(
      "aria-selected",
      "true",
    );
    expect(screen.getByTestId("resume-rewrites-tab")).toBeInTheDocument();
  });

  it("rerun requests carry the flat { resumeId, targetJobId, mode } and no version ids", async () => {
    const client = buildClient("default");
    const requestSpy = vi
      .spyOn(client, "requestResumeTailor")
      .mockResolvedValueOnce(
        requestResumeTailorFixture.scenarios.default.response.body as never,
      );

    renderDetailWithClient(client, {
      name: "resume_versions",
      params: {
        resumeId: RESUME_ID,
        targetJobId: "01918fa0-0000-7000-8000-000000002000",
        tab: "rewrites",
      },
    });

    await waitFor(() => {
      expect(screen.getByTestId("resume-rewrites-tab")).toBeInTheDocument();
    });
    await userEvent
      .setup()
      .click(screen.getByTestId("resume-rewrites-rerun-tailor"));

    await waitFor(() => {
      expect(requestSpy).toHaveBeenCalledTimes(1);
    });
    const requestArg = requestSpy.mock.calls[0]![0] as unknown as Record<
      string,
      unknown
    >;
    expect(requestArg).toMatchObject({
      resumeId: RESUME_ID,
      targetJobId: "01918fa0-0000-7000-8000-000000002000",
      mode: "bullet_suggestions",
    });
    expect(requestArg).not.toHaveProperty("resumeAssetId");
    expect(requestArg).not.toHaveProperty("resumeVersionId");
  });

  it("keeps the header Export PDF action available on Rewrites and Edit tabs and copies text from the preview tab", async () => {
    const client = buildClient("default");
    const exportSpy = vi.spyOn(client, "exportResume");
    const writeText = vi.fn().mockResolvedValue(undefined);
    const originalNavigator = Object.getOwnPropertyDescriptor(
      window,
      "navigator",
    );
    Object.defineProperty(window, "navigator", {
      configurable: true,
      value: {
        ...window.navigator,
        clipboard: { writeText },
      },
    });

    try {
      renderDetailWithClient(client, {
        name: "resume_versions",
        params: { resumeId: RESUME_ID, tab: "rewrites" },
      });

      await waitFor(() => {
        expect(screen.getByTestId("resume-rewrites-tab")).toBeInTheDocument();
      });

      // Copy Text lives only on the preview tab; Export PDF persists in the
      // header across every tab. From Rewrites, only the header export exists.
      expect(
        screen.queryByTestId("resume-detail-copy-text"),
      ).not.toBeInTheDocument();
      const headerActions = screen.getByTestId("resume-detail-header-actions");
      fireEvent.click(within(headerActions).getByTestId("resume-detail-export-pdf"));
      await waitFor(() => {
        expect(exportSpy).toHaveBeenCalledWith(
          RESUME_ID,
          expect.objectContaining({
            idempotencyKey: expect.stringMatching(/^v1\.\d+\./),
          }),
        );
      });

      // Edit tab also keeps the header export action.
      fireEvent.click(screen.getByTestId("resume-detail-tab-edit"));
      await waitFor(() => {
        expect(screen.getByTestId("resume-edit-tab")).toBeInTheDocument();
      });
      expect(
        within(headerActions).getByTestId("resume-detail-export-pdf"),
      ).toBeInTheDocument();

      // Preview tab is where Copy Text is wired; clicking it writes plain text.
      fireEvent.click(screen.getByTestId("resume-detail-tab-preview"));
      await waitFor(() => {
        expect(
          screen.getByTestId("resume-detail-preview-content"),
        ).toBeInTheDocument();
      });
      fireEvent.click(screen.getByTestId("resume-detail-copy-text"));
      await waitFor(() => {
        expect(writeText).toHaveBeenCalledWith(
          expect.stringContaining("Senior frontend engineer"),
        );
      });
    } finally {
      if (originalNavigator) {
        Object.defineProperty(window, "navigator", originalNavigator);
      }
    }
  });

  it("clicking a tab updates the active selection and shows that tab's content", async () => {
    renderDetail("default", {
      name: "resume_versions",
      params: { resumeId: RESUME_ID, tab: "rewrites" },
    });

    await waitFor(() => {
      expect(
        screen.getByTestId("resume-detail-tab-preview"),
      ).toBeInTheDocument();
    });

    const user = userEvent.setup();
    await user.click(screen.getByTestId("resume-detail-tab-preview"));

    await waitFor(() => {
      expect(screen.getByTestId("resume-detail-tab-preview")).toHaveAttribute(
        "aria-selected",
        "true",
      );
    });
    expect(
      screen.getByTestId("resume-detail-preview-content"),
    ).toBeInTheDocument();
  });

  it("builds the original modal text from the loaded resume's parsed snapshot (no extra getResume call)", async () => {
    const client = buildClient("default");
    const getResumeSpy = vi.spyOn(client, "getResume");

    renderDetailWithClient(client, {
      name: "resume_versions",
      params: { resumeId: RESUME_ID, tab: "preview" },
    });

    await waitFor(() => {
      expect(
        screen.getByTestId("resume-detail-view-original"),
      ).toBeInTheDocument();
    });
    await userEvent
      .setup()
      .click(screen.getByTestId("resume-detail-view-original"));

    const modal = await screen.findByTestId(
      "resume-detail-original-modal-content",
    );
    expect(modal).toHaveTextContent("Original resume parsed text snapshot");
    expect(modal).toHaveTextContent(
      "Led platform release guardrail work across frontend surfaces",
    );
    // The detail view loads the resume exactly once; the modal reuses that data.
    expect(getResumeSpy).toHaveBeenCalledTimes(1);
  });

  it("persists accepted rewrites into structuredProfile sections when overwriting", async () => {
    const client = buildClient("default");
    const resumeWithBullet = {
      ...(getResumeFixture.scenarios.default.response.body as Resume),
      structuredProfile: {
        headline: "Senior frontend engineer",
        sections: [
          {
            title: "Experience",
            bullets: ["Led design-system migration.", "Kept unrelated bullet."],
          },
        ],
      },
    } satisfies Resume;
    const updateSpy = vi.spyOn(client, "updateResume").mockResolvedValueOnce({
      ...resumeWithBullet,
      structuredProfile: {
        ...resumeWithBullet.structuredProfile,
        sections: [
          {
            title: "Experience",
            bullets: [
              "Led design-system migration across 12 teams; reduced UI defect rate by 38% over 6 weeks.",
              "Kept unrelated bullet.",
            ],
          },
        ],
      },
    } as Resume);
    vi.spyOn(client, "getResume").mockResolvedValue(resumeWithBullet);

    renderDetailWithClient(client, {
      name: "resume_versions",
      params: {
        resumeId: RESUME_ID,
        tab: "rewrites",
        tailorRunId: getResumeTailorRunFixture.scenarios.default.response.body.id,
      },
    });

    await waitFor(() => {
      expect(screen.getByTestId("resume-rewrites-action-accept")).toBeEnabled();
    }, { timeout: 3000 });
    const user = userEvent.setup();
    await user.click(screen.getByTestId("resume-rewrites-action-accept"));
    await user.click(screen.getByTestId("resume-rewrites-preview-save"));
    await user.click(screen.getByTestId("resume-rewrites-save-confirm"));

    await waitFor(() => {
      expect(updateSpy).toHaveBeenCalledTimes(1);
    });
    const body = updateSpy.mock.calls[0]![1] as Record<string, unknown>;
    expect(body.structuredProfile).toMatchObject({
      sections: [
        {
          title: "Experience",
          bullets: [
            "Led design-system migration across 12 teams; reduced UI defect rate by 38% over 6 weeks.",
            "Kept unrelated bullet.",
          ],
        },
      ],
    });
    expect(
      JSON.stringify(body.structuredProfile),
    ).not.toContain("acceptedRewrites");
  });

  it("persists accepted rewrites into flat experience, experiences, and projects bullets", async () => {
    const client = buildClient("default");
    const resumeWithFlatBullets = {
      ...(getResumeFixture.scenarios.default.response.body as Resume),
      structuredProfile: {
        headline: "Senior frontend engineer",
        experience: [
          {
            company: "Acme",
            bullets: ["Led design-system migration.", "Kept unrelated bullet."],
          },
        ],
        experiences: [
          {
            company: "Legacy Acme",
            bullets: ["Led design-system migration."],
          },
        ],
        projects: [
          {
            name: "Design System",
            bullets: ["Led design-system migration."],
          },
        ],
      },
    } satisfies Resume;
    const updateSpy = vi
      .spyOn(client, "updateResume")
      .mockResolvedValueOnce(resumeWithFlatBullets);
    vi.spyOn(client, "getResume").mockResolvedValue(resumeWithFlatBullets);

    renderDetailWithClient(client, {
      name: "resume_versions",
      params: {
        resumeId: RESUME_ID,
        tab: "rewrites",
        tailorRunId: getResumeTailorRunFixture.scenarios.default.response.body.id,
      },
    });

    await waitFor(() => {
      expect(screen.getByTestId("resume-rewrites-action-accept")).toBeEnabled();
    }, { timeout: 3000 });
    const user = userEvent.setup();
    await user.click(screen.getByTestId("resume-rewrites-action-accept"));
    await user.click(screen.getByTestId("resume-rewrites-preview-save"));
    await user.click(screen.getByTestId("resume-rewrites-save-confirm"));

    await waitFor(() => {
      expect(updateSpy).toHaveBeenCalledTimes(1);
    });
    const body = updateSpy.mock.calls[0]![1] as Record<string, unknown>;
    expect(body.structuredProfile).toMatchObject({
      experience: [
        {
          company: "Acme",
          bullets: [
            "Led design-system migration across 12 teams; reduced UI defect rate by 38% over 6 weeks.",
            "Kept unrelated bullet.",
          ],
        },
      ],
      experiences: [
        {
          company: "Legacy Acme",
          bullets: [
            "Led design-system migration across 12 teams; reduced UI defect rate by 38% over 6 weeks.",
          ],
        },
      ],
      projects: [
        {
          name: "Design System",
          bullets: [
            "Led design-system migration across 12 teams; reduced UI defect rate by 38% over 6 weeks.",
          ],
        },
      ],
    });
  });

  it("does not crash the original preview fallback when structuredProfile is omitted", async () => {
    const client = buildClient("default");
    const resumeWithoutTextOrProfile = {
      ...(getResumeFixture.scenarios.default.response.body as Resume),
      parsedTextSnapshot: null,
      originalText: null,
    };
    delete resumeWithoutTextOrProfile.structuredProfile;
    vi.spyOn(client, "getResume").mockResolvedValueOnce(
      resumeWithoutTextOrProfile,
    );

    renderDetailWithClient(client, {
      name: "resume_versions",
      params: { resumeId: RESUME_ID, tab: "preview" },
    });

    await waitFor(() => {
      expect(
        screen.getByTestId("resume-detail-view-original"),
      ).toBeInTheDocument();
    });
    await userEvent
      .setup()
      .click(screen.getByTestId("resume-detail-view-original"));

    expect(
      screen.getByTestId("resume-detail-original-modal-content"),
    ).toBeInTheDocument();
  });

  it("shows a retryable detail error for non-404 getResume failures", async () => {
    const client = buildClient("default");
    const getResumeSpy = vi
      .spyOn(client, "getResume")
      .mockRejectedValueOnce(new Error("HTTP 500 fixture outage"))
      .mockResolvedValueOnce(
        getResumeFixture.scenarios.default.response.body as Resume,
      );

    renderDetailWithClient(client, {
      name: "resume_versions",
      params: { resumeId: RESUME_ID, tab: "preview" },
    });

    await waitFor(() => {
      expect(screen.getByTestId("resume-detail-error")).toBeInTheDocument();
    });

    await userEvent.setup().click(screen.getByTestId("resume-detail-retry"));

    await waitFor(() => {
      expect(getResumeSpy).toHaveBeenCalledTimes(2);
      expect(screen.getByTestId("resume-detail-crumb")).toBeInTheDocument();
    });
  });

  it("renders NotFoundEmptyState when getResume returns 404", async () => {
    renderDetail("not-found", {
      name: "resume_versions",
      params: { resumeId: RESUME_ID, tab: "preview" },
    });

    await waitFor(() => {
      expect(
        screen.getByTestId("resume-detail-not-found"),
      ).toBeInTheDocument();
    });
    expect(
      screen.getByTestId("resume-detail-not-found-back"),
    ).toBeInTheDocument();
  });
});
