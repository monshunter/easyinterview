// @vitest-environment jsdom
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

import { EasyInterviewClient } from "../../api/generated/client";
import {
  createFixtureBackedFetch,
  createFixtureRegistry,
} from "../../api/mockTransport";
import { App } from "../App";

import getRuntimeConfigFixture from "../../../../openapi/fixtures/Auth/getRuntimeConfig.json";
import getMeFixture from "../../../../openapi/fixtures/Auth/getMe.json";
import listResumesFixture from "../../../../openapi/fixtures/Resumes/listResumes.json";
import getResumeFixture from "../../../../openapi/fixtures/Resumes/getResume.json";

const FIXTURES = [
  getRuntimeConfigFixture,
  getMeFixture,
  listResumesFixture,
  getResumeFixture,
];

const FIRST_RESUME_ID =
  listResumesFixture.scenarios.default.response.body.items[0]?.id ?? "";
const SECOND_RESUME_ID =
  listResumesFixture.scenarios.default.response.body.items[1]?.id ?? "";

interface ToastCall {
  message: string;
  tone?: string;
}

let toastCalls: ToastCall[] = [];

beforeEach(() => {
  toastCalls = [];
  (
    window as unknown as {
      eiToast?: (msg: string, opts?: { tone?: string }) => void;
    }
  ).eiToast = (message, opts) => {
    toastCalls.push({ message, tone: opts?.tone });
  };
});

afterEach(() => {
  delete (
    window as unknown as {
      eiToast?: (msg: string, opts?: { tone?: string }) => void;
    }
  ).eiToast;
});

function buildClient(): EasyInterviewClient {
  return new EasyInterviewClient({
    fetch: createFixtureBackedFetch(
      createFixtureRegistry(FIXTURES),
      { scenario: "default" },
    ),
  });
}

function renderApp(authMode: "authenticated" | "unauthenticated") {
  return render(
    <App
      client={buildClient()}
      requestOptions={{
        getMe: { headers: { Prefer: `example=${authMode}` } },
      }}
      initialRoute={{ name: "resume_versions", params: {} }}
    />,
  );
}

describe("E2E.P0.036 resume list is flat (no tree toggle) + auth boundary", () => {
  it("unauthenticated visit to /resume_versions routes to login, not the list, and triggers no Resume API", async () => {
    const client = buildClient();
    const listSpy = vi.spyOn(client, "listResumes");
    const getSpy = vi.spyOn(client, "getResume");
    render(
      <App
        client={client}
        requestOptions={{
          getMe: { headers: { Prefer: "example=unauthenticated" } },
        }}
        initialRoute={{ name: "resume_versions", params: {} }}
      />,
    );

    await waitFor(() =>
      expect(screen.getByTestId("route-auth_login")).toBeInTheDocument(),
    );
    expect(screen.getByTestId("auth-side-pending-action")).toBeInTheDocument();
    expect(screen.queryByTestId("resume-workshop-list")).not.toBeInTheDocument();
    expect(
      screen.queryByTestId("resume-workshop-auth-gate"),
    ).not.toBeInTheDocument();
    expect(listSpy).not.toHaveBeenCalled();
    expect(getSpy).not.toHaveBeenCalled();
  });

  it("authenticated default view renders the flat ResumeWorkshopScreen list with one row per resume and NO tree / stats / view-switcher chrome", async () => {
    renderApp("authenticated");

    // The `resume-workshop-list` testid is shared by the loading and loaded
    // states; wait for the loaded flat table specifically so this assertion is
    // deterministic regardless of fixture-fetch timing under parallel workers.
    await waitFor(() => {
      expect(screen.getByTestId("resume-workshop-table")).toBeInTheDocument();
    });
    // Resume workshop screen replaces the placeholder.
    expect(
      screen.queryByTestId("route-resume_versions"),
    ).not.toBeInTheDocument();
    expect(screen.getByTestId("resume-workshop-screen")).toBeInTheDocument();
    expect(screen.getByTestId("resume-workshop-list")).toBeInTheDocument();

    // One flat row per resume, derived from the fixture body.
    expect(
      screen.getByTestId(`resume-list-row-${FIRST_RESUME_ID}`),
    ).toBeInTheDocument();
    expect(
      screen.getByTestId(`resume-list-row-${SECOND_RESUME_ID}`),
    ).toBeInTheDocument();

    // D-20 flatten: stats strip, view switcher, and version tree are gone.
    expect(
      screen.queryByTestId("resume-workshop-stats-originals"),
    ).not.toBeInTheDocument();
    expect(
      screen.queryByTestId("resume-workshop-stats-versions"),
    ).not.toBeInTheDocument();
    expect(
      screen.queryByTestId("resume-workshop-view-switcher-tree"),
    ).not.toBeInTheDocument();
    expect(
      screen.queryByTestId("resume-workshop-view-switcher-flat"),
    ).not.toBeInTheDocument();
    expect(
      screen.queryByTestId(`resume-tree-row-${FIRST_RESUME_ID}`),
    ).not.toBeInTheDocument();
    expect(
      screen.queryByTestId("resume-workshop-selected-tree-helper"),
    ).not.toBeInTheDocument();
    expect(
      toastCalls.some((c) => /即将开放|coming soon/i.test(c.message)),
    ).toBe(false);
  });

  it("opening a flat row navigates to that resume's detail with the preview tab as default", async () => {
    renderApp("authenticated");

    await waitFor(() => {
      expect(
        screen.getByTestId(`resume-list-open-${FIRST_RESUME_ID}`),
      ).toBeInTheDocument();
    });

    const user = userEvent.setup();
    await user.click(screen.getByTestId(`resume-list-open-${FIRST_RESUME_ID}`));

    await waitFor(() => {
      expect(screen.getByTestId("resume-workshop-detail")).toBeInTheDocument();
    });
    const detail = screen.getByTestId("resume-workshop-detail");
    expect(detail).toHaveAttribute("data-resume-id", FIRST_RESUME_ID);
    expect(detail).toHaveAttribute("data-tab", "preview");
  });

  it("retired prototype testids and runtime imports are absent from the resume-workshop source", () => {
    // Static gate already enforced in ResumeWorkshopPrivacy.test.ts. We assert
    // here that the rendered DOM never surfaces retired-route testids that
    // would indicate the legacy welcome / mistakes / drill / voice modules
    // sneaked back in.
    renderApp("authenticated");
    for (const forbidden of [
      "route-welcome",
      "route-mistakes",
      "route-drill",
      "route-followup",
      "route-onboarding",
      "route-experiences",
      "route-star",
      "route-voice",
    ]) {
      expect(screen.queryByTestId(forbidden)).not.toBeInTheDocument();
    }
  });
});
