import {
  createContext,
  useContext,
  useMemo,
  type FC,
  type ReactNode,
} from "react";

import type { LooseRoute } from "../normalizeRoute";

/**
 * NavigationProvider exposes the App's `navigate` function to descendants so
 * pending-action triggers (`立即面试`, `复练当前轮`, etc.) can reach it
 * without prop drilling. The provider is intentionally minimal — App owns
 * the route state, this context only forwards the imperative navigate call.
 */
export interface NavigationValue {
  navigate: (next: LooseRoute) => void;
  replaceRoute: (next: LooseRoute) => void;
}

const NavigationContext = createContext<NavigationValue | null>(null);

export const NavigationProvider: FC<{
  value: Omit<NavigationValue, "replaceRoute"> &
    Partial<Pick<NavigationValue, "replaceRoute">>;
  children: ReactNode;
}> = ({ value, children }) => {
  const contextValue = useMemo<NavigationValue>(
    () => ({
      navigate: value.navigate,
      replaceRoute: value.replaceRoute ?? value.navigate,
    }),
    [value.navigate, value.replaceRoute],
  );
  return (
    <NavigationContext.Provider value={contextValue}>
      {children}
    </NavigationContext.Provider>
  );
};

export function useNavigation(): NavigationValue {
  const ctx = useContext(NavigationContext);
  if (!ctx) {
    throw new Error("useNavigation must be used inside <NavigationProvider>");
  }
  return ctx;
}
