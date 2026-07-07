---
name: init-docs
description: Initialize docs directory structure with README.md, TEMPLATES.md, INDEX.md, and supporting templates for work journals, spec-centric plans, specs, and other documentation. Run this skill once when setting up a new project or adding documentation infrastructure. Triggers on /init-docs.
---

# Initialize Docs Skill

Sets up the `docs/` directory structure with separated rule docs and template assets.

## When to Use

- Setting up a new project
- Adding documentation infrastructure to an existing project
- Before using `/work-journal`, `/create-doc`, `/tdd`, or `/bug-report`

## Directory Structure Created

```
docs/
├── README.md              # Main navigation hub
├── work-journal/
│   ├── README.md          # Rules and maintenance checklist
│   ├── TEMPLATES.md       # Copyable journal templates
│   └── INDEX.md           # Journal index
├── spec/
│   ├── README.md          # Spec-centric subject rules
│   ├── TEMPLATES.md       # Spec/plan/checklist/context templates
│   ├── INDEX.md           # Spec subject index
│   └── ${subspec}/
│       ├── spec.md
│       ├── history.md
│       └── plans/
│           ├── INDEX.md                   # Plan index scoped to this subspec
│           └── ${NNN-plan}/
│               ├── context.yaml
│               ├── plan.md
│               ├── checklist.md
│               ├── test-plan.md        # conditional
│               ├── test-checklist.md   # conditional
│               ├── bdd-plan.md         # conditional
│               └── bdd-checklist.md    # conditional
├── reports/
│   ├── README.md          # Report rules
│   ├── TEMPLATES.md       # Report templates
│   └── INDEX.md           # Report index
├── apis/
│   ├── README.md          # API doc rules
│   ├── TEMPLATES.md       # API templates
│   └── INDEX.md           # API index
├── discuss/
│   ├── README.md          # Discussion archive rules
│   ├── TEMPLATES.md       # Discussion templates
│   └── INDEX.md           # Discussion index
└── bugs/
    ├── README.md          # Bug record rules
    ├── TEMPLATES.md       # Bug record template
    ├── INDEX.md           # Bug index
    └── PATTERNS.md        # Bug pattern library
```

`README.md`、`TEMPLATES.md` 和 INDEX.md 的职责必须分离：

README.md、`TEMPLATES.md` 和 INDEX.md 分别承载规则、模板和索引，不得混排。

- README 只承载目录规范、命名规则、流程约束和检查清单
- `TEMPLATES.md` 只承载可复制模板和结构示例
- `INDEX.md` 只承载目录索引

## Workflow

### Step 1: Check existing structure

Check whether `docs/` exists and which subdirectories/files are already present.

### Step 2: Create missing directories

Create any missing subdirectories from the list above.

### Step 3: Create scaffold files

For each subdirectory, create README.md, `TEMPLATES.md` (when applicable), and INDEX.md from templates:

| Directory | README Template | `TEMPLATES.md` Template | INDEX Template |
|-----------|-----------------|-------------------------|----------------|
| `docs/` | [docs-readme.md](./templates/docs-readme.md) | N/A | N/A |
| `work-journal/` | [work-journal-readme.md](./templates/work-journal-readme.md) | [work-journal-templates.md](./templates/work-journal-templates.md) | [work-journal-index.md](./templates/work-journal-index.md) |
| `spec/` | [spec-readme.md](./templates/spec-readme.md) | [spec-templates.md](./templates/spec-templates.md) | [spec-index.md](./templates/spec-index.md) |
| `spec/<subspec>/plans/` | N/A — use `docs/spec/README.md` | N/A — use `docs/spec/TEMPLATES.md` | [subspec-plans-index.md](./templates/subspec-plans-index.md) |
| `reports/` | [reports-readme.md](./templates/reports-readme.md) | [reports-templates.md](./templates/reports-templates.md) | [reports-index.md](./templates/reports-index.md) |
| `apis/` | [apis-readme.md](./templates/apis-readme.md) | [apis-templates.md](./templates/apis-templates.md) | [apis-index.md](./templates/apis-index.md) |
| `discuss/` | [discuss-readme.md](./templates/discuss-readme.md) | [discuss-templates.md](./templates/discuss-templates.md) | [discuss-index.md](./templates/discuss-index.md) |
| `bugs/` | [bugs-readme.md](./templates/bugs-readme.md) | [bugs-templates.md](./templates/bugs-templates.md) | [bugs-index.md](./templates/bugs-index.md) + [bugs-patterns.md](./templates/bugs-patterns.md) |

Template rule:

- `docs/*/README.md` and the matching `*-readme.md` template must stay semantically aligned.
- `docs/*/TEMPLATES.md` and the matching `*-templates.md` template must stay semantically aligned when the directory owns a local template asset.
- New project scaffold 默认只输出当前项目契约，不应在 README 或 `TEMPLATES.md` 中混入临时兼容 patch 正文。
- Latest flow is spec-centric: executable plans live under `docs/spec/<subspec>/plans/<NNN-plan>/`; only the per-subspec plan index lives under `docs/spec/<subspec>/plans/INDEX.md`; plan rules live in `docs/spec/README.md`; plan templates live only in `docs/spec/TEMPLATES.md`; do not create top-level `docs/plan/`.

### Step 4: Report results

Report which files were created and which already existed.

## Options

User can specify which directories to initialize:

- `all` (default) - Initialize all directories
- `work-journal` - Only the work-journal directory
- `spec` - Only the spec directory
- `minimal` - Only work-journal and spec
- `subspec-plans` - Only the per-subspec `plans/INDEX.md` scaffold when a subject is created
- `test-framework` - Scaffold `test/scenarios/` framework directory

## Existing Files

Default behavior is non-destructive: if README.md, `TEMPLATES.md`, INDEX.md, or `PATTERNS.md` already exists, skip it and report to the user.

When the user explicitly says this is a new project and asks to reinitialize docs with the latest flow, reset the scaffold files from templates and remove generated spec/plan subject documents that came from an incorrect structure. Preserve committed source inputs and work-journal history unless the user explicitly asks to delete them.

## Post-Initialization

After running `/init-docs`, you can use:

- `/work-journal` - Record daily work progress
- `/create-doc` - Create new documents in any directory
- `/tdd` - Follow TDD workflow with plan checklists
- `/bug-report` - Create Bug knowledge base records
