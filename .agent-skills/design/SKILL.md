---
name: design
description: "Design crystallizer: turn a converged design discussion into the minimal sufficient set of spec/plan/test/BDD documents with explicit main-path, edge-condition, failure-path, and regression coverage. IMPORTANT: Invoke this skill automatically when a design discussion has converged and needs to be formalized into project documents. Triggers on /design, or when the user says 'crystallize this design', 'turn this into a spec', 'create spec and plan from this discussion', 'formalize this design', or otherwise indicates that a design conversation should produce actionable documents. Also triggers when the user has a design document or discussion file and wants to generate the matching plan/checklist/test suite."
---

# /design Skill

Turn a converged design discussion into the minimal sufficient set of project documents.
Instead of manually guiding the AI through 6-8 file creation steps, one command produces
exactly the documents the current need requires — no more, no less.

## Usage

```
/design {subject}                    # Extract from current conversation
/design {subject} --from {path}      # Extract from a docs/discuss/ file
/design -h                           # Show help
```

## Parameters

| Parameter | Required | Description |
|-----------|----------|-------------|
| `{subject}` | Yes | Topic identifier used for file naming (e.g., `sealed-secrets`, `wiki-sync`) |
| `--from {path}` | No | Path to a discussion/design file to extract from instead of conversation context |
| `-h` / `--help` | No | Show usage and stop |

## Workflow

### Step 1: Extract Design Brief

Gather structured information from the conversation context (or `--from` file). Map each
piece of discussion content to the Brief fields below.

| Brief Field | Source | Maps To |
|-------------|--------|---------|
| `subject` / `title` | Parameter + discussion title | File naming |
| `goals` | Stated design objectives | spec 2 |
| `decisions` | Confirmed choices (ID / decision / conclusion / rationale) | spec 3 |
| `acceptance_criteria` | criterion-id / Given / When / Then / Phase | spec Acceptance Criteria |
| `requirements` | Functional needs and constraints | spec 4-7 |
| `non_goals` | Explicitly excluded scope | spec Non-goals |
| `interfaces` | API / data structure definitions | spec 5 |
| `source_spec` | Existing spec path to reuse when no new spec is generated | context.yaml `spec` + summary |
| `components` | Involved components mapped to main code areas | plan phases + implementation scope |
| `coverage_matrix` | Primary flows, alternate flows, failure paths, edge conditions, regressions, and out-of-scope boundaries | plan/checklist/test-plan/bdd coverage |
| `test_surface` | Unit / contract / integration / scenario / lint / drift / smoke verification needed for each coverage row | test-plan + checklist + BDD-Gate |
| `risks` | Risks and mitigations | plan Risks |
| `ui_design_contract` | Relevant `docs/ui-design/` documents, active spec constraints, and formal `frontend/` implementation boundaries when user-visible UI is in scope | UI behavior, responsive, accessibility, and browser-verification rows in coverage matrix |
| `open_questions` | Unresolved items | spec Open Questions |
| `inferred_outputs` | Recommended document set (see Step 3) | Output scope |
| `bdd_scenarios` | User-observable behaviors, verification layers, domain Behavior IDs, and any justified real E2E ID reservations | bdd-plan + bdd-checklist + checklist BDD-Gate |
| `bdd_strategy` | Whether BDD phase gates are needed + rationale | BDD document generation |

The Brief is a transient confirmation artifact — it is never persisted as a file.

### Step 2: User Confirmation

Present the Brief to the user in a structured format:

1. **Summary table** with subject, goals, key decisions
2. **Recommended output scope** — which documents will be generated and why (see Step 3 logic)
3. **BDD strategy** — whether BDD phase gates are recommended, which user behaviors they describe, and whether each behavior is proved by a domain behavior test or a real API/UI E2E flow
4. **Coverage matrix** — primary flows, important edge conditions, failure paths, regression/non-current-negative checks, and which plan/test/BDD artifact will cover each one
5. **Acceptance criteria summary** (if any criteria were extracted) + BDD behavior matrix (when BDD is active)

Wait for explicit user confirmation before proceeding. If the user cancels, stop without
generating any files.

### Step 2.5: Branch Guard Before Document Mutation

`/design` may extract, summarize, and ask for confirmation on any branch, but it must not
create or revise `docs/` files on the default parent branch. Run this guard after the user
confirms the Brief and before Step 4, before invoking `/create-doc`, before updating any
INDEX, and before writing spec / plan / checklist / test / BDD / context files:

1. Check the current branch and worktree with `git status --short --branch`.
2. If the current branch already matches the session feature branch, treat the run as a
   retry/resume and continue.
3. If the current branch is the default parent branch and the worktree is clean, update
   the parent branch with fast-forward-only semantics, then create or switch to a feature
   branch before editing files. Use the repository branch prefix convention, for example
   `design/{subject}` or another concise `spec-design/` branch name tied to the design
   subject. The prefix must describe the work type or domain; never create new
   `codex/`, `claude/`, `gemini/`, `agent/`, or other tool-name branches.
4. If the fast-forward-only parent update fails, stop before file edits and report the
   blocker. Do not generate documents from a stale parent branch.
5. If dirty changes already came from the current session while still on the default
   parent branch, create the feature branch immediately while preserving those changes,
   report the recovery, and continue only after the branch switch.
6. If the default parent branch is dirty for unclear or user-owned reasons, stop and ask
   the user before creating or switching branches.
7. If the current session is on a freshly created tool-name branch that has not been
   pushed or shared, rename it to the semantic repository prefix before editing files.
   If it may already be externally referenced, stop and ask the user.
8. If a non-parent branch has unrelated dirty changes and does not match the session
   feature branch, stop and ask the user before mutating anything.

Never invoke `/create-doc`, create spec / plan directories, revise completed owner docs
back to `active`, or update INDEX files on the default parent branch. For pure proposal
or backlog guidance with no file edits, no branch switch is required.

### Step 3: Infer Minimal Sufficient Output Scope

Based on the Brief content, determine the smallest document set that fully serves the need.
Do not expose `--scope` or `--skip-bdd` flags — the scope is an LLM judgment call, not a
user toggle. This prevents users from accidentally under-scoping or over-generating.

**Decision logic:**

1. **Needs design landing?** (new architecture, trade-offs, interface contracts, schema changes)
   - Yes -> generate spec
   - No (pure implementation of existing design) -> skip spec, but require an explicit reusable spec path from current context or the user and store it as `source_spec`

2. **Needs execution plan?** (multi-step implementation, checklist tracking, team coordination)
   - Yes -> generate `docs/spec/{subspec}/plans/{NNN-plan}/plan.md` + `checklist.md` + `context.yaml`
   - No (spec-only crystallization) -> stop after spec

3. **Needs test plan?** (unit testable components, complex logic, regression risk)
   - Code plan requires TDD: any plan that introduces front-end, back-end, tooling, migration, codegen, or test helper logic must carry a TDD strategy and executable test assertions for each checklist item.
   - If the tests need independent phase mapping -> generate a dedicated `{NNN-unit-test}` plan directory with `plan.md` + `checklist.md`, or generate `test-plan.md` / `test-checklist.md` in the implementation plan directory when that is the tighter fit.
   - If the plan is docs-only -> skip test plan and record `TDD 策略: 不适用：docs-only`.

4. **Needs BDD phase gates?** (behavior phases that can be independently deployed and verified,
   acceptance criteria with Given/When/Then, end-to-end verification requirements)
   - Feature plan requires BDD: any plan that introduces user-visible UI, API behavior, or business workflow must generate `bdd-plan.md` and `bdd-checklist.md`, add `BDD-Gate:` items keyed by Behavior ID or justified real E2E ID, and add `bddPlan` and `bddChecklist` to `context.yaml`.
   - BDD describes user-observable behavior; it does not imply an E2E directory. Use a domain Behavior ID and code-level domain behavior test when that is the smallest sufficient proof.
   - Allocate an `E2E.*` ID only after confirming that the verification drives an already running frontend/backend through real HTTP API calls or browser UI interactions. Read `test/scenarios/README.md` plus the E2E `README.md` / `INDEX.md` only for that real-E2E branch.
   - Pure configuration defaults, internal contracts, tooling, migration, codegen, lint, fixtures, and build orchestration are `BDD-N/A` unless they introduce a distinct user-observable flow. Record the reason and substitute gate in the main plan, generate no BDD files, and leave `bddPlan` / `bddChecklist` out of `context.yaml`.

**Output the reasoning** for each decision so the user can challenge it in Step 2.

### Step 3.5: Build the Coverage Matrix

Before generating plan/checklist/test/BDD artifacts, build an explicit coverage matrix from the Brief.
This matrix is the guardrail that prevents plan documents from covering only the happy path.

For every behavior, invariant, interface, data transition, or risk in scope, classify rows using the smallest useful set of these categories:

- **Primary path**: the expected user or system flow that proves the feature works.
- **Alternate path**: legitimate variants such as unauthenticated/authenticated, role or permission differences, config variants, language/theme/mode differences, optional inputs, provider/profile choices, or feature-disabled behavior.
- **Failure / recovery path**: validation errors, missing or malformed input, downstream unavailable, timeout/retry, partial state, transaction failure, stale data, conflict, cancellation, or cleanup after failure.
- **Boundary condition**: empty/min/max values, duplicate records, ordering, pagination, concurrency/idempotency, rerun behavior, migration on non-empty data, unknown enum/route/provider/config, and data retention/deletion edges.
- **Cross-layer contract**: API/schema/OpenAPI/shared type/codegen parity, generated client behavior, fixture/mock parity, event/job contract, database constraint, runtime config, and scenario data contract.
- **Privacy / security / observability**: auth boundary, sensitive data redaction, audit/log/metric expectations, OWASP-relevant input handling, and no secret/token persistence.
- **UX quality**: loading/empty/error states, accessibility semantics, localization fallback, display preference behavior, responsive layout, and copy visible to the user.
- **Regression / non-current-negative**: non-current route/module/tag/table/event/config/model/provider/feature flag terminology must not reappear when the current design rejects it.

For UI plans whose truth source is `ui-design/`, add explicit source-level parity rows instead of relying only on visual similarity:

- **UI source structure parity**: DOM composition, component nesting, control type, menu/popover hierarchy, icons, labels, aria state, keyboard/close behavior, and primary interaction paths must map from `ui-design/src/*.jsx`, `ui-design/src/app.jsx`, and `ui-design/src/primitives.jsx` to concrete frontend components/tests.
- **UI visual geometry parity**: computed style, spacing, typography, colors, responsive layout, bounding boxes, and screenshot/baseline checks must be verified separately from source structure parity. A passing screenshot or bounding-box gate is not sufficient evidence for source-level replication.
- **UI stale-contract negative**: old positive UI contracts such as non-current `data-testid`s, old route labels, old dropdown/select controls, old screen names, and old prototype shorthand must be searched across spec/plan/README/scenario/test/runtime files when they conflict with the current truth source.

Each coverage row must map to one or more concrete artifacts:

| Coverage row field | Requirement |
|--------------------|-------------|
| `source` | Spec requirement, decision, acceptance criterion, risk, or explicit inference from the Brief |
| `category` | One of the categories above |
| `plan_phase` | The phase/checklist item that implements or verifies it |
| `verification` | Unit test, contract test, lint/drift check, migration check, smoke, domain Behavior ID, or real E2E ID |
| `negative_scope` | Non-current or intentionally excluded behavior to search for, when relevant |
| `ui_source_anchor` | Required for UI parity rows; cite the concrete `ui-design/src/*.jsx` function/component/constant or docs/ui-design section that owns the target shape |

Do not create synthetic edge cases for irrelevant categories, but do not silently omit a category just because the user did not name it. If a high-risk category is not applicable, write a short `N/A` rationale in the plan quality gate or test plan.

### Step 4: Generate Documents

Generate only the documents determined in Step 3. For each document type, follow the
corresponding template from the project documentation — reference the template, do not
inline-copy it (prevents template drift).

Before creating or modifying any file under `docs/`, invoke `/create-doc` and follow the
target directory README/INDEX rules. `/design` decides the minimal sufficient scope; `/create-doc`
owns the repository's document creation mechanics.

#### 4.1 Spec Document

- Path: `docs/spec/{subspec}/spec.md`
- Supporting history path: `docs/spec/{subspec}/history.md`
- Template source: `docs/spec/TEMPLATES.md`
- Naming: follow `docs/spec/README.md` spec-centric v2 naming rules
- Include Acceptance Criteria section (8) when criteria are present in the Brief
- Keep Acceptance Criteria descriptive; do not use `AC-*` as BDD-Gate machine references in new documents
- Use the spec's section numbering: 1-Overview through 10-Related Documents
- When Step 3 chose spec reuse instead of new spec generation, do not create a new spec directory; keep the reused path in `source_spec`

#### 4.2 Implementation Plan + Checklist

- Plan path: `docs/spec/{subspec}/plans/{NNN-plan}/plan.md`
- Checklist path: `docs/spec/{subspec}/plans/{NNN-plan}/checklist.md`
- Template source: `docs/spec/TEMPLATES.md` plan/checklist sections
- Ensure `docs/spec/{subspec}/plans/INDEX.md` exists using `/init-docs` `subspec-plans` scaffold before writing the first plan for a subject; do not create local `plans/README.md` or `plans/TEMPLATES.md`
- When BDD is needed: add `BDD-Gate:` items at the end of each behavior phase per `docs/spec/TEMPLATES.md`, using Behavior IDs or justified real E2E IDs from `bdd-plan.md`; track behavior evidence and execution in `bdd-checklist.md`
- Phase design must follow the phase closability principle from spec 4.4:
  each behavior phase is a vertical behavior slice that can be independently deployed and verified
- Every implementation plan must include `## 3 质量门禁分类` with Plan 类型, TDD 策略, BDD 策略, and 替代验证 gate.
- Every non-docs checklist item must name its verification source: a unit/contract/integration test, lint/drift gate, migration check, smoke, or BDD-Gate that covers the row in the coverage matrix.
- UI implementation checklist items that migrate from `ui-design/` must include both source-structure parity and visual-geometry parity verification. The checklist must name the source anchors, target components, and tests that fail on control-type or interaction-shape drift (for example select/dropdown vs menu/toggle).
- Each implementation phase must cover its primary path and any directly coupled failure, cleanup, idempotency, privacy/security, or non-current-negative checks before the phase can be marked closable.
- If an edge condition is deferred, the plan must say which later phase owns it and why the current phase remains safely deployable without it.

#### 4.3 Unit Test Plan + Checklist

- Default path: `docs/spec/{subspec}/plans/{NNN-unit-test}/plan.md`
- Checklist path: `docs/spec/{subspec}/plans/{NNN-unit-test}/checklist.md`
- Use phase-number section headings that directly mirror the implementation checklist by default.
- Write unit-test completion items against the planned test set itself, for example `Phase N 本计划定义的单元测试项全部通过`.
- Do not generate hard coverage-percentage gates in acceptance criteria or checklist items.
- If coverage is mentioned at all, keep it as observational background rather than a completion, commit, or phase-exit condition.
- Test plans must include a coverage matrix that maps primary, alternate, failure/recovery, boundary, cross-layer contract, privacy/security/observability, UX quality, and regression/non-current-negative rows to concrete test files or commands.
- Unit/contract test checklists must include negative and boundary assertions for meaningful risks, not only success assertions. Prefer deterministic checks for malformed input, empty data, unknown identifiers, duplicate/conflict handling, rerun/idempotency, config fallback, generated contract drift, and non-current terminology reintroduction when those risks are in scope.
- UI-related test plans must consider loading, empty, error, auth, localization fallback, display preference, accessibility, responsive-state risks, and source-level UI parity; include only the rows that matter for the subject and mark high-risk exclusions explicitly.
- For `ui-design/` parity, test plans must split assertions into source-structure tests (DOM shape, control type, menu/popover hierarchy, icons, labels, aria state, primary interactions), visual-geometry tests (computed style, bounding boxes, responsive layout, screenshots), and stale-contract negative searches. Do not treat pixel/screenshot parity as a substitute for DOM/interaction parity.
- Backend/tooling/migration test plans must consider validation errors, persistence failures, transaction/concurrency behavior, retry/idempotency, non-empty data, rerun safety, generated artifacts, logs/metrics/audit redaction, and drift gates.

#### 4.4 BDD Test Plan + Checklist

- BDD files live inside the relevant plan directory:
  `docs/spec/{subspec}/plans/{NNN-plan}/bdd-plan.md` and
  `docs/spec/{subspec}/plans/{NNN-plan}/bdd-checklist.md`
- Only generated when Step 3 inferred a real user-observable behavior; `BDD-N/A` plans generate neither file
- `bdd-plan.md` contains detailed Given/When/Then behaviors grouped by Phase and does not contain execution progress checkboxes
- Give each behavior a stable domain Behavior ID such as `BDD.AUTH.001`. A Behavior ID may be verified by a code-level domain behavior test and does not reserve or require an E2E scenario.
- Allocate an ID such as `E2E.P0.001` only when the verification drives the running product through real HTTP/UI. Before allocating that ID, read `test/scenarios/README.md` and the E2E `README.md` / `INDEX.md`.
- `bdd-checklist.md` records the chosen verification entrypoint, executable behavior assertions, result, and evidence. Only a real E2E entry includes scenario-directory, data-isolation, setup/trigger/verify/cleanup, and environment tasks.
- BDD behavior selection must cover the primary user journey plus the highest-risk alternate or failure/recovery journey for each independently deployable behavior phase. Do not push unit-level edge cases into BDD, but do include user-visible auth, permission, empty/error, recovery, and non-current-negative flows when they define product correctness.
- `bdd-plan.md` must include a behavior matrix that labels each behavior as primary, alternate, failure/recovery, or regression/non-current-negative, maps it to the plan phase and checklist BDD-Gate, and names `domain behavior test` or `real API/UI E2E` as its evidence layer.
- A domain behavior test stays in its code owner and is executed through the normal code-test gates; do not create a `test/scenarios/e2e/` shell wrapper around Go, Vitest, npm test, pytest, lint, source-contract, fixture, or build commands.
- A real E2E scenario must exercise the running frontend/backend without mock transport or request interception replacing the backend.

#### 4.5 context.yaml

- Path: `docs/spec/{subspec}/plans/{NNN-plan}/context.yaml`
- Generated whenever implementation plan is generated
- Template reference: `docs/spec/TEMPLATES.md` context section
- Generator reference: `.agent-skills/implement/shared/scripts/generate_context_yaml.py`
- `spec` must point to either `../../spec.md` or `source_spec`; if neither exists, stop and ask for clarification before writing the plan set
- Include `bddPlan` and `bddChecklist` fields only when BDD artifacts are generated
- `metadata` must contain only `name`; derive subject and plan order from the path
- Do not add top-level or target-level `discovery`, target `references`, branch hints, Spec versions, commands, or custom fields

### Step 5: Update INDEX Files

Update the relevant INDEX files:

- `docs/spec/INDEX.md` when a spec subject is generated or revised
- `docs/spec/{subspec}/plans/INDEX.md` when a plan is generated or revised

Plans remain inside the subject directory and are indexed only by that subject's
local `plans/INDEX.md`.

### Step 6: Verify and Summarize

Run validation and present a summary:

1. **context.yaml validation** (when generated):
   ```bash
   python3 .agent-skills/implement/shared/scripts/validate_context.py \
     --context docs/spec/{subspec}/plans/{NNN-plan}/context.yaml \
     --docs-root docs \
     --target {default-target}
   ```

2. **BDD reference integrity** (when BDD is active):
   - Every Behavior ID or real E2E ID in the checklist has a corresponding entry in `bdd-plan.md`
   - Every ID in `bdd-plan.md` has a corresponding evidence/execution section in `bdd-checklist.md`
   - Every real E2E ID follows `test/scenarios/e2e/README.md` / `INDEX.md`; domain Behavior IDs do not create scenario directories
   - Every E2E entry proves real running-product HTTP/UI interaction and does not wrap code-level test, lint, fixture, or build commands
   - Report any gaps as errors

3. **Coverage integrity**:
   - Every non-docs checklist item names a concrete verification source
   - Every coverage matrix row maps to a plan phase and at least one verification artifact
   - Every UI source parity row maps to a `ui_source_anchor`, a target component/file, and at least one source-structure test plus one visual-geometry or explicit N/A rationale
   - Every BDD-Gate item maps to a behavior labeled in the BDD behavior matrix
   - Any high-risk category marked `N/A` includes a rationale
   - Regression/non-current-negative rows include explicit search targets when non-current terminology, routes, modules, schemas, events, configs, or model/provider assumptions are part of the risk

4. **Output summary**:
   - Files generated (with paths)
   - Reused spec input (when Step 3 skipped new spec generation)
   - Output scope decision rationale
   - BDD strategy rationale
   - Coverage matrix summary, including any high-risk `N/A` rationale
   - INDEX entries added
   - Any warnings or items needing attention

## Prohibited Actions

- Generating documents beyond what the Brief justifies (over-generation wastes effort and
  creates maintenance burden)
- Skipping spec generation when the Brief contains unrecorded trade-offs or architecture
  decisions (these need a persistent home)
- Inlining template content from `TEMPLATES.md` files (always reference the template source; inlined
  copies drift from the canonical template over time)
- Generating implementation plan/context artifacts without a concrete spec path (either a newly
  generated spec file or an explicit `source_spec` reuse path)
- Generating `bddPlan` / `bddChecklist` in context.yaml without generating the matching `bdd-plan.md` / `bdd-checklist.md` files (or vice versa)
- Generating BDD files or retaining `bddPlan` / `bddChecklist` fields for a pure internal/config/tooling `BDD-N/A` plan
- Generating BDD-Gate checklist items with `AC-*` references for new documents
- Allocating an `E2E.*` ID or creating a scenario directory for a code-level Go/Vitest/pytest/lint/fixture/build gate instead of a real running-product HTTP/UI flow
- Inventing real E2E IDs without first consulting `test/scenarios/e2e/README.md` and `INDEX.md`
- Creating or updating files under `docs/` without invoking `/create-doc`
- Invoking `/create-doc`, creating spec / plan directories, or updating INDEX files before
  the Step 2.5 branch guard succeeds
- Generating test plan acceptance criteria or checklist items that use raw code coverage percentages as hard gates
- Generating plan/checklist/test-plan/BDD artifacts that only verify the primary path while omitting obvious failure, boundary, security/privacy, UX, contract, or regression/non-current-negative risks
- Marking edge coverage as "later" without assigning a concrete owner phase and explaining why the current phase is still independently deployable
- Proceeding past Step 2 without explicit user confirmation of the Brief
- Persisting the Brief as a file (it is a transient conversation artifact)
