# Rubric Truth Source

This directory is the F3 rubric registry truth source. Every baseline rubric
shipped to staging or production is a YAML file under
`<feature_key>/<version>[.<language>].yaml`.

The `prompt-rubric-registry` Go package, the `scripts/lint/rubric_lint.py`
linter, and the seed migration in `migrations/` all derive their behavior from
the rules in this README. Any change to the field set, dimension allowlist,
or weight tolerance must update this README first; the Go loader and Python
lint share this description as their canonical specification.

## 1 File layout

```
config/rubrics/
  README.md                                 (this file — schema description)
  <feature_key>/
    <version>.yaml                          (language: multi default)
    <version>.<language>.yaml               (optional semantic language override)
```

- `<feature_key>` is one of the 6 baseline keys frozen in
  `docs/spec/prompt-rubric-registry/spec.md` §3.1.1 and must match the parent
  directory of the matching prompt under `config/prompts/`.
- `<version>` is a SemVer literal of the form `vMAJOR.MINOR.PATCH`; baseline
  starts at `v0.1.0`.
- `<language>` is either `multi` (omitted from the filename) or an ISO-639
  lowercase code such as `en` or `zh`. Baseline rubrics ship only the
  canonical `multi` coordinate. Add a concrete language override only when
  the evaluation standard itself genuinely differs by language or region.
  User-visible translation of rubric labels belongs in UI/i18n surfaces, not
  duplicated rubric truth-source files.

## 2 YAML schema

```yaml
feature_key: <string>
version: <semver-literal>
language: <multi | iso-639>
status: <active | inactive>
dimensions:
  - name: <allowlisted-dimension-name>
    weight: <0.0-1.0>
    score_levels:
      - label: <string>
        threshold: <numeric>
        description: <string>
```

Rules:

1. **`feature_key`** must equal the parent directory name and the matching
   prompt's `feature_key`.
2. **`version`** matches the lint regex
   `^v\d+\.\d+\.\d+(-[A-Za-z0-9\.-]+)?(\+[A-Za-z0-9\.-]+)?$`.
3. **`language`** is `multi` or a lowercase ISO-639 code; the filename suffix
   must match this field exactly.
4. **`dimensions`** is a non-empty array. Each dimension must have:
   - `name`: a string from the allowlist in §3.
   - `weight`: a non-negative float. The sum of all `weight` values within a
     rubric must equal `1.0` with absolute tolerance `±0.001`.
   - `score_levels`: at least three entries, each with `label`, `threshold`,
     `description`.
5. **`status`** is activation metadata and must be `active` or `inactive`.
   Exactly one version is active for each `(feature_key, language)` coordinate.
   Published dimensions, weights, score levels, version, and language are
   immutable; a gated release may change only status metadata.

## 3 Dimension name allowlist

Spec §4.1 reserves the dimension namespace for F1/F3 quality metrics plus
business-domain extensions. The lint gate accepts the following names:

**F1/F3 quality metrics** (cross-feature):

- `followup_relevance`
- `report_specificity`
- `score_outlier`
- `language_consistency`

**Business-domain dimensions** (per feature_key family):

- Practice family: `practice_depth`, `practice_dimension_coverage`,
  `practice_signal_strength`, `practice_clarity`
- Report family: `report_evidence`, `report_action_quality`,
  `report_calibration`
- Resume family: `resume_match`, `resume_impact`, `resume_clarity`,
  `resume_truthfulness`
- Target family: `target_extraction_completeness`, `target_field_accuracy`

A dimension name not on this list fails `make lint-rubrics`. Adding a new name
requires (a) adding it here, (b) bumping the linter's allowlist constant, and
(c) referencing the spec section that introduced it.

## 4 Score level conventions

- Each `score_levels` array is ordered low → high by `threshold`.
- `threshold` is the lower-bound score that activates the level.
- `label` is a short human-readable name such as `weak`, `developing`,
  `proficient`, `strong`. Labels are not enumerated by the linter so feature
  owners may keep a domain-specific vocabulary, but they should be stable
  across optional language overrides of the same feature_key.
- `description` is a one-sentence operational definition. It must not
  reference user-identifiable information or include model/provider strings.

## 5 Language coordinate convention

- `<feature_key>/<version>.yaml` is the canonical baseline (`language:
  multi`).
- Optional overrides are `<feature_key>/<version>.<language>.yaml`. The
  filename language tag must match the YAML `language` field.
- Baseline feature keys must ship the `multi` coordinate. They must not ship
  duplicate `en`, `zh`, or other language copies unless a spec/plan records
  the semantic reason for the override. When an override exists, the prompt
  and rubric language coordinate sets for that feature_key must match so
  `ResolveActive` never returns a prompt without the corresponding rubric.

## 6 Weight sum tolerance

`sum(weight)` must equal `1.0` within absolute tolerance `±0.001`. The lint
script computes the running sum in float64 and compares against
`abs(total - 1.0) <= 0.001`. Implementations may use higher precision but
must remain within this bound. Authors should keep weights at three decimal
places to stay well inside the tolerance window.

## 7 Forbidden values

The lint gate rejects:

- Module names outside current product scope. The exact list lives
  in `scripts/lint/rubric_lint.py` so this README can stay under the same
  recursive grep scan as the rubrics it governs.
- Provider names, model identifiers, or secret material in `description`
  fields.

## 8 Activation and rollback

The loader validates the complete prompt/rubric version snapshot before one
atomic pointer swap. The coordinated report/practice v0.2 release keeps both
v0.2 rubrics `active` and their v0.1 rollback coordinates `inactive`; a
rollback reverses all four rubric statuses together with the paired prompt
statuses and reloads the whole snapshot. Database activation is a separate
transactional truth substrate; this file contract makes no cross-media
atomicity claim.

## 9 References

- Spec: `docs/spec/prompt-rubric-registry/spec.md`
- Plans: `docs/spec/prompt-rubric-registry/plans/001-baseline/plan.md`,
  `docs/spec/prompt-rubric-registry/plans/003-language-coordinate-simplification/plan.md`,
  `docs/spec/prompt-rubric-registry/plans/004-real-model-profile-and-evals/plan.md`
- DB schema: `migrations/000001_create_baseline.up.sql` (`rubric_versions` table)
- Lint: `scripts/lint/rubric_lint.py`
- Go loader: `backend/internal/ai/registry/loader.go`
- Sibling prompt README: `config/prompts/README.md`
