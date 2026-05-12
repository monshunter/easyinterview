// @vitest-environment jsdom
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { act, render, screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

import { EasyInterviewClient } from "../../../api/generated/client";
import {
  createFixtureBackedFetch,
  createFixtureRegistry,
} from "../../../api/mockTransport";
import { DisplayPreferencesProvider } from "../../display/DisplayPreferencesProvider";
import { NavigationProvider } from "../../navigation/NavigationProvider";
import { AppRuntimeProvider } from "../../runtime/AppRuntimeProvider";
import type { Route } from "../../routes";
import { ResumeWorkshopScreen } from "./ResumeWorkshopScreen";

import getRuntimeConfigFixture from "../../../../../openapi/fixtures/Auth/getRuntimeConfig.json";
import getMeFixture from "../../../../../openapi/fixtures/Auth/getMe.json";
import listResumesFixture from "../../../../../openapi/fixtures/Resumes/listResumes.json";
import listResumeVersionsFixture from "../../../../../openapi/fixtures/Resumes/listResumeVersions.json";
import getResumeVersionFixture from "../../../../../openapi/fixtures/Resumes/getResumeVersion.json";

const FIXTURES = [
  getRuntimeConfigFixture,
  getMeFixture,
  listResumesFixture,
  listResumeVersionsFixture,
  getResumeVersionFixture,
];

const FIRST_ASSET_ID =
  listResumesFixture.scenarios.default.response.body.items[0]?.id ?? "";
const VERSION_ID =
  getResumeVersionFixture.scenarios.default.response.body.id;

interface RecordedRequest {
  url: string;
  headers: Record<string, string>;
}

let recorded: RecordedRequest[] = [];

beforeEach(() => {
  recorded = [];
  window.localStorage.removeItem("ei-lang");
});

afterEach(() => {
  vi.unstubAllGlobals();
  window.localStorage.removeItem("ei-lang");
});

function buildClientWithSpy(
  scenario: string,
  initialLang?: string,
): EasyInterviewClient {
  if (initialLang) {
    window.localStorage.setItem("ei-lang", initialLang);
  }
  const baseFetch = createFixtureBackedFetch(
    createFixtureRegistry(FIXTURES),
    { scenario },
  );
  const fetchSpy = (
    input: RequestInfo | URL,
    init?: RequestInit,
  ): Promise<Response> => {
    const url = typeof input === "string" ? input : input.toString();
    const headers = (init?.headers as Record<string, string>) ?? {};
    recorded.push({ url, headers });
    return baseFetch(input, init);
  };
  return new EasyInterviewClient({ fetch: fetchSpy });
}

function renderScreen(client: EasyInterviewClient, route: Route) {
  return render(
    <DisplayPreferencesProvider>
      <AppRuntimeProvider
        client={client}
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

describe("ResumeWorkshop i18n + Accept-Language + a11y (Phase 4)", () => {
  it("EN locale renders English copy on the list view", async () => {
    const client = buildClientWithSpy("default", "en");
    renderScreen(client, { name: "resume_versions", params: {} });

    await waitFor(() => {
      expect(
        screen.getByTestId("resume-workshop-stats-originals"),
      ).toBeInTheDocument();
    });
    const originals = screen.getByTestId("resume-workshop-stats-originals");
    expect(originals).toHaveTextContent("Originals");
    expect(
      screen.getByTestId("resume-workshop-view-switcher-tree"),
    ).toHaveTextContent("Group by original");
  });

  it("ZH locale renders Chinese copy on the list view", async () => {
    const client = buildClientWithSpy("default", "zh");
    renderScreen(client, { name: "resume_versions", params: {} });

    await waitFor(() => {
      expect(
        screen.getByTestId("resume-workshop-stats-originals"),
      ).toBeInTheDocument();
    });
    const originals = screen.getByTestId("resume-workshop-stats-originals");
    expect(originals).toHaveTextContent("原始简历");
    expect(
      screen.getByTestId("resume-workshop-view-switcher-tree"),
    ).toHaveTextContent("按原始分组");
  });

  it("listResumes request carries Accept-Language header derived from the active lang", async () => {
    const client = buildClientWithSpy("default", "en");
    renderScreen(client, { name: "resume_versions", params: {} });

    await waitFor(() => {
      expect(
        screen.getByTestId("resume-workshop-stats-originals"),
      ).toBeInTheDocument();
    });

    const listCall = recorded.find((req) => req.url.endsWith("/resumes"));
    expect(listCall, "expected a listResumes call").toBeTruthy();
    expect(listCall!.headers["Accept-Language"]).toBe("en");
  });

  it("getResumeVersion request also carries Accept-Language", async () => {
    const client = buildClientWithSpy("default", "zh");
    renderScreen(client, {
      name: "resume_versions",
      params: { versionId: VERSION_ID, tab: "preview" },
    });

    await waitFor(() => {
      expect(
        screen.getByTestId("resume-detail-breadcrumb"),
      ).toBeInTheDocument();
    });
    const getCall = recorded.find((req) =>
      req.url.includes(`/resume-versions/${VERSION_ID}`) &&
      !req.url.endsWith("/exports"),
    );
    expect(getCall, "expected a getResumeVersion call").toBeTruthy();
    expect(getCall!.headers["Accept-Language"]).toBe("zh");
  });

  it("ViewSwitcher buttons expose role=tab and aria-selected reflecting the active group", async () => {
    const client = buildClientWithSpy("default");
    renderScreen(client, { name: "resume_versions", params: {} });
    await waitFor(() => {
      expect(
        screen.getByTestId("resume-workshop-view-switcher-tree"),
      ).toBeInTheDocument();
    });
    const tree = screen.getByTestId("resume-workshop-view-switcher-tree");
    const flat = screen.getByTestId("resume-workshop-view-switcher-flat");
    expect(tree).toHaveAttribute("role", "tab");
    expect(flat).toHaveAttribute("role", "tab");
    expect(tree).toHaveAttribute("aria-selected", "true");
    expect(flat).toHaveAttribute("aria-selected", "false");

    await userEvent.setup().click(flat);
    await waitFor(() =>
      expect(flat).toHaveAttribute("aria-selected", "true"),
    );
    expect(tree).toHaveAttribute("aria-selected", "false");
  });

  it("Tree row toggle exposes aria-expanded that flips on click (keyboard accessible)", async () => {
    const client = buildClientWithSpy("default");
    renderScreen(client, { name: "resume_versions", params: {} });
    await waitFor(() => {
      expect(
        screen.getByTestId(`resume-tree-row-${FIRST_ASSET_ID}-toggle`),
      ).toBeInTheDocument();
    });
    const toggle = screen.getByTestId(
      `resume-tree-row-${FIRST_ASSET_ID}-toggle`,
    );
    expect(toggle).toHaveAttribute("aria-expanded", "true");
    expect(toggle.tagName.toLowerCase()).toBe("button");

    await act(async () => {
      toggle.focus();
    });
    await userEvent.setup().keyboard(" ");
    await waitFor(() =>
      expect(toggle).toHaveAttribute("aria-expanded", "false"),
    );
  });

  it("Detail view tabs expose role=tab and aria-selected and the breadcrumb has an aria-label", async () => {
    const client = buildClientWithSpy("default");
    renderScreen(client, {
      name: "resume_versions",
      params: { versionId: VERSION_ID, tab: "preview" },
    });
    await waitFor(() => {
      expect(
        screen.getByTestId("resume-detail-breadcrumb"),
      ).toBeInTheDocument();
    });
    expect(screen.getByTestId("resume-detail-breadcrumb")).toHaveAttribute(
      "aria-label",
    );
    const previewTab = screen.getByTestId("resume-detail-tab-preview");
    expect(previewTab).toHaveAttribute("role", "tab");
    expect(previewTab).toHaveAttribute("aria-selected", "true");
    const rewritesTab = screen.getByTestId("resume-detail-tab-rewrites");
    expect(rewritesTab).toHaveAttribute("aria-selected", "false");
  });

  it("clicking View original opens the modal with focus on the close button (focus management)", async () => {
    const client = buildClientWithSpy("default");
    renderScreen(client, {
      name: "resume_versions",
      params: { versionId: VERSION_ID, tab: "preview" },
    });
    await waitFor(() => {
      expect(
        screen.getByTestId("resume-detail-view-original"),
      ).toBeInTheDocument();
    });

    await userEvent.setup().click(
      screen.getByTestId("resume-detail-view-original"),
    );
    const dialog = await screen.findByTestId("resume-detail-original-modal");
    const closeBtn = within(dialog).getByTestId(
      "resume-detail-original-modal-close",
    );
    await waitFor(() => {
      expect(document.activeElement).toBe(closeBtn);
    });
  });
});
