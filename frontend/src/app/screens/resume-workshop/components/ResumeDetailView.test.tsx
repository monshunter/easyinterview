// @vitest-environment jsdom
import { describe, expect, it, vi } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
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

const FIXTURES = [getRuntimeConfigFixture, getMeFixture, getResumeFixture];

function buildClient(scenario: string): EasyInterviewClient {
  return new EasyInterviewClient({
    fetch: createFixtureBackedFetch(createFixtureRegistry(FIXTURES), {
      scenario,
    }),
  });
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

function renderDetail(
  scenario: string,
  route: Route,
  nav: ReturnType<typeof vi.fn> = vi.fn(),
) {
  return renderDetailWithClient(buildClient(scenario), route, nav);
}

const RESUME_ID = getResumeFixture.scenarios.default.response.body.id;

describe("ResumeDetailView read-only contract", () => {
  it("renders the resume itself with no secondary edit, export, copy, or original-preview controls", async () => {
    renderDetail("default", {
      name: "resume_versions",
      params: { resumeId: RESUME_ID },
    });

    await waitFor(() => {
      expect(screen.getByTestId("resume-detail-crumb")).toBeInTheDocument();
    });
    expect(screen.getByTestId("resume-detail-preview-content")).toHaveTextContent(
      "Original resume parsed text snapshot",
    );
    expect(screen.getByTestId("resume-detail-preview-content")).toHaveTextContent(
      "Led platform release guardrail work across frontend surfaces",
    );
    expect(screen.getByTestId("resume-detail-preview-content")).not.toHaveTextContent(
      "Senior frontend engineer for platform-heavy product teams",
    );
    expect(screen.queryByRole("tablist")).not.toBeInTheDocument();
    for (const forbidden of [
      "resume-detail-tab-preview",
      "resume-detail-tab-rewrites",
      "resume-detail-tab-edit",
      "resume-detail-header-actions",
      "resume-detail-export-pdf",
      "resume-detail-copy-text",
      "resume-detail-view-original",
      "resume-detail-original-modal",
      "resume-rewrites-tab",
      "resume-edit-tab",
    ]) {
      expect(screen.queryByTestId(forbidden)).not.toBeInTheDocument();
    }
  });

  it("ignores legacy tab=rewrites and still shows the same read-only resume content", async () => {
    renderDetail("default", {
      name: "resume_versions",
      params: { resumeId: RESUME_ID, tab: "rewrites" },
    });

    await waitFor(() => {
      expect(
        screen.getByTestId("resume-detail-preview-content"),
      ).toBeInTheDocument();
    });
    expect(screen.queryByTestId("resume-rewrites-tab")).not.toBeInTheDocument();
    expect(screen.queryByTestId("resume-detail-tab-rewrites")).not.toBeInTheDocument();
    expect(screen.getByTestId("resume-workshop-detail")).not.toHaveAttribute(
      "data-tab",
    );
  });

  it("does not invoke exportResume, requestResumeTailor, or updateResume while loading a detail", async () => {
    const client = buildClient("default");
    const exportSpy = vi.spyOn(client, "exportResume");
    const tailorSpy = vi.spyOn(client, "requestResumeTailor");
    const updateSpy = vi.spyOn(client, "updateResume");

    renderDetailWithClient(client, {
      name: "resume_versions",
      params: { resumeId: RESUME_ID, tab: "edit" },
    });

    await waitFor(() => {
      expect(screen.getByTestId("resume-detail-crumb")).toBeInTheDocument();
    });
    expect(exportSpy).not.toHaveBeenCalled();
    expect(tailorSpy).not.toHaveBeenCalled();
    expect(updateSpy).not.toHaveBeenCalled();
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
      params: { resumeId: RESUME_ID },
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
      params: { resumeId: RESUME_ID },
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
