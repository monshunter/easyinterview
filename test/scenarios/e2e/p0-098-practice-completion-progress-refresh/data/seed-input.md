# Seed Input

- `live-round-refresh-seed.sql` creates one fixed real-browser user
  (`p0-098-live-round-refresh@example.test`) with `user_settings`, a ready
  resume, one TargetJob with canonical non-contiguous sequences `1, 2, 4`, an
  exact round-1 plan, and a waiting round-1 session containing one completed
  user/assistant turn so the real completion endpoint is reportable.
- The browser uses the production email-code and Mailpit login path. It does not
  insert a raw auth session or expose a verification code in logs.
- After completing round 1, the browser selects the ready TargetJob card body
  once from Workspace and once from Home. Both destinations are the read-only
  `/workspace?targetJobId=019f6098-0000-7000-8000-000000000003` detail.
- The seeded Chinese UI language makes the detail-state labels deterministic:
  `已进行`, `即将进行`, `未进行`.
- `live-round-refresh-cleanup.sql` removes the fixed rows before every seed and
  again during scenario cleanup; no shared volume reset is performed.
