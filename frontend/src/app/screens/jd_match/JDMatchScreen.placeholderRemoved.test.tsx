// @vitest-environment jsdom
import { describe, expect, it, vi } from "vitest";
import { render, screen } from "@testing-library/react";

import { DisplayPreferencesProvider } from "../../display/DisplayPreferencesProvider";
import { MESSAGES } from "../../i18n/messages";
import { NavigationProvider } from "../../navigation/NavigationProvider";
import { JDMatchScreen } from "./JDMatchScreen";

const PLACEHOLDER_TESTIDS = [
  "jdmatch-placeholder",
  "jdmatch-placeholder-cta",
] as const;

const PLACEHOLDER_KEYS = [
  "jdMatch.placeholderTitle",
  "jdMatch.placeholderCopy",
  "jdMatch.placeholderCta",
] as const;

function wrap(ui: React.ReactElement, navigate = vi.fn()) {
  return (
    <NavigationProvider value={{ navigate }}>
      <DisplayPreferencesProvider>{ui}</DisplayPreferencesProvider>
    </NavigationProvider>
  );
}

describe("JDMatchScreen plan-001 placeholder removal (item 2.1)", () => {
  it("does not render placeholder testids in DOM", () => {
    render(
      wrap(<JDMatchScreen route={{ name: "jd_match", params: {} }} />),
    );

    for (const id of PLACEHOLDER_TESTIDS) {
      expect(screen.queryByTestId(id)).toBeNull();
    }
  });

  it("removes placeholder i18n keys from zh dictionary", () => {
    for (const key of PLACEHOLDER_KEYS) {
      expect(Object.prototype.hasOwnProperty.call(MESSAGES.zh, key)).toBe(
        false,
      );
    }
  });

  it("removes placeholder i18n keys from en dictionary", () => {
    for (const key of PLACEHOLDER_KEYS) {
      expect(Object.prototype.hasOwnProperty.call(MESSAGES.en, key)).toBe(
        false,
      );
    }
  });
});
