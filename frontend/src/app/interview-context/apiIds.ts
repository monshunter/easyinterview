import { isServerId } from "../../lib/ids";

export function normalizeServerBoundId(id?: string | null): string | undefined {
  const value = id?.trim();
  if (!value || !isServerId(value)) return undefined;
  return value;
}
