import {
  createContext,
  useContext,
  useMemo,
  useState,
  type FC,
  type ReactNode,
} from "react";

/**
 * Global display preferences (Spec D-4 / docs/ui-design/auth-and-entry.md §4):
 * theme color / dark mode / UI language are held by the App shell (TopBar
 * surfaces the controls) and are intentionally independent of the auth state.
 * Signing in or out must NEVER reset these preferences.
 *
 * `fontPreset` is owned by the settings page (Phase 4.2), not by TopBar; it is
 * therefore not exposed here.
 */

export type Theme = "warm" | "forest" | "ocean" | "plum";
export type Lang = "zh" | "en";

export interface DisplayPreferences {
  theme: Theme;
  dark: boolean;
  lang: Lang;
  setTheme: (next: Theme) => void;
  setDark: (next: boolean) => void;
  setLang: (next: Lang) => void;
}

const DEFAULTS = {
  theme: "warm" as Theme,
  dark: false,
  lang: "zh" as Lang,
};

const Context = createContext<DisplayPreferences | null>(null);

export interface DisplayPreferencesProviderProps {
  children: ReactNode;
  initial?: { theme?: Theme; dark?: boolean; lang?: Lang };
}

export const DisplayPreferencesProvider: FC<
  DisplayPreferencesProviderProps
> = ({ children, initial }) => {
  const [theme, setTheme] = useState<Theme>(initial?.theme ?? DEFAULTS.theme);
  const [dark, setDark] = useState<boolean>(initial?.dark ?? DEFAULTS.dark);
  const [lang, setLang] = useState<Lang>(initial?.lang ?? DEFAULTS.lang);

  const value = useMemo<DisplayPreferences>(
    () => ({ theme, dark, lang, setTheme, setDark, setLang }),
    [theme, dark, lang],
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
