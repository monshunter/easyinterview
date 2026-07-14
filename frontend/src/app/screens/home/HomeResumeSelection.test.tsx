// @vitest-environment jsdom
import { StrictMode } from "react";
import { describe, expect, it, vi } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

import { EasyInterviewClient } from "../../../api/generated/client";
import { createFixtureBackedFetch, createFixtureRegistry } from "../../../api/mockTransport";
import { DisplayPreferencesProvider } from "../../display/DisplayPreferencesProvider";
import { NavigationProvider } from "../../navigation/NavigationProvider";
import { AppRuntimeProvider } from "../../runtime/AppRuntimeProvider";
import { HomeScreen } from "./HomeScreen";

import getRuntimeConfigFixture from "../../../../../openapi/fixtures/Auth/getRuntimeConfig.json";
import getMeFixture from "../../../../../openapi/fixtures/Auth/getMe.json";
import listResumesFixture from "../../../../../openapi/fixtures/Resumes/listResumes.json";
import importTargetJobFixture from "../../../../../openapi/fixtures/TargetJobs/importTargetJob.json";
import listTargetJobsFixture from "../../../../../openapi/fixtures/TargetJobs/listTargetJobs.json";

type ListResumesResponse = Awaited<ReturnType<EasyInterviewClient["listResumes"]>>;

const defaultListResumesResponse = listResumesFixture.scenarios.default.response
  .body as ListResumesResponse;
const emptyListResumesResponse = listResumesFixture.scenarios.empty.response
  .body as ListResumesResponse;
const readableNonReadyListResumesResponse = {
  ...defaultListResumesResponse,
  items: [
    {
      ...defaultListResumesResponse.items[0]!,
      id: "01918fa0-0000-7000-8000-000000001101",
      title: "failed-readable.pdf",
      displayName: "Readable Failed Resume",
      parseStatus: "failed",
      sourceType: "upload",
      summaryHeadline: "Recovered PDF text",
      hasReadableContent: true,
      updatedAt: "2026-05-15T08:00:00Z",
    },
    {
      ...defaultListResumesResponse.items[0]!,
      id: "01918fa0-0000-7000-8000-000000001102",
      title: "Queued Paste Source",
      displayName: "Queued Paste Source",
      parseStatus: "queued",
      sourceType: "paste",
      summaryHeadline: "Queued paste resume",
      hasReadableContent: true,
      updatedAt: "2026-05-14T08:00:00Z",
    },
    {
      ...defaultListResumesResponse.items[0]!,
      id: "01918fa0-0000-7000-8000-000000001103",
      title: "Processing Markdown Source",
      displayName: "Processing Markdown Source",
      parseStatus: "processing",
      sourceType: "upload",
      summaryHeadline: "Processing markdown resume",
      hasReadableContent: true,
      updatedAt: "2026-05-13T08:00:00Z",
    },
  ],
} satisfies ListResumesResponse;

function createClient(scenario?: string) {
  const fetch = createFixtureBackedFetch(
    createFixtureRegistry([
      getRuntimeConfigFixture,
      getMeFixture,
      listResumesFixture,
      importTargetJobFixture,
    ]),
    scenario ? { scenario } : undefined,
  );
  const client = new EasyInterviewClient({ fetch });
  vi.spyOn(client, "listResumes").mockResolvedValue(
    scenario === "empty" ? emptyListResumesResponse : defaultListResumesResponse,
  );
  return client;
}

function renderHome(client: EasyInterviewClient, options?: { strict?: boolean }) {
  const navigate = vi.fn();
  const home = (
    <DisplayPreferencesProvider initial={{ lang: "zh" }}>
      <AppRuntimeProvider
        client={client}
        requestOptions={{
          getMe: { headers: { Prefer: "example=authenticated" } },
        }}
      >
        <NavigationProvider value={{ navigate }}>
          <HomeScreen route={{ name: "home", params: {} }} />
        </NavigationProvider>
      </AppRuntimeProvider>
    </DisplayPreferencesProvider>
  );
  return {
    navigate,
    ...render(options?.strict ? <StrictMode>{home}</StrictMode> : home),
  };
}

describe("Home resume selection", () => {
  it("issues one jobs GET and one resumes GET on an authenticated StrictMode mount", async () => {
    const fixtureFetch = createFixtureBackedFetch(
      createFixtureRegistry([
        getRuntimeConfigFixture,
        getMeFixture,
        listResumesFixture,
        listTargetJobsFixture,
      ]),
    );
    const requestCounts = new Map<string, number>();
    const fetch: typeof globalThis.fetch = async (input, init) => {
      const method = (init?.method ?? "GET").toUpperCase();
      const path = new URL(String(input), "http://easyinterview.local").pathname;
      const key = `${method} ${path}`;
      requestCounts.set(key, (requestCounts.get(key) ?? 0) + 1);
      return fixtureFetch(input, init);
    };
    const client = new EasyInterviewClient({ fetch });

    renderHome(client, { strict: true });

    await screen.findByTestId("home-resume-select");
    await screen.findByTestId(
      "home-recent-mock-card-01918fa0-0000-7000-8000-000000002000",
    );
    expect(requestCounts.get("GET /api/v1/resumes")).toBe(1);
    expect(requestCounts.get("GET /api/v1/targets")).toBe(1);
    console.info(
      "E2E.P0.014 Home StrictMode transport PASS listTargetJobs=1 listResumes=1",
    );
  });

  it("renders the home quick-start copy without the out-of-scope hero sub or CTA", async () => {
    const client = createClient("default");
    renderHome(client);

    expect(screen.queryByTestId("home-hero-sub")).not.toBeInTheDocument();
    expect(screen.getByTestId("home-jd-submit")).toHaveTextContent("立即面试");
    expect(screen.queryByText("解析并确认面试")).not.toBeInTheDocument();

    await waitFor(() => {
      expect(screen.getByTestId("home-resume-select")).toBeInTheDocument();
    });
    expect(screen.getByTestId("home-resume-select").tagName).toBe("SELECT");
    expect(screen.getByTestId("home-resume-select")).toHaveRole("combobox");
  });

  it("requires an explicit ready resume selection before importing a pasted JD", async () => {
    const client = createClient("paste-primary");
    const listSpy = vi.spyOn(client, "listResumes");
    const importSpy = vi.spyOn(client, "importTargetJob");
    const { navigate } = renderHome(client);

    await waitFor(() => {
      expect(listSpy).toHaveBeenCalledTimes(1);
    });

    await userEvent.type(
      screen.getByTestId("home-jd-textarea"),
      "Senior Frontend Engineer needed",
    );

    expect(screen.getByTestId("home-jd-submit")).toBeDisabled();
    expect(importSpy).not.toHaveBeenCalled();

    await screen.findByTestId(
      "home-resume-option-01918fa0-0000-7000-8000-000000001000",
    );
    const resumeSelect = screen.getByTestId("home-resume-select");
    expect(resumeSelect.tagName).toBe("SELECT");
    expect(
      screen.queryByRole("button", { name: /Alice Example/i }),
    ).not.toBeInTheDocument();

    await userEvent.selectOptions(
      resumeSelect,
      "01918fa0-0000-7000-8000-000000001000",
    );
    expect(resumeSelect).toHaveValue("01918fa0-0000-7000-8000-000000001000");
    expect(screen.getByTestId("home-jd-submit")).not.toBeDisabled();

    await userEvent.click(screen.getByTestId("home-jd-submit"));

    await waitFor(() => {
      expect(importSpy).toHaveBeenCalledTimes(1);
    });

    await waitFor(() => {
      expect(navigate).toHaveBeenCalledWith(
        expect.objectContaining({
          name: "parse",
          params: {
            targetJobId: "01918fa0-0000-7000-8000-000000002001",
          },
        }),
      );
    });
  });

  it("keeps readable existing resumes selectable when parseStatus is not ready", async () => {
    const client = createClient("default");
    vi.mocked(client.listResumes).mockResolvedValue(
      readableNonReadyListResumesResponse,
    );
    renderHome(client);

    const resumeSelect = await screen.findByTestId("home-resume-select");

    await waitFor(() => {
      expect(resumeSelect).not.toBeDisabled();
    });
    expect(screen.queryByTestId("home-resume-empty")).not.toBeInTheDocument();
    expect(
      await screen.findByTestId(
        "home-resume-option-01918fa0-0000-7000-8000-000000001101",
      ),
    ).toHaveTextContent("Readable Failed Resume");
    expect(
      screen.getByTestId(
        "home-resume-option-01918fa0-0000-7000-8000-000000001102",
      ),
    ).toHaveTextContent("Queued Paste Source");
    expect(
      screen.getByTestId(
        "home-resume-option-01918fa0-0000-7000-8000-000000001103",
      ),
    ).toHaveTextContent("Processing Markdown Source");

    await userEvent.selectOptions(
      resumeSelect,
      "01918fa0-0000-7000-8000-000000001101",
    );

    expect(resumeSelect).toHaveValue(
      "01918fa0-0000-7000-8000-000000001101",
    );
    expect(screen.getByTestId("home-resume-selection-status")).toHaveTextContent(
      "Readable Failed Resume",
    );
  });

  it("keeps immediate interview disabled when no ready resume exists and keeps the create CTA", async () => {
    const client = createClient("empty");
    const importSpy = vi.spyOn(client, "importTargetJob");
    const { navigate } = renderHome(client);

    await waitFor(() => {
      expect(screen.getByTestId("home-resume-empty")).toBeInTheDocument();
    });

    await userEvent.type(
      screen.getByTestId("home-jd-textarea"),
      "Senior Frontend Engineer needed",
    );
    await userEvent.click(screen.getByTestId("home-jd-submit"));

    expect(screen.getByTestId("home-jd-submit")).toBeDisabled();
    expect(importSpy).not.toHaveBeenCalled();

    await userEvent.click(screen.getByTestId("home-resume-create"));

    expect(navigate).toHaveBeenCalledWith({
      name: "resume_versions",
      params: { flow: "create" },
    });
  });

  it("does not expose the resume service error when resume loading fails", async () => {
    const client = createClient("default");
    vi.mocked(client.listResumes).mockRejectedValue(
      new Error("HTTP 503 RESUME_PROVIDER_UNAVAILABLE"),
    );

    renderHome(client);

    expect(await screen.findByText("简历暂时无法读取，请稍后重试。")).toBeInTheDocument();
    expect(screen.queryByText("HTTP 503 RESUME_PROVIDER_UNAVAILABLE")).not.toBeInTheDocument();
    expect(screen.getByTestId("home-jd-submit")).toBeDisabled();
  });
});
