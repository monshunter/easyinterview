# Expected outcome

- Replay sends exactly `{goal: "retry_current_round", sourceReportId}` for both empty generic and valid non-empty issue-backed report focus; no client focus/evidence/identity/settings field is present.
- Next sends exactly `{goal: "next_round", sourceReportId}` when frozen `hasNextRound=true`; no `roundId`, duration or successor inference is sent by the client.
- The client trusts the returned server-derived plan, starts one fresh session and navigates directly to Practice; it never re-reads mutable TargetJob/Resume/report context.
- Empty focus remains a valid generic Replay. Invalid non-empty focus fails before CTA use; server-owned projection/isolation remains composed from P0.070/P0.072.
- Terminal next-round state is accessibly disabled. Either pending CTA locks both actions and repeated clicks create at most one plan/session.
- Signed-out entry runs no report/plan/session side effect and the pending action restores the same report route.
- Raw conversation text, focus codes and evidence never enter request, URL, pending action or logs.
