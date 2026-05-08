---
name: implement
description: "IMPORTANT: Invoke this skill automatically when the user asks to implement an existing plan. Do NOT implement a plan without invoking this skill first. Thin plan execution entry point: resolves a spec-centric plan at docs/spec/<subspec>/plans/<plan>/context.yaml, validates referenced content through implement-owned shared scripts, reads the referenced markdown directly, then hands checklist execution to /tdd. Triggers on /implement or when user says 'implement plan', 'execute plan', 'start implementing', 'continue implementing', or wants to resume an in-flight plan, including terse follow-ups like '继续' when the current session or branch already identifies the plan owner. Supports /implement <subspec>/<plan> [target] syntax."
---

# Implement Skill

Thin entry point for executing an existing plan. `/implement` owns plan selection,
`context.yaml` validation, context summary, sequential execution routing,
handoff to `/tdd`, and delivery close-out. It does **not** own L1/L2 review and
does **not** run extra markdown format checkers.

## Usage

- `/implement` - List latest candidate plans and let user choose
- `/implement <subspec>/<plan>` - Implement the default target of the named spec-centric plan
- `/implement <subspec>/<plan> <target>` - Implement a specific target (for example `frontend`, `unit-test`)
- `/implement` - Resume the current in-flight plan when the current session or branch already identifies the owner
- `/implement -h` - Show help only, do not execute workflow
- `/implement -h -v` - Show verbose help (including full workflow), do not execute

## Shared Resources

`/implement` owns the shared plan-context resources used by sibling skills:

- `.agent-skills/implement/shared/scripts/list_context_candidates.py`
- `.agent-skills/implement/shared/scripts/validate_context.py`
- `.agent-skills/implement/shared/scripts/generate_context_yaml.py`
- `.agent-skills/implement/shared/scripts/detect_session_branch.py`
- `.agent-skills/implement/shared/references/plan-context-contract.md`

Use the shared reference for common file-role mapping and error templates.

## Removed Legacy Flags

The following `/implement` flags no longer exist:

- `--review`
- `--fix`
- `--review-code`
- `--fix-code`

If the user asks for one of these removed modes, stop and redirect:

- L1 document review/fix → `/plan-review`
- L2 code review/fix → `/plan-code-review`

## Prerequisites

- Plan directory exists under `docs/spec/<subspec>/plans/<plan>/`
- Manifest exists at `docs/spec/<subspec>/plans/<plan>/context.yaml`
- `python3` is available and can run bundled scripts
- `PyYAML` is installed (required by shared validator and candidate scripts)

## Workflow

### Step 0: Handle help flags

- If user passes `-h` or `--help`, show: skill name, description, usage.
- If user passes `-h -v` or `--help --verbose`, also show full workflow.
- In both cases, stop after help output.
- If the user passes any removed legacy review flag, stop and tell them to use `/plan-review` or `/plan-code-review`.

### Step 1: Resolve plan name

**With argument** (`/implement target-workspace/001-frontend`):

1. Check if `docs/spec/{subspec}/plans/{plan}/` exists when the argument contains `/`.
2. If the argument is a bare name, fuzzy-match against spec-centric candidates from `docs/spec/*/plans/*/context.yaml`.
3. If multiple matches, list candidates and ask user to choose.
4. If no match, stop and report available plan names.

**Without argument** (`/implement`):

1. Run:

```bash
python3 .agent-skills/implement/shared/scripts/list_context_candidates.py \
  --plan-root docs
```

2. Display numbered candidates with reasons.
3. Ask user to select one number.
4. If no candidates, show script output and stop.
5. If input is invalid, re-display list and wait (do not proceed).

### Step 2: Read manifest and determine target scope

1. Read `docs/spec/{subspec}/plans/{plan}/context.yaml`.
2. Determine target:
   - If user passed `<target>`, use it.
   - Otherwise use `spec.defaultTarget`.
3. If target is not defined in `spec.targets`, stop and show available targets.

### Step 3: Validate manifest and collect normalized file set

Run validator for the selected target:

```bash
python3 .agent-skills/implement/shared/scripts/validate_context.py \
  --context docs/spec/{subspec}/plans/{plan}/context.yaml \
  --docs-root docs \
  --target {target}
```

Expected behavior:

- Exit code `0`: stdout is normalized JSON with `files[]`.
- Non-zero: print stderr and stop.
- Validation scope is limited to `context.yaml` schema/content plus referenced
  markdown path existence and `docs/` boundary checks. Do not add separate
  markdown structure validation.

Map the validator output using
`.agent-skills/implement/shared/references/plan-context-contract.md`:

- Locate exactly one `role == "checklist"` path as `/tdd --file` input.
- If a `role == "test-checklist"` file exists, extract its path as `/tdd --test-checklist`.
- Treat `role == "bdd-plan"` and `role == "bdd-checklist"` as BDD references for `/tdd` gate verification.
- Treat all other validated markdown files as `/tdd --references`.
- Keep file order stable and deduplicate by absolute path.

### Step 3.1: Lock Declarative Boundaries Before Coding

Before handing off to `/tdd`, inspect the loaded spec/plan/checklist set for any
change that affects declarative truth sources or contract boundaries.

If the implementation changes a persisted spec, public schema, or runtime-derived
state model, lock these decisions from the documents first:

1. Which document-owned structure is the source of truth
2. Which fields are runtime-derived, status-only, or otherwise non-authoritative
3. Which values must never be written back into the declarative source
4. Whether the change assumes a clean break, compatibility layer, or migration path

Rules:

- Treat spec/plan/checklist documents as the only truth for these boundaries.
- Do not invent intermediate persistence models, shadow fields, or write-back behavior unless the documents explicitly require them.
- If the loaded documents are ambiguous or contradictory on any of the four points above, stop and confirm with the user before continuing.

### Step 4: Load context and present summary

Read every validated markdown file from Step 3 directly, then summarize:

1. Plan name, selected target, and default target.
2. Loaded file list grouped by role (`plan`, `checklist`, `spec`, ...).
3. Checklist progress (checked/total).
4. Plan Header lifecycle status (`draft`, `active`, ...).

### Step 4.1: Reopened Plan Ownership Gate

If the selected plan was revised in place from a previously completed delivery:

1. Treat the original plan directory as the current delivery owner.
2. If docs were just updated in the current session or the newly added checklist items still have zero progress, continue directly into execution routing unless a concrete blocker prevents it.
3. Do not stop after the summary with "docs updated" as the only result.
4. If execution cannot start, explicitly report that the original plan was revised in place but remains unlaunched; do not present the session as normally closed.

### Step 4.2: Quality Gate Completeness Check

Before branch resolution or `/tdd` handoff, inspect the loaded plan/checklist/context set for the plan's `## 3 质量门禁分类` section.

Required document-level rules:

1. Code plan requires TDD: if the plan introduces front-end, back-end, tooling, migration, codegen, or test helper logic, the quality gate section must name a TDD strategy and checklist items must carry executable test assertions.
2. Feature plan requires BDD: if the plan introduces user-visible UI, API behavior, business workflow, or end-to-end flow, the validated file set must include `bdd-plan` and `bdd-checklist`, and the main checklist must include scenario-ID `BDD-Gate:` items.
3. Internal code plans without BDD must explicitly state why BDD is not applicable and name a substitute verification gate such as contract test, lint, drift check, migration check, or smoke.

If any required classification, BDD file, BDD-Gate, TDD strategy, or substitute gate is missing, stop and route to `/plan-review --fix` before coding. Do not infer missing quality gates during implementation.

### Step 4.3: Frontend / Backend Contract Preflight

Before branch resolution or `/tdd` handoff, check whether the selected target or
validated files touch `frontend/`, `backend/`, `openapi/`, `migrations/`,
`config/ai-*`, `deploy/dev-stack/`, or `test/scenarios/`.

If yes, read and summarize the applicable execution contracts:

1. `docs/development.md` §2 Frontend / Backend Contract Workflow.
2. Every relevant module README, at minimum the README for each touched root
   directory (`frontend/README.md`, `backend/README.md`, `openapi/README.md`,
   `deploy/dev-stack/README.md`, `test/scenarios/README.md`, etc.).
3. For UI-visible work, the relevant `docs/ui-design/` document and
   `ui-design/src/*.jsx` source files.
4. For API/fixture/handler work, `openapi/openapi.yaml`, related
   `openapi/fixtures/<tag>/<operationId>.json`, generated artifacts, and the
   plan's operation matrix.

If a feature/API/cross-layer data plan lacks the operation matrix required by
`docs/development.md` §2.1, stop and route to `/plan-review --fix` before
coding. Do not hand off to `/tdd` while frontend mock progress and backend real
implementation status are ambiguous.

### Step 4.5: Branch Resolution

Insert branch creation and checkout between Step 4 and Step 5.

1. Start Step 4.5 by checking `git status`.
2. Resolve the current branch name and run:

```bash
python3 .agent-skills/implement/shared/scripts/detect_session_branch.py \
  --plan-name {subspec}-{plan} \
  --current-branch "$(git branch --show-current)"
```

If `metadata.branch` exists, pass it as `--branch-stem`.

3. If the current branch already matches the session feature branch, treat the run as retry/resume in place. If the current branch is already the session feature branch, treat the run as retry/resume and continue without creating a new branch.
4. A dirty working tree on that branch is a valid resume state; continue into `/tdd` instead of blocking branch creation.
5. If the working tree has uncommitted changes and the current branch does not match the session feature branch, stop before creating or switching branches.
6. Resolve the base branch in this priority order:
   - `context.yaml` `metadata.baseBranch`
   - `AGENTS.md` project-level Git branch strategy
   - Git default branch auto-detection
7. Before creating a new session feature branch, update the resolved base branch to the latest upstream state with fast-forward-only semantics.
   - Checkout the resolved base branch.
   - Fetch its upstream remote and fast-forward the local base branch (for example `git pull --ff-only` or an equivalent explicit `fetch` + fast-forward).
   - If the base branch has no upstream, cannot be fast-forwarded, or the update fails, stop before creating the feature branch and report the blocker.
   - Do not perform this base-branch update when the current branch already matches the session feature branch and the run is a retry/resume.
8. Resolve the feature branch stem from `metadata.branch` when present; otherwise derive it from `{subspec}-{plan}`.
9. Otherwise, create or switch branches from the updated base branch using the naming convention below.

Branch naming convention:

- `{type}/{subspec}-{plan}-{MMDD}`
- Collision handling: append `-{N}` after the date suffix.
- type inference: `fix/`, `opt/`, `docs/`, otherwise `feat/`.

### Step 5: Execute through sequential `/tdd`

`/implement` does not perform DAG parsing, Wave dispatch, teammate fan-out, or
markdown-format linting. All plans execute through the same sequential `/tdd`
path using checklist order as the source of truth.

Invoke `/tdd` directly:

```text
/tdd --file {checklist-path} --references {ref1},{ref2},... --phase-commit {subspec}/{plan}
```

If a `test-checklist` exists, add `--test-checklist`.

Rules:

- `--file` uses the validated checklist path only.
- `--test-checklist` uses the validated test-checklist path when present.
- `--references` includes all other validated markdown files.
- Never write implementation code against checklist items before `/tdd` takes over.

### Step 6: Completion check

After `/tdd` returns:

1. Read the checklist and confirm all items are checked.
2. Confirm tests were actually run and passed.
3. If the selected plan was revised in place for this session and the newly added checklist items still show zero progress, report it as an unlaunched in-place revision and stop without normal close-out.
4. If blockers remain, report unresolved checklist items and the next required phase.
5. If all items are complete, proceed with the normal `/tdd` completion lifecycle sync.
6. If delivery is complete and verification passed, `/implement` owns the retrospective trigger before final close-out.
