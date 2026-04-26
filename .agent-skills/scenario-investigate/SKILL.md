---
name: scenario-investigate
description: Investigate scenario test failures by checking test design, environment, and code issues in priority order. Triggers on /scenario-investigate.
---

# Scenario Investigate Skill

Investigates scenario test failures by systematically checking potential causes in priority order.

## Trigger Command

```
/scenario-investigate E2E.P0.003                               # Investigate latest failure
/scenario-investigate E2E.P0.003 --run 20260120-153045-a1b2    # Investigate specific run
```

## Reference Documents

- Scenario contract: the failing scenario's `README.md`
- Framework truth source: `test/scenarios/README.md`
- Suite truth source: `test/scenarios/e2e/README.md` and `INDEX.md`

Use these documents for concrete environment topology, cleanup expectations, consumer surfaces,
and suite-specific troubleshooting commands. Keep this skill focused on investigation order and
decision points.

## Workflow

### Step 1: Gather Failure Context

1. Locate the run output directory:
   - Latest: `{TEST_OUTPUT_DIR}/latest/e2e/{scenario_id}/`
   - Specific run: `{TEST_OUTPUT_DIR}/runs/{run_id}/e2e/{scenario_id}/`

2. Read failure artifacts:
   - `result.json` - Scenario result with failed step info
   - `steps/*.log` - Step execution logs
   - Scenario `README.md` - Expected behavior reference
   - Framework README and suite README - environment behavior, cleanup rules, and live verification reference
   - Compare elapsed runtime with any documented long-running / expected-duration notes before classifying the symptom as a hang

### Step 2: Priority 1 - Check the Test Itself

Most failures are test issues. Check:

- Is the scenario design correct?
- Is the test data valid?
- Are the steps in correct order?
- Are the expected results reasonable?
- Is the wait time sufficient?
- Was setup executed?
- Is the scenario simply still within the suite-documented duration budget for a known long-running case?

### Step 3: Priority 2 - Check the Environment and Shared Control Plane

If the scenario design looks correct, investigate environment-level causes before blaming code:

- Is the failure explained by incomplete cleanup, residual scenario resources, or blocked convergence?
- Is the failure coming from a shared environment component rather than the scenario under test?
- Do the framework/suite README documents describe a known cleanup verification step or contamination source that was skipped?
- If the symptom looks like “controller stuck”, check manager logs for worker count, repeated wait loops, and whether the affected object key is blocked by a long-running reconcile instead of a dead controller.
- For StatefulSet ordinal pods or other same-name replacement patterns, check whether the wait path is observing a stale cached Pod view instead of a fresh read of the current UID.

Use the framework README and suite README for the concrete commands and suite-specific signals.

### Step 4: Priority 3 - Check Consumer Artifact Drift

If the environment looks healthy, verify that the test is hitting current consumer artifacts instead of stale ones:

- locally built binaries
- generated clients or schema artifacts
- deployed components
- user-facing entrypoints that proxy or front the underlying service
- deployed controller image / chart revision / manager flags when the suspected issue is runtime-behavior drift

Treat consumer drift as distinct from code bugs. Use the suite README to identify the concrete surfaces that must align in that environment.

### Step 5: Priority 4 - Check for Code Bugs

Only after the test contract, environment, and consumer artifacts look correct should you attribute the failure to product code.

### Step 6: Generate Investigation Report

Write `investigation.json` to the scenario output directory:

```json
{
  "investigation": {
    "scenario_id": "E2E.P0.003",
    "run_id": "20260120-153045-a1b2",
    "investigated_at": "2026-01-20T16:00:00Z",
    "failure_symptom": "Description of what failed",
    "investigation_steps": [
      {
        "priority": 1,
        "area": "test",
        "findings": "Summary of test-level findings"
      }
    ],
    "root_cause": "Identified root cause",
    "recommendation": "Recommended fix or next action"
  }
}
```
