import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useRef,
  useState,
  type FC,
  type ReactNode,
} from "react";

import { computeCustomAccentOverrides } from "../theme/customAccent";
import {
  COLOR_TOKENS,
  ROOT_CUSTOM_ACCENT_ATTR,
  ROOT_MODE_ATTR,
  ROOT_THEME_ATTR,
  type Theme,
} from "../theme/tokens";
import {
  DEFAULT_LANG,
  resolveSupportedLocale,
  type Lang,
} from "../i18n/localeCatalog";

export type { Lang } from "../i18n/localeCatalog";
export type { Theme } from "../theme/tokens";

/**
 * Global display preferences (product-scope D-21 / UI architecture): dark
 * mode and UI language remain App-shell controls, while theme color is an
 * account preference edited in Settings Appearance. The provider owns both
 * the server-confirmed account value and the local preview draft.
 *
 * D2 (Phase 1.2) extends the provider with `customAccent` and roots the
 * theme / mode / custom-accent state on `<html>` via `data-theme` /
 * `data-mode` / `data-custom-accent`. The CSS palette in
 * `src/app/theme/themes.css` keys off those attributes; custom accent only
 * overrides `--ei-color-accent` and `--ei-color-accent-soft`, leaving the
 * rest of the EI_THEMES base palette untouched (matches formal frontend implementation
 * customAccent overlay model).
 *
 * `fontPreset` is owned by the settings page (Phase 4.2), not by TopBar; it is
 * therefore not exposed here.
 */

export interface CustomAccent {
  /** Hue in degrees, [0, 360). Out-of-range inputs are normalized. */
  h: number;
  /** Chroma in [0, 0.28]. Out-of-range inputs are clamped. */
  c: number;
}

export interface AccountDisplayPreferences {
  theme: Theme;
  customAccent: CustomAccent | null;
}

const DEFAULT_ACCOUNT_DISPLAY_PREFERENCES: AccountDisplayPreferences = {
  theme: "ocean",
  customAccent: null,
};

export function normalizeAccountDisplayPreferences(
  value: unknown,
): AccountDisplayPreferences {
  if (!hasExactKeys(value, ["theme", "customAccent"])) {
    return DEFAULT_ACCOUNT_DISPLAY_PREFERENCES;
  }
  if (value.theme !== "ocean" && value.theme !== "plum" && value.theme !== "forest") {
    return DEFAULT_ACCOUNT_DISPLAY_PREFERENCES;
  }
  if (value.customAccent === null) {
    return { theme: value.theme, customAccent: null };
  }
  if (!hasExactKeys(value.customAccent, ["h", "c"])) {
    return DEFAULT_ACCOUNT_DISPLAY_PREFERENCES;
  }
  const { h, c } = value.customAccent;
  if (
    typeof h !== "number" ||
    !Number.isFinite(h) ||
    h < 0 ||
    h >= 360 ||
    typeof c !== "number" ||
    !Number.isFinite(c) ||
    c < 0 ||
    c > 0.28
  ) {
    return DEFAULT_ACCOUNT_DISPLAY_PREFERENCES;
  }
  return { theme: value.theme, customAccent: { h, c } };
}

export interface DisplayPreferences {
  theme: Theme;
  dark: boolean;
  lang: Lang;
  customAccent: CustomAccent | null;
  confirmedTheme: Theme;
  confirmedCustomAccent: CustomAccent | null;
  setTheme: (next: Theme) => void;
  setDark: (next: boolean) => void;
  setLang: (next: Lang) => void;
  setCustomAccent: (next: CustomAccent | null) => void;
  commitAccountPreferences: (next: { theme: Theme; customAccent: CustomAccent | null }) => void;
  restoreConfirmedAccountPreferences: () => void;
}

const DEFAULTS = {
  // product-scope D-21: ocean is the default theme; invalid or
  // missing theme values must also fall back to ocean.
  theme: "ocean" as Theme,
  dark: false,
  lang: DEFAULT_LANG,
  customAccent: null as CustomAccent | null,
};

const Context = createContext<DisplayPreferences | null>(null);
const LANG_STORAGE_KEY = "ei-lang";

export interface DisplayPreferencesProviderProps {
  children: ReactNode;
  initial?: {
    theme?: Theme;
    dark?: boolean;
    lang?: Lang;
    customAccent?: CustomAccent | null;
  };
}

export const DisplayPreferencesProvider: FC<
  DisplayPreferencesProviderProps
> = ({ children, initial }) => {
  const [theme, setTheme] = useState<Theme>(initial?.theme ?? DEFAULTS.theme);
  const [dark, setDark] = useState<boolean>(initial?.dark ?? DEFAULTS.dark);
  const [lang, setLangState] = useState<Lang>(
    initial?.lang ?? getStoredLang() ?? getBrowserLang(),
  );
  const [customAccent, setCustomAccentState] = useState<CustomAccent | null>(
    initial?.customAccent ?? DEFAULTS.customAccent,
  );
  const confirmedRef = useRef({
    theme: initial?.theme ?? DEFAULTS.theme,
    customAccent: initial?.customAccent ?? DEFAULTS.customAccent,
  });
  const [confirmed, setConfirmed] = useState(confirmedRef.current);

  const setLang = useCallback((next: Lang) => {
    setLangState(next);
    writeStoredLang(next);
  }, []);

  const setCustomAccent = useCallback((next: CustomAccent | null) => {
    setCustomAccentState(next);
  }, []);

  const commitAccountPreferences = useCallback((next: { theme: Theme; customAccent: CustomAccent | null }) => {
    confirmedRef.current = next;
    setConfirmed(next);
    setTheme(next.theme);
    setCustomAccentState(next.customAccent);
  }, []);

  const restoreConfirmedAccountPreferences = useCallback(() => {
    setTheme(confirmedRef.current.theme);
    setCustomAccentState(confirmedRef.current.customAccent);
  }, []);

  // Root-element wiring: write data-theme / data-mode / data-custom-accent on
  // the <html> element so the static palette in `themes.css` activates the
  // correct combination, and apply customAccent overrides as inline overlay.
  useEffect(() => {
    const root = document.documentElement;
    root.setAttribute(ROOT_THEME_ATTR, theme);
    root.setAttribute(ROOT_MODE_ATTR, dark ? "dark" : "light");
    const overrides = computeCustomAccentOverrides(
      customAccent ? { ...customAccent, dark } : null,
    );
    if (overrides) {
      root.setAttribute(ROOT_CUSTOM_ACCENT_ATTR, "active");
      root.style.setProperty(
        COLOR_TOKENS.accent.base,
        overrides[COLOR_TOKENS.accent.base],
      );
      root.style.setProperty(
        COLOR_TOKENS.accent.soft,
        overrides[COLOR_TOKENS.accent.soft],
      );
    } else {
      root.removeAttribute(ROOT_CUSTOM_ACCENT_ATTR);
      root.style.removeProperty(COLOR_TOKENS.accent.base);
      root.style.removeProperty(COLOR_TOKENS.accent.soft);
    }
  }, [theme, dark, customAccent]);

  const value = useMemo<DisplayPreferences>(
    () => ({
      theme,
      dark,
      lang,
      customAccent,
      confirmedTheme: confirmed.theme,
      confirmedCustomAccent: confirmed.customAccent,
      setTheme,
      setDark,
      setLang,
      setCustomAccent,
      commitAccountPreferences,
      restoreConfirmedAccountPreferences,
    }),
    [theme, dark, lang, customAccent, confirmed, setLang, setCustomAccent, commitAccountPreferences, restoreConfirmedAccountPreferences],
  );

  return <Context.Provider value={value}>{children}</Context.Provider>;
};

export function useDisplayPreferences(): DisplayPreferences {
  const ctx = useContext(Context);
  if (!ctx) {
    throw new Error(
      "useDisplayPreferences must be used inside <DisplayPreferencesProvider>",
    );
  }
  return ctx;
}

export function useDisplayPreferencesOptional(): DisplayPreferences | null {
  return useContext(Context);
}

function getBrowserLang(): Lang {
  const candidates = [
    ...(globalThis.navigator?.languages ?? []),
    globalThis.navigator?.language,
  ];
  for (const candidate of candidates) {
    const normalized = resolveSupportedLocale(candidate);
    if (normalized) {
      return normalized;
    }
  }
  return DEFAULTS.lang;
}

function getStoredLang(): Lang | null {
  try {
    return resolveSupportedLocale(globalThis.localStorage?.getItem(LANG_STORAGE_KEY));
  } catch {
    return null;
  }
}

function writeStoredLang(next: Lang): void {
  try {
    globalThis.localStorage?.setItem(LANG_STORAGE_KEY, next);
  } catch {
    // Display preferences still work for browsers that block localStorage.
  }
}

function hasExactKeys<K extends string>(
  value: unknown,
  expected: readonly K[],
): value is Record<K, unknown> {
  if (typeof value !== "object" || value === null || Array.isArray(value)) {
    return false;
  }
  const keys = Object.keys(value);
  return keys.length === expected.length && expected.every((key) => keys.includes(key));
}
