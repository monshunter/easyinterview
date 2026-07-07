---
name: tdd
description: "IMPORTANT: Invoke this skill automatically for any code logic implementation following a plan checklist, including front-end, back-end, tooling, migration, codegen, or test helper logic. Do NOT write implementation code against a checklist without invoking this skill first. Test-Driven Development workflow with strict checklist progression, immediate checklist sync, and document lifecycle coordination. Use when implementing features from implementation/unit-test/e2e/frontend checklists. Triggers on /tdd or when working against checklist-based plans."
---

# TDD Development Skill

Execute checklist-driven development with mandatory Red-Green-Refactor cycles.

Any code logic implementation must use this skill: front-end, back-end, tooling, migration, codegen, or test helper logic all require Red-Green-Refactor execution and passing test evidence before checklist completion.

## Usage

- `/tdd` - Use checklist already present in current context
- `/tdd --file <checklist>` - Explicit checklist path
- `/tdd --file <checklist> --references f1,f2,...` - Checklist + additional references
- `/tdd --file <checklist> --test-checklist <path> --references f1,f2,...` - With associated test checklist
- `/tdd --file <checklist> --phase-commit <plan-name> --references f1,f2,...` - Enable phase-boundary auto-commit
- `/tdd --file <checklist> --section <prefix> --references f1,f2,...` - Section-scoped execution
- `/tdd -h` - Show help only, do not execute workflow
- `/tdd -h -v` - Show verbose help (including workflow), do not execute

## Argument Contract

- `--file <path>`: checklist markdown file
- `--test-checklist <path>`: associated test checklist; prefer section headings that mirror implementation phase numbers, with `<!-- phase-mapping: -->` compatibility support
- `--phase-commit <plan-name>`: enable phase-boundary auto-commit via `/work-journal --auto`
- `--references <p1>,<p2>,...`: comma-separated markdown references (plan/spec/test-plan)
- `--section <prefix>`: Section ID/prefix (for example `1` or `W1.Auth`) to scope execution to a single checklist section

Rules:

- When arguments are provided, `--file` is required.
- Split `--references` by comma, trim whitespace, ignore empty segments.
- If a referenced file is missing, stop and report error before coding.

### `--section` mode

When `--section <prefix>` is provided, `/tdd` operates in **section-scoped mode**:

1. **Section matching**: Find the `## ` heading line that starts with `## {prefix}` (e.g., `## W1.Auth:`). Process only the checkbox items (`- [ ] ...` / `- [x] ...`) between that heading and the next `## ` heading (or end of file).
2. **Step 2 (lifecycle status) is skipped**. The caller (for example `/implement` or another higher-level remediation/orchestration skill) is responsible for lifecycle management.
3. **Step 3 (select next item)** only considers items within the matched section.
4. **Steps 4-8** are unchanged (Red-Green-Refactor + immediate checklist update).
5. **Step 10 (completion lifecycle sync) is no longer used**: when all items in the section are checked, report `Section {prefix} complete` and stop. Do NOT trigger global lifecycle sync.
6. **Prohibited**: modifying checklist items outside the matched section.

When `--section` and `--phase-commit` are both present, trigger Step 9.5 exactly once when that section's last item and mapped test items are complete.

## Workflow

### Step 0: Handle help flags

- `-h`/`--help`: show skill name, description, usage, then stop.
- `-h -v`/`--help --verbose`: show usage + full workflow, then stop.

### Step 1: Load working documents

**With arguments**:

1. Read checklist from `--file`.
2. Read each file from `--references`.
3. If exactly one loaded reference has basename `bdd-plan.md` or `bdd-test-plan.md`,
   classify it as the BDD scenario source for Step 5B. New canonical plans key BDD verification by behavior-oriented
   scenario IDs such as `E2E.P0.001` / `E2E.P0.004`;
   some active plan files may still carry `AC-*` mappings as compatibility input.
4. If exactly one loaded reference has basename `bdd-checklist.md`, classify it as the BDD
   asset/execution checklist for Step 5B.
5. If multiple loaded references match either BDD role, stop and ask the user to
   disambiguate before continuing.
6. If `--test-checklist` is provided:
   a. Read the test checklist file.
   b. Build a mapping table by reading each `## ` section heading:
      - default: infer the implementation phase from the heading itself (for example `## Phase 2: API tests` → `2`)
      - compatibility: if an immediate `<!-- phase-mapping: {id} -->` comment exists, use that explicit mapping instead
   c. Build a mapping table: `{impl-phase-id} → [test-section-heading, ...]`.
   d. Log the mapping summary (e.g., "Phase 2 → test sections 1, 2, 14").

**Without arguments**:

1. Locate checklist from current context.
2. If multiple candidates exist, ask user to choose one.
3. Read the chosen checklist and existing context references.

### Step 1.5: Frontend / Backend Contract Preflight

Before selecting the next checklist item or editing files, inspect the loaded
checklist/references and intended write scope. If the work involves `frontend/`,
`backend/`, `openapi/`, `migrations/`, `config/ai-*`, `deploy/dev-stack/`, or
`test/scenarios/`, read the current contracts that govern that scope:

1. `docs/development.md` §2 Frontend / Backend Contract Workflow.
2. The relevant root README files for the touched modules.
3. UI-visible work: relevant `docs/ui-design/` document plus
   `ui-design/src/*.jsx` / `ui-design/src/primitives.jsx` /
   `ui-design/src/app.jsx` sources.
4. API/fixture/handler work: `openapi/openapi.yaml`, related fixture files,
   generated artifacts, and the operation matrix.
5. Local integration or scenario work: `deploy/dev-stack/README.md` and
   `test/scenarios/README.md`; distinguish Docker Compose external dependencies,
   host-run app commands, and repo-tracked local scenario runners.

If the current checklist item requires a missing operation matrix or contradicts
these contracts, stop and route back to `/plan-review --fix` or user-approved
plan revision. Do not continue by relying on memory of prior repository state.

### Step 2: Check plan/checklist lifecycle status

After loading documents:

1. Read Header `状态` from checklist and its associated plan.
2. If first implementation run and status is `draft`, ask user:
   > "Plan/checklist is `draft`. Switch to `active` now?"
3. If user approves:
   - Update Header `状态` to `active`.
   - Update `更新日期` to today (`YYYY-MM-DD`).
   - If Header field order/enum/date is non-compliant, invoke `/sync-doc-index --fix-header` first.
   - Invoke `/sync-doc-index --fix-index` to sync `docs/spec/INDEX.md`.

### Step 3: Select next checklist item (strict order)

1. Find next unchecked item in original checklist order.
2. Announce the exact item before coding:

```text
Executing: {checklist-file} -> {section/item-id} {item-title}
```

3. Never skip unchecked items.

### Step 3B: BDD-Gate detection

After selecting the next item in Step 3, check if the item text contains the `BDD-Gate:` prefix:

- **No** → Continue with Steps 4-8 (normal Red-Green-Refactor cycle).
- **Yes** → Skip Steps 4-8 entirely. Jump to Step 5B (Deploy-Verify protocol).

### Step 3C: Hard Coverage Gate detection

After Step 3/3B, inspect the current checklist item text:

- If the item is a raw code coverage threshold gate (for example `coverage >= 75%`, `覆盖率 ≥ 80%`, `line coverage` targets), do **not** try to satisfy it by inventing extra tests.
- Treat it as a document-owned issue: stop, report that the checklist is using a forbidden completion gate, and route back to `/plan-review --fix` or an explicit user-approved plan revision.
- Only continue implementation after the plan/checklist is repaired to use execution-based completion criteria.

### Step 4: Red phase (MANDATORY)

For the current checklist item, before touching non-test source files:

1. Determine target test file using project conventions (`*_test.go`, `*.test.ts`, etc.).
2. Add/adjust a test case for this item.
3. Run a focused test command and verify Red:
   - Go: `go test ./path/to/pkg -run TestName -count=1`
   - JS/TS: use repository-configured focused command
4. If test passes immediately, record explicit note:

```text
Red phase note: test already passes (pre-existing implementation). Continue with coverage adequacy check.
```

### Step 5: Green phase (minimal implementation)

1. Modify non-test source files only after Step 4.
2. Implement the minimal change required for current item.
3. Re-run focused test; it must pass.

### Step 5B: Deploy-Verify protocol (BDD-Gate items only)

This step applies only when Step 3B identified a `BDD-Gate:` item. It replaces the Red-Green-Refactor cycle (Steps 4-8) for that item.

1. **Parse scenario references**: Extract scenario identifiers from the item text (for example `BDD-Gate: 验证 E2E.P0.001, E2E.P0.004 通过` → `[E2E.P0.001, E2E.P0.004]`).
2. **Load BDD scenarios**: Prefer the loaded reference whose basename is `bdd-plan.md` or `bdd-test-plan.md`; when the caller handed off validated files from `context.yaml`, this is the first-class `bddPlan` document. Use the parsed scenario IDs as the primary lookup key. If the current checklist item is an AC-style compatibility gate, treat that as compatibility input and resolve through the AC mapping in `bdd-plan.md`, `bdd-test-plan.md`, or the spec §验收标准 table.
3. **Check BDD checklist prerequisite**: If a loaded `bdd-checklist.md` exists, find the checklist section/items for the parsed scenario IDs. Every asset and execution item for those scenarios must already be checked before the main checklist `BDD-Gate` can be marked complete. If any related BDD checklist item is unchecked, stop and report the exact missing items instead of running or marking the gate.
4. **Deploy and verify**:
   a. If a matching scenario test directory exists under `test/scenarios/` for the parsed scenario IDs → require the documented script contract and execute the scenario scripts (setup → trigger → verify → cleanup) after confirming the framework-defined environment is ready.
   b. If the scenario directory exists but the required scripts are missing → stop and report that the BDD asset contract is incomplete. Do not fall back to manual verification for repo-defined scenarios.
   c. If no scenario directory exists and the plan/framework explicitly documents a manual-only compatibility verification path → execute that manual verification and record evidence.
5. **Judge result**:
   - **Pass** → Mark the BDD-Gate item complete. Append verification evidence as an HTML comment on the line below: `<!-- verified: YYYY-MM-DD method={scenario|manual} bddChecklist=complete -->`.
   - **Fail** → Do NOT mark the item complete. Report failure reason and return to fix the implementation. The current phase cannot advance until the BDD-Gate passes.
5. After marking complete (or reporting failure), continue to Step 8 (checklist update) or back to the failing implementation item.

### Step 6: Refactor phase

1. Refactor for readability/maintainability without behavior change.
2. Re-run focused tests after refactor.

### Step 7: Verification for current item

1. Run focused tests for the item.
2. Run adjacent/regression scope tests (same package/module) as needed.
3. If tests fail, item remains incomplete.

### Step 7.5: Structural Contract Change Gate

Before marking the current item complete, apply this gate whenever the item changes a checked-in contract such as field names, schema keys, wire shapes, generated API types, config keys, or other consumer-facing structures.

1. Update all repo-tracked direct consumers required by the change before checking the item off.
2. Rebuild or regenerate any repo-tracked artifacts that derive from the changed contract.
3. Run focused verification on the direct consumer surfaces named by the checklist, test checklist, or references.
4. Run a repo-wide search for the non-current contract shape and confirm any remaining matches are intentional.
5. If verification depends on built binaries, deployed components, or cluster-installed schemas, confirm the consumer artifact under test is freshly rebuilt or re-synced rather than stale.

Environment-specific consumer surfaces and live verification steps belong to the relevant framework README or layer README; use those documents as the source of truth.

### Step 8: Update checklist immediately

After current item is verified green:

1. Mark the exact checklist checkbox as complete.
2. Save checklist changes immediately (no batch update).
3. Continue to next unchecked item.

### Step 9: Execute mapped test items (when --test-checklist provided)

When `--test-checklist` is provided and the current implementation phase is complete
(all items in the current implementation section are checked):

1. Look up the phase-mapping table built in Step 1.
2. Collect all test checklist sections mapped to the just-completed implementation phase.
3. For each mapped test section, execute its unchecked items using the standard
   Red-Green-Refactor cycle (Steps 4–8), but updating the test checklist instead.
4. When all mapped test items are checked, continue to the next implementation phase.
5. If no test sections map to the current phase, skip this step.

Phase completion detection:

- New canonical format: the implementation checklist is grouped by `## Phase N: ...` sections. When the last unchecked item in a section is marked complete, Step 9 triggers.
- Sequential and parallel section headings remain readable for compatibility.
- Test checklist mapping prefers section-heading inference and only falls back to `<!-- phase-mapping: {id} -->` when that compatibility annotation is present.

### Step 9.5: Phase Commit Gate

When `--phase-commit` is present and the current implementation phase is complete
(including any mapped test sections executed in Step 9):

1. call `/work-journal --auto --plan <name> --phase <heading>` on the feature branch
2. keep the current feature branch checked out; do not checkout, merge, or ff-only merge the base branch automatically
3. if the phase commit succeeds, continue to the next implementation phase

Base branch integration is a separate explicit operation owned by the user,
merge/rebase workflow, or PR review stage. Do not perform it at a phase boundary.

If `/work-journal --auto` or drift repair fails, stop the current `/tdd` run immediately.
Preserve the current branch and working tree for retry or manual intervention.
Do not advance to the next phase or Step 10 while the phase-commit failure remains unresolved.

Only after Step 9.5 succeeds may `/tdd` continue to the next implementation phase.

### Step 10: Completion lifecycle sync

When all checklist items are checked:

1. If `--test-checklist` was provided, verify all mapped test checklist sections are also fully checked. If any mapped items remain unchecked, report the gap and continue executing them via Step 9 before proceeding.
2. Ask user:
   > "All checklist items are complete. Switch plan/checklist to `completed` and sync INDEX?"
3. If approved:
   - Set Header `状态` to `completed` on plan and checklist.
   - Update `更新日期` to today (`YYYY-MM-DD`).
   - If Header field order/enum/date is non-compliant, invoke `/sync-doc-index --fix-header` first.
   - Invoke `/sync-doc-index --fix-index` to sync INDEX grouping.
4. Post-pass retrospective owner rule:
   - If the current delivery entered through a higher-level caller that explicitly owns close-out, that caller owns the retrospective trigger and `/tdd` must not invoke it separately.
   - `/implement` is such a caller for full plan delivery.
   - Remediation callers such as `/plan-code-review --fix` are not delivery owners; they may route work through `/tdd --section`, but they do not own global completion or retrospective by themselves.
   - Otherwise, once completion and passing verification are established, invoke `/retrospective --this` before final close-out.

## Prohibited Actions

- Writing non-test implementation code before the corresponding Red test
- Skipping checklist items or changing execution order without approval
- Batch-checking multiple checklist items after the fact
- Claiming "tests passed" without actually running tests
- Modifying plan semantics/checklist scope without user approval
- Adding speculative tests only to satisfy a raw coverage-percentage gate instead of the planned test scope
- In `--section` mode: modifying checklist items outside the matched section scope
- Marking a structural contract change complete while repo-tracked consumers, generated artifacts, or test-facing consumer artifacts are still stale
- Marking a BDD-Gate item complete without actually executing verification (Deploy-Verify protocol)
- Marking a BDD-Gate item complete while related `bdd-checklist.md` scenario asset/execution items remain unchecked
- Marking a BDD-Gate item complete with `method=static-contract`, `method=unit-equivalent`, or any other non-runtime evidence
- Advancing to the next phase while a BDD-Gate item in the current phase remains unchecked
- When resuming an in-flight plan, do not continue implementation outside `/tdd`.
- Re-enter through `/implement` or the current `/tdd` owner path so Step 9.5 remains active.

## Plan Change Protocol

If plan/checklist content is wrong or outdated:

1. Stop current coding.
2. Explain the mismatch and proposed change.
3. Wait for explicit user approval.
4. Resume only after decision is confirmed.

Hard coverage-percentage checklist gates count as a plan/checklist mismatch and must be repaired before continuing.

## Bug Fix Protocol

When the checklist item is a bug fix:

1. Add a reproducer test first (Red).
2. Verify reproducer fails.
3. Apply fix (Green).
4. Verify test passes.
5. Evaluate and invoke `/bug-report` if bug knowledge entry is needed (mandatory project protocol after bug fix).
6. `/bug-report` remains independent from `/retrospective`; do not treat one as a substitute for the other.

## Test Completeness Requirements

- Every checklist item has at least one corresponding test assertion.
- Normal path, boundary path, and error path are covered where applicable.
- Checklist completion must be backed by passing test evidence.
- Do not invent extra coverage thresholds unless the plan explicitly requires them.
