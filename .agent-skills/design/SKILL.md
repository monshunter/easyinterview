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
| `components` | Involved components mapped to main code areas | plan phases + context discovery |
| `coverage_matrix` | Primary flows, alternate flows, failure paths, edge conditions, regressions, and out-of-scope boundaries | plan/checklist/test-plan/bdd coverage |
| `test_surface` | Unit / contract / integration / scenario / lint / drift / smoke verification needed for each coverage row | test-plan + checklist + BDD-Gate |
| `risks` | Risks and mitigations | plan Risks |
| `open_questions` | Unresolved items | spec Open Questions |
| `inferred_outputs` | Recommended document set (see Step 3) | Output scope |
| `bdd_scenarios` | Target test layers, numbering rules, and scenario IDs/reservations | bdd-plan + bdd-checklist + checklist BDD-Gate |
| `bdd_strategy` | Whether BDD phase gates are needed + rationale | BDD document generation |

The Brief is a transient confirmation artifact — it is never persisted as a file.

### Step 2: User Confirmation

Present the Brief to the user in a structured format:

1. **Summary table** with subject, goals, key decisions
2. **Recommended output scope** — which documents will be generated and why (see Step 3 logic)
3. **BDD strategy** — whether BDD phase gates are recommended and the reasoning
4. **Coverage matrix** — primary flows, important edge conditions, failure paths, regression/legacy-negative checks, and which plan/test/BDD artifact will cover each one
5. **Acceptance criteria summary** (if any criteria were extracted) + BDD scenario matrix (when BDD is active)

Wait for explicit user confirmation before proceeding. If the user cancels, stop without
generating any files.

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
   - Feature plan requires BDD: any plan that introduces user-visible UI, API behavior, business workflow, or end-to-end flow must read `test/scenarios/README.md` plus the relevant layer `README.md` / `INDEX.md`, generate bdd-plan.md and bdd-checklist.md with scenario IDs that follow those conventions, add BDD-Gate items keyed by scenario IDs, and add `bddPlan` and `bddChecklist` to context.yaml.
   - BDD is not a discretionary optional artifact for user behavior. Skip BDD only for docs-only or internal contract/tooling/migration/codegen plans that do not create a user behavior flow; record the reason and the substitute verification gate.

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
- **Regression / legacy-negative**: retired route/module/tag/table/event/config/model/provider/feature flag terminology must not reappear when the current design rejects it.

Each coverage row must map to one or more concrete artifacts:

| Coverage row field | Requirement |
|--------------------|-------------|
| `source` | Spec requirement, decision, acceptance criterion, risk, or explicit inference from the Brief |
| `category` | One of the categories above |
| `plan_phase` | The phase/checklist item that implements or verifies it |
| `verification` | Unit test, contract test, lint/drift check, migration check, smoke, or BDD scenario ID |
| `negative_scope` | Deprecated or intentionally excluded behavior to search for, when relevant |

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
- When BDD is needed: add `BDD-Gate:` items at the end of each behavior phase per `docs/spec/TEMPLATES.md`, using scenario IDs from `bdd-plan.md`; track scenario asset readiness and execution in `bdd-checklist.md`
- Phase design must follow the phase closability principle from spec 4.4:
  each behavior phase is a vertical behavior slice that can be independently deployed and verified
- Every implementation plan must include `## 3 质量门禁分类` with Plan 类型, TDD 策略, BDD 策略, and 替代验证 gate.
- Every non-docs checklist item must name its verification source: a unit/contract/integration test, lint/drift gate, migration check, smoke, or BDD-Gate that covers the row in the coverage matrix.
- Each implementation phase must cover its primary path and any directly coupled failure, cleanup, idempotency, privacy/security, or legacy-negative checks before the phase can be marked closable.
- If an edge condition is deferred, the plan must say which later phase owns it and why the current phase remains safely deployable without it.

#### 4.3 Unit Test Plan + Checklist

- Default path: `docs/spec/{subspec}/plans/{NNN-unit-test}/plan.md`
- Checklist path: `docs/spec/{subspec}/plans/{NNN-unit-test}/checklist.md`
- Use phase-number section headings that directly mirror the implementation checklist by default.
- Write unit-test completion items against the planned test set itself, for example `Phase N 本计划定义的单元测试项全部通过`.
- Do not generate hard coverage-percentage gates in acceptance criteria or checklist items.
- If coverage is mentioned at all, keep it as observational background rather than a completion, commit, or phase-exit condition.
- Test plans must include a coverage matrix that maps primary, alternate, failure/recovery, boundary, cross-layer contract, privacy/security/observability, UX quality, and regression/legacy-negative rows to concrete test files or commands.
- Unit/contract test checklists must include negative and boundary assertions for meaningful risks, not only success assertions. Prefer deterministic checks for malformed input, empty data, unknown identifiers, duplicate/conflict handling, rerun/idempotency, config fallback, generated contract drift, and deprecated terminology reintroduction when those risks are in scope.
- UI-related test plans must consider loading, empty, error, auth, localization fallback, display preference, accessibility, and responsive-state risks; include only the rows that matter for the subject and mark high-risk exclusions explicitly.
- Backend/tooling/migration test plans must consider validation errors, persistence failures, transaction/concurrency behavior, retry/idempotency, non-empty data, rerun safety, generated artifacts, logs/metrics/audit redaction, and drift gates.

#### 4.4 BDD Test Plan + Checklist

- BDD files live inside the relevant plan directory:
  `docs/spec/{subspec}/plans/{NNN-plan}/bdd-plan.md` and
  `docs/spec/{subspec}/plans/{NNN-plan}/bdd-checklist.md`
- Only generated when Step 3 inferred BDD is needed
- Before allocating IDs, read `test/scenarios/README.md` and the target suite `README.md` / `INDEX.md`
- `bdd-plan.md` contains detailed Given/When/Then scenarios grouped by Phase and does not contain execution progress checkboxes
- `bdd-checklist.md` contains scenario asset and execution tasks for each scenario ID: create scenario directory, prepare data, implement setup/trigger/verify/cleanup, execute verification, and record evidence
- Each scenario uses a behavior-oriented scenario ID such as `E2E.P0.001` or `E2E.P1.003`; if needed, map back to spec acceptance criteria inside `bdd-plan.md`, not in `BDD-Gate` items
- BDD scenario selection must cover the primary user journey plus the highest-risk alternate or failure/recovery journey for each independently deployable behavior phase. Do not push unit-level edge cases into BDD, but do include user-visible auth, permission, empty/error, recovery, and legacy-negative flows when they define product correctness.
- `bdd-plan.md` must include a scenario matrix that labels each scenario as primary, alternate, failure/recovery, or regression/legacy-negative, and maps it to the plan phase and checklist BDD-Gate.
- `bdd-checklist.md` must make setup, data isolation, cleanup, pollution recovery, execution command, and evidence capture explicit for every scenario.

#### 4.5 context.yaml

- Path: `docs/spec/{subspec}/plans/{NNN-plan}/context.yaml`
- Generated whenever implementation plan is generated
- Template reference: `docs/spec/TEMPLATES.md` context section
- Generator reference: `.agent-skills/implement/shared/scripts/generate_context_yaml.py`
- `spec` must point to either `../../spec.md` or `source_spec`; if neither exists, stop and ask for clarification before writing the plan set
- Include `bddPlan` and `bddChecklist` fields only when BDD artifacts are generated
- Fill `spec.discovery` with aliases and keywords from the Brief
- Fill `spec.targets.<target>.discovery.packages` from the component list
- Add `uiRoutes` / `apiNames` only when they are stable retrieval identifiers
- Do not write `commands` into `context.yaml`; runtime commands belong to repo README / scenario docs

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
   - Every scenario ID in the checklist has a corresponding entry in `bdd-plan.md`
   - Every scenario ID in `bdd-plan.md` has a corresponding asset/execution section in `bdd-checklist.md`
   - Every generated scenario ID follows the relevant layer `README.md` / `INDEX.md` numbering convention
   - Report any gaps as errors

3. **Coverage integrity**:
   - Every non-docs checklist item names a concrete verification source
   - Every coverage matrix row maps to a plan phase and at least one verification artifact
   - Every BDD-Gate item maps to a scenario labeled in the BDD scenario matrix
   - Any high-risk category marked `N/A` includes a rationale
   - Regression/legacy-negative rows include explicit search targets when retired terminology, routes, modules, schemas, events, configs, or model/provider assumptions are part of the risk

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
- Generating BDD-Gate checklist items with `AC-*` references for new documents
- Inventing BDD scenario IDs without first consulting the relevant `test/scenarios/<layer>/README.md` and `INDEX.md` when such conventions exist
- Creating or updating files under `docs/` without invoking `/create-doc`
- Generating test plan acceptance criteria or checklist items that use raw code coverage percentages as hard gates
- Generating plan/checklist/test-plan/BDD artifacts that only verify the primary path while omitting obvious failure, boundary, security/privacy, UX, contract, or regression/legacy-negative risks
- Marking edge coverage as "later" without assigning a concrete owner phase and explaining why the current phase is still independently deployable
- Proceeding past Step 2 without explicit user confirmation of the Brief
- Persisting the Brief as a file (it is a transient conversation artifact)
