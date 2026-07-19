---
name: change-intake
description: "IMPORTANT: Use this skill when the user reports a bug, regression, broken behavior, screenshot-based issue, or asks to revise an existing feature without naming the exact plan. This is the default entry point for issue-driven work: it locates the most relevant plan/spec/bug context, decides whether the problem is implementation drift or a design/feature change, revises the matched spec/plan/checklist in place when the subject already exists, and routes the session to the right owner skill."
---

# Change Intake Skill

Route issue-driven requests before coding. `/change-intake` is the user-facing
entry point for:

- bug reports without a plan name
- screenshots or vague UI/API failures
- regressions against recently delivered work
- feature revisions that may require spec/plan updates first

This skill does **not** replace `/implement`, `/plan-review`, or
`/plan-code-review`. It decides which one should run next.

## Usage

- `/change-intake "login page blank after refresh"`
- `/change-intake "secret edit page still posts intentBlocks"`
- `/change-intake "need to revise selector behavior for authored blocks"`

## Required Inputs

The user can provide any combination of:

- free-form issue text
- screenshot / visible error text
- API name, route, CLI command, field name, or BUG ID

If a screenshot is present, extract any visible strings, labels, field names,
or request/response errors before ranking candidates.

## Shared Inputs To Read First

Before matching a plan, read:

1. `docs/work-journal/INDEX.md`
2. the latest work-journal entry relevant to the topic
3. `docs/bugs/PATTERNS.md` when the request is a bug or regression

Use the new matcher script for deterministic candidate ranking:

```bash
python3 .agent-skills/change-intake/scripts/match_change_context.py \
  --plan-root docs \
  --query "<issue text>"
```

## Branch Guard Before Mutation

`/change-intake` may inspect files on any branch, but it must not mutate files on
the default parent branch. Run this guard before the first file edit, including
spec / plan / checklist revision, generated document creation, formatting,
`bug-report`, `retrospective`, or downstream implementation handoff that will
write files:

1. Check the current branch and worktree with `git status --short --branch`.
2. If the current branch already matches the session feature branch, treat the
   run as a retry/resume and continue.
3. If the current branch is the default parent branch and the worktree is clean,
   update the parent branch with fast-forward-only semantics, then create or
   switch to a feature branch before editing files. The branch prefix must
   express the work type or domain, such as `fix/`, `docs/`, `design/`, or
   `spec-design/`; do not create `codex/`, `claude/`, `gemini/`, `agent/`, or
   other tool-name branches.
4. If dirty changes already came from the current session while still on the
   default parent branch, create the feature branch immediately while preserving
   those changes, report the recovery, and continue only after the branch switch.
5. If the default parent branch is dirty for unclear or user-owned reasons, stop
   and ask the user before creating or switching branches.
6. If the current session is on a freshly created tool-name branch that has not
   been pushed or shared, rename it to the semantic repository prefix before
   mutating files. If it may already be externally referenced, stop and ask the
   user.
7. If a non-parent branch has unrelated dirty changes and does not match the
   session feature branch, stop and ask the user before mutating anything.

Never revise spec / plan / checklist on the default parent branch. For completed
plans that require in-place revision, the branch guard must complete before
setting the plan back to `active` or changing any owner document. For pure
proposal/backlog guidance with no file edits, no branch switch is required.

## Matching Workflow

### Step 1: Collect signals

Extract and normalize:

- user wording
- screenshot OCR / visible UI text
- API names
- routes
- commands
- field names
- BUG IDs

Build one combined query string and pass it to the matcher script.

### Step 2: Rank candidate plan targets

Interpret matcher output as:

- `high` confidence: top candidate is clear; load it directly
- `medium` confidence: top candidate is preferred, but compare the next result
- `low` confidence: treat candidates as hypotheses and verify them with live repo search before selecting an owner
- `none`: fall back to manual repo search and explain the gap

Exact scenario README Owner evidence outranks generic API / route keyword overlap.
Active / draft status is only a bounded tie-breaker for otherwise close semantic
matches; a completed plan with an exact BUG or scenario owner remains eligible.
If the recommendation conflicts with the current active plan or an explicit Owner
link, inspect both artifacts before mutation even when confidence is `high`.

Rules:

- Prefer the `recommended` candidate only for `high`/`medium` confidence and when it does not conflict with stronger owner evidence
- If confidence is `low`, search current Markdown, code, routes, API identifiers, BUG/scenario Owner links and Git evidence before presenting a choice. Only ask the user after live evidence remains ambiguous
- If the matcher finds nothing, search by BUG ID / API / route / command manually

### Step 3: Decide plan lifecycle handling

After selecting a candidate:

1. Every candidate must have a minimal `context.yaml`; inspect its `contextPath` and owner files. Missing `context.yaml` is a document-contract gap, not a supported owner bypass: route it to the current plan/docs owner for repair before implementation or review handoff.
2. Read the candidate `context.yaml` and validate it with:

```bash
python3 .agent-skills/implement/shared/scripts/validate_context.py \
  --context docs/spec/<subspec>/plans/<plan>/context.yaml \
  --docs-root docs \
  --target <target>
```

   Then read the validated plan / checklist / spec and optional first-class test/BDD documents. If `plan` and `checklist` resolve to the same single-plan file, preserve both roles but read the body once.
3. Inspect the selected plan Header `状态`.

Routing rule:

- `active` / `draft`: the plan is still live; continue with the original plan context
- `completed`: revise the original spec / plan / checklist in place before coding
- Do not create sibling follow-up / bugfix docs for same-subject revisions by default.

## Classification Workflow

### Step 4: Determine change type

Classify the request before coding:

- `implementation drift`
  - intended design is still correct
  - code, generated artifacts, tests, or deployment drifted from the plan/spec
- `design/feature change`
  - user expectation changes the design, contract, workflow, or target behavior
  - new user-visible behavior, schema, API, or compatibility rule is needed
- `uncertain`
  - the issue may be design drift, but the documents are ambiguous or stale

### Step 5: Route to the next skill

#### A. Active/Draft/Completed + implementation drift

- If the selected plan is `completed`, first revise the original plan directory in place:
  - update the original spec / plan / checklist together
  - increment the affected document versions
  - set the plan/checklist `状态` back to `active` before execution
- If the issue is already within the current plan scope, continue with `/implement`
- If the scope is already implemented but needs remediation, prefer `/plan-code-review --fix`
- If document drift is the blocker, use `/plan-review --fix` first

#### B. Design/Feature change

Do not code first.

1. Revise the original spec first
2. Revise the original plan/checklist in the same directory
3. Run `/plan-review --fix` if the updated docs need consistency cleanup
4. Only then continue to implementation

Create a new spec/plan subject only when:

- no existing plan/spec cleanly matches the requested change
- the user explicitly requests a separate standalone workstream

#### C. Uncertain classification

Stop and show:

- the candidate plan
- the conflicting evidence
- whether the ambiguity is in spec, plan, or current code

If the ambiguity is document-owned, resolve it through spec/plan updates before
coding.

## In-place Revision Contract

When `/change-intake` revises an existing subject:

- update spec/plan/checklist before code changes
- keep using the original spec-centric plan directory and `context.yaml`
- keep `context.yaml` limited to the minimal link contract; do not add routing vocabulary or branch/version metadata
- add or refresh a `## 修订记录` section when the change benefits from an explicit delta trail
- preserve links to related bug records and reports in the owning Markdown documents when relevant

If the user currently wants proposal/backlog guidance only, stop before mutating
the original docs and present the recommended in-place revision scope instead.

Do not create sibling follow-up docs as passive notes for the same subject.

## Close-out Workflow

After the fix or revision is complete and verified:

1. run `plan-review` on the relevant plan if spec/plan changed
2. run `sync-doc-index` when Header / INDEX projections changed
3. evaluate `bug-report` when the session fixed a real bug
4. invoke `/retrospective --this` before final close-out

`/change-intake` is an entry skill, not a delivery owner. Once it has routed the
session into a concrete plan flow, the downstream delivery skill still owns its
normal testing and lifecycle duties. Once `/change-intake` has aligned the
original docs in place, it must hand off to the next owner in the same session.
