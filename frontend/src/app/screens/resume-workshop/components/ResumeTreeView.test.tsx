// @vitest-environment jsdom
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

import { EasyInterviewClient } from "../../../../api/generated/client";
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
import listResumesFixture from "../../../../../../openapi/fixtures/Resumes/listResumes.json";
import listResumeVersionsFixture from "../../../../../../openapi/fixtures/Resumes/listResumeVersions.json";

const FIRST_ASSET_ID =
  listResumesFixture.scenarios.default.response.body.items[0]?.id ?? "";
const FIRST_VERSION_ID =
  listResumeVersionsFixture.scenarios.default.response.body.items[0]?.id ?? "";

function buildClient(): EasyInterviewClient {
  return new EasyInterviewClient({
    fetch: createFixtureBackedFetch(
      createFixtureRegistry([
        getRuntimeConfigFixture,
        getMeFixture,
        listResumesFixture,
        listResumeVersionsFixture,
      ]),
      { scenario: "default" },
    ),
  });
}

function renderTree() {
  const route: Route = { name: "resume_versions", params: {} };
  return render(
    <DisplayPreferencesProvider>
      <AppRuntimeProvider
        client={buildClient()}
        requestOptions={{
          getMe: { headers: { Prefer: "example=authenticated" } },
        }}
      >
        <NavigationProvider value={{ navigate: vi.fn() }}>
          <ResumeWorkshopScreen route={route} />
        </NavigationProvider>
      </AppRuntimeProvider>
    </DisplayPreferencesProvider>,
  );
}

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

describe("ResumeTreeView interactions", () => {
  it("renders the tree row toggle and collapses / expands the version list", async () => {
    renderTree();

    await waitFor(() => {
      expect(
        screen.getByTestId(`resume-tree-row-${FIRST_ASSET_ID}-toggle`),
      ).toBeInTheDocument();
    });
    // initially expanded — version row visible
    await waitFor(() => {
      expect(
        screen.getByTestId(`resume-version-row-${FIRST_VERSION_ID}`),
      ).toBeInTheDocument();
    });
    expect(
      screen.getByTestId(`resume-tree-row-${FIRST_ASSET_ID}-toggle`),
    ).toHaveAttribute("aria-expanded", "true");

    const user = userEvent.setup();
    await user.click(
      screen.getByTestId(`resume-tree-row-${FIRST_ASSET_ID}-toggle`),
    );

    await waitFor(() => {
      expect(
        screen.queryByTestId(`resume-version-row-${FIRST_VERSION_ID}`),
      ).not.toBeInTheDocument();
    });
    expect(
      screen.getByTestId(`resume-tree-row-${FIRST_ASSET_ID}-toggle`),
    ).toHaveAttribute("aria-expanded", "false");
  });

  it("renders the use-as-base / new-version buttons and their click fires the coming-soon toast", async () => {
    renderTree();

    await waitFor(() => {
      expect(
        screen.getByTestId(`resume-tree-row-${FIRST_ASSET_ID}-use-as-base`),
      ).toBeInTheDocument();
    });
    expect(
      screen.getByTestId(`resume-tree-row-${FIRST_ASSET_ID}-new-version`),
    ).toBeInTheDocument();

    const user = userEvent.setup();
    await user.click(
      screen.getByTestId(`resume-tree-row-${FIRST_ASSET_ID}-use-as-base`),
    );
    await user.click(
      screen.getByTestId(`resume-tree-row-${FIRST_ASSET_ID}-new-version`),
    );

    await waitFor(() => {
      expect(toastCalls.length).toBeGreaterThanOrEqual(2);
    });
    expect(
      toastCalls.some((call) => /即将开放|coming soon/i.test(call.message)),
    ).toBe(true);
  });
});
