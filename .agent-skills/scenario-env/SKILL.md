---
name: scenario-env
description: "IMPORTANT: Invoke this skill automatically when user asks to create, verify, cleanup, or inspect the local scenario test environment. Do NOT run environment commands without invoking this skill first. Manage the scenario environment defined by the repository's framework docs. Read `test/scenarios/README.md` and the active suite README before commands, prefer repo-tracked scripts when present, and automate setup, verification, cleanup, and status. Triggers on /scenario-env or when user asks to create, verify, cleanup, or check status of a test environment."
---

# Scenario Environment Skill

Manage the local scenario environment defined by the repository's framework docs.
Treat the framework and suite README documents as the only source of truth for
environment topology, bootstrap, verify, and cleanup behavior.

## Usage

```text
/scenario-env setup
/scenario-env verify
/scenario-env cleanup
/scenario-env cleanup --with-helpers
/scenario-env status
/scenario-env -h
```

## Reference Documents

- Framework source of truth: `test/scenarios/README.md`
- Active suite source of truth: the active suite `README.md`
- Active scenario index: the active suite `INDEX.md`

Read these documents before any environment command. Do not infer cluster topology,
component names, namespaces, helper scripts, or deploy commands from historical projects.

## Workflow Rules

- Always read the framework README and active suite README first.
- Prefer repo-tracked helper scripts when present; otherwise follow the README steps manually.
- Treat shared-environment contamination separately from scenario-specific cleanup.
- Use this recovery ladder:
  1. targeted cleanup of dirty resources
  2. targeted redeploy of the failing component or dependency
  3. full environment rebuild
- Do not jump straight to full rebuild unless the README says the environment is unrecoverable.

## Workflow: setup

### Step 1: Read framework and suite README

Read:
- `test/scenarios/README.md`
- active suite `README.md`

Capture these runtime facts from the docs before executing commands:
- environment identity / active context
- bootstrap entrypoint
- optional helper scripts
- required smoke checks

### Step 2: Check local prerequisites

Confirm the documented prerequisites are available.
If helper scripts exist, only use them when the framework or suite docs explicitly document them.

### Step 3: Run setup entrypoint

Preferred order:

1. repo-tracked helper script if the README uses it
2. the exact bootstrap command documented in the README
3. the 手动引导 (manual bootstrap) commands from the suite README — use this fallback only when the framework and suite docs explicitly declare that no repo-tracked env script exists

If none of the above are available, report the gap and stop.

Treat documented setup scripts as re-entrant unless the README explicitly says otherwise.

### Step 4: Handle result

**On success**:
- report environment identity
- report primary endpoints or access commands from the README
- report the next recommended smoke check

**On failure**:
1. identify the failing stage from setup output
2. gather the documented quick diagnostics
3. narrow the follow-up using the recovery ladder
4. only recommend full rebuild after narrower recovery steps are exhausted

## Workflow: verify

### Step 1: Read framework and suite README

Read the same two truth sources again before verification.

### Step 2: Confirm environment exists

Use the environment identity documented by the suite README.

If the target environment does not exist, report that and suggest `/scenario-env setup`.

### Step 3: Run health checks

Preferred order:

1. suite-specific verify script if documented
2. component verify helpers when documented
3. README-defined smoke checks

### Step 4: Report health summary

Summarize:
- environment presence
- core services / pods readiness
- failing components or smoke checks
- recommended next action (targeted cleanup, targeted redeploy, or rebuild)

## Workflow: cleanup

### Step 1: Read framework and suite README

Read both truth sources before cleanup.

### Step 2: Run cleanup entrypoint

Preferred order:

1. repo-tracked cleanup helper if present and documented
2. the exact cleanup sequence from the README

### Step 3: Optionally stop helper services

If `--with-helpers` is set, follow the documented helper cleanup flow when present. Otherwise, skip silently.

### Step 4: Verify cleanup result

Confirm:
- the documented cluster/context is gone, or returned to the documented clean state
- no obvious shared-resource contamination remains

## Workflow: status

### Step 1: Read framework and suite README

Read the two truth sources before summarizing status.

### Step 2: Collect status

Prefer documented status helpers. Otherwise inspect:
the documented status surfaces for the environment.

### Step 3: Summarize

Report:
- active cluster/context
- readiness summary
- obvious contamination or drift signals
- next recommended action

## Key Paths

| Path | Role |
|------|------|
| `test/scenarios/README.md` | framework truth source |
| active suite `README.md` | active suite truth source |
| active suite `INDEX.md` | scenario index |
| `test/scenarios/env-setup.sh` | optional setup entrypoint |
| `test/scenarios/env-verify.sh` | optional verify entrypoint |
| `test/scenarios/env-cleanup.sh` | optional cleanup entrypoint |
| `test/scenarios/_shared/scripts/` | optional shared helpers |
