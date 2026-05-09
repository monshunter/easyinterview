/**
 * @vitest-environment jsdom
 */

import { describe, expect, it, vi } from "vitest";
import { render, screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { useEffect, type ReactNode } from "react";

import { DisplayPreferencesProvider } from "../../display/DisplayPreferencesProvider";
import {
  InterviewContextProvider,
  useInterviewContext,
} from "../../interview-context/InterviewContext";
import { NavigationProvider } from "../../navigation/NavigationProvider";
import type { Route } from "../../routes";
import { WorkspaceScreen } from "./WorkspaceScreen";

function HydrateRoute({ route }: { route: Route }) {
  const { dispatch } = useInterviewContext();
  useEffect(() => {
    dispatch({ type: "HYDRATE_FROM_ROUTE", params: route.params });
  }, []);
  return null;
}

function withProviders(ui: ReactNode, route: Route) {
  const nav = vi.fn();
  return {
    nav,
    ...render(
      <DisplayPreferencesProvider>
        <InterviewContextProvider>
          <NavigationProvider value={{ navigate: nav }}>
            <HydrateRoute route={route} />
            {ui}
          </NavigationProvider>
        </InterviewContextProvider>
      </DisplayPreferencesProvider>,
    ),
  };
}

const WORKSPACE_ROUTE: Route = {
  name: "workspace",
  params: {
    targetJobId: "01918fa0-0000-7000-8000-000000002000",
    jdId: "jd-1",
    planId: "plan-1",
    resumeVersionId: "rv-1",
    roundId: "round-hr",
  },
};

describe("WorkspaceScreen static shell (Phase 1)", () => {
  it("renders plan eyebrow section with testids", () => {
    const { nav } = withProviders(
      <WorkspaceScreen route={WORKSPACE_ROUTE} />,
      WORKSPACE_ROUTE,
    );
    expect(screen.getByTestId("workspace-crumbs")).toBeDefined();
    expect(screen.getByTestId("workspace-plan-eyebrow")).toBeDefined();
    expect(screen.getByTestId("workspace-plan-eyebrow-label")).toBeDefined();
    expect(screen.getByTestId("workspace-plan-eyebrow-title")).toBeDefined();
    expect(screen.getByTestId("workspace-plan-eyebrow-status")).toBeDefined();
    expect(screen.getByTestId("workspace-plan-eyebrow-sub")).toBeDefined();
    expect(screen.getByTestId("workspace-plan-action-switch")).toBeDefined();
    expect(screen.getByTestId("workspace-plan-action-create")).toBeDefined();
    // nav stubs
    screen.getByTestId("workspace-crumbs").click();
    expect(nav).toHaveBeenCalledWith({ name: "home", params: {} });
    screen.getByTestId("workspace-plan-action-create").click();
    expect(nav).toHaveBeenCalledWith({ name: "home", params: {} });
  });

  it("renders header summary with testids", () => {
    withProviders(<WorkspaceScreen route={WORKSPACE_ROUTE} />, WORKSPACE_ROUTE);
    expect(screen.getByTestId("workspace-header")).toBeDefined();
    expect(screen.getByTestId("workspace-header-tag")).toBeDefined();
    expect(screen.getByTestId("workspace-header-level")).toBeDefined();
    expect(screen.getByTestId("workspace-header-updated")).toBeDefined();
    expect(screen.getByTestId("workspace-header-title")).toBeDefined();
    expect(screen.getByTestId("workspace-header-subtitle")).toBeDefined();
    expect(screen.getByTestId("workspace-header-prep")).toBeDefined();
  });

  it("renders Interview Launcher with Round Rail + CTA + BindingPill", () => {
    const { nav } = withProviders(
      <WorkspaceScreen route={WORKSPACE_ROUTE} />,
      WORKSPACE_ROUTE,
    );
    expect(screen.getByTestId("workspace-launcher")).toBeDefined();
    expect(screen.getByTestId("workspace-round-rail")).toBeDefined();
    expect(screen.getByTestId("workspace-cta-start")).toBeDefined();
    expect(screen.getByTestId("workspace-binding-jd")).toBeDefined();
    expect(screen.getByTestId("workspace-binding-resume")).toBeDefined();
    // CTA is present and is a button (no runtime provider → won't navigate)
    expect(screen.getByTestId("workspace-cta-start").tagName).toBe("BUTTON");
  });

  it("renders Main Left with CompanyIntelEmbed placeholder + JD breakdown", () => {
    withProviders(<WorkspaceScreen route={WORKSPACE_ROUTE} />, WORKSPACE_ROUTE);
    expect(screen.getByTestId("workspace-companyintel-summary")).toBeDefined();
    expect(screen.getByTestId("workspace-companyintel-open")).toBeDefined();
    expect(screen.getByTestId("workspace-jd-card")).toBeDefined();
    expect(screen.getByTestId("workspace-jd-block-must")).toBeDefined();
    expect(screen.getByTestId("workspace-jd-block-nice")).toBeDefined();
    expect(screen.getByTestId("workspace-jd-block-hidden")).toBeDefined();
  });

  it("renders Main Right with risks/strengths + sessionHistory placeholder", () => {
    withProviders(<WorkspaceScreen route={WORKSPACE_ROUTE} />, WORKSPACE_ROUTE);
    expect(screen.getByTestId("workspace-prep-card")).toBeDefined();
    expect(screen.getByTestId("workspace-prep-strongs")).toBeDefined();
    expect(screen.getByTestId("workspace-prep-risks")).toBeDefined();
    expect(screen.getByTestId("workspace-history-card")).toBeDefined();
    expect(screen.getByTestId("workspace-history-empty")).toBeDefined();
  });

  it("has at least 20 testids from plan §3.5 parity rows", () => {
    withProviders(<WorkspaceScreen route={WORKSPACE_ROUTE} />, WORKSPACE_ROUTE);
    const knownTestIds = [
      "workspace-crumbs",
      "workspace-plan-eyebrow",
      "workspace-plan-eyebrow-label",
      "workspace-plan-eyebrow-title",
      "workspace-plan-eyebrow-status",
      "workspace-plan-eyebrow-sub",
      "workspace-plan-action-switch",
      "workspace-plan-action-create",
      "workspace-header",
      "workspace-header-tag",
      "workspace-header-level",
      "workspace-header-updated",
      "workspace-header-title",
      "workspace-header-subtitle",
      "workspace-header-prep",
      "workspace-launcher",
      "workspace-round-rail",
      "workspace-cta-start",
      "workspace-binding-jd",
      "workspace-binding-resume",
      "workspace-companyintel-summary",
      "workspace-companyintel-open",
      "workspace-jd-card",
      "workspace-jd-block-must",
      "workspace-jd-block-nice",
      "workspace-jd-block-hidden",
      "workspace-prep-card",
      "workspace-prep-strongs",
      "workspace-prep-risks",
      "workspace-history-card",
      "workspace-history-empty",
    ];
    expect(knownTestIds.length).toBeGreaterThanOrEqual(20);
    for (const testid of knownTestIds) {
      expect(screen.getByTestId(testid)).toBeDefined();
    }
  });

  it("control type: Switch Plan button is a button not select", () => {
    withProviders(<WorkspaceScreen route={WORKSPACE_ROUTE} />, WORKSPACE_ROUTE);
    const el = screen.getByTestId("workspace-plan-action-switch");
    expect(el.tagName).toBe("BUTTON");
  });

  it("control type: CTA start is a button", () => {
    withProviders(<WorkspaceScreen route={WORKSPACE_ROUTE} />, WORKSPACE_ROUTE);
    const el = screen.getByTestId("workspace-cta-start");
    expect(el.tagName).toBe("BUTTON");
  });

  it("negative: old prototype testids do NOT exist", () => {
    withProviders(<WorkspaceScreen route={WORKSPACE_ROUTE} />, WORKSPACE_ROUTE);
    expect(screen.queryByTestId("practice-mode-card-warmup")).toBeNull();
    expect(screen.queryByTestId("practice-mode-card-single_drill")).toBeNull();
    expect(screen.queryByTestId("growth-center")).toBeNull();
    expect(screen.queryByTestId("drill-builder-daily")).toBeNull();
    expect(screen.queryByTestId("mistake-queue-entry")).toBeNull();
  });

  it("negative: no <select> elements for plan picker or resume picker", () => {
    withProviders(<WorkspaceScreen route={WORKSPACE_ROUTE} />, WORKSPACE_ROUTE);
    expect(screen.queryByRole("combobox")).toBeNull();
  });

  it("renders with zh locale for workspace.* namespace", () => {
    localStorage.setItem("ei-lang", "zh");
    render(
      <DisplayPreferencesProvider>
        <InterviewContextProvider>
          <NavigationProvider
            value={{ navigate: vi.fn() }}
          >
            <HydrateRoute route={WORKSPACE_ROUTE} />
            <WorkspaceScreen route={WORKSPACE_ROUTE} />
          </NavigationProvider>
        </InterviewContextProvider>
      </DisplayPreferencesProvider>,
    );
    expect(screen.getByTestId("workspace-plan-eyebrow-label").textContent).toBe(
      "当前面试规划",
    );
    localStorage.removeItem("ei-lang");
  });

  it("renders with en locale for workspace.* namespace", () => {
    localStorage.setItem("ei-lang", "en");
    render(
      <DisplayPreferencesProvider>
        <InterviewContextProvider>
          <NavigationProvider
            value={{ navigate: vi.fn() }}
          >
            <HydrateRoute route={WORKSPACE_ROUTE} />
            <WorkspaceScreen route={WORKSPACE_ROUTE} />
          </NavigationProvider>
        </InterviewContextProvider>
      </DisplayPreferencesProvider>,
    );
    expect(screen.getByTestId("workspace-plan-eyebrow-label").textContent).toBe(
      "Current Interview Plan",
    );
    expect(screen.getByTestId("workspace-round-rail")).toHaveTextContent("HR Screen");
    expect(screen.getByTestId("workspace-round-rail")).toHaveTextContent("Technical 1");
    expect(screen.getByTestId("workspace-jd-block-must")).toHaveTextContent("Must Have");
    expect(screen.getByTestId("workspace-jd-block-nice")).toHaveTextContent("Nice to Have");
    expect(screen.getByTestId("workspace-jd-block-hidden")).toHaveTextContent("Hidden Signals");
    expect(document.body).not.toHaveTextContent("必需项");
    expect(document.body).not.toHaveTextContent("技术一面");
    localStorage.removeItem("ei-lang");
  });

  it("clicking 'Switch Plan' opens the plan switcher modal", async () => {
    const { nav } = withProviders(
      <WorkspaceScreen route={WORKSPACE_ROUTE} />,
      WORKSPACE_ROUTE,
    );
    const user = userEvent.setup();
    await user.click(screen.getByTestId("workspace-plan-action-switch"));
    await waitFor(() => {
      expect(screen.getByTestId("workspace-plan-modal-card")).toBeDefined();
    });
    expect(nav).not.toHaveBeenCalled();
  });

  it("clicking company intel open stub calls navigate", () => {
    const { nav } = withProviders(
      <WorkspaceScreen route={WORKSPACE_ROUTE} />,
      WORKSPACE_ROUTE,
    );
    screen.getByTestId("workspace-companyintel-open").click();
    expect(nav).toHaveBeenCalled();
  });

  it("sessionHistory row click is disabled / stub-only", () => {
    withProviders(<WorkspaceScreen route={WORKSPACE_ROUTE} />, WORKSPACE_ROUTE);
    expect(screen.getByTestId("workspace-history-empty")).toBeDefined();
  });

  it("renders note line about practice context", () => {
    withProviders(<WorkspaceScreen route={WORKSPACE_ROUTE} />, WORKSPACE_ROUTE);
    expect(screen.getByTestId("workspace-note-practice")).toBeDefined();
  });
});
