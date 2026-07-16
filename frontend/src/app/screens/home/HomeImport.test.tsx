// @vitest-environment jsdom
import { describe, expect, it, vi } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

import { createFixtureBackedFetch, createFixtureRegistry } from "../../../api/mockTransport";
import { EasyInterviewClient } from "../../../api/generated/client";
import { DisplayPreferencesProvider } from "../../display/DisplayPreferencesProvider";
import type { Lang } from "../../i18n/messages";
import { NavigationProvider } from "../../navigation/NavigationProvider";
import { AppRuntimeProvider } from "../../runtime/AppRuntimeProvider";
import { HomeScreen } from "./HomeScreen";

import getRuntimeConfigFixture from "../../../../../openapi/fixtures/Auth/getRuntimeConfig.json";
import getMeFixture from "../../../../../openapi/fixtures/Auth/getMe.json";
import listResumesFixture from "../../../../../openapi/fixtures/Resumes/listResumes.json";
import importTargetJobFixture from "../../../../../openapi/fixtures/TargetJobs/importTargetJob.json";

type ListResumesResponse = Awaited<ReturnType<EasyInterviewClient["listResumes"]>>;
type RuntimeConfigResponse = Awaited<ReturnType<EasyInterviewClient["getRuntimeConfig"]>>;

const defaultListResumesResponse = listResumesFixture.scenarios.default.response
  .body as ListResumesResponse;
const defaultRuntimeConfigResponse = getRuntimeConfigFixture.scenarios.default.response
  .body as RuntimeConfigResponse;
const RESUME_ID = "01918fa0-0000-7000-8000-000000001000";

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
  vi.spyOn(client, "getRuntimeConfig").mockResolvedValue(defaultRuntimeConfigResponse);
  vi.spyOn(client, "listResumes").mockResolvedValue(defaultListResumesResponse);
  return client;
}

async function selectDefaultResume() {
  await screen.findByTestId(
    "home-resume-option-01918fa0-0000-7000-8000-000000001000",
  );
  await userEvent.selectOptions(
    screen.getByTestId("home-resume-select"),
    RESUME_ID,
  );
}

function renderHome(client: EasyInterviewClient, options?: { lang?: Lang }) {
  const navigate = vi.fn();
  return {
    navigate,
    ...render(
      <DisplayPreferencesProvider initial={{ lang: options?.lang ?? "zh" }}>
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
      </DisplayPreferencesProvider>,
    ),
  };
}

describe("HomeImport — paste-only", () => {
  it("submits the exact flattened request with trimmed raw text", async () => {
    const client = createClient("paste-primary");
    const spy = vi.spyOn(client, "importTargetJob");
    const uploadSpy = vi.spyOn(client, "createUploadPresign");

    renderHome(client);
    await selectDefaultResume();

    await userEvent.type(
      screen.getByTestId("home-jd-textarea"),
      "  Senior Frontend Engineer needed  ",
    );
    await userEvent.click(screen.getByTestId("home-jd-submit"));

    await waitFor(() => {
      expect(spy).toHaveBeenCalledTimes(1);
    });

    const callBody = spy.mock.calls[0]?.[0];
    expect(callBody).toEqual({
      rawText: "Senior Frontend Engineer needed",
      resumeId: RESUME_ID,
      targetLanguage: "zh-CN",
    });
    expect(uploadSpy).not.toHaveBeenCalled();

    const callOpts = spy.mock.calls[0]?.[1];
    expect(callOpts?.idempotencyKey).toBeTruthy();
    expect(typeof callOpts?.idempotencyKey).toBe("string");
  });

  it("uses the current English UI locale as targetLanguage", async () => {
    const client = createClient("paste-primary");
    const spy = vi.spyOn(client, "importTargetJob");

    renderHome(client, { lang: "en" });
    await selectDefaultResume();

    await userEvent.type(
      screen.getByTestId("home-jd-textarea"),
      "Senior Frontend Engineer needed",
    );
    await userEvent.click(screen.getByTestId("home-jd-submit"));

    await waitFor(() => {
      expect(spy).toHaveBeenCalledTimes(1);
    });

    expect(spy.mock.calls[0]?.[0]).toMatchObject({
      targetLanguage: "en",
    });
  });

  it("blocks whitespace-only raw text before dispatch", async () => {
    const client = createClient("paste-primary");
    const spy = vi.spyOn(client, "importTargetJob");

    renderHome(client);
    await selectDefaultResume();

    await userEvent.type(screen.getByTestId("home-jd-textarea"), "   \n  ");

    expect(screen.getByTestId("home-jd-submit")).toBeDisabled();
    expect(spy).not.toHaveBeenCalled();
  });

  it("navigates to parse with targetJobId as the sole command locator", async () => {
    const client = createClient("paste-primary");
    const { navigate } = renderHome(client);
    await selectDefaultResume();

    await userEvent.type(
      screen.getByTestId("home-jd-textarea"),
      "Senior Frontend Engineer needed",
    );
    await userEvent.click(screen.getByTestId("home-jd-submit"));

    await waitFor(() => {
      expect(navigate).toHaveBeenCalledWith({
        name: "parse",
        params: {
          targetJobId: "01918fa0-0000-7000-8000-000000002001",
        },
      });
    });
  });

  it("renders a user-safe import failure and keeps the paste ready to retry", async () => {
    const client = createClient("paste-primary");
    const importError = new Error("HTTP 422: VALIDATION_FAILED");
    const importSpy = vi
      .spyOn(client, "importTargetJob")
      .mockRejectedValue(importError);
    const { navigate } = renderHome(client);
    await selectDefaultResume();

    const textarea = screen.getByTestId("home-jd-textarea");
    await userEvent.type(textarea, "Senior Frontend Engineer needed");
    await userEvent.click(screen.getByTestId("home-jd-submit"));

    expect(await screen.findByTestId("home-import-error")).toHaveTextContent(
      "面试规划创建失败，请稍后再试。",
    );
    expect(screen.queryByText("HTTP 422: VALIDATION_FAILED")).not.toBeInTheDocument();
    expect(textarea).toHaveValue("Senior Frontend Engineer needed");
    expect(screen.getByTestId("home-jd-submit")).toBeEnabled();
    expect(importSpy).toHaveBeenCalledTimes(1);
    expect(navigate).not.toHaveBeenCalled();
  });
});

describe("HomeImport — privacy", () => {
  const jdText = "Senior Frontend Engineer needed for design system team";

  it("does not log JD raw text to console", async () => {
    const client = createClient("paste-primary");
    const logSpy = vi.spyOn(console, "log").mockImplementation(() => {});

    renderHome(client);
    await selectDefaultResume();

    await userEvent.type(screen.getByTestId("home-jd-textarea"), jdText);
    await userEvent.click(screen.getByTestId("home-jd-submit"));

    await waitFor(() => {
      expect(screen.getByTestId("home-jd-textarea")).toBeInTheDocument();
    });

    const logCalls = logSpy.mock.calls.flat().join(" ");
    expect(logCalls).not.toContain(jdText);
    expect(logCalls).not.toContain("rawText");
    expect(logCalls).not.toContain("rawDescription");

    logSpy.mockRestore();
  });

  it("does not put JD raw text in localStorage", async () => {
    const client = createClient("paste-primary");
    const setItemSpy = vi.spyOn(Storage.prototype, "setItem");

    renderHome(client);
    await selectDefaultResume();

    await userEvent.type(screen.getByTestId("home-jd-textarea"), jdText);
    await userEvent.click(screen.getByTestId("home-jd-submit"));

    await waitFor(() => {
      expect(screen.getByTestId("home-jd-textarea")).toBeInTheDocument();
    });

    for (const call of setItemSpy.mock.calls) {
      const val = call[1] ?? "";
      expect(val).not.toContain(jdText);
    }

    setItemSpy.mockRestore();
  });

  it("excludes JD raw text from navigation params", async () => {
    const client = createClient("paste-primary");
    const { navigate } = renderHome(client);
    await selectDefaultResume();

    await userEvent.type(screen.getByTestId("home-jd-textarea"), jdText);
    await userEvent.click(screen.getByTestId("home-jd-submit"));

    await waitFor(() => {
      expect(navigate).toHaveBeenCalled();
    });

    const navCall = navigate.mock.calls[0]?.[0];
    const serialized = JSON.stringify(navCall);

    expect(serialized).not.toContain(jdText);
    expect(serialized).not.toContain("rawText");
    expect(serialized).not.toContain("rawDescription");
    expect(navCall.params?.rawText).toBeUndefined();
    expect(navCall.params?.rawDescription).toBeUndefined();
  });

  it("mockTransport spy does not record request bodies", async () => {
    const client = createClient("paste-primary");
    const importSpy = vi.spyOn(client, "importTargetJob");

    renderHome(client);
    await selectDefaultResume();

    await userEvent.type(
      screen.getByTestId("home-jd-textarea"),
      "Secret principal engineer role",
    );
    await userEvent.click(screen.getByTestId("home-jd-submit"));

    await waitFor(() => {
      expect(importSpy).toHaveBeenCalled();
    });

    const callBody = importSpy.mock.calls[0]?.[0];
    expect(callBody).toBeTruthy();
  });
});
