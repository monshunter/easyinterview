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

When a concrete target is selected, the normalized payload may also include:

- `discovery`: top-level `spec.discovery` metadata
- `targetDiscovery`: target-level `spec.targets.<target>.discovery` metadata
- `baseBranch`: optional `metadata.baseBranch` string from `context.yaml`
- `branch`: optional `metadata.branch` string from `context.yaml`

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
| `bddPlan` | `bdd-plan` | BDD scenario design document keyed by scenario IDs | `/tdd` BDD-Gate reference |
| `bddChecklist` | `bdd-checklist` | BDD scenario asset and execution checklist | `/tdd` BDD-Gate prerequisite reference |
| `references[]` | `reference` | Additional markdown references (array) | review + execution references |

> YAML field keys are strictly camelCase; normalized output roles are strictly kebab-case. Do not mix the two conventions. `validate_context.py` silently ignores kebab-case YAML keys, which drops files from the normalized set.

Spec-centric full example:

```yaml
apiVersion: plancontext.agent.dev/v1alpha1
kind: PlanContext
metadata:
  subspec: example
  name: 001-backend
  sequence: 1
  supersedes: []
  specVersion:
    from: null
    to: 1.0
spec:
  defaultTarget: backend
  targets:
    backend:
      plan: ./plan.md
      checklist: ./checklist.md
      spec: ../../spec.md
      testPlan: ./test-plan.md            # optional
      testChecklist: ./test-checklist.md  # optional, feeds /tdd --test-checklist
      bddPlan: ./bdd-plan.md              # optional, feeds /tdd BDD-Gate
      bddChecklist: ./bdd-checklist.md    # optional, BDD assets + verification progress
      references:                         # optional, read-only refs
        - ./release-notes.md
```

Role mapping rules:

- There must be exactly one `checklist`.
- `test-checklist`, when present, is passed to `/tdd --test-checklist`.
- `bdd-plan`, when present, is passed as a reference to `/tdd` for BDD-Gate item verification; `/tdd` treats `bdd-plan.md` as the BDD scenario source keyed by layer scenario IDs.
- `bdd-checklist`, when present, is passed as a reference to `/tdd`; all referenced scenario asset/execution items must be complete before the related main checklist `BDD-Gate` can be marked complete.
- All other markdown files remain read-only references.
- Consumers must keep file order stable and deduplicate by absolute path.

## Discovery Metadata

`context.yaml` may also include discovery metadata for issue-intake routing:

- `spec.discovery`
- `spec.targets.<target>.discovery`

Execution-focused consumers such as `/implement`, `/plan-review`, and
`/plan-code-review` must treat these fields as read-only metadata and must not
change their execution semantics based on them.

The shared validator applies type checks when these fields are present; their
absence is allowed.

Allowed discovery fields:

- top-level `spec.discovery`: `aliases`, `keywords`, `relatedSpecs`, `relatedBugs`
- target-level `spec.targets.<target>.discovery`: `packages`, `uiRoutes`, `apiNames`

`context.yaml` is a retrieval index, not an execution runbook. Do not store
`commands`, Make targets, shell snippets, or manual operation steps in it. Those
belong to repo README / INDEX / scenario documents.

## Branch Metadata

`context.yaml` may declare optional branch lifecycle hints under `metadata`:

| Field | Type | Required | Meaning |
|------|------|----------|---------|
| `metadata.baseBranch` | string | No | Base branch and merge target used by `/implement` Step 4.5 |
| `metadata.branch` | string | No | Feature branch name stem used by `/implement` Step 4.5 before the date/collision suffix is appended |

Rules:

- Both fields are optional and must be strings when present.
- These fields are not markdown references and are not subject to `.md` suffix checks or `docs/` boundary checks.
- `validate_context.py` must preserve these fields in normalized JSON output when present.
- `generate_context_yaml.py` must preserve these fields during reconciliation.
- `/implement` consumes them as execution metadata, not as document references.

`metadata.baseBranch` priority is:

1. `context.yaml` `metadata.baseBranch`
2. `AGENTS.md` project-level Git branch strategy
3. Git default branch auto-detection

`metadata.branch` only overrides the human-authored branch stem. `/implement` still owns
the generated suffixes such as `-{MMDD}` and collision suffixes like `-{MMDD}-2`.

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
