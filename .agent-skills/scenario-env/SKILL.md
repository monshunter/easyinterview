---
name: scenario-env
description: "IMPORTANT: Invoke this skill automatically when user asks to create, start, verify, cleanup, inspect, rebuild, or redeploy the local scenario or local frontend/backend integration environment. Do NOT run environment commands without invoking this skill first. Manage the environment defined by the repository's framework docs, preferring top-level `test/scenarios/env-*.sh` scripts and root `scenario-env-*` Make targets. Read `test/scenarios/README.md` and the active suite README before commands. Triggers on /scenario-env or when user asks to setup/start/status/verify/cleanup/rebuild/redeploy a test environment."
---

# Scenario Environment Skill

Manage the local scenario/local integration environment defined by the
repository's framework docs. Treat `test/scenarios/README.md`, the active suite
README, and the top-level env scripts as the source of truth for environment
topology, bootstrap, verify, cleanup, rebuild, and redeploy behavior.

## Usage

```text
/scenario-env setup
/scenario-env verify
/scenario-env cleanup
/scenario-env cleanup --with-volumes
/scenario-env status
/scenario-env redeploy [deps|backend|frontend|all]
/scenario-env rebuild [backend|frontend|all]
/scenario-env -h
```

## Reference Documents

- Framework source of truth: `test/scenarios/README.md`
- Active suite source of truth: the active suite `README.md`
- Active scenario index: the active suite `INDEX.md`

Read these documents before any environment command. Do not infer cluster
topology, component names, namespaces, helper scripts, or deploy commands from
historical projects.

In this repo the environment is host-run: Docker Compose provides external
dependencies, backend/frontend processes are normally started by host commands,
and repo-tracked scenario runners consume that environment. Do not promise to
start long-running backend/frontend processes when required local secrets are
not present; use redeploy/rebuild for build artifacts and report the runbook
command boundary instead.

## Workflow Rules

- Always read the framework README and active suite README first.
- Prefer top-level repo-tracked environment scripts when present:
  `test/scenarios/env-setup.sh`, `test/scenarios/env-status.sh`,
  `test/scenarios/env-verify.sh`, `test/scenarios/env-cleanup.sh`, and
  `test/scenarios/env-redeploy.sh`.
- Use root Make aliases when a user asks for Makefile integration:
  `make scenario-env-setup`, `make scenario-env-status`,
  `make scenario-env-verify`, `make scenario-env-cleanup`, and
  `make scenario-env-redeploy`.
- Do not extract shared environment bootstrap from a specific scenario
  directory. Specific scenario scripts may consume the environment, but they do
  not own shared setup/status/verify/cleanup/redeploy.
- Treat shared-environment contamination separately from scenario-specific
  cleanup.
- Use this recovery ladder:
  1. targeted cleanup of dirty resources
  2. targeted redeploy of the failing component or dependency
  3. full environment rebuild
- Do not jump straight to full rebuild unless the README says the environment
  is unrecoverable.

## Workflow: setup

1. Read `test/scenarios/README.md` and the active suite `README.md`.
2. Capture environment identity, bootstrap entrypoint, optional helper scripts,
   and required smoke checks from the docs.
3. Check documented local prerequisites.
4. Run setup in this order:
   - `test/scenarios/env-setup.sh`
   - `make scenario-env-setup` when the user explicitly wants the Makefile entrypoint
   - the exact bootstrap command documented in the README
   - the 手动引导 (manual bootstrap) commands from the suite README, only when no
     repo-tracked env script exists
5. On success, report environment identity, primary endpoints/access commands,
   and the next recommended smoke check.
6. On failure, identify the failing stage, gather documented quick diagnostics,
   and narrow recovery through the ladder above.

Treat documented setup scripts as re-entrant unless the README explicitly says
otherwise.

## Workflow: verify

1. Read the same two truth sources again before verification.
2. Confirm the documented environment exists; if not, suggest `/scenario-env setup`.
3. Run health checks in this order:
   - `test/scenarios/env-verify.sh`
   - `make scenario-env-verify` when the user explicitly wants the Makefile entrypoint
   - suite-specific verify script if documented
   - component verify helpers when documented
   - README-defined smoke checks
4. Summarize environment presence, readiness, failing components, and the next
   action.

## Workflow: cleanup

1. Read `test/scenarios/README.md` and the active suite README.
2. Run cleanup in this order:
   - `test/scenarios/env-cleanup.sh`
   - `make scenario-env-cleanup` when the user explicitly wants the Makefile entrypoint
   - the exact cleanup sequence from the README
3. Use `--with-volumes` only when the user explicitly asks to reset/delete
   shared volumes or the README says a full reset is required. Default cleanup
   preserves named volumes.
4. If the README documents helper services outside the default env scripts,
   follow that helper cleanup flow when requested. Otherwise skip silently.
5. Verify that the environment returned to the documented clean state and that
   no obvious shared-resource contamination remains.

## Workflow: status

1. Read `test/scenarios/README.md` and the active suite README.
2. Collect status in this order:
   - `test/scenarios/env-status.sh`
   - `make scenario-env-status` when the user explicitly wants the Makefile entrypoint
   - the documented status surfaces for the environment
3. Report active context, readiness summary, contamination/drift signals, and
   next recommended action.

## Workflow: redeploy / rebuild

Use this workflow when the user asks to rebuild, redeploy, refresh, or recompile
the local scenario/local integration environment.

1. Read `test/scenarios/README.md` and the active suite README.
2. Resolve the requested target:
   - `deps`: refresh Docker Compose external dependencies via env redeploy.
   - `backend`: refresh backend build artifacts.
   - `frontend`: refresh frontend build artifacts.
   - `all`: refresh dependencies, backend, and frontend.
3. Run in this order:
   - `test/scenarios/env-redeploy.sh [deps|backend|frontend|all]`
   - `make scenario-env-redeploy TARGET=<target>`
   - suite-documented exact rebuild/redeploy command
4. In the current host-run topology, redeploy/rebuild is not a Kind, Helm, or
   cluster rollout. It refreshes local dependencies and build artifacts. If the
   user needs a long-running backend/frontend process for a hybrid real-provider
   scenario, point to the documented scenario README command and keep secrets in
   local ignored files.
5. Run `/scenario-env verify` after redeploy unless the user asked for dry-run
   or inspection-only status.

## Key Paths

| Path | Role |
|------|------|
| `test/scenarios/README.md` | framework truth source |
| active suite `README.md` | active suite truth source |
| active suite `INDEX.md` | scenario index |
| `test/scenarios/env-setup.sh` | shared environment setup entrypoint |
| `test/scenarios/env-status.sh` | shared environment status entrypoint |
| `test/scenarios/env-verify.sh` | shared environment verify entrypoint |
| `test/scenarios/env-cleanup.sh` | shared environment cleanup entrypoint |
| `test/scenarios/env-redeploy.sh` | shared host-run rebuild/redeploy entrypoint |
| `make scenario-env-redeploy` | Makefile alias for rebuild/redeploy |
| `test/scenarios/_shared/scripts/` | optional shared helpers |
