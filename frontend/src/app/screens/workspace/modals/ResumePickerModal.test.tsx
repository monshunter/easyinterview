/**
 * @vitest-environment jsdom
 */

import { describe, expect, it, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import type { ReactNode } from "react";

import { EasyInterviewClient } from "../../../../api/generated/client";
import {
  createFixtureBackedFetch,
  createFixtureRegistry,
} from "../../../../api/mockTransport";
import { DisplayPreferencesProvider } from "../../../display/DisplayPreferencesProvider";
import { InterviewContextProvider } from "../../../interview-context/InterviewContext";
import { NavigationProvider } from "../../../navigation/NavigationProvider";
import { AppRuntimeContext } from "../../../runtime/AppRuntimeProvider";
import { ResumePickerModal } from "./ResumePickerModal";

import listResumesFixture from "../../../../../../openapi/fixtures/Resumes/listResumes.json";

function buildClient(): EasyInterviewClient {
  return new EasyInterviewClient({
    fetch: createFixtureBackedFetch(
      createFixtureRegistry([listResumesFixture]),
      { scenario: "default" },
    ),
  });
}

function withProviders(ui: ReactNode, client = buildClient()) {
  const nav = vi.fn();
  return {
    client,
    nav,
    ...render(
      <DisplayPreferencesProvider>
        <InterviewContextProvider>
          <AppRuntimeContext.Provider
            value={{
              client,
              runtime: {
                status: "ready",
                config: {
                  analyticsEnabled: false,
                  appVersion: "test",
                  defaultUiLanguage: "zh",
                  featureFlags: {},
                },
              },
              auth: {
                status: "authenticated",
                user: {
                  id: "user-1",
                  emailMasked: "u***@example.com",
                  displayName: "User",
                  preferredPracticeLanguage: "zh",
                  profileCompletionRequired: false,
                  uiLanguage: "zh",
                },
              },
              refreshAuth: vi.fn(),
            }}
          >
            <NavigationProvider value={{ navigate: nav }}>
              {ui}
            </NavigationProvider>
          </AppRuntimeContext.Provider>
        </InterviewContextProvider>
      </DisplayPreferencesProvider>,
    ),
  };
}

const DEFAULT_RESUME_ID =
  listResumesFixture.scenarios.default.response.body.items[0]!.id;
const SECOND_RESUME_ID =
  listResumesFixture.scenarios.default.response.body.items[1]!.id;

describe("ResumePickerModal", () => {
  it("renders the active flat resume list from listResumes", async () => {
    withProviders(
      <ResumePickerModal open onClose={vi.fn()} boundResumeId={DEFAULT_RESUME_ID} />,
    );
    expect(screen.getByTestId("workspace-resume-modal-overlay")).toBeDefined();
    expect(screen.getByTestId("workspace-resume-modal-card")).toBeDefined();
    expect(
      await screen.findByTestId(`workspace-resume-modal-option-${DEFAULT_RESUME_ID}`),
    ).toBeDefined();
    expect(
      screen.getByTestId(`workspace-resume-modal-option-${SECOND_RESUME_ID}`),
    ).toBeDefined();
    expect(
      screen.queryByTestId("workspace-resume-modal-disabled-note"),
    ).not.toBeInTheDocument();
    expect(screen.getByTestId("workspace-resume-modal-close")).toBeDefined();
    expect(screen.getByTestId("workspace-resume-modal-cancel")).toBeDefined();
    expect(screen.getByTestId("workspace-resume-modal-confirm")).toBeDefined();
  });

  it("does NOT render when open=false", () => {
    const { container } = withProviders(
      <ResumePickerModal open={false} onClose={vi.fn()} />,
    );
    expect(container.querySelector("[data-testid]")).toBeNull();
  });

  it("closes on Cancel button click", async () => {
    const onClose = vi.fn();
    withProviders(<ResumePickerModal open onClose={onClose} />);
    const user = userEvent.setup();
    await user.click(screen.getByTestId("workspace-resume-modal-cancel"));
    expect(onClose).toHaveBeenCalled();
  });

  it("confirms the selected resume id", async () => {
    const onClose = vi.fn();
    const onSelectResume = vi.fn();
    withProviders(
      <ResumePickerModal
        open
        onClose={onClose}
        boundResumeId={DEFAULT_RESUME_ID}
        onSelectResume={onSelectResume}
      />,
    );
    const user = userEvent.setup();
    await user.click(
      await screen.findByTestId(`workspace-resume-modal-option-${SECOND_RESUME_ID}`),
    );
    await user.click(screen.getByTestId("workspace-resume-modal-confirm"));
    expect(onSelectResume).toHaveBeenCalledWith(SECOND_RESUME_ID);
    expect(onClose).toHaveBeenCalled();
  });

  it("does not confirm synthetic or stale bound resume ids", async () => {
    const onClose = vi.fn();
    const onSelectResume = vi.fn();
    withProviders(
      <ResumePickerModal
        open
        onClose={onClose}
        boundResumeId="resume-unbound"
        onSelectResume={onSelectResume}
      />,
    );

    const user = userEvent.setup();
    await screen.findByTestId(`workspace-resume-modal-option-${DEFAULT_RESUME_ID}`);
    expect(screen.getByTestId("workspace-resume-modal-confirm")).toBeDisabled();

    await user.click(screen.getByTestId("workspace-resume-modal-confirm"));
    expect(onSelectResume).not.toHaveBeenCalled();
    expect(onClose).not.toHaveBeenCalled();
  });

  it("has aria-modal attribute", () => {
    withProviders(<ResumePickerModal open onClose={vi.fn()} />);
    expect(screen.getByTestId("workspace-resume-modal-card")).toHaveAttribute(
      "aria-modal",
      "true",
    );
  });
});
