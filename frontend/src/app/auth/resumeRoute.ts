import { normalizeRoute, type LooseRoute } from "../normalizeRoute";
import {
  decodePendingActionRoute,
  PENDING_ACTION_INTERVIEW_KEYS,
} from "./pendingAction";

export function buildResumeRoute(params: Record<string, string>): LooseRoute {
  const decoded = decodePendingActionRoute(params);
  if (decoded) return decoded;

  const resumeParams: Record<string, string> = {};
  for (const key of PENDING_ACTION_INTERVIEW_KEYS) {
    const value = params[key];
    if (value !== undefined) resumeParams[key] = value;
  }
  const returnTo = params.returnTo;
  if (returnTo) {
    const parsed = new URL(returnTo, "http://easyinterview.local");
    for (const key of PENDING_ACTION_INTERVIEW_KEYS) {
      const value = parsed.searchParams.get(key);
      if (value !== null) resumeParams[key] = value;
    }
    const candidate = parsed.pathname.replace(/^\/+/, "") || "home";
    return normalizeRoute({ name: candidate, params: resumeParams });
  }
  return { name: "home", params: resumeParams };
}
