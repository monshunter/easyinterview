---
name: scenario-create
description: Create new scripted BDD scenario cases from spec/plan documents. Generate the repository-defined scenario asset set for the active suite, including fixtures, executable scripts, and isolation metadata, and append INDEX entries. Use when adding scenario coverage or scaffolding new scenario directories. Triggers on /scenario-create.
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
4. Read the reference document and extract the target behavior, phase, and acceptance intent.
5. Allocate the next scenario ID using the suite's documented numbering convention.
6. Create the scenario directory and the full documented file set for that suite:
   - `README.md`
   - `data/seed-input.md`
   - `data/expected-outcome.md`
   - `scripts/setup.sh`
   - `scripts/trigger.sh`
   - `scripts/verify.sh`
   - `scripts/cleanup.sh`
7. Default new scenarios to the suite's documented safe baseline. If the suite does not say otherwise, use `shared-cluster` + `parallel-safe=No`.
8. Make the generated shell scripts executable when the suite contract uses shell entrypoints.
9. Append the new row to the suite `INDEX.md`.

## Rules

- Prefer the active suite documented by the repo; do not invent new suite or layer directories.
- Prefer behavior names and user value in the slug.
- The generated README must follow the suite's documented scenario structure.
- `Ready` means the scenario is AI-agent-executable, not just documented. Do not mark or describe a scenario as `Ready` unless the full fixture + script set exists.
- Source data and expected outcome must live in `data/`; do not leave critical fixtures only in README prose.
- Unless the suite README explicitly documents stronger evidence, new scenarios start as non-parallel-safe.
- When the suite tracks isolation metadata, write the same `隔离级别` / parallel-safety decision into both the scenario `README.md` metadata and the suite `INDEX.md`.
- Cleanup must be idempotent when the suite contract requires cleanup.
- Do not invent scenario IDs, directory shapes, or scripts that are not documented by the framework or suite docs.
- Do not generate manual-only scenarios under `test/scenarios/` unless the framework README explicitly declares a manual compatibility path.
- Keep top-level framework helpers generic. If the scenario needs suite-specific resource names, environment defaults, or assertions, place that logic under `test/scenarios/<suite-id>/_shared/`.
