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
import importTargetJobFixture from "../../../../../openapi/fixtures/TargetJobs/importTargetJob.json";
import createUploadPresignFixture from "../../../../../openapi/fixtures/Uploads/createUploadPresign.json";

function createClient(scenario?: string) {
  const fetch = createFixtureBackedFetch(
    createFixtureRegistry([
      getRuntimeConfigFixture,
      getMeFixture,
      importTargetJobFixture,
      createUploadPresignFixture,
    ]),
    scenario ? { scenario } : undefined,
  );
  return new EasyInterviewClient({ fetch });
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

describe("HomeImport — paste (manual_text)", () => {
  it("calls importTargetJob with manual_text discriminator on Submit", async () => {
    const client = createClient("manual-text-primary");
    const spy = vi.spyOn(client, "importTargetJob");

    renderHome(client);

    await userEvent.type(
      screen.getByTestId("home-jd-textarea"),
      "Senior Frontend Engineer needed",
    );
    screen.getByTestId("home-jd-submit").click();

    await waitFor(() => {
      expect(spy).toHaveBeenCalledTimes(1);
    });

    const callBody = spy.mock.calls[0]?.[0];
    expect(callBody).toMatchObject({
      source: {
        type: "manual_text",
        rawText: "Senior Frontend Engineer needed",
      },
      targetLanguage: "zh-CN",
    });

    const callOpts = spy.mock.calls[0]?.[1];
    expect(callOpts?.idempotencyKey).toBeTruthy();
    expect(typeof callOpts?.idempotencyKey).toBe("string");
  });

  it("uses the current English UI locale as targetLanguage", async () => {
    const client = createClient("manual-text-primary");
    const spy = vi.spyOn(client, "importTargetJob");

    renderHome(client, { lang: "en" });

    await userEvent.type(
      screen.getByTestId("home-jd-textarea"),
      "Senior Frontend Engineer needed",
    );
    screen.getByTestId("home-jd-submit").click();

    await waitFor(() => {
      expect(spy).toHaveBeenCalledTimes(1);
    });

    expect(spy.mock.calls[0]?.[0]).toMatchObject({
      targetLanguage: "en",
    });
  });

  it("navigates to parse on successful paste import", async () => {
    const client = createClient("manual-text-primary");
    const { navigate } = renderHome(client);

    await userEvent.type(
      screen.getByTestId("home-jd-textarea"),
      "Senior Frontend Engineer needed",
    );
    screen.getByTestId("home-jd-submit").click();

    await waitFor(() => {
      expect(navigate).toHaveBeenCalledWith(
        expect.objectContaining({
          name: "parse",
          params: expect.objectContaining({
            targetJobId: "01918fa0-0000-7000-8000-000000002001",
            source: "paste",
          }),
        }),
      );
    });
  });
});

describe("HomeImport — url import", () => {
  it("opens URL modal and calls importTargetJob with url discriminator", async () => {
    const client = createClient("default");
    const spy = vi.spyOn(client, "importTargetJob");

    renderHome(client);

    screen.getByText("URL").click();

    const urlInput = await screen.findByTestId("home-modal-url-input");
    await userEvent.type(urlInput, "https://acme.example/careers/senior");

    screen.getByTestId("home-modal-url-continue").click();

    await waitFor(() => {
      expect(spy).toHaveBeenCalledTimes(1);
    });

    const callBody = spy.mock.calls[0]?.[0];
    expect(callBody).toMatchObject({
      source: {
        type: "url",
        url: "https://acme.example/careers/senior",
      },
      targetLanguage: "zh-CN",
    });

    const callOpts = spy.mock.calls[0]?.[1];
    expect(callOpts?.idempotencyKey).toBeTruthy();
  });

  it("navigates to parse on successful url import", async () => {
    const client = createClient("default");
    const { navigate } = renderHome(client);

    screen.getByText("URL").click();
    const urlInput = await screen.findByTestId("home-modal-url-input");
    await userEvent.type(urlInput, "https://acme.example/careers/senior");

    screen.getByTestId("home-modal-url-continue").click();

    await waitFor(() => {
      expect(navigate).toHaveBeenCalledWith(
        expect.objectContaining({
          name: "parse",
          params: expect.objectContaining({ source: "url" }),
        }),
      );
    });
  });

  it("shows inline error on 4xx import response", async () => {
    const client = createClient("url-invalid-source");
    renderHome(client);

    screen.getByText("URL").click();
    const urlInput = await screen.findByTestId("home-modal-url-input");
    await userEvent.type(urlInput, "http://192.0.2.1/job");
    screen.getByTestId("home-modal-url-continue").click();

    await waitFor(() => {
      expect(screen.getByTestId("home-import-error")).toBeInTheDocument();
    });
  });
});

describe("HomeImport — upload flow", () => {
  it("calls createUploadPresign then importTargetJob on upload modal confirm", async () => {
    const client = createClient("default");
    const presignSpy = vi.spyOn(client, "createUploadPresign");
    const importSpy = vi.spyOn(client, "importTargetJob");

    renderHome(client);

    const uploadBtn = screen.getByTestId("home-upload-trigger");
    uploadBtn.click();

    const continueBtn = await screen.findByTestId("home-modal-upload-continue");
    continueBtn.click();

    await waitFor(() => {
      expect(presignSpy).toHaveBeenCalled();
    });

    const presignBody = presignSpy.mock.calls[0]?.[0];
    expect(presignBody).toMatchObject({
      purpose: "target_job_attachment",
      fileName: expect.stringMatching(/^placeholder\.pdf$/),
      contentType: "application/pdf",
      byteSize: 0,
    });

    await waitFor(() => {
      expect(importSpy).toHaveBeenCalled();
    });

    const importBody = importSpy.mock.calls[0]?.[0];
    expect(importBody).toMatchObject({
      source: {
        type: "file",
        fileObjectId: "01918fa0-0000-7000-8000-000000001100",
      },
      targetLanguage: "zh-CN",
    });
  });
});

describe("HomeImport — privacy", () => {
  const jdText = "Senior Frontend Engineer needed for design system team";

  it("does not log JD raw text to console", async () => {
    const client = createClient("manual-text-primary");
    const logSpy = vi.spyOn(console, "log").mockImplementation(() => {});

    renderHome(client);

    await userEvent.type(screen.getByTestId("home-jd-textarea"), jdText);
    screen.getByTestId("home-jd-submit").click();

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
    const client = createClient("manual-text-primary");
    const setItemSpy = vi.spyOn(Storage.prototype, "setItem");

    renderHome(client);

    await userEvent.type(screen.getByTestId("home-jd-textarea"), jdText);
    screen.getByTestId("home-jd-submit").click();

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
    const client = createClient("manual-text-primary");
    const { navigate } = renderHome(client);

    await userEvent.type(screen.getByTestId("home-jd-textarea"), jdText);
    screen.getByTestId("home-jd-submit").click();

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
    const client = createClient("default");
    const importSpy = vi.spyOn(client, "importTargetJob");

    renderHome(client);

    screen.getByText("URL").click();
    const urlInput = await screen.findByTestId("home-modal-url-input");
    await userEvent.type(urlInput, "https://example.com/jd/secret-role");
    screen.getByTestId("home-modal-url-continue").click();

    await waitFor(() => {
      expect(importSpy).toHaveBeenCalled();
    });

    const callBody = importSpy.mock.calls[0]?.[0];
    expect(callBody).toBeTruthy();
  });
});
