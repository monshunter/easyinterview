/**
 * @vitest-environment jsdom
 */

import { describe, expect, it, vi } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import type { ReactNode } from "react";

import { EasyInterviewClient } from "../../../../api/generated/client";
import {
  createFixtureBackedFetch,
  createFixtureRegistry,
} from "../../../../api/mockTransport";
import { AppRuntimeProvider } from "../../../runtime/AppRuntimeProvider";
import { InterviewContextProvider } from "../../../interview-context/InterviewContext";
import { NavigationProvider } from "../../../navigation/NavigationProvider";
import { PlanSwitcherModal } from "./PlanSwitcherModal";

import listTargetJobsFixture from "../../../../../../openapi/fixtures/TargetJobs/listTargetJobs.json";

function buildClient() {
  return new EasyInterviewClient({
    fetch: createFixtureBackedFetch(
      createFixtureRegistry([listTargetJobsFixture]),
      { scenario: "default" },
    ),
  });
}

function withProviders(ui: ReactNode) {
  const client = buildClient();
  const nav = vi.fn();
  return {
    client,
    nav,
    ...render(
      <InterviewContextProvider>
        <AppRuntimeProvider client={client}>
          <NavigationProvider value={{ navigate: nav }}>
            {ui}
          </NavigationProvider>
        </AppRuntimeProvider>
      </InterviewContextProvider>,
    ),
  };
}

describe("PlanSwitcherModal (Phase 3.4)", () => {
  it("renders when open with correct testids", async () => {
    const onSelect = vi.fn();
    withProviders(
      <PlanSwitcherModal open onClose={vi.fn()} onSelectPlan={onSelect} />,
    );

    await waitFor(() => {
      expect(screen.getByTestId("workspace-plan-modal-card")).toBeDefined();
    });

    expect(screen.getByTestId("workspace-plan-modal-overlay")).toBeDefined();
    expect(screen.getByTestId("workspace-plan-modal-close")).toBeDefined();
    expect(screen.getByTestId("workspace-plan-modal-create")).toBeDefined();
    expect(screen.getByTestId("workspace-plan-modal-cancel")).toBeDefined();
    expect(screen.getByTestId("workspace-plan-modal-confirm")).toBeDefined();
  });

  it("renders plan cards from listTargetJobs data", async () => {
    const { client } = withProviders(
      <PlanSwitcherModal open onClose={vi.fn()} onSelectPlan={vi.fn()} />,
    );

    // Verify listTargetJobs was called
    const spy = vi.spyOn(client, "listTargetJobs");

    // Wait for the modal to fully render
    await waitFor(() => {
      expect(spy).toHaveBeenCalled();
    });
  });

  it("Create Plan CTA navigates to home", async () => {
    const { nav } = withProviders(
      <PlanSwitcherModal open onClose={vi.fn()} onSelectPlan={vi.fn()} />,
    );
    const user = userEvent.setup();

    await waitFor(() => {
      expect(screen.getByTestId("workspace-plan-modal-create")).toBeDefined();
    });

    await user.click(screen.getByTestId("workspace-plan-modal-create"));
    expect(nav).toHaveBeenCalledWith({ name: "home", params: {} });
  });

  it("calls onSelectPlan when clicking a plan card", async () => {
    const onSelect = vi.fn();
    const { container } = withProviders(
      <PlanSwitcherModal open onClose={vi.fn()} onSelectPlan={onSelect} />,
    );
    const user = userEvent.setup();

    await waitFor(() => {
      expect(screen.getByTestId("workspace-plan-modal-create")).toBeDefined();
    });

    const firstCard = container.querySelector(
      "[data-testid^='workspace-plan-modal-card-']",
    ) as HTMLElement;
    expect(firstCard).toBeDefined();
    await user.click(firstCard);
    expect(onSelect).toHaveBeenCalled();
  });

  it("has aria-modal attribute", async () => {
    withProviders(
      <PlanSwitcherModal open onClose={vi.fn()} onSelectPlan={vi.fn()} />,
    );

    await waitFor(() => {
      expect(screen.getByTestId("workspace-plan-modal-card")).toBeDefined();
    });

    expect(screen.getByTestId("workspace-plan-modal-card")).toHaveAttribute(
      "aria-modal",
      "true",
    );
  });
});
