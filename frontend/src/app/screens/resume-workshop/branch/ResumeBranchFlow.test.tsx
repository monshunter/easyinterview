// @vitest-environment jsdom
import { describe, expect, it, vi } from "vitest";
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
import { ResumeBranchFlow } from "./ResumeBranchFlow";

import getRuntimeConfigFixture from "../../../../../../openapi/fixtures/Auth/getRuntimeConfig.json";
import getMeFixture from "../../../../../../openapi/fixtures/Auth/getMe.json";
import listResumesFixture from "../../../../../../openapi/fixtures/Resumes/listResumes.json";
import listResumeVersionsFixture from "../../../../../../openapi/fixtures/Resumes/listResumeVersions.json";

const RESUME_FIXTURES = [
  getRuntimeConfigFixture,
  getMeFixture,
  listResumesFixture,
  listResumeVersionsFixture,
];

const buildClient = (): EasyInterviewClient =>
  new EasyInterviewClient({
    fetch: createFixtureBackedFetch(
      createFixtureRegistry(RESUME_FIXTURES),
      { scenario: "default" },
    ),
  });

const renderBranchFlow = (
  branchOriginalId: string | null,
  options?: {
    authMode?: "authenticated" | "unauthenticated";
    nav?: ReturnType<typeof vi.fn>;
  },
) => {
  const client = buildClient();
  const nav = options?.nav ?? vi.fn();
  return {
    client,
    nav,
    ...render(
      <DisplayPreferencesProvider>
        <AppRuntimeProvider
          client={client}
          requestOptions={{
            getMe: {
              headers: {
                Prefer: `example=${options?.authMode ?? "authenticated"}`,
              },
            },
          }}
        >
          <NavigationProvider value={{ navigate: nav }}>
            <ResumeBranchFlow branchOriginalId={branchOriginalId} />
          </NavigationProvider>
        </AppRuntimeProvider>
      </DisplayPreferencesProvider>,
    ),
  };
};

const FIXTURE_MASTER_ASSET_ID = listResumesFixture.scenarios.default.response
  .body.items[0]!.id as string;

describe("ResumeBranchFlow source resolution", () => {
  it("shows the missing-id fallback panel when branchOriginalId is null", () => {
    renderBranchFlow(null);
    expect(screen.getByTestId("resume-branch-missing-id")).toBeInTheDocument();
    expect(screen.queryByTestId("resume-branch-flow-form")).not.toBeInTheDocument();
  });

  it("shows the not-found fallback when the originalId does not match any asset", async () => {
    renderBranchFlow("01918fa0-0000-7000-8000-000000009999");
    await waitFor(() => {
      expect(screen.getByTestId("resume-branch-not-found")).toBeInTheDocument();
    });
    expect(screen.queryByTestId("resume-branch-flow-form")).not.toBeInTheDocument();
  });

  it("renders the form once the original asset and master version are resolved", async () => {
    renderBranchFlow(FIXTURE_MASTER_ASSET_ID);
    await waitFor(() => {
      expect(screen.getByTestId("resume-branch-flow-form")).toBeInTheDocument();
    });
    const form = screen.getByTestId("resume-branch-flow-form");
    expect(form).toHaveAttribute(
      "data-branch-original-id",
      FIXTURE_MASTER_ASSET_ID,
    );
    expect(screen.getByTestId("resume-branch-from-card")).toBeInTheDocument();
    expect(
      screen.getByTestId("resume-branch-from-original-name"),
    ).toHaveTextContent(/.+/);
    expect(
      screen.getByTestId("resume-branch-from-master-name"),
    ).toHaveTextContent(/.+/);
  });

  it("clicking back navigates to resume_versions list", async () => {
    const nav = vi.fn();
    renderBranchFlow(FIXTURE_MASTER_ASSET_ID, { nav });
    const backBtn = await screen.findByTestId("resume-branch-back");
    await userEvent.setup().click(backBtn);
    expect(nav).toHaveBeenCalledWith({
      name: "resume_versions",
      params: {},
    });
  });
});

describe("ResumeBranchFlow form behaviour", () => {
  it("starts with platform focus + copy_master seed defaults and submit disabled when name/target are empty", async () => {
    renderBranchFlow(FIXTURE_MASTER_ASSET_ID);
    await waitFor(() => {
      expect(screen.getByTestId("resume-branch-flow-form")).toBeInTheDocument();
    });
    const form = screen.getByTestId("resume-branch-flow-form");
    expect(form).toHaveAttribute("data-branch-focus", "platform");
    expect(form).toHaveAttribute("data-branch-seed", "copy_master");
    expect(form).toHaveAttribute("data-branch-can-submit", "false");

    const submit = screen.getByTestId("resume-branch-submit");
    expect(submit).toBeDisabled();
    expect(submit).toHaveAttribute("aria-disabled", "true");

    expect(
      screen
        .getByTestId("resume-branch-focus-chip-platform")
        .getAttribute("data-selected"),
    ).toBe("true");
    expect(
      screen
        .getByTestId("resume-branch-seed-card-copy_master")
        .getAttribute("data-selected"),
    ).toBe("true");
  });

  it("enables submit once both name and target inputs have non-blank content", async () => {
    renderBranchFlow(FIXTURE_MASTER_ASSET_ID);
    const user = userEvent.setup();
    const nameInput = await screen.findByTestId("resume-branch-field-name");
    const targetInput = screen.getByTestId("resume-branch-field-target");

    await user.type(nameInput, "v3 ByteDance FE Platform");
    expect(screen.getByTestId("resume-branch-submit")).toBeDisabled();

    await user.type(targetInput, "ByteDance Frontend Platform");
    await waitFor(() => {
      expect(screen.getByTestId("resume-branch-submit")).toBeEnabled();
    });
    expect(
      screen.getByTestId("resume-branch-flow-form"),
    ).toHaveAttribute("data-branch-can-submit", "true");
  });

  it("treats whitespace-only inputs as invalid and keeps submit disabled", async () => {
    renderBranchFlow(FIXTURE_MASTER_ASSET_ID);
    const user = userEvent.setup();
    const nameInput = await screen.findByTestId("resume-branch-field-name");
    const targetInput = screen.getByTestId("resume-branch-field-target");
    await user.type(nameInput, "   ");
    await user.type(targetInput, "\t\t  ");
    expect(screen.getByTestId("resume-branch-submit")).toBeDisabled();
    expect(
      screen.getByTestId("resume-branch-flow-form"),
    ).toHaveAttribute("data-branch-can-submit", "false");
  });

  it("toggles focus chips across all five focus angles", async () => {
    renderBranchFlow(FIXTURE_MASTER_ASSET_ID);
    const user = userEvent.setup();
    await screen.findByTestId("resume-branch-flow-form");

    for (const focus of [
      "collaboration",
      "fullstack",
      "leadership",
      "custom",
      "platform",
    ] as const) {
      await user.click(screen.getByTestId(`resume-branch-focus-chip-${focus}`));
      expect(
        screen.getByTestId("resume-branch-flow-form"),
      ).toHaveAttribute("data-branch-focus", focus);
      expect(
        screen
          .getByTestId(`resume-branch-focus-chip-${focus}`)
          .getAttribute("data-selected"),
      ).toBe("true");
    }
  });

  it("toggles seed cards across copy_master, blank, and ai_select", async () => {
    renderBranchFlow(FIXTURE_MASTER_ASSET_ID);
    const user = userEvent.setup();
    await screen.findByTestId("resume-branch-flow-form");

    for (const seed of ["blank", "ai_select", "copy_master"] as const) {
      await user.click(screen.getByTestId(`resume-branch-seed-card-${seed}`));
      expect(
        screen.getByTestId("resume-branch-flow-form"),
      ).toHaveAttribute("data-branch-seed", seed);
      expect(
        screen
          .getByTestId(`resume-branch-seed-card-${seed}`)
          .getAttribute("data-selected"),
      ).toBe("true");
    }
  });

  it("invokes onSubmitDraft with trimmed name/target plus selected focus and seed", async () => {
    const onSubmitDraft = vi.fn().mockResolvedValue({
      kind: "version",
      version: {} as never,
    });
    const client = buildClient();
    const nav = vi.fn();
    render(
      <DisplayPreferencesProvider>
        <AppRuntimeProvider
          client={client}
          requestOptions={{
            getMe: { headers: { Prefer: "example=authenticated" } },
          }}
        >
          <NavigationProvider value={{ navigate: nav }}>
            <ResumeBranchFlow
              branchOriginalId={FIXTURE_MASTER_ASSET_ID}
              onSubmitDraft={onSubmitDraft}
            />
          </NavigationProvider>
        </AppRuntimeProvider>
      </DisplayPreferencesProvider>,
    );

    const user = userEvent.setup();
    const nameInput = await screen.findByTestId("resume-branch-field-name");
    const targetInput = screen.getByTestId("resume-branch-field-target");
    await user.type(nameInput, "  v3 ByteDance  ");
    await user.type(targetInput, " ByteDance Frontend Platform ");
    await user.click(screen.getByTestId("resume-branch-focus-chip-fullstack"));
    await user.click(screen.getByTestId("resume-branch-seed-card-ai_select"));

    await user.click(screen.getByTestId("resume-branch-submit"));

    await waitFor(() => {
      expect(onSubmitDraft).toHaveBeenCalledTimes(1);
    });
    expect(onSubmitDraft).toHaveBeenCalledWith({
      name: "v3 ByteDance",
      target: "ByteDance Frontend Platform",
      focus: "fullstack",
      seed: "ai_select",
    });
  });

  it("does not trigger network calls to protected resume APIs in the missing-id state", async () => {
    const client = buildClient();
    const listSpy = vi.spyOn(client, "listResumes");
    const versionsSpy = vi.spyOn(client, "listResumeVersions");
    const nav = vi.fn();
    render(
      <DisplayPreferencesProvider>
        <AppRuntimeProvider
          client={client}
          requestOptions={{
            getMe: { headers: { Prefer: "example=authenticated" } },
          }}
        >
          <NavigationProvider value={{ navigate: nav }}>
            <ResumeBranchFlow branchOriginalId={null} />
          </NavigationProvider>
        </AppRuntimeProvider>
      </DisplayPreferencesProvider>,
    );
    expect(screen.getByTestId("resume-branch-missing-id")).toBeInTheDocument();
    // listResumes still fires from useResumeAssets in authenticated state,
    // but listResumeVersions must stay quiet without an originalId.
    expect(versionsSpy).not.toHaveBeenCalled();
    // Ensure auth-restoring listResumes is the only protected call from this branch path.
    await waitFor(() => {
      expect(listSpy).toHaveBeenCalledTimes(1);
    });
  });
});
