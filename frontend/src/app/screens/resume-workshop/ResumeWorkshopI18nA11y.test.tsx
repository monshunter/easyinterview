// @vitest-environment jsdom
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { render, screen, waitFor, within } from "@testing-library/react";

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
import getResumeFixture from "../../../../../openapi/fixtures/Resumes/getResume.json";

const FIXTURES = [
  getRuntimeConfigFixture,
  getMeFixture,
  listResumesFixture,
  getResumeFixture,
];

const RESUME_ID = getResumeFixture.scenarios.default.response.body.id;

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
      expect(screen.getByTestId("resume-workshop-card-grid")).toBeInTheDocument();
    });
    const list = within(screen.getByTestId("resume-workshop-list"));
    expect(list.getByText("Resume Workshop")).toBeInTheDocument();
    expect(screen.getByTestId("resume-workshop-create")).toHaveTextContent(
      "New resume",
    );
  });

  it("ZH locale renders Chinese copy on the list view", async () => {
    const client = buildClientWithSpy("default", "zh");
    renderScreen(client, { name: "resume_versions", params: {} });

    await waitFor(() => {
      expect(screen.getByTestId("resume-workshop-card-grid")).toBeInTheDocument();
    });
    const list = within(screen.getByTestId("resume-workshop-list"));
    expect(list.getByText("简历工坊")).toBeInTheDocument();
    expect(screen.getByTestId("resume-workshop-create")).toHaveTextContent(
      "新建简历",
    );
  });

  it("listResumes request carries Accept-Language header derived from the active lang", async () => {
    const client = buildClientWithSpy("default", "en");
    renderScreen(client, { name: "resume_versions", params: {} });

    await waitFor(() => {
      expect(screen.getByTestId("resume-workshop-card-grid")).toBeInTheDocument();
    });

    const listCall = recorded.find((req) => req.url.endsWith("/resumes"));
    expect(listCall, "expected a listResumes call").toBeTruthy();
    expect(listCall!.headers["Accept-Language"]).toBe("en");
  });

  it("getResume request also carries Accept-Language", async () => {
    const client = buildClientWithSpy("default", "zh");
    renderScreen(client, {
      name: "resume_versions",
      params: { resumeId: RESUME_ID, tab: "preview" },
    });

    await waitFor(() => {
      expect(screen.getByTestId("resume-detail-crumb")).toBeInTheDocument();
    });
    const getCall = recorded.find((req) =>
      req.url.endsWith(`/resumes/${RESUME_ID}`),
    );
    expect(getCall, "expected a getResume call").toBeTruthy();
    expect(getCall!.headers["Accept-Language"]).toBe("zh");
  });

  it("the card grid exposes list items and distinct card actions for assistive tech", async () => {
    const client = buildClientWithSpy("default");
    renderScreen(client, { name: "resume_versions", params: {} });
    await waitFor(() => {
      expect(screen.getByTestId("resume-workshop-card-grid")).toBeInTheDocument();
    });
    const list = within(screen.getByTestId("resume-workshop-card-grid"));
    expect(list.getAllByRole("listitem")).toHaveLength(2);
    expect(list.queryByRole("row")).not.toBeInTheDocument();
    expect(list.queryByRole("columnheader")).not.toBeInTheDocument();
    expect(list.getAllByRole("button", { name: /^(打开|Open) / })).toHaveLength(2);
    expect(
      list.getAllByRole("button", { name: /^(删除简历|Delete resume) / }),
    ).toHaveLength(2);
  });

  it("Detail view exposes the read-only resume article without tab semantics", async () => {
    const client = buildClientWithSpy("default");
    renderScreen(client, {
      name: "resume_versions",
      params: { resumeId: RESUME_ID, tab: "preview" },
    });
    await waitFor(() => {
      expect(screen.getByTestId("resume-detail-crumb")).toBeInTheDocument();
    });
    expect(screen.queryByRole("tablist")).not.toBeInTheDocument();
    expect(screen.queryByRole("tab")).not.toBeInTheDocument();
    expect(screen.getByTestId("resume-detail-preview-content")).toBeInTheDocument();
  });

  it("Detail view does not expose a separate original-preview modal trigger", async () => {
    const client = buildClientWithSpy("default");
    renderScreen(client, {
      name: "resume_versions",
      params: { resumeId: RESUME_ID, tab: "preview" },
    });
    await waitFor(() => {
      expect(screen.getByTestId("resume-detail-crumb")).toBeInTheDocument();
    });
    expect(screen.queryByTestId("resume-detail-view-original")).not.toBeInTheDocument();
    expect(screen.queryByTestId("resume-detail-original-modal")).not.toBeInTheDocument();
  });
});
