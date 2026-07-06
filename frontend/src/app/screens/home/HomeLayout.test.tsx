// @vitest-environment jsdom
import { describe, expect, it, vi } from "vitest";
import { render, screen } from "@testing-library/react";

import { DisplayPreferencesProvider } from "../../display/DisplayPreferencesProvider";
import { NavigationProvider } from "../../navigation/NavigationProvider";
import { HomeScreen } from "./HomeScreen";

function wrap(ui: React.ReactElement) {
  return (
    <NavigationProvider value={{ navigate: vi.fn() }}>
      <DisplayPreferencesProvider>{ui}</DisplayPreferencesProvider>
    </NavigationProvider>
  );
}

describe("Home layout", () => {
  it("keeps JD upload and URL source actions inside the input card", () => {
    render(wrap(<HomeScreen route={{ name: "home", params: {} }} />));

    const inputCard = screen.getByTestId("home-jd-input-card");
    const textarea = screen.getByTestId("home-jd-textarea");
    const sourceControls = screen.getByTestId("home-jd-source-controls");
    const uploadTrigger = screen.getByTestId("home-upload-trigger");
    const urlTrigger = screen.getByTestId("home-url-trigger");

    expect(screen.queryByTestId("home-source-layout")).not.toBeInTheDocument();
    expect(screen.queryByTestId("home-upload-source-panel")).not.toBeInTheDocument();
    expect(inputCard).toContainElement(textarea);
    expect(inputCard).toContainElement(sourceControls);
    expect(sourceControls).toContainElement(uploadTrigger);
    expect(sourceControls).toContainElement(urlTrigger);
  });

  it("keeps resume selection compact with the create CTA on the same row", () => {
    render(wrap(<HomeScreen route={{ name: "home", params: {} }} />));

    const resumeRow = screen.getByTestId("home-resume-row");
    const resumeSelect = screen.getByTestId("home-resume-select");
    const createCta = screen.getByTestId("home-resume-create");

    expect(resumeRow).toContainElement(resumeSelect);
    expect(resumeRow).toContainElement(createCta);
    expect(resumeSelect).toHaveStyle({ width: "360px" });
    expect(resumeSelect).toHaveStyle({ maxWidth: "100%" });
    expect(resumeSelect).not.toHaveStyle({ width: "100%" });
  });

  it("places immediate interview below resume selection and outside the JD input card", () => {
    render(wrap(<HomeScreen route={{ name: "home", params: {} }} />));

    const inputCard = screen.getByTestId("home-jd-input-card");
    const resumeRow = screen.getByTestId("home-resume-row");
    const submitRow = screen.getByTestId("home-submit-row");
    const submitButton = screen.getByTestId("home-jd-submit");

    expect(submitRow).toContainElement(submitButton);
    expect(inputCard).not.toContainElement(submitButton);
    expect(
      resumeRow.compareDocumentPosition(submitRow) &
        Node.DOCUMENT_POSITION_FOLLOWING,
    ).toBeTruthy();
  });
});
