---
name: scenario-run
description: Execute repository-defined scenario tests. Supports single scenario, full suite, and rerun modes. Use parallel sub-agent dispatch only when the framework and suite docs explicitly mark the target scenarios as `parallel-safe`; otherwise default to serial execution. Read the framework README and target suite README before execution. Triggers on /scenario-run.
---

# Scenario Run Skill

Execute scenario tests following the repository's active scenario framework.

## Prerequisites

- The local scenario environment is ready and has passed verification (`/scenario-env setup` followed by `/scenario-env verify`, or the equivalent repo-tracked env entrypoints documented by the framework)
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
7. Determine direct vs parallel mode:
   - default to serial execution
   - only enable parallel mode when the framework and suite docs explicitly mark every resolved scenario as `parallel-safe`
   - if any resolved scenario is `shared-cluster`, keep the run serial
   - if any resolved scenario is `exclusive`, run that scenario alone with no overlap

Run artifact layout:

```text
.test-output/runs/{run-id}/{suite-id}/{scenario-id}/
```

### Step 3A: Execute scenarios — direct mode

For each scenario:

1. Read the scenario `README.md`
2. Treat `scripts/` as the execution contract for repo-defined scenarios
3. If a `Ready` or `Verified` scenario is missing any of `setup.sh`, `trigger.sh`, `verify.sh`, or `cleanup.sh`, stop and mark the scenario as `ERROR`; do not fall back to manual verification
4. Run `scripts/setup.sh`
5. Run `scripts/trigger.sh`
6. Run `scripts/verify.sh`
7. Always run `scripts/cleanup.sh`, even after failure
8. Verify cleanup expectations from the framework/suite/scenario docs
9. Record the scenario result in the run output

Notes:
- For suite-documented long-running scenarios, lack of new stdout between phase boundaries is not by itself an execution failure signal.
- Do not escalate to environment rebuild only because a scenario is “taking a while” until it exceeds the documented duration budget.

### Step 3B: Execute scenarios — parallel mode

#### 3B.1 Batch preparation

Only enter this section when Step 2 determined that the resolved scenario set is explicitly `parallel-safe`.

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
4. If the scenario is marked `Ready` or `Verified` but any required script is missing, fail the scenario as `ERROR` instead of inventing a manual fallback.
5. Run `setup.sh`, `trigger.sh`, `verify.sh`, `cleanup.sh` in order.
6. Always perform cleanup verification after cleanup runs.
7. Write a final result artifact to the output directory.
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
- suggest rerunning those scenarios individually

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
- cleanup contamination findings
- paths to the run artifacts

## Output Conventions

- Prefer framework helper functions for result files when available
- Otherwise write plain JSON artifacts that at minimum include `scenario_id`, `result`, and `error` (if any)
- Keep scenario output under `{TEST_OUTPUT_DIR}/runs/{run_id}/{suite_id}/{scenario_id}/`
