import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
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
} from "../theme/tokens";

/**
 * Global display preferences (Spec D-4 / docs/ui-design/auth-and-entry.md §4):
 * theme color / dark mode / UI language are held by the App shell (TopBar
 * surfaces the controls) and are intentionally independent of the auth state.
 * Signing in or out must NEVER reset these preferences.
 *
 * D2 (Phase 1.2) extends the provider with `customAccent` and roots the
 * theme / mode / custom-accent state on `<html>` via `data-theme` /
 * `data-mode` / `data-custom-accent`. The CSS palette in
 * `src/app/theme/themes.css` keys off those attributes; custom accent only
 * overrides `--ei-color-accent` and `--ei-color-accent-soft`, leaving the
 * rest of the EI_THEMES base palette untouched (matches ui-design/src/app.jsx
 * customAccent overlay model).
 *
 * `fontPreset` is owned by the settings page (Phase 4.2), not by TopBar; it is
 * therefore not exposed here.
 */

export type Theme = "warm" | "forest" | "ocean" | "plum";
export type Lang = "zh" | "en";

export interface CustomAccent {
  /** Hue in degrees, [0, 360). Out-of-range inputs are normalized. */
  h: number;
  /** Chroma in [0, 0.28]. Out-of-range inputs are clamped. */
  c: number;
}

export interface DisplayPreferences {
  theme: Theme;
  dark: boolean;
  lang: Lang;
  customAccent: CustomAccent | null;
  setTheme: (next: Theme) => void;
  setDark: (next: boolean) => void;
  setLang: (next: Lang) => void;
  setCustomAccent: (next: CustomAccent | null) => void;
}

const DEFAULTS = {
  theme: "warm" as Theme,
  dark: false,
  lang: "en" as Lang,
  customAccent: null as CustomAccent | null,
};

const Context = createContext<DisplayPreferences | null>(null);

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
    initial?.lang ?? getBrowserLang(),
  );
  const [customAccent, setCustomAccentState] = useState<CustomAccent | null>(
    initial?.customAccent ?? DEFAULTS.customAccent,
  );

  const setLang = useCallback((next: Lang) => {
    setLangState(next);
  }, []);

  const setCustomAccent = useCallback((next: CustomAccent | null) => {
    setCustomAccentState(next);
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
      setTheme,
      setDark,
      setLang,
      setCustomAccent,
    }),
    [theme, dark, lang, customAccent, setLang, setCustomAccent],
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
    const normalized = normalizeLang(candidate);
    if (normalized) return normalized;
  }
  return DEFAULTS.lang;
}

function normalizeLang(tag: string | undefined | null): Lang | null {
  const lower = tag?.trim().toLowerCase();
  if (!lower) return null;
  if (lower === "en" || lower.startsWith("en-")) return "en";
  if (lower === "zh" || lower.startsWith("zh-")) return "zh";
  return null;
}
