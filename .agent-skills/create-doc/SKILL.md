---
name: create-doc
description: "IMPORTANT: Invoke this skill automatically when creating any document under docs/. Do NOT create documents in docs/ without invoking this skill first. Create project documentation with proper formatting and index updates. Use when creating spec documents, plans, reports, or API docs. Triggers on /create-doc or when asked to create documentation in docs/ directory."
---

# Create Documentation Skill

Creates project documentation following conventions for each document type.

## Prerequisites

Before using this skill, ensure `docs/` directory structure exists with README.md, `TEMPLATES.md` (when applicable), and INDEX.md files.
If not, run `/init-docs` first to initialize the documentation structure.

## Document Types and Locations

| Type | Directory | When to Use |
|------|-----------|-------------|
| Architecture/Module design | `docs/spec/` | System design, module specs |
| Implementation/Refactor/Test plans | `docs/plan/${subject}/` | Work plans with checklists |
| Code review/Validation reports | `docs/reports/` | Review results, validation, post-pass retrospectives |
| Agent analysis discussions | `docs/discuss/` | Analysis, decision records |
| API definitions | `docs/apis/` | Interface specifications |
| Work journals | `docs/work-journal/` | Daily progress (use /work-journal) |
| Bug records | `docs/bugs/` | Bug diagnosis records (use /bug-report) |

## Workflow

### Step 1: Read the specification

**Before creating any document**, read the README.md in the target directory for rules, then read `TEMPLATES.md` when the directory provides one:

| Activity | Read First |
|----------|------------|
| Architecture/Module design | `docs/spec/README.md` + `docs/spec/TEMPLATES.md` |
| Implementation/Test plans | `docs/plan/README.md` + `docs/plan/TEMPLATES.md` |
| Code review/Validation reports | `docs/reports/README.md` + `docs/reports/TEMPLATES.md` |
| Agent analysis discussions | `docs/discuss/README.md` + `docs/discuss/TEMPLATES.md` |
| API definitions | `docs/apis/README.md` + `docs/apis/TEMPLATES.md` |
| Bug records | `docs/bugs/README.md` + `docs/bugs/TEMPLATES.md` (prefer using `/bug-report` skill) |

README contains naming conventions, lifecycle rules, and checklist items. `TEMPLATES.md` contains copyable structure examples.

### Step 2: Create document with standard Header

Follow naming conventions from the README and structure examples from `TEMPLATES.md`.

**Standard Header is mandatory.** Every new document must include the Header in the exact field order defined below.

**Spec documents** (`docs/spec/*.md`):

```markdown
> **版本**: 1.0
> **状态**: draft
> **更新日期**: YYYY-MM-DD
```

**Plan documents** (`docs/plan/${subject}/*.md`, excluding checklist):

```markdown
> **版本**: 1.0
> **状态**: draft
> **更新日期**: YYYY-MM-DD
```

**Checklist documents** (`*-checklist.md`):

```markdown
> **版本**: 1.0
> **状态**: draft
> **更新日期**: YYYY-MM-DD
```

Valid status values: `draft`, `active`, `completed`, `superseded`, `deprecated`.
New documents default to `draft`. Field order is fixed and must not be rearranged.

**For plans**: Create a directory `docs/plan/${subject}/` containing:
- `context.yaml` - **Required** plan-context manifest for `/implement`, `/plan-review`, and `/plan-code-review`
- `implementation.md` - Implementation plan
- `implementation-checklist.md` - Implementation checklist
- `unit-test-plan.md` (optional) - Unit test plan
- `unit-test-plan-checklist.md` (optional) - Unit test checklist
- `bdd-plan.md` (optional) - BDD scenario plan
- `bdd-checklist.md` (optional) - BDD scenario asset and execution checklist

**context.yaml** must be generated with the plan. Minimal template lives in `docs/plan/TEMPLATES.md`.

Minimal shape:

```yaml
apiVersion: plancontext.agent.dev/v1alpha1
kind: PlanContext
metadata:
  name: ${subject}
spec:
  defaultTarget: backend
  discovery:
    aliases:
      - ${subject}
    keywords:
      - TODO: add issue keywords
  targets:
    backend:
      plan: ./implementation.md
      checklist: ./implementation-checklist.md
      spec: ../../spec/${subject}-design.md
      discovery:
        packages:
          - TODO: add primary packages/modules
```

`context.yaml` 仅用于 plan 文档关联和问题检索索引。不要写入 `commands`、脚本名、Make target 或人工操作步骤；若需要稳定检索标识，可补充 `uiRoutes` / `apiNames`。

If creating frontend or unit-test sub-plans, add corresponding targets.
The plan-context manifest is the shared contract consumed by the implement-owned validator at
`.agent-skills/implement/shared/scripts/validate_context.py`.

For revisions to an existing `completed` plan:

- update the original plan directory instead of creating a sibling follow-up directory
- revise the original spec / plan / checklist together
- increment affected document versions
- set the plan/checklist `状态` back to `active` while execution is pending, then restore `completed` after verification
- keep `context.yaml` as the same subject manifest and refresh discovery metadata only when needed
- use a `## 修订记录` block when an explicit delta trail is useful
- only create a new plan subject when no existing subject matches or the user explicitly requests a separate workstream

For **new plans**, follow the canonical sequential forms from `docs/plan/README.md` and `docs/plan/TEMPLATES.md`:
- plan phase heading: `### Phase N: ...`
- plan task heading: `#### N.M ...`
- checklist section heading: `## Phase N: ...`
- checklist item ID: `- [ ] N.M ...`

If a test checklist is created, use matching phase-number section headings by default (for example `## Phase 2: API tests`) so `/tdd` can infer section mapping without `<!-- phase-mapping: -->`.

For test plans and test checklists:

- Use completion language tied to the enumerated tests, for example `Phase 2 本计划定义的单元测试项全部通过`.
- Do not create checklist items or acceptance criteria that use hard code coverage percentages such as `coverage >= 75%` or `覆盖率 ≥ 80%`.
- If coverage is worth recording, keep it as observational context only; never make it a completion, commit, or phase-exit gate.

Legacy forms such as `### 3.1 Phase 1: ...`, `#### 3.1.1 ...`, `## 1 Phase 1: ...`, `### W1.Auth: ...`, or explicit `<!-- phase-mapping: -->` comments are compatibility formats for existing documents, not the default for new ones.

For **spec documents**, if the design contains confirmed tradeoffs or open product/architecture choices, explicitly include:
- `设计决策记录` for decisions already confirmed
- `待确认事项` or `用户决策` for unresolved choices that require user arbitration

### Step 3: Validate Header, then update INDEX.md

**Before updating INDEX**, verify the new document's Header:
1. All required fields are present.
2. Fields are in the correct order.
3. `状态` is a valid enum value.

If Header is invalid, **abort INDEX update** and fix the Header first.

**For `docs/spec/INDEX.md`**:
- Add the new document to the appropriate **domain group** (e.g., "核心组件", "AI Agent 工程实践").
- Fill in `版本`, `状态`, `更新日期` columns from the Header.
- Domain groups are managed manually — do not create new groups without user approval.

**For `docs/plan/INDEX.md`**:
- Add the new document to the **status group** matching the Header `状态`:
  - `draft` → "草稿（Draft）" section
  - `active` → "进行中（Active）" section
  - `completed` → "已完成（Completed）" section
  - `superseded` → "已取代（Superseded）" section
- Fill in `版本` and `更新日期` columns from the Header.
- `↳` sub-plan rows follow their parent plan's group.

## Markdown Format

```markdown
# Document Title     (main title, one per doc)
## 1 First Section   (numbered)
### 1.1 Subsection   (hierarchical)
#### 1.1.1 Detail    (hierarchical)
```

- Heading levels must be consecutive (no skipping)
- Use numbered sections for structure

## Checklist Principle

Checklists are the single source of truth for task completion:

1. **Atomic updates**: Update checklist when modifying plan content
2. **Completion criteria**: All items checked = task complete
3. **No false marking**: Never mark task `#completed` if checklist incomplete
4. **Skip with reason**: If skipping an item, annotate why in checklist

## Prohibited Actions

- Creating documents without reading the corresponding README.md and `TEMPLATES.md` when available
- Forgetting to update INDEX.md after creation
- Skipping heading levels in markdown
- Creating plan files directly in `docs/plan/` instead of a subdirectory
- Writing test plan acceptance criteria or checklist items that treat code coverage percentages as hard gates
