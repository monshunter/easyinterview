import { createContext, useContext, type FC, type ReactNode } from "react";

import type { LooseRoute } from "../normalizeRoute";

/**
 * NavigationProvider exposes the App's `navigate` function to descendants so
 * pending-action triggers (`立即面试`, `复练当前轮`, etc.) can reach it
 * without prop drilling. The provider is intentionally minimal — App owns
 * the route state, this context only forwards the imperative navigate call.
 */
export interface NavigationValue {
  navigate: (next: LooseRoute) => void;
}

const NavigationContext = createContext<NavigationValue | null>(null);

export const NavigationProvider: FC<{
  value: NavigationValue;
  children: ReactNode;
}> = ({ value, children }) => (
  <NavigationContext.Provider value={value}>
    {children}
  </NavigationContext.Provider>
);

export function useNavigation(): NavigationValue {
  const ctx = useContext(NavigationContext);
  if (!ctx) {
    throw new Error("useNavigation must be used inside <NavigationProvider>");
  }
  return ctx;
}
