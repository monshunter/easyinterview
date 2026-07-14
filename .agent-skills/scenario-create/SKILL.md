---
name: scenario-create
description: Create new real-environment E2E scenario cases from spec/plan/BDD documents. Generate repository-defined assets only after proving the flow drives an already running frontend/backend through real HTTP API calls or browser UI. Reject domain tests, config/lint/build gates, and mock-backed browser flows as E2E. Triggers on /scenario-create.
---

# Scenario Create Skill

Creates new scenario cases by reading a reference document and generating the
directory structure required by the repository's active suite.

## Usage

```text
/scenario-create --suite <suite-id> --ref docs/spec/<subspec>/plans/<plan>/bdd-plan.md
/scenario-create --suite <suite-id> --ref docs/spec/<subspec>/spec.md
```

## Required Inputs

| Flag | Required | Description |
|------|----------|-------------|
| `--suite <suite-id>` | Yes | Target scenario suite defined by the repo |
| `--ref <path>` | Yes | Source spec / plan / bdd-plan / bdd-test-plan document |

## Workflow

1. Read `test/scenarios/README.md`, the target suite `README.md`, and the suite `INDEX.md`.
2. Extract the suite-specific numbering, directory naming, and required file matrix from those docs.
3. Extract the suite's isolation / parallel-safety metadata contract from the suite `README.md` and `INDEX.md`.
4. Read the reference document and extract the target behavior, phase, acceptance intent, and proposed verification entrypoint.
5. Apply the **real-E2E eligibility gate before allocating an ID or creating a directory**:
   - require a concrete trigger that calls the running backend over HTTP or drives a real browser against the running frontend while business requests reach the real backend
   - require evidence from a real response, persisted state, or user-visible state
   - reject code-level domain behavior tests, `go test`, Vitest/npm test, pytest, lint, source-contract, fixture parity, build, package smoke, pure config/default wiring, and internal/tooling-only checks
   - reject fixture transport, dev mock, jsdom, or browser request interception/mocking that replaces the business backend
   - when rejected, stop with `ERROR: not a real API/UI E2E flow`, keep the Behavior ID/test in its code owner, and create no E2E ID, directory, or INDEX row
6. Allocate the next E2E scenario ID using the suite's documented numbering convention only after Step 5 passes.
7. Create the scenario directory and the full documented file set for that suite:
   - `README.md`
   - `data/seed-input.md`
   - `data/expected-outcome.md`
   - `scripts/setup.sh`
   - `scripts/trigger.sh`
   - `scripts/verify.sh`
   - `scripts/cleanup.sh`
8. Implement `trigger.sh` as a real HTTP/UI action and `verify.sh` as a check of real response/persistence/visible state. Do not place code test, lint, fixture, or build commands in either script or a helper/browser spec they invoke.
9. Default new scenarios to the suite's documented safe baseline. If the suite does not say otherwise, use `shared-cluster` + `parallel-safe=No`.
10. Make the generated shell scripts executable when the suite contract uses shell entrypoints.
11. Append the new row to the suite `INDEX.md`.

## Rules

- Prefer the active suite documented by the repo; do not invent new suite or layer directories.
- `test/scenarios/e2e/` is not a generic home for BDD or integration tests. A domain Behavior ID can be fully valid without an E2E asset.
- Prefer behavior names and user value in the slug.
- The generated README must follow the suite's documented scenario structure.
- `Ready` means the scenario is AI-agent-executable, not just documented. Do not mark or describe a scenario as `Ready` unless the full fixture + script set exists.
- Source data and expected outcome must live in `data/`; do not leave critical fixtures only in README prose.
- Unless the suite README explicitly documents stronger evidence, new scenarios start as non-parallel-safe.
- When the suite tracks isolation metadata, write the same `隔离级别` / parallel-safety decision into both the scenario `README.md` metadata and the suite `INDEX.md`.
- Cleanup must be idempotent when the suite contract requires cleanup.
- Do not invent scenario IDs, directory shapes, or scripts that are not documented by the framework or suite docs.
- Never allocate an E2E ID merely because a plan has a `BDD-Gate:`. The verification must independently pass the real-E2E eligibility gate.
- Never wrap `make test`, focused Go/Vitest/pytest tests, lint, build, fixture parity, or package smoke in E2E scripts or evidence.
- Browser automation is E2E only when it reaches the real business backend; backend mocking/interception is a code-level browser test.
- Do not generate manual-only scenarios under `test/scenarios/` unless the framework README explicitly declares a manual compatibility path.
- Keep top-level framework helpers generic. If the scenario needs suite-specific resource names, environment defaults, or assertions, place that logic under `test/scenarios/<suite-id>/_shared/`.
