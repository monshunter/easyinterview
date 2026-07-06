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
  it("separates JD paste and upload sources instead of keeping upload inside the textarea card", () => {
    render(wrap(<HomeScreen route={{ name: "home", params: {} }} />));

    const sourceLayout = screen.getByTestId("home-source-layout");
    const pastePanel = screen.getByTestId("home-jd-paste-panel");
    const uploadPanel = screen.getByTestId("home-upload-source-panel");
    const inputCard = screen.getByTestId("home-jd-input-card");
    const textarea = screen.getByTestId("home-jd-textarea");
    const uploadTrigger = screen.getByTestId("home-upload-trigger");

    expect(sourceLayout).toContainElement(pastePanel);
    expect(sourceLayout).toContainElement(uploadPanel);
    expect(pastePanel).toContainElement(inputCard);
    expect(inputCard).toContainElement(textarea);
    expect(uploadPanel).toContainElement(uploadTrigger);
    expect(inputCard).not.toContainElement(uploadTrigger);
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
