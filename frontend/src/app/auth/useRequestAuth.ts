import { useCallback } from "react";

import { useNavigation } from "../navigation/NavigationProvider";
import { useAppRuntimeOptional } from "../runtime/AppRuntimeProvider";
import { encodePendingAction, type PendingAction } from "./pendingAction";

/**
 * Returns a `requestAuth` callback that implements the auth-and-entry §6
 * lightweight gate:
 *   - signed in: navigate to the pending action route + params directly.
 *   - signed out: navigate to `auth_login` carrying the encoded pending
 *     action so verify can restore the target.
 *
 * The hook intentionally treats `loading` like `unauthenticated` so an
 * in-flight `/me` does not allow a privileged action to slip through; once
 * `/me` resolves the user can retry from the login screen.
 */
export function useRequestAuth(): (action: PendingAction) => void {
  const { navigate } = useNavigation();
  const runtime = useAppRuntimeOptional();
  return useCallback(
    (action: PendingAction) => {
      if (runtime?.auth.status === "authenticated") {
        navigate({ name: action.route, params: action.params });
        return;
      }
      navigate({
        name: "auth_login",
        params: encodePendingAction(action),
      });
    },
    [navigate, runtime?.auth.status],
  );
}
