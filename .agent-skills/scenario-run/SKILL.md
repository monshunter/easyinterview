---
name: scenario-run
description: Execute repository-defined real API/UI E2E scenarios. Before environment setup or scenario execution, reject any scenario that wraps Go/Vitest/pytest/lint/build code gates or replaces the business backend with browser mocks/interception. Supports single, suite, rerun, and hybrid modes; parallelize only explicitly parallel-safe scenarios. Triggers on /scenario-run.
---

# Scenario Run Skill

Execute scenario tests following the repository's active scenario framework.

## Prerequisites

- The local scenario environment is prepared and verified before scenario scripts run. Prefer `test/scenarios/env-setup.sh` followed by `test/scenarios/env-verify.sh`, or the equivalent repo-tracked entrypoints documented by the framework.
- Read `test/scenarios/README.md` before first use
- If the framework or suite README documents expected duration / long-running scenarios, treat that budget as the first-line signal for “still running vs actually stalled”

## Reference Documents

- Framework truth source: `test/scenarios/README.md`
- Active suite source of truth: the selected suite `README.md`
- Active suite index: the selected suite `INDEX.md`
- Scenario truth source: the scenario's own `README.md`

Use these documents for scenario semantics, prerequisites, troubleshooting, cleanup
expectations, and contamination rules. Keep this skill focused on execution protocol.

## Usage

```text
/scenario-run -i E2E.P0.001
/scenario-run -s <suite-id>
/scenario-run -s <suite-id> --from E2E.P0.003
/scenario-run -r {run-id}
```

| Flag | Description |
|------|-------------|
| `-i` | Scenario ID (for example `E2E.P0.001`) |
| `-s` | Suite ID defined by the repo |
| `--from` | When used with `-s`, skip earlier scenarios |
| `-r` | Rerun failed scenarios from a previous run |
| `--batch-size` | Max parallel sub-agents per batch (default: 8) |

Execution modes:
- **Direct Mode**: default mode; execute sequentially
- **Parallel Mode**: use only when every resolved scenario is explicitly marked `parallel-safe`

## Workflow

### Step 1: Parse arguments and resolve scenarios

**Single scenario** (`/scenario-run -i E2E.P0.001`):
- read the target suite `INDEX.md`
- locate the scenario row by ID
- resolve the directory from the linked path in that row

**Suite mode** (`/scenario-run -s <suite-id>`):
- read the target suite `INDEX.md`
- collect scenarios with status `Ready` or `Verified`
- collect isolation / parallel-safety metadata when the suite `INDEX.md` provides it
- if `--from` is present, skip earlier IDs

**Rerun mode** (`/scenario-run -r {run-id}`):
- read `{TEST_OUTPUT_DIR}/runs/{run_id}/summary.json`
- collect failed or errored scenarios

### Step 2: Initialize run

1. Set `TEST_OUTPUT_DIR` (default `.test-output/`)
2. Generate or load a run ID
3. Create run output directories
4. Read the framework README and target suite README once before execution
5. If `test/scenarios/_shared/scripts/common.sh` exists, prefer its helpers for run initialization and result writing
6. Surface any documented long-running scenarios or expected duration notes before execution starts
7. Run the **E2E eligibility preflight before starting the environment or executing scripts** for every resolved scenario:
   - read its README plus `setup.sh`, `trigger.sh`, `verify.sh`, and `cleanup.sh`
   - follow repo-tracked helpers, browser specs, and commands invoked by those scripts; eligibility cannot be hidden one call away
   - require `trigger.sh` to drive a real HTTP API or browser UI against the running product and require `verify.sh` to assert a real response, persistence result, or user-visible state
   - if `trigger.sh` or `verify.sh` invokes `go test`, Vitest/npm test, pytest, lint, source-contract, fixture parity, build, package smoke, or root `make test`, mark the scenario `ERROR` with reason `not a real E2E: wraps code-level gate`; do not execute it as E2E
   - if a browser flow or referenced browser spec uses fixture transport, dev mock, jsdom, route fulfillment/interception, or request mocking to replace the business backend, mark it `ERROR` with reason `not a real E2E: backend is mocked`; do not execute it as E2E
   - do not reinterpret a rejected scenario as PASS, MANUAL_REQUIRED, or a domain behavior test. Report that its code-level test should remain with the code owner and that the stale E2E asset must be removed or redesigned
8. Run environment preflight only for scenarios that passed Step 7:
   - prefer `test/scenarios/env-setup.sh`, adding framework/suite documented arguments such as `--with-migrations` when required by the target scenario README
   - then run `test/scenarios/env-verify.sh`
   - if a suite or scenario documents additional local configuration, such as real provider env files or host-run backend/frontend commands, treat that as scenario setup evidence; do not invent secrets or start long-running processes without the required local env
9. Determine direct vs parallel mode:
   - default to serial execution
   - only enable parallel mode when the framework and suite docs explicitly mark every resolved scenario as `parallel-safe`
   - if any resolved scenario is `shared-cluster`, keep the run serial
   - if any resolved scenario is `exclusive`, run that scenario alone with no overlap

Run artifact layout:

```text
.test-output/runs/{run-id}/{suite-id}/{scenario-id}/
```

### Step 3A: Execute scenarios — direct mode

For each scenario that passed Step 2 eligibility (skip rejected scenarios and preserve their `ERROR` result):

1. Read the scenario `README.md`
2. Treat `scripts/` as the execution contract for repo-defined scenarios
3. If a `Ready` or `Verified` scenario is missing any of `setup.sh`, `trigger.sh`, `verify.sh`, or `cleanup.sh`, stop and mark the scenario as `ERROR`; do not fall back to manual verification
4. Reconfirm the Step 2 eligibility result immediately before execution. If scripts changed after preflight or runtime output reveals a forbidden code-test wrapper/backend mock, stop, mark `ERROR`, and do not treat its output as E2E evidence.
5. Run `scripts/setup.sh`
6. Run `scripts/trigger.sh`
7. Run `scripts/verify.sh`
8. Always run `scripts/cleanup.sh`, even after failure
9. Verify cleanup expectations from the framework/suite/scenario docs
10. Record the scenario result in the run output. If the scenario writes a valid result artifact with `result=MANUAL_REQUIRED`, keep that state distinct from PASS/FAIL/ERROR.

Notes:
- For suite-documented long-running scenarios, lack of new stdout between phase boundaries is not by itself an execution failure signal.
- Do not escalate to environment rebuild only because a scenario is “taking a while” until it exceeds the documented duration budget.

### Step 3B: Execute scenarios — parallel mode

#### 3B.1 Batch preparation

Only enter this section when Step 2 determined that the eligible scenario set is explicitly `parallel-safe`. Do not dispatch scenarios already rejected by the eligibility preflight.

Split the resolved scenarios into batches of `N` (`--batch-size`, default `8`).

#### 3B.2 Dispatch each batch

For each scenario in the batch, dispatch a sub-agent with this prompt shape:

```text
You are executing a single repository-defined scenario integration test.

Scenario:
- ID: {scenario_id}
- Directory: {scenario_dir}
- README: {scenario_dir}/README.md
- Output directory: {output_dir}/{scenario_id}

Run context:
- run_id: {run_id}
- suite: {suite_id}
- TEST_OUTPUT_DIR: {test_output_dir}

Execution protocol:
1. Read the scenario README and the framework/suite README context already provided.
2. If `test/scenarios/_shared/scripts/common.sh` exists, source it at the start of every shell call.
3. Inspect `scripts/` first and treat those files as the execution contract.
4. Reapply the real-E2E eligibility gate: fail as `ERROR` without execution if scripts wrap Go/Vitest/pytest/lint/build/root-make-test gates or if browser mocks/interception replace the business backend.
5. If the scenario is marked `Ready` or `Verified` but any required script is missing, fail the scenario as `ERROR` instead of inventing a manual fallback.
6. Run `setup.sh`, `trigger.sh`, `verify.sh`, `cleanup.sh` in order.
7. Always perform cleanup verification after cleanup runs.
8. Write a final result artifact to the output directory.
9. For `hybrid` scenarios, AI Agent execution comes first: run the scripts and accept a `MANUAL_REQUIRED` result when local real-provider credentials, browser actions, or human observation evidence are intentionally missing. Do not convert that state into PASS or ERROR.
```

After all sub-agents finish:
- read each scenario result
- mark missing/invalid result artifacts as `ERROR`
- report batch progress

### Step 4: Handle failures

If any scenario has result `FAIL`:
- recommend or trigger `/scenario-investigate {scenario_id}`

If any scenario has result `ERROR`:
- report which scenarios had execution/tooling errors
- distinguish real-E2E eligibility errors from environment/tool failures; eligibility errors require deleting or redesigning the stale E2E asset, not rerunning it unchanged
- suggest rerunning those scenarios individually

If any scenario has result `MANUAL_REQUIRED`:
- report that the AI Agent preflight completed but human/browser-agent evidence remains
- point to the scenario README, checklist, and output directory
- do not call it passed until the same scenario result artifact reports PASS

If cleanup output indicates shared-environment contamination:
- prefer targeted cleanup first
- only escalate to full environment rebuild when narrower recovery steps fail

If a scenario appears stalled:
- first compare elapsed time with the framework/suite README duration notes
- only after exceeding that budget should you recommend `/scenario-investigate` or broader environment recovery

### Step 5: Generate summary

Summarize:
- run ID
- passed / failed / errored scenarios
- manual-required hybrid scenarios
- cleanup contamination findings
- paths to the run artifacts

## Output Conventions

- Prefer framework helper functions for result files when available
- Otherwise write plain JSON artifacts that at minimum include `scenario_id`, `result`, and `error` (if any). Valid result states are `PASS`, `FAIL`, `ERROR`, and `MANUAL_REQUIRED`.
- Keep scenario output under `{TEST_OUTPUT_DIR}/runs/{run_id}/{suite_id}/{scenario_id}/`
