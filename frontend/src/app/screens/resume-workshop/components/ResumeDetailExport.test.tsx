// @vitest-environment jsdom
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { render, screen, waitFor, within } from "@testing-library/react";
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
import getResumeFixture from "../../../../../../openapi/fixtures/Resumes/getResume.json";
import exportResumeFixture from "../../../../../../openapi/fixtures/Resumes/exportResume.json";

const FIXTURES = [
  getRuntimeConfigFixture,
  getMeFixture,
  getResumeFixture,
  exportResumeFixture,
];

const RESUME_ID = getResumeFixture.scenarios.default.response.body.id;

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

describe("exportResume P0 fallback (Phase 3.7)", () => {
  it("clicking Export PDF on the preview tab passes a generated Idempotency-Key header to the request and surfaces the not-available toast", async () => {
    let capturedHeaders: Record<string, string> | null = null;
    const baseFetch = createFixtureBackedFetch(
      createFixtureRegistry(FIXTURES),
      { scenario: "default" },
    );
    const fetchSpy = (
      input: RequestInfo | URL,
      init?: RequestInit,
    ): Promise<Response> => {
      const url = typeof input === "string" ? input : input.toString();
      if (url.includes("/exports")) {
        const headers = init?.headers as Record<string, string> | undefined;
        capturedHeaders = headers ?? {};
      }
      return baseFetch(input, init);
    };

    const client = new EasyInterviewClient({ fetch: fetchSpy });
    const exportSpy = vi.spyOn(client, "exportResume");

    const route: Route = {
      name: "resume_versions",
      params: { resumeId: RESUME_ID, tab: "preview" },
    };

    render(
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

    // Export PDF appears in both header and preview tab; click the preview one.
    const previewContent = await screen.findByTestId(
      "resume-detail-preview-content",
    );
    const exportBtn = within(previewContent).getByTestId(
      "resume-detail-export-pdf",
    );
    await userEvent.setup().click(exportBtn);

    await waitFor(() => {
      expect(exportSpy).toHaveBeenCalled();
    });

    const args = exportSpy.mock.calls[0]!;
    expect(args[0]).toBe(RESUME_ID);
    const opts = args[1] as { idempotencyKey?: string } | undefined;
    expect(opts?.idempotencyKey).toMatch(/^v1\.\d+\./);

    expect(capturedHeaders).not.toBeNull();
    expect(capturedHeaders!["Idempotency-Key"]).toBeTruthy();
    expect(capturedHeaders!["Idempotency-Key"]).toMatch(/^v1\.\d+\./);

    await waitFor(() => {
      expect(
        toastCalls.some((call) =>
          /即将开放|not available|P0/i.test(call.message),
        ),
      ).toBe(true);
    });

    // Privacy red line: no blob URL, no localStorage write.
    const localStorageOffenders: string[] = [];
    for (let i = 0; i < window.localStorage.length; i++) {
      const key = window.localStorage.key(i);
      if (key && /resume|export|pdf/i.test(key)) {
        localStorageOffenders.push(key);
      }
    }
    expect(localStorageOffenders).toEqual([]);
  });
});
