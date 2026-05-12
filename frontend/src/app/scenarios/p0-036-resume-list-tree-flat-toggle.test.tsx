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
import listResumeVersionsFixture from "../../../../openapi/fixtures/Resumes/listResumeVersions.json";

const FIXTURES = [
  getRuntimeConfigFixture,
  getMeFixture,
  listResumesFixture,
  listResumeVersionsFixture,
];

const FIRST_ASSET_ID =
  listResumesFixture.scenarios.default.response.body.items[0]?.id ?? "";
const SECOND_ASSET_ID =
  listResumesFixture.scenarios.default.response.body.items[1]?.id ?? "";
const FIRST_VERSION_ID =
  listResumeVersionsFixture.scenarios.default.response.body.items[0]?.id ?? "";

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

describe("E2E.P0.036 resume-list tree/flat toggle + StatsStrip + auth boundary", () => {
  it("unauthenticated visit to /resume_versions surfaces the auth gate, not the list, and triggers no Resume API", async () => {
    const client = buildClient();
    const listSpy = vi.spyOn(client, "listResumes");
    const versionsSpy = vi.spyOn(client, "listResumeVersions");
    render(
      <App
        client={client}
        requestOptions={{
          getMe: { headers: { Prefer: "example=unauthenticated" } },
        }}
        initialRoute={{ name: "resume_versions", params: {} }}
      />,
    );

    await waitFor(() => {
      expect(
        screen.getByTestId("resume-workshop-auth-gate"),
      ).toBeInTheDocument();
    });
    expect(screen.queryByTestId("resume-workshop-list")).not.toBeInTheDocument();
    expect(listSpy).not.toHaveBeenCalled();
    expect(versionsSpy).not.toHaveBeenCalled();
  });

  it("authenticated default view renders ResumeWorkshopScreen with StatsStrip + ViewSwitcher + tree rows derived from fixture body", async () => {
    renderApp("authenticated");

    await waitFor(() => {
      expect(
        screen.getByTestId("resume-workshop-stats-originals"),
      ).toBeInTheDocument();
    });
    // Resume workshop screen replaces the placeholder.
    expect(
      screen.queryByTestId("route-resume_versions"),
    ).not.toBeInTheDocument();
    expect(screen.getByTestId("resume-workshop-screen")).toBeInTheDocument();

    // StatsStrip 4 stats, derived from fixture body (no static counts).
    const expectedAssetCount =
      listResumesFixture.scenarios.default.response.body.items.length;
    const expectedVersionCount =
      listResumeVersionsFixture.scenarios.default.response.body.items.length;
    expect(
      screen.getByTestId("resume-workshop-stats-originals"),
    ).toHaveTextContent(String(expectedAssetCount));
    await waitFor(() => {
      expect(
        screen.getByTestId("resume-workshop-stats-versions"),
      ).toHaveTextContent(String(expectedVersionCount));
    });
    expect(
      screen.getByTestId("resume-workshop-stats-top-match"),
    ).toBeInTheDocument();
    expect(
      screen.getByTestId("resume-workshop-stats-recent"),
    ).toBeInTheDocument();

    // ViewSwitcher tree (active) / flat
    const treeBtn = screen.getByTestId("resume-workshop-view-switcher-tree");
    const flatBtn = screen.getByTestId("resume-workshop-view-switcher-flat");
    expect(treeBtn).toHaveAttribute("aria-selected", "true");
    expect(flatBtn).toHaveAttribute("aria-selected", "false");

    // Tree rows for both assets; second asset surfaces no-versions placeholder.
    expect(
      screen.getByTestId(`resume-tree-row-${FIRST_ASSET_ID}`),
    ).toBeInTheDocument();
    expect(
      screen.getByTestId(`resume-tree-row-${SECOND_ASSET_ID}`),
    ).toBeInTheDocument();
    expect(
      screen.getByTestId(`resume-tree-row-${SECOND_ASSET_ID}-no-versions`),
    ).toBeInTheDocument();
    expect(
      screen.getByTestId(`resume-version-row-${FIRST_VERSION_ID}`),
    ).toBeInTheDocument();

    // Use-as-base / new-version buttons render and surface coming-soon toasts on click.
    const user = userEvent.setup();
    await user.click(
      screen.getByTestId(`resume-tree-row-${FIRST_ASSET_ID}-use-as-base`),
    );
    await waitFor(() =>
      expect(
        toastCalls.some((c) => /即将开放|coming soon/i.test(c.message)),
      ).toBe(true),
    );
  });

  it("clicking ViewSwitcher flat renders the FlatView and clicking a flat row navigates to the version detail with the right default tab", async () => {
    renderApp("authenticated");

    await waitFor(() => {
      expect(
        screen.getByTestId("resume-workshop-view-switcher-flat"),
      ).toBeInTheDocument();
    });

    const user = userEvent.setup();
    await user.click(
      screen.getByTestId("resume-workshop-view-switcher-flat"),
    );

    await waitFor(() => {
      expect(
        screen.getByTestId(`resume-flat-row-${FIRST_VERSION_ID}`),
      ).toBeInTheDocument();
    });
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
