// @vitest-environment jsdom
import { describe, expect, it, vi } from "vitest";
import { render, screen } from "@testing-library/react";

import { MockInterviewCard } from "./MockInterviewCard";
import type { TargetJob } from "../../../api/generated/types";

const mockJob: TargetJob = {
  id: "job-001",
  title: "Senior Frontend Engineer",
  companyName: "Acme Corp",
  locationText: "San Francisco, CA",
  status: "preparing",
  analysisStatus: "ready",
  sourceType: "manual_text",
  targetLanguage: "en",
  createdAt: "2026-01-01T00:00:00Z",
  updatedAt: "2026-05-01T00:00:00Z",
  requirements: [],
  openQuestionIssueCount: 0,
};

const provenance = {
  modelId: "fixture-model:target-import-parse",
  promptVersion: "v0.1.0",
  rubricVersion: "v0.1.0",
  dataSourceVersion: "registry.v1",
  featureFlag: "none",
  language: "en",
};

describe("MockInterviewCard", () => {
  it("renders card with testid", () => {
    render(
      <MockInterviewCard job={mockJob} onClick={() => {}} />,
    );

    expect(
      screen.getByTestId("home-recent-mock-card-job-001"),
    ).toBeInTheDocument();
  });

  it("displays job title and company name", () => {
    render(
      <MockInterviewCard job={mockJob} onClick={() => {}} />,
    );

    expect(screen.getByText("Senior Frontend Engineer")).toBeInTheDocument();
    expect(screen.getByText(/ACME CORP/)).toBeInTheDocument();
  });

  it("displays location or fallback", () => {
    render(
      <MockInterviewCard job={mockJob} onClick={() => {}} />,
    );

    expect(screen.getByText("San Francisco, CA")).toBeInTheDocument();
  });

  it("shows fallback location when locationText is empty", () => {
    const noLocationJob = { ...mockJob, locationText: "" };
    render(
      <MockInterviewCard job={noLocationJob} onClick={() => {}} />,
    );

    expect(screen.getByText("Location not set")).toBeInTheDocument();
  });

  it("renders MiniRoundRail testid", () => {
    render(
      <MockInterviewCard job={mockJob} onClick={() => {}} />,
    );

    expect(
      screen.getByTestId("home-recent-mock-rail-job-001"),
    ).toBeInTheDocument();
  });

  it("renders MiniRoundRail labels from target-job structured interview rounds", () => {
    render(
      <MockInterviewCard
        job={{
          ...mockJob,
          summary: {
            coreThemes: [],
            interviewRounds: [
              {
                sequence: 1,
                type: "hr",
                name: "Recruiter screen",
                durationMinutes: 30,
                focus: "LLM HR screen probes motivation fit",
              },
              {
                sequence: 2,
                type: "technical",
                name: "Frontend architecture interview",
                durationMinutes: 55,
                focus: "LLM technical round probes frontend architecture",
              },
            ],
            provenance,
          },
        }}
        onClick={() => {}}
      />,
    );

    const rail = screen.getByTestId("home-recent-mock-rail-job-001");
    expect(rail).toHaveTextContent("Recruiter screen · 30m");
    expect(rail).toHaveTextContent("Frontend architecture interview · 55m");
    expect(rail).not.toHaveTextContent("R1 Phone Screen");
    expect(rail).not.toHaveTextContent("HR screen · 20m");
  });

  it("calls onClick when clicked", () => {
    let clicked = false;
    render(
      <MockInterviewCard
        job={mockJob}
        onClick={() => {
          clicked = true;
        }}
      />,
    );

    screen.getByTestId("home-recent-mock-card-job-001").click();
    expect(clicked).toBe(true);
  });

  it("supports workspace-owned testids and shared quick-start/delete actions", () => {
    const cardClick = vi.fn();
    const start = vi.fn();
    const remove = vi.fn();

    render(
      <MockInterviewCard
        job={mockJob}
        onClick={cardClick}
        cardTestId="workspace-plan-list-card-job-001"
        bodyTestId="workspace-plan-list-card-body-job-001"
        railTestId="workspace-plan-list-rail-job-001"
        footerTestId="workspace-plan-list-card-footer-job-001"
        primaryAction={{
          label: "Start interview now",
          testId: "workspace-plan-list-start-job-001",
          onClick: start,
        }}
        deleteAction={{
          label: "Delete",
          testId: "workspace-plan-list-delete-job-001",
          onClick: remove,
        }}
      />,
    );

    expect(
      screen.getByTestId("workspace-plan-list-card-job-001"),
    ).toBeInTheDocument();
    expect(
      screen.getByTestId("workspace-plan-list-card-body-job-001"),
    ).toBeInTheDocument();
    expect(
      screen.getByTestId("workspace-plan-list-rail-job-001"),
    ).toBeInTheDocument();
    expect(
      screen.getByTestId("workspace-plan-list-card-footer-job-001"),
    ).toHaveTextContent("Start interview now");
    expect(
      screen.getByTestId("workspace-plan-list-card-footer-job-001"),
    ).not.toHaveTextContent("Open plan");
    expect(
      screen
        .getByTestId("workspace-plan-list-card-footer-job-001")
        .querySelector("[data-testid='workspace-plan-list-delete-job-001']"),
    ).toBeNull();

    screen.getByTestId("workspace-plan-list-start-job-001").click();
    expect(start).toHaveBeenCalledTimes(1);
    expect(cardClick).not.toHaveBeenCalled();

    const deleteButton = screen.getByTestId("workspace-plan-list-delete-job-001");
    expect(deleteButton).toHaveAttribute("aria-label", "Delete");
    expect((deleteButton as HTMLElement).style.position).toBe("absolute");
    expect((deleteButton as HTMLElement).style.right).toBe("14px");
    expect((deleteButton as HTMLElement).style.top).toBe("14px");
    expect(deleteButton.querySelector('[data-icon="trash"]')).not.toBeNull();
    deleteButton.click();
    expect(remove).toHaveBeenCalledTimes(1);
    expect(cardClick).not.toHaveBeenCalled();
  });

  it("can render a quick-start action without a delete action for Home recent cards", () => {
    render(
      <MockInterviewCard
        job={mockJob}
        primaryAction={{
          label: "Start interview now",
          testId: "home-recent-mock-start-job-001",
          onClick: () => {},
        }}
      />,
    );

    expect(
      screen.getByTestId("home-recent-mock-start-job-001"),
    ).toHaveTextContent("Start interview now");
    expect(
      screen.queryByTestId("home-recent-mock-delete-job-001"),
    ).toBeNull();
  });
});
