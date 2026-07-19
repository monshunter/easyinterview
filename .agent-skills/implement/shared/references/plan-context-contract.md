# Plan Context Contract

Shared plan-context contract owned by `/implement` and reused by sibling review skills.

## Manifest Header Contract

`context.yaml` must use:

- `apiVersion: plancontext.agent.dev/v1alpha1`
- `kind: PlanContext`

This contract belongs to the shared skill layer rather than any single repo namespace.

## Shared Scripts

- `scripts/list_context_candidates.py`
- `scripts/validate_context.py`
- `scripts/generate_context_yaml.py`

## Validator Output Contract

`validate_context.py` returns normalized JSON with `files[]`.

Its validation scope is intentionally narrow: it validates `context.yaml`
structure/content plus referenced markdown paths. It does not parse or lint the
markdown files themselves; consumers read those documents directly and rely on
the `doc-init` templates for baseline document shape.

When a concrete target is selected, the normalized payload contains only the
selected target name, default target, manifest name, and validated `files[]`.

Each `files[]` item contains:

- `role`
- `path`

Recognized roles (output-side) and their source YAML keys (input-side):

| YAML key in `context.yaml` | Normalized role | Meaning | Typical Consumer |
|----------------------------|-----------------|---------|------------------|
| `plan` | `plan` | Main plan markdown | review + execution |
| `checklist` | `checklist` | Main execution checklist | `/tdd --file` |
| `spec` | `spec` | Design/spec markdown | review + execution references |
| `testPlan` | `test-plan` | Supporting test plan | `/tdd --references` |
| `testChecklist` | `test-checklist` | Test checklist mapped to impl phases | `/tdd --test-checklist` |
| `bddPlan` | `bdd-plan` | User-behavior design keyed by domain Behavior IDs and, only when justified, real E2E IDs | `/tdd` BDD-Gate reference |
| `bddChecklist` | `bdd-checklist` | Behavior evidence and execution checklist; E2E assets are conditional | `/tdd` BDD-Gate prerequisite/reference |
Unknown metadata, spec, and target fields fail validation; the validator never
silently preserves or ignores manifest extensions.

Spec-centric full example:

```yaml
apiVersion: plancontext.agent.dev/v1alpha1
kind: PlanContext
metadata:
  name: 001-backend
spec:
  defaultTarget: backend
  targets:
    backend:
      plan: ./plan.md
      checklist: ./checklist.md
      spec: ../../spec.md
      testPlan: ./test-plan.md            # optional
      testChecklist: ./test-checklist.md  # optional, feeds /tdd --test-checklist
      bddPlan: ./bdd-plan.md              # optional; user-observable behavior only
      bddChecklist: ./bdd-checklist.md    # optional; behavior evidence + progress
```

Role mapping rules:

- There must be exactly one `checklist`.
- `test-checklist`, when present, is passed to `/tdd --test-checklist`.
- `bdd-plan`, when present, is passed as a reference to `/tdd` for BDD-Gate verification. It distinguishes domain Behavior IDs backed by code-level behavior tests from real E2E IDs backed by running-product HTTP/UI flows.
- `bdd-checklist`, when present, is passed as a reference to `/tdd`; its behavior definition and evidence-entry prerequisites must be complete before execution, and `/tdd` records the execution/evidence result before the related main checklist `BDD-Gate` is marked complete.
- A domain Behavior ID does not create a `test/scenarios/e2e/` directory. An `E2E.*` ID requires a real running frontend/backend flow and cannot wrap Go/Vitest/pytest/lint/build commands.
- All other markdown files remain read-only references.
- Consumers must keep file order stable and deduplicate by absolute path.

## Conditional Test/BDD Document Rules

The validator intentionally checks only manifest shape, path boundaries, and referenced file existence. It does not decide whether a plan should have test or BDD documents.

Document-level consumers enforce these rules:

- Code plan requires TDD: plans that introduce front-end, back-end, tooling, migration, codegen, or test helper logic must declare a TDD strategy and executable test assertions.
- Feature plan requires BDD: plans that introduce user-visible UI, API behavior, or business workflow must include `bddPlan`, `bddChecklist`, and Behavior-ID or justified real-E2E-ID `BDD-Gate:` items.
- Pure internal/config/tooling/migration/codegen plans without a user behavior flow must declare `BDD-N/A`, document the substitute verification gate, generate no BDD files, and retain no `bddPlan` / `bddChecklist` fields.

`/design`, `/create-doc`, `/plan-review`, and `/implement` enforce these document-level rules; `validate_context.py` remains limited to schema and path validation.

## Exact Minimal Schema

`metadata` contains exactly one field: `name`. Subject, sequence, Spec version,
base branch, and feature-branch stem are derived from the plan path, current Spec
Header, `AGENTS.md`, and Git rather than duplicated in the manifest.

`spec.discovery` is forbidden; target-level `discovery` and `references` are forbidden.
The only allowed target keys are `plan`, `checklist`, `spec`,
`testPlan`, `testChecklist`, `bddPlan`, and `bddChecklist`.

Branch resolution priority is:

1. `AGENTS.md` project-level Git branch strategy
2. Git default branch auto-detection

Before creating a new feature branch, `/implement` must update the resolved base
branch to the latest upstream state with fast-forward-only semantics. If the base
branch cannot be updated cleanly, `/implement` must stop before branch creation.

`/implement` derives the branch stem from `{subspec}-{plan}` and still owns the
type/date/collision suffixes.

## Shared Error Templates

### Missing `context.yaml`

```text
ERROR: docs/spec/{subspec}/plans/{plan}/context.yaml not found.
Create a context.yaml manifest to use plan-context skills. See docs/spec/TEMPLATES.md for template.
```

### Unknown target

```text
ERROR: Target '{target}' not found.
Available targets: {list of target keys}
```

### Declared file missing

```text
ERROR: Referenced file does not exist:
  - {path1}
  - {path2}
Fix the paths in context.yaml and retry.
```

### Path escapes `docs/`

```text
ERROR: Path escapes docs/ boundary:
  - {path} resolves outside docs/
Fix the relative paths in context.yaml.
```
