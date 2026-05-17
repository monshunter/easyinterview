// @vitest-environment jsdom
import { describe, expect, it, vi } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

import { EasyInterviewClient } from "../../../../api/generated/client";
import type { ResumeAsset, ResumeVersion } from "../../../../api/generated/types";
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
import getResumeVersionFixture from "../../../../../../openapi/fixtures/Resumes/getResumeVersion.json";
import exportResumeVersionFixture from "../../../../../../openapi/fixtures/Resumes/exportResumeVersion.json";

const FIXTURES = [
  getRuntimeConfigFixture,
  getMeFixture,
  getResumeVersionFixture,
  exportResumeVersionFixture,
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

const TARGETED_VERSION_ID =
  getResumeVersionFixture.scenarios.default.response.body.id;
const MASTER_VERSION_ID =
  getResumeVersionFixture.scenarios["master-default"].response.body.id;

describe("ResumeDetailView container (Phase 3.1)", () => {
  it("renders breadcrumb, branch graph, and three tabs once the version loads", async () => {
    renderDetail("default", {
      name: "resume_versions",
      params: { versionId: TARGETED_VERSION_ID, tab: "preview" },
    });

    await waitFor(() => {
      expect(
        screen.getByTestId("resume-detail-breadcrumb"),
      ).toBeInTheDocument();
    });
    expect(
      screen.getByTestId("resume-detail-branch-graph"),
    ).toBeInTheDocument();
    expect(screen.getByTestId("resume-detail-tab-preview")).toBeInTheDocument();
    expect(screen.getByTestId("resume-detail-tab-rewrites")).toBeInTheDocument();
    expect(screen.getByTestId("resume-detail-tab-edit")).toBeInTheDocument();
  });

  it("MASTER version (master-default scenario) defaults the active tab to preview", async () => {
    renderDetail("master-default", {
      name: "resume_versions",
      params: { versionId: MASTER_VERSION_ID },
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

  it("TARGETED version with explicit tab=rewrites in the URL keeps rewrites active and renders ResumeRewritesTab (plan 003 replaces ComingSoonTab)", async () => {
    renderDetail("default", {
      name: "resume_versions",
      params: { versionId: TARGETED_VERSION_ID, tab: "rewrites" },
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
    expect(
      screen.queryByTestId("resume-detail-tab-content-coming-soon-rewrites"),
    ).not.toBeInTheDocument();
  });

  it("clicking a tab updates the active selection and shows that tab's content (preview from default route)", async () => {
    renderDetail("default", {
      name: "resume_versions",
      params: { versionId: TARGETED_VERSION_ID, tab: "rewrites" },
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

  it("loads the original resume asset for the original modal instead of reusing structuredProfile copy", async () => {
    const client = buildClient("default");
    const originalAsset: ResumeAsset = {
      id: "01918fa0-0000-7000-8000-000000001000",
      title: "Original uploaded resume",
      language: "zh-CN",
      parseStatus: "ready",
      sourceType: "upload",
      status: "active",
      parsedSummary: { headline: "Original asset headline" },
      parsedTextSnapshot:
        "Original asset line A\nOriginal asset line B from parsed snapshot",
      originalText: "Raw original asset fallback",
      createdAt: "2026-04-22T09:30:00Z",
      updatedAt: "2026-04-28T12:00:00Z",
    };
    const getResumeSpy = vi
      .spyOn(client, "getResume")
      .mockResolvedValue(originalAsset);

    renderDetailWithClient(client, {
      name: "resume_versions",
      params: { versionId: TARGETED_VERSION_ID, tab: "preview" },
    });

    await waitFor(() => {
      expect(
        screen.getByTestId("resume-detail-view-original"),
      ).toBeInTheDocument();
    });
    await userEvent.setup().click(screen.getByTestId("resume-detail-view-original"));

    await waitFor(() => {
      expect(getResumeSpy).toHaveBeenCalledWith(
        originalAsset.id,
        expect.objectContaining({
          headers: expect.objectContaining({ "Accept-Language": "en" }),
        }),
      );
    });
    const modal = await screen.findByTestId("resume-detail-original-modal-content");
    expect(modal).toHaveTextContent("Original asset line A");
    expect(modal).toHaveTextContent("Original asset line B from parsed snapshot");
    expect(modal).not.toHaveTextContent(
      "Highlights reliability, cross-functional tradeoff work",
    );
  });

  it("shows a retryable detail error for non-404 getResumeVersion failures", async () => {
    const client = buildClient("default");
    const getVersionSpy = vi
      .spyOn(client, "getResumeVersion")
      .mockRejectedValueOnce(new Error("HTTP 500 fixture outage"))
      .mockResolvedValueOnce(
        getResumeVersionFixture.scenarios.default.response.body as ResumeVersion,
      );

    renderDetailWithClient(client, {
      name: "resume_versions",
      params: { versionId: TARGETED_VERSION_ID, tab: "preview" },
    });

    await waitFor(() => {
      expect(screen.getByTestId("resume-detail-error")).toBeInTheDocument();
    });
    expect(screen.queryByText(/加载版本中|Loading version/i)).not.toBeInTheDocument();

    await userEvent.setup().click(screen.getByTestId("resume-detail-retry"));

    await waitFor(() => {
      expect(getVersionSpy).toHaveBeenCalledTimes(2);
      expect(
        screen.getByTestId("resume-detail-breadcrumb"),
      ).toBeInTheDocument();
    });
  });

  it("keeps the original modal in loading state instead of showing structuredProfile fallback while getResume is pending", async () => {
    const client = buildClient("default");
    const originalAsset: ResumeAsset = {
      id: "01918fa0-0000-7000-8000-000000001000",
      title: "Original uploaded resume",
      language: "zh-CN",
      parseStatus: "ready",
      sourceType: "upload",
      status: "active",
      parsedSummary: { headline: "Original asset headline" },
      parsedTextSnapshot: "Original asset line after pending resolves",
      originalText: "Raw original asset fallback",
      createdAt: "2026-04-22T09:30:00Z",
      updatedAt: "2026-04-28T12:00:00Z",
    };
    let resolveOriginal: (asset: ResumeAsset) => void = () => undefined;
    vi.spyOn(client, "getResume").mockReturnValue(
      new Promise<ResumeAsset>((resolve) => {
        resolveOriginal = resolve;
      }),
    );

    renderDetailWithClient(client, {
      name: "resume_versions",
      params: { versionId: TARGETED_VERSION_ID, tab: "preview" },
    });

    await waitFor(() => {
      expect(
        screen.getByTestId("resume-detail-view-original"),
      ).toBeInTheDocument();
    });
    await userEvent.setup().click(screen.getByTestId("resume-detail-view-original"));

    await waitFor(() => {
      expect(
        screen.getByTestId("resume-detail-original-modal-loading"),
      ).toBeInTheDocument();
    });
    expect(
      screen.getByTestId("resume-detail-original-modal-content"),
    ).not.toHaveTextContent(
      "Highlights reliability, cross-functional tradeoff work",
    );

    resolveOriginal(originalAsset);

    await waitFor(() => {
      expect(
        screen.getByTestId("resume-detail-original-modal-content"),
      ).toHaveTextContent("Original asset line after pending resolves");
    });
  });

  it("shows original-source errors without substituting structuredProfile fallback text", async () => {
    const client = buildClient("default");
    const getResumeSpy = vi
      .spyOn(client, "getResume")
      .mockRejectedValue(new Error("HTTP 500 source outage"));

    renderDetailWithClient(client, {
      name: "resume_versions",
      params: { versionId: TARGETED_VERSION_ID, tab: "preview" },
    });

    await waitFor(() => {
      expect(
        screen.getByTestId("resume-detail-view-original"),
      ).toBeInTheDocument();
    });
    await userEvent.setup().click(screen.getByTestId("resume-detail-view-original"));

    await waitFor(() => {
      expect(
        screen.getByTestId("resume-detail-original-modal-error"),
      ).toBeInTheDocument();
    });
    expect(getResumeSpy).toHaveBeenCalledTimes(1);
    expect(
      screen.getByTestId("resume-detail-original-modal-content"),
    ).not.toHaveTextContent(
      "Highlights reliability, cross-functional tradeoff work",
    );
  });
});
