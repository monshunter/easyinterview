// @vitest-environment jsdom
import { afterEach, describe, expect, it } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { useState, type FC } from "react";

import {
  DisplayPreferencesProvider,
  normalizeAccountDisplayPreferences,
  useDisplayPreferences,
} from "./DisplayPreferencesProvider";

const Probe: FC<{ signedIn: boolean }> = ({ signedIn }) => {
  const prefs = useDisplayPreferences();
  return (
    <div>
      <span data-testid="theme">{prefs.theme}</span>
      <span data-testid="dark">{String(prefs.dark)}</span>
      <span data-testid="lang">{prefs.lang}</span>
      <span data-testid="signed-in">{String(signedIn)}</span>
      <button
        type="button"
        data-testid="set-theme-plum"
        onClick={() => prefs.setTheme("plum")}
      >
        theme plum
      </button>
      <button
        type="button"
        data-testid="set-dark-true"
        onClick={() => prefs.setDark(true)}
      >
        dark on
      </button>
      <button
        type="button"
        data-testid="set-lang-en"
        onClick={() => prefs.setLang("en")}
      >
        lang en
      </button>
    </div>
  );
};

const Harness: FC = () => {
  const [signedIn, setSignedIn] = useState(false);
  return (
    <DisplayPreferencesProvider>
      <button
        type="button"
        data-testid="toggle-auth"
        onClick={() => setSignedIn((v) => !v)}
      >
        toggle auth
      </button>
      <Probe signedIn={signedIn} />
    </DisplayPreferencesProvider>
  );
};

function setNavigatorLanguages(language: string, languages = [language]) {
  Object.defineProperty(window.navigator, "language", {
    value: language,
    configurable: true,
  });
  Object.defineProperty(window.navigator, "languages", {
    value: languages,
    configurable: true,
  });
}

describe("DisplayPreferencesProvider", () => {
  afterEach(() => {
    setNavigatorLanguages("en-US", ["en-US"]);
    window.localStorage.clear();
  });

  it("starts with the documented defaults and follows browser locale", () => {
    setNavigatorLanguages("zh-CN", ["zh-CN", "en-US"]);
    render(<Harness />);
    // product-scope D-21 (v2.1): the default theme is ocean.
    expect(screen.getByTestId("theme")).toHaveTextContent("ocean");
    expect(screen.getByTestId("dark")).toHaveTextContent("false");
    expect(screen.getByTestId("lang")).toHaveTextContent("zh");
  });

  it("fails closed for malformed account theme projections", () => {
    expect(normalizeAccountDisplayPreferences({ theme: "plum", customAccent: { h: 120, c: 0.18 } })).toEqual({
      theme: "plum",
      customAccent: { h: 120, c: 0.18 },
    });

    for (const projection of [
      null,
      { theme: "forest", customAccent: null },
      { theme: "plum", customAccent: { h: -1, c: 0.18 } },
      { theme: "plum", customAccent: { h: 360, c: 0.18 } },
      { theme: "plum", customAccent: { h: 120, c: -0.001 } },
      { theme: "plum", customAccent: { h: 120, c: 0.281 } },
      { theme: "plum", customAccent: { h: 120, c: 0.18, extra: true } },
      { theme: "plum", customAccent: null, extra: true },
    ]) {
      expect(normalizeAccountDisplayPreferences(projection)).toEqual({
        theme: "ocean",
        customAccent: null,
      });
    }
  });

  it("falls back to English for unsupported browser locales", () => {
    setNavigatorLanguages("fr-FR", ["fr-FR"]);
    render(<Harness />);
    expect(screen.getByTestId("lang")).toHaveTextContent("en");
  });

  it("uses a stored user language before browser locale", () => {
    window.localStorage.setItem("ei-lang", "en");
    setNavigatorLanguages("zh-CN", ["zh-CN", "en-US"]);
    render(<Harness />);
    expect(screen.getByTestId("lang")).toHaveTextContent("en");
  });

  it("ignores unsupported stored languages and persists explicit choices", async () => {
    window.localStorage.setItem("ei-lang", "de-DE");
    setNavigatorLanguages("zh-CN", ["zh-CN", "en-US"]);
    render(<Harness />);
    const user = userEvent.setup();

    expect(screen.getByTestId("lang")).toHaveTextContent("zh");

    await user.click(screen.getByTestId("set-lang-en"));
    expect(screen.getByTestId("lang")).toHaveTextContent("en");
    expect(window.localStorage.getItem("ei-lang")).toBe("en");
  });

  it("preserves theme / dark / lang across signed-in state transitions", async () => {
    render(<Harness />);
    const user = userEvent.setup();

    await user.click(screen.getByTestId("set-theme-plum"));
    await user.click(screen.getByTestId("set-dark-true"));
    await user.click(screen.getByTestId("set-lang-en"));

    expect(screen.getByTestId("theme")).toHaveTextContent("plum");
    expect(screen.getByTestId("dark")).toHaveTextContent("true");
    expect(screen.getByTestId("lang")).toHaveTextContent("en");

    await user.click(screen.getByTestId("toggle-auth"));
    expect(screen.getByTestId("signed-in")).toHaveTextContent("true");
    expect(screen.getByTestId("theme")).toHaveTextContent("plum");
    expect(screen.getByTestId("dark")).toHaveTextContent("true");
    expect(screen.getByTestId("lang")).toHaveTextContent("en");

    await user.click(screen.getByTestId("toggle-auth"));
    expect(screen.getByTestId("signed-in")).toHaveTextContent("false");
    expect(screen.getByTestId("theme")).toHaveTextContent("plum");
    expect(screen.getByTestId("dark")).toHaveTextContent("true");
    expect(screen.getByTestId("lang")).toHaveTextContent("en");
  });

  it("throws when used outside a provider", () => {
    const Orphan: FC = () => {
      useDisplayPreferences();
      return null;
    };
    // Suppress React's expected error log so the test output stays focused.
    const original = console.error;
    console.error = () => {};
    try {
      expect(() => render(<Orphan />)).toThrow(
        /useDisplayPreferences must be used inside/i,
      );
    } finally {
      console.error = original;
    }
  });
});
