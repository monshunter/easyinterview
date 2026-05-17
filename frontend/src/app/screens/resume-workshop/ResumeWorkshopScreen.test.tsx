// @vitest-environment jsdom
import { describe, expect, it } from "vitest";
import { render, screen } from "@testing-library/react";

import { App } from "../../App";

const renderResumeWorkshop = (params: Record<string, string>) =>
  render(
    <App initialRoute={{ name: "resume_versions", params }} />,
  );

describe("ResumeWorkshopScreen route param parsing", () => {
  it("defaults flow to list when no flow param is provided", () => {
    renderResumeWorkshop({});
    const root = screen.getByTestId("resume-workshop-screen");
    expect(root).toHaveAttribute("data-flow", "list");
    expect(screen.getByTestId("resume-workshop-list")).toBeInTheDocument();
    expect(screen.queryByTestId("resume-workshop-detail")).not.toBeInTheDocument();
    expect(screen.queryByTestId("resume-workshop-not-implemented")).not.toBeInTheDocument();
  });

  it("dispatches flow=create to ResumeCreateFlow without blocking the screen", () => {
    renderResumeWorkshop({ flow: "create" });
    const root = screen.getByTestId("resume-workshop-screen");
    expect(root).toHaveAttribute("data-flow", "create");
    expect(screen.getByTestId("resume-create-flow")).toBeInTheDocument();
    expect(
      screen.queryByTestId("resume-workshop-not-implemented"),
    ).not.toBeInTheDocument();
    expect(screen.queryByTestId("resume-workshop-list")).not.toBeInTheDocument();
  });

  it("forwards createMode route param to ResumeCreateFlow", () => {
    renderResumeWorkshop({ flow: "create", createMode: "paste" });
    const root = screen.getByTestId("resume-workshop-screen");
    expect(root).toHaveAttribute("data-create-mode", "paste");
    const flow = screen.getByTestId("resume-create-flow");
    expect(flow).toHaveAttribute("data-create-mode", "paste");
  });

  it("dispatches flow=branch to ResumeBranchFlow and preserves branchOriginalId", () => {
    renderResumeWorkshop({
      flow: "branch",
      branchOriginalId: "01918fa0-0000-7000-8000-000000001000",
    });
    const root = screen.getByTestId("resume-workshop-screen");
    expect(root).toHaveAttribute("data-flow", "branch");
    expect(root).toHaveAttribute(
      "data-branch-original-id",
      "01918fa0-0000-7000-8000-000000001000",
    );
    const branchRoot = screen.getByTestId("resume-branch-flow");
    expect(branchRoot).toBeInTheDocument();
    expect(branchRoot).toHaveAttribute(
      "data-branch-original-id",
      "01918fa0-0000-7000-8000-000000001000",
    );
    // Phase 1: NotImplementedPlaceholder no longer participates in flow=branch.
    expect(
      screen.queryByTestId("resume-workshop-not-implemented"),
    ).not.toBeInTheDocument();
  });

  it("renders the detail container when versionId is set without an explicit tab", () => {
    renderResumeWorkshop({ versionId: "0195f2d0-0001-7000-8000-000000000201" });
    expect(screen.getByTestId("resume-workshop-detail")).toBeInTheDocument();
    expect(screen.getByTestId("resume-workshop-detail")).toHaveAttribute(
      "data-resume-version-id",
      "0195f2d0-0001-7000-8000-000000000201",
    );
    expect(screen.queryByTestId("resume-workshop-list")).not.toBeInTheDocument();
  });

  it("preserves an explicit tab=preview on the detail container", () => {
    renderResumeWorkshop({
      versionId: "0195f2d0-0001-7000-8000-000000000201",
      tab: "preview",
    });
    const detail = screen.getByTestId("resume-workshop-detail");
    expect(detail).toHaveAttribute("data-tab", "preview");
  });

  it("preserves an explicit tab=rewrites on the detail container (TARGETED entries are not rewritten to preview)", () => {
    renderResumeWorkshop({
      versionId: "0195f2d0-0001-7000-8000-000000000202",
      tab: "rewrites",
    });
    const detail = screen.getByTestId("resume-workshop-detail");
    expect(detail).toHaveAttribute("data-tab", "rewrites");
    expect(detail).not.toHaveAttribute("data-tab", "preview");
  });
});

describe("ResumeWorkshopScreen flow=branch dispatch (plan 003)", () => {
  it("renders ResumeBranchFlow for flow=branch and never falls back to NotImplementedPlaceholder", () => {
    renderResumeWorkshop({
      flow: "branch",
      branchOriginalId: "01918fa0-0000-7000-8000-000000001000",
    });
    expect(screen.getByTestId("resume-branch-flow")).toBeInTheDocument();
    expect(
      screen.queryByTestId("resume-workshop-not-implemented"),
    ).not.toBeInTheDocument();
  });

  it("shows the missing-id fallback panel when flow=branch is opened without a branchOriginalId", () => {
    renderResumeWorkshop({ flow: "branch" });
    expect(screen.getByTestId("resume-branch-flow")).toBeInTheDocument();
    expect(screen.getByTestId("resume-branch-missing-id")).toBeInTheDocument();
  });

  it("does not render ResumeBranchFlow when flow=create is in effect (CreateFlow takes over)", () => {
    renderResumeWorkshop({ flow: "create" });
    expect(screen.queryByTestId("resume-branch-flow")).not.toBeInTheDocument();
    expect(screen.getByTestId("resume-create-flow")).toBeInTheDocument();
  });

  it("does not render ResumeBranchFlow when flow=list is in effect", () => {
    renderResumeWorkshop({});
    expect(screen.queryByTestId("resume-branch-flow")).not.toBeInTheDocument();
    expect(screen.getByTestId("resume-workshop-list")).toBeInTheDocument();
  });
});
