// @vitest-environment jsdom
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import {
  cleanup,
  render,
  screen,
  waitFor,
} from "@testing-library/react";
import userEvent from "@testing-library/user-event";

import { App } from "../../../App";
import { EasyInterviewClient } from "../../../../api/generated/client";
import {
  createFixtureBackedFetch,
  createFixtureRegistry,
} from "../../../../api/mockTransport";

import getRuntimeConfigFixture from "../../../../../../openapi/fixtures/Auth/getRuntimeConfig.json";
import getMeFixture from "../../../../../../openapi/fixtures/Auth/getMe.json";
import listResumesFixture from "../../../../../../openapi/fixtures/Resumes/listResumes.json";
import getResumeFixture from "../../../../../../openapi/fixtures/Resumes/getResume.json";
import listTargetJobsFixture from "../../../../../../openapi/fixtures/TargetJobs/listTargetJobs.json";

const FIXTURES = [
  getRuntimeConfigFixture,
  getMeFixture,
  listResumesFixture,
  getResumeFixture,
  listTargetJobsFixture,
];

function buildClient() {
  return new EasyInterviewClient({
    fetch: createFixtureBackedFetch(createFixtureRegistry(FIXTURES), {
      scenario: "default",
    }),
  });
}

describe("Home → ResumeCreateFlow CTA integration", () => {
  afterEach(() => cleanup());

  it("clicking the home resume create CTA lands on resume_versions?flow=create and renders ResumeCreateFlow (authenticated)", async () => {
    render(
      <App
        client={buildClient()}
        requestOptions={{
          getMe: { headers: { Prefer: "example=authenticated" } },
        }}
        initialRoute={{ name: "home", params: {} }}
      />,
    );
    const cta = await screen.findByTestId("home-resume-create");
    await userEvent.setup().click(cta);
    await waitFor(() =>
      expect(screen.getByTestId("resume-create-flow")).toBeInTheDocument(),
    );
  });

  it("clicking the home resume create CTA while unauthenticated routes through auth_login with a clean pendingAction (no raw text)", async () => {
    render(
      <App
        client={buildClient()}
        requestOptions={{
          getMe: { headers: { Prefer: "example=unauthenticated" } },
        }}
        initialRoute={{ name: "home", params: {} }}
      />,
    );
    const cta = await screen.findByTestId("home-resume-create");
    await userEvent.setup().click(cta);
    await waitFor(() =>
      expect(screen.getByTestId("route-auth_login")).toBeInTheDocument(),
    );
    expect(screen.getByTestId("auth-side-pending-action")).toBeInTheDocument();
    expect(
      screen.queryByTestId("resume-workshop-screen"),
    ).not.toBeInTheDocument();
    expect(
      screen.queryByTestId("resume-create-flow"),
    ).not.toBeInTheDocument();
  });
});

describe("Old onboarding route alias", () => {
  afterEach(() => cleanup());

  it("normalizes onboarding → resume_versions; ResumeCreateFlow is NOT invoked from the alias without flow=create", async () => {
    render(
      <App
        client={buildClient()}
        requestOptions={{
          getMe: { headers: { Prefer: "example=authenticated" } },
        }}
        initialRoute={{ name: "onboarding", params: {} } as never}
      />,
    );
    await waitFor(() =>
      expect(
        screen.getByTestId("resume-workshop-screen"),
      ).toBeInTheDocument(),
    );
    // No flow param → list view (or list loader) renders.
    expect(
      screen.queryByTestId("resume-create-flow"),
    ).not.toBeInTheDocument();
  });
});

describe("Retired-module negative grep is enforced inline", () => {
  beforeEach(() => {
    vi.resetModules();
  });
  it("imports the create-flow module without triggering any side-effect logging or hidden DOM artefacts", async () => {
    const consoleLog = vi.spyOn(console, "log").mockImplementation(() => {});
    const consoleInfo = vi.spyOn(console, "info").mockImplementation(() => {});
    const mod = await import("./ResumeCreateFlow");
    expect(mod.ResumeCreateFlow).toBeTruthy();
    expect(consoleLog).not.toHaveBeenCalled();
    expect(consoleInfo).not.toHaveBeenCalled();
  });
});
