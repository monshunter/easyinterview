# Seed Input

- The real PostgreSQL integration harness creates an isolated user, ready
  resume, multi-round TargetJob, exact round-identified plans, sessions, and a
  report row using fixed test-only UUIDv7 values.
- It then appends real `session_completed` facts in first-round, duplicate,
  out-of-order, and final-round order and reads through the production TargetJob
  store/service Get and List paths after every stage.
- Practice plan integration uses a separate fixed test user and exercises
  baseline, retry-current, next-round, equal-duration, non-contiguous sequence,
  mismatch, and all-complete transactions through the production SQL store.
- Each integration test deletes its isolated user before and after execution;
  no scenario data volume reset is required.
- `live-round-refresh-seed.sql` creates one fixed real-browser user
  (`p0-098-live-round-refresh@example.test`) with `user_settings`, a ready
  resume, one TargetJob with canonical non-contiguous sequences `1, 2, 4`, an
  exact round-1 plan, and a waiting round-1 session.
- The live browser uses the production email-code + Mailpit login path. It does
  not insert a raw auth session or expose a verification code in logs.
- The SQL fixture is removed by `live-round-refresh-cleanup.sql` before every
  seed and again by scenario cleanup; no shared volume reset is performed.
