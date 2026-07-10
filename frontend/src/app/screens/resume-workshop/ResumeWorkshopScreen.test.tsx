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

  it("renders the detail container when resumeId is set without an explicit tab", () => {
    renderResumeWorkshop({ resumeId: "0195f2d0-0001-7000-8000-000000000201" });
    expect(screen.getByTestId("resume-workshop-detail")).toBeInTheDocument();
    expect(screen.getByTestId("resume-workshop-detail")).toHaveAttribute(
      "data-resume-id",
      "0195f2d0-0001-7000-8000-000000000201",
    );
    expect(screen.queryByTestId("resume-workshop-list")).not.toBeInTheDocument();
  });

  it("ignores an explicit tab=preview on the detail container", () => {
    renderResumeWorkshop({
      resumeId: "0195f2d0-0001-7000-8000-000000000201",
      tab: "preview",
    });
    const detail = screen.getByTestId("resume-workshop-detail");
    expect(detail).not.toHaveAttribute("data-tab");
  });

  it("ignores an explicit tab=rewrites on the detail container", () => {
    renderResumeWorkshop({
      resumeId: "0195f2d0-0001-7000-8000-000000000202",
      tab: "rewrites",
    });
    const detail = screen.getByTestId("resume-workshop-detail");
    expect(detail).not.toHaveAttribute("data-tab");
  });
});

describe("ResumeWorkshopScreen flow dispatch (D-20 flat Resume)", () => {
  it("renders the flat list for flow=list", () => {
    renderResumeWorkshop({});
    expect(screen.getByTestId("resume-workshop-list")).toBeInTheDocument();
  });

  it("renders the create flow for flow=create", () => {
    renderResumeWorkshop({ flow: "create" });
    expect(screen.getByTestId("resume-create-flow")).toBeInTheDocument();
    expect(screen.queryByTestId("resume-workshop-list")).not.toBeInTheDocument();
  });

  it("does not expose an out-of-scope branch flow surface for any flow value", () => {
    renderResumeWorkshop({ flow: "branch" });
    // D-20 keeps branch flow outside current scope. An unknown flow falls back to
    // the flat list; the branch surface must never render.
    expect(screen.queryByTestId("resume-branch-flow")).not.toBeInTheDocument();
    expect(screen.getByTestId("resume-workshop-list")).toBeInTheDocument();
  });
});
