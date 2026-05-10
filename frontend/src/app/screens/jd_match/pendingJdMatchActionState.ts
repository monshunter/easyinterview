export type PendingJdMatchSearchAction =
  | { action: "run_search"; query: string }
  | { action: "create_saved_search"; query: string; label: string };

const pendingActions = new Map<string, PendingJdMatchSearchAction>();

function randomId(): string {
  if (
    typeof crypto !== "undefined" &&
    typeof crypto.randomUUID === "function"
  ) {
    return crypto.randomUUID();
  }
  return `${Date.now()}-${Math.random().toString(36).slice(2)}`;
}

export function storePendingJdMatchAction(
  action: PendingJdMatchSearchAction,
): string {
  const id = `pending-jd-match-${randomId()}`;
  pendingActions.set(id, action);
  return id;
}

export function consumePendingJdMatchAction(
  id: string,
): PendingJdMatchSearchAction | null {
  const action = pendingActions.get(id) ?? null;
  pendingActions.delete(id);
  return action;
}

export function clearPendingJdMatchActionsForTests(): void {
  pendingActions.clear();
}
