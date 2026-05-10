// @vitest-environment jsdom
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import type { MockInstance } from "vitest";
import { render, screen, waitFor, fireEvent } from "@testing-library/react";
import type { ReactNode } from "react";

import { EasyInterviewClient } from "../../../api/generated/client";
import {
  createFixtureBackedFetch,
  createFixtureRegistry,
} from "../../../api/mockTransport";
import { AppRuntimeProvider } from "../../runtime/AppRuntimeProvider";
import { DisplayPreferencesProvider } from "../../display/DisplayPreferencesProvider";
import { NavigationProvider } from "../../navigation/NavigationProvider";

import getJobMatchProfileFixture from "../../../../../openapi/fixtures/JobMatch/getJobMatchProfile.json";
import getAgentScanStatusFixture from "../../../../../openapi/fixtures/JobMatch/getAgentScanStatus.json";
import listJobRecommendationsFixture from "../../../../../openapi/fixtures/JobMatch/listJobRecommendations.json";
import getMeFixture from "../../../../../openapi/fixtures/Auth/getMe.json";
import getRuntimeConfigFixture from "../../../../../openapi/fixtures/Auth/getRuntimeConfig.json";

import { JDMatchScreen } from "./JDMatchScreen";

function buildClient(opts: { signedIn: boolean }) {
  const registry = createFixtureRegistry([
    getJobMatchProfileFixture,
    getAgentScanStatusFixture,
    listJobRecommendationsFixture,
    getMeFixture,
    getRuntimeConfigFixture,
  ]);
  return new EasyInterviewClient({
    fetch: async (input, init) => {
      const url =
        typeof input === "string" ? input : (input as URL | Request).toString();
      const inner = createFixtureBackedFetch(registry, undefined);
      const headers = new Headers(init?.headers ?? {});
      if (url.includes("/me")) {
        headers.set(
          "Prefer",
          opts.signedIn ? "example=authenticated" : "example=unauthenticated",
        );
      }
      return inner(input, { ...init, headers });
    },
  });
}

function wrap(ui: ReactNode, opts: { signedIn: boolean }) {
  const client = buildClient(opts);
  const navigate = vi.fn();
  const tree = (
    <DisplayPreferencesProvider initial={{ lang: "en" }}>
      <AppRuntimeProvider client={client}>
        <NavigationProvider value={{ navigate }}>{ui}</NavigationProvider>
      </AppRuntimeProvider>
    </DisplayPreferencesProvider>
  );
  return { navigate, ...render(tree) };
}

let openSpy: MockInstance<typeof window.open>;
let toastSpy: ReturnType<typeof vi.fn>;

beforeEach(() => {
  openSpy = vi.spyOn(window, "open").mockImplementation(() => null);
  toastSpy = vi.fn();
  (window as unknown as { eiToast?: typeof toastSpy }).eiToast = toastSpy;
});

afterEach(() => {
  openSpy.mockRestore();
  delete (window as unknown as { eiToast?: unknown }).eiToast;
  vi.restoreAllMocks();
});

describe("JDMatchAuthGate — pending-action surface (item 3.7)", () => {
  it("Confirm interview while unauthenticated → navigate(auth_login) with encoded pendingAction (type=jd_match_action, action=confirm_interview)", async () => {
    const { navigate } = wrap(
      <JDMatchScreen route={{ name: "jd_match", params: {} }} />,
      { signedIn: false },
    );
    const confirmBtn = await screen.findByTestId(
      "jdmatch-detail-action-confirm",
    );
    fireEvent.click(confirmBtn);
    await waitFor(() => expect(navigate).toHaveBeenCalledTimes(1));
    const call = navigate.mock.calls[0]![0];
    expect(call.name).toBe("auth_login");
    const params = call.params as Record<string, string>;
    expect(params.pendingType).toBe("jd_match_action");
    expect(params.pendingRoute).toBe("jd_match");
    expect(params.action).toBe("confirm_interview");
    expect(params.selectedJobMatchId).toBe(
      "01918fa0-0000-7000-8000-00000000a001",
    );
    expect(params.tab).toBe("recommended");
    // pendingAction must NOT carry private fields
    expect(params.query).toBeUndefined();
    expect(params.freeNote).toBeUndefined();
    expect(params.sourceUrl).toBeUndefined();
    expect(params.label).toBeUndefined();
  });

  it("Save while unauthenticated → pendingAction.action='save'", async () => {
    const { navigate } = wrap(
      <JDMatchScreen route={{ name: "jd_match", params: {} }} />,
      { signedIn: false },
    );
    const saveBtn = await screen.findByTestId("jdmatch-detail-action-save");
    fireEvent.click(saveBtn);
    await waitFor(() => expect(navigate).toHaveBeenCalled());
    const params = navigate.mock.calls[0]![0].params as Record<string, string>;
    expect(params.pendingType).toBe("jd_match_action");
    expect(params.action).toBe("save");
  });

  it("Mark not relevant while unauthenticated → pendingAction.action='dismiss'", async () => {
    const { navigate } = wrap(
      <JDMatchScreen route={{ name: "jd_match", params: {} }} />,
      { signedIn: false },
    );
    const dismissBtn = await screen.findByTestId(
      "jdmatch-detail-action-dismiss",
    );
    fireEvent.click(dismissBtn);
    await waitFor(() => expect(navigate).toHaveBeenCalled());
    const params = navigate.mock.calls[0]![0].params as Record<string, string>;
    expect(params.action).toBe("dismiss");
  });

  it("Source button while unauthenticated does NOT trigger pendingAction (still calls window.open)", async () => {
    const { navigate } = wrap(
      <JDMatchScreen route={{ name: "jd_match", params: {} }} />,
      { signedIn: false },
    );
    const sourceBtn = await screen.findByTestId(
      "jdmatch-detail-action-source",
    );
    fireEvent.click(sourceBtn);
    await waitFor(() => expect(openSpy).toHaveBeenCalledTimes(1));
    // Source path must NOT enter pendingAction surface
    for (const call of navigate.mock.calls) {
      const name = call[0].name;
      expect(name).not.toBe("auth_login");
    }
  });

  it("Saved-state Unsave click while unauthenticated → pendingAction.action='unsave'", async () => {
    const { navigate } = wrap(
      <JDMatchScreen route={{ name: "jd_match", params: {} }} />,
      { signedIn: false },
    );
    // Select the second card (saved=true in the default fixture)
    const secondCard = await screen.findByTestId(
      "jdmatch-card-01918fa0-0000-7000-8000-00000000a002",
    );
    fireEvent.click(secondCard);
    fireEvent.click(screen.getByTestId("jdmatch-detail-action-save"));
    await waitFor(() => expect(navigate).toHaveBeenCalled());
    const params = navigate.mock.calls.at(-1)![0].params as Record<
      string,
      string
    >;
    expect(params.action).toBe("unsave");
    expect(params.selectedJobMatchId).toBe(
      "01918fa0-0000-7000-8000-00000000a002",
    );
  });

  it("Confirm interview while authenticated → navigates directly to parse (no auth detour)", async () => {
    const { navigate } = wrap(
      <JDMatchScreen route={{ name: "jd_match", params: {} }} />,
      { signedIn: true },
    );
    const confirmBtn = await screen.findByTestId(
      "jdmatch-detail-action-confirm",
    );
    fireEvent.click(confirmBtn);
    await waitFor(() => expect(navigate).toHaveBeenCalled());
    const call = navigate.mock.calls[0]![0];
    expect(call.name).toBe("parse");
    expect((call.params as Record<string, string>).source).toBe("jd_match");
  });
});
