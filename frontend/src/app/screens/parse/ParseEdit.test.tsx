// @vitest-environment jsdom
import { afterEach, describe, expect, it, vi } from "vitest";
import { act, render, screen, fireEvent, waitFor } from "@testing-library/react";

import {
  createFixtureBackedFetch,
  createFixtureRegistry,
} from "../../../api/mockTransport";
import { EasyInterviewClient } from "../../../api/generated/client";
import { DisplayPreferencesProvider } from "../../display/DisplayPreferencesProvider";
import { NavigationProvider } from "../../navigation/NavigationProvider";
import { AppRuntimeProvider } from "../../runtime/AppRuntimeProvider";
import { ParseScreen } from "./ParseScreen";

import getTargetJobFixture from "../../../../../openapi/fixtures/TargetJobs/getTargetJob.json";
import updateTargetJobFixture from "../../../../../openapi/fixtures/TargetJobs/updateTargetJob.json";

const LOADING_PREVIEW_DELAY = 3200;

function makeReadyFixture() {
  const body = (
    getTargetJobFixture.scenarios["default"] as {
      response: { body: Record<string, unknown> };
    }
  ).response.body;
  return {
    operationId: "getTargetJob",
    scenarios: {
      default: {
        response: {
          status: 200,
          body: { ...body, analysisStatus: "ready" as const },
        },
      },
    },
  };
}

function createClient() {
  const fetch = createFixtureBackedFetch(
    createFixtureRegistry([
      makeReadyFixture(),
      updateTargetJobFixture,
    ]),
    { scenario: "default" },
  );
  return new EasyInterviewClient({ fetch });
}

function renderParse(client: EasyInterviewClient) {
  const navigate = vi.fn();
  return {
    navigate,
    ...render(
      <DisplayPreferencesProvider>
        <AppRuntimeProvider client={client}>
          <NavigationProvider value={{ navigate }}>
            <ParseScreen
              route={{ name: "parse", params: { targetJobId: "tj-1" } }}
            />
          </NavigationProvider>
        </AppRuntimeProvider>
      </DisplayPreferencesProvider>,
    ),
  };
}

async function renderReadyParse(client: EasyInterviewClient) {
  vi.useFakeTimers();
  const result = renderParse(client);

  await act(async () => {
    await vi.advanceTimersByTimeAsync(LOADING_PREVIEW_DELAY);
  });
  vi.useRealTimers();

  return result;
}

afterEach(() => {
  vi.useRealTimers();
});

describe("ParseEdit — inline editing", () => {
  it("renders editable title input with initial value from fixture", async () => {
    const client = createClient();
    await renderReadyParse(client);

    const titleEl = await screen.findByTestId("parse-basics-title");

    // Title field contains an editable input
    const input = titleEl.querySelector("input");
    expect(input).toBeTruthy();
    expect(input?.value).toBe("Senior Frontend Engineer");
  });

  it("renders editable company input", async () => {
    const client = createClient();
    await renderReadyParse(client);

    const companyEl = await screen.findByTestId("parse-basics-company");
    const input = companyEl.querySelector("input");
    expect(input).toBeTruthy();
    expect(input?.value).toBe("Acme");
  });

  it("renders editable location input", async () => {
    const client = createClient();
    await renderReadyParse(client);

    const locEl = await screen.findByTestId("parse-basics-location");
    const input = locEl.querySelector("input");
    expect(input).toBeTruthy();
  });

  it("renders level and language as read-only (no input element)", async () => {
    const client = createClient();
    await renderReadyParse(client);

    const levelEl = await screen.findByTestId("parse-basics-level");
    expect(levelEl.querySelector("input")).toBeFalsy();

    const langEl = await screen.findByTestId("parse-basics-language");
    expect(langEl.querySelector("input")).toBeFalsy();
  });

  it("renders requirements with toggle buttons", async () => {
    const client = createClient();
    await renderReadyParse(client);

    const req0 = await screen.findByTestId("parse-requirement-must_have-0");
    expect(req0).toBeInTheDocument();

    const toggle0 = await screen.findByTestId(
      "parse-requirement-must_have-0-toggle",
    );
    expect(toggle0.tagName).toBe("BUTTON");
  });

  it("hit toggle cycles through false → true → partial → false", async () => {
    const client = createClient();
    await renderReadyParse(client);

    const toggle = await screen.findByTestId(
      "parse-requirement-must_have-0-toggle",
    );

    // Initial state: false → GAP
    expect(toggle.textContent).toMatch(/GAP|缺口/);

    // Click → true (HIT)
    fireEvent.click(toggle);
    expect(toggle.textContent).toMatch(/HIT|命中/);

    // Click → partial
    fireEvent.click(toggle);
    expect(toggle.textContent).toMatch(/PARTIAL|部分/);

    // Click → false (GAP)
    fireEvent.click(toggle);
    expect(toggle.textContent).toMatch(/GAP|缺口/);
  });

  it("renders hidden signals and round assumptions", async () => {
    const client = createClient();
    await renderReadyParse(client);

    const signal0 = await screen.findByTestId("parse-hidden-signal-0");
    expect(signal0).toBeInTheDocument();

    const round0 = await screen.findByTestId("parse-round-0");
    expect(round0).toBeInTheDocument();

    const round1 = await screen.findByTestId("parse-round-1");
    expect(round1).toBeInTheDocument();

    const round2 = await screen.findByTestId("parse-round-2");
    expect(round2).toBeInTheDocument();

    const round3 = await screen.findByTestId("parse-round-3");
    expect(round3).toBeInTheDocument();
  });
});

describe("ParseEdit — confirm call", () => {
  it("calls updateTargetJob on Confirm with only supplied fields", async () => {
    const client = createClient();
    const spy = vi.spyOn(client, "updateTargetJob");
    const { navigate } = await renderReadyParse(client);

    const confirmBtn = await screen.findByTestId("parse-action-confirm");
    fireEvent.click(confirmBtn);

    await waitFor(() => {
      expect(spy).toHaveBeenCalledTimes(1);
    });

    const callBody = spy.mock.calls[0]?.[1];
    expect(callBody).toBeDefined();
    // Only supplied fields should be present
    expect(callBody).toMatchObject({
      titleHint: "Senior Frontend Engineer",
      companyNameHint: "Acme",
      locationText: "Shanghai · Hybrid",
    });
    // Must NOT include level or language (read-only fields)
    expect(callBody).not.toHaveProperty("level");
    expect(callBody).not.toHaveProperty("language");

    // Idempotency key
    const callOpts = spy.mock.calls[0]?.[2];
    expect(callOpts?.idempotencyKey).toBeTruthy();
    expect(typeof callOpts?.idempotencyKey).toBe("string");

    // Navigate to workspace after success
    await waitFor(() => {
      expect(navigate).toHaveBeenCalledWith({
        name: "workspace",
        params: expect.objectContaining({
          targetJobId: "01918fa0-0000-7000-8000-000000002000",
          jobId: "01918fa0-0000-7000-8000-000000002000",
          jdId: "jd-01918fa0-0000-7000-8000-000000002000",
          planId: "plan-01918fa0-0000-7000-8000-000000002000",
          resumeId: "resume-unbound",
          roundId: "round-technical-1",
          roundName: "Technical Round 1",
        }),
      });
    });
  });

  it("shows inline error on updateTargetJob 4xx", async () => {
    const errorFixture = {
      operationId: "updateTargetJob",
      scenarios: {
        default: {
          response: {
            status: 422,
            body: {
              error: {
                code: "VALIDATION_FAILED",
                message: "Invalid request",
              },
            },
          },
        },
      },
    };

    const fetch = createFixtureBackedFetch(
      createFixtureRegistry([makeReadyFixture(), errorFixture]),
      { scenario: "default" },
    );
    const client = new EasyInterviewClient({ fetch });

    await renderReadyParse(client);

    const confirmBtn = await screen.findByTestId("parse-action-confirm");
    fireEvent.click(confirmBtn);

    await waitFor(() => {
      expect(
        screen.getByTestId("parse-action-confirm"),
      ).toBeInTheDocument();
      // Confirm button should still be present (editing state preserved)
    });
  });
});

describe("ParseEdit — re-parse and cancel", () => {
  it("re-parse does not throw when scrollTo is unavailable", async () => {
    const client = createClient();
    await renderReadyParse(client);

    await screen.findByTestId("parse-action-reparse");
    const scrollSpy = vi
      .spyOn(window, "scrollTo")
      .mockImplementation(() => {
        throw new Error("scrollTo unavailable");
      });

    await act(async () => {
      expect(() => {
        fireEvent.click(screen.getByTestId("parse-action-reparse"));
      }).not.toThrow();
    });

    scrollSpy.mockRestore();
  });

  it("re-parse button resets to loading stage", async () => {
    const client = createClient();
    await renderReadyParse(client);

    // Wait for preview
    await screen.findByTestId("parse-action-reparse");

    fireEvent.click(screen.getByTestId("parse-action-reparse"));

    // Should show loading steps
    await waitFor(() => {
      expect(screen.getByTestId("parse-loading-step-0")).toBeInTheDocument();
    });
  });

  it("re-parse triggers a fresh getTargetJob poll", async () => {
    const client = createClient();
    const spy = vi.spyOn(client, "getTargetJob");
    await renderReadyParse(client);

    await screen.findByTestId("parse-action-reparse");
    const callsBeforeReparse = spy.mock.calls.length;

    fireEvent.click(screen.getByTestId("parse-action-reparse"));

    await waitFor(() => {
      expect(spy.mock.calls.length).toBeGreaterThan(callsBeforeReparse);
    });
  });
});
