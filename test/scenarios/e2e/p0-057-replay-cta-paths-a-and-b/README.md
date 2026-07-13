# E2E.P0.057 — Server-owned Replay / next-round CTA paths

> **Owner**: frontend-report-dashboard/001-report-screen-and-generating-handoff
> **Coverage tags**: C-3 / C-4 / C-11
> **Execution**: six focused frontend Vitest files

## Given / When / Then

- **Given** a ready report whose first action is Replay and whose retry focus is either an explicit empty array (generic same-round practice) or a non-empty set backed by a same-code `needs_work` dimension and issue; another report has `hasNextRound=true` and a next-round action.
- **When** the authenticated user chooses Replay or Next, including repeated clicks and signed-out pending-action recovery.
- **Then** the client sends exactly `{goal, sourceReportId}`, starts one fresh server-derived plan/session and navigates directly to Practice. It never sends competency/dimension focus, evidence gaps, identity, round settings, duration, language, persona or difficulty.

An empty focus is a valid generic Replay, not a missing-data failure. A non-empty focus remains report-local and issue-backed, but the backend alone projects it into the derived plan. Registry-owned P0.070/P0.072 provide the server projection/isolation evidence; this UI scenario proves the client cannot become a second focus authority.

## Covered paths

- Replay with empty generic focus and Replay with valid non-empty issue-backed focus use the same closed request shape.
- Next uses only frozen `context.hasNextRound` for availability and sends no round identity.
- Either in-flight CTA locks both buttons; repeated clicks create at most one plan and session.
- Signed-out entry is auth-gated before report reads/mutations and returns to the same `reportId` route.
- Missing/invalid/nonready source and server projection mismatch fail without client fallback; backend details belong to P0.070/P0.072.

## Privacy

No answer/question/hint text, anchors, focus codes, evidence text, mutable target/resume fields or provider output enters route params, pending action or `createPracticePlan` request.
