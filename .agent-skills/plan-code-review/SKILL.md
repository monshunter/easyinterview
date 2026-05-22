---
name: plan-code-review
description: "Review or fix code against spec/plan/checklist context. Use when the user wants L2 code review or remediation for already-implemented checklist phases, especially after product/UI spec changes, historical implementation drift, or requests to ignore old checklist/PASS state. Reuses implement-owned shared context validator, then performs artifact-level semantic review against validated markdown, current truth sources, generated artifacts, tests, fixtures, scripts, coverage-matrix expectations, and negative legacy-scope searches. Supports /plan-code-review SUBSPEC/PLAN [target] [--base-rev REV] [--fix]."
---

# Plan Code Review Skill

L2 code review for `code ↔ spec/plan/checklist`. This skill checks completed phases,
reports drift, and, when the user confirms fixes, routes remediation through
`/tdd --section`. It is not an end-to-end delivery entry point.

## Usage

- `/plan-code-review <subspec>/<plan>` - Review the default target of the named spec-centric plan
- `/plan-code-review <subspec>/<plan> <target>` - Review a specific target
- `/plan-code-review <subspec>/<plan> [target] --base-rev <git-rev>` - Include git diff context
- `/plan-code-review <subspec>/<plan> [target] --fix` - Review, preview, confirm, and fix via `/tdd --section`
- `/plan-code-review <subspec>/<plan> [target] --fix --base-rev <git-rev>` - Same with git diff context
- `/plan-code-review -h` - Show help only
- `/plan-code-review -h -v` - Show verbose help (including workflow)

Flag rules:

- Plan name is mandatory
- `--fix` implies a review pass first
- Review is advisory; fix remains preview-only until user confirmation

## Shared Resources

Use the implement-owned shared resources:

- `.agent-skills/implement/shared/scripts/validate_context.py`
- `.agent-skills/implement/shared/references/plan-context-contract.md`

Reviewer rule:

- Trust the post-`doc-init` templates for markdown structure unless the user is
  explicitly fixing template drift.
- Do not add parser-only gates or markdown-format checkers before the semantic
  code review.
- New plan docs are sequential-only by default.
- Checked checklist items define the primary implementation scope.
- Task `**文件**:` declarations are optional hints, not required scope contracts.
- Historical `completed` status, checked checklist items, previous PASS reports,
  and small git diffs are never sufficient evidence. Treat them as leads to
  verify against current artifacts.
- When product or UI scope may have changed, derive current semantic invariants
  from active spec, `docs/ui-design/`, and `ui-design/`, then audit the code and
  generated artifacts against those invariants.

## Workflow

### Step 0: Handle help flags

- `-h`/`--help`: show skill name, description, usage, then stop.
- `-h -v`/`--help --verbose`: show usage + full workflow, then stop.

### Step 1: Resolve plan and target

1. Require an explicit `subspec/plan` name.
2. Read `docs/spec/{subspec}/plans/{plan}/context.yaml`.
3. Determine target from the explicit argument or `spec.defaultTarget`.
4. If target is missing, stop and show available targets.

### Step 2: Validate manifest and collect normalized file set

Run:

```bash
python3 .agent-skills/implement/shared/scripts/validate_context.py \
  --context docs/spec/{subspec}/plans/{plan}/context.yaml \
  --docs-root docs \
  --target {target}
```

Use the shared plan-context contract for role mapping and common errors.

Validation scope is limited to manifest shape/content and referenced markdown
paths. After validation, read the returned markdown files directly.

### Step 2.5: Frontend / Backend Contract Preflight

Before deriving findings, check whether the target, validated files, diff scope,
or discovered artifacts involve `frontend/`, `backend/`, `openapi/`,
`migrations/`, `config/ai-*`, `deploy/dev-stack/`, or `test/scenarios/`.

If yes, read the current execution contracts and include them in `Deep Evidence`:

1. `docs/development.md` §2 Frontend / Backend Contract Workflow.
2. Relevant module README files for the touched roots.
3. UI-visible targets: `docs/ui-design/` plus relevant `ui-design/src/*.jsx`,
   `ui-design/src/primitives.jsx`, or `ui-design/src/app.jsx` source.
4. API/fixture/handler targets: `openapi/openapi.yaml`, related fixtures,
   generated artifacts, and the operation matrix.
5. Local integration/scenario targets: `deploy/dev-stack/README.md` and
   `test/scenarios/README.md`, with Docker Compose vs Kind boundaries checked
   against current docs rather than historical reports.

If the reviewed plan lacks the operation matrix required by
`docs/development.md` §2.1, record a blocking finding. In `--fix` mode, map the
finding to a plan/document repair first; do not treat code or fixture presence as
evidence of frontend/backend closure.

### Step 3: Determine code review scope

1. Main scope: checklist phases with at least one `[x]` item.
2. Gather concrete code scope from the strongest available sources, in this order:
   - `--base-rev` git diff filtered to files relevant to the current target
   - target-level discovery in `context.yaml` (`packages`, `uiRoutes`, `apiNames`, `commands`)
   - plan task declarations such as `**文件**:`
3. Expand the scope from implementation files to artifact files that prove the
   behavior: generated output, fixtures, baselines, migrations/DDL, config,
   scripts, Make targets, README docs, smoke scripts, and tests.
4. Missing phase file declarations do not invalidate the review by themselves.
5. If no concrete file set can be derived, fall back to target-level advisory review.
6. `--fix` requires a concrete checklist-section mapping; target-level-only findings stay preview-only until the user confirms the section.

Code scope sources:

- **Source A (git diff)**: if `--base-rev` is provided, `git diff --name-only {base}..HEAD`
- **Source B (context discovery)**: target `packages` plus other target discovery hints from validated `context.yaml`
- **Source C (plan declarations)**: `**文件**:` exact paths when present

### Step 4: Execute L2 semantic review directly

For each in-scope phase:

1. Read code files for the phase.
2. Read relevant spec sections.
3. Read plan task descriptions.
4. Read checklist items and completion status.
5. Merge git diff and target discovery context when present.
6. Build an artifact map that connects each completed checklist item to concrete
   source code, generated output, fixtures, baselines, DDL/config, scripts,
   README entries, and tests.
   - For lifecycle / runtime capabilities, reverse-audit the production
     entrypoint instead of stopping at package-level behavior. If a spec, plan,
     checklist, or discovered artifact declares a worker, dispatcher, outbox
     loop, scheduler, bootstrap hook, shutdown/drain path, background runner, or
     runtime kernel, inspect the real startup path such as `cmd/api`, `main`,
     server boot, worker launch, Docker Compose service, or Kind deployment
     entrypoint. Verify a production-wiring test, smoke, or scenario proves the
     startup path constructs, attaches/registers, starts, and shuts down the
     runtime capability. Treat internal package tests alone as insufficient
     evidence for production lifecycle closure; in `--fix` mode, add the
     missing production-wiring test or reopen the mapped checklist item before
     accepting the phase as complete.
7. Reconstruct the expected coverage matrix from the validated spec/plan/checklist,
   test-plan/test-checklist, bdd-plan/bdd-checklist, quality-gate classification,
   non-goals, risks, and active product/UI truth sources. For each completed
   checklist item, verify the artifact map proves the relevant primary,
   alternate, failure/recovery, boundary, cross-layer contract, privacy/security/
   observability, UX, and regression/legacy-negative rows.
8. Run negative legacy-scope searches relevant to the target. At minimum, cover
   stale route/tag/schema/table/event/job/config flag names, vendor/model
   assumptions, feature-key routing assumptions, and product modules that the
   current spec/UI has dropped.
   - For UI parity targets, include stale positive UI contract searches across
     `docs/`, package READMEs, scenario assets, tests, and runtime code. Examples
     include old `data-testid` names, old control types (`select`/dropdown vs
     menu/toggle), old screen labels, old component shorthands, and outdated
     prototype route names.
9. Check whether existing gates prove the current semantic contract. If a gate
   only proves structure counts or historical expectations, record the gap and
   prefer adding lint, unit tests, negative fixtures, smoke tests, or drift checks
   before moving to the next target.
   - For `ui-design` source-level parity, computed style, bounding-box, and
     screenshot checks are necessary but not sufficient. Also reverse-audit
     `ui-design/src/*.jsx`, `ui-design/src/app.jsx`, and
     `ui-design/src/primitives.jsx` for DOM shape, control type, menu/popup
     hierarchy, icons, labels, aria state, and primary interaction paths; then
     verify official frontend tests fail if any of those source-level structures
     drift.
10. For completed code phases, verify actual test evidence exists for the implemented checklist scope, including meaningful negative/boundary assertions where the coverage matrix marks them in scope.
    - When a checklist, plan, report, or previous handoff cites a focused `go test ... -run TestX` gate, verify that the command executed at least one intended test before treating it as evidence.
    - Use `go test -list` or source search for `func Test...` to confirm the named test or regex exists when the command output is unavailable or ambiguous.
    - Treat `testing: warning: no tests to run`, package output ending in `[no tests to run]`, or a focused `-run` pattern with zero matching tests as a gate failure and record a finding. Do not count it as PASS even if the command exits 0.
    - In `--fix` mode, map the finding to either adding the missing executable test, correcting the focused gate name, or reopening the checklist item whose evidence was no-op.
11. For completed feature phases, verify BDD evidence exists: `bdd-plan` / `bdd-checklist` references, completed scenario asset/execution items, a passed `BDD-Gate:` verification note, and scenario coverage for the primary journey plus the highest-risk alternate or failure/recovery journey per deployable phase.
    - Treat `completed` plan/checklist/test/BDD documents that still contain unchecked BDD items, `partial`/`pending`/`next pass` comments, or "asset readiness" language as blocking evidence drift, not as PASS.
    - Read scenario `trigger.sh` / `verify.sh` scripts directly. Scenario directory, README, or INDEX presence does not prove coverage unless the trigger executes the dedicated tests and the verifier asserts the relevant runtime/negative conditions.
    - Treat scenario wrapper scripts as evidence artifacts in their own right, not just launchers for the test body. Do not stop after reading the Go test body; wrapper process-success proof is separate evidence. For shell wrappers around focused Go tests, verify the trigger preserves the real test process exit status (`pipefail` where the shell supports it, or explicit status capture around `tee`) and record a D-series finding if it uses `go test | tee` without preserving the `go test` status. The verifier must require the intended test to start, require a passing marker such as `--- PASS`, require package-level `ok`, and reject `--- FAIL`, package `FAIL`, and `no tests to run`.
    - Read the dedicated test bodies named by BDD scenarios and map every material BDD checklist assertion to concrete assertions in code. A focused `go test -run '^ExactScenario$'` wrapper only proves the named test executed; it does not prove that DB side effects, replay privacy, no-op absence, task-run metadata, or negative paths listed in `bdd-checklist.md` were asserted.
    - In `--fix` mode, first add the missing executable test or script assertion, then update the original checklist evidence and lifecycle state only after the gate passes.

Coverage rows to verify:

- **Primary path**: implemented behavior matches the spec/plan and has current passing evidence.
- **Alternate path**: auth/permission, config/provider/profile, locale/theme/mode,
  optional input, and feature-disabled variants are implemented or explicitly N/A.
- **Failure / recovery path**: invalid input, missing data, downstream failure,
  timeout/retry, partial state, conflict, cancellation, cleanup, and recovery are
  handled and tested where user/system correctness depends on them.
- **Boundary condition**: empty/min/max, duplicate, ordering, pagination,
  concurrency/idempotency, rerun safety, migration on non-empty data, unknown
  enum/route/config/provider, and retention/deletion cases are covered where in scope.
- **Cross-layer contract**: API/schema/OpenAPI/shared type/codegen, fixtures/mock
  parity, event/job, DDL/config, generated artifacts, README/Make/script, and
  scenario data contracts remain aligned.
- **Privacy / security / observability**: authz, sensitive data redaction,
  secret/token persistence, audit/log/metric behavior, and unsafe input handling
  are checked against code and tests.
- **UX quality**: UI loading/empty/error states, accessibility, localization
  fallback, display preferences, responsive-state behavior, and visible copy are
  checked against current UI truth sources when relevant.
- **Regression / legacy-negative**: retired routes/modules/tags/schema names,
  events/jobs/config flags, feature keys, and model/provider assumptions are
  absent from active code or guarded by explicit drift gates.

Review dimensions:

- `R-series`: consistency with spec definitions
- `P-series`: completeness against plan tasks and error handling
- `E-series`: best-practice code quality, tests, naming, security
- `D-series`: deep reconcile evidence, artifact coverage, negative legacy-scope
  search, and semantic gate adequacy
- `C-series`: coverage matrix proof; primary, alternate, failure/recovery,
  boundary, cross-layer contract, privacy/security/observability, UX, and
  regression-negative rows each map to current artifacts or an explicit N/A
  rationale

Output rules:

- Each finding includes check ID, severity, description, and `file:line`.
- Cite the relevant spec/plan evidence.
- Distinguish drift from acceptable design-preserving extension.
- Put any extra findings under `Extended Findings` with `X-L2-*` IDs.
- Include a `Deep Evidence` section listing artifact map coverage, negative
  searches, focused gates/tests run, and any gate gaps discovered or hardened.
  For every cited `go test -run` focused gate, state the matching test name(s)
  or explicitly record the no-op finding if zero tests ran.
  For scenario wrapper evidence, state how `trigger.sh` preserves the real test
  process exit status and how `verify.sh` proves pass/fail/no-op status rather
  than merely grepping a test name or package path.
  For lifecycle / runtime capabilities, state the production entrypoint audited
  and the production-wiring test, smoke, or scenario that proves startup and
  shutdown/drain behavior.
- Include a `Coverage Matrix Evidence` section summarizing which coverage rows
  are proven by current artifacts, which are explicitly N/A, and which are gaps.

### Step 5: Branch by mode

**Review mode**:

1. Review all in-scope phases.
2. Aggregate the report.
3. If only target-level scope could be derived, say so explicitly.
4. Stop.

**Fix mode**:

1. Generate fix proposals with diff-style previews.
2. Explicitly degrade target-level-only findings to preview-only when no checklist section can be mapped automatically.
3. Wait for user confirmation.
4. Run the `AGENTS.md` §7 branch guard before the first remediation file edit.
   The branch prefix must express the work type or domain, such as `fix/`,
   `docs/`, or `feat/`; never create `codex/`, `claude/`, `gemini/`,
   `agent/`, or other tool-name branches.
5. For each accepted finding, map it to a checklist section.
6. Execute remediation only through `/tdd --section`.

### Step 6: Fix via `/tdd --section`

When the user accepts a fix proposal:

1. Map each accepted finding to a checklist section; unmappable findings stay preview-only.
2. If an item is incorrectly marked `[x]`, the preview must offer either:
   - reopen the item, or
   - append a remediation item
3. When a finding is supported only by target-level discovery or git diff, ask the user to confirm the checklist section before routing to `/tdd --section`.
4. Invoke:

```text
/tdd --file {checklist-path} --section {phase-prefix} --references {ref1},{ref2},...
```

5. After `/tdd` completes, run focused tests and adjacent/regression tests.
6. Verify compilation succeeds for affected packages.
7. Any test or compile failure means the fix is not applied.

## Guardrails

- Findings with no concrete checklist-section mapping are never auto-fixed.
- `/plan-code-review --fix` is remediation only; it does not own plan lifecycle sync or retrospective.
- Do not edit code directly outside the `/tdd` Red-Green-Refactor loop once fix mode is accepted.
- Do not conclude PASS from checklist state, old reports, previous green tests,
  or small diffs. PASS requires current artifact evidence and current semantic
  gate coverage.
- If the review exposes a shallow workflow or gate blind spot, update the
  relevant rule, lint/test gate, or plan review notes before proceeding to the
  next target.
