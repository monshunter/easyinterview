---
name: plan-review
description: "Review or fix spec/plan/checklist/context documents for a spec-centric plan target. Use when the user wants L1 review or document remediation for spec/plan/checklist consistency, including spec-owned issues that should be repaired in the same fix pass. Reuses implement-owned shared context validator to resolve the target docs, then performs semantic review directly on the validated markdown instead of running parser-heavy format checkers. Supports /plan-review [subspec/plan] [target] [--fix]."
---

# Plan Review Skill

L1 document review for `spec ↔ plan ↔ checklist ↔ context.yaml`. This skill owns
review output and document remediation. It uses the shared `context.yaml`
validator to resolve the current target, then reads the validated markdown
directly. It does **not** implement code and does **not** hand off to `/tdd`.

## Usage

- `/plan-review` - List latest candidate plans and let user choose, then run review
- `/plan-review <subspec>/<plan>` - Review the default target of the named spec-centric plan
- `/plan-review <subspec>/<plan> <target>` - Review a specific target
- `/plan-review <subspec>/<plan> [target] --fix` - Preview, confirm, and apply spec/plan/checklist/context fixes
- `/plan-review -h` - Show help only
- `/plan-review -h -v` - Show verbose help (including workflow)

Flag rules:

- `--fix` requires an explicit plan name
- No-plan mode exists only for review, not for fix
- Review is advisory and read-only

## Shared Resources

Use the implement-owned shared resources:

- `.agent-skills/implement/shared/scripts/list_context_candidates.py`
- `.agent-skills/implement/shared/scripts/validate_context.py`
- `.agent-skills/implement/shared/references/plan-context-contract.md`

Reviewer rule:

- Trust the post-`doc-init` templates for markdown structure unless the user is
  explicitly fixing template drift.
- Do not add python markdown-format validators or parser-only gatekeeping for
  plan/checklist/spec documents.
- New plan docs are sequential-only by default.

## Workflow

### Step 0: Handle help flags

- `-h`/`--help`: show skill name, description, usage, then stop.
- `-h -v`/`--help --verbose`: show usage + full workflow, then stop.

### Step 1: Resolve plan name

**With argument**:

1. Check if `docs/spec/{subspec}/plans/{plan}/` exists when the argument contains `/`.
2. If the argument is a bare name, fuzzy-match against spec-centric candidates from `docs/spec/*/plans/*/context.yaml`.
3. If multiple matches, ask the user to choose.
4. If no match, stop and report available plan names.

**Without argument**:

1. Run:

```bash
python3 .agent-skills/implement/shared/scripts/list_context_candidates.py \
  --plan-root docs
```

2. Display numbered candidates with reasons.
3. Ask user to select one number.
4. If no candidates, show script output and stop.
5. If input is invalid, re-display the list and wait.

### Step 2: Read manifest and determine target scope

1. Read `docs/spec/{subspec}/plans/{plan}/context.yaml`.
2. Determine target:
   - If user passed `<target>`, use it.
   - Otherwise use `spec.defaultTarget`.
3. If target is not defined in `spec.targets`, stop and show available targets.

### Step 3: Validate manifest and collect normalized file set

Run:

```bash
python3 .agent-skills/implement/shared/scripts/validate_context.py \
  --context docs/spec/{subspec}/plans/{plan}/context.yaml \
  --docs-root docs \
  --target {target}
```

Use the shared plan-context contract to interpret `files[]`.

Validation scope is limited to manifest shape, target selection, path boundary,
and referenced markdown existence. After validation, read the returned markdown
files directly; do not run separate markdown structure checkers.

### Step 4: Load current target context

Read the validated files for the current target:

- `plan`
- `checklist`
- `spec` when present
- `test-plan`, `test-checklist`, `bdd-plan`, `bdd-checklist`, and `reference` files when present
- `docs/spec/{subspec}/plans/{plan}/context.yaml` itself when the finding or fix involves target
  wiring, branch metadata, or target discovery fields

### Step 5: Branch by mode

**Review mode**:

1. Read the validated markdown/context set and run the full L1 semantic review.
2. Treat document structure issues as findings only when they materially harm
   readability, consistency, or execution.
3. Produce the final report, then stop.

**Fix mode**:

1. Generate semantic fix suggestions across the current target's
   `spec`/`plan`/`checklist`/`context.yaml` documents.
2. Display the preview grouped by document type and wait for user confirmation.
3. On confirmation:
   - write only the fixes explicitly confirmed by the user
   - if both spec and downstream docs need updates, write spec first, then
     update plan/checklist/context consumers
4. Re-read the changed files and summarize the post-fix status, then stop.

### Step 6: L1 semantic analysis

Baseline rule:

- Always execute the full semantic review once the target docs are validated.
- Do not pre-filter the review through deterministic markdown-format checks.

Review dimensions:

- `Consistency`: terminology, state model, target scoping, truth-source boundaries, fit with repo patterns
- `Completeness`: success path, error path, boundary cases, non-goals, verification/test coverage
- `Best Practice`: security/privacy/authz, minimal scope, idempotency, concurrency, testability, maintainability

Baseline checks:

- `S-001`: spec section coverage
- `S-002`: error path coverage
- `S-003`: description alignment
- `S-004`: orphan judgment
- `S-005`: test completion gates are execution-based; flag plan/checklist items that use raw code coverage percentages as completion, commit, or phase-exit criteria

Extension review:

- `X-L1-Value`: practical value / operator workflow closure
- `X-L1-Landing`: delivery and landing feasibility
- `X-L1-Risk`: security, privacy, and operational risk

### Step 7: Report contract

Final L1 report must use this structure:

1. `Findings`
2. `Dimension Coverage`
3. `Strengths`
4. `Open Questions / Assumptions`
5. `Optimization Opportunities`

Rules:

- Order findings by severity.
- Each finding must cite impact, evidence, and remediation direction.
- If a dimension has no material finding, explicitly say `No material finding`.
- Do not merge `Strengths` or `Optimization Opportunities` into `Findings`.

## Fix Guardrails

- `--fix` without plan name is an error.
- Fix mode only updates the current target's `spec` / `plan` / `checklist` /
  optional review docs / `context.yaml`; it never edits source code.
- Do not edit documents outside the current target's validated file set plus the
  selected `context.yaml`.
- If a fix would change design or plan semantics beyond the user's confirmed intent rather than repair drift or requested document-owned issues, stop and confirm with the user first.
